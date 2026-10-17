# charm-toast

Componente de notificação toast para [Bubble Tea](https://github.com/charmbracelet/bubbletea) — mensagens de status com descarte automático.

> **[Read in English](README.md)**

## Funcionalidades

- **Quatro tipos de toast**: Info, Success, Warning, Error
- **Descarte automático** com duração configurável (os toasts expirados são removidos durante a renderização)
- **Integração com o Bubble Tea** via `ToastManager.Tick()` / `toast.ExpiredMsg`
- **ToastManager** para gerenciar múltiplas notificações
- Estilizado com [Lip Gloss](https://github.com/charmbracelet/lipgloss) — bordas coloridas e ícones
- Gerenciamento thread-safe (todas as operações do gerenciador são protegidas por um mutex)
- Renderização alinhada à direita para notificações não-intrusivas
- Renderização limitada por largura — conteúdos mais largos que a largura pedida são truncados, de modo que a view nunca ultrapassa o terminal

## Instalação

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

    // Gerenciador para múltiplas notificações
    tm := toast.NewToastManager()
    tm.AddInfo("Carregando...", 5*time.Second)
    tm.AddSuccess("Concluído!", 3*time.Second)
    tm.AddError("Falha na conexão", 10*time.Second)
    fmt.Println(tm.Render(80))
}
```

## Integração com o Bubble Tea

`ToastManager.Tick()` agenda um `tea.Cmd` que emite `toast.ExpiredMsg` quando o próximo
toast expira. Arme o tick em `Init()` e rearme-o a cada mensagem, para que os toasts
adicionados depois também sejam considerados:

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
            m.tm.AddInfo("Processando...", 3*time.Second)
        case "q":
            return m, tea.Quit
        }
    case toast.ExpiredMsg:
        // Descarta os toasts que não são mais válidos.
        m.tm.Cleanup()
    }
    // Rearma o tick de expiração a cada atualização.
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

> O gerenciador também descarta os toasts expirados durante a renderização, portanto um
> toast sai da view assim que expira — mesmo que a aplicação nunca chame `Cleanup()`.
> O tick existe para dar à aplicação a chance de redesenhar no momento certo.

## Tipos de Toast

| Tipo | Ícone | Cor |
|------|-------|-----|
| `Info` | `[i]` | Azul |
| `Success` | `[v]` | Verde |
| `Warning` | `[!]` | Amarelo |
| `Error` | `[x]` | Vermelho |

## Comportamento

- **Expiração** — um toast expira quando a `Duration` decorre desde o `CreatedAt`. Uma
  `Duration` menor ou igual a zero torna o toast *persistente*: ele nunca expira sozinho
  e precisa ser removido com `Remove` ou `Clear`.
- **O `CreatedAt` é preservado** — `ToastManager.Add` mantém o instante informado pelo
  chamador, de modo que um toast já expirado não é revivido ao ser adicionado.
- **Ordem de exibição** — o toast mais recente é renderizado no topo e os mais antigos
  abaixo. `MaxVisible` mantém os toasts mais recentes, portanto a janela visível é sempre
  a dos mais novos.
- **Largura** — `Render(width)` trata `width` como um limite superior: a mensagem é
  truncada e a caixa é alinhada à direita, de modo que o resultado nunca ultrapassa a
  largura informada. Uma `width` igual a zero (ou menor) significa "sem limite".

## API

### Toast

```go
toast.NewToast(mensagem string, duracao time.Duration) Toast
toast.NewToast(...).WithType(ToastType) Toast
toast.NewToast(...).WithMessage(string) Toast
toast.NewToast(...).WithDuration(time.Duration) Toast
toast.NewToast(...).WithID(string) Toast
toast.IsExpired() bool
toast.Persistent() bool
toast.Remaining() time.Duration
toast.ExpiresAt() time.Time
toast.Render(largura int) string
```

`WithType`, `WithMessage`, `WithDuration` e `WithID` são **métodos** de `Toast` — devolvem
uma cópia modificada do toast.

### ToastManager

```go
tm := toast.NewToastManager()
tm.Add(t Toast)
tm.AddInfo(mensagem string, duracao time.Duration)
tm.AddSuccess(mensagem string, duracao time.Duration)
tm.AddWarning(mensagem string, duracao time.Duration)
tm.AddError(mensagem string, duracao time.Duration)
tm.Toasts() []Toast             // Toasts ativos, do mais novo para o mais antigo (cópia)
tm.Remove(id string) bool       // Remove um toast pelo identificador
tm.Update(id string, fn func(toast.Toast) toast.Toast) bool
tm.SetMaxVisible(n int)         // Limite de toasts visíveis (0 ou menos = sem limite)
tm.MaxVisible() int             // Limite atual
tm.Cleanup()                    // Remove os toasts expirados
tm.Clear()                      // Remove todos os toasts
tm.Count() int
tm.Render(largura int) string
tm.Tick() tea.Cmd               // Agenda a mensagem toast.ExpiredMsg
tm.NextExpiry() time.Duration
```

### Mensagens do Bubble Tea

```go
toast.ExpiredMsg{IDs: []string{"toast-1"}}
```

## Migração do antigo campo `MaxVisible`

`ToastManager.MaxVisible` era um campo público, que podia ser lido sob o mutex enquanto
era escrito pela aplicação sem nenhuma trava — uma condição de corrida. O campo foi
substituído por um par de métodos:

```go
// antes
tm.MaxVisible = 3

// depois
tm.SetMaxVisible(3)
limite := tm.MaxVisible()
```

## Licença

[MIT](LICENSE) - junhinhow
