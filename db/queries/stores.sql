-- name: UpsertStore :exec
INSERT INTO stores (
    cnpj, name, trade_name, state_reg,
    street, number, complement, district, city, state, zip_code
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (cnpj) DO UPDATE SET
    name       = EXCLUDED.name,
    trade_name = EXCLUDED.trade_name,
    state_reg  = EXCLUDED.state_reg,
    street     = EXCLUDED.street,
    number     = EXCLUDED.number,
    complement = EXCLUDED.complement,
    district   = EXCLUDED.district,
    city       = EXCLUDED.city,
    state      = EXCLUDED.state,
    zip_code   = EXCLUDED.zip_code;

-- name: GetStoreByCNPJ :one
SELECT * FROM stores WHERE cnpj = $1;
