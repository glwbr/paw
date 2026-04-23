package nfce

import "time"

// Receipt is the unified output of parsing any NFC-e, regardless of state.
// All monetary values use the Money type (centavos) to avoid float precision issues.
// This struct is the contract — consumers of this library depend on this shape.
type Receipt struct {
	Source       string    `json:"source"`
	AccessKey    string    `json:"access_key"` // 44-digit NFC-e chave de acesso
	Number       int       `json:"number"`
	Series       int       `json:"series"`
	Model        int       `json:"model"` // 65 for NFC-e
	IssuedAt     time.Time `json:"issued_at"`
	AuthProtocol string    `json:"auth_protocol"`

	Store Store `json:"store"`

	Items []Item `json:"items"`

	Totals Totals `json:"totals"`

	Payments []Payment `json:"payments"`

	State    UF        `json:"state"`
	RawHTML  []byte    `json:"-"` // original HTML (not serialized)
	ParsedAt time.Time `json:"parsed_at"`
}

type Address struct {
	Street     string `json:"street,omitempty"`
	Number     string `json:"number,omitempty"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	ZipCode    string `json:"zip_code,omitempty"`
}

type Store struct {
	CNPJ      string  `json:"cnpj"`
	Name      string  `json:"name"`
	TradeName string  `json:"trade_name,omitempty"`
	StateReg  string  `json:"state_reg,omitempty"`
	Address   Address `json:"address"`
}

type Totals struct {
	Subtotal       Money `json:"subtotal,omitempty"`
	DiscountAmount Money `json:"discount_amount,omitempty"`
	TotalAmount    Money `json:"total_amount"`
	ApproxTaxTotal Money `json:"approx_tax_total"` // Lei 12.741
}

type Payment struct {
	Method PaymentMethod `json:"method"`
	Amount Money         `json:"amount"`
}

func (r *Receipt) Validate() error {
	if r.AccessKey == "" {
		return &ValidationError{Receipt: r, Err: ErrMissingAccessKey}
	}
	if len(r.AccessKey) != 44 {
		return &ValidationError{Receipt: r, Err: ErrInvalidAccessKey}
	}
	if r.Store.CNPJ == "" {
		return &ValidationError{Receipt: r, Err: ErrMissingCNPJ}
	}
	if len(r.Items) == 0 {
		return &ValidationError{Receipt: r, Err: ErrNoItems}
	}
	if r.Totals.TotalAmount.IsZero() {
		return &ValidationError{Receipt: r, Err: ErrZeroTotal}
	}
	return nil
}

func (r *Receipt) ItemCount() int { return len(r.Items) }

func (r *Receipt) TotalQuantity() Quantity {
	var total Quantity
	for _, item := range r.Items {
		total += item.Quantity
	}
	return total
}
