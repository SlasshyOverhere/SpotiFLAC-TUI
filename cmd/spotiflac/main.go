package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
	"spotiflac-cli/internal/selfcheck"
	"spotiflac-cli/internal/ui"
	"spotiflac-cli/internal/update"
)

func main() {
	for _, a := range os.Args[1:] {
		if a == "--selfcheck" {
			if err := selfcheck.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "selfcheck: ", err)
				os.Exit(1)
			}
			fmt.Println("selfcheck ok")
			return
		}
	}

	cfg := app.Default()
	cfg.Load()

	if err := backend.Init(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}

	qm := queue.New()
	go update.Check(cfg)

	if _, err := tea.NewProgram(ui.New(cfg, qm), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
