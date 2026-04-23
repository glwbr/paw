package ba

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func parseItems(content *goquery.Selection, result *nfce.Result) {
	if content.Length() == 0 {
		result.Warnings = append(result.Warnings, "produtos tab content not found")
		return
	}

	htmlparse.EachToggleGroup(content, func(toggle, detail *goquery.Selection) {
		item, err := buildItem(toggle, detail)
		if err != nil {
			result.Errors = append(result.Errors, "items: "+err.Error())
			return
		}
		result.Receipt.Items = append(result.Receipt.Items, item)
	})
}

func buildItem(toggle, detail *goquery.Selection) (nfce.Item, error) {
	seq, _ := strconv.Atoi(htmlparse.LabeledField(toggle, "Número"))
	desc := htmlparse.LabeledField(toggle, "Descrição")
	unit := htmlparse.LabeledField(toggle, "Unidade Comercial")
	totalRaw := htmlparse.LabeledField(toggle, "Valor (R$)")
	qtyRaw := htmlparse.LabeledField(toggle, "Qtd.")

	total, err := nfce.ParseMoney(totalRaw)
	if err != nil {
		return nfce.Item{}, fmt.Errorf("item %d total: %w", seq, err)
	}
	qty, err := nfce.ParseQuantity(qtyRaw)
	if err != nil {
		return nfce.Item{}, fmt.Errorf("item %d qty: %w", seq, err)
	}

	productCode := htmlparse.LabeledField(detail, "Código do Produto")
	ncm := htmlparse.LabeledField(detail, "Código NCM")
	cest := htmlparse.LabeledField(detail, "Código CEST")
	cfop := htmlparse.LabeledField(detail, "CFOP")
	gtinComm := htmlparse.LabeledField(detail, "Código EAN Comercial")
	unitPriceRaw := htmlparse.LabeledField(detail, "Valor unitário de comercialização")
	approxTaxRaw := htmlparse.LabeledField(detail, "Valor Aproximado dos Tributos")

	unitPrice, _ := nfce.ParseMoney(unitPriceRaw)
	approxTax, _ := nfce.ParseMoney(approxTaxRaw)

	if strings.EqualFold(gtinComm, "SEM GTIN") {
		gtinComm = ""
	}

	item := nfce.Item{
		Sequence:        seq,
		ProductCode:     productCode,
		Description:     desc,
		Quantity:        qty,
		Unit:            unit,
		NormalizedUnit:  nfce.NormalizeUnit(unit),
		UnitPrice:       unitPrice,
		TotalPrice:      total,
		NCM:             ncm,
		CEST:            cest,
		CFOP:            cfop,
		GTINCommercial:  gtinComm,
		ApproxTaxAmount: approxTax,
		Taxes:           parseTaxes(detail),
	}
	return item, nil
}

func parseTaxes(detail *goquery.Selection) *nfce.ItemTaxes {
	taxes := &nfce.ItemTaxes{}

	icmsTable := htmlparse.SectionTable(detail, "ICMS Normal")
	if icmsTable.Length() > 0 {
		taxes.ICMS = &nfce.ICMS{}
		if base, err := nfce.ParseMoney(htmlparse.LabeledField(icmsTable, "Base de Cálculo do ICMS Normal")); err == nil {
			taxes.ICMS.BaseAmount = base
		}
		if rate, err := nfce.ParseRate(htmlparse.LabeledField(icmsTable, "Alíquota do ICMS Normal")); err == nil {
			taxes.ICMS.Rate = rate
		}
		if amount, err := nfce.ParseMoney(htmlparse.LabeledField(icmsTable, "Valor do ICMS Normal")); err == nil {
			taxes.ICMS.Amount = amount
		}
	}

	pisTable := htmlparse.SectionTable(detail, "PIS")
	if pisTable.Length() > 0 {
		taxes.PIS = parsePISCOFINS(pisTable)
	}

	cofinsTable := htmlparse.SectionTable(detail, "COFINS")
	if cofinsTable.Length() > 0 {
		raw := parsePISCOFINS(cofinsTable)
		taxes.COFINS = &nfce.COFINS{
			Rate:       raw.Rate,
			BaseAmount: raw.BaseAmount,
			Amount:     raw.Amount,
		}
	}

	return taxes
}

func parsePISCOFINS(sel *goquery.Selection) *nfce.PIS {
	p := &nfce.PIS{}
	if base, err := nfce.ParseMoney(htmlparse.LabeledField(sel, "Base de Cálculo")); err == nil {
		p.BaseAmount = base
	}
	if rate, err := nfce.ParseRate(htmlparse.LabeledField(sel, "Alíquota")); err == nil {
		p.Rate = rate
	}
	if amount, err := nfce.ParseMoney(htmlparse.LabeledField(sel, "Valor")); err == nil {
		p.Amount = amount
	}
	return p
}
