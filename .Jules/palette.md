## 2026-07-10 - Hypermedia for Asynchronous API Discoverability
**Learning:** In headless or backend-only projects, the "User Experience" is primarily the Developer Experience (DX). When dealing with multi-step asynchronous processes (like receipt imports with captchas), hypermedia links (HATEOAS) significantly improve discoverability by guiding the client to the next required action directly from the resource state.
**Action:** Always provide resource-specific hypermedia links in API responses when the resource is in a state that requires further user interaction.
