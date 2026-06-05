# charm-toast

Componente de notificacao toast para [Bubble Tea](https://github.com/charmbracelet/bubbletea) — mensagens de status com auto-dismiss.

> **[Read in English](README.md)**

## Funcionalidades

- **Quatro tipos de toast**: Info, Success, Warning, Error
- **Auto-dismiss** com duracao configuravel
- **ToastManager** para gerenciar multiplas notificacoes
- Estilizado com [Lip Gloss](https://github.com/charmbracelet/lipgloss) — bordas coloridas e icones
- Gerenciamento thread-safe
- Renderizacao alinhada a direita para notificacoes nao-intrusivas

## Instalacao

```bash
go get github.com/junhinhow/charm-toast@latest
```

## Uso

```go
package main

import (
    "fmt"
    "time"
    "github.com/junhinhow/charm-toast"
)

func main() {
    // Toast individual
    t := toast.NewToast("Arquivo salvo com sucesso", 3*time.Second).WithType(toast.Success)
    fmt.Println(t.Render(80))

    // Gerenciador para multiplas notificacoes
    tm := toast.NewToastManager()
    tm.AddInfo("Carregando...", 5*time.Second)
    tm.AddSuccess("Concluido!", 3*time.Second)
    tm.AddError("Falha na conexao", 10*time.Second)
    fmt.Println(tm.Render(80))
}
```

## Tipos de Toast

| Tipo | Icone | Cor |
|------|-------|-----|
| `Info` | `[i]` | Azul |
| `Success` | `[v]` | Verde |
| `Warning` | `[!]` | Amarelo |
| `Error` | `[x]` | Vermelho |

## API

### Toast

```go
toast.NewToast(mensagem string, duracao time.Duration) Toast
toast.WithType(ToastType) Toast
toast.WithMessage(string) Toast
toast.WithDuration(time.Duration) Toast
toast.IsExpired() bool
toast.Remaining() time.Duration
toast.Render(largura int) string
```

### ToastManager

```go
tm := toast.NewToastManager()
tm.Add(t Toast)
tm.AddInfo(mensagem, duracao)
tm.AddSuccess(mensagem, duracao)
tm.AddWarning(mensagem, duracao)
tm.AddError(mensagem, duracao)
tm.Cleanup()    // Remove toasts expirados
tm.Clear()      // Remove todos os toasts
tm.Count() int
tm.Render(largura int) string
```

## Licenca

[MIT](LICENSE) - junhinhow
