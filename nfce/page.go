package nfce

// Page is a pre-fetched HTML page from a SEFAZ portal.
type Page struct {
	Name    string // optional — improves error messages (filename, URL)
	Content []byte
}
