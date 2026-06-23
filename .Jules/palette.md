## 2025-05-14 - Hypermedia links for async processes
**Learning:** In headless/backend-only projects, API UX is defined by discoverability; provide resource-specific hypermedia links (like `captcha_url`) to guide consumers through multi-step asynchronous processes.
**Action:** Always include hypermedia links for state transitions in long-running or multi-step operations.

## 2025-05-14 - Immediate feedback on identifiers
**Learning:** Immediate validation of domain-specific identifiers (like 44-digit NFC-e access keys) in the API layer provides actionable feedback before background processing starts, improving the failure-fast UX.
**Action:** Validate high-integrity identifiers at the entry point using domain logic (e.g., `nfce.ParseAccessKey`).
