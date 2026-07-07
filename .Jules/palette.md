# Palette's UX Journal

This journal tracks critical UX and accessibility learnings for the project.

## 2026-07-06 - [Hypermedia in Headless APIs]
**Learning:** In headless/backend-only projects, API UX is defined by discoverability; providing resource-specific hypermedia links (like `captcha_url`) guides consumers through multi-step asynchronous processes without requiring them to hardcode URL patterns.
**Action:** Use anonymous structs in `MarshalJSON` to conditionally add hypermedia links based on the resource's current state.
