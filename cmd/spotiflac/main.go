package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	gobackend "github.com/zarz/spotiflac_android/go_backend"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
	"spotiflac-cli/internal/selfcheck"
	"spotiflac-cli/internal/ui"
	"spotiflac-cli/internal/update"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "--selfcheck" {
		if err := selfcheck.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "selfcheck: ", err)
			os.Exit(1)
		}
		fmt.Println("selfcheck ok")
		return
	}

	cfg := app.Default()
	cfg.Load()

	if err := backend.Init(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}

	if len(args) > 0 && args[0] == "--dl" {
		os.Exit(runHeadless(args[1:], cfg))
	}

	qm := queue.New()
	go update.Check(cfg)

	logFile := redirectStderr(cfg)
	if logFile != nil {
		defer logFile.Close()
	}

	if _, err := tea.NewProgram(ui.New(cfg, qm),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func redirectStderr(cfg app.Config) *os.File {
	logDir := cfg.BackendDataDir()
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil
	}
	f, err := os.OpenFile(filepath.Join(logDir, "spotiflac.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil
	}
	// Redirect both the Go-level variable and the underlying fd 2 so that
	// any C/goja code writing directly to stderr also lands in the log file.
	os.Stderr = f
	if err := syscall.Dup2(int(f.Fd()), syscall.Stderr); err != nil {
		f.Close()
		return nil
	}
	return f
}

// runHeadless implements scriptable mode:
//
//	spotiflac --dl <url> [--registry <url>] [--ext <id>]... [-o dir]
func runHeadless(args []string, cfg app.Config) int {
	var url, registry, out string
	var exts []string
	debug := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--registry":
			if i+1 < len(args) {
				registry = args[i+1]
				i++
			}
		case "--ext":
			if i+1 < len(args) {
				exts = append(exts, args[i+1])
				i++
			}
		case "-o", "--out":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		case "--debug":
			debug = true
		default:
			if url == "" {
				url = args[i]
			}
		}
	}

	if registry != "" {
		if err := backend.AddRegistry(registry); err != nil {
			fmt.Fprintln(os.Stderr, "registry:", err)
			return 1
		}
		fmt.Println("registry:", registry)
	}
	for _, id := range exts {
		if err := backend.InstallRepoExt(id); err != nil {
			fmt.Fprintln(os.Stderr, "install "+id+":", err)
			return 1
		}
		fmt.Println("installed:", id)
	}
	if out != "" {
		cfg.OutputDir = out
	}
	if url == "" {
		fmt.Fprintln(os.Stderr, "usage: spotiflac --dl <url> [--registry <url>] [--ext <id>]... [-o dir]")
		return 2
	}

	resolved, err := backend.ResolveURL(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve:", err)
		return 1
	}
	if len(resolved.Tracks) == 0 {
		fmt.Fprintln(os.Stderr, "no tracks resolved from", url)
		return 1
	}
	fmt.Printf("resolved: %s — %d track(s)\n", resolved.Name, len(resolved.Tracks))
	if debug {
		raw, _ := json.MarshalIndent(resolved, "", "  ")
		fmt.Println(string(raw))
	}

	ok := 0
	for _, t := range resolved.Tracks {
		req := backend.DownloadRequestForTrack(t, cfg)
		raw, err := backend.Download(backend.DownloadRequestJSON(req))
		if err != nil {
			fmt.Fprintln(os.Stderr, "download failed:", err)
			continue
		}
		var resp gobackend.DownloadResponse
		_ = json.Unmarshal([]byte(raw), &resp)
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "download failed: %s (%s)\n", resp.Error, resp.ErrorType)
			fmt.Fprintln(os.Stderr, backend.Logs())
			continue
		}
		fmt.Printf("OK  %s — %s  (%s %d-bit/%dHz)\n",
			t.Name, resp.FilePath, resp.AudioCodec, resp.ActualBitDepth, resp.ActualSampleRate)
		ok++
	}
	if ok == 0 {
		fmt.Fprintln(os.Stderr, "no downloads succeeded")
		return 1
	}
	return 0
}
