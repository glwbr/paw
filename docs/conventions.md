# Conventions

Engineering decisions we landed on during the soft reset. These are the "why" behind the current shape of the code — if you change any of them, make it a deliberate decision, not a drift.

## Value objects (`nfce/Money`, `Quantity`, `Rate`, `NormalizedUnit`)

All monetary/numeric values are integer-backed custom types, never `float64`:

- `Money` — `int64` centavos. `R$ 12,90 → Money(1290)`.
- `Quantity` — `int64` scaled by 10⁴. `2.5 → Quantity(25000)`.
- `Rate` — `int64` scaled by 100. `20.50% → Rate(2050)`.
- `NormalizedUnit` — tagged struct (`Kind` + `MeasureUnit` or `PackUnit`) whose `String()` returns a canonical code (`KG`, `L`, `PC`, `UNKNOWN`, …).

### Scanner / Valuer contract

All four implement `database/sql/driver.Valuer` and `sql.Scanner`:

- `Value()` returns `int64` for numeric types, `string` for `NormalizedUnit` — the types pgx knows how to marshal without reflection hops.
- `Scan(src any)` accepts **`int64`, `int32`, `int`, `nil`** for integer-backed types. Don't assume pgx only hands back `int64` — drivers, column types, and aggregate expressions can produce any of them. Defensive matching is cheap and saves the next debugging round.
- `Scan` treats `nil` as the zero value (`Money(0)`, `Quantity(0)`, `Rate(0)`, `NormalizedUnit{Kind: UnitKindUnknown}`). No `pgtype.Null*` wrappers in consumer code.

### Constructor naming (`ParseX`, not `Parse`)

Go's "avoid stutter" rule says a package-qualified identifier shouldn't repeat the package name (`bytes.Buffer`, not `bytes.BytesBuffer`). **It applies when a package exposes one dominant type.** The `nfce` package holds `Money`, `Quantity`, `Rate`, `AccessKey`, `Receipt` — we can't have five `Parse` functions. `ParseMoney`, `ParseQuantity`, `ParseRate`, `ParseAccessKey` is the correct form: the type _is_ the disambiguator.

If you ever split these into one-type-per-package (`nfce/money/`, `nfce/quantity/`), then `money.Parse` becomes correct. Not worth the restructure today — the types call each other (`Rate.ApplyTo(Money)`), so subpackages would either duplicate code or create circular imports.

### Parse functions stay minimal

Each `ParseX`:

