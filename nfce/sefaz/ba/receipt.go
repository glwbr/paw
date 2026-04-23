package ba

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func parseReceipt(content *goquery.Selection, result *nfce.Result) {
	if content.Length() == 0 {
		result.Warnings = append(result.Warnings, "nfe tab content not found")
		return
	}

	dataRow := content.Find("tr.col-6").First()
	if model, err := strconv.Atoi(htmlparse.LabeledField(dataRow, "Modelo")); err == nil {
		result.Receipt.Model = model
	}
	if series, err := strconv.Atoi(htmlparse.LabeledField(dataRow, "Série")); err == nil {
		result.Receipt.Series = series
	}
	if n, err := strconv.Atoi(htmlparse.LabeledField(dataRow, "Número")); err == nil {
		result.Receipt.Number = n
	}
	if rawDate := htmlparse.LabeledField(dataRow, "Data de Emissão"); rawDate != "" {
		if t, err := htmlparse.Date(rawDate); err == nil {
			result.Receipt.IssuedAt = t
		} else {
			result.Warnings = append(result.Warnings, "nfe: could not parse IssuedAt: "+err.Error())
		}
	}

	// Full store info comes from parseIssuer; only set if not already populated.
	emitente := htmlparse.SectionTable(content, "Emitente")
	if result.Receipt.Store.CNPJ == "" {
		result.Receipt.Store.CNPJ = htmlparse.DigitsOnly(htmlparse.LabeledField(emitente, "CNPJ"))
	}
	if result.Receipt.Store.Name == "" {
		result.Receipt.Store.Name = htmlparse.LabeledField(emitente, "Nome / Razão Social")
	}
	if result.Receipt.Store.StateReg == "" {
		result.Receipt.Store.StateReg = htmlparse.LabeledField(emitente, "Inscrição Estadual")
	}

	situacao := htmlparse.SectionTable(content, "Situação")
	protocolRows := situacao.Find("tr.col-3")
	if protocolRows.Length() >= 2 {
		dataRow := protocolRows.Eq(1)
		spans := dataRow.Find("span.linha")
		if spans.Length() >= 2 {
			result.Receipt.AuthProtocol = htmlparse.Text(spans.Eq(1).Text())
		}
		if spans.Length() >= 3 {
			rawDate := htmlparse.Text(spans.Eq(2).Text())
			rawDate = strings.Replace(rawDate, " às ", " ", 1)
			if t, err := htmlparse.Date(rawDate); err == nil {
				_ = t // AuthorizedAt not in Receipt struct; stored in AuthProtocol
			}
		}
	}
}
