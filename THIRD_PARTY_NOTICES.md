# Third Party Notices

SpotiFLAC-CLI relies on the following third-party software. Licenses are
reproduced here in accordance with their terms.

## Upstream engine

### SpotiFLAC Mobile (`go_backend`) — MIT License

The entire download, encoding, metadata, lyrics, and extension runtime is
inherited from the upstream open-source project:

- Project: https://github.com/spotiflacapp/SpotiFLAC-Mobile
- Author / copyright: Copyright (c) 2026 zarzet
- License: MIT (reproduced in `SpotiFLAC-Mobile/LICENSE`)

```
MIT License

Copyright (c) 2026 zarzet

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Direct Go dependencies — MIT License

The following Go modules are distributed under the MIT License. Full license
texts are available at their respective repositories.

- **github.com/charmbracelet/bubbles** — https://github.com/charmbracelet/bubbles
- **github.com/charmbracelet/bubbletea** — https://github.com/charmbracelet/bubbletea
- **github.com/charmbracelet/lipgloss** — https://github.com/charmbracelet/lipgloss

### MIT License (Charmbracelet)

```
MIT License

Copyright (c) 2020-2025 Charmbracelet, Inc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Other dependencies

Additional transitive Go modules (see `go.mod` for the full list) are licensed
under the MIT License, BSD-3-Clause License, Apache License 2.0, or a
compatible OSI-approved license. Full texts are available in the module
source directories within the Go module cache:

```
go env GOMODCACHE
```
