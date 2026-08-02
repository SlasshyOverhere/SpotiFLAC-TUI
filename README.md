# SpotiFLAC-CLI

An interactive terminal (TUI) client for downloading **lossless (FLAC) music**.
It uses the **same engine and extension ecosystem as SpotiFLAC Mobile** — a
keyboard-driven replica of the mobile app, not a rewrite. The heavy lifting
(extension runtime, FLAC encoding, tag embedding, cover art, lyrics, provider
fallback, signed sessions) is inherited from the shared `go_backend` module via
a git submodule, so your favorite mobile extensions work here too.

![Tabs](https://img.shields.io/badge/tabs-Home%20%7C%20Store%20%7C%20Extensions%20%7C%20Downloads%20%7C%20Settings-81g)

---

## Install

Grab the latest binary for your platform from the
[Releases](https://github.com/SlasshyOverhere/SpotiFLAC-TUI/releases) page.

| Platform | Arch | Asset |
|---|---|---|
| Linux | amd64 / arm64 / arm / 386 | `spotiflac_linux_<arch>.tar.gz` |
| macOS | amd64 / arm64 | `spotiflac_darwin_<arch>.tar.gz` |
| Windows | amd64 / arm64 | `spotiflac_windows_<arch>.zip` |
| FreeBSD / OpenBSD / NetBSD / illumos | amd64 | `spotiflac_<os>_<arch>.tar.gz` |

Example (Linux x64):

```bash
curl -sL https://github.com/SlasshyOverhere/SpotiFLAC-TUI/releases/latest/download/spotiflac_linux_amd64.tar.gz | tar xz
./spotiflac_linux_amd64
```

Each asset ships with a `.sha256` checksum for verification.

> **Note:** the release robot builds a fresh binary from upstream source every
> night when the backend changes, then tiles it our binary with a versioned tag
> (`vYYYY.MM.DD.HHMM`). The CLI also **self-updates at runtime** — it checks for
> a newer release and swaps itself in place, so you usually never update
> manually.

---

## Build from source

Requires **Go 1.26.5+** and an initialized submodule.

```bash
git clone --recurse-submodules https://github.com/SlasshyOverhere/SpotiFLAC-TUI.git
cd SpotiFLAC-TUI

# if you already cloned without --recurse-submodules
git submodule update --init --recursive

go build -o spotiflac ./cmd/spotiflac
./spotiflac
```

The backend is pulled in via a `replace` directive pointing at the
`SpotiFLAC-Mobile` submodule (see `go.mod`). Cross-compile for other platforms
with the usual `GOOS`/`GOARCH` env vars:

```bash
GOOS=windows GOARCH=amd64 go build -o spotiflac.exe ./cmd/spotiflac
GOOS=darwin GOARCH=arm64 go build -o spotiflac ./cmd/spotiflac
```

---

## Usage

Navigate with `Tab` / `Shift-Tab` or `1`–`5`. Mouse works too.

| Tab | Keys | What it does |
|---|---|---|
| **1 Home** | `e` edit, `enter` search, `j/k` move, `d` download, `a` all | Paste a track/album/playlist/artist URL **or** type a search query |
| **2 Store** | `r` add repo, `i` install, `R` refresh, `s` switch repo | Browse and install extensions from a registry |
| **3 Extensions** | `x` toggle, `j/k` reorder, `m` save, `v` verify | Enable/disable providers and set priority; signed-session auth |
| **4 Downloads** | `j/k` move, `c` cancel | Live queue with per-item progress bars |
| **5 Settings** | `e` edit dir, `q` quality, `m` metadata, `l` lyrics, `s` save | Output dir, default quality, metadata/lyrics toggles |

Global: `q` quit, `?` help, `ctrl+c` exit.

> Signed-session providers (e.g. Tidal/Qobuz) may prompt for interactive
> verification via the `v` key. Some extensions need the host `ffmpeg` binary
> for container conversion — the app degrades gracefully if it's missing.

---

## How updates work

1. **Extensions** auto-update live inside the running TUI from the Store.
2. **Backend** changes land via a pinned `SpotiFLAC-Mobile` submodule, bumped
   automatically by [Renovate](renovate.json).
3. **Binary** publishes via a scheduled robot job; the CLI self-updates at
   runtime. See `.github/workflows/release.yml`.

---

## Development

- `internal/ui/` — bubbletea model: views, keyboard, mouse, rendering.
- `internal/backend/` — thin wrappers over the shared `gobackend` JSON API.
- `internal/app/` — persistent config (stored as JSON).
- `internal/selfcheck/` — one runnable end-to-end check (downloads a known
  track, asserts FLAC magic bytes + a tag frame).

See [PLAN.md](PLAN.md) for the design and roadmap.

---

## License

Same as upstream SpotiFLAC Mobile.