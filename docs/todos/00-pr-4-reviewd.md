# PR #4 review: redundancies to slim

Source: [PR #4](https://github.com/yule/beep/pull/4) (`webpush`). Subscribe / settings / test-push path is fine. Fat is the unused send-security layer, plus two subscription sources and two UA parsers on the web client.

Suggested order: SSRF/DNS/pool gem → drop localStorage → collapse UA/status. First two cut the most lines.

---

## 1. Drop SSRF / DNS for this PR

Docs already say: create is HTTPS + host allowlist only; **send does not pin IP yet**. The PR still ships:

- `SsrfProtection` (~100 lines) + `ssrf_protection_test.rb` (~140 lines)
- `Push::Subscription#resolved_endpoint_ip` and four model tests
- Global `DnsTestHelper`; controller `setup` always calls `stub_web_push_dns_resolution` even though create/index/destroy never resolve DNS
- `net-http-persistent` in the Gemfile (for unwritten `WebPush::Pool`)

Test push uses `WebPush.payload_send` and never reads `resolved_endpoint_ip`. Host allowlist already blocks arbitrary URLs. IP pin is Next step 4; add it with the pool later.

This cut is ~300+ lines with no current-feature change.

If `SsrfProtection` is kept later, drop duplicate ranges: `IPAddr#private?` / `#loopback?` / `#link_local?` already cover RFC1918, loopback, and link-local. `DISALLOWED_IP_RANGES` repeats `10/8`, `127/8`, `169.254/16`, `172.16/12`, `192.168/16`. Keep only extras (CGNAT, TEST-NET, multicast, …).

- [ ] Remove `SsrfProtection`, `resolved_endpoint_ip`, DNS test helper, and related tests
- [ ] Remove unused `net-http-persistent` until the pool exists
- [ ] Stop stubbing DNS in subscription controller tests

---

## 2. Drop localStorage as a second source of truth

`web-push.ts` stores `{id, endpoint}` under `beep.pushSubscription.{slug}`, and also uses PushManager + the server list.

The UI already marks “This browser” with `record.endpoint === currentEndpoint`. `isSubscribedForAccount` requires **both** a localStorage row **and** a PushManager subscription. If they diverge, the card looks unsubscribed while the device list still shows this browser.

Use only:

- PushManager endpoint
- whether that endpoint is in `GET push_subscriptions`

`disable` / `remove` look up id by endpoint. Then delete `read/store/clearStoredSubscription` and `STORAGE_PREFIX`. `sendTestPush` need not `create` (upsert) before every test.

- [ ] Subscribe state = PushManager endpoint ∈ server list
- [ ] Remove localStorage helpers
- [ ] Test push: call `POST .../test` with the existing row id

---

## 3. One UA / platform model

`WebPushStatus` has `ios`, `macos`, `platform`, and `browserName`.

- `status.macos` has **no callers**
- `needsIosInstall` uses `status.ios` ≡ `status.platform === "ios"`
- `isIosDevice` / `isMacOS` and `notificationPlatform` / `describePushDevice` each scan UA

Keep on status: `supported`, `permission`, `subscribed`, `standalone`, `browserName`, `platform`. Reuse `notificationPlatform` inside `describePushDevice` instead of a second iPhone/Android/Mac/Windows/Linux pass.

iOS copy is duplicated: settings card Home Screen text vs Tips dialog `ios` step.

- [ ] Drop `ios` / `macos` from status; use `platform`
- [ ] Share one UA → platform/browser helper
- [ ] One iOS Home Screen sentence

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

- [ ] Delete or inline the wrappers above
- [ ] Fold hook mutations into one helper

---

## Leave as-is

- Settings route, sidebar entry, unhashed SW, `web_push` in reserved slugs
- `upsert_for!` SQLite lock comment
- jbuilder partials (repo convention)
