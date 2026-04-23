package nfce_test

import (
	"testing"

	"github.com/glwbr/paw/nfce"
)

func TestParseAccessKey(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantUF     int
		wantYM     string
		wantCNPJ   string
		wantModel  int
		wantSeries int
		wantNum    int
		wantErr    bool
	}{
		{
			name:      "ba issuer key with spaces",
			raw:       "2924 0112 3456 7800 0190 6500 1000 0034 0263 1101 9296",
			wantUF:    29,
			wantYM:    "2401",
			wantCNPJ:  "12345678000190",
			wantModel: 65,
		},
		{
			name:    "too short",
			raw:     "12345",
			wantErr: true,
		},
		{
			name:    "contains non-digits",
			raw:     "ABCD 0112 3456 7800 0190 6500 1000 0034 0263 1101 9296",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nfce.ParseAccessKey(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.UFCode != tc.wantUF {
				t.Errorf("UFCode = %d, want %d", got.UFCode, tc.wantUF)
			}
			if got.YearMonth != tc.wantYM {
				t.Errorf("YearMonth = %q, want %q", got.YearMonth, tc.wantYM)
			}
			if got.CNPJ != tc.wantCNPJ {
				t.Errorf("CNPJ = %q, want %q", got.CNPJ, tc.wantCNPJ)
			}
			if got.Model != tc.wantModel {
				t.Errorf("Model = %d, want %d", got.Model, tc.wantModel)
			}
		})
	}
}

func TestUFFromCode(t *testing.T) {
	tests := []struct {
		code   int
		want   nfce.UF
		wantOK bool
	}{
		{29, nfce.BA, true},
		{31, nfce.MG, true},
		{35, "SP", true},
		{99, "", false},
	}
	for _, tc := range tests {
		got, ok := nfce.UFFromCode(tc.code)
		if ok != tc.wantOK {
			t.Errorf("UFFromCode(%d) ok = %v, want %v", tc.code, ok, tc.wantOK)
		}
		if got != tc.want {
			t.Errorf("UFFromCode(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}
