package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gobackend "github.com/zarz/spotiflac_android/go_backend"

	"spotiflac-cli/internal/app"
)

type ResolvedURL struct {
	Type   string  `json:"type"`
	Name   string  `json:"name"`
	Track  *Track  `json:"track"`
	Tracks []Track `json:"tracks"`
}

func Search(query string, limit int) ([]Track, error) {
	raw, err := gobackend.SearchTracksWithMetadataProvidersJSON(query, limit, true)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	if err := unmarshal(raw, &tracks); err != nil {
		return nil, fmt.Errorf("parse search results: %w", err)
	}
	return tracks, nil
}

func ResolveURL(url string) (*ResolvedURL, error) {
	var resolved ResolvedURL
	for attempt := 0; attempt < 10; attempt++ {
		raw, err := gobackend.HandleURLWithExtensionJSON(url)
		if err != nil {
			return nil, err
		}
		resolved = ResolvedURL{}
		if err := unmarshal(raw, &resolved); err != nil {
			return nil, fmt.Errorf("parse resolve result: %w", err)
		}
		if resolved.Track != nil {
			resolved.Tracks = []Track{*resolved.Track}
		}
		if resolved.complete() {
			return &resolved, nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return &resolved, nil
}

// complete reports whether every resolved track has real metadata. Some
// extensions (e.g. ytmusic) return a placeholder track on first resolve and
// fill the cache asynchronously, so we poll until it's ready.
func (r *ResolvedURL) complete() bool {
	if len(r.Tracks) == 0 {
		return false
	}
	for _, t := range r.Tracks {
		if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Artists) == "" {
			return false
		}
	}
	return true
}

func expandDir(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, dir[2:])
		}
	}
	return dir
}

func DownloadRequestForTrack(t Track, cfg app.Config) gobackend.DownloadRequest {
	return gobackend.DownloadRequest{
		ISRC:               t.ISRC,
		SpotifyID:          t.SpotifyID,
		TrackName:          t.Name,
		ArtistName:         t.Artists,
		AlbumName:          t.AlbumName,
		AlbumArtist:        t.AlbumArtist,
		CoverURL:           t.CoverURL,
		OutputDir:          expandDir(cfg.OutputDir),
		FilenameFormat:     "",
		Quality:            cfg.Quality,
		EmbedMetadata:      cfg.EmbedMetadata,
		EmbedLyrics:        cfg.EmbedLyrics,
		EmbedMaxQualityCover: true,
		TrackNumber:        t.TrackNumber,
		TotalTracks:        t.TotalTracks,
		DiscNumber:         t.DiscNumber,
		TotalDiscs:         t.TotalDiscs,
		ReleaseDate:        t.ReleaseDate,
		ItemID:             t.ID,
		DurationMS:         t.DurationMS,
		Composer:           t.Composer,
		UseExtensions:      true,
		UseFallback:        true,
	}
}

func DownloadRequestJSON(req gobackend.DownloadRequest) string {
	b, _ := json.Marshal(req)
	return string(b)
}

func Download(reqJSON string) (string, error) {
	return gobackend.DownloadByStrategy(reqJSON)
}
