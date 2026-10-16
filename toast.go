// Package toast fornece notificações toast para aplicações Bubble Tea.
//
// O pacote oferece dois níveis de uso:
//
//   - Toast: uma notificação individual, com mensagem, tipo, duração e
//     instante de criação, renderizada como uma caixa alinhada à direita.
//   - ToastManager: um gerenciador thread-safe que mantém várias notificações,
//     descarta as expiradas, limita quantas são exibidas e oferece integração
//     com o Bubble Tea através de Tick e ExpiredMsg.
//
// Os toasts são descartados automaticamente (auto-dismiss) depois que a
// Duration decorre. Durações menores ou iguais a zero tornam a notificação
// persistente, ou seja, ela nunca expira sozinha.
package toast

import (
	"image/color"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/lipgloss/v2"
)

// ToastType representa o tipo da notificação toast.
type ToastType int

const (
	// Info notificação informativa.
	Info ToastType = iota
	// Success notificação de sucesso.
	Success
	// Warning notificação de aviso.
	Warning
	// Error notificação de erro.
	Error
)

// String retorna a representação textual do tipo.
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

// Icon retorna o ícone associado ao tipo da notificação.
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

// Color retorna a cor associada ao tipo da notificação.
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

// Toast representa uma notificação toast individual.
type Toast struct {
	// ID identificador único da notificação. É gerado automaticamente por
	// NewToast e por ToastManager.Add quando informado vazio.
	ID string
	// Message texto da notificação.
	Message string
	// Duration duração antes do auto-dismiss. Valores menores ou iguais a
	// zero tornam a notificação persistente, isto é, ela nunca expira sozinha.
	Duration time.Duration
	// Type tipo da notificação (Info, Success, Warning, Error).
	Type ToastType
	// CreatedAt momento de criação da notificação. É usado para calcular a
	// expiração e é preservado por ToastManager.Add quando informado.
	CreatedAt time.Time
}

// NewToast cria uma nova notificação toast do tipo Info.
func NewToast(message string, duration time.Duration) Toast {
	return Toast{
		ID:        nextToastID(),
		Message:   message,
		Duration:  duration,
		Type:      Info,
		CreatedAt: time.Now(),
	}
}

// WithType retorna uma cópia do toast com o tipo alterado.
func (t Toast) WithType(typ ToastType) Toast {
	t.Type = typ
	return t
}

// WithMessage retorna uma cópia do toast com a mensagem alterada.
func (t Toast) WithMessage(msg string) Toast {
	t.Message = msg
	return t
}

// WithDuration retorna uma cópia do toast com a duração alterada.
func (t Toast) WithDuration(d time.Duration) Toast {
	t.Duration = d
	return t
}

// WithID retorna uma cópia do toast com o identificador alterado.
func (t Toast) WithID(id string) Toast {
	t.ID = id
	return t
}

// IsExpired verifica se o toast já expirou.
//
// Durações menores ou iguais a zero correspondem a toasts persistentes, que
// nunca expiram sozinhos.
func (t Toast) IsExpired() bool {
	if t.Duration <= 0 {
		return false
	}
	if t.CreatedAt.IsZero() {
		return false
	}
	return time.Since(t.CreatedAt) >= t.Duration
}

// Persistent informa se o toast nunca expira sozinho.
func (t Toast) Persistent() bool {
	return t.Duration <= 0
}

// Remaining retorna o tempo restante antes do auto-dismiss.
//
// Devolve zero para toasts persistentes e para toasts já expirados.
func (t Toast) Remaining() time.Duration {
	if t.Duration <= 0 {
		return 0
	}
	remaining := t.Duration - time.Since(t.CreatedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ExpiresAt retorna o instante em que o toast expira.
//
// Para toasts persistentes devolve o valor zero de time.Time.
func (t Toast) ExpiresAt() time.Time {
	if t.Persistent() || t.CreatedAt.IsZero() {
		return time.Time{}
	}
	return t.CreatedAt.Add(t.Duration)
}

// Render renderiza o toast como uma caixa posicionada à direita na largura
// fornecida.
//
// A largura informada é um limite superior: conteúdos mais largos são
// truncados para caber, de modo que a view nunca ultrapassa a largura
// informada. Quando width é zero ou negativo, nenhum limite é aplicado.
func (t Toast) Render(width int) string {
	styles := styleFor(t.Type)

	icon := styles.icon.Render("[" + t.Type.Icon() + "]")

	// Largura disponível para o texto da mensagem.
	textWidth := 0
	if width > 0 {
		textWidth = width - toastBorderWidth - toastIconWidth
		if textWidth < 1 {
			// Não há espaço para o ícone ao lado da borda: usa apenas a
			// mensagem.
			icon = ""
			textWidth = width - toastBorderWidth
			if textWidth < 1 {
				// Largura insuficiente até para a borda: devolve somente o
				// texto truncado.
				return lipgloss.NewStyle().MaxWidth(width).Render(t.Message)
			}
		}
	}

	messageStyle := styles.message
	if textWidth > 0 {
		messageStyle = messageStyle.MaxWidth(textWidth)
	}

	content := messageStyle.Render(t.Message)
	if icon != "" {
		content = lipgloss.JoinHorizontal(lipgloss.Center, icon, content)
	}

	boxed := styles.border.Render(content)
	if width > 0 {
		boxed = lipgloss.PlaceHorizontal(width, lipgloss.Right, boxed)
	}
	return boxed
}

// ToastManager gerencia múltiplas notificações toast.
//
// Todas as operações públicas são protegidas por um mutex, portanto o
// gerenciador pode ser utilizado por múltiplas goroutines com segurança.
type ToastManager struct {
	mu     sync.Mutex
	toasts []Toast

	// maxVisible número máximo de toasts exibidos simultaneamente.
	// Valores menores ou iguais a zero significam "sem limite".
	maxVisible int
}

// NewToastManager cria um novo gerenciador de toasts.
//
// Por padrão até 5 toasts são exibidos simultaneamente; use SetMaxVisible
// para alterar esse limite.
func NewToastManager() *ToastManager {
	return &ToastManager{
		maxVisible: defaultMaxVisible,
	}
}

// SetMaxVisible define o número máximo de toasts exibidos simultaneamente.
//
// O valor é protegido pelo mesmo mutex das demais operações, evitando
// condições de corrida com Render. Valores menores ou iguais a zero removem o
// limite, exibindo todos os toasts ativos.
func (tm *ToastManager) SetMaxVisible(n int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.maxVisible = n
}

// MaxVisible retorna o número máximo de toasts exibidos simultaneamente.
//
// Zero ou menos significa que não há limite.
func (tm *ToastManager) MaxVisible() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.maxVisible
}

// Add adiciona uma nova notificação ao gerenciador.
//
// O CreatedAt informado pelo usuário é preservado; somente toasts sem
// CreatedAt recebem o instante atual. Dessa forma, um toast já expirado não
// é revivido pela adição.
func (tm *ToastManager) Add(t Toast) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.ID == "" {
		t.ID = nextToastID()
	}
	tm.toasts = append(tm.toasts, t)
}

