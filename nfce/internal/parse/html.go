package parse

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Doc parses raw HTML bytes into a goquery document.
func Doc(b []byte) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(bytes.NewReader(b))
}

// HeaderField returns the trimmed text of the element with the given id.
func HeaderField(doc *goquery.Document, id string) string {
	return Text(doc.Find("#" + id).Text())
}

// LabeledField returns the text of the field labeled with labelText.
func LabeledField(sel *goquery.Selection, labelText string) string {
	var found string
	sel.Find("label").EachWithBreak(func(_ int, l *goquery.Selection) bool {
		if Text(l.Text()) == labelText {
			span := l.NextFiltered("span")
			if span.Length() == 0 {
				span = l.Parent().Find("span").First()
			}
			found = Text(span.Text())
			return false
		}
		return true
	})
	return found
}

// SectionTable returns the table for the section with the given title.
func SectionTable(root *goquery.Selection, title string) *goquery.Selection {
	var found *goquery.Selection
	root.Find("td.table-titulo-aba-interna").EachWithBreak(func(_ int, td *goquery.Selection) bool {
		if strings.Contains(td.Text(), title) {
			found = td.Closest("table").Next()
			return false
		}
		return true
	})
	if found == nil {
		return root.End()
	}
	return found
}

// EachToggleGroup iterates over toggle/detail pairs within sel, calling fn with each pair.
func EachToggleGroup(sel *goquery.Selection, fn func(toggle, detail *goquery.Selection)) {
	sel.Find("table.toggle").Each(func(_ int, toggle *goquery.Selection) {
		detail := toggle.Next()
		if detail.Length() > 0 && detail.HasClass("toggable") {
			fn(toggle, detail)
		}
	})
}

// CodePrefix extracts the leading code from "XX - description" patterns.
func CodePrefix(s string) string {
	s = Text(s)
	if idx := strings.Index(s, " - "); idx > 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}
