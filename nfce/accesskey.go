package nfce

import (
	"fmt"
	"strconv"
	"strings"
)

// AccessKey is a decomposed NFC-e 44-digit chave de acesso.
type AccessKey struct {
	Raw       string
	UFCode    int    // positions 1-2: IBGE state code (29=BA, 31=MG, 35=SP, …)
	YearMonth string // positions 3-6: AAMM (e.g. "2401" = Jan 2024)
	CNPJ      string // positions 7-20: emitter CNPJ, 14 digits unformatted
	Model     int    // positions 21-22: 65 for NFC-e
	Series    int    // positions 23-25: 3-digit series number
	Number    int    // positions 26-34: 9-digit sequential document number
}

// ParseAccessKey parses a raw access key string (spaces are ignored).
func ParseAccessKey(raw string) (AccessKey, error) {
	digits := strings.ReplaceAll(raw, " ", "")
	if len(digits) != 44 {
		return AccessKey{}, fmt.Errorf("parse access key: %w", ErrInvalidAccessKeyFormat)
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return AccessKey{}, fmt.Errorf("parse access key: %w", ErrInvalidAccessKeyFormat)
		}
	}

	ufCode, _ := strconv.Atoi(digits[0:2])
	model, _ := strconv.Atoi(digits[20:22])
	series, _ := strconv.Atoi(digits[22:25])
	number, _ := strconv.Atoi(digits[25:34])

	return AccessKey{
		Raw:       digits,
		UFCode:    ufCode,
		YearMonth: digits[2:6],
		CNPJ:      digits[6:20],
		Model:     model,
		Series:    series,
		Number:    number,
	}, nil
}
