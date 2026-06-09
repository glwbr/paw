# Palette's UX & Accessibility Journal

This journal tracks critical UX and accessibility learnings discovered during the development of this project.

## 2025-05-15 - API Discoverability and Fast-Fail Validation
**Learning:** In headless/backend-only projects, UX is defined by discoverability and immediate feedback. Providing resource-specific hypermedia links (like `captcha_url`) guides consumers through asynchronous flows, and regex-based fast-fail validation prevents unnecessary background processing for malformed inputs.
**Action:** Always include validation regex in API handlers for well-defined formats (like NFC-e access keys) and use hypermedia links to bridge multi-step processes.
