# PR #4 review: redundancies to slim

Source: [PR #4](https://github.com/yule/beep/pull/4) (`webpush`). Subscribe / settings / test-push path is fine. Fat is the unused send-security layer, plus two subscription sources and two UA parsers on the web client.

Suggested order: drop localStorage → collapse UA/status. First cuts the most lines.

---

## 1. ~~Drop SSRF / DNS~~ — decided: keep as-is

Original concern: create is HTTPS + host allowlist only and **send does not pin IP yet**, yet the PR ships `SsrfProtection`, `resolved_endpoint_ip`, a global `DnsTestHelper`, and unused `net-http-persistent` — ~300+ lines with no current-feature change.

Decision: keep everything verbatim. `SsrfProtection` is copied 1:1 from [fizzy](https://github.com/basecamp/fizzy/blob/main/app/models/ssrf_protection.rb) (same duplicate RFC1918/loopback/link-local ranges alongside the `IPAddr` predicates — that duplication is upstream). Staying in lockstep with upstream beats trimming it here. IP pinning ships with `WebPush::Pool` in Next step 4, where `resolved_endpoint_ip` becomes actually used.

---

## 2. Drop localStorage as a second source of truth

`web-push.ts` stores `{id, endpoint}` under `beep.pushSubscription.{slug}`, and also uses PushManager + the server list.

The UI already marks “This browser” with `record.endpoint === currentEndpoint`. `isSubscribedForAccount` requires **both** a localStorage row **and** a PushManager subscription. If they diverge, the card looks unsubscribed while the device list still shows this browser.

Use only:

- PushManager endpoint
- whether that endpoint is in `GET push_subscriptions`

`disable` / `remove` look up id by endpoint. Then delete `read/store/clearStoredSubscription` and `STORAGE_PREFIX`. `sendTestPush` need not `create` (upsert) before every test.

- [x] Subscribe state = PushManager endpoint ∈ server list
- [x] Remove localStorage helpers
- [x] Test push: call `POST .../test` with the existing row id

---

## 3. One UA / platform model

`WebPushStatus` has `ios`, `macos`, `platform`, and `browserName`.

- `status.macos` has **no callers**
- `needsIosInstall` uses `status.ios` ≡ `status.platform === "ios"`
- `isIosDevice` / `isMacOS` and `notificationPlatform` / `describePushDevice` each scan UA

Keep on status: `supported`, `permission`, `subscribed`, `standalone`, `browserName`, `platform`. Reuse `notificationPlatform` inside `describePushDevice` instead of a second iPhone/Android/Mac/Windows/Linux pass.

iOS copy is duplicated: settings card Home Screen text vs Tips dialog `ios` step.

- [x] Drop `ios` / `macos` from status; use `platform`
- [x] Share one UA → platform/browser helper
- [x] One iOS Home Screen sentence

---

## 4. Thin wrappers

| Place                                                             | Suggestion                                                                                         |
| ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `listPushSubscriptions`                                           | Unpack of `fetchPushSubscriptions`; hook can call the API                                          |
| `getPushPermission`                                               | Marked `async`, just reads `Notification.permission`                                               |
| `hasBrowserPushSubscription`                                      | ≈ `getBrowserPushEndpoint() !== null`                                                              |
| `upsert_push_subscription`                                        | One-liner; call `upsert_for!` in the controller                                                    |
| `useWebPush` enable/disable/sendTest/remove                       | Same `setError` + try/catch + `refresh` + finally → one `run(action)`                              |
| `refresh` `setStatus({ ios, macos, standalone, browserName, … })` | UA fields do not change between refreshes; only update `supported` / `permission` / `subscribed` |

`GET /api/v1/web_push` as its own controller is fine (VAPID is user-scoped, no account slug). Do not merge with subscriptions.

- [x] Delete or inline the wrappers above
- [x] Fold hook mutations into one helper

---

## Leave as-is

- Settings route, sidebar entry, unhashed SW, `web_push` in reserved slugs
- `upsert_for!` SQLite lock comment
- jbuilder partials (repo convention)
