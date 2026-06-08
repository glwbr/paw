## 2025-05-15 - Improving API Discoverability with Dynamic Links
**Learning:** In backend-only projects, API UX is defined by discoverability. Providing resource-specific hypermedia links (like `captcha_url`) guides consumers through multi-step asynchronous processes without requiring them to hardcode URL patterns.
**Action:** When an API resource enters a state requiring user interaction (like `waiting_captcha`), include a field with the direct URL for the next action.
