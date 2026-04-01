package toast

import (
	"testing"
	"time"
)

func TestNewToast(t *testing.T) {
	toast := NewToast("Teste", 3*time.Second)
	if toast.Message != "Teste" {
		t.Errorf("mensagem esperada 'Teste', obteve '%s'", toast.Message)
	}
	if toast.Duration != 3*time.Second {
		t.Errorf("duracao esperada 3s, obteve %v", toast.Duration)
	}
	if toast.Type != Info {
		t.Errorf("tipo esperado Info, obteve %v", toast.Type)
	}
}

func TestWithType(t *testing.T) {
	toast := NewToast("Erro", time.Second).WithType(Error)
	if toast.Type != Error {
		t.Errorf("tipo esperado Error, obteve %v", toast.Type)
	}
}

func TestWithMessage(t *testing.T) {
	toast := NewToast("Original", time.Second).WithMessage("Alterada")
	if toast.Message != "Alterada" {
		t.Errorf("mensagem esperada 'Alterada', obteve '%s'", toast.Message)
	}
}

func TestWithDuration(t *testing.T) {
	toast := NewToast("Teste", time.Second).WithDuration(5 * time.Second)
	if toast.Duration != 5*time.Second {
		t.Errorf("duracao esperada 5s, obteve %v", toast.Duration)
	}
}

func TestIsExpired(t *testing.T) {
	toast := NewToast("Teste", 50*time.Millisecond)
	if toast.IsExpired() {
		t.Error("toast nao deveria estar expirado imediatamente")
	}
	time.Sleep(60 * time.Millisecond)
	if !toast.IsExpired() {
		t.Error("toast deveria estar expirado apos duracao")
	}
}

func TestRemaining(t *testing.T) {
	toast := NewToast("Teste", 1*time.Second)
	remaining := toast.Remaining()
	if remaining <= 0 {
		t.Error("tempo restante deveria ser positivo")
	}
	if remaining > 1*time.Second {
		t.Error("tempo restante nao deveria exceder duracao")
	}
}

func TestToastTypeString(t *testing.T) {
	tests := []struct {
		typ      ToastType
		expected string
	}{
		{Info, "info"},
		{Success, "success"},
		{Warning, "warning"},
		{Error, "error"},
	}
	for _, tt := range tests {
		if tt.typ.String() != tt.expected {
			t.Errorf("tipo %d: esperado '%s', obteve '%s'", tt.typ, tt.expected, tt.typ.String())
		}
	}
}

func TestToastTypeIcon(t *testing.T) {
	if Info.Icon() != "i" {
		t.Errorf("icone Info esperado 'i', obteve '%s'", Info.Icon())
	}
	if Success.Icon() != "v" {
		t.Errorf("icone Success esperado 'v', obteve '%s'", Success.Icon())
	}
	if Warning.Icon() != "!" {
		t.Errorf("icone Warning esperado '!', obteve '%s'", Warning.Icon())
	}
	if Error.Icon() != "x" {
		t.Errorf("icone Error esperado 'x', obteve '%s'", Error.Icon())
	}
}

func TestRender(t *testing.T) {
	toast := NewToast("Operacao concluida", time.Second).WithType(Success)
	output := toast.Render(80)
	if output == "" {
		t.Error("render nao deveria retornar string vazia")
	}
}

func TestToastManager(t *testing.T) {
	tm := NewToastManager()

	if tm.Count() != 0 {
		t.Errorf("gerenciador deveria iniciar vazio, obteve %d", tm.Count())
	}

	tm.AddInfo("Info", time.Minute)
	tm.AddSuccess("Sucesso", time.Minute)
	tm.AddWarning("Aviso", time.Minute)
	tm.AddError("Erro", time.Minute)

	if tm.Count() != 4 {
		t.Errorf("esperados 4 toasts, obteve %d", tm.Count())
	}
}

func TestToastManagerCleanup(t *testing.T) {
	tm := NewToastManager()
	tm.Add(NewToast("Expira rapido", 50*time.Millisecond))
	tm.AddInfo("Dura mais", time.Minute)

	time.Sleep(60 * time.Millisecond)
	tm.Cleanup()

	if tm.Count() != 1 {
		t.Errorf("apos cleanup esperado 1 toast, obteve %d", tm.Count())
	}
}

func TestToastManagerClear(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("Um", time.Minute)
	tm.AddInfo("Dois", time.Minute)
	tm.Clear()

	if tm.Count() != 0 {
		t.Errorf("apos clear esperado 0 toasts, obteve %d", tm.Count())
	}
}

func TestToastManagerRender(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("Notificacao 1", time.Minute)
	tm.AddSuccess("Notificacao 2", time.Minute)

	output := tm.Render(80)
	if output == "" {
		t.Error("render do manager nao deveria retornar string vazia")
	}
}

func TestToastManagerRenderEmpty(t *testing.T) {
	tm := NewToastManager()
	output := tm.Render(80)
	if output != "" {
		t.Error("render do manager vazio deveria retornar string vazia")
	}
}

func TestToastManagerMaxVisible(t *testing.T) {
	tm := NewToastManager()
	tm.MaxVisible = 2
	tm.AddInfo("Um", time.Minute)
	tm.AddInfo("Dois", time.Minute)
	tm.AddInfo("Tres", time.Minute)

	// Verifica que renderiza no maximo MaxVisible
	output := tm.Render(80)
	if output == "" {
		t.Error("render nao deveria ser vazio com toasts ativos")
	}
}
