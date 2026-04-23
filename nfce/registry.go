package nfce

import (
	"fmt"
	"sync"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// StateParser is implemented by each state-specific SEFAZ parser.
type StateParser interface {
	State() UF
	// CanParse returns true if the HTML looks like a page from this state's portal.
	CanParse(page []byte) bool
	// Assemble builds a Result from one or more pre-fetched HTML pages.
	Assemble(pages ...Page) *Result
}

var (
	mu       sync.RWMutex
	registry = map[UF]StateParser{}
)

// Register adds a StateParser to the global registry.
// Typically called from an init() function in a state sub-package.
func Register(p StateParser) {
	mu.Lock()
	defer mu.Unlock()
	registry[p.State()] = p
}

// Parse auto-detects the state from the pages and delegates to the matching StateParser.
// It first tries to extract the access key from page #lbl_chave_acesso to determine state via UFFromCode,
// then falls back to CanParse() on each registered parser.
func Parse(pages ...Page) *Result {
	for _, pg := range pages {
		doc, err := parse.Doc(pg.Content)
		if err != nil {
			continue
		}
		raw := parse.HeaderField(doc, "lbl_chave_acesso")
		if raw == "" {
			continue
		}
		key, err := ParseAccessKey(raw)
		if err != nil {
			continue
		}
		uf, ok := UFFromCode(key.UFCode)
		if !ok {
			continue
		}
		return ParseState(uf, pages...)
	}

	mu.RLock()
	defer mu.RUnlock()
	for _, p := range registry {
		for _, pg := range pages {
			if p.CanParse(pg.Content) {
				return p.Assemble(pages...)
			}
		}
	}

	r := &Result{Receipt: &Receipt{}}
	r.Errors = append(r.Errors, ErrUndetectedState.Error())
	return r
}

// ParseState bypasses auto-detection and uses the parser for the given state.
func ParseState(state UF, pages ...Page) *Result {
	mu.RLock()
	p, ok := registry[state]
	mu.RUnlock()
	if !ok {
		r := &Result{Receipt: &Receipt{}}
		r.Errors = append(r.Errors, fmt.Sprintf("%s: %s", ErrUnsupportedState, state))
		return r
	}
	return p.Assemble(pages...)
}
