# SpotiFLAC-CLI — Plan (TUI)

## Goal
An **interactive terminal TUI** that downloads lossless (FLAC) music using the
**same engine and extension ecosystem as SpotiFLAC Mobile** — a keyboard-driven
replica of the mobile app, not a rewrite.

## Core insight
The heavy lifting already exists as a standalone Go module:

- **Module:** `github.com/zarz/spotiflac_android/go_backend`
  (`go_backend/` inside `spotiflacapp/SpotiFLAC-Mobile`), package `gobackend`.
- It already exports, as plain Go funcs, everything a TUI needs:
  - `HandleURLWithExtensionJSON(url)` — resolve a track/album/playlist/artist
    URL to full metadata **through extensions** (the mobile "paste URL" flow).
  - `SearchTracksWithMetadataProvidersJSON(query, limit, includeExts)` — search.
  - `DownloadWithExtensionsJSON(reqJSON)` / `DownloadByStrategy(reqJSON)` —
    download, FLAC encode, embed metadata/cover/lyrics, provider fallback.
  - `InitExtensionSystem(dir, dataDir)`, `LoadExtensionFromPath`,
    `GetInstalledExtensions`, `SetExtensionEnabledByID`, provider priority.
  - `InitExtensionRepoJSON(cacheDir)`, `SetRepoRegistryURLJSON`,
    `GetRepoExtensionsJSON`, `DownloadRepoExtensionJSON` — the Store/registry.
  - `GetAllDownloadProgress()`, `GetAllDownloadProgressDelta()`, `CancelDownload`.

So the TUI is a **thin Go shell**: a bubbletea UI over JSON in/out of these
funcs. We do NOT reimplement: goja extension runtime, FLAC encoding, tag
embedding, cover art, lyrics, provider fallback, quality probing, URL matching,
signed sessions.

**Not chosen:** desktop SpotiFLAC (afkarxyz) — GUI, no headless mode; and a pure
CLI — user wants a TUI.

## Repo layout
```
SpotiFLAC-CLI/
├── go.mod                     # module spotiflac-cli; go 1.26.5+
│                              #   require github.com/zarz/spotiflac_android/go_backend
│                              #   replace  ... => ../SpotiFLAC-Mobile/go_backend  (submodule)
│                              #   require github.com/charmbracelet/bubbletea  (+ lipgloss)
├── .gitmodules                # submodule: spotiflacapp/SpotiFLAC-Mobile
├── renovate.json              # auto-bump the submodule on upstream changes
├── .github/workflows/verify.yml
├── cmd/spotiflac/main.go      # bootstrap: init backend, start bubbletea program
└── internal/
    ├── app/                   # persistent state: output dir, quality default,
    │                          #   provider priority, registries (JSON file)
    ├── backend/               # thin wrappers around gobackend JSON funcs
    ├── ui/                    # bubbletea model: views, keyboard, render
    │   ├── home.go            #   search box + paste-URL box
    │   ├── store.go           #   repo registry + extension browser
    │   ├── extensions.go      #   installed extensions: enable/disable, priority
    │   ├── downloads.go       #   queue list + per-item progress bars
    │   ├── settings.go        #   output dir, default quality, metadata/lyrics toggles
    │   └── queue.go           #   background worker goroutine + tea messages
    └── selfcheck/             # one runnable end-to-end check (assert FLAC + tags)
```

Toolchain: **Go 1.26.5+** (go_backend requires it). Dependencies: `dop251/goja`,
`go-flac/*`, `utls`, `x/net`, `x/crypto` (from backend) + **bubbletea/lipgloss**
(for the TUI). All cross-platform; iOS-tagged backend files are excluded on
Linux by build tags.

## UI: views (mirror the mobile app)
Tab navigation: `1 Home · 2 Store · 3 Extensions · 4 Downloads · 5 Settings`

- **Home** — text input to paste a track/album/playlist URL **or** a search
  query. Enter → `HandleURLWithExtensionJSON` (URL) or provider search →
  results list (track name / artist / album / quality source). Select → enqueue.
- **Store** — manages repo registry URLs; browse `GetRepoExtensionsJSON`,
  `ext install` with one keypress, shows installed/update states.
- **Extensions** — installed list; enable/disable, set **provider priority**
  (drag ordering), per-extension settings + signed-session verification flow
  (prints auth URL, polls `GetPendingAuthRequestsJSON` until done).
- **Downloads** — queue of `DownloadRequest`s; live progress bars per item
  (`GetAllDownloadProgress`), pause/cancel (`CancelDownload`), per-item result
  (path, bit depth, sample rate, codec).
- **Settings** — output dir (default `~/Music/SpotiFLAC`), default quality,
  embed metadata/lyrics/cover/replaygain toggles, filename format.

Background worker: one goroutine owns the download queue; on completion it
sends a `tea.Cmd`/message back into the bubbletea loop. UI polls progress via a
ticker (bubbletea `tea.Tick`) so the UI never blocks on a download.

