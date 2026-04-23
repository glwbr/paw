-- name: InsertReceiptItem :exec
INSERT INTO receipt_items (
    receipt_id, sequence, product_code, description,
    quantity, raw_unit, unit, unit_price, total_price,
    ncm, cest, cfop, gtin_commercial, approx_tax_amount,
    icms_rate, icms_base_amount, icms_amount,
    pis_rate, pis_base_amount, pis_amount,
    cofins_rate, cofins_base_amount, cofins_amount
) VALUES (
    @receipt_id, @sequence, @product_code, @description,
    @quantity, @raw_unit, @unit, @unit_price, @total_price,
    NULLIF(@ncm_code::text, ''),
    @cest, @cfop, @gtin_commercial, @approx_tax_amount,
    @icms_rate, @icms_base_amount, @icms_amount,
    @pis_rate, @pis_base_amount, @pis_amount,
    @cofins_rate, @cofins_base_amount, @cofins_amount
);

-- name: ListItemsByReceipt :many
SELECT * FROM receipt_items
WHERE receipt_id = $1
ORDER BY sequence;
