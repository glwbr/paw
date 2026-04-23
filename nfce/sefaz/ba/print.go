package ba

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
)

func parsePrint(doc *goquery.Document, result *nfce.Result) {
	parseReceipt(doc.Find("#NFe"), result)
	parseIssuer(doc.Find("#Emitente"), result)
	parseItems(doc.Find("#Prod"), result)
	parseCharges(doc.Find("#Cobranca"), result)
	parseTotals(doc.Find("#Totais"), result)
}
