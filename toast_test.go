package toast

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
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
	if toast.ID == "" {
		t.Error("identificador nao deveria estar vazio")
	}
	if toast.CreatedAt.IsZero() {
		t.Error("instante de criacao nao deveria estar zerado")
	}
}

func TestNewToastGeraIdentificadoresUnicos(t *testing.T) {
	primeiro := NewToast("a", time.Second)
	segundo := NewToast("b", time.Second)
	if primeiro.ID == segundo.ID {
		t.Errorf("identificadores deveriam ser unicos, ambos '%s'", primeiro.ID)
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

func TestWithID(t *testing.T) {
	toast := NewToast("Teste", time.Second).WithID("custom")
	if toast.ID != "custom" {
		t.Errorf("identificador esperado 'custom', obteve '%s'", toast.ID)
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

// TestIsExpiredDuracaoZeroNaoExpira garante que um toast sem duracao definida
// nao expira imediatamente e de forma silenciosa.
func TestIsExpiredDuracaoZeroNaoExpira(t *testing.T) {
	toast := NewToast("Persistente", 0)
	if toast.IsExpired() {
		t.Error("toast com duracao zero nao deveria estar expirado")
	}
	if !toast.Persistent() {
		t.Error("toast com duracao zero deveria ser persistente")
	}

	negativo := NewToast("Persistente", -time.Second)
	if negativo.IsExpired() {
		t.Error("toast com duracao negativa nao deveria estar expirado")
	}
	if !negativo.Persistent() {
		t.Error("toast com duracao negativa deveria ser persistente")
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

func TestRemainingDeToastPersistente(t *testing.T) {
	toast := NewToast("Teste", 0)
	if toast.Remaining() != 0 {
		t.Errorf("tempo restante de toast persistente deveria ser 0, obteve %v", toast.Remaining())
	}
}

func TestExpiresAt(t *testing.T) {
	toast := NewToast("Teste", time.Minute)
	expires := toast.ExpiresAt()
	if expires.IsZero() {
		t.Fatal("instante de expiracao nao deveria estar zerado")
	}
	if delta := expires.Sub(toast.CreatedAt); delta != time.Minute {
		t.Errorf("expiracao deveria ocorrer 1m apos a criacao, obteve %v", delta)
	}

	persistente := NewToast("Teste", 0)
	if !persistente.ExpiresAt().IsZero() {
		t.Error("toast persistente nao deveria ter instante de expiracao")
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

func TestToastTypeColor(t *testing.T) {
	for _, typ := range []ToastType{Info, Success, Warning, Error} {
		if typ.Color() == nil {
			t.Errorf("tipo %v nao deveria ter cor nula", typ)
		}
	}
}

func TestRender(t *testing.T) {
	toast := NewToast("Operacao concluida", time.Second).WithType(Success)
	output := toast.Render(80)
	if output == "" {
		t.Fatal("render nao deveria retornar string vazia")
	}
	if !strings.Contains(output, "Operacao concluida") {
		t.Errorf("render deveria conter a mensagem, obteve:\n%s", output)
	}
}

// TestRenderRespeitaLargura garante que o conteudo mais largo que a largura
// pedida seja truncado, evitando que a view ultrapasse o terminal.
func TestRenderRespeitaLargura(t *testing.T) {
	mensagem := strings.Repeat("mensagem muito longa ", 20)
	toast := NewToast(mensagem, time.Second)

	for _, largura := range []int{20, 40, 80} {
		output := toast.Render(largura)
		if got := lipgloss.Width(output); got > largura {
			t.Errorf("largura %d: esperava no maximo %d colunas, obteve %d", largura, largura, got)
		}
		if got := lipgloss.Width(output); got == 0 {
			t.Errorf("largura %d: render nao deveria ter largura zero", largura)
		}
	}
}

// TestRenderRespeitaLarguraComUnicode garante que a medicao usa a largura de
// exibicao e nao a contagem de bytes.
func TestRenderRespeitaLarguraComUnicode(t *testing.T) {
	toast := NewToast("日本語のテキストと絵文字😀", time.Second)
	output := toast.Render(24)
	if got := lipgloss.Width(output); got > 24 {
		t.Errorf("esperava no maximo 24 colunas, obteve %d", got)
	}
}

func TestRenderComMensagemMultilinhaRespeitaLargura(t *testing.T) {
	toast := NewToast("primeira linha\nsegunda linha bem longa", time.Second)
	output := toast.Render(30)
	for _, linha := range strings.Split(output, "\n") {
		if got := lipgloss.Width(linha); got > 30 {
			t.Errorf("linha com %d colunas excede o limite de 30: %q", got, linha)
		}
	}
}

func TestRenderSemLimiteQuandoLarguraZero(t *testing.T) {
	mensagem := strings.Repeat("x", 120)
	toast := NewToast(mensagem, time.Second)
	output := toast.Render(0)
	if got := lipgloss.Width(output); got <= 40 {
		t.Errorf("sem limite de largura esperava mais de 40 colunas, obteve %d", got)
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
		t.Fatal("render do manager nao deveria retornar string vazia")
	}
	if !strings.Contains(output, "Notificacao 1") || !strings.Contains(output, "Notificacao 2") {
		t.Errorf("render deveria conter as duas notificacoes, obteve:\n%s", output)
	}
}

func TestToastManagerRenderEmpty(t *testing.T) {
	tm := NewToastManager()
	output := tm.Render(80)
	if output != "" {
		t.Error("render do manager vazio deveria retornar string vazia")
	}
}

// TestToastManagerMaxVisible verifica que no maximo MaxVisible toasts sao
// renderizados.
func TestToastManagerMaxVisible(t *testing.T) {
	tm := NewToastManager()
	tm.SetMaxVisible(2)
	tm.AddInfo("primeiro", time.Minute)
	tm.AddInfo("segundo", time.Minute)
	tm.AddInfo("terceiro", time.Minute)

	output := tm.Render(80)
	if output == "" {
		t.Fatal("render nao deveria ser vazio com toasts ativos")
	}

	if !strings.Contains(output, "terceiro") || !strings.Contains(output, "segundo") {
		t.Errorf("render deveria conter os dois toasts mais novos, obteve:\n%s", output)
	}
	if strings.Contains(output, "primeiro") {
		t.Errorf("render nao deveria conter toasts alem do limite, obteve:\n%s", output)
	}

	// Cada toast ocupa tres linhas (borda superior, conteudo, borda inferior).
	linhas := strings.Split(output, "\n")
	if len(linhas) != 6 {
		t.Errorf("esperava 2 toasts (6 linhas), obteve %d linhas:\n%s", len(linhas), output)
	}
}

func TestToastManagerMaxVisibleSemLimite(t *testing.T) {
	tm := NewToastManager()
	tm.SetMaxVisible(0)
	for i := range 8 {
		tm.AddInfo(fmt.Sprintf("toast-%d", i), time.Minute)
	}

	output := tm.Render(80)
	for i := range 8 {
		if !strings.Contains(output, fmt.Sprintf("toast-%d", i)) {
			t.Errorf("sem limite esperava o toast-%d na renderizacao", i)
		}
	}
}

func TestMaxVisibleAccessor(t *testing.T) {
	tm := NewToastManager()
	if tm.MaxVisible() != defaultMaxVisible {
		t.Errorf("esperava %d por padrao, obteve %d", defaultMaxVisible, tm.MaxVisible())
	}
	tm.SetMaxVisible(7)
	if tm.MaxVisible() != 7 {
		t.Errorf("esperava 7 apos SetMaxVisible, obteve %d", tm.MaxVisible())
	}
}

// TestToastManagerOrdemMaisNovoPrimeiro verifica que os toasts mais recentes
// aparecem no topo da pilha, de forma coerente com o corte de MaxVisible.
func TestToastManagerOrdemMaisNovoPrimeiro(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("antigo", time.Minute)
	tm.AddInfo("meio", time.Minute)
	tm.AddInfo("novo", time.Minute)

	output := tm.Render(80)
	idxNovo := strings.Index(output, "novo")
	idxMeio := strings.Index(output, "meio")
	idxAntigo := strings.Index(output, "antigo")

	if idxNovo < 0 || idxMeio < 0 || idxAntigo < 0 {
		t.Fatalf("render deveria conter as tres notificacoes, obteve:\n%s", output)
	}
	if !(idxNovo < idxMeio && idxMeio < idxAntigo) {
		t.Errorf("esperava ordem novo, meio, antigo; obteve posicoes %d, %d, %d", idxNovo, idxMeio, idxAntigo)
	}
}

// TestToastManagerRenderDescartaExpirados garante que a renderizacao remove
// os toasts expirados, evitando crescimento ilimitado do gerenciador.
func TestToastManagerRenderDescartaExpirados(t *testing.T) {
	tm := NewToastManager()
	for i := range 20 {
		expirado := NewToast(fmt.Sprintf("expirado-%d", i), time.Second)
		expirado.CreatedAt = time.Now().Add(-time.Hour)
		tm.Add(expirado)
	}
	tm.AddInfo("ativo", time.Minute)

	tm.Render(80)

	if len(tm.toasts) != 1 {
		t.Errorf("esperava 1 toast retido apos a renderizacao, obteve %d", len(tm.toasts))
	}
	if tm.Count() != 1 {
		t.Errorf("esperava 1 toast ativo, obteve %d", tm.Count())
	}
}

// TestAddPreservaCreatedAt garante que um toast ja expirado nao e revivido
// pela adicao e que o instante informado pelo usuario e preservado.
func TestAddPreservaCreatedAt(t *testing.T) {
	tm := NewToastManager()

	toast := NewToast("ja expirado", time.Second)
	toast.CreatedAt = time.Now().Add(-time.Hour)
	tm.Add(toast)

	if tm.Count() != 0 {
		t.Errorf("toast expirado nao deveria ser contabilizado como ativo, obteve %d", tm.Count())
	}
	if output := tm.Render(80); output != "" {
		t.Errorf("toast expirado nao deveria ser renderizado, obteve:\n%s", output)
	}

	// O instante de criacao informado deve ser preservado.
	novo := NewToast("com instante proprio", time.Minute)
	novo.CreatedAt = time.Now().Add(-10 * time.Second)
	tm.Add(novo)

	ativos := tm.Toasts()
	if len(ativos) != 1 {
		t.Fatalf("esperava 1 toast ativo, obteve %d", len(ativos))
	}
	if !ativos[0].CreatedAt.Equal(novo.CreatedAt) {
		t.Errorf("instante de criacao deveria ser preservado: esperado %v, obteve %v", novo.CreatedAt, ativos[0].CreatedAt)
	}
}

func TestToastManagerToasts(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("primeiro", time.Minute)
	tm.AddInfo("segundo", time.Minute)

	ativos := tm.Toasts()
	if len(ativos) != 2 {
		t.Fatalf("esperava 2 toasts, obteve %d", len(ativos))
	}
	// A ordem é do mais novo para o mais antigo.
	if ativos[0].Message != "segundo" {
		t.Errorf("esperava 'segundo' primeiro, obteve '%s'", ativos[0].Message)
	}
	if ativos[1].Message != "primeiro" {
		t.Errorf("esperava 'primeiro' em segundo, obteve '%s'", ativos[1].Message)
	}
}

func TestToastManagerToastsNaoExpoeEstadoInterno(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("original", time.Minute)

	ativos := tm.Toasts()
	ativos[0].Message = "vazado"

	novos := tm.Toasts()
	if novos[0].Message != "original" {
		t.Errorf("alteracao na fatia retornada vazou para o estado interno, obteve '%s'", novos[0].Message)
	}
}

func TestToastManagerRemove(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("primeiro", time.Minute)
	tm.AddInfo("segundo", time.Minute)

	alvo := tm.Toasts()[0]
	if !tm.Remove(alvo.ID) {
		t.Fatal("remocao deveria informar sucesso")
	}
	if tm.Count() != 1 {
		t.Errorf("esperava 1 toast apos a remocao, obteve %d", tm.Count())
	}
	if tm.Remove("inexistente") {
		t.Error("remocao de id inexistente deveria informar falha")
	}
}

func TestToastManagerUpdate(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("antes", time.Minute)

	alvo := tm.Toasts()[0]
	if !tm.Update(alvo.ID, func(toast Toast) Toast {
		return toast.WithMessage("depois")
	}) {
		t.Fatal("update deveria informar sucesso")
	}

	ativos := tm.Toasts()
	if ativos[0].Message != "depois" {
		t.Errorf("mensagem atualizada incorreta, obteve '%s'", ativos[0].Message)
	}
	if ativos[0].ID != alvo.ID {
		t.Errorf("update deveria preservar o identificador, obteve '%s'", ativos[0].ID)
	}
	if !ativos[0].CreatedAt.Equal(alvo.CreatedAt) {
		t.Error("update deveria preservar o instante de criacao")
	}
}

func TestToastManagerUpdateInexistente(t *testing.T) {
	tm := NewToastManager()
	tm.AddInfo("unico", time.Minute)
	if tm.Update("inexistente", func(toast Toast) Toast { return toast }) {
		t.Error("update de id inexistente deveria informar falha")
	}
	if tm.Update("inexistente", nil) {
		t.Error("update com funcao nula deveria informar falha")
	}
}

func TestToastManagerConcorrencia(t *testing.T) {
	tm := NewToastManager()

	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for i := range 50 {
				toast := NewToast(fmt.Sprintf("toast-%d-%d", n, i), time.Minute)
				tm.Add(toast)
				tm.SetMaxVisible(1 + n%5)
				_ = tm.MaxVisible()
				_ = tm.Render(80)
				_ = tm.Count()
				_ = tm.Toasts()
				tm.Remove(toast.ID)
				tm.Cleanup()
			}
		}(worker)
	}
	wg.Wait()

	if got := tm.MaxVisible(); got < 0 {
		t.Errorf("valor de MaxVisible invalido: %d", got)
	}
}
