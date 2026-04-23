package ba

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/glwbr/paw/nfce"
	htmlparse "github.com/glwbr/paw/nfce/internal/parse"
)

func parseCharges(content *goquery.Selection, result *nfce.Result) {
	if content.Length() == 0 {
		result.Warnings = append(result.Warnings, "cobranca tab content not found")
		return
	}

	htmlparse.EachToggleGroup(content, func(toggle, detail *goquery.Selection) {
		methodRaw := htmlparse.Text(toggle.Find("td").Eq(1).Find("span.linha").Text())
		method := nfce.NormalizePayment(zeroPadCode(methodRaw))

		amountRaw := findPaymentAmount(detail)
		amount, err := nfce.ParseMoney(amountRaw)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("charges: amount %q: %v", amountRaw, err))
			return
		}

		result.Receipt.Payments = append(result.Receipt.Payments, nfce.Payment{
			Method: method,
			Amount: amount,
		})
	})
}

// zeroPadCode zero-pads the leading code from "X - Description" to match payment method map keys.
func zeroPadCode(s string) string {
	code := htmlparse.CodePrefix(s)
	if n, err := strconv.Atoi(code); err == nil {
		return fmt.Sprintf("%02d", n)
	}
	return s
}

func findPaymentAmount(detail *goquery.Selection) string {
	var amount string
	detail.Find("label").EachWithBreak(func(_ int, l *goquery.Selection) bool {
		if strings.TrimSpace(l.Text()) == "Valor do Pagamento" {
			row := l.Closest("tr").Next()
			amount = htmlparse.Text(row.Find("td").First().Find("span.linha").Text())
			return false
		}
		return true
	})
	return amount
}
