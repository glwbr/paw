## 2026-07-10 - API Hypermedia for Asynchronous Flows
**Learning:** In headless/backend-only projects, API UX is defined by discoverability. Providing resource-specific hypermedia links (like `captcha_url`) guides consumers through multi-step asynchronous processes without requiring them to hardcode URL patterns.
**Action:** Always include conditional hypermedia links in API responses for state-dependent actions.

## 2026-07-10 - Fail-Fast Validation for DX
**Learning:** Returning immediate validation errors in asynchronous "Accept-then-Process" flows improves the Developer Experience by providing instant feedback on malformed inputs before the resource is even created.
**Action:** Use existing domain validation logic (e.g., `nfce.ParseAccessKey`) in the API layer to fail fast.
