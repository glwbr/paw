package nfce

// Item represents a single line item on a receipt.
type Item struct {
	Sequence    int    `json:"sequence"`
	ProductCode string `json:"product_code"` // internal store SKU (cProd)
	Description string `json:"description"`  // raw POS description

	Quantity       Quantity       `json:"quantity"` // scaled by QuantityScale (×10000)
	Unit           string         `json:"unit"`     // raw unit: "PC", "Kg", "UN", "LT"
	NormalizedUnit NormalizedUnit `json:"normalized_unit"`
	UnitPrice      Money          `json:"unit_price"`
	TotalPrice     Money          `json:"total_price"`

	NCM  string `json:"ncm"`
	CEST string `json:"cest,omitempty"`
	CFOP string `json:"cfop"`

	GTINCommercial string `json:"gtin_commercial,omitempty"`

	Taxes *ItemTaxes `json:"taxes"`

	ApproxTaxAmount Money `json:"approx_tax_amount"` // transparency law
}

// ItemTaxes contains tax information for an item (ICMS, PIS, COFINS).
type ItemTaxes struct {
	ICMS   *ICMS   `json:"icms,omitempty"`
	PIS    *PIS    `json:"pis,omitempty"`
	COFINS *COFINS `json:"cofins,omitempty"`
}

// ICMS holds Brazil's state-level VAT (Imposto sobre Circulação de Mercadorias e Serviços) data for an item.
type ICMS struct {
	Rate       Rate  `json:"rate"`
	BaseAmount Money `json:"base_amount"`
	Amount     Money `json:"amount"`
}

// PIS represents PIS tax information.
type PIS struct {
	Rate       Rate  `json:"rate"`
	BaseAmount Money `json:"base_amount"`
	Amount     Money `json:"amount"`
}

// COFINS represents COFINS tax information.
type COFINS struct {
	Rate       Rate  `json:"rate"`
	BaseAmount Money `json:"base_amount"`
	Amount     Money `json:"amount"`
}
