// Package store persists parsed NFC-e receipts to the database.
package store

import (
	"context"
	"errors"

	"github.com/glwbr/paw/internal/db"
	"github.com/glwbr/paw/nfce"
	"github.com/jackc/pgx/v5"
)

// SaveReceipt persists a parsed receipt and its related records.
// It returns the new receipt ID, or 0 if the receipt already existed (idempotent).
func SaveReceipt(ctx context.Context, q *db.Queries, r nfce.Receipt) (int64, error) {
	if err := q.UpsertStore(ctx, db.UpsertStoreParams{
		Cnpj:       r.Store.CNPJ,
		Name:       r.Store.Name,
		TradeName:  r.Store.TradeName,
		StateReg:   r.Store.StateReg,
		Street:     r.Store.Address.Street,
		Number:     r.Store.Address.Number,
		Complement: r.Store.Address.Complement,
		District:   r.Store.Address.District,
		City:       r.Store.Address.City,
		State:      r.Store.Address.State,
		ZipCode:    r.Store.Address.ZipCode,
	}); err != nil {
		return 0, err
	}

	receiptID, err := q.InsertReceipt(ctx, db.InsertReceiptParams{
		AccessKey:      r.AccessKey,
		Number:         r.Number,
		Series:         r.Series,
		Model:          r.Model,
		IssuedAt:       r.IssuedAt,
		AuthProtocol:   r.AuthProtocol,
		Source:         r.Source,
		State:          r.State,
		StoreCnpj:      r.Store.CNPJ,
		Subtotal:       r.Totals.Subtotal,
		DiscountAmount: r.Totals.DiscountAmount,
		TotalAmount:    r.Totals.TotalAmount,
		ApproxTaxTotal: r.Totals.ApproxTaxTotal,
		ParsedAt:       r.ParsedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil // already exists (ON CONFLICT DO NOTHING)
	}
	if err != nil {
		return 0, err
	}

	for _, item := range r.Items {
		p := db.InsertReceiptItemParams{
			ReceiptID:       receiptID,
			Sequence:        item.Sequence,
			ProductCode:     item.ProductCode,
			Description:     item.Description,
			Quantity:        item.Quantity,
			RawUnit:         item.Unit,
			Unit:            item.NormalizedUnit,
			UnitPrice:       item.UnitPrice,
			TotalPrice:      item.TotalPrice,
			NcmCode:         item.NCM,
			Cest:            item.CEST,
			Cfop:            item.CFOP,
			GtinCommercial:  item.GTINCommercial,
			ApproxTaxAmount: item.ApproxTaxAmount,
		}
		if t := item.Taxes; t != nil {
			if t.ICMS != nil {
				p.IcmsRate = &t.ICMS.Rate
				p.IcmsBaseAmount = &t.ICMS.BaseAmount
				p.IcmsAmount = &t.ICMS.Amount
			}
			if t.PIS != nil {
				p.PisRate = &t.PIS.Rate
				p.PisBaseAmount = &t.PIS.BaseAmount
				p.PisAmount = &t.PIS.Amount
			}
			if t.COFINS != nil {
				p.CofinsRate = &t.COFINS.Rate
				p.CofinsBaseAmount = &t.COFINS.BaseAmount
				p.CofinsAmount = &t.COFINS.Amount
			}
		}
		if err := q.InsertReceiptItem(ctx, p); err != nil {
			return 0, err
		}
	}

	for _, pay := range r.Payments {
		if err := q.InsertReceiptPayment(ctx, db.InsertReceiptPaymentParams{
			ReceiptID: receiptID,
			Method:    pay.Method,
			Amount:    pay.Amount,
		}); err != nil {
			return 0, err
		}
	}

	return receiptID, nil
}
