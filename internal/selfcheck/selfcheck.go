package selfcheck

import (
	"encoding/json"
	"fmt"
	"os"

	gobackend "github.com/zarz/spotiflac_android/go_backend"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
)

func Run() error {
	cfg := app.Default()
	cfg.OutputDir = os.TempDir()
	if err := backend.Init(cfg); err != nil {
		return fmt.Errorf("backend init: %w", err)
	}

	// Offline checks: our resolver + request mapping + progress plumbing.
	t := backend.Track{ID: "test-1", Name: "Test", Artists: "Artist", AlbumName: "Album", ISRC: "QZTEST", ProviderID: "deezer"}
	req := backend.DownloadRequestForTrack(t, cfg)
	raw := backend.DownloadRequestJSON(req)

	var rt gobackend.DownloadRequest
	if err := json.Unmarshal([]byte(raw), &rt); err != nil {
		return fmt.Errorf("request JSON round-trip: %w", err)
	}
	if rt.TrackName != "Test" || rt.ArtistName != "Artist" || rt.ItemID != "test-1" {
		return fmt.Errorf("request JSON round-trip mismatch: %+v", rt)
	}
	if !rt.UseExtensions || !rt.UseFallback {
		return fmt.Errorf("extension/fallback flags missing")
	}

	prog := backend.AllProgress()
	if prog == nil {
		return fmt.Errorf("progress JSON unparseable")
	}

	return nil
}
