# charm-toast

Toast notification component for [Bubble Tea](https://github.com/charmbracelet/bubbletea) — auto-dismissing status messages.

> **[Leia em Portugues (PT-BR)](README.pt-br.md)**

## Features

- **Four toast types**: Info, Success, Warning, Error
- **Auto-dismiss** with configurable duration
- **ToastManager** for managing multiple notifications
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss) — colored borders and icons
- Thread-safe toast management
- Right-aligned rendering for non-intrusive notifications

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

## Toast Types

| Type | Icon | Color |
|------|------|-------|
| `Info` | `[i]` | Blue |
| `Success` | `[v]` | Green |
| `Warning` | `[!]` | Yellow |
| `Error` | `[x]` | Red |

## API

### Toast

```go
toast.NewToast(message string, duration time.Duration) Toast
toast.WithType(ToastType) Toast
toast.WithMessage(string) Toast
toast.WithDuration(time.Duration) Toast
toast.IsExpired() bool
toast.Remaining() time.Duration
toast.Render(width int) string
```

### ToastManager

```go
tm := toast.NewToastManager()
tm.Add(t Toast)
tm.AddInfo(message, duration)
tm.AddSuccess(message, duration)
tm.AddWarning(message, duration)
tm.AddError(message, duration)
tm.Cleanup()    // Remove expired toasts
tm.Clear()      // Remove all toasts
tm.Count() int
tm.Render(width int) string
```

## License

[MIT](LICENSE) - junhinhow
