---
"@go-ios/sdk": minor
---

Initial release of the go-ios TypeScript SDK: typed client generated from the
OpenAPI 3.1 spec plus an ergonomic facade (`IosClient`) covering device info,
screenshots, activation, pairing, condition inducers, developer disk images,
configuration profiles, accessibility/location reset, location simulation, app
management (`apps`), and WebDriverAgent sessions (`wda`). SSE endpoints
(`syslog`, `notifications`, `ostrace`, `listen`) are exposed as typed async
iterators with heartbeat handling, unknown-event surfacing, and `AbortSignal`
cancellation.
