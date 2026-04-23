package nfce_test

import (
	"testing"

	"github.com/glwbr/paw/nfce"
	_ "github.com/glwbr/paw/nfce/sefaz/ba"
)

func TestParseStateUnsupported(t *testing.T) {
	result := nfce.ParseState(nfce.MG)
	if !result.HasErrors() {
		t.Fatal("expected error for unsupported state MG")
	}
}

func TestParseNoPagesReturnsError(t *testing.T) {
	result := nfce.Parse()
	if !result.HasErrors() {
		t.Fatal("expected error when no pages given")
	}
}

func TestParseAutoDetectsBA(t *testing.T) {
	// Use the issuer page which has a clean lbl_chave_acesso for BA (UF code 29).
	html := []byte(`<html><body>
		<span id="lbl_chave_acesso">2924 0112 3456 7800 0190 6500 1000 0034 0263 1101 9296</span>
		<span class="barra_cinza">SECRETARIA DA FAZENDA DO ESTADO DA BAHIA</span>
	</body></html>`)

	result := nfce.Parse(nfce.Page{Name: "test.html", Content: html})
	if result.Receipt == nil {
		t.Fatal("expected non-nil Receipt")
	}
	if result.Receipt.State != nfce.BA {
		t.Errorf("State = %q, want BA", result.Receipt.State)
	}
}