## The only new code
1. **URL → DownloadRequest mapping** — `HandleURLWithExtensionJSON(url)`
   returns normalized track metadata JSON; map onto the `DownloadRequest`
   struct (`track_name`, `artist_name`, `album_name`, `isrc`, `item_id`,
   `service`, `source`, `spotify_id`, `deezer_id`, `quality`, `output_dir`,
   `filename_format`, `embed_metadata`, `embed_lyrics`, `use_extensions: true`).
   Mobile does this in Dart; we do it in Go.
2. **The TUI glue** — views + queue worker (bubbletea model).

## Auto-update (your requirement) — three layers
Goal: when `spotiflacapp/SpotiFLAC-Mobile` changes, we get it with **zero manual
work**. Three layers, because "the source" and "the binary" update differently.

### Layer 1 — Extensions: auto-update live, inside the running TUI
Extensions are interpreted JS (`.sflx`) loaded from disk — **not compiled in**.
New sources, provider fixes, features all ship as extension updates through the
Store and hot-reload. This is inherited from `go_backend` for free and covers
the majority of real-world updates. Nothing to build.

### Layer 2 — Dev: submodule + Renovate + CI gate
`go_backend` is compiled into our binary, so upstream changes need a rebuild.

1. **Git submodule** `SpotiFLAC-Mobile` pinned to upstream `main` commit.
2. **Renovate** (`renovate.json`) with the `git-submodules` manager enabled
   (default on in Renovate). Scheduled; detects upstream commits; opens a PR
   bumping the submodule commit hash. Free, like Dependabot.
   - Alternative if we avoid Renovate: a scheduled GitHub Action running
     `git submodule update --remote` + `git commit` + open PR.
3. **CI gate** (`.github/workflows/verify.yml`): on every PR, `go build ./...`
   + `go vet` + run `internal/selfcheck` (download one known track to a temp
   dir, assert FLAC magic bytes `fLaC` + a tag frame). If upstream broke the
   build, the bump PR fails — never auto-merge a red build.
4. Renovate `automerge` only when CI passes → updates land hands-free.

### Layer 3 — Release + runtime self-update of the TUI binary
The constraint: `go_backend` is compiled into our binary; upstream ships mobile
binaries (APK/iOS) that can't run on desktop, so a fresh Linux binary must be
**built from upstream source** whenever it changes. Solution: a **robot is the
publisher — you never release or update manually again.**

1. **Watch upstream (already Layer 2):** Renovate auto-merges submodule bumps,
   so our `main` always tracks upstream.
2. **Robot releases** (`.github/workflows/release.yml`, scheduled nightly):
   compares the backend SHA in `main` to the last release. If changed, it
   builds the TUI, tags it (`backend-<sha>`), and publishes binary +
   `sha256sum` to **our** GitHub Releases. No human publishes anything.
3. **Runtime self-update** in the TUI: on launch (and every N hours via
   `tea.Tick`), check our latest release; if newer than the running version:
   download asset → verify checksum → atomically replace the running binary
   (`os.Rename` works over a live binary on Linux/macOS) → `syscall.Exec` the
   new binary in place. Hand-rolled, ~60 lines, no new dependency.

**Result:** upstream pushes → within ~a day the TUI is running the new backend.
You do nothing: no manual releases, no manual updates.

**Alternative if we don't want even our own releases:** the TUI watches upstream
source directly and **recompiles itself at runtime** (download source → `go
build` using Go's auto-fetched `toolchain` directive → swap → re-exec). No CI,
no releases — but needs a Go toolchain on the machine, slower/fragile at update
time. Recommended only as a fallback.

> Honest ceiling: the compiled backend cannot be hot-swapped inside a live
> process — the swap happens between check and re-exec. Extensions, which are
> the frequent-change layer, update live with no restart.

> If upstream renames the module path or moves `go_backend`, the fix is a
> one-line `replace` update in `go.mod` — surfaced immediately by CI.

## Milestones
1. **Scaffold:** submodule + go.mod replace + `main.go` that imports
   `gobackend` and renders a minimal bubbletea shell. Verify `go build`.
2. **Download one track:** Home URL paste → resolve → download → selfcheck green.
3. **Downloads view:** queue + live progress bars + cancel.
4. **Store + Extensions views:** repo add, install, enable/disable, priority.
5. **Settings + verification flow:** output dir, quality flags, signed-session
   auth.
6. **Auto-update:** submodule + Renovate + CI + selfcheck wired and green.
7. **Self-update:** release CD + runtime check/swap/re-exec.

## Risks / notes
- **Go 1.26.5** toolchain is very new; CI must pin it (or use Go's `toolchain`
  directive to auto-download).
- Signed-session extensions (Tidal/Qobuz) need interactive verification —
  plan a `verify` flow in the Extensions view (print auth URL, poll).
- Some extensions need the host `ffmpeg` binary for container conversion
  (same as the mobile runtime) — degrade gracefully if absent.
- Quality/source availability varies per extension (same as the app: provider
  priority + enabling more providers).
- TUI must stay responsive while downloading → queue worker in a goroutine,
  progress via `tea.Tick`, never block the bubbletea loop on network/JSON calls.

## Non-goals (v1)
- No GUI, no library scanning UI, no embedded player.
- No reimplementation of any backend feature.