1. `parse.Text` → trim / collapse whitespace / strip NBSP.
2. Strip prefix/suffix specific to that type (`R$ `, `%`, leading `-`).
3. `parse.BRNumber` → `"1.197,00"` to `"1197.00"` (shared helper, don't duplicate).
4. `strconv.ParseFloat`, then `Type(math.Round(f * scale))`.

Idioms to keep:

- **Direct float → named-int conversion.** `Quantity(math.Round(f * 10000))` works — Go allows float → any named integer type. Don't write `Quantity(int64(math.Round(...)))`; the `int64` is redundant.
- **`ParseMoney` owns Brazilian currency parsing.** Don't shift the cleanup "up" to the HTML parser — you just relocate the work and every other caller loses the convenience. The function is 8 substantive lines; that's the right size.
- **`parse.BRNumber` is the single source** for "strip thousands separator, swap `,` for `.`". If you find yourself writing those two `ReplaceAll`s again, call the helper instead.

## Column types

### BIGINT for integer-backed value types

`Money`, `Quantity`, `Rate` → `BIGINT`. Rationale:

- All three are `type X int64` in Go. `BIGINT` is the 1:1 match — pgx hands the int64 straight to `Scan`.
- Downsizing to INT saves 4 bytes per column while forcing Scanners to handle int32 paths. The overflow headroom (R$ 21M for INT centavos, 214K raw quantity) evaporates if we ever sum across many receipts.
- Consistency: one rule — "if it's an `nfce` value-object column, it's `BIGINT`."

### INT for simple counters

`receipts.number`, `receipts.series`, `receipts.model`, `receipt_items.sequence` stay `INT`. Mapped by the global `pg_catalog.int4 → int` override — they read as plain `int` in Go, not `int32`, so no casts either.

### Why we don't keep a units table

`units` + `unit_aliases` used to hold conversion factors and alias rows. Everything there is now in Go:

- Canonical codes live in `nfce/unit.go` (`MeasureKg`, `PackUnitPiece`, …).
- Conversion factors live in `MeasureUnit.ConversionRatio()` and `PricePerBaseUnit`.
- Fuzzy alias matching (`"KILO" → KG`) lives in `NormalizeUnit`.

The DB's only job is to **refuse invalid codes**, which a Postgres enum does better than a FK + alias lookup. Hence `CREATE TYPE unit_code AS ENUM (...)` and `receipt_items.unit unit_code NOT NULL DEFAULT 'UNKNOWN'`. `NormalizedUnit.Scan` matches exact enum values only — no fuzzy re-normalization — because trusted DB input is already canonical.

**Pattern:** domain logic in Go, DB for validation. Don't duplicate Go-side rules as SQL tables.

## sqlc overrides

Config lives in `db/sqlc.yaml`. The two knobs that do the heavy lifting:

- `emit_pointers_for_null_types: true` — nullable columns become `*T` instead of `pgtype.Text` / `pgtype.Int8`. Only with `pgx/v5` + `pgx/v4`.
- Per-column overrides with `nullable: true` + `go_type: { ..., pointer: true }` — map nullable custom types to `*CustomType`.

What this buys us: no more `pgtype.Int8{Int64: int64(t.ICMS.Rate), Valid: true}` wrapping in `store/receipt.go`. Nullable tax fields assign directly:

```go
p.IcmsRate       = &t.ICMS.Rate        // *nfce.Rate
p.IcmsBaseAmount = &t.ICMS.BaseAmount  // *nfce.Money
```

### Rules when adding columns

- **Non-nullable custom type** → single override, no `nullable:`, no `pointer:`. Generates `T`.
- **Nullable custom type** → add `nullable: true` and `pointer: true`. Generates `*T`.
- sqlc matches one override per (column, nullability). If you want the same custom type for both, write two overrides — one with `nullable: true`, one without.
- Enum columns (`payment_method`, `unit_code`) → override to the Go type that owns the enum (`nfce.PaymentMethod`, `nfce.NormalizedUnit`). Their Scanner/Valuer handles the string round-trip.

## Error handling at API boundaries

### Problem

An error like `pgx: failed to connect to host=...` reaching an HTTP response body is a data leak — it exposes internal topology. Same concern for file paths, library internals, or any string built from runtime data.

### Two-package design

The contract splits along a library / application line:

| Package                  | Importable by | Purpose                                                                       |
| ------------------------ | ------------- | ----------------------------------------------------------------------------- |
| `errs/`                  | anyone        | `Public` interface + `PublicMessage(err)` chain-walker                        |
| `internal/api/errors/`   | this app only | `HTTPError` type + `BadRequest`/`NotFound`/`Conflict`/`Internal` constructors |

Import the HTTP errors package with an alias to avoid shadowing the stdlib `errors` package:

```go
import apierrors "github.com/glwbr/paw/internal/api/errors"
```

**Why two packages:** `sefaz/` and `nfce/` are library-shaped (importable, no HTTP dependency). They need the `Public` contract to declare which of their errors are safe to expose. They cannot import `internal/` packages — Go's toolchain forbids it. So `errs/` lives at the top level, zero deps. `HTTPError` is an application concept; libraries must never produce HTTP errors, so it stays under `internal/`.

### Pattern: errors declare themselves safe

```go
// errs/errs.go
type Public interface {
    error
    PublicMessage() string
}
```

Every `sefaz.*` and `nfce.*` error type that is safe to surface implements `PublicMessage()`. Boundary code asks "is this Public?" via one `errors.As` call — no dispatch table, no central registry. The policy lives with the error.

When you introduce a new error type, decide up front: does it contain only information safe to show a user? If yes, implement `PublicMessage()`. If no — raw text, internal addresses, DB details — don't, and the default "internal error" path catches it automatically.

`ParseError` in `nfce/` is a deliberate example of an error that must stay internal: its `RawText` field may contain portal or user data.

### HTTP handlers return `error`

Handlers in `internal/api/` have the signature `func(w, r) error`. A `handle()` adapter wraps each one:

```go
func handle(fn handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil {
            writeError(w, r, err)
        }
    }
}
```

Handlers return `apierrors.BadRequest(...)` / `apierrors.NotFound(...)` for client errors, or bubble domain errors up unwrapped — `writeError` handles the mapping.

### Central error adapter (`writeError`)

Defined in `internal/api/respond.go`. The only place in the codebase that converts an error into an HTTP response:

1. Map known domain sentinels (`api.ErrNotFound`, `api.ErrCaptchaNotPending`) to the appropriate `apierrors.HTTPError` — once, here.
2. If the error is (or wraps) an `*apierrors.HTTPError`: use its status + public message.
3. If any error in the chain is `errs.Public`: respond 400 with that message.
4. Otherwise: log the full wrapping chain server-side, respond `500 {"error":"internal error"}`.

All responses include a `request_id` field that correlates the client response with the server log line. In `APP_ENV=dev` a `detail` field with the full error chain is also included.

Response shape:

```json
{ "error": "import not found", "request_id": "abc123" }
```

### Receipt-import error separation

`ReceiptImport.fail(clientMsg, start, internal)` (in `internal/api/receipt_imports.go`) keeps two channels distinct:

- `clientMsg` → stored on `ReceiptImport.errMsg`, served via `GET /receipts/imports/{id}`. Must be safe. Static strings only.
- `internal` → logged via `slog.Error`. Never touches the wire.

For sefaz fetcher errors, `fetchClientMessage` uses `errs.PublicMessage(err)` to extract the safe message if one exists. For DB errors, the static string `"failed to save receipt"` is the client message.

### Sentinel errors for flow control

`api.ErrNotFound`, `api.ErrCaptchaNotPending` — these are behaviour signals, not user messages. They live as package-level `var`s and are checked with `errors.Is`. Don't use `fmt.Errorf` with runtime data (like an import id) to signal a condition — the caller can't pattern-match on it, and the error string may end up in a log or a response.

### Request ID middleware

Every request gets an `X-Request-ID` header — read from the client if present, generated otherwise. It is stored in the request context and echoed on every response. Correlate a `request_id` from a 500 response body with the corresponding `slog.Error` line to find the full wrapping chain.

## Comments

### Exported identifiers — always doc-commented

Every exported type, function, method, var, and const must have a doc comment — Go tooling (`go doc`, `pkgsite`, `golangci-lint`) uses them. Follow the [Go doc comment](https://go.dev/doc/comment) conventions:

- Start with the identifier name: `// ParseMoney parses Brazilian currency strings…`
- Use complete sentences (capital first letter, period at end).
- Package docs go in a single `// Package X …` comment above `package X`.

### Unexported code — comments only for non-obvious _why_

Internal symbols (lowercase) do not need doc comments. Add one only when it explains something the code itself cannot: a constraint, a trade-off, a platform quirk, or a deliberate omission. Ask _"would a reader who sees the code for the first time wonder why?"_ If yes, comment. If the function name and body already make it obvious, don't.

**Remove:**
- Narrating comments: `// Loop over items`, `// Return result`, `// getImportByID returns the import with the given ID` — these restate the code in prose.
- Unexported function/type docs that only describe _what_, never _why_.

**Keep:**
- Inline annotations for business rules, platform quirks, or deliberate non-obvious choices: `// Lei 12.741`, `// SEFAZ BA returns image bytes with a wrong Content-Type`, `// already exists (ON CONFLICT DO NOTHING)`.
- `//nolint:…` directives with a brief rationale.
- `// TODO:` / `// FIXME:` when tracked work isn't captured elsewhere.

## File summary

| Topic                                 | Canonical location                                                    |
| ------------------------------------- | --------------------------------------------------------------------- |
| Integer value types                   | `nfce/money.go`, `nfce/quantity.go`, `nfce/rate.go`                   |
| Unit canonical codes + fuzzy matching | `nfce/unit.go`                                                        |
| BR-number cleanup helper              | `nfce/internal/parse/text.go` (`BRNumber`)                            |
| sqlc overrides                        | `db/sqlc.yaml`                                                        |
| Public error contract (library)       | `errs/errs.go` (`Public` interface, `PublicMessage`)                  |
| HTTP error types (app-only)           | `internal/api/errors/errors.go` (`HTTPError`, constructors)           |
| Central HTTP error adapter            | `internal/api/respond.go` (`handle`, `writeError`)                    |
| Request ID middleware                 | `internal/api/middleware.go`                                          |
| Receipt-import error separation       | `internal/api/receipt_imports.go` (`fail`, `fetchClientMessage`)      |
