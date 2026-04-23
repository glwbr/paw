package ba

import (
	"os"
	"testing"
)

func TestParseFormState_validHTML(t *testing.T) {
	t.Parallel()

	html, err := os.ReadFile("testdata/form_state.html")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	fs, err := parseFormState(html)
	if err != nil {
		t.Fatalf("parseFormState() error: %v", err)
	}

	if !fs.valid() {
		t.Fatal("expected valid form state")
	}
	if fs.viewState != "dGVzdHZpZXdzdGF0ZQ==" {
		t.Errorf("viewState = %q, want %q", fs.viewState, "dGVzdHZpZXdzdGF0ZQ==")
	}
	if fs.viewStateGenerator != "0760F948" {
		t.Errorf("viewStateGenerator = %q, want %q", fs.viewStateGenerator, "0760F948")
	}
	if fs.eventValidation != "dGVzdHZhbGlkYXRpb24=" {
		t.Errorf("eventValidation = %q, want %q", fs.eventValidation, "dGVzdHZhbGlkYXRpb24=")
	}
}

func TestParseFormState_missingFields(t *testing.T) {
	t.Parallel()

	html := []byte(`<html><body><form></form></body></html>`)
	fs, err := parseFormState(html)
	if err != nil {
		t.Fatalf("parseFormState() error: %v", err)
	}

	if fs.valid() {
		t.Error("expected invalid form state when ViewState is missing")
	}
}

func TestParseFormState_malformedHTML(t *testing.T) {
	t.Parallel()

	html := []byte(`<html><body><input name="__VIEWSTATE" value="abc">`)
	fs, err := parseFormState(html)
	if err != nil {
		t.Fatalf("parseFormState() error: %v", err)
	}

	if fs.viewState != "abc" {
		t.Errorf("viewState = %q, want %q", fs.viewState, "abc")
	}
}

func TestBuildForm_mergesState(t *testing.T) {
	t.Parallel()

	fs := &formState{
		viewState:          "vs",
		viewStateGenerator: "vsg",
		eventValidation:    "ev",
	}

	extra := map[string]string{
		"txt_chave_acesso": "12345",
		"btn_consulta":     "Consultar",
	}

	vals := buildForm(fs, extra)

	if got := vals.Get("__VIEWSTATE"); got != "vs" {
		t.Errorf("__VIEWSTATE = %q, want %q", got, "vs")
	}
	if got := vals.Get("txt_chave_acesso"); got != "12345" {
		t.Errorf("txt_chave_acesso = %q, want %q", got, "12345")
	}
	if got := vals.Get("btn_consulta"); got != "Consultar" {
		t.Errorf("btn_consulta = %q, want %q", got, "Consultar")
	}
}

func TestBuildForm_nilState(t *testing.T) {
	t.Parallel()

	extra := map[string]string{"key": "value"}
	vals := buildForm(nil, extra)

	if got := vals.Get("key"); got != "value" {
		t.Errorf("key = %q, want %q", got, "value")
	}
	if got := vals.Get("__VIEWSTATE"); got != "" {
		t.Errorf("__VIEWSTATE should be empty, got %q", got)
	}
}
