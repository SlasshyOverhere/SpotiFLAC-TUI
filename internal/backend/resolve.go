package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	raw, err := gobackend.HandleURLWithExtensionJSON(url)
	if err != nil {
		return nil, err
	}
	var resolved ResolvedURL
	if err := unmarshal(raw, &resolved); err != nil {
		return nil, fmt.Errorf("parse resolve result: %w", err)
	}
	if resolved.Track != nil {
		resolved.Tracks = []Track{*resolved.Track}
	}
	return &resolved, nil
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
		Source:             t.ProviderID,
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
