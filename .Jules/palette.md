## 2026-06-06 - [API Discoverability]
**Learning:** In backend-only projects, API UX is defined by discoverability; provide resource-specific hypermedia links (like `captcha_url`) to guide consumers through multi-step asynchronous processes.
**Action:** Always check if an async process has intermediate steps that could be exposed as links in the resource representation.

## 2026-06-06 - [Strict validation over normalization]
**Learning:** For identifiers with strict formats (like 44-digit NFC-e keys), early validation with clear error messages is better than silent failure or complex backend normalization.
**Action:** Use regex patterns in API handlers to fast-fail invalid identifier formats.
