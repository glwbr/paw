## 2026-06-28 - Headless API UX: Discoverability and Fail-Fast Validation
**Learning:** In projects without a frontend, UX is defined by the Developer Experience (DX) of the API. Key pillars include:
1. **Discoverability:** Use hypermedia links (like `captcha_url`) to guide clients through multi-step asynchronous processes (LROs), reducing the need for out-of-band documentation.
2. **Actionable Feedback:** Implement fail-fast validation (e.g., checking NFC-e access keys immediately) to provide instant, helpful error messages rather than failing silently in background goroutines.

**Action:** When working on headless projects, focus on improving the API layer's "interactivity" through rich JSON responses and proactive input validation.
