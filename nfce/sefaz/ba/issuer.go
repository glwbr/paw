package ba

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func parseIssuer(content *goquery.Selection, result *nfce.Result) {
	if content.Length() == 0 {
		result.Warnings = append(result.Warnings, "emitente tab content not found")
		return
	}

	result.Receipt.Store.Name = htmlparse.LabeledField(content, "Nome / Razão Social")
	result.Receipt.Store.TradeName = htmlparse.LabeledField(content, "Nome Fantasia")
	result.Receipt.Store.CNPJ = htmlparse.DigitsOnly(htmlparse.LabeledField(content, "CNPJ"))
	result.Receipt.Store.StateReg = htmlparse.LabeledField(content, "Inscrição Estadual")

	street, number := splitAddress(htmlparse.LabeledField(content, "Endereço"))
	result.Receipt.Store.Address = nfce.Address{
		Street:   street,
		Number:   number,
		District: htmlparse.LabeledField(content, "Bairro / Distrito"),
		ZipCode:  htmlparse.DigitsOnly(htmlparse.LabeledField(content, "CEP")),
		City:     cityName(htmlparse.LabeledField(content, "Município")),
		State:    htmlparse.LabeledField(content, "UF"),
	}
}

func splitAddress(raw string) (street, number string) {
	idx := strings.LastIndex(raw, ",")
	if idx < 0 {
		return raw, ""
	}
	num := strings.TrimSpace(raw[idx+1:])
	if len(num) > 10 {
		return raw, "" // unlikely to be just a number
	}
	return strings.TrimSpace(raw[:idx]), num
}

// cityName extracts the city name from a formatted code-name string.
func cityName(raw string) string {
	if idx := strings.Index(raw, " - "); idx >= 0 {
		return strings.TrimSpace(raw[idx+3:])
	}
	return raw
}
