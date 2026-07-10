## 2026-07-10 - [Hypermedia for API discoverability]
**Learning:** In headless or backend-only projects, UX is synonymous with DX (Developer Experience). Providing resource-specific hypermedia links (like `captcha_url`) guides the consumer through multi-step asynchronous processes and improves self-discoverability without requiring the client to hardcode URL patterns.
**Action:** When implementing asynchronous or multi-step API resources, always include hypermedia links to the next logical actions in the resource's current state.
