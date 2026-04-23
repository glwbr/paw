package parse_test

import (
	"testing"

	"github.com/glwbr/paw/nfce/internal/parse"
)

const sampleHTML = `<html><body>
<span id="lbl_chave_acesso" class="labelConteudo">2924 0112 3456 7800 0190 6500 1000 0034 0263 1101 9296</span>
<div id="Emitente">
  <table><tr><td><label>CNPJ</label><span class="linha">12.345.678/0001-90</span></td></tr></table>
  <table><tr>
    <td class="table-titulo-aba-interna">Dados</td>
  </tr></table>
  <table><tr><td><label>Nome</label><span class="linha">Loja Teste</span></td></tr></table>
</div>
</body></html>`

func TestHeaderField(t *testing.T) {
	doc, err := parse.Doc([]byte(sampleHTML))
	if err != nil {
		t.Fatal(err)
	}
	got := parse.HeaderField(doc, "lbl_chave_acesso")
	want := "2924 0112 3456 7800 0190 6500 1000 0034 0263 1101 9296"
	if got != want {
		t.Errorf("HeaderField = %q, want %q", got, want)
	}
}

func TestLabeledField(t *testing.T) {
	doc, err := parse.Doc([]byte(sampleHTML))
	if err != nil {
		t.Fatal(err)
	}
	content := doc.Find("#Emitente")
	got := parse.LabeledField(content, "CNPJ")
	if got != "12.345.678/0001-90" {
		t.Errorf("LabeledField(CNPJ) = %q, want %q", got, "12.345.678/0001-90")
	}
}

func TestSectionTable(t *testing.T) {
	doc, err := parse.Doc([]byte(sampleHTML))
	if err != nil {
		t.Fatal(err)
	}
	content := doc.Find("#Emitente")
	tbl := parse.SectionTable(content, "Dados")
	if tbl.Length() == 0 {
		t.Error("SectionTable returned empty selection for existing section")
	}
	got := parse.LabeledField(tbl, "Nome")
	if got != "Loja Teste" {
		t.Errorf("LabeledField after SectionTable = %q, want %q", got, "Loja Teste")
	}
}

func TestCodePrefix(t *testing.T) {
	tests := []struct{ in, want string }{
		{"00 - Tributada integralmente", "00"},
		{"3 - Cartão de Crédito", "3"},
		{"0 - Nacional", "0"},
		{"noprefx", "noprefx"},
	}
	for _, tc := range tests {
		got := parse.CodePrefix(tc.in)
		if got != tc.want {
			t.Errorf("CodePrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
