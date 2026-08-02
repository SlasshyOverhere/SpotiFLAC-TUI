package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	OutputDir        string   `json:"output_dir"`
	Quality          string   `json:"quality"`
	EmbedMetadata    bool     `json:"embed_metadata"`
	EmbedLyrics      bool     `json:"embed_lyrics"`
	Registries       []string `json:"registries"`
	ProviderPriority []string `json:"provider_priority"`
	UpdateCheck      bool     `json:"update_check"`
}

func Default() Config {
	return Config{
		OutputDir:     "~/Music/SpotiFLAC",
		Quality:       "lossless",
		EmbedMetadata: true,
		EmbedLyrics:   true,
		UpdateCheck:   true,
	}
}

func DataDir() string {
	if d := os.Getenv("SPOTIFLAC_DATA_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".spotiflac-cli"
	}
	return filepath.Join(home, ".local", "share", "spotiflac-cli")
}

func configPath() string {
	return filepath.Join(DataDir(), "config.json")
}

func (c *Config) Load() {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, c)
}

func (c Config) Save() error {
	if err := os.MkdirAll(DataDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0o644)
}

func (c Config) ExtensionsDir() string {
	return filepath.Join(DataDir(), "extensions")
}

func (c Config) BackendDataDir() string {
	return filepath.Join(DataDir(), "data")
}

func (c Config) RepoCacheDir() string {
	return filepath.Join(DataDir(), "repo-cache")
}
