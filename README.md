# SpotiFLAC-CLI

> **⚠️ Legal disclaimer — please read before using.**
>
> This project is a **terminal user interface (TUI)** built *for* the users of
> the open-source **SpotiFLAC** ecosystem. It is **not** a service that hosts,
> distributes, streams, or stores any copyrighted content, and it does **not**
> provide access to any premium or DRM-protected catalog. The download,
> encoding, metadata, and extension capabilities come **entirely from the
> upstream open-source project**:
>
> - **SpotiFLAC Mobile** — https://github.com/spotiflacapp/SpotiFLAC-Mobile
>
> which this project only wraps in a keyboard-driven terminal client. **The
> actual author and copyright holder of the underlying engine is
> [`spotiflacapp`](https://github.com/spotiflacapp)** (MIT licensed, see
> [LICENSE](LICENSE)). This TUI does not ship, bundle, or redistribute any
> music or content of its own.
>
> **You are solely responsible for** the content you download and for ensuring
> you have the rights to use it (e.g. your own purchases, public-domain
> material, or content you are otherwise licensed to access). **This project is
> provided for interoperability with the open-source ecosystem only** and makes
> no representation that any particular use is lawful in your jurisdiction.
> Do not use it to infringe copyright. If you are a rights holder and believe
> this project misuses your work, open an issue and we will address it.

An interactive terminal (TUI) client for **lossless (FLAC) music** built on the
**same engine and extension ecosystem as SpotiFLAC Mobile** — a keyboard-driven
replica of the mobile app, not a rewrite. The heavy lifting (extension runtime,
FLAC encoding, tag embedding, cover art, lyrics, provider fallback, signed
sessions) is inherited from the shared `go_backend` module via a git submodule,
so your favorite mobile extensions work here too.

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

> **Note:** the release robot polls upstream every 30 minutes, auto-syncs the
> submodule to the latest `SpotiFLAC-Mobile` commit, and publishes a fresh
> binary with a versioned tag (`vYYYY.MM.DD.HHMMSS`) on every upstream commit.
> The CLI also **self-updates at runtime** — it checks for a newer release and
> swaps itself in place, so you usually never update manually.

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
2. **Backend** changes land via the pinned `SpotiFLAC-Mobile` submodule, which
   the release robot auto-syncs to the latest upstream commit (polled every
   30 minutes) and commits to `main`.
3. **Binary** publishes via the same robot job on every upstream commit, so
   there is always a build matching the latest upstream. The CLI self-updates
   at runtime. See `.github/workflows/release.yml`.

---

## Development

- `internal/ui/` — bubbletea model: views, keyboard, mouse, rendering.
- `internal/backend/` — thin wrappers over the shared `gobackend` JSON API.
- `internal/app/` — persistent config (stored as JSON).
- `internal/selfcheck/` — one runnable end-to-end check (downloads a known
  track, asserts FLAC magic bytes + a tag frame).

See [PLAN.md](PLAN.md) for the design and roadmap.

---

## Attribution & Licensing

- **Engine / backend:** Copyright (c) 2026 [zarzet](https://github.com/spotiflacapp),
  released under the **MIT License** by the upstream project
  [SpotiFLAC Mobile](https://github.com/spotiflacapp/SpotiFLAC-Mobile).
- **This TUI wrapper:** MIT licensed — see [LICENSE](LICENSE) for the full text,
  and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for the list of bundled
  dependencies and their licenses.
- **Notice:** This project is not affiliated with or endorsed by the upstream
  author or any music service. It is an independent, community-maintained
  terminal client that consumes the upstream open-source engine.