## 2026-06-27 - Hypermedia links and early validation in headless projects
**Learning:** In backend-only projects, API UX is defined by discoverability and failure-fast validation. Providing hypermedia links (like `captcha_url`) guides consumers through asynchronous processes, while immediate validation of identifiers (like `access_key`) prevents wasted cycles and provides better DX.
**Action:** Always look for multi-step processes where hypermedia can reduce client complexity, and use existing domain logic to validate inputs at the API entry point.
