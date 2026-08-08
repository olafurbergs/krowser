# krowser

```
██╗  ██╗ ██████╗  ██████╗  ██╗    ██╗ ███████╗ ███████╗ ██████╗ 
██║ ██╔╝ ██╔══██╗ ██╔═══██╗ ██║ ██╗ ██║ ██╔════╝ ██╔════╝ ██╔══██╗
█████╔╝  ██████╔╝ ██║   ██║ ██║╚██╗██║ ███████╗ █████╗   ██████╔╝
██╔═██╗  ██╔══██╗ ██║   ██║ ██║ ╚████║ ╚════██║ ██╔══╝   ██╔══██╗
██║  ██╗ ██║  ██║ ╚██████╔╝ ██║  ╚███║ ███████║ ███████╗ ██║  ██║
╚═╝  ╚═╝ ╚═╝  ╚═╝  ╚═════╝ ╚═╝   ╚══╝ ╚══════╝ ╚══════╝ ╚═╝  ╚═╝
```

[![asciicast](https://asciinema.org/a/9uYbHxn0acq5NC20.svg)](https://asciinema.org/a/9uYbHxn0acq5NC20)

A slick Kubernetes TUI browser built with [Bubble Tea v2](https://charm.land/bubbletea/v2),
[Bubbles v2](https://charm.land/bubbles/v2), and [Lip Gloss v2](https://charm.land/lipgloss/v2).

Browse clusters, namespaces, and resources, stream syntax-highlighted logs, scale and delete
workloads, and manage port-forwards — all from your terminal.

## Features

- **Navigation** — context, namespace, and resource (kind) pickers with k9s-inspired keybindings
- **Resource tables** — typed columns, `/` fuzzy filtering, and auto-refresh
- **Detail views** — YAML and `describe` output with selectable Secret data (decoded in place)
- **Pod logs** — live follow, container switching, timestamp toggle, and
  **syntax highlighting** for log levels, timestamps, numbers, URLs, IPs, UUIDs, and more
- **Mutating actions** — delete, scale, rollout restart, and edit, all behind a confirm dialog
- **Port-forwards** — open and manage port-forwards from the cluster
- **CPU/memory gauges** — a live `top` view with gradient progress bars when `metrics-server` is available
- **Themes** — 10 color themes, selectable with `T` or `--theme`, with the selection persisted across runs
- **Extras** — toast notifications, focus blur, help overlay

## Demo

A ~45s walkthrough recording showing a `kind` cluster being browsed with krowser:
resource tables, syntax-highlighted live logs, YAML detail, theme switching, and the
CPU/memory gauges.

```sh
asciinema play docs/krowser-demo.cast
```

The cast is also ready to upload for an embedded player:

```sh
asciinema upload docs/krowser-demo.cast
```

## Requirements

- Go 1.25+ (the `go 1.25` directive triggers an automatic toolchain download on 1.21+)
- A kubeconfig with access to a cluster (KUBECONFIG or `~/.kube/config`)
- A terminal with truecolor support for the full palette

## Install

```sh
go install github.com/olafurb/krowser@latest
```

## Build & run

```sh
make build
./bin/krowser --kubeconfig ~/.kube/config
```

The Makefile also provides `make test` (with `-race`), `make lint`, and `make fmt`.

### CLI flags

| Flag | Description |
| --- | --- |
| `--kubeconfig PATH` | path to the kubeconfig (default: `$KUBECONFIG` or `~/.kube/config`) |
| `--context NAME` | kubeconfig context to use |
| `--namespace NAME` | namespace to browse (default: the context namespace, or all namespaces if none is set) |
| `--all-namespaces` | browse all namespaces |
| `--theme NAME` | theme name (default: `Monokai`) |

## Usage

### Keybindings

| Key | Action |
| --- | --- |
| `↑` / `↓` | navigate rows |
| `Enter` | open detail (a pod/node in `top` too) |
| `y` / `d` | YAML / describe |
| `l` | stream logs |
| `/` | fuzzy filter (type to filter, `esc` clears) |
| `n` / `g` | namespace picker / all namespaces |
| `u` | switch context |
| `k` | pick a kind |
| `r` | refresh |
| `x` / `s` / `R` / `e` | delete / scale / restart / edit |
| `f` | port-forward |
| `t` | top (metrics) |
| `T` | themes |
| `?` | help overlay |
| `q` / `esc` | back |
| `Ctrl-C` | quit |

### Logs screen

| Key | Action |
| --- | --- |
| `f` | toggle follow mode |
| `t` | toggle timestamps |
| `c` | pick a container |
| `q` / `esc` / `←` | back |

## Theming

Press `T` to open the theme picker or pass `--theme "Dracula"` on startup. Themes: Monokai,
One Dark, Dracula, Nord, Gruvbox Dark, Solarized Dark, Catppuccin Mocha, Tokyo Night,
Solarized Light, and Catppuccin Latte.

The theme you select is persisted to `~/.local/share/krowser/theme`
(`$XDG_DATA_HOME/krowser/theme` if set) and restored on the next launch. A `--theme` flag
takes precedence over the saved selection.

## Project layout

```
cmd/krowser       CLI entry point (flags, kubeconfig wiring)
internal/k8s      Kubernetes client helpers (list, logs, top, forwards, actions)
internal/tui      Bubble Tea application
  app.go          root model, navigation, theming, chrome
  resource.go     typed resource tables with filtering
  detail.go       YAML / describe / secret views
  logs.go         streamed, syntax-highlighted pod logs
  top.go          live CPU/memory gauges
  picker.go       contexts, namespaces, kinds, containers, themes
  theme(s).go     theme registry and palette
  loghl/          log syntax-highlighting engine (see Acknowledgments)
```

## Testing

```sh
make test
```

The suite covers the resource tables, pickers, themes, metrics formatting, and the
log-highlighting engine (including loglit's own pattern-gate and overlap tests).

## Acknowledgments

- **Log highlighting engine** — vendored from
  [madmaxieee/loglit](https://github.com/madmaxieee/loglit) (MIT, © 2025 Tsng, Kahiok).
  loglit keeps its engine in Go `internal/` packages that cannot be imported from other
  modules, so it is copied verbatim into `internal/tui/loghl/` (patterns, syntax gating,
  keyword matching, and ANSI rendering) and recolored from krowser's themes. Its license
  lives in `internal/tui/loghl/LICENSE`.
- **UI framework** — [Bubble Tea v2](https://charm.land/bubbletea/v2),
  [Bubbles v2](https://charm.land/bubbles/v2), [Lip Gloss v2](https://charm.land/lipgloss/v2).
- **Fuzzy filtering** — [sahilm/fuzzy](https://github.com/sahilm/fuzzy).
- **Theme inspiration** — loglit's palettes draw on the Tokyo Night and
  [log-highlight.nvim](https://github.com/fei6409/log-highlight.nvim) themes.

## License

[MIT](LICENSE). The vendored loglit engine is also MIT licensed; see
`internal/tui/loghl/LICENSE`.
