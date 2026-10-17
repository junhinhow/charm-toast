package toast_test

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/junhinhow/charm-toast"
)

// Exemplo de uso isolado, sem Bubble Tea.
func ExampleNewToast() {
	t := toast.NewToast("Arquivo salvo com sucesso", 3*time.Second).WithType(toast.Success)
	fmt.Println(t.Render(40))
}

// Exemplo do gerenciador com vários toasts.
func ExampleToastManager() {
	tm := toast.NewToastManager()
	tm.AddInfo("Carregando...", 5*time.Second)
	tm.AddSuccess("Concluído!", 3*time.Second)
	tm.AddError("Falha na conexão", 10*time.Second)

	fmt.Println(tm.Render(60))
}

// Exemplo do modelo Bubble Tea mostrado no README.
type exemploModel struct {
	tm *toast.ToastManager
}

func (m exemploModel) Init() tea.Cmd {
	return m.tm.Tick()
}

func (m exemploModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.tm.AddInfo("Processando...", 3*time.Second)
		case "q":
			return m, tea.Quit
		}
	case toast.ExpiredMsg:
		m.tm.Cleanup()
	}
	return m, m.tm.Tick()
}

func (m exemploModel) View() tea.View {
	return tea.NewView(m.tm.Render(80))
}

func ExampleToastManager_Tick() {
	p := tea.NewProgram(exemploModel{tm: toast.NewToastManager()})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
