## 2026-07-10 - API Discoverability and Failure-Fast Validation in Headless Projects

**Learning:** In backend-only (headless) projects, UX is synonymous with DX (Developer Experience). Since there is no visual interface, clarity is provided through actionable error messages and hypermedia links that guide the consumer through multi-step asynchronous processes.

**Action:** When implementing asynchronous operations that require user intervention (like CAPTCHAs), provide explicit links in the resource representation to discoverable endpoints. Implement early validation at the API boundary to catch obvious input errors before they enter the background pipeline.