// AddInfo adiciona uma notificação do tipo Info.
func (tm *ToastManager) AddInfo(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration))
}

// AddSuccess adiciona uma notificação do tipo Success.
func (tm *ToastManager) AddSuccess(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Success))
}

// AddWarning adiciona uma notificação do tipo Warning.
func (tm *ToastManager) AddWarning(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Warning))
}

// AddError adiciona uma notificação do tipo Error.
func (tm *ToastManager) AddError(message string, duration time.Duration) {
	tm.Add(NewToast(message, duration).WithType(Error))
}

// Cleanup remove os toasts expirados do gerenciador.
func (tm *ToastManager) Cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.pruneExpired()
}

// Count retorna o número de toasts ativos (não expirados).
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

// Toasts retorna uma cópia dos toasts ativos, do mais novo para o mais
// antigo.
//
// A cópia impede que o chamador altere o estado interno do gerenciador.
func (tm *ToastManager) Toasts() []Toast {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.pruneExpired()

	active := make([]Toast, 0, len(tm.toasts))
	for i := len(tm.toasts) - 1; i >= 0; i-- {
		active = append(active, tm.toasts[i])
	}
	return active
}

// Remove remove o toast com o identificador informado.
//
// Devolve true quando o toast existia e foi removido.
func (tm *ToastManager) Remove(id string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, t := range tm.toasts {
		if t.ID != id {
			continue
		}
		tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
		return true
	}
	return false
}

// Update altera o toast com o identificador informado usando a função update.
//
// A função recebe uma cópia do toast e devolve a versão atualizada. O
// identificador e o instante de criação são preservados caso a função os
// zere. Devolve true quando o toast existia e foi atualizado.
func (tm *ToastManager) Update(id string, update func(Toast) Toast) bool {
	if update == nil {
		return false
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, t := range tm.toasts {
		if t.ID != id {
			continue
		}
		updated := update(t)
		if updated.ID == "" {
			updated.ID = t.ID
		}
		if updated.CreatedAt.IsZero() {
			updated.CreatedAt = t.CreatedAt
		}
		tm.toasts[i] = updated
		return true
	}
	return false
}

// Clear remove todas as notificações.
func (tm *ToastManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.toasts = nil
}

// Render renderiza as notificações ativas empilhadas verticalmente.
//
// Os toasts expirados são descartados durante a renderização, evitando que o
// gerenciador cresça indefinidamente. Os mais recentes aparecem no topo e, no
// máximo, MaxVisible toasts são exibidos.
func (tm *ToastManager) Render(width int) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.pruneExpired()

	limit := tm.maxVisible
	if limit <= 0 {
		limit = len(tm.toasts)
	}

	// Percorre do mais novo para o mais antigo, reservando o espaço uma
	// única vez para evitar realocações sucessivas da fatia.
	lines := make([]string, 0, min(limit, len(tm.toasts)))
	for i := len(tm.toasts) - 1; i >= 0 && len(lines) < limit; i-- {
		lines = append(lines, tm.toasts[i].Render(width))
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

// pruneExpired remove os toasts expirados da lista interna.
//
// Deve ser chamado com o mutex bloqueado.
func (tm *ToastManager) pruneExpired() {
	if len(tm.toasts) == 0 {
		return
	}

	kept := tm.toasts[:0]
	for _, t := range tm.toasts {
		if !t.IsExpired() {
			kept = append(kept, t)
		}
	}

	// Libera as referências dos toasts descartados para não reter memória.
	for i := len(kept); i < len(tm.toasts); i++ {
		tm.toasts[i] = Toast{}
	}
	tm.toasts = kept
}

// toastCounter gera identificadores únicos para os toasts.
var toastCounter atomic.Uint64

// nextToastID devolve um identificador único para um novo toast.
func nextToastID() string {
	return "toast-" + strconv.FormatUint(toastCounter.Add(1), 10)
}
