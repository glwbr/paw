## 2025-05-14 - [Hypermedia in Headless API]
**Learning:** In backend-only projects, API UX is defined by discoverability; provide resource-specific hypermedia links (like `captcha_url`) to guide consumers through multi-step asynchronous processes.
**Action:** Always include hypermedia links for state transitions in Long Running Operation (LRO) patterns.

## 2025-05-14 - [Failure-Fast API UX]
**Learning:** Immediate validation of identifier-based parameters (like access keys) in async API handlers provides actionable feedback before background processing starts, improving the Developer Experience (DX).
**Action:** Implement regex or format validation for identifiers at the entry point of the API.
