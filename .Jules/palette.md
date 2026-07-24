## 2026-07-10 - Hypermedia for Headless API Discoverability
**Learning:** In headless or backend-only projects, API UX is defined by discoverability. Providing resource-specific hypermedia links (like `captcha_url`) guides consumers through multi-step asynchronous processes dynamically, reducing client guesswork and hardcoding.
**Action:** When designing asynchronous workflows, always provide self-discoverable hypermedia links in resource representations to indicate the next logical step (e.g., CAPTCHA entry, polling locations).
