package toast

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// expiryGrace é a margem adicionada ao agendamento da expiração para
// garantir que o toast já esteja vencido quando a mensagem chega ao
// aplicativo.
const expiryGrace = time.Millisecond

// ExpiredMsg é entregue pelo Bubble Tea quando chega o instante de expiração
// do próximo toast.
//
// Ao receber a mensagem, o aplicativo deve chamar ToastManager.Cleanup para
// descartar os toasts expirados e renderizar a view novamente.
type ExpiredMsg struct {
	// IDs contém os identificadores dos toasts que já estavam expirados no
	// momento do disparo. A lista pode estar vazia quando o aplicativo já
	// removeu os toasts expirados antes de tratar a mensagem.
	IDs []string
}

// Tick retorna um tea.Cmd que dispara ExpiredMsg quando o próximo toast
// expira.
//
// Devolve nil quando não há toasts com expiração programada (por exemplo,
// apenas toasts persistentes). O comando deve ser rearmado a cada ExpiredMsg
// e sempre que um novo toast for adicionado, conforme o exemplo de integração
// do README.
func (tm *ToastManager) Tick() tea.Cmd {
	delay := tm.NextExpiry()
	if delay <= 0 {
		return nil
	}

	return tea.Tick(delay+expiryGrace, func(time.Time) tea.Msg {
		return ExpiredMsg{IDs: tm.expiredIDs()}
	})
}

// NextExpiry retorna o tempo restante até a expiração do próximo toast.
//
// Devolve zero quando não há toasts que expiram sozinhos.
func (tm *ToastManager) NextExpiry() time.Duration {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var next time.Duration
	for _, t := range tm.toasts {
		if t.Persistent() {
			continue
		}
		remaining := t.Remaining()
		if remaining <= 0 {
			continue
		}
		if next == 0 || remaining < next {
			next = remaining
		}
	}
	return next
}

// expiredIDs devolve os identificadores dos toasts já expirados.
//
// Deve ser chamado com o mutex bloqueado.
func (tm *ToastManager) expiredIDs() []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var ids []string
	for _, t := range tm.toasts {
		if t.IsExpired() {
			ids = append(ids, t.ID)
		}
	}
	return ids
}
