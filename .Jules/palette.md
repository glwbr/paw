## 2025-05-15 - [API Hypermedia for Asynchronous Flows]
**Learning:** In headless/backend-only projects, API UX is defined by discoverability. Providing resource-specific hypermedia links (like `captcha_url`) guides clients through multi-step asynchronous processes and improves self-discoverability of the API.
**Action:** Always include "next step" resource links in API responses for asynchronous operations when the state transitions to a state requiring user intervention.
