package ba

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// formState holds ASP.NET hidden fields extracted from each portal response.
type formState struct {
	viewState          string
	viewStateGenerator string
	eventValidation    string
	lastFocus          string
	eventTarget        string
	eventArgument      string
}

func (fs *formState) valid() bool {
	return fs.viewState != ""
}

// parseFormState extracts ASP.NET hidden fields from HTML.
func parseFormState(html []byte) (*formState, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing form state: %w", err)
	}

	get := func(name string) string {
		val, _ := doc.Find(fmt.Sprintf("input[name='%s']", name)).Attr("value")
		return val
	}

	return &formState{
		viewState:          get("__VIEWSTATE"),
		viewStateGenerator: get("__VIEWSTATEGENERATOR"),
		eventValidation:    get("__EVENTVALIDATION"),
		lastFocus:          get("__LASTFOCUS"),
		eventTarget:        get("__EVENTTARGET"),
		eventArgument:      get("__EVENTARGUMENT"),
	}, nil
}

// buildForm merges form state hidden fields with additional key-value pairs
// into url.Values ready for POST submission.
func buildForm(fs *formState, extra map[string]string) url.Values {
	vals := make(url.Values, 12)

	if fs != nil {
		if fs.viewState != "" {
			vals.Set("__VIEWSTATE", fs.viewState)
		}
		if fs.viewStateGenerator != "" {
			vals.Set("__VIEWSTATEGENERATOR", fs.viewStateGenerator)
		}
		if fs.eventValidation != "" {
			vals.Set("__EVENTVALIDATION", fs.eventValidation)
		}
		vals.Set("__LASTFOCUS", fs.lastFocus)
		vals.Set("__EVENTTARGET", fs.eventTarget)
		vals.Set("__EVENTARGUMENT", fs.eventArgument)
	}

	for k, v := range extra {
		vals.Set(k, v)
	}

	return vals
}
