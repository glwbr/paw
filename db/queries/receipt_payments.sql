-- name: InsertReceiptPayment :exec
INSERT INTO receipt_payments (receipt_id, method, amount)
VALUES ($1, $2, $3);

-- name: ListPaymentsByReceipt :many
SELECT * FROM receipt_payments
WHERE receipt_id = $1;
