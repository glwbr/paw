## 2025-05-15 - [API Hypermedia for Discovery]
**Learning:** In headless or backend-only projects, the "User Experience" is synonymous with "Developer Experience" (DX). Discoverability is key: providing hypermedia links (like `captcha_url`) in API responses guides consumers through multi-step asynchronous processes without them having to hardcode URL patterns.
**Action:** Always look for opportunities to provide "next step" links in LRO (Long Running Operation) status responses.
