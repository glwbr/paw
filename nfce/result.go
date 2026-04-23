package nfce

// Result is the output of parsing a set of NFC-e pages.
// Errors are per-page failures; processing continues on error so
// Warnings and partial Receipt data are still available.
type Result struct {
	Receipt  *Receipt
	Warnings []string // non-fatal: unknown tab, empty section, duplicate page
	Errors   []string // per-page parse failures
}

// HasErrors reports whether any parse errors occurred.
func (r *Result) HasErrors() bool { return len(r.Errors) > 0 }
