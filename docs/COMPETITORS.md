# Competitors

Competitive landscape for Beep: one-shot (and later recurring) reminders with a short message and a fire time, delivered via channels such as Email / Web Push.

Beep today (from product + [`TERMS.md`](TERMS.md)):

- Core loop: **message + time** (`kind: once`, `run_at`)
- Server-side fire (Poller → Deliver), not browser-tab timers
- MVP channels: Email, Web Push
- Multi-tenant Account; recurring (`cron`) planned

This is closer to “send a reminder to future me” than a full todo / GTD app.

---

## Competitor map

| Tier                         | Product                 | URL                                                                                                           | Model                                              | Create UX                         | Delivery                         | Account           | Relation to Beep                                      |
| ---------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- | --------------------------------- | -------------------------------- | ----------------- | ----------------------------------------------------- |
| Closest                      | Xtime.pro               | https://xtime.pro/                                                                                            | Self-created one-shot reminder                     | Text + date + time                | Browser Push, Telegram           | Optional / light  | Same form shape                                       |
| Closest (richer)             | YouGot                  | https://www.yougot.ai/                                                                                        | AI reminder SaaS                                   | Natural language                  | Push, SMS, WhatsApp, Email       | Required          | Same job, heavier product                             |
| Closest (zero account)       | Reme.io                 | https://reme.io/                                                                                              | Email-only reminder                                | Text + `@time` + email            | Email                            | No signup needed  | Shortest create path                                  |
| Same name, different job     | beep.me                 | https://beep.me/                                                                                              | Event discovery + email subscribe                  | Tap event → email                 | Email                            | Email only        | Not self-authored reminders; B2B embed “Remind me”    |
| System baseline              | Apple Reminders         | System app                                                                                                    | Native lists / reminders                           | Form + Siri NL                    | Device push                      | Apple ID          | User mental baseline                                  |
| System baseline              | Google Tasks / Assistant| System / web                                                                                                  | Tasks + voice                                      | Form + voice                      | Device push                      | Google account    | User mental baseline                                  |
| Reminder specialist          | Due                     | https://www.dueapp.com/                                                                                       | Persistent nag reminders + timers                  | Fast time chips, NL-ish typing    | Device push (auto-snooze)        | No cloud account  | Best interaction reference for “can’t miss”           |
| Browser light                | Rune Quick Reminder     | https://rune.codes/tools/productivity/quick-reminder                                                          | Minute-delay, one active reminder                  | Title + minutes from now          | In-tab toast / sound / Push      | None              | Only for “in N minutes”; tab must stay open           |
| Browser light                | Time.now Friendly Reminder | https://time.now/tool/friendly-reminder/                                                                   | Local one-shot / recurring                         | Date + time + repeat              | Browser Push + optional sound    | None (local)      | No server reliability                                 |
| Browser extension            | Best Reminder App       | https://chromewebstore.google.com/detail/best-reminder-app-set-tab/dnpkpjllkijgiiedcbjjkccmhcgoebbf            | Tab / note reminders                               | Quick chips (30m / 1h / 1d / 1w)  | Extension notifications          | None              | Relative-time UX worth copying                        |
| Desktop novelty              | guguFly (咕咕机长)      | https://github.com/pumf/guguFly                                                                               | Local desktop timed / holiday / anniversary        | Task list + schedule              | On-screen plane animation        | Local-first       | Experience experiment, not SaaS                       |
| Full todo suite              | TickTick (滴答清单) 等  | Product sites                                                                                                 | Tasks + calendar + habits                          | Full productivity UI              | App / push / email               | Required          | Overlap on reminders; different product center        |

---

## Interaction patterns

| Dimension              | Common patterns                                      | Beep now                         | What to notice when trying others        |
| ---------------------- | ---------------------------------------------------- | -------------------------------- | ---------------------------------------- |
| Time input             | Split Date + Time; `datetime-local`; NL              | Split Date + Time picker         | Faster than native datetime-local?       |
| Relative time          | `+30m` / `+1h` / `+3h` / tomorrow 9:00 chips         | None yet                         | Due / Chrome extension speed             |
| Natural language       | YouGot NL; Reme `@1 hour`                            | None yet                         | Helpful or noise for Beep?               |
| Default time           | Often +1h or next round hour                         | Default +1h                      | Feels right on first create?             |
| Post-create feedback   | List / toast / confirm page                          | Back to beep list                | “Safe to close the tab” feeling?         |
| Delivery channels      | Email most reliable; Push needs permission; SMS loud | Email + Web Push (planned)       | Still works after closing browser?       |
| Missed reminder        | Due nag; YouGot multi-channel                        | Single fire                      | What if user ignores once?               |
| Account friction       | Zero-account vs full signup                          | Login required                   | Too heavy for first beep?                |

---

## Experience checklist (priority)

### First tier — closest to Beep’s form

| # | Product           | Focus while trying                                                                 |
| - | ----------------- | ---------------------------------------------------------------------------------- |
| 1 | Xtime.pro         | Layout of text + date + time; Push / Telegram permission; works after closing tab? |
| 2 | YouGot            | NL vs form speed; free-tier limits; channel picker (Push / SMS / Email)            |
| 3 | Reme.io           | No-signup path; `@time` syntax; email delay and email design                       |
| 4 | Apple / Google    | System default = subconscious baseline                                             |

### Second tier — learn interaction, don’t copy product

| # | Product              | Focus while trying                                      |
| - | -------------------- | ------------------------------------------------------- |
| 5 | Due                 | Time chips; notification snooze; persistent reminders   |
| 6 | Rune Quick Reminder  | Minimal UI for “in N minutes”                           |
| 7 | beep.me              | How light subscribe feels; naming / positioning contrast|

### Trial notes template

```
Product:
Create “Call mom in 1 hour”: ___ seconds, ___ steps
Time UX: split / combined / NL / chips
Account required: yes / no
Still delivers after closing browser/tab: yes / no
Confidence after create (safe to leave): 1–5
Best moment:
Worst moment:
```

---

## Positioning notes

| Angle                    | Competitors mostly…                         | Beep opportunity                                      |
| ------------------------ | ------------------------------------------- | ----------------------------------------------------- |
| Job-to-be-done           | Personal “remind me later”                  | Same job; keep scope tight vs TickTick                |
| Reliability              | Browser tools fail when tab closed          | Server fire + Email/Push is a real differentiator     |
| Team / tenancy           | Mostly personal tools                       | Account / team shared beeps later                     |
| AI / NL                  | YouGot leads                                | Optional later; don’t block MVP                       |
| Naming                   | beep.me = event subscribe                   | Clarify Beep = self-authored reminder SaaS            |

Short-term UX ideas (not committed):

1. Relative chips: `+30m` `+1h` `+3h` `tomorrow 9:00`
2. Clear success copy: “We’ll remind you Aug 14, 6:00 PM”
3. List shows relative countdown (“in 2 hours”)

---

## Sources

Gathered from public product sites / stores (Aug 2026). Re-check live UX when trying each product — marketing pages drift.
