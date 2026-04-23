-- Core schema for NFC-e receipt ingestion.
-- Scope: stores, receipts, receipt_items, receipt_payments.
-- Unit normalization lives in Go (nfce.NormalizedUnit). GTIN enrichment and
-- product/brand/NCM resolution belong to other services.

-- ---------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------

CREATE TYPE payment_method AS ENUM (
    'cash',
    'check',
    'credit_card',
    'debit_card',
    'store_credit',
    'food_voucher',
    'meal_voucher',
    'gift_voucher',
    'fuel_voucher',
    'bank_slip',
    'bank_deposit',
    'pix',
    'bank_transfer',
    'loyalty_credit',
    'no_payment',
    'other'
);

-- Canonical unit codes. Matches nfce.NormalizedUnit.String() exactly.
CREATE TYPE unit_code AS ENUM (
    'KG', 'G', 'L', 'ML', 'M',  -- measure
    'PC', 'CX', 'PCT', 'DZ',    -- pack
    'UNKNOWN'
);

-- ---------------------------------------------------------------------------
-- Common trigger
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- Stores
-- ---------------------------------------------------------------------------

CREATE TABLE stores (
    cnpj       CHAR(14)    PRIMARY KEY,
    name       TEXT        NOT NULL,
    trade_name TEXT        NOT NULL DEFAULT '',
    state_reg  TEXT        NOT NULL DEFAULT '',
    street     TEXT        NOT NULL DEFAULT '',
    number     TEXT        NOT NULL DEFAULT '',
    complement TEXT        NOT NULL DEFAULT '',
    district   TEXT        NOT NULL DEFAULT '',
    city       TEXT        NOT NULL DEFAULT '',
    state      CHAR(2)     NOT NULL DEFAULT '',
    zip_code   CHAR(8)     NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_stores_updated_at
    BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Receipts
-- ---------------------------------------------------------------------------

CREATE TABLE receipts (
    id               BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    access_key       CHAR(44)    NOT NULL UNIQUE,
    number           INT         NOT NULL,
    series           INT         NOT NULL,
    model            INT         NOT NULL DEFAULT 65,
    issued_at        TIMESTAMPTZ NOT NULL,
    auth_protocol    TEXT        NOT NULL DEFAULT '',
    source           TEXT        NOT NULL DEFAULT '',
    state            CHAR(2)     NOT NULL,
    store_cnpj       CHAR(14)    NOT NULL REFERENCES stores(cnpj),
    subtotal         BIGINT      NOT NULL DEFAULT 0,
    discount_amount  BIGINT      NOT NULL DEFAULT 0,
    total_amount     BIGINT      NOT NULL,
    approx_tax_total BIGINT      NOT NULL DEFAULT 0,
    parsed_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_receipts_store_cnpj ON receipts(store_cnpj);
CREATE INDEX idx_receipts_issued_at  ON receipts(issued_at);
CREATE INDEX idx_receipts_state      ON receipts(state);

-- ---------------------------------------------------------------------------
-- Receipt items
-- ---------------------------------------------------------------------------

CREATE TABLE receipt_items (
    id                 BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    receipt_id         BIGINT      NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    sequence           INT         NOT NULL,
    product_code       TEXT        NOT NULL DEFAULT '',
    description        TEXT        NOT NULL,
    quantity           BIGINT      NOT NULL,
    raw_unit           TEXT        NOT NULL DEFAULT '',
    unit               unit_code   NOT NULL DEFAULT 'UNKNOWN',
    unit_price         BIGINT      NOT NULL,
    total_price        BIGINT      NOT NULL,
    ncm                CHAR(8),
    cest               TEXT        NOT NULL DEFAULT '',
    cfop               TEXT        NOT NULL DEFAULT '',
    gtin_commercial    TEXT        NOT NULL DEFAULT '',
    approx_tax_amount  BIGINT      NOT NULL DEFAULT 0,
    icms_rate          BIGINT,
    icms_base_amount   BIGINT,
    icms_amount        BIGINT,
    pis_rate           BIGINT,
    pis_base_amount    BIGINT,
    pis_amount         BIGINT,
    cofins_rate        BIGINT,
    cofins_base_amount BIGINT,
    cofins_amount      BIGINT,
    UNIQUE (receipt_id, sequence)
);

CREATE INDEX idx_receipt_items_receipt_id ON receipt_items(receipt_id);
CREATE INDEX idx_receipt_items_gtin
    ON receipt_items(gtin_commercial) WHERE gtin_commercial != '';

-- ---------------------------------------------------------------------------
-- Receipt payments
-- ---------------------------------------------------------------------------

CREATE TABLE receipt_payments (
    id         BIGINT         GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    receipt_id BIGINT         NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    method     payment_method NOT NULL,
    amount     BIGINT         NOT NULL
);

CREATE INDEX idx_receipt_payments_receipt_id ON receipt_payments(receipt_id);
