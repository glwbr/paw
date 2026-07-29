package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/internal/db"
	"github.com/glwbr/paw/nfce"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m mockRow) Scan(dest ...any) error {
	return m.scanFunc(dest...)
}

type mockDBTX struct {
	t            *testing.T
	upsertCalled bool
	insertCalled bool
	getByKey     bool
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "INSERT INTO stores") {
		m.upsertCalled = true
		return pgconn.CommandTag{}, nil
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "INSERT INTO receipts") {
		m.insertCalled = true
		return mockRow{scanFunc: func(dest ...any) error {
			return pgx.ErrNoRows
		}}
	}
	if strings.Contains(sql, "SELECT r.id, r.access_key") {
		m.getByKey = true
		return mockRow{scanFunc: func(dest ...any) error {
			if len(dest) < 18 {
				return fmt.Errorf("expected at least 18 fields, got %d", len(dest))
			}
			idVal, ok := dest[0].(*int64)
			if !ok {
				return fmt.Errorf("expected field 0 to be *int64, got %T", dest[0])
			}
			*idVal = 456
			return nil
		}}
	}
	return mockRow{scanFunc: func(dest ...any) error { return nil }}
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

type testParser struct {
	receipt *nfce.Receipt
}

func (tp *testParser) State() nfce.UF {
	return nfce.UF("TEST")
}

func (tp *testParser) CanParse(page []byte) bool {
	return string(page) == "test-html-page"
}

func (tp *testParser) Assemble(pages ...nfce.Page) *nfce.Result {
	return &nfce.Result{
		Receipt: tp.receipt,
	}
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := &ReceiptImport{
		ID:        "test-id",
		status:    StatusWaitingCaptcha,
		AccessKey: "12345678901234567890123456789012345678901234",
	}

	data, err := ri.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error marshaling: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("unexpected error unmarshaling: %v", err)
	}

	if captchaURL, ok := res["captcha_url"]; !ok || captchaURL != "/receipts/imports/test-id/captcha" {
		t.Errorf("expected captcha_url to be '/receipts/imports/test-id/captcha', got %v", captchaURL)
	}

	res = make(map[string]any)
	ri.status = StatusPending
	data, err = ri.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error marshaling: %v", err)
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("unexpected error unmarshaling: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("expected captcha_url to be omitted when not waiting for captcha")
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, nil)

	reqBody := `{"access_key": "invalid-key-short"}`
	req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	err := s.createImport(w, req)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var he *apierrors.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected apierrors.HTTPError, got %T", err)
	}

	if he.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", he.Status)
	}

	if !strings.Contains(he.Error(), "access key must be 44 numeric digits") {
		t.Errorf("expected access key validation message, got %q", he.Error())
	}
}

type mockFetcher struct {
	fetchFunc func(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error)
}

func (m *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return m.fetchFunc(ctx, req)
}

func TestReceiptImport_Run_IdempotentDuplicate(t *testing.T) {
	mockReceipt := &nfce.Receipt{
		AccessKey: "12345678901234567890123456789012345678901234",
		Store: nfce.Store{
			CNPJ: "12345678901234",
			Name: "Test Store",
		},
		Items: []nfce.Item{
			{
				Sequence:    1,
				ProductCode: "001",
				Description: "Product 1",
			},
		},
		Totals: nfce.Totals{
			TotalAmount: 100,
		},
	}

	tp := &testParser{receipt: mockReceipt}
	nfce.Register(tp)

	ri := &ReceiptImport{
		ID:        "import-id",
		AccessKey: "12345678901234567890123456789012345678901234",
		status:    StatusPending,
	}

	mockFetcherInst := &mockFetcher{
		fetchFunc: func(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
			return &sefaz.Result{
				Page: []byte("test-html-page"),
			}, nil
		},
	}

	factory := func(solver captcha.Solver) sefaz.Fetcher {
		return mockFetcherInst
	}

	mDB := &mockDBTX{t: t}
	queries := db.New(mDB)

	ri.run(factory, queries)

	if ri.status != StatusCompleted {
		t.Errorf("expected status completed, got %v", ri.status)
	}

	if !mDB.upsertCalled {
		t.Errorf("expected UpsertStore to be called")
	}

	if !mDB.insertCalled {
		t.Errorf("expected InsertReceipt to be called")
	}

	if !mDB.getByKey {
		t.Errorf("expected GetReceiptByAccessKey to be called")
	}

	if ri.ReceiptID == nil || *ri.ReceiptID != 456 {
		t.Errorf("expected ReceiptID to be set to 456, got %v", ri.ReceiptID)
	}
}
