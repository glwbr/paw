package ba

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func parseTotals(content *goquery.Selection, result *nfce.Result) {
	if content.Length() == 0 {
		result.Warnings = append(result.Warnings, "totais tab content not found")
		return
	}

	if v, err := nfce.ParseMoney(htmlparse.LabeledField(content, "Valor Total dos Produtos")); err == nil {
		result.Receipt.Totals.Subtotal = v
	}
	if v, err := nfce.ParseMoney(htmlparse.LabeledField(content, "Valor Total dos Descontos")); err == nil {
		result.Receipt.Totals.DiscountAmount = v
	}
	if v, err := nfce.ParseMoney(htmlparse.LabeledField(content, "Valor Total da NFe")); err == nil {
		result.Receipt.Totals.TotalAmount = v
	}
	if v, err := nfce.ParseMoney(htmlparse.LabeledField(content, "Valor Aproximado dos Tributos")); err == nil {
		result.Receipt.Totals.ApproxTaxTotal = v
	}
}
