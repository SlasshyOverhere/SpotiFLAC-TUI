package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"spotiflac-cli/internal/app"
)

var (
	Version = "dev" // set via -ldflags
	Repo    = ""    // set via -ldflags, e.g. "you/SpotiFLAC-CLI"
)

func Check(cfg app.Config) {
	if !cfg.UpdateCheck || Repo == "" || strings.HasPrefix(Version, "dev") {
		return
	}
	rel, err := latestRelease()
	if err != nil {
		return
	}
	if !newer(rel.Tag, Version) {
		return
	}
	assetURL, shaURL := pickAssets(rel)
	if assetURL == "" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	body, err := download(assetURL)
	if err != nil {
		return
	}
	if shaURL != "" && !matches(body, shaURL) {
		return
	}
	if err := replace(exe, body); err != nil {
		return
	}
	fmt.Printf("\nupdated to %s — restarting\n", rel.Tag)
	_ = syscall.Exec(exe, os.Args, os.Environ())
}

type release struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestRelease() (*release, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "spotiflac-cli")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func pickAssets(rel *release) (assetURL, shaURL string) {
	suffix := runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	for _, a := range rel.Assets {
		switch {
		case strings.HasSuffix(a.Name, ".sha256") && strings.Contains(a.Name, runtime.GOARCH):
			shaURL = a.URL
		case strings.HasSuffix(a.Name, suffix):
			assetURL = a.URL
		}
	}
	return assetURL, shaURL
}

func download(u string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download %s: status %d", u, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func matches(data []byte, shaURL string) bool {
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	raw, err := download(shaURL)
	if err != nil {
		return false
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return false
	}
	return strings.EqualFold(fields[0], actual)
}

func replace(exe string, data []byte) error {
	tmp := filepath.Join(filepath.Dir(exe), ".spotiflac-update")
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, exe)
}

func newer(tag, current string) bool {
	if current == "dev" {
		return false
	}
	t := strings.TrimPrefix(tag, "v")
	c := strings.TrimPrefix(current, "v")
	tv := strings.Split(t, ".")
	cv := strings.Split(c, ".")
	for i := 0; i < len(tv) && i < len(cv); i++ {
		ti, ci := num(tv[i]), num(cv[i])
		if ti != ci {
			return ti > ci
		}
	}
	return len(tv) > len(cv)
}

func num(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
