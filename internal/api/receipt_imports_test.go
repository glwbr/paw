package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON_ReceiptID(t *testing.T) {
	t.Parallel()

	t.Run("with receipt ID", func(t *testing.T) {
		id := int64(42)
		ri := &ReceiptImport{
			ID:        "test-import-id",
			status:    StatusCompleted,
			ReceiptID: &id,
		}

		data, err := json.Marshal(ri)
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var res map[string]any
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		got, ok := res["receipt_id"]
		if !ok {
			t.Fatal("expected receipt_id in JSON, but it was missing")
		}

		gotFloat, ok := got.(float64)
		if !ok {
			t.Fatalf("expected receipt_id to be a number, got %T", got)
		}

		if int64(gotFloat) != 42 {
			t.Errorf("expected receipt_id to be 42, got %v", gotFloat)
		}
	})

	t.Run("without receipt ID", func(t *testing.T) {
		ri := &ReceiptImport{
			ID:     "test-import-id",
			status: StatusPending,
		}

		data, err := json.Marshal(ri)
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var res map[string]any
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if _, ok := res["receipt_id"]; ok {
			t.Error("expected receipt_id to be omitted from JSON when nil")
		}
	})
}
