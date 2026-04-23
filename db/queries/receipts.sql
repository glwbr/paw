-- name: InsertReceipt :one
INSERT INTO receipts (
    access_key, number, series, model, issued_at, auth_protocol,
    source, state, store_cnpj,
    subtotal, discount_amount, total_amount, approx_tax_total,
    parsed_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13,
    $14
)
ON CONFLICT (access_key) DO NOTHING
RETURNING id;

-- name: GetReceiptByAccessKey :one
SELECT r.*, s.name AS store_name, s.trade_name AS store_trade_name
FROM receipts r
JOIN stores s ON s.cnpj = r.store_cnpj
WHERE r.access_key = $1;

-- name: ListReceiptsByStore :many
SELECT id, access_key, number, series, issued_at, state,
       subtotal, discount_amount, total_amount, approx_tax_total, parsed_at
FROM receipts
WHERE store_cnpj = $1
ORDER BY issued_at DESC
LIMIT $2 OFFSET $3;

-- name: ListReceiptsByDateRange :many
SELECT id, access_key, number, series, issued_at, state, store_cnpj,
       total_amount, parsed_at
FROM receipts
WHERE issued_at >= $1 AND issued_at < $2
ORDER BY issued_at DESC
LIMIT $3 OFFSET $4;

-- name: ReceiptExistsByAccessKey :one
SELECT EXISTS(SELECT 1 FROM receipts WHERE access_key = $1);
