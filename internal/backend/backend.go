package backend

import (
	"encoding/json"
	"fmt"
	"strings"

	gobackend "github.com/zarz/spotiflac_android/go_backend"

	"spotiflac-cli/internal/app"
)

type Track struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artists    string `json:"artists"`
	AlbumName  string `json:"album_name"`
	AlbumArtist string `json:"album_artist"`
	AlbumID    string `json:"album_id"`
	DurationMS int    `json:"duration_ms"`
	CoverURL   string `json:"cover_url"`
	ReleaseDate string `json:"release_date"`
	TrackNumber int   `json:"track_number"`
	TotalTracks int   `json:"total_tracks"`
	DiscNumber  int   `json:"disc_number"`
	TotalDiscs  int   `json:"total_discs"`
	ISRC       string `json:"isrc"`
	ProviderID string `json:"provider_id"`
	SpotifyID  string `json:"spotify_id"`
	Composer   string `json:"composer"`
	Explicit   bool   `json:"explicit"`
}

type RepoExt struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Version         string `json:"version"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	Downloads       int    `json:"downloads"`
	IsInstalled     bool   `json:"is_installed"`
	InstalledVersion string `json:"installed_version"`
	HasUpdate       bool   `json:"has_update"`
}

type InstalledExt struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name"`
	Version       string   `json:"version"`
	Description   string   `json:"description"`
	Types         []string `json:"types"`
	Enabled       bool     `json:"enabled"`
	Status        string   `json:"status"`
	HasMetadataProvider bool `json:"has_metadata_provider"`
	HasDownloadProvider bool `json:"has_download_provider"`
	HasLyricsProvider   bool `json:"has_lyrics_provider"`
	Error         string   `json:"error_message"`
}

func Init(cfg app.Config) error {
	gobackend.SetLoggingEnabled(true)
	if err := gobackend.InitExtensionSystem(cfg.ExtensionsDir(), cfg.BackendDataDir()); err != nil {
		return err
	}
	if err := gobackend.InitExtensionRepoJSON(cfg.RepoCacheDir()); err != nil {
		return err
	}
	if len(cfg.ProviderPriority) > 0 {
		_ = gobackend.SetProviderPriorityJSON(mustJSON(cfg.ProviderPriority))
	}
	if len(cfg.Registries) > 0 {
		_ = gobackend.SetRepoRegistryURLJSON(cfg.Registries[len(cfg.Registries)-1])
	}
	_, _ = gobackend.LoadExtensionsFromDir(cfg.ExtensionsDir())
	exts, _ := ListInstalled()
	for _, e := range exts {
		if !e.Enabled {
			_ = gobackend.SetExtensionEnabledByID(e.ID, true)
		}
	}
	syncProviderPriority()
	return nil
}

// syncProviderPriority seeds the backend's download-provider priority from
// installed extensions when nothing is configured yet, otherwise the download
// fallback has zero providers to try.
func syncProviderPriority() {
	prio, err := ProviderPriority()
	if err == nil && len(prio) > 0 {
		return
	}
	exts, err := ListInstalled()
	if err != nil {
		return
	}
	var dl []string
	for _, e := range exts {
		if e.HasDownloadProvider {
			dl = append(dl, e.ID)
		}
	}
	if len(dl) > 0 {
		_ = gobackend.SetProviderPriorityJSON(mustJSON(dl))
		cfg := app.Default()
		cfg.Load()
		cfg.ProviderPriority = dl
		_ = cfg.Save()
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func Logs() string {
	var out struct {
		Logs []struct {
			Level   string `json:"level"`
			Tag     string `json:"tag"`
			Message string `json:"message"`
		} `json:"logs"`
	}
	_ = unmarshal(gobackend.GetLogsSince(0), &out)
	var b strings.Builder
	for _, l := range out.Logs {
		fmt.Fprintf(&b, "[%s] %s: %s\n", l.Level, l.Tag, l.Message)
	}
	return b.String()
}

func unmarshal(s string, v any) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}

// ---- Registry / Store ----

func AddRegistry(url string) error {
	regs, _ := ListRegistries()
	for _, r := range regs {
		if r == url {
			return nil
		}
	}
	if err := gobackend.SetRepoRegistryURLJSON(url); err != nil {
		return err
	}
	regs = append(regs, url)
	cfg := app.Default()
	cfg.Load()
	cfg.Registries = regs
	return cfg.Save()
}

func ListRegistries() ([]string, error) {
	cfg := app.Default()
	cfg.Load()
	return cfg.Registries, nil
}

func SetRegistry(index int) error {
	regs, err := ListRegistries()
	if err != nil || index < 0 || index >= len(regs) {
		return fmt.Errorf("invalid registry index")
	}
	return gobackend.SetRepoRegistryURLJSON(regs[index])
}

func ListRepoExtensions(force bool) ([]RepoExt, error) {
	raw, err := gobackend.GetRepoExtensionsJSON(force)
	if err != nil {
		return nil, err
	}
	var exts []RepoExt
	if err := unmarshal(raw, &exts); err != nil {
		return nil, fmt.Errorf("parse repo extensions: %w", err)
	}
	return exts, nil
}

func InstallRepoExt(id string) error {
	installed, _ := ListInstalled()
	for _, e := range installed {
		if e.ID == id {
			return nil // already installed
		}
	}
	dest := mustDownloadDir()
	path, err := gobackend.DownloadRepoExtensionJSON(id, dest)
	if err != nil {
		return err
	}
	if _, err = gobackend.LoadExtensionFromPath(path); err != nil {
		return err
	}
	if err = gobackend.SetExtensionEnabledByID(id, true); err != nil {
		return err
	}
	ext, err := ListInstalled()
	if err != nil {
		return err
	}
	for _, e := range ext {
		if e.ID == id && e.HasDownloadProvider {
			syncProviderPriority()
			break
		}
	}
	return nil
}

func InstallLocalExt(path string) error {
	_, err := gobackend.LoadExtensionFromPath(path)
	return err
}

func mustDownloadDir() string {
	cfg := app.Default()
	cfg.Load()
	return cfg.ExtensionsDir()
}

// ---- Installed extensions ----

func ListInstalled() ([]InstalledExt, error) {
	raw, err := gobackend.GetInstalledExtensions()
	if err != nil {
		return nil, err
	}
	var exts []InstalledExt
	if err := unmarshal(raw, &exts); err != nil {
		return nil, fmt.Errorf("parse installed extensions: %w", err)
	}
	return exts, nil
}

func SetEnabled(id string, enabled bool) error {
	return gobackend.SetExtensionEnabledByID(id, enabled)
}

func ProviderPriority() ([]string, error) {
	raw, err := gobackend.GetProviderPriorityJSON()
	if err != nil {
		return nil, err
	}
	var p []string
	if err := unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return p, nil
}

func SetProviderPriority(p []string) error {
	if err := gobackend.SetProviderPriorityJSON(mustJSON(p)); err != nil {
		return err
	}
	cfg := app.Default()
	cfg.Load()
	cfg.ProviderPriority = p
	return cfg.Save()
}
