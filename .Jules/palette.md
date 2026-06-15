# Palette's UX Journal - PAW

## 2025-05-15 - API Discoverability via Hypermedia
**Learning:** In backend-only projects, API UX is defined by discoverability. When a process is multi-step (like async imports requiring captcha), providing resource-specific hypermedia links (like `captcha_url`) guides consumers and makes the API feel "self-documenting" and easier to integrate.
**Action:** Always look for opportunities to add hypermedia links in long-running or multi-state API resources to guide the client to the next logical action.
