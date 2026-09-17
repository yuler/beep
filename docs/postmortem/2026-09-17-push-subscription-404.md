# Postmortem: Production `Push::Subscription` 404 (2026-09-17)

| Field        | Detail                                                                                            |
|--------------|---------------------------------------------------------------------------------------------------|
| Symptom      | `DELETE /api/v1/:slug/push_subscriptions/:id` and `POST .../:id/test` return 404                   |
| Error        | `Couldn't find Push::Subscription with 'id'="2q1tne…" [WHERE "channels"."kind" = ? AND "channels"."user_id" = ?]` |
| Reporter     | qichenwx (production, beep.yuler.cc, personal account `qichenwx9620`)                              |
| Root cause   | Backfill migration wrote dashed-UUID **TEXT** primary keys via raw SQL, bypassing the app's base36/BLOB(16) UUID codec |
| Blast radius | Every migration-created `channels` row: all email channels + backfilled web_push rows; all find-by-id endpoints on them |
| Data fix     | Migration `20260917140205_fix_channels_text_uuid_ids.rb` (runs automatically on deploy)            |
| Status       | Pending: close after merge + deploy                                                                |

## 1. Timeline

| Time              | Event                                                                                   |
|-------------------|-----------------------------------------------------------------------------------------|
| 2026-08-17        | #4 ships web push with a standalone `push_subscriptions` table (UUID PKs, Rails-made)   |
| 2026-08-20        | The failing subscription row is created in the old table (FCM endpoint, Chrome/Linux)   |
| 2026-09-11        | Migration `20260911110000_backfill_email_and_web_push_channels.rb` merges with #47      |
| 2026-09-15        | #47 ships: web push moves into the `channels` table; old rows are backfilled (new ids, `created_at` preserved) |
| 2026-09-17 morning | User clicks delete/test on the settings page, 404 reproduces consistently, reported     |
| 2026-09-17 afternoon | Root cause found; fix SQL verified by hand on a local copy; converted to a proper migration |

## 2. Root cause

The app's UUID strategy (borrowed from basecamp/fizzy, two pieces working together):

- `core/lib/rails_ext/active_record_type_uuid.rb` — Ruby-side ids are **base36, 25-char, time-ordered**
  strings (`Type::Uuid.generate` transcodes `SecureRandom.uuid_v7`, e.g. `03gvtja60ojxfq685e7rk3xug`),
  stored in SQLite as **16-byte BLOBs**, converted by `serialize`/`deserialize`.
- `core/config/initializers/uuid_primary_keys.rb` — maps uuid columns to `blob(16)` at table creation.

The backfill migration wrote `channels.id` as dashed-UUID **TEXT** via raw SQL
(`lower(hex(randomblob(…)))`), bypassing the codec. Consequences (SQLite is dynamically
typed, `BLOB ≠ TEXT`):

| Operation              | Result                                                                                            |
|------------------------|---------------------------------------------------------------------------------------------------|
| `index` (no id filter) | Fine (this is why the row always showed up in list pages)                                         |
| `user_id` scope        | Fine (`account_id`/`user_id` were copied verbatim as BLOBs from users)                            |
| Reading `id`           | `deserialize` hex-encodes the 36 ASCII chars as if they were binary → a stable **56-char** string (what the API and the error show; the stored value is a plain UUID) |
| `find(id)`             | Bound blob never equals the stored TEXT → `RecordNotFound`, 404                                   |

In one line: **the row is in the database, but Rails can never find it by id**. This has nothing
to do with the "web push scoped under account/user" design — the `kind + user_id` conditions are
both correct.

## 3. Ruled-out hypotheses (so nobody re-walks them)

| Suspect                                     | Verdict                                                        |
|---------------------------------------------|----------------------------------------------------------------|
| Server-side token revocation/expiry         | None. Neither PATs nor channel tokens expire                   |
| `channel disconnect` / `logout` / `unset`   | Code audit: every `SaveFile` call preserves the other fields   |
| Wrong account / identity mismatch           | `/me` shows a single personal account; PAT and session see the same row under the same cookie |
| Migration lost data                         | No. The failing id exists in `channels`, `user_id` preserved   |
| "Production ids are not UUIDs" (mid-debug misread) | **Corrected**: storage holds plain 36-char UUID TEXT; the 56-char form is a read-side mis-decode |

## 4. Blast radius

- Every migration-created `channels` row: each user's email channel + web_push rows backfilled from the old table.
- Hit endpoints: `push_subscriptions` destroy/test, `channels` show/destroy/test.
- Unaffected: Rails-created rows (cli etc., BLOB ids); listing/polling/delivery (never look up by id).
- Frontend `disableWebPush` already swallows destroy 404s, so user-facing impact is limited.

## 5. Fix

Data fix converged into a regular migration (runs automatically on deploy, no manual SQL):

`core/db/migrate/20260917140205_fix_channels_text_uuid_ids.rb`

- `UPDATE channels SET id = x'…'`: restores TEXT to the 16 bytes the codec expects
  (`36^25 > 2^128`, so the base36↔int↔hex round-trip is lossless).
- Also re-points `channel_deliveries` / `channel_authorizations` FKs stored as
  `serialize(misdecoded_id)` garbage blobs back to the new ids (deterministically recomputed
  via the inverse codec; both blob lengths are matched against Binary truncation; anything
  unmatched is left alone with a warning).
- Rows already stored as BLOB(16) are a no-op; `down` is empty (same irreversible precedent as 20260911110000).
- Verified by hand on production-imported data locally with the equivalent SQL: `find`, scoped find,
  and delivery associations all recovered.

## 6. Follow-up action items

| # | Item                                                                                            | Status   |
|---|-------------------------------------------------------------------------------------------------|----------|
| 1 | Merge this PR, deploy, close the loop by clicking test on a web push subscription               | Pending  |
| 2 | Rule: raw-SQL migrations touching uuid PKs must write `BLOB(16)` (`unhex` or Ruby-generated)    | To adopt |
| 3 | Regression test: `find` backfilled rows via models after running migration SQL                  | To add   |
| 4 | Old `push_subscriptions` table is unread by anyone; drop it in a separate migration             | Backlog  |

## Appendix: key evidence

- Production repro under one cookie: `GET index` returns 1 row, `POST .../test` with the same id 404s.
- Local repro: `Channel.find` on a migrated row 404s (before fix) / works (after fix).
- Local storage measurements: migrated `channels.id` rows are `text/36`, Rails-made rows `blob/16`; `users.id` all `blob`.
- Key files: `core/lib/rails_ext/active_record_type_uuid.rb`,
  `core/config/initializers/uuid_primary_keys.rb`,
  `core/db/migrate/20260911110000_backfill_email_and_web_push_channels.rb`,
  `core/app/controllers/api/v1/push_subscriptions_controller.rb`,
  `core/app/controllers/api/v1/push_subscriptions/tests_controller.rb`.
