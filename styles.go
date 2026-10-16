package toast

import "charm.land/lipgloss/v2"

const (
	// defaultMaxVisible é o número padrão de toasts exibidos ao mesmo tempo.
	defaultMaxVisible = 5

	// toastBorderWidth corresponde às colunas ocupadas pela borda
	// (esquerda + direita).
	toastBorderWidth = 2
	// toastIconWidth corresponde às colunas ocupadas pelo ícone e pelo seu
	// padding horizontal.
	toastIconWidth = 5

	// Cores fixas usadas por todos os toasts.
	iconForeground    = "#1E1E1E"
	messageForeground = "#FAFAFA"
	messageBackground = "#2D2D2D"
)

// renderStyles reúne os estilos necessários para desenhar um toast.
type renderStyles struct {
	icon    lipgloss.Style
	message lipgloss.Style
	border  lipgloss.Style
}

// toastStyles guarda os estilos prontos de cada tipo suportado.
//
// Os estilos são construídos uma única vez, na inicialização do pacote, para
// que a renderização não precise instanciar novos lipgloss.Style a cada
// chamada.
var toastStyles = map[ToastType]renderStyles{
	Info:    newRenderStyles(Info),
	Success: newRenderStyles(Success),
	Warning: newRenderStyles(Warning),
	Error:   newRenderStyles(Error),
}

// unknownStyles são usados para tipos fora do conjunto suportado.
var unknownStyles = newRenderStyles(ToastType(-1))

// newRenderStyles cria os estilos de um tipo de toast.
func newRenderStyles(typ ToastType) renderStyles {
	accent := typ.Color()
	return renderStyles{
		icon: lipgloss.NewStyle().
			Foreground(lipgloss.Color(iconForeground)).
			Background(accent).
			Bold(true).
			Padding(0, 1),
		message: lipgloss.NewStyle().
			Foreground(lipgloss.Color(messageForeground)).
			Background(lipgloss.Color(messageBackground)).
			Padding(0, 1),
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent),
	}
}

// styleFor devolve os estilos cacheados do tipo informado.
func styleFor(typ ToastType) renderStyles {
	if styles, ok := toastStyles[typ]; ok {
		return styles
	}
	return unknownStyles
}
