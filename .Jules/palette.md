# Palette's Journal

## 2026-07-05 - Improving API Discoverability with Hypermedia
**Learning:** In headless projects, UX is Developer Experience (DX). Guiding the consumer through asynchronous, multi-step processes (like captcha solving) using hypermedia links (e.g., `captcha_url`) significantly improves the discoverability and usability of the API.
**Action:** Always look for opportunities to provide resource-specific links in API responses to guide the next steps.

## 2026-07-05 - Failure-Fast Validation for Better Feedback
**Learning:** Immediate validation of complex identifiers (like 44-digit NFC-e access keys) in the API handler provides actionable feedback before background processing starts, improving the overall UX/DX.
**Action:** Implement validation at the entry point for critical identifiers using existing domain logic.
