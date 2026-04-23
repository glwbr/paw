package ba_test

import (
	"os"
	"testing"

	"github.com/glwbr/paw/nfce"
	_ "github.com/glwbr/paw/nfce/sefaz/ba"
)

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func TestAssembleBA_PrintPage(t *testing.T) {
	pages := []nfce.Page{
		{Name: "print.html", Content: mustRead(t, "../../testdata/ba/print.html")},
	}

	result := nfce.Parse(pages...)

	if result.HasErrors() {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Receipt == nil {
		t.Fatal("Receipt is nil")
	}

	r := result.Receipt

	if len(r.AccessKey) != 44 {
		t.Errorf("AccessKey len = %d, want 44", len(r.AccessKey))
	}
	if r.State != nfce.BA {
		t.Errorf("State = %q, want BA", r.State)
	}
	if r.Model != 65 {
		t.Errorf("Model = %d, want 65", r.Model)
	}
	if len(r.Store.CNPJ) != 14 {
		t.Errorf("Store.CNPJ = %q, want 14 digits", r.Store.CNPJ)
	}
	if r.Store.Name == "" {
		t.Error("Store.Name is empty")
	}
	if r.Store.Address.City == "" {
		t.Error("Store.Address.City is empty")
	}
	if r.Totals.TotalAmount.IsZero() {
		t.Error("Totals.TotalAmount is zero")
	}
	if len(r.Items) == 0 {
		t.Error("Items is empty")
	}
	for i, item := range r.Items {
		if item.Description == "" {
			t.Errorf("Items[%d].Description is empty", i)
		}
		if item.TotalPrice.IsZero() {
			t.Errorf("Items[%d].TotalPrice is zero", i)
		}
		if item.Taxes == nil {
			t.Errorf("Items[%d].Taxes is nil", i)
		}
	}
	if len(r.Payments) == 0 {
		t.Error("Payments is empty")
	}
	for i, p := range r.Payments {
		if p.Amount.IsZero() {
			t.Errorf("Payments[%d].Amount is zero", i)
		}
	}
}

func TestCanParse(t *testing.T) {
	page := mustRead(t, "../../testdata/ba/print.html")
	result := nfce.Parse(nfce.Page{Name: "print.html", Content: page})
	if result.Receipt == nil {
		t.Fatal("Receipt is nil")
	}
	if result.Receipt.State != nfce.BA {
		t.Errorf("State = %q, want BA", result.Receipt.State)
	}
}
