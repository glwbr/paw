package nfce

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// PaymentMethod represents the NFC-e payment method codes (tPag).
// The NFC-e standard defines numeric codes, but SEFAZ HTML pages
// often show the text description. We normalize both.
type PaymentMethod string

const (
	PaymentCash          PaymentMethod = "cash"
	PaymentCheck         PaymentMethod = "check"
	PaymentCreditCard    PaymentMethod = "credit_card"
	PaymentDebitCard     PaymentMethod = "debit_card"
	PaymentStoreCredit   PaymentMethod = "store_credit"
	PaymentFoodVoucher   PaymentMethod = "food_voucher"
	PaymentMealVoucher   PaymentMethod = "meal_voucher"
	PaymentGiftVoucher   PaymentMethod = "gift_voucher"
	PaymentFuelVoucher   PaymentMethod = "fuel_voucher"
	PaymentBankSlip      PaymentMethod = "bank_slip"
	PaymentBankDeposit   PaymentMethod = "bank_deposit"
	PaymentPix           PaymentMethod = "pix"
	PaymentBankTransfer  PaymentMethod = "bank_transfer"
	PaymentLoyaltyCredit PaymentMethod = "loyalty_credit"
	PaymentNoPayment     PaymentMethod = "no_payment"
	PaymentOther         PaymentMethod = "other"
)

// NFC-e tPag codes → PaymentMethod
var tPagMap = map[string]PaymentMethod{
	"01": PaymentCash,
	"02": PaymentCheck,
	"03": PaymentCreditCard,
	"04": PaymentDebitCard,
	"05": PaymentStoreCredit,
	"10": PaymentFoodVoucher,
	"11": PaymentMealVoucher,
	"12": PaymentGiftVoucher,
	"13": PaymentFuelVoucher,
	"15": PaymentBankSlip,
	"16": PaymentBankDeposit,
	"17": PaymentPix,
	"18": PaymentBankTransfer,
	"19": PaymentLoyaltyCredit,
	"90": PaymentNoPayment,
	"99": PaymentOther,
}

func (pm PaymentMethod) String() string { return string(pm) }

// NormalizePayment maps both numeric codes and free-text descriptions
// to a canonical PaymentMethod.
func NormalizePayment(raw string) PaymentMethod {
	raw = strings.TrimSpace(raw)

	if pm, ok := tPagMap[raw]; ok {
		return pm
	}

	normalized := parse.Normalize(raw)

	switch {
	case parse.ContainsAny(normalized, "DINHEIRO"):
		return PaymentCash

	case parse.ContainsAny(normalized, "CHEQUE"):
		return PaymentCheck

	case parse.ContainsAny(normalized, "CARTAO CREDITO", "CARTAO DE CREDITO"):
		return PaymentCreditCard

	case parse.ContainsAny(normalized, "CARTAO DEBITO", "CARTAO DE DEBITO"):
		return PaymentDebitCard

	case parse.ContainsAny(normalized, "PIX"):
		return PaymentPix

	case parse.ContainsAny(normalized, "BOLETO"):
		return PaymentBankSlip

	case parse.ContainsAny(normalized, "DEPOSITO"):
		return PaymentBankDeposit

	case parse.ContainsAny(normalized, "TRANSFERENCIA", "TED", "DOC"):
		return PaymentBankTransfer

	case parse.ContainsAny(normalized, "CREDITO LOJA", "CREDITO INTERNO"):
		return PaymentStoreCredit

	case parse.ContainsAny(normalized, "CASHBACK", "CREDITO FIDELIDADE"):
		return PaymentLoyaltyCredit

	case parse.ContainsAny(normalized, "VALE ALIMENTACAO"):
		return PaymentFoodVoucher

	case parse.ContainsAny(normalized, "VALE REFEICAO"):
		return PaymentMealVoucher

	case parse.ContainsAny(normalized, "VALE COMBUSTIVEL"):
		return PaymentFuelVoucher

	case parse.ContainsAny(normalized, "VALE PRESENTE"):
		return PaymentGiftVoucher

	case parse.ContainsAny(normalized, "VALE"):
		return PaymentOther

	case parse.ContainsAny(normalized, "SEM PAGAMENTO", "GRATUITO"):
		return PaymentNoPayment

	default:
		return PaymentOther
	}
}

func (pm PaymentMethod) Value() (driver.Value, error) { return string(pm), nil }

func (pm *PaymentMethod) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*pm = PaymentMethod(v)
		return nil
	case []byte:
		*pm = PaymentMethod(v)
		return nil
	case nil:
		*pm = PaymentOther
		return nil
	default:
		return fmt.Errorf("cannot scan %T into PaymentMethod", src)
	}
}
