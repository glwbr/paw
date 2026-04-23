// Package ba implements the NFC-e parser for the Bahia (BA) SEFAZ portal.
// Import it with a blank import to register the parser:
//
//	import _ "github.com/glwbr/paw/nfce/sefaz/ba"
package ba

import (
	"time"

	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func init() {
	nfce.Register(&Parser{})
}

// Parser implements nfce.StateParser for the Bahia SEFAZ portal.
type Parser struct{}

func (p *Parser) State() nfce.UF { return nfce.BA }

// CanParse returns true if the page is from the Bahia SEFAZ portal.
func (p *Parser) CanParse(page []byte) bool {
	doc, err := htmlparse.Doc(page)
	if err != nil {
		return false
	}
	return doc.Find("#lbl_cod_chave_acesso").Length() > 0
}

// Assemble parses a print page into a single Receipt.
func (p *Parser) Assemble(pages ...nfce.Page) *nfce.Result {
	result := &nfce.Result{
		Receipt: &nfce.Receipt{
			State:    nfce.BA,
			ParsedAt: time.Now(),
		},
	}

	for _, pg := range pages {
		doc, err := htmlparse.Doc(pg.Content)
		if err != nil {
			result.Errors = append(result.Errors, pg.Name+": "+err.Error())
			continue
		}

		raw := htmlparse.HeaderField(doc, "lbl_cod_chave_acesso")
		if raw == "" {
			result.Warnings = append(result.Warnings, pg.Name+": access key not found")
			continue
		}
		key, err := nfce.ParseAccessKey(raw)
		if err != nil {
			result.Warnings = append(result.Warnings, pg.Name+": bad access key: "+err.Error())
			continue
		}
		result.Receipt.AccessKey = key.Raw
		result.Receipt.Model = key.Model
		result.Receipt.Series = key.Series
		result.Receipt.Number = key.Number

		parsePrint(doc, result)
	}

	return result
}
