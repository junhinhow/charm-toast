// Package toast fornece notificacoes toast para aplicacoes Bubble Tea.
//
// Suporta mensagens de status com auto-dismiss, com tipos Info, Success,
// Warning e Error, e um gerenciador para multiplas notificacoes.
package toast

import (
	"image/color"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
)

// ToastType representa o tipo da notificacao toast.
type ToastType int

const (
	// Info notificacao informativa.
	Info ToastType = iota
	// Success notificacao de sucesso.
	Success
	// Warning notificacao de aviso.
	Warning
	// Error notificacao de erro.
	Error
)

// String retorna a representacao textual do tipo.
func (t ToastType) String() string {
	switch t {
	case Info:
		return "info"
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// Icon retorna o icone associado ao tipo da notificacao.
func (t ToastType) Icon() string {
	switch t {
	case Info:
		return "i"
	case Success:
		return "v"
	case Warning:
		return "!"
	case Error:
		return "x"
	default:
		return "*"
	}
}

// Color retorna a cor associada ao tipo da notificacao.
func (t ToastType) Color() color.Color {
	switch t {
	case Info:
		return lipgloss.Color("#61AFEF")
	case Success:
		return lipgloss.Color("#98C379")
	case Warning:
		return lipgloss.Color("#E5C07B")
	case Error:
		return lipgloss.Color("#E06C75")
	default:
		return lipgloss.Color("#ABB2BF")
	}
}

// Toast representa uma notificacao toast individual.
type Toast struct {
	// Message texto da notificacao.
	Message string
	// Duration duracao antes do auto-dismiss.
	Duration time.Duration
	// Type tipo da notificacao (Info, Success, Warning, Error).
	Type ToastType
	// CreatedAt momento de criacao da notificacao.
	CreatedAt time.Time
}

// NewToast cria uma nova notificacao toast do tipo Info.
func NewToast(message string, duration time.Duration) Toast {
	return Toast{
		Message:   message,
		Duration:  duration,
		Type:      Info,
		CreatedAt: time.Now(),
	}
}

// WithType retorna uma copia do toast com o tipo alterado.
func (t Toast) WithType(typ ToastType) Toast {
	t.Type = typ
	return t
}

// WithMessage retorna uma copia do toast com a mensagem alterada.
func (t Toast) WithMessage(msg string) Toast {
	t.Message = msg
	return t
}

// WithDuration retorna uma copia do toast com a duracao alterada.
func (t Toast) WithDuration(d time.Duration) Toast {
	t.Duration = d
	return t
}

// IsExpired verifica se o toast ja expirou.
func (t Toast) IsExpired() bool {
	return time.Since(t.CreatedAt) >= t.Duration
}

// Remaining retorna o tempo restante antes do auto-dismiss.
func (t Toast) Remaining() time.Duration {
	remaining := t.Duration - time.Since(t.CreatedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Render renderiza o toast como string posicionada na largura fornecida.
func (t Toast) Render(width int) string {
	color := t.Type.Color()
	icon := t.Type.Icon()

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E1E")).
		Background(color).
		Bold(true).
		Padding(0, 1)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#2D2D2D")).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color)

	iconRendered := iconStyle.Render("[" + icon + "]")
	msgRendered := messageStyle.Render(t.Message)

	content := lipgloss.JoinHorizontal(lipgloss.Center, iconRendered, msgRendered)
	boxed := borderStyle.Render(content)

	// Alinha a direita se largura especificada
	if width > 0 {
		boxed = lipgloss.PlaceHorizontal(width, lipgloss.Right, boxed)
	}

	return boxed
}

// ToastManager gerencia multiplas notificacoes toast com auto-dismiss.
type ToastManager struct {
	mu     sync.Mutex
	toasts []Toast
	// MaxVisible numero maximo de toasts visiveis simultaneamente.
	MaxVisible int
}

// NewToastManager cria um novo gerenciador de toasts.
func NewToastManager() *ToastManager {
	return &ToastManager{
		MaxVisible: 5,
	}
}

// Add adiciona uma nova notificacao ao gerenciador.
func (tm *ToastManager) Add(t Toast) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	t.CreatedAt = time.Now()
	tm.toasts = append(tm.toasts, t)
}

// AddInfo adiciona uma notificacao do tipo Info.
func (tm *ToastManager) AddInfo(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration))
}

// AddSuccess adiciona uma notificacao do tipo Success.
func (tm *ToastManager) AddSuccess(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Success))
}

// AddWarning adiciona uma notificacao do tipo Warning.
func (tm *ToastManager) AddWarning(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Warning))
}

// AddError adiciona uma notificacao do tipo Error.
func (tm *ToastManager) AddError(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Error))
}

// Cleanup remove toasts expirados.
func (tm *ToastManager) Cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	active := make([]Toast, 0, len(tm.toasts))
	for _, t := range tm.toasts {
		if !t.IsExpired() {
			active = append(active, t)
		}
	}
	tm.toasts = active
}

// Count retorna o numero de toasts ativos (nao expirados).
func (tm *ToastManager) Count() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	count := 0
	for _, t := range tm.toasts {
		if !t.IsExpired() {
			count++
		}
	}
	return count
}

// Clear remove todas as notificacoes.
func (tm *ToastManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.toasts = nil
}

// Render renderiza todas as notificacoes ativas empilhadas verticalmente.
func (tm *ToastManager) Render(width int) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var lines []string
	visible := 0

	for i := len(tm.toasts) - 1; i >= 0; i-- {
		if visible >= tm.MaxVisible {
			break
		}
		t := tm.toasts[i]
		if !t.IsExpired() {
			lines = append([]string{t.Render(width)}, lines...)
			visible++
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}
