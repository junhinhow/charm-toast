# charm-toast

Toast notification component for [Bubble Tea](https://github.com/charmbracelet/bubbletea) — auto-dismissing status messages.

> **[Leia em Português (PT-BR)](README.pt-br.md)**

## Features

- **Four toast types**: Info, Success, Warning, Error
- **Auto-dismiss** with configurable duration (expired toasts are discarded during rendering)
- **Bubble Tea integration** via `ToastManager.Tick()` / `toast.ExpiredMsg`
- **ToastManager** for managing multiple notifications
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss) — colored borders and icons
- Thread-safe toast management (every manager operation is protected by a mutex)
- Right-aligned rendering for non-intrusive notifications
- Width-limited rendering — content wider than the requested width is truncated, so the view never overflows the terminal

## Install

```bash
go get github.com/junhinhow/charm-toast@latest
```

## Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/junhinhow/charm-toast"
)

func main() {
    // Single toast
    t := toast.NewToast("File saved successfully", 3*time.Second).WithType(toast.Success)
    fmt.Println(t.Render(80))

    // Toast manager for multiple notifications
    tm := toast.NewToastManager()
    tm.AddInfo("Loading...", 5*time.Second)
    tm.AddSuccess("Done!", 3*time.Second)
    tm.AddError("Connection failed", 10*time.Second)
    fmt.Println(tm.Render(80))
}
```

## Bubble Tea integration

`ToastManager.Tick()` schedules a `tea.Cmd` that emits `toast.ExpiredMsg` when the next
toast expires. Arm the tick in `Init()` and re-arm it after every message so newly added
toasts are picked up as well:

```go
package main

import (
    "time"

    tea "charm.land/bubbletea/v2"

    "github.com/junhinhow/charm-toast"
)

type model struct {
    tm *toast.ToastManager
}

func (m model) Init() tea.Cmd {
    return m.tm.Tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "a":
            m.tm.AddInfo("Working...", 3*time.Second)
        case "q":
            return m, tea.Quit
        }
    case toast.ExpiredMsg:
        // Drop the toasts that are no longer valid.
        m.tm.Cleanup()
    }
    // Re-arm the expiration tick on every update.
    return m, m.tm.Tick()
}

func (m model) View() tea.View {
    return tea.NewView(m.tm.Render(80))
}

func main() {
    p := tea.NewProgram(model{tm: toast.NewToastManager()})
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

> The manager also discards expired toasts while rendering, so a toast disappears from
> the view as soon as it expires — even if the application never calls `Cleanup()`.
> The tick exists so the application gets a chance to redraw at the right moment.

## Toast Types

| Type | Icon | Color |
|------|------|-------|
| `Info` | `[i]` | Blue |
| `Success` | `[v]` | Green |
| `Warning` | `[!]` | Yellow |
| `Error` | `[x]` | Red |

## Behavior

- **Expiration** — a toast expires once its `Duration` has elapsed since `CreatedAt`.
  A `Duration` smaller than or equal to zero makes the toast *persistent*: it never
  expires on its own and must be removed with `Remove` or `Clear`.
- **`CreatedAt` is preserved** — `ToastManager.Add` keeps the timestamp provided by the
  caller, so an already expired toast is not revived by being added.
- **Display order** — the most recent toast is rendered at the top, older ones below.
  `MaxVisible` keeps the most recent toasts, so the visible window is always the newest
  ones.
- **Width** — `Render(width)` treats `width` as an upper bound: the message is truncated
  and the box is aligned to the right, so the result never exceeds the given width.
  A `width` of zero (or less) means "no limit".

## API

### Toast

```go
toast.NewToast(message string, duration time.Duration) Toast
toast.NewToast(...).WithType(ToastType) Toast
toast.NewToast(...).WithMessage(string) Toast
toast.NewToast(...).WithDuration(time.Duration) Toast
toast.NewToast(...).WithID(string) Toast
toast.IsExpired() bool
toast.Persistent() bool
toast.Remaining() time.Duration
toast.ExpiresAt() time.Time
toast.Render(width int) string
```

`WithType`, `WithMessage`, `WithDuration` and `WithID` are methods of `Toast` — they
return a modified copy of the toast.

### ToastManager

```go
tm := toast.NewToastManager()
tm.Add(t Toast)
tm.AddInfo(message string, duration time.Duration)
tm.AddSuccess(message string, duration time.Duration)
tm.AddWarning(message string, duration time.Duration)
tm.AddError(message string, duration time.Duration)
tm.Toasts() []Toast             // Active toasts, newest first (copy)
tm.Remove(id string) bool       // Remove a toast by ID
tm.Update(id string, fn func(toast.Toast) toast.Toast) bool
tm.SetMaxVisible(n int)         // Limit of visible toasts (0 or less = no limit)
tm.MaxVisible() int             // Current limit
tm.Cleanup()                    // Remove expired toasts
tm.Clear()                      // Remove all toasts
tm.Count() int
tm.Render(width int) string
tm.Tick() tea.Cmd               // Schedules toast.ExpiredMsg
tm.NextExpiry() time.Duration
```

### Bubble Tea messages

```go
toast.ExpiredMsg{IDs: []string{"toast-1"}}
```

## Migrating from the previous `MaxVisible` field

`ToastManager.MaxVisible` used to be a public field, which could be read under the mutex
while being written by the application without any lock — a data race. The field was
replaced by a pair of methods:

```go
// before
tm.MaxVisible = 3

// after
tm.SetMaxVisible(3)
limit := tm.MaxVisible()
```

## License

[MIT](LICENSE) - junhinhow
