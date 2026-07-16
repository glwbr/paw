## 2026-07-10 - Hypermedia and Fail-Fast Validation for API UX

**Learning:** In headless or backend-only projects, the "User Interface" is the API itself. Micro-UX improvements for developers (DX) focus on discoverability and clear, immediate feedback. Adding hypermedia links like `captcha_url` guides the consumer through multi-step asynchronous processes without them needing to guess or hardcode the next endpoint. Similarly, validating critical identifiers like NFC-e access keys immediately at the entry point, rather than deep in a background goroutine, provides a much better experience by failing fast with actionable errors.

**Action:** Always look for opportunities to add hypermedia links in API responses to improve resource discoverability. Prioritize immediate validation for complex identifiers at the API layer to provide faster feedback.
