# Impex Line · Backend specification

Status: 27.11.2026 · written after design round 3 (voice, D199); stack parts rewritten for **Go + PostgreSQL** (D202, user 27.11.2026) · for the Claude Code agent that builds the backend.
Companion file: `backend/schema.sql` (PostgreSQL 16 DDL; the goose baseline migration `0001`, sqlc-friendly; loads clean in one transaction; its views reproduce the golden values of §11.7 on the Hîrbovăț extract).
Stack in one line: **a Go API service (net/http, pgx/v5 + sqlc, goose) on PostgreSQL 16; the Vue/Nuxt online-only PWA is only the front end and calls it over `/api`** (§1, §12).

---

## 0. How to use this document

**Reading order.** §1 scope and stack → §2 glossary → §3 roles → §4 data model with `schema.sql` open beside it → §5 calculations → §6 rules and refusals → §7 API (JSON conventions and Go types in §7.0) → §8 voice → §9 live updates → §10 exports/imports → §11 non-functional + golden tests → §12 Go implementation (stack, module layout, sqlc, worked example, build plan) → §13 open questions. §5, §6 and §11.7 are the parts tests are written against; §1–§11 say *what*, §12 says *how in Go*.

**Source-of-truth files** (all paths from the repo root; read them for detail, this spec wins where it is more precise):

| File | What it decides |
|---|---|
| `design/kit/decisions-current.md` | Every decision `Dnn` cited here. **Later decisions win** over older ones and over spec text (e.g. D200 observer sees no prices; D201 manual cycle may pick the series; D151 week rule replaces D02/D26; D172 whole-save corrections replace D101; D183 egg categories replace D164; D184 guide lay % replaces RealGhid). |
| `design/kit/PRODUCT.md` | Product, users, constraints (online only, Romanian, counts from events). |
| `design/kit/data-model.md` | Logical tables, cycle rules, "calculated, not stored" formulas (round 2 text is current). |
| `design/kit/spec.md` | Functional spec per feature (A1–A2, B1–B4, C1–C9, D1–D6, E1–E2, F1 voice §8 with the 81-intent catalogue). |
| `design/kit/screens.md` | Screens S1–S22, what each shows per role. |
| `design/kit/components.md` | Exact UI rules where the backend must compute what a component shows (EntryCard refusal bounds, StockReadout, CostBlock, GuideImportPreview). |
| `design/kit/open-questions.md` | Unresolved product questions 1–21 (see §13). |
| `design/data/extracts/*.json` + `README.md` + `key-figures-18-11.md` | Real data to 18.11.2026 ("today" on every board, D152); seed and golden values. |
| `design/data/scripts/data03_build.py`, `data03_check.py` | Workbook → data mapping, sale/death split (D155), lay basis (D156/D159), Cat 1–3 (D183), stock pairing; the reference for the one-time migration (§10.3). They stay Python (workbook → extracts, already run); the Go command `cmd/migrate-history` reads their output. |
| `design/canvas/scripts/realdata.py` | Reference implementation of every derived figure the screens show (alive, lay, feed g, stock days, alerts, missing). Python reference only: port the formulas to Go (`internal/calc`), don't call it. |
| `design/canvas/project/*.dc.html` | The boards; exact Romanian copy of every message quoted in §6 was taken from them. |

**Conventions in this document.** "Romanian strings" in `"…"` are returned or shown verbatim (keep diacritics: ă â î ș ț, the comma-below forms). `{x}` are placeholders filled with formatted values (§11.3). Copy marked *(proposed)* has no board yet: keep it, the user reviews it. Figures are the real Hîrbovăț Seria 2026 data at 18.11.2026 unless marked *illustrative*.

---

## 1. Scope and non-goals

**Stack and deployment (D202, user 27.11.2026)**
- **Backend = one Go service** (`cmd/api`, §12) on **PostgreSQL 16**. Every rule of §3–§11 lives in it: auth, permissions, calculations, validation, audit, live events, voice, exports. It is a separate deployable with its own Dockerfile (§11.13) and its own repo directory (`backend/`, a Go module).
- **Front end = Vue/Nuxt, an online-only PWA** (D08). Nuxt only renders and calls the Go API over HTTPS JSON (+ one SSE stream, + one WebSocket for voice). Nuxt has **no** server API routes, no database access and no business rules; whether Nuxt runs SSR or ships a static build is a front-end decision. Client-side form checks are a convenience: the Go API re-validates everything.
- **Same origin (recommended)**: one public origin, e.g. `https://app.example`, behind a reverse proxy (Caddy, nginx or a managed load balancer): `/api/*` → Go, everything else → Nuxt. The `il_session` cookie is then first-party, `SameSite=Lax` works, no CORS is needed, the CSP stays `'self'`. The proxy must: pass `/api/live` unbuffered and without a short read timeout (SSE, §9), pass the WebSocket upgrade on `/api/voice/stream` (§8.2), forward `X-Forwarded-For`/`X-Forwarded-Proto` (the Go side trusts them only from `TRUSTED_PROXIES`, §11.10), and keep request bodies ≤ 2 MB.
- **If the API must live on another origin** (e.g. `https://api.example` beside `https://app.example`): it must be the **same site** (same registrable domain), otherwise browsers treat `il_session` as a third-party cookie and drop it; a different site is not supported. CORS on the Go side, for the configured `APP_ORIGIN` only: `Access-Control-Allow-Origin: <APP_ORIGIN>` (never `*`), `Access-Control-Allow-Credentials: true`, `Vary: Origin`, allowed methods `GET, POST, PUT, PATCH, DELETE`, allowed headers `Content-Type, Last-Event-ID, X-Voice-Run`, preflight `Access-Control-Max-Age: 600`; preflights from any other origin get no CORS headers. The front end sends `credentials: 'include'` and opens SSE with `new EventSource(url, { withCredentials: true })`. The cookie stays host-only on the API host (no `Domain=`), `SameSite=Lax`. The Origin checks of §11.4 compare against `APP_ORIGIN` exactly as on one origin.

**In scope (MVP)**
- Sign-in, users, farm access (A1, A2, S1, S22).
- Setup: farms, sections (halls/compartments), guides with import, feed types, series with placement, cycles (B1–B4, S18–S21).
- Entry: Zilnic (mortality, culls with reason, sexing errors, eggs: total via Cat. 1–3 + Incubabile, Jos, egg weight), Hrană zilnică (g/bird per section × sex), Cântărire, Transfer, Vânzare, Livrare hrană, Prețuri hrană (C1–C7), Introdus azi, Lipsesc date (C8, C9).
- Views: series, farm (cycle), section pages with tabs Rezumat · Păsări · Hrană · Greutate · Producție, each table-first (PeriodTable), reports Producție and Hrană (Stoc · Livrări · Consum · Cost), Acasă per role (D1–D4, S2–S5, S11, S14).
- Corrections of past data with Motiv and the change log Istoric modificări (E1, E2, S16, S17).
- Export of every table to .xlsx (D5, D191).
- Voice commands (F1, D199): navigate/filter/answer at once, prefill forms, never save.
- Live updates of open pages (D197).
- Egg categories Cat. 1, Cat. 2, Cat. 3 as counts per hall per day, shown as % of eggs in reports (D183, D186).
- One-time developer migration of the history (D09) from the two workbooks via the extracts.

**Out of scope** (do not build tables, endpoints or flags for them)
- Offline entry, background sync, service-worker data caches (D08; PWA install only).
- Notifications of any kind (push, e-mail, Telegram) — alerts are shown in the app only.
- Incubator: `Incubare` batches and `BD Incubator` history (D182).
- An import screen for history (D09) and for anything but guides.
- Voice: wake word / always-on listening, any language but Romanian, voice before sign-in, speech synthesis (spec §7, §8).
- Compare cycles (D6 "later"), egg/hen and hatchability guide lines (D184).
- Self-service password reset (A1).

---

## 2. Glossary (Romanian ↔ English ↔ code)

| Romanian (UI) | English | Code / column |
|---|---|---|
| fermă | farm | `farm`; `type` growth/laying |
| creștere / ouat | growth / laying (farm and guide type) | `farm_type` `growth`/`laying` |
| hală | hall (holds compartments; eggs are counted per hall) | `section` with children |
| secțiune / compartiment | section / compartment (holds birds) | `section` without children (`H5.1`, Telița `H1`, `H5-1`) |
| serie (Seria 2026) | series = one yearly flock | `series` |
| ciclu (deschis / închis) | cycle = one stay of a series in a farm (open / closed) | `cycle` |
| Începe ciclu nou | start a new cycle manually | `cycle.started_by = manual` |
| populare | placement (birds put into the growth farm) | `bird_event.type = placement` |
| transfer / Mută păsări | transfer between sections (any farms) | `transfer` (2 rows per sex) |
| vânzare / golire | sale / emptying (emptying is a sale, D155) | `sale` |
| mortalitate / morți | deaths | `mortality` |
| sacrificare (motiv: boală, șchiopătare, altul) | cull with reason | `cull`, `cull_reason` |
| eroare sexare | sexing error (signed per sex) | `sexing_error` |
| diferență registru | register difference (migration only) | `adjustment` |
| în viață / vii | alive | derived (`v_section_sex_day`) |
| la începutul zilei | at the start of the day (= end of the previous day) | `alive_start` |
| găini ♀ / cocoși ♂ | hens / roosters | `sex` `F` / `M` |
| hrană zilnică, g/pasăre | daily feed, grams per bird | `feed_allocation.g_per_bird` |
| livrare hrană, kg, preț lei/kg, total lei, fără preț | feed delivery, kg, price, invoice total, unpriced | `feed_delivery` |
| stoc, zile rămase, se termină | stock, days left, run-out date | derived (§5.9) |
| tip de hrană (Breeder 2, Cocoși, Starter…) | feed type | `feed_type` |
| cântărire, masă medie, CV, uniformitate, mărime probă | weighing, mean weight g, CV %, uniformity %, sample size | `weighing` |
| ouă (total), incubabile, jos, Cat. 1 / 2 / 3, masa medie | eggs (total = Cat. 1+2+3 + incubabile), hatching eggs, floor eggs (apart), categories, mean egg weight | `egg_production` |
| % ouat | lay % | §5.5 |
| ghid | guide (targets by age week and sex) | `guide`, `guide_week` |
| SAPT / sapt. | the birds' age week | §5.1 |
| Vârstă (zile), vârstă la populare | age in days, age at placement | §5.1, `series.age_at_placement_days` |
| în normă / abatere / refuzat / informație / lipsă / oprit | status words: in norm / deviation / refused / info / missing / off | §5.12, §6 |
| Motiv | reason (≤ 200 characters) | `audit_log.reason` |
| Corectează / corectat / modificat de | correct / corrected / changed by | §7.9 |
| Istoric modificări | change log | `audit_log`, §7.10 |
| Introdus azi | entered today | §7.8 |
| Lipsesc date | missing entries | §5.11 |
| De verificat | to check (alerts) | §5.12 |
| Proprietar / Manager / Veterinar (Medic veterinar) / Observator | owner / manager / vet / viewer | `user_role` |
| din voce | filled from voice (VoiceMark) | §8 |

---

## 3. Roles and permissions

### 3.1 Scope
- **owner** sees and acts on every farm, including farms added later; has no `user_farm` rows.
- **vet, viewer, manager** see and act only on farms in `user_farm` (D03). Every endpoint filters by it; a request naming another farm gets `403 FARM_FORBIDDEN` "Nu ai acces la această fermă." (VO-6, S4 fără acces). A user with no farm gets the S2 state "Nu ai acces la nicio fermă."
- A **transfer** touches two farms: the manager needs both.
- **Series pages** show only the cycles on the user's farms.
- **Prices** (lei/kg, total lei, Cost, lei columns, price endpoints): owner read/write; manager read-only; **viewer (Observator) none anywhere; vet none** (D200, user 26.11.2026, supersedes D91's observer-price part; D06, D64, spec §1 table, the S11 Observator board). The *count* "n livrări fără preț" is not a price and is shown to the owner (Acasă) and on S11 to manager and viewer.
- **Closed series** are read-only for everyone ("nu mai primește înregistrări sau corecții", S18 Închide seria Dialog): writes dated inside a cycle of a closed series → `409 SERIES_CLOSED` *(proposed)* "Seria 2025 e închisă: nu mai primește înregistrări sau corecții."

### 3.2 Time windows (all on the Europe/Chisinau day, §11.1)
| Who | Create | Edit / delete without Motiv | Edit / delete with Motiv (Corectează / Șterge) |
|---|---|---|---|
| vet | only `date = today` (D07, spec C1 "Pentru zilele trecute cere managerului.") | own entries while `today = local date of created_at` (until 23:59, D07); logged with `same_day_author=true`, no reason (D70 A) | never (D71: no action on others' entries; own past entries show "blocat") |
| manager | any `date ≤ today` on own farms (past days: filling missing rows, D65) | own entries the same creation day (D70 B) | any entry on own farms, any past date, Motiv required (E1); someone else's entry from today also goes through Corectează (D70 B) |
| owner | any `date ≤ today` | own entries the same creation day | any entry, Motiv required |
| viewer | – | – | – |

- Future dates are refused for every role: "{date} e în viitor. Alege o dată până la {today}." (S17/S18 boards; also the D37 form "Alege o dată până la 18.11.2026.").
- Undo on a save Toast ("Anulează", D84/D108/D122) = the author deletes or restores within the same window; logged `action=undo`.
- Price undo (S12 Toast) returns the priced deliveries to "fără preț" the same day (D122); later price changes need a Motiv (C6, D69).

### 3.3 Permission matrix (✓ allowed · – refused with 403 · ○ read-only)
| Resource · action | vet | viewer | manager | owner | Notes |
|---|---|---|---|---|---|
| Acasă (S2) | ✓ vet layout | ✓ viewer layout (D89) | ✓ manager layout | ✓ owner layout | §7.6 |
| Zilnic, Hrană zilnică, Cântărire: create | ✓ today | – | ✓ | ✓ | §3.2 |
| Transfer, Vânzare, Livrare hrană: create | – | – | ✓ | ✓ | C3–C5 |
| Livrare: price set / change | – | – | ○ (sees, cannot set) | ✓ | C6, D69 |
| Introdus azi (S15) | ✓ | – | ✓ | ✓ | the caller's own entries |
| Corectează / Șterge with Motiv (S16) | – | – | ✓ | ✓ | E1, D70 |
| Scope pages S3–S5, tabs, tables | – (D64: vets never reach them) | ✓ | ✓ | ✓ | vets read only through entry forms and Acasă |
| Rapoarte Producție (S14) | – | ✓ | ✓ | ✓ | |
| Rapoarte Hrană (S11): Stoc, Livrări (kg), Consum | – | ✓ | ✓ | ✓ | |
| Rapoarte Hrană: lei, Cost block, price endpoints | – | – | ○ | ✓ | D64, D200 |
| Prețuri hrană (S12) | – | – | – | ✓ | C6 |
| Export .xlsx | – | ✓ (no lei columns) | ✓ | ✓ | where Exportă exists |
| Istoric modificări (S17) | – "Nu ai acces la Istoric modificări." | – same | ✓ rows touching own farms | ✓ all | D104 |
| "modificat de …" mark on a record | plain text | plain text | link to S17 | link | D105, D178 |
| Începe ciclu nou, Schimbă ghidul ciclului | – | – | ✓ | ✓ | B4 |
| Serii: list, Serie nouă, age change, Închide seria | – | – | – | ✓ | D107; "Nu ai acces la Serii." |
| Ferme și secțiuni, Ghiduri, Tipuri de hrană, Utilizatori (write) | – | – | – | ✓ | "Nu ai acces la Ferme și secțiuni." / "… la Ghiduri." / "… la Tipuri de hrană." / "… la Utilizatori." |
| Voice commands | ✓ within the rows above | ✓ | ✓ | ✓ | §8: voice never adds a power |

### 3.4 Voice limits
Voice resolves to the same permission checks (§8.6). In addition, voice **never**: saves, deletes, deactivates, reactivates, retires, closes a series, resets/sets/reveals a password, undoes a saved record, opens or confirms a confirm Dialog, signs in; it never shows prices to vet/viewer or attaches a price to a manager's delivery ("Prețul îl adaugă proprietarul în Prețuri hrană.").

---

## 4. Data model

The DDL is `backend/schema.sql`. This section explains every table, the choices behind it and the rules the DDL cannot express (enforced in the service layer).

### 4.1 Overview
```
app_user ─< user_farm >─ farm ─< section (hall ─< compartment)
                          │  └─ guide_id → guide ─< guide_week (week × sex)
series ─< cycle >─ farm      cycle.guide_id → guide
entry (one save) ─< bird_event · egg_production · feed_allocation · weighing · feed_delivery · feed_stock_adjustment · legacy_mortality
feed_type ─< feed_delivery, feed_allocation
audit_log (one row per save action) · live_event (SSE outbox) · request_replay (idempotent replies) · voice_usage · app_setting · user_session · login_attempt
```
IDs are UUID v4 (`gen_random_uuid()`), so the client can create an idempotency key before saving.

### 4.2 Counts are derived, never stored (D01)
- Birds alive for a section, sex and day = Σ `bird_event.qty` of live rows up to and including that day (`v_section_sex_day.alive_end`). Start of day = end of the previous day (`alive_start`).
- Hall / farm / series totals = sums of their sections (D156: never a register "Total" row).
- `bird_event.qty` is signed: placement +; mortality, cull, sale −; transfer −out/+in; sexing_error ± as recorded; adjustment ± (migration only).
- Optional cache (not needed at this scale: ~13 sections × 2 sexes × 400 days ≈ 10 k rows; the dense view answers in ~25 ms): a table `section_day_snapshot(section_id, sex, date, alive_end, …)` may be added later; it must be rebuilt for `(section, sex, date ≥ D)` in the same transaction as any write dated D, and a nightly job compares it with the view and alerts on drift. Never read a cache for validation.

### 4.3 Users and access
| Table | Key columns | Why |
|---|---|---|
| `app_user` | `login` unique `[a-z0-9.]{3,40}`; `name` ≤ 60; `role`; `active`; `password_hash` argon2id; `password_generated_at/by`; `last_login_at` | A2, D78, D118. Users are never deleted (entries point to them). Login is read-only after creation (D118). |
| `user_farm` | (user_id, farm_id) | D03. Owners have no rows. |
| `user_session` | `token_hash` (sha256 of a 32-byte random cookie), `expires_at` (sliding 30 days), `revoked_at` | A1 "session lasts 30 days on that device". Revoked on logout, password reset, deactivation, and **any change of the user's role or farms (all of that user's sessions are revoked; the user signs in again)**. |
| `login_attempt` | login, ip, at, success | A1: 5 failures → 5 min wait. |

**One-time passwords** (D78, D118, D119) need no table: the server generates the password, stores only its argon2id hash, returns the plain text once in the create/reset response, and never again. There is no forced change at first sign-in (A1 has none).

### 4.4 Farms and sections
- `farm`: `name` unique ignoring case and diacritics (`name_key` = lower-case, diacritics removed: "Hirbovat" collides with "Hîrbovăț", D111). `type` growth/laying; changeable only while the farm has no open cycle (D110). `guide_id` nullable ("fără ghid"), composite FK `(guide_id, type) → guide(id, type)` enforces "same type" (B1). A guide change applies to the **next** cycle only (B1).
- `section`: `code` (`H1`, `H5-1`, `H5`, `H5.1`), unique per farm among active sections, case-insensitive. `parent_id` = the hall (one level only, D75; parent in the same farm via composite FK). `sort_order` for reorder among siblings (D75). `retired_at`: retire never deletes, no reactivation (D75; "O secțiune retrasă nu se mai poate reactiva.").
- **Roles of a section** (derived, `v_section_label`): a section **with children is a hall**: it never holds birds, it carries the hall's eggs; a section **without children holds birds**. **Egg unit** of a laying farm = the hall, or the section itself when it has neither parent nor children (`v_egg_unit`).
- Real codes (D152): Telița `H1 H2 H3 H6 H5-1 H5-2` (no halls); Hîrbovăț halls `H5 H6 H7` with compartments `H5.1 H5.2 H6.1 H6.2 H7.1 H7.2` (old codes H1/H3/H4 map by position).

### 4.5 Guides
- `guide(name, name_key unique, type)`; `guide_week(guide_id, week, sex)` PK = "duplicate week + sex refused" (B2). Columns: `feed_g`, `body_weight_g` (weight at the **end** of that week, week 1 = day 7, D157), `mortality_cum_pct` (cumulative; laying guide counts from the start of lay, D161), `lay_pct` (laying ♀ only; Productivitate col I, D184), `hatching_pct` (% incubabile of all eggs, not hatchability, D153), `egg_weight_g`. Null = "– fără date în ghid" (D29).
- `week` = the birds' **age week** SAPT, no offset (D151): "Ross 308 ouat" SAPT 19–65 (lay % 23–62), "Ross 308 creștere" SAPT 1–25 (guide names illustrative, D22).
- **Replace** (Înlocuiește ghidul, D113) replaces all rows in place; every cycle using the guide compares against the new rows from then on; type stays fixed while used. **Delete** only when no cycle uses it ("Folosit de 2 cicluri: îl poți înlocui, nu șterge.").

### 4.6 Series and cycles
- `series(name unique, year unique, status, placement_date, placement_farm_id, placement_inferred, age_at_placement_days 0–140)`. `placement_date` = the date of the first placement (stored because the history has an inferred one: Seria 2025 = 01.11.2024, D36). A closed series is read-only.
- `cycle(series_id, farm_id, guide_id, start_date, end_date, started_by auto|manual|migration, start_inferred, ended_by emptied|manual|migration, start_reason)`.
  - DB guarantees: one open cycle per farm (partial unique index); no two cycles of a farm overlap (`exclude using gist` on `daterange(start,end,'[]')`).
  - A dated record of a farm belongs to the cycle whose `[start_date, end_date]` contains its date (open cycle: up to today). Records do **not** store `cycle_id`: when a correction moves a boundary, the records follow automatically.
  - `guide_id` is copied from `farm.guide_id` when the cycle opens; "Schimbă ghidul ciclului" changes it for the whole cycle ("applies from the cycle start", B4) and is audited.
- The week belongs to the birds (series), not to the cycle: a new cycle never restarts SAPT (D151).
- Auto open/close is maintained by `reconcileCycles` (§6.4) after every write that moves birds.

### 4.7 Entries: one save = one record (D168, D172, D178)
`entry` is the header of one form save. Its figures are child rows pointing to it:

| `entry.kind` | Form | Child rows |
|---|---|---|
| `daily` | Zilnic (S6) for one farm and day; any subset of its sections | `bird_event` (mortality, cull + reason, sexing_error) per section × sex; `egg_production` per hall |
| `feed_daily` | Hrană zilnică (S7) | `feed_allocation` per section × sex (only confirmed rows, D13) |
| `weighing` | Cântărire (S8) | one `weighing` |
| `transfer` | Transfer (S9): one Din, one În, ♀ and ♂ | 2 `bird_event` rows per sex with a count (−n at Din, +n at În, `peer_section_id` = the other side); `farm_id` = Din farm, `peer_farm_id` = În farm |
| `sale` | Vânzare (S10): one farm, rows per section, buyer, note | `bird_event` type `sale` per section × sex |
| `delivery` | Livrare hrană (S13) | one `feed_delivery` |
| `placement` | Serie nouă (S18) placement rows | `bird_event` type `placement` per section × sex; `series_id` set |
| `migration` | one-time import | anything |

- **Uniqueness of daily values**: at most one live value per (section, day, sex, type) for mortality/cull/sexing_error (`bird_event_daily_uq`), per (egg unit, day) for eggs, per (section, day, sex) for feed and weighings. A second create for a key that already has a live value is refused with `409 ALREADY_ENTERED` + the owning `entryId`, and the client opens that entry in Editează or Corectează (spec C1 "A second save replaces the earlier entry"; D77 "deja introdus"). Two people can therefore save different sections of the same farm-day as two entries.
- **Empty is null, never 0**: a field left empty creates no row; "Fără pierderi azi" creates rows with 0 (D53).
- **Soft delete**: `entry.deleted_at/by/delete_reason`; the same transaction sets `deleted_at` on every child row. An edit that removes a value soft-deletes its child row; a changed value is updated in place; a new value inserts a row. Old/new values live in `audit_log` (§4.10).
- `entry.source = 'voice'` when the save still carried at least one "din voce" mark (D199 note); `client_request_id` = the form draft's idempotency key (§7.0).
- `entry.corrected = true` after a Corectează (Tag "corectat"; EventRow "modificat de Ion Rusu · 18.11.2026, 10:12").

### 4.8 Birds, eggs, feed, weighings
- `bird_event`: `(entry_id, farm_id, section_id, date, sex, type, qty, cull_reason, peer_section_id, note, deleted_at)`. Only sections without children; never on a retired section. `cull_reason` required when a cull qty ≠ 0 (stored ASCII `boala|schiopatare|altul`, shown "boală / șchiopătare / altul").
- `egg_production`: per egg unit and day: `floor_eggs` (Jos), `cat1`, `cat2`, `cat3`, `hatching_eggs` (Incubabile), `avg_egg_weight_g`; `total_eggs` is a generated column = Cat. 1 + Cat. 2 + Cat. 3 + Incubabile, null until all four are filled (D18, D53, D183). Only at laying farms.
- `feed_type(name unique ignoring case/diacritics, for_sex nullable = ambele sexe, active)`: global list (D116). For-sex fixed once used; retire refused while any farm has stock; delete only when unused.
- `feed_delivery(farm, feed_type, date, qty_kg > 0, note, price_per_kg, total_lei, price_input, priced_by, priced_at, price_from_register)`: the manager enters type/date/kg (no price, C5); the owner adds lei/kg **or** total lei and the other is derived (C6). Both price columns null = "fără preț".
- `feed_stock_adjustment`: migration only (the workbook's −300 kg and the 3 000 kg ♀→♂ move, DATA03).
- `feed_allocation(section, date, sex, feed_type, g_per_bird 1–300, kg_recorded)`: kg is **derived** (§5.8) unless `kg_recorded` (migration: the register's typed kg, D163) is set.
- `weighing(section, date, sex, avg_weight_g, cv_pct, uniformity_pct, sample_size)`; CV, uniformity, sample optional (D95).
- `legacy_mortality`: deaths per compartment-day of past cycles that have no flock count (Seria 2025, D54), display only; `emptying` rows are never shown as mortality (D52/D92).

### 4.9 Indexes and why
All per-farm reads filter `deleted_at is null` and a date range: partial indexes `(farm_id, date)` on every child table, `(section_id, sex, date)` on `bird_event` for the running count, `(created_by, created_at)` on `entry` for Introdus azi, GIN on `audit_log.farm_ids` for the manager's Istoric filter, `(at)` on `live_event` for SSE replay.

### 4.10 Audit log (E1, E2, D70, D172, D178, D199)
One row per **save action** (not per field):
| Column | Content |
|---|---|
| `action` | `create`, `update`, `delete`, `undo`, `restore`, `price_set`, `price_change`, `cycle_*`, `series_*`, `setup_change`, `user_change`, `password_reset`, `user_deactivate`, `user_reactivate` |
| `entity`, `entity_id`, `entry_id`, `entry_kind` | what changed |
| `record_types` | the S17 chips this row matches: `mortalitate`, `sacrificare`, `eroare_sexare`, `transfer`, `populare`, `vanzare`, `hrana`, `livrare`, `pret`, `cantarire`, `oua`, `ciclu`, `setari` (D178: a move between sections counts as Transfer, a priced delivery as Preț hrană) |
| `farm_ids` | every farm the change touches (transfer: both) |
| `record_date` | the record's day |
| `reason` | Motiv; null only for same-day author edits (D70 A) and system rows (auto cycle open/close) |
| `same_day_author` | "editat"/"șters" before 23:59 by the author |
| `summary` | the Jurnal name of the save: "Mortalitate · 4 secțiuni" (D178) |
| `changes` | `[{ "key": "H5.2 · găini moarte", "section": "H5.2", "sex": "F", "field": "mortality", "old": 1, "new": 3 }]` — the old → new pairs Istoric shows |
| `old_values` / `new_values` | full JSON snapshot of the save (child rows) before/after |
| `source`, `voice_transcript` | `voice` + the command text only when the saved form still carried a "din voce" mark (D199) |

Rules: every write is logged, creates included (needed for "who entered what" and for voice transcripts). Istoric (§7.10) lists every action except plain `create` from the forms, but it does list a `create` that carries a voice transcript (the first-use notice promises the text "rămâne în Istoric modificări") *(assumption, open question 18)*. Price set and price change are logged (C6). Auto cycle open/close are logged with `user_id = null`. The table is append-only (revoke UPDATE/DELETE from the app role).

### 4.11 Live outbox, voice usage, settings
- `live_event(id bigserial, at, farm_ids[], type, payload)`: written in the same transaction as the change together with `select pg_notify('live_event', id::text)` (PostgreSQL delivers a notification only when the transaction commits, and drops it on rollback); the SSE handler replays from `Last-Event-ID` (24 h retention). §9.
- `voice_usage(user_id, day, commands, audio_seconds, llm_tokens_in/out)`: cost tracking without text (§8.9).
- `app_setting`: D19 thresholds and limits (seeded in `schema.sql`); code reads them once per request.
- `request_replay(client_request_id pk, user_id, route, status, body bytea, created_at)`: the exact response bytes of every create that carried a `clientRequestId`, written in the save's transaction; a repeat returns them unchanged (§7.0, §12.5). `bytea`, not `jsonb`, so key order and spacing never drift. Purged after 7 days by the maintenance role.

### 4.12 Views in `schema.sql`
| View | Columns (main) | Used for |
|---|---|---|
| `v_cycle` | cycle + series name, `age_days_start`, `sapt_start`, `sapt_end_or_today`, `is_open` | cycle lists, track, "SAPT 20–59" spans |
| `v_bird_event` | live events + `count_removed` | event log (Jurnal) |
| `v_section_sex_delta` | per section × sex × day: placed, deaths, culls, sexing_error, transfer_in, transfer_out, sold, adjustment, delta | daily rows |
| `v_section_sex_day` | dense days (first event → max(last event, today)) with the above + `alive_start`, `alive_end` | every count, the never-below-0 check |
| `v_farm_sex_day` | farm sums | farm tables |
| `v_egg_unit`, `v_egg_day` | eggs per unit/day + `hens_start`, `lay_pct` | Producție |
| `v_feed_consumption` | allocation + `birds_start`, `kg` | Hrană, stock |
| `v_feed_stock` | per farm × feed type × movement day: delivered, consumed, adjusted, `stock_end` | Stoc, days left |
| `v_section_label` | code, hall code, is_hall, retired | labels and audit keys |
SQL functions: `app_today()`, `age_days(series, date)`, `sapt(series, date)`.

---

## 5. Derived calculations

Notation: `d` a local day; `s` a section; `x` a sex; `scope` = farm, hall, section or series; Σ over live rows only. Every formula below is computed **server-side** from full-precision values and rounded **once, at the end**, half away from zero (D82; §5.15). Reference values are real (Hîrbovăț, Seria 2026) unless marked.

### 5.1 Age, week, week boundaries (D151, D159)
```
age_days(series, d) = series.age_at_placement_days + (d − series.placement_date)      -- integer days, 0 on placement day
SAPT(series, d)     = floor(age_days / 7) + 1                                          -- 13 d → 2, 14 d → 3
week_first_day(k)   = series.placement_date + 7·(k − 1) − age_at_placement_days
week_last_day(k)    = week_first_day(k) + 6
```
- Seria 2026: placed 07.10.2025 (a Tuesday), age 0 → weeks run Tue–Mon; 18.11.2026 = age **407**, **SAPT 59**; SAPT 59 = 17.11–23.11.2026; 24.02.2026 = age 140, SAPT 21; Hîrbovăț cycle start 22.02.2026 = age 138, SAPT 20.
- The table column "Vârstă (zile)" = `age_days` (407), not the workbook's Zile (408 = age_days + 1, D159).
- The week of a page/scope = the week of the series of the cycle shown. On the series page every row uses the series' week. A record's SAPT never depends on the farm or the cycle (transfers keep age; a new cycle doesn't restart it).
- Weekday labels in day rows: `Lu Ma Mi Jo Vi Sâ Du` ("Ma 17.11 · 406 · 59").
- **Partial week**: a week row covers only the days that are inside the period **and** the cycle **and** ≤ today; it carries `daysInRow` and the label "2 din 7 zile" (SAPT 59 on 18.11.2026; Hîrbovăț's first week SAPT 20 holds 22.02–23.02 = 2 din 7). Sums cover those days; rates use those days only.
- **Age change** (B3, owner): changing `age_at_placement_days` from a to b gives `age_days'(d) = age_days(d) + (b − a)` and `SAPT'(d) = floor(age_days'(d) / 7) + 1` for every day of the series (Seria 2026, 0 → 7: today age 407 → 414, SAPT 59 → 60); every comparison is re-read from the guides by the new SAPT. Nothing is stored per week, so nothing else changes.

### 5.2 Periods (FilterBar, D191, D192, D45)
| Period | Weeks/days covered |
|---|---|
| `cycle` "Ciclul curent" (default on S4, S5, S11, S14) | SAPT(cycle.start) … SAPT(min(cycle.end, today)); days clipped to the cycle |
| `cycle` on a past cycle ("Tot ciclul", D52) | the whole closed cycle |
| `weeks:n` ("4 săpt.", "8 săpt.", "13 săpt.") | the current SAPT and the n − 1 before it, clipped to the cycle ("4 săpt." on 18.11.2026 = SAPT 56–59) |
| `interval:from,to` | days from..to (to ≤ today, else "Până la {to} e în viitor. Alege o dată până la {today}."); weeks touched; partial weeks clipped |
| series page (no FilterBar, D63) | the whole series: SAPT of the first placement … today/last day |
- Every table opens on the **last 4 weeks** of the period; older weeks load 4 at a time (§7.12).
- The record-type chips/column groups of a period = the record types that have at least one live record in the **whole period** (D45, D192 note), so columns don't change while weeks load.

### 5.3 Birds alive (D01, D53, D162)
```
alive_end(s, x, d)   = Σ bird_event.qty  (section s, sex x, date ≤ d)            -- v_section_sex_day.alive_end
alive_start(s, x, d) = alive_end(s, x, d − 1)
alive(scope)         = Σ over the sections of the scope
```
- Day row "În viață" = `alive_end` of that day. Week row = `alive_end` of the week row's last day. Foot total = `alive_end` of the latest loaded day.
- Entry forms show "în viață la începutul zilei" = `alive_start` (D53, D181); "Toate" in Transfer/Vânzare fills the current count = `alive_start + today's movements already saved`.
- 18.11.2026: start ♀ 14 395 · ♂ 1 083 → end ♀ 12 561 · ♂ 794 (H5.1 4 263 · 342, H5.2 4 281 · 323, H6.1 3 542 · 124, H6.2 475 · 5, H7 0 · 0).
- A section emptied in the period shows `–` after its last day with the group head "goală din 14.11" (H7, D158/D161); `empty_since(scope)` = the last day its count went from > 0 to 0 and stayed 0.
- **Sexes present** in a section on d = sexes with `alive_start > 0` (or a movement that day). Forms show only those ("H1 are doar găini.").

### 5.4 Deaths, culls, sales, transfers (D155, D165)
| Column | Day value | Week / total |
|---|---|---|
| Morți (Mortalitate) | Σ −qty of `mortality` | Σ days |
| Sacrificare | Σ −qty of `cull` (reason in the event log) | Σ |
| Eroare sexare | Σ qty of `sexing_error`, signed as recorded | Σ |
| Transfer | Σ qty of `transfer`, signed (+ in, − out); on the farm scope internal moves net to 0 | Σ |
| Populare | Σ qty of `placement` | Σ |
| Vânzare / vândute | Σ −qty of `sale` | Σ |
| Rezumat "Morți" | mortality **+ culls** (as the register counts them, D165) | Σ |
- **Sales are never mortality** (D155). Mortality % excludes sales. Alive falls by both.
- Migration rule only (not an app rule): Hîrbovăț compartment-days from 12.11.2026 with ≥ 40 birds booked as mortality are sales (DATA03): 12–18.11 ♀ 12 801 · ♂ 1 182 sold; 18.11 real deaths ♀ 3 · ♂ 0; sold ♀ 1 831 · ♂ 289.

### 5.5 Cumulative mortality % (D161, D162)
```
base(scope, x, cycle)   = Σ birds that ENTERED the scope from outside it during the cycle
                           (placement + transfer in whose peer section is outside the scope), up to d
deaths(scope, x, d)     = Σ mortality + Σ culls of the scope in [cycle.start, d]        -- real deaths only, never sales
cum_mort_pct(scope,x,d) = 100 × deaths / base          (null when base = 0)
```
- Hîrbovăț to 18.11.2026: ♀ 1 521 ÷ 26 883 = **5,66 %**, ♂ 1 048 ÷ 3 024 = **34,66 %** (cycle deaths from arrival; the culls of the real data are inside these deaths).
- Shown on Păsări week rows and day rows (group "% cumulat", one decimal "5,7%") and as the Stat "Mortalitate cumulată".
- Guide: `guide_week.mortality_cum_pct` for SAPT(d), ♀ (laying ♀ only: 7,25 at SAPT 59). Status: actual > guide → `abatere`, else `în normă`; the Stat adds "(ghid din intrarea în ouat)" because the guide counts from the start of lay (D161). No ♂ guide → "– fără date în ghid".
- A past cycle without a flock count (Seria 2025) shows deaths as counts from `legacy_mortality` (emptying days excluded) and `–` for the % with the reason (D54).

### 5.6 Lay %, eggs, categories, egg weight (D18, D156, D159, D162, D176, D183, D184, D186)
```
eggs(u, d)          = egg_production.total_eggs = cat1 + cat2 + cat3 + hatching_eggs      (Jos apart, D18)
hens_start(u, d)    = Σ alive_start(s, F, d) over the compartments of egg unit u
lay_day(u, d)       = 100 × eggs(u,d) / hens_start(u,d)                                     -- D156: START of day
hen_days(u, W)      = Σ_{d ∈ W, eggs(u,d) not null} hens_start(u, d)                       -- D159
lay_week(u, W)      = 100 × Σ_{d ∈ W} eggs(u,d) / hen_days(u, W)                            -- never the mean of daily %
farm / hall         = the same with Σ over egg units (section sums, never a register Total row)
inc_pct  = 100 × Σ hatching_eggs / Σ eggs        jos_pct  = 100 × Σ floor_eggs / Σ eggs
catk_pct = 100 × Σ catk / Σ eggs                  (cat1_pct + cat2_pct + cat3_pct + inc_pct = 100)
egg_weight(W)       = Σ (eggs × avg_egg_weight_g) / Σ eggs   over unit-days with both values  -- D176, empty days excluded
guide_lay(d)        = guide_week.lay_pct for (cycle.guide, SAPT(d), F)                       -- Productivitate col I (D184)
guide_lay(period)   = Σ guide_lay(d) × hens_start(d) / Σ hens_start(d)                       -- weighted by hen-days (D162)
diff_pts            = lay − guide_lay        status = |diff_pts| > 3 → abatere, else în normă (D19)
```
- 18.11.2026 farm: 5 914 ÷ 14 395 = **41,1 %** (guide 57,6 % → −16,5 pct, abatere); H5 45,3 % (−12,3), H6 34,9 % (−22,7). The workbook's 47,1 % (end-of-day count) is not used.
- SAPT 59 (17–18.11, 2 din 7 zile): 11 806 ÷ 29 221 hen-days = **40,4 %** (between its days 39,7 % and 41,1 %; the end-of-day basis 43,8 % is rejected). SAPT 58: 47 512 ÷ 156 418 = 30,4 %.
- Eggs 18.11: 5 914 · incubabile 5 308 (89,8 %) · jos 158 (2,7 %). Categories 18.11: H5 Cat. 1 13 · Cat. 2 417 · Cat. 3 27 + Incubabile 3 415 = 3 872; H6 8 · 110 · 31 + 1 893 = 2 042; farm SAPT 59 Cat. 0,3 % · 6,2 % · 1,1 %.
- Egg weight 18.11 farm = (3 872 × 67,04 + 2 042 × 65,08) ÷ 5 914 = 66,4 g; SAPT 58 H7 66,3 g, farm 66,6 g.
- "Găini" on a Producție week row = hens at the **start of the week** (D162); on a day row = `hens_start`.
- Egg data only exist per egg unit: a compartment's Producție shows its hall's figures with the note "ouăle se numără pe hală (H5)".

### 5.7 Feed: kg, g/bird, guide (D38, D46, D163, D165)
```
kg(row)              = g_per_bird × alive_start(s, x, d) / 1000      (or kg_recorded for migrated rows)   -- start of day (DATA03)
kg(scope, x, d)      = Σ kg(row)
g_day(section, x, d) = the typed g_per_bird
g_day(scope, x, d)   = 1000 × Σ kg / Σ alive_start over the fed rows of that day                         -- bird-weighted (D38)
g_week(scope, x, W)  = 1000 × Σ_W kg / Σ_W bird-days on fed days                                         -- D38, D165
guide_g(x, d)        = guide_week.feed_g for (cycle.guide, SAPT(d), x);  period: weighted by bird-days
deviation_pct        = 100 × (g / guide_g − 1);  |deviation| > 10 → abatere (D19)
```
- 18.11.2026: ♀ 14 395 × 160 g = **2 303,2 kg**; ♂ 154,7 kg (144/143/142/141 g by section); guide SAPT 59 ♀ 157 g · ♂ 156 g. SAPT 58: ♀ 22 230,9 kg · ♂ 1 682,7 kg, g ♀ 155,9 · ♂ 142,2.
- A day with no allocation for a section × sex with birds is **missing**, never carried over (data-model). Week rows cover fed days only and name the gap ("6 din 7 zile").
- Like-for-like totals (D46): Σ daily g over the fed days vs the guide over the same days.
- Migrated anomaly kept as typed: 02–06.11.2026 kg = 460 per compartment while g × birds ≈ 690 (D163, open question 12); the Hrană foot note says so.

### 5.8 Stock, days left, run-out (data-model; SYS04; D40, D49, D50)
```
stock(farm, type, d)   = Σ deliveries(date ≤ d) − Σ kg of allocations of that type in the farm's sections (date ≤ d) + Σ adjustments(date ≤ d)
avg_daily(farm, type)  = Σ consumption over the last 3 complete days ÷ 3
                         (window ends today when every row with birds has today's Hrană zilnică, else yesterday)
days_left              = stock ÷ avg_daily        (one decimal)
run_out_date           = window_end + floor(days_left) days
```
- 18.11.2026: Breeder 2 **31 664,9 kg** (= 1 114 110 delivered − 300 − 3 000 adjusted − 1 079 145,1 consumed), avg 16–18.11 2 392 kg/zi → **13,2 zile**, se termină 01.12.2026; Cocoși **1 781 kg**, 171 kg/zi → 10,4 zile, 28.11.2026.
- Lamp: stock < 0 → `abatere` "stoc sub 0 · verifică livrările"; days_left < 5 → `abatere` "sub pragul de 5 zile"; avg = 0 → `oprit` "– fără consum în ultimele 3 zile"; farm without an open cycle → `oprit` "fără ciclu deschis" (Telița never shows its workbook −716 kg).
- Hrană zilnică "Stoc după azi" = stock at the end of today including the form's confirmed rows (draft), "din {stock now} kg".

### 5.9 Feed cost (D05, D49, CostBlock)
```
deliveries(cycle)  = farm deliveries dated in (end_date of the farm's previous cycle, cycle.end_date or today]   -- pre-stocking before the first birds belongs to the cycle
lei(cycle)         = Σ total_lei of priced deliveries;  kg_priced = Σ their kg
unpriced(cycle)    = count, Σ kg, first–last date of deliveries without a price ("fără preț")
lei_per_bird       = lei ÷ birds at the cycle start (placed, or arrived by transfer: base ♀ + ♂ of §5.5 at the farm)
lei_per_egg        = lei ÷ Σ eggs of the cycle to date        (growth: "–" "creștere, fără ouă")
series total       = Σ lei ÷ series placement (♀ + ♂);  ÷ Σ eggs of the series
```
Whole lei for totals, 2 decimals for unit costs. Seria 2026: Telița 263 740 kg · 2 031 768 lei · 64,32 lei/pasăre (÷ 31 590); Hîrbovăț 1 174 480 kg priced · 9 562 437 lei · **319,74 lei/pasăre** (÷ 29 907) · **1,88 lei/ou** (÷ 5 073 155 eggs to 18.11) + 18 160 kg fără preț (4 livrări, 17.02–12.11.2026); total 11 594 205 lei · 367,02 lei/pasăre · 2,29 lei/ou.

### 5.10 Weighing vs guide (D157)
```
guide_weight = guide_week.body_weight_g for (cycle.guide, SAPT(d) − 1, x)       -- weight at the END of the week just ended
diff_pct     = 100 × (avg_weight_g − guide_weight) / guide_weight               -- one decimal
status       = |diff_pct| > 5 → abatere, else în normă;  no guide row → "– fără date în ghid"
```
- 24.02.2026 Telița H1 ♀ 2 553 g, SAPT 21 → guide week 20 2 370 g → **+7,7 %, abatere** (the old D95 comparison with week 21 is superseded).
- A Greutate week row shows that week's weighing; at farm/hall scope several sections' weighings of the same sex in a week are combined as the bird-weighted mean (weights × `alive_end` on the weighing day) *(assumption)*. A week without one opens to one muted row "fără cântărire în această săptămână" (D167).

### 5.11 Missing entries · Lipsesc date (C9, D65, D181)
For each day d in [today − 7, today] (today only from 12:00 local), each section s with birds and each kind:
```
has_birds(s, x, d)  = alive_start(s, x, d) > 0 and alive_end(s, x, d) > 0          -- the sale day that empties s is not missing
zilnic missing      = ∃ x with has_birds and no live mortality row for (s, x, d)      -- one row per section × day
hrana  missing      = has_birds(s, x, d) and no live allocation for (s, x, d)         -- one row per section × sex × day
```
An allocation is keyed by section × sex × day and carries exactly **one** feed type (unique index `feed_allocation_uq`; one ration per row, so a section × sex can't split a day over two types — assumption A15). Hence a Hrană missing row is section × sex × day (its feed type, if shown, is the row's last-used type), and stock per feed type sums the allocations of that type. One row per section × kind (× sex for Hrană) × day, newest first; each row opens its form at that day (manager) or shows Tag "blocat" + "o completează managerul" for a vet on a past day (D65). Example: H7.1 / H7.2 ♀ Hrană zilnică 13.11.2026 (g in the register, no kg).

### 5.12 Alerts · De verificat (D19, D169, D181, realdata.alerts)
One alert set per farm and day, computed once; S2 shows all the user's farms, S4 Rezumat the farm's subset. Display order:
| Kind | Condition (day d) | Title example | Roles |
|---|---|---|---|
| `lay` | hall daily lay < guide − 3 pts; `since` = first day of the uninterrupted run | "H5 · ouat 45,3% · ghid 57,6% · −12,3 pct" / "18.11.2026 · SAPT 59 · sub ghid din 01.11.2026" · link "Vezi producția" · Lamp abatere | all but vet |
| `feed` | section × sex g deviates > 10 % from guide | "H5.2 ♂ · 173 g · ghid 156 g · +11%" *(illustrative)* | all but vet |
| `missing` | §5.11 rows | "H7.1 · hrană ♀ – lipsă" / "13.11.2026 · SAPT 58" | owner, viewer (manager and vet have the Lipsesc date block instead, D86, D169) |
| `stock` | days_left < 5 or stock < 0 | "Breeder 2 · 3,8 zile" *(illustrative)*, "consumul depășește livrările" · "Deschide stocul" | owner, manager, viewer |
| `unpriced` | any delivery without price | "4 livrări fără preț · 18 160 kg" | owner only (own UnpricedCount block) |
Guide-based kinds are skipped when the cycle has no guide. An empty farm shows "nimic de verificat" + "Telița e goală din 28.02.2026; alertele apar cu ciclul nou."

### 5.13 Rezumat readout, stats and section tiles (S4)
- **Key readout**: laying → current week lay % vs guide ("SAPT 59 · efectiv 40,4% · ghid 57,6% · −17,2 pct") with the weekly series for the chart; growth → latest ♀ weight vs guide (SAPT − 1).
- **Stats** (follow the period, cycle by default, D191): alive ♀ · ♂; cumulative mortality ♀ · ♂ vs guide ♀; weight ♀ / ♂ (latest weighing, its date and section); feed g today ♀ · ♂ (bird-weighted) vs guide; feed days left; % lay today vs guide.
- **Section tiles** (per hall group, D41): code, ♀ · ♂ alive, deaths this SAPT, g today, weight vs guide, hall lay % today; lamps "lipsă" (any missing kind today) and "abatere" (lay/feed); an empty hall is one muted line "goală din 14.11.2026 · H7.1, H7.2" (D161).
- **Latest records**: the last 15 event-log rows (§7.5 Jurnal).

### 5.14 PeriodTable figures per row type (D154, D159, D162, D166, D190–D192)
| Column type | Day row | Week row | Foot "Total SAPT a–b" (loaded weeks) |
|---|---|---|---|
| counts (deaths, culls, sexing, transfer, placed, sold, eggs, incubabile, jos, Cat. counts, kg, deliveries kg, lei) | the day | Σ days in row | Σ loaded weeks |
| levels (în viață, stoc) | end of day | end of the row's last day | end of the latest loaded day |
| Găini (Producție) | start of day | start of the week | start of the oldest loaded week *(assumption)* |
| rates (% ouat, % incub., % jos, % Cat., g/pasăre, egg weight) | the day | Σnum ÷ Σden over the row's days | Σnum ÷ Σden over all loaded days |
| guide columns | guide of SAPT(d) | weighted by hen-/bird-days | weighted by hen-/bird-days |
| % cumulat | at end of day | at the row's last day | at the latest loaded day |
| stare (status lamp) | from the day's diff | from the week's diff | from the total's diff |
| Greutate | weighing of that day | the week's weighing | – |
Reference: foot "Total SAPT 56–59" (Hîrbovăț): eggs 226 472 · hen-days 545 038 · lay 41,6 % · guide 59,9 % · deaths ♀ 534 · ♂ 77 · sold ♀ 12 801 · ♂ 1 182 · % incub. 97,7 % · kg ♀ 78 898,6 · ♂ 5 996,6 · alive ♀ 12 561 · ♂ 794.

### 5.15 Rounding and units
- Counts: integers. Grams per bird: 1 decimal max (`160`, `141`). kg: one decimal where a record is named ("2 457,9 kg", D181), whole kg in stock readouts and the Rezumat feed column (S4N default). Lei: totals whole, unit prices 2 decimals (lei/kg shows 2: "7,76"; stored 4). Percentages: 1 decimal ("41,1%"), guide mortality 2 decimals ("7,25", D180), weight diff 1 decimal, points 1 decimal ("−16,5 pct").
- Round half away from zero (D82) from full precision; never round an intermediate (SAPT 30 peak 168 122 ÷ 186 088 = 90,345… → **90,3 %**; rounding the stored 90,35 first would give 90,4).
- "Shown parts add up to shown totals" (D82): when a table shows parts and their total (kg ♀ + ♂), round the parts, then show their sum as the total *(rule for the client formatter; the API returns full precision)*.
- The API returns numbers at full precision (decimals as JSON numbers with ≤ 4 decimals) plus, where a label is needed server-side (voice answers, xlsx head lines, Toast text), preformatted strings.

---

## 6. Business rules and validation

Conventions: **refuse** = HTTP 422 with `refusals[]`, nothing saved; **warn** = saved (or saveable), `warnings[]` returned; every message names the section, the number and the fix (spec §1 copy voice). `{sex_word}`: găini/cocoși; agreement "cele {n} găini" / "cei {n} cocoși". Dates `dd.mm.yyyy`, numbers with U+202F thousands (§11.3).

### 6.1 Counts never below 0 (PRODUCT; spec §1, C3, E1; D53)
Applies to **every** write that changes `bird_event` rows: create, edit, correction, delete, undo, restore, series age is not affected.
Algorithm (inside the write transaction, after applying the change, before commit):
0. **Lock first**: `select pg_advisory_xact_lock(hashtextextended(farm_id::text, 0))` for every farm the write touches, **in ascending farm-id order** (a transfer locks both farms in that order; no deadlock), before reading counts. The check, `reconcileCycles` and the write then see a stable farm; two saves on the same farm serialize, saves on different farms run in parallel. The lock is released at commit/rollback.
1. `affected` = every (section, sex) whose rows changed; `from` = the earliest changed date.
2. `bad` = rows of `v_section_sex_day` for `affected` with `date ≥ from and alive_end < 0`.
3. If `bad` is empty → commit. Else roll back and return 422 with, per (section, sex): `firstNegativeDate`, `lastNegativeDate`, `minAliveEnd`, and the bound for the edited field: `max = typed + min(alive_end over date ≥ from)` for a removal (min is negative: 4 292 + (−8) = 4 284), `min = typed − min(…)` for an inflow being lowered (4 381 − (−2) = 4 383).

Messages:
| Form | Field message | Banner |
|---|---|---|
| Zilnic (per-day bound = `alive_start`, D53) | "H5.2 ♂: Mortalitate + Sacrificare + Eroare sexare = 3 230, peste cei 323 cocoși în viață pe 18.11.2026. La Mortalitate scrie cel mult 323." | SaveBar "refuzat · H5.2 ♂ 3 230 · cel mult 323" |
| Transfer | "H2: 4 500 depășește cele 4 383 găini în viață la Telița pe 28.02.2026. Introdu cel mult 4 383." | "Transfer refuzat: Telița › H2 ar ajunge la ♀ −117 pe 28.02.2026. Nicio secțiune nu poate scădea sub 0." |
| Vânzare | "H6.2: 98 depășește cei 93 cocoși în viață pe 18.11.2026. Introdu cel mult 93." | "Vânzare refuzată: Hîrbovăț › H6.2 ar ajunge la ♂ −5 pe 18.11.2026. Nicio secțiune nu poate scădea sub 0." |
| Corectează (removal raised) | "H5.2 ♀ 4 292: ar lăsa −8 găini azi, 18.11.2026. Scrie cel mult 4 284." | "Corecția ar lăsa H5.2 cu −6 găini pe 17.11.2026 și cu −8 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." (short form, spec E1: "Corecția ar lăsa H6 cu −4 găini pe 16.02.2026.") |
| Corectează (inflow lowered) | "H7.2 ♀ 4 381: vânzarea din 14.11.2026 a scos 2 691 găini din H7.2. Scrie cel puțin 4 383." | "Corecția ar lăsa Hîrbovăț H7.2 cu −2 găini de pe 14.11.2026 până azi: vânzarea din 14.11.2026 a golit secțiunea. Nicio secțiune nu poate scădea sub 0." |
| Șterge / undo / restore | – | *(proposed)* "Ștergerea ar lăsa {section} cu {n} {sex_word} pe {date}. Nicio secțiune nu poate scădea sub 0." |
Zilnic's per-day rule: for each section × sex, `mortality + cull + max(0, −sexing_error) ≤ alive_start(d) + inflows saved that day`.

### 6.2 Zilnic (C1, D18, D36, D53, D77, D79, D183)
| # | Condition | Effect | Message |
|---|---|---|---|
| Z1 | vet and `date ≠ today` | refuse 403 `PAST_DAY_VET` | "Pentru zilele trecute cere managerului." |
| Z2 | values not integers ≥ 0 (sexing error: integer, signed) | refuse | *(proposed)* "{sec} {sex}: scrie un număr întreg." |
| Z3 | a value for a section × sex without birds that day, or on a hall / retired section | refuse 422 | *(proposed)* "{sec} nu are {sex_word} pe {date}." |
| Z4 | cull > 0 without reason | refuse | "H5.1 ♀: Motiv lipsă pentru 2 găini sacrificate pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul." |
| Z5 | §6.1 per-day bound | refuse | see §6.1 |
| Z6 | `mortality > 3 × mean7` where `mean7` = Σ deaths of that section × sex over the 7 days before d ÷ 7 (one decimal; literal: a 0,0 mean warns on any death, D53) | warn | "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva." |
| Z7 | eggs on a section that isn't the egg unit, or at a growth farm | refuse 422 | *(proposed)* "Ouăle se scriu pe hală: {hall}." |
| Z8 | egg counts not integers ≥ 0; Masa medie outside 20–120 g | refuse | *(proposed)* "{hall}: masa medie {v} g e în afara limitei. Scrie o valoare între 20 și 120 g." |
| Z9 | value key already live in another entry | refuse 409 `ALREADY_ENTERED` + `entryId` | client opens Editează / Corectează |
| Z10 | hall lay today vs guide | info (not a warning): Lamp + "ghid 57,6% · −12,3 puncte" | – |
| – | "Fără pierderi azi" | rows with 0 for mortality and cull for the sexes present | Toast "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0" |
Hints the form needs (returned by the form-state endpoint, §7.7): `alive_start` ♀ · ♂, yesterday's deaths and `mean7` ("ieri ♀ 3 · media 7 zile 1,0"), guide lay % for SAPT(d).

### 6.3 Hrană zilnică (C2, D13, D40, D50, D57, D62)
| # | Condition | Effect | Message |
|---|---|---|---|
| H1 | g outside 1–300 | refuse | "H1.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g." (SaveBar "scrie 1–300 g") |
| H2 | feed type inactive, or `for_sex` ≠ row sex | refuse | *(proposed)* "{type} nu e pentru {sex_word}. Alege alt tip." |
| H3 | row for a section × sex without birds at the start of the day | refuse | *(proposed)* "{sec} nu are {sex_word} la începutul zilei {date}." |
| H4 | `|g / guide − 1| > 10 %` | warn | "H5.2 ♂: 173 g, cu 11% peste ghidul de 156 g. Verifică gramajul; se poate salva." ("sub ghidul" when below) |
| H5 | stock of a type at the end of d (or any later day) < 0 | warn (D40) | "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg. Verifică livrările." |
| H6 | vet past day | refuse | "Pentru zilele trecute cere managerului." |
- Only confirmed rows are sent (an unconfirmed suggestion is not saved; the section stays missing, D13). The suggestion source = yesterday's g and feed type per row (`GET` form state).
- Toast: "Hrană zilnică salvată · 8 rânduri · 2 457,9 kg · stoc Breeder 2: 31 665 kg".

### 6.4 Cycles (B4, data-model Cycle rules, D17, D54, D63)
**reconcileCycles(farm, from)** — run in the same transaction after every write that adds, changes or removes `bird_event` rows of the farm (both farms for a transfer):
1. Load farm totals `T(d) = Σ alive_end` for d ≥ the start of the cycle containing `from` (or the first event).
2. Walk days. Open auto cycle C: if `T(d) = 0` and the farm had birds earlier in C → set `C.end_date = d`, `ended_by = emptied` (unless a manual cycle starts on d+1).
3. No open cycle and an inflow (placement or transfer in from another farm) on d with `T(d) > 0` → open a cycle: `start_date = d`, `series` = the series of the arriving birds (placement entry's series; transfer: the series of the source cycle), `guide_id = farm.guide_id`, `started_by = auto`. Audit `cycle_auto_open`; live event; Toast/preview "Ciclu nou la Hîrbovăț · Seria 2026 · sapt. 20 de la 22.02.2026 · ghid Ross 308 ouat."
4. A previously emptied cycle whose farm no longer reaches 0 (a correction raised a count) → `end_date = null` (reopen; audit `cycle_reopen`), **unless** a later cycle of that farm exists → refuse the write *(proposed)* "Corecția ar redeschide ciclul {series} la {farm}, dar ferma are un ciclu nou din {date}. Corectează întâi ciclul nou."
5. Manual cycles (`started_by = manual`) keep their dates; reconcile never moves a manual start.
Refusals and rules:
| # | Rule | Message |
|---|---|---|
| C1 | One open cycle per farm; a series has at most one open cycle per farm | (DB) |
| C2 | Inflow (transfer in) of birds of **another series** than the farm's open cycle into a farm that holds birds. Passes when the open cycle's series = the arriving birds' series — e.g. after a manual new cycle that picked that series (D201) | *(proposed)* "{farm} are încă {n} păsări din {series}. Mută-le sau vinde-le înainte, sau începe un ciclu nou manual." |
| C3 | Începe ciclu nou (manager, owner): date ≤ today, ≥ the farm's last event of the previous cycle, Motiv required; guide of the farm's type | "A cycle can't start before the farm's last event of the previous cycle" → *(proposed)* "Ciclul nou poate începe cel mai devreme pe {date}, după ultima înregistrare a ciclului curent." |
| C4 | Manual start on date X: old cycle `end_date = X − 1`, `ended_by = manual`; new cycle `start_date = X`, `started_by = manual`, `series_id` = the series the user picks (default: the open cycle's series; any **active** series, D201). From X every record of the farm — including the birds that remain from the old series — is counted in the new cycle and dated by the new series' age (SAPT) | dialog note: "Ciclul curent (sapt. 59) se închide pe 18.11.2026. Săptămâna rămâne cea de vârstă a păsărilor." (the close is effective from X) |
| C5 | Schimbă ghidul ciclului: guide of the farm's type or none; applies from the cycle start (all comparisons recomputed) | "Ghidul „Ross 308 ouat” este pentru ouat. Ferma Telița este de creștere. Alege un ghid de creștere sau schimbă tipul." |
| C6 | Empty farm state | "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21." / no cycle ever: "Ferma e goală. Ciclul pornește la prima populare sau transfer." |

### 6.5 Transfer (C3, D83, S9)
| # | Condition | Message |
|---|---|---|
| T1 | Din = În, Din/În a hall or retired, ♀ + ♂ = 0 | *(proposed)* "Alege Din, În și câte păsări muți." (board hint text) |
| T2 | manager without both farms | "Nu ai acces la această fermă." |
| T3 | count > alive (§6.1) | see §6.1 |
| T4 | C2 (other series at the destination) | see §6.4 |
Preview (dry-run) lines, one per farm: "Ciclu închis la Telița, sapt. 21" + "ferma ajunge la 0 păsări"; "Ciclul la Hîrbovăț continuă, sapt. 21"; "Ciclu nou la Telița, sapt. 59" + "ferma e goală"; per section "rămân ♀ 12 086 · ♂ 789". Toast: "Transfer salvat · ♀ 4 383 · Telița H2 → Hîrbovăț H7.2 · ciclul Telița închis". Undo the same day (manager).

### 6.6 Vânzare (C4, D84, D85, D155, D158b, D168)
- One farm, one date, rows per section (each section once), counts ≥ 0, total > 0 ("Scrie câte vinzi"); buyer ≤ 120, note ≤ 500 optional; §6.1 bounds.
- Preview: "Ciclul la Hîrbovăț continuă, sapt. 59" + "rămân ♀ 12 561 · ♂ 794", or when every section of the farm reaches 0: "Hîrbovăț: ♀ 12 561 · ♂ 794 → 0 · ciclul se închide" and the client shows the confirm Dialog "Înregistrează vânzarea și închide ciclul la Hîrbovăț" / "Vinzi toate păsările din Hîrbovăț pe 18.11.2026: ♀ 12 561 · ♂ 794. Ferma ajunge la 0 păsări și ciclul ei se închide." The API requires `confirmEmptiesFarm: true` in that case (409 `CONFIRM_REQUIRED` otherwise) — voice can never set it.
- Toast: "Vânzare înregistrată · ♀ 1 831 · ♂ 289 · Hîrbovăț · ciclul continuă".

### 6.7 Livrare hrană (C5, D73)
- Farm (manager's farms), date ≤ today, active feed type, `qty_kg > 0` ("Scrie câte kg de Breeder 2 s-au livrat la Hîrbovăț."), note ≤ 200. A manager's request containing price fields → 403 `PRICE_FORBIDDEN` ("Prețul îl adaugă proprietarul în Prețuri hrană.").
- Toast: "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg · fără preț" (stock of that type at the end of the delivery date).

### 6.8 Prețuri (C6, D67–D69, D72, D122)
- Owner only. Each item: exactly one of `pricePerKg > 0` or `totalLei > 0`; the other = `total = round(kg × price, 2)` or `price = total ÷ kg` (4 decimals). Partial batches allowed ("Poți completa doar o parte din rânduri.").
- Changing an existing price: Motiv required, ≤ 200; logged `price_change` with old/new ("Prețul vechi (7,76 lei/kg), cel nou și motivul rămân în jurnal, în Istoric modificări."). Hint text: "Preț salvat: 7,76 lei/kg · 199 000 lei. Scrie prețul nou în lei/kg sau totalul; celălalt se calculează din 25 640 kg."
- Toast: "Prețuri salvate · 4 livrări · 18 160 kg intră în cost" (+ Anulează the same day). Empty list: "Toate livrările au preț."

### 6.9 Cântărire (C7, D95–D97, D157)
| # | Condition | Effect / message |
|---|---|---|
| W1 | section without birds of that sex | refuse *(proposed)* "{sec} nu are {sex_word} pe {date}." (one-sex section hint "H1 are doar găini.") |
| W2 | weight missing or outside 1–10 000 g | refuse *(proposed range)* "Scrie masa medie în grame." |
| W3 | CV or uniformitate outside 0–100 | refuse "H5.1 ♀: uniformitate 790% e peste limită. Scrie o valoare între 0 și 100%." |
| W4 | |diff vs guide SAPT − 1| > 5 % | warn "H1 ♀: 2 553 g, cu 7,7% peste ghidul de 2 370 g de la sfârșitul SAPT 20 (prag ±5%). Verifică proba; poți salva." |
| W5 | key already live | 409 `ALREADY_ENTERED` |
Live comparison line: "ghid la sfârșitul SAPT 20: 2 370 g · 2 553 g (+7,7%)"; no guide: "ghid la sfârșitul SAPT 57: – fără date în ghid". Toast "Cântărire salvată · H5.1 ♀ 3 960 g".

### 6.10 Users (A2, D78, D118–D120)
| # | Rule | Message |
|---|---|---|
| U1 | name required, ≤ 60 | "Scrie numele persoanei, cum vrei să apară în listă, de exemplu Ion Rusu." |
| U2 | login `[a-z0-9.]{3,40}` | hint "Cu el intră în aplicație: litere mici fără diacritice, cifre și punct, 3–40 caractere, de exemplu ion.rusu." |
| U3 | login unique (deactivated included) | "Loginul „petrea” este deja folosit, de contul dezactivat al lui Petrea. Reactivează-l din listă sau alege alt login, de exemplu petrea.2." (active holder: "Loginul „petrea” este deja folosit.") |
| U4 | role required | "Alege rolul: Veterinar, Observator, Manager sau Proprietar. De el depinde ce vede și ce introduce." |
| U5 | ≥ 1 farm unless owner | *(hint)* "Vede și introduce doar la fermele bifate; cel puțin una." |
| U6 | at least one active owner; the only active owner can't change own role | "Ești singurul proprietar activ, așa că rolul tău rămâne Proprietar. Ca să-l schimbi, dă întâi rolul Proprietar altui utilizator, de exemplu lui Ion Rusu." |
| U7 | can't deactivate yourself | "Nu îți poți dezactiva propriul cont." |
| U8 | a deactivated user can't get a new password | "Contul e dezactivat. Reactivează-l ca să-i dai o parolă nouă." |
| U9 | login is read-only after creation | "Loginul nu se schimbă după ce contul e creat." |
- Password: generated, 12 characters from `abcdefghjkmnpqrstuvwxyz23456789` (a–z without i, l, o + 2–9), CSPRNG; returned once; reset revokes all sessions ("Parola de acum nu va mai merge."). Deactivate revokes sessions ("Petrea nu va mai putea intra. Datele introduse de el rămân. Poți reactiva contul oricând."). Reactivate keeps the old password ("Contul lui Petrea e activ din nou · intră cu parola lui de dinainte").

### 6.11 Farms and sections (B1, D75, D110–D112)
| # | Rule | Message |
|---|---|---|
| F1 | name required | "Scrie numele fermei, de exemplu Telița 2." |
| F2 | name unique ignoring case/diacritics | "Numele e deja folosit de ferma Hîrbovăț (literele mari și diacriticele nu contează). Alege alt nume." |
| F3 | type required | "Alege tipul fermei: Creștere sau Ouat. Lista de ghiduri depinde de el." |
| F4 | type change while a cycle is open | "Tipul se schimbă doar cât ferma e goală: ciclul de ouat e deschis din 22.02.2026." |
| F5 | guide of the other type | "Ghidul „Ross 308 ouat” este pentru ouat. Ferma Telița este de creștere. Alege un ghid de creștere sau schimbă tipul." |
| F6 | section code unique in the farm (active) | "H5.2 există deja în H5. Scrie alt cod, de exemplu H5.3." |
| F7 | section code empty on rename | "Scrie codul nou al secțiunii H5-2 sau apasă Anulează." |
| F8 | retire a section with birds (now or on any later dated record) | "H5.1 are păsări. O poți retrage când e goală." |
| F9 | retire a hall with active compartments | "H7 are compartimente active (H7.1, H7.2). O poți retrage după ce le retragi pe ele." / with birds "H5 are compartimente cu păsări. O poți retrage după ce le retragi pe ele." |
| F10 | one compartment level; a compartment's parent must be a hall without birds of its own | *(proposed)* "Un compartiment nu poate avea compartimente." |
| F11 | adding a compartment under a section that holds birds (it would become a hall) | refuse *(proposed)* "{code} are păsări: mută-le întâi în compartimente." |
- Toast: "Telița salvată · ghid Ross 308 creștere de la următorul ciclu al fermei". Farms are not retired or deleted (D110).

### 6.12 Guides: import and preview (B2, D76, D113–D115, D180, D184, D185)
Accepted input: `.xlsx` (first sheet), `.csv` (`,` or `;`), or a pasted TSV table. First non-empty row = header. Header matching ignores case, diacritics, spaces and units; required `SAPT` and `Sex`; optional `Hrană g`, `Greutate g`, `Mortalitate %`, `Ouat %`, `Incubabile %`, `Masă ou g`. Each data row = one preview row with the file's own line number (D76 B).
| # | Check (per row unless noted) | Result | Message |
|---|---|---|---|
| G1 | file not .xlsx/.csv | file refused | "Ross 308 creștere.pdf nu se poate citi. Încarcă un fișier .xlsx sau .csv, sau lipește tabelul." |
| G2 | no header with SAPT and Sex | file refused | "Ross 308 creștere.csv nu are antetul cu coloanele SAPT și Sex. Pune pe primul rând numele coloanelor de sub chenar și încarcă fișierul din nou." |
| G3 | header only | file empty | "Ross 308 creștere.csv are doar antetul, niciun rând de ghid." + "Scrie rândurile ghidului sub antet sau alege alt fișier." |
| G4 | SAPT not an integer 1–120 | row refused | *(proposed)* "SAPT „{v}” nu e o săptămână. Scrie un număr întreg de la 1." |
| G5 | Sex missing/unknown (accepted: G/C, F/M, ♀/♂, găini/cocoși) | row refused | "Lipsește sexul. Scrie G (♀ găini) sau C (♂ cocoși)." |
| G6 | duplicate SAPT + Sex | row refused (both cells marked) | "Rândul 8 are deja sapt. 4 ♀. Păstrează un singur rând." |
| G7 | number cell not a number | row refused | "Greutatea „#REF!” nu e număr. Scrie greutatea în grame." |
| G8 | a % > 100 or < 0 | row refused | "Ouat % „645” e peste 100. Scrie procentul cu virgulă, între 0 și 100." |
| G9 | blank cells, absent columns | info, saveable | "Lipsurile nu blochează salvarea. Greutatea lipsește la sapt. 24–25 (rândurile 48–51), iar fișierul nu are coloanele Mortalitate %, Ouat %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid." |
| G10 | type = ouat and the first row is SAPT 1 | file-level warn, saveable (D180) | "Ghidul de ouat începe la SAPT 1. Săptămâna 1 e prima săptămână de viață. Ghidurile de ouat încep de obicei la SAPT 19–23; verifică dacă fișierul numără săptămânile de la transfer." |
| G11 | any refused row | save blocked (D76 A) | "1 rând refuzat. Corectează-l în tabel și lipește-l din nou; până atunci ghidul nu se poate salva." / "3 rânduri refuzate. Corectează-le în fișier și încarcă-l din nou; până atunci ghidul nu se poate salva." |
| G12 | name empty / duplicate (ignoring case and diacritics) | refuse | "Scrie numele ghidului, de exemplu Ross 308 ouat 2027." / "Numele e deja folosit de ghidul Ross 308 ouat (literele mari și diacriticele nu contează). Scrie alt nume, de exemplu cu anul ediției, sau deschide Ross 308 ouat și apasă Înlocuiește ghidul." |
| G13 | type empty | refuse | "Alege tipul ghidului: Creștere sau Ouat. De el depinde cum se citește tabelul." |
| G14 | replace: type fixed while used | info | "Tipul rămâne ouat: ghidul e folosit de Hîrbovăț." |
| G15 | delete a used guide | refuse | "Folosit de 2 cicluri: îl poți înlocui, nu șterge." |
Number parsing: decimal comma or point, thousands separators (space, U+202F, `.` when followed by 3 digits and a comma decimal), a trailing `%` or `g`. Lay % and mortality on ♂ rows are ignored (shown `–`). Counts line "51 rânduri citite · 48 valide · 3 refuzate". Save Toast: "Ross 308 creștere salvat · 50 de rânduri înlocuite · Telița · Seria 2026 se compară cu ele".

### 6.13 Feed types (D116, D117)
| # | Rule | Message |
|---|---|---|
| FT1 | name required | "Scrie numele tipului, de exemplu Breeder 3." |
| FT2 | unique ignoring case/diacritics | "Numele e deja folosit de tipul Cocoși (literele mari și diacriticele nu contează). Alege alt nume, de exemplu Cocoși 2, sau editează Cocoși din listă." |
| FT3 | for-sex required while new/unused; fixed once used | "Alege pentru cine e hrana: ambele sexe, găini sau cocoși. În Hrană zilnică apare doar la sexul ales." / "Nu se mai schimbă: Breeder 2 are livrări la Hîrbovăț. Pentru alt sex, adaugă un tip nou." |
| FT4 | retire while any farm's stock of it > 0 | "Breeder 2 nu se poate retrage: Hîrbovăț mai are 31 665 kg în stoc. Bifează Activ la loc; îl retragi când stocul ajunge la 0 kg." |
| FT5 | delete a used type | "Are livrări: rămâne în rapoarte, așa că nu se poate șterge, doar retrage." |
Toasts: "Breeder 3 adăugat · pentru găini, apare în Livrare hrană și Hrană zilnică"; "Starter activ din nou · pentru ambele sexe, apare în Livrare hrană și Hrană zilnică".

### 6.14 Series (B3, D107–D109, D151, D179)
| # | Rule | Message |
|---|---|---|
| S1 | name required, unique | "Scrie numele seriei, de exemplu Seria 2027." |
| S2 | year unique | "Anul 2026 e deja folosit de Seria 2026. Scrie anul seriei noi, de exemplu 2027." |
| S3 | farm = growth farm | (list shows growth farms only) |
| S4 | date ≤ today | "19.11.2026 e în viitor. Alege Azi, …" (S18 board) |
| S5 | age at placement required, whole, 0–140 | "Scrie vârsta la populare în zile, 0 pentru pui de o zi." / "Vârsta la populare e 160 de zile. Scrie o valoare între 0 și 140." / "Scrie un număr întreg de zile, de exemplu 0." |
| S6 | ≥ 1 placement row with a count; each section once | "Niciun rând completat. Scrie câte găini sau cocoși intră în cel puțin o secțiune a Teliței." / "H1 e deja în listă. Alege altă secțiune sau elimină rândul." |
| S7 | farm holds birds | "Telița are încă 29 907 păsări din Seria 2026. Mută-le sau vinde-le înainte, sau începe un ciclu nou manual." (text kept, D201). The third option: the request may carry `manualCycle: { reason }`; then the creation closes the farm's open cycle on `date − 1` and opens a **manual** cycle for the new series from `date` in the same transaction (D201 applied at creation; Claude detail, to review) |
| S8 | Închide seria while a cycle of it is open | "Seria 2026 se poate închide după ce se încheie ciclul de la Telița (deschis, sapt. 59)." |
| S9 | age change: owner, Motiv | preview "Toate săptămânile Seriei 2026 se renumerotează (0 → 7 zile: azi SAPT 59 → SAPT 60) și se recompară cu ghidurile." |
Live line: "La populare: vârstă 0 zile · SAPT 1." (21 → "vârstă 21 de zile · SAPT 4"). Toast: "Seria 2026 creată · ciclu nou la Telița · sapt. 1 de la 07.10.2025 · ghid Ross 308 creștere".

### 6.15 Motiv, corrections, deletes (E1, D101–D103, D172, D173)
- Motiv required for every correction, delete by a non-author or after the author window, price change, Începe ciclu nou, series age change; trimmed length 1–200.
  - missing: "Motiv lipsă. Scrie de ce se schimbă H5.2 ♀ din 1 în 3; fără motiv corecția nu se salvează."
  - too long (soft limit, never truncated): "Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți salva."
- A correction covers the whole save (D172): any figure of any section × sex in it; the date is not editable; §6.1 runs over every later day; a correction that changes nothing → 422 *(proposed)* "Nu ai schimbat nicio cifră."
- Effect preview lines (dry-run): per change and day "H2 · găini în viață de pe 03.11.2026 se schimbă cu +12." and the totals "morți ♀ −5 → −7" and "Hîrbovăț ♀ azi 12 563 → 12 561".
- Delete Dialog text (whole save): "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:07: 4 secțiuni · ♀ −5 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 563 → ♀ 12 568. Ștergerea rămâne în Istoric modificări."
- Vet same-day delete of a section's rows from Introdus azi (D98): removes that section's Zilnic values only (mortality, culls, sexing errors; eggs have their own row).

### 6.16 Sign-in (A1, D80)
| Condition | HTTP | Message |
|---|---|---|
| empty login / password | 400 | "Scrie loginul." / "Scrie parola." |
| wrong login or password | 401 | "Login sau parolă greșită." |
| 5 failures for the login (or 20 for the IP) within 15 min | 429 + `retryAfter` | "Prea multe încercări. Încearcă din nou peste 5 minute." |
| deactivated | 403 | "Contul tău este dezactivat. Vorbește cu administratorul." |
| offline (client only) | – | "Nu există conexiune. Verifică internetul și încearcă din nou." |
The deactivated check runs only after a correct password (no account probing). Length rules (password ≥ 8) apply on creation only (generated passwords are 12).

---

## 7. API

### 7.0 Conventions
- **Transport**: JSON over HTTPS, served by the Go API under `/api` (routes registered on the stdlib `http.ServeMux`, module layout §12.3); the Nuxt front end calls it same-origin through the reverse proxy (§1). Request and response bodies are camelCase. Dates `YYYY-MM-DD` (local days), timestamps ISO 8601 with offset (`2026-11-18T13:16:00+02:00`). Sex `"F"`/`"M"`. Numbers are JSON numbers at full precision (≤ 4 decimals); formatting is client-side (§11.3) except the few preformatted strings named below (`toast`, `message`, `label`). The Go realisation of these conventions is §7.0a; validation §7.0b.
- **Auth**: HttpOnly cookie `il_session` (§11.4). Every route except `POST /api/auth/login` requires it → `401 UNAUTHENTICATED`.
- **Authorization**: §3 on every route; farm-scoped routes check `farmId` (and `peerFarmId`) against the caller's farms → `403 FARM_FORBIDDEN` "Nu ai acces la această fermă.".
- **Errors**: one envelope:
```json
{ "error": { "code": "REFUSED", "message": "Transfer refuzat: Telița › H2 ar ajunge la ♀ −117 pe 28.02.2026. Nicio secțiune nu poate scădea sub 0.",
  "refusals": [ { "field": "F", "sectionCode": "H2", "sex": "F", "max": 4383,
                  "message": "H2: 4 500 depășește cele 4 383 găini în viață la Telița pe 28.02.2026. Introdu cel mult 4 383." } ],
  "warnings": [] } }
```
  Codes: `VALIDATION` 400 (shape: malformed JSON, unknown field, wrong type, missing required field; the problems in `issues[]`, §7.0b), `UNAUTHENTICATED` 401, `FORBIDDEN` / `FARM_FORBIDDEN` / `PAST_DAY_VET` / `PRICE_FORBIDDEN` / `ACCOUNT_DISABLED` 403, `NOT_FOUND` 404, `ALREADY_ENTERED` / `CONFIRM_REQUIRED` / `SERIES_CLOSED` / `STALE` 409, `PAYLOAD_TOO_LARGE` 413 (body over the route's limit: 64 KiB JSON, 2 MB guide file, 1 MB audio; *(proposed)* "Datele trimise sunt prea mari."), `REFUSED` 422 (business rules §6), `RATE_LIMITED` 429 (+ `retryAfterSeconds`), `VOICE_UNAVAILABLE` 503.
- **Warnings** never block: successful responses carry `warnings: [{ code, sectionCode?, sex?, message }]`.
- **Dry run**: every write that shows an effect preview accepts `"dryRun": true` and returns `{ effects, warnings, refusals }` without writing (transfer, sale, correction, delete, new cycle, series create, series age change, guide save; the entry creates accept it too). A dry run answers **200** with the refusals **in the body** (never 422): the server runs the real writes and checks in a transaction (child rows, §6.1 counts, `ReconcileCycles`) and then rolls it back, so the preview is exactly what a save would do. No audit, live event or replay row survives it. The client calls it on change (debounced 300 ms).
- **Idempotency**: create requests carry `clientRequestId` (UUID generated when the form draft starts). A repeat with the same id returns the first result (`200` with the same body, byte for byte: the original response is stored with the key in `request_replay`, §4.11, §12.5), so "Reîncearcă" never saves twice.
- **Optimistic concurrency**: entry reads return `version` (= `updated_at ?? created_at`); edits send `ifVersion`; mismatch → `409 STALE` *(proposed)* "Înregistrarea s-a schimbat între timp. Reîncarcă și încearcă din nou."
- **Every successful write** returns `toast` (the exact Toast text), the affected entity, `asOf`, and emits live events (§9).
- **Pagination**: cursor-based (`cursor` opaque, `nextCursor` null at the end) except the PeriodTable, whose cursor is the oldest loaded SAPT (§7.4).

### 7.0a JSON in Go (struct tags, numbers, dates, null vs omitted)
Every request and response body is a named Go struct in the module's `dto.go` (§12.3) with an explicit `json:"…"` tag on **every** field; sqlc row types (§12.4) never reach the wire. `encoding/json` from the standard library is the codec.
| Topic | Rule (Go) |
|---|---|
| Names | camelCase tags exactly as the JSON in this spec (`json:"clientRequestId"`, `json:"F"`, `json:"M"`); no reliance on Go field names. Maps only where the keys are data: PeriodTable `cells`/`status`/`reasons` (keys `"{leaf}"`, `"{leaf}@{sectionCode}"`), `voice` `args`, `form.values`. |
| Decoding | One helper `httpx.Decode[T](w, r, maxBytes)`: `http.MaxBytesReader` (64 KiB for JSON bodies; 2 MB guide upload §7.17; 1 MB audio §8.2), `Content-Type` must be `application/json`, `dec.DisallowUnknownFields()` (a misspelt field is a 400, not silently ignored), exactly one JSON value (a second `Decode` must return `io.EOF`). Decode errors become `400 VALIDATION` with one issue naming the JSON path; a body over the limit (`*http.MaxBytesError`) becomes `413 PAYLOAD_TOO_LARGE`. |
| Encoding | One helper `httpx.JSON(w, status, v)`: `Content-Type: application/json; charset=utf-8`, `enc.SetEscapeHTML(false)` (keeps `›`, `<`, `&` readable in toasts), `Cache-Control: no-store` on every `/api` response. |
| Integers | Counts, eggs, SAPT, age days: `int` (fits `int32` in the DB). In **requests**, user-typed counts decode into `*num.Num` (below) so that `2.5` reaches `Validate()` and gets the Romanian sentence naming the section (Z2) instead of a generic decode error; `Validate()` converts to `int32`. |
| Decimals | Type `num.Num` (wraps `shopspring/decimal.Decimal`, §11.2): `MarshalJSON` writes a **JSON number** (never quoted) rounded half away from zero to ≤ 4 decimals, trailing zeros dropped (`40.4024…` → `40.4024`, `2303.2`); `UnmarshalJSON` accepts a JSON number only (a string is a 400). No `float64` for kg, lei, grams or percentages anywhere on the server; charts too use `num.Num`. |
| Days | Type `timex.Day` (a civil date, no time zone): JSON `"YYYY-MM-DD"`, parsed with `time.Parse("2006-01-02", s)` and checked (`2026-02-30` → 400). It implements the pgx date scanner/valuer interfaces so sqlc can map `date` to it (verify the interface names at implementation). Never a `time.Time` for a local day (a UTC midnight shifts the day). |
| Instants | `time.Time` in Go and `timestamptz` in SQL, stored UTC. Type `timex.LocalTime` marshals `t.In(timex.Chisinau).Format(time.RFC3339)` → `"2026-11-18T13:16:00+02:00"` (seconds, offset, no fraction, never `Z`). `asOf`, `at`, `createdAt`, `now` use it. |
| Enums | `"F"`/`"M"`, roles, kinds, statuses are Go string types generated by sqlc from the PostgreSQL enums (`dbq.Sex`, `dbq.UserRole`, …) or small string types in the DTO package; requests check membership in `Validate()`. |
| IDs | `uuid.UUID` (`github.com/google/uuid`), JSON string. |
| `null` vs omitted | **`null`** = the value exists in the shape but is unknown or empty: the client shows `–` (`"endDate": null`, `"pricePerKg": null` on an unpriced delivery for owner/manager, table cells, `"more": null`, `"start": null`). Go: a pointer (`*num.Num`, `*timex.Day`, `*uuid.UUID`, `*string`) **without** an omit option. **Omitted** = the field doesn't apply to this row/state (`partial`, `startInferred`, `emptySince`, `tag`, `noGuide`, `empty`): tag option `omitzero` (Go ≥ 1.24; omits the zero value, so a nil slice/map is omitted but an empty non-nil one is written as `[]`/`{}`). **Never sent** = fields the caller may not see (lei for vet/viewer, D200): not a nil pointer but a **different struct type** chosen by role (`reports.DeliveryRow` vs `reports.DeliveryRowPriced`, which embeds it and adds `PricePerKg`, `TotalLei` and the `pricedBy`/`actions` fields), so a forgotten check can't leak a price; tests assert the viewer JSON has no `pricePerKg`, `totalLei`, `lei*`, `cost` key. |
| Lists | Always `[]`, never `null` (`warnings: []`, `weeks: []`): constructors initialise slices with `make([]T, 0)`; a test walks every golden response and fails on a JSON `null` array. |
| Requests: absent vs `null` | For create bodies both mean "empty" (spec C1: empty is null, never 0): a nil pointer creates no row, `0` creates a 0 row. For PATCH bodies where absent = "leave as is" and `null` = "clear" (`PATCH /api/cycles/:id/guide { guideId \| null }`, `PATCH /api/farms/:id { guideId? }`) use `httpx.Optional[T]` (`Set bool; Null bool; V T`, `UnmarshalJSON` sets `Set` when the key appears). |

Shared Go types used below (package `httpx` unless named):
```go
type ErrorBody struct {
	Error APIError `json:"error"`
}
type APIError struct {
	Code              string     `json:"code"`                       // VALIDATION, REFUSED, FARM_FORBIDDEN, …
	Message           string     `json:"message"`                    // the Romanian sentence (Banner / SaveBar)
	Refusals          []Refusal  `json:"refusals,omitzero"`          // REFUSED: always non-nil (written as [] when empty)
	Warnings          []Warning  `json:"warnings,omitzero"`          // REFUSED: always non-nil
	Issues            []Issue    `json:"issues,omitzero"`            // VALIDATION
	EntryID           *uuid.UUID `json:"entryId,omitzero"`           // ALREADY_ENTERED
	Dialog            *Dialog    `json:"dialog,omitzero"`            // CONFIRM_REQUIRED (§6.6 title/body)
	RetryAfterSeconds *int       `json:"retryAfterSeconds,omitzero"` // RATE_LIMITED
}
type Refusal struct {
	Field             string     `json:"field,omitzero"` // form path ("F", "sections[2].F.mortality")
	SectionCode       string     `json:"sectionCode,omitzero"`
	Sex               string     `json:"sex,omitzero"`
	Max               *int       `json:"max,omitzero"` // §6.1 bound for a removal
	Min               *int       `json:"min,omitzero"` // §6.1 bound for a lowered inflow
	FirstNegativeDate *timex.Day `json:"firstNegativeDate,omitzero"`
	LastNegativeDate  *timex.Day `json:"lastNegativeDate,omitzero"`
	MinAliveEnd       *int       `json:"minAliveEnd,omitzero"`
	Message           string     `json:"message"`
}
type Warning struct {
	Code        string `json:"code"` // MORTALITY_SPIKE, FEED_GUIDE, STOCK_NEGATIVE, WEIGHT_GUIDE, …
	SectionCode string `json:"sectionCode,omitzero"`
	Sex         string `json:"sex,omitzero"`
	Message     string `json:"message"`
}
type Issue struct {
	Path    string `json:"path"`    // JSON path: "sections[0].F.mortality"
	Code    string `json:"code"`    // not_integer, out_of_range, required, unknown_field, bad_date, bad_type, too_long
	Message string `json:"message"` // Romanian: the §6 sentence where one exists, else a short generic one
}
type SexPair[T any] struct {
	F T `json:"F"`
	M T `json:"M"`
}
type WriteMeta struct { // embedded in every create body (fields are promoted in JSON)
	ClientRequestID *uuid.UUID `json:"clientRequestId"`
	Source          string     `json:"source"` // "manual" (default) | "voice"
	VoiceTranscript *string    `json:"voiceTranscript"`
	VoiceToken      *string    `json:"voiceToken"`
	DryRun          bool       `json:"dryRun"`
}
type EntryRef struct {
	ID      uuid.UUID `json:"id"`
	Kind    string    `json:"kind"`
	Version string    `json:"version"` // coalesce(updated_at, created_at) as RFC 3339 with microseconds (opaque to the client)
}
```
`httpx.WriteError(w, err)` is the only way errors leave a handler: it maps typed Go errors (`*httpx.Error`, `*entries.RefusedError`, `pgconn.PgError` codes `23505` → `ALREADY_ENTERED`/name clash by constraint name, `23P01` → cycle overlap refusal, `context.DeadlineExceeded` → 503) to the envelope; anything unexpected → `500 INTERNAL` "A apărut o eroare. Încearcă din nou." *(proposed)*, logged with the request id, no details to the client.

### 7.0b Validation (one approach: explicit `Validate()` methods)
Every request struct has a method `Validate(v *httpx.Validator)` written in plain Go; **go-playground/validator is not used**. Why: the messages are exact Romanian sentences with section codes, sex-word agreement ("cele 4 383 găini" / "cei 93 cocoși"), formatted numbers and dates (§6, §11.3); struct tags can express "1 ≤ g ≤ 300" but not "H1.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g.", so tags would need a second translation layer keyed by field and tag, which is where messages drift. Plain methods keep each rule next to its sentence and are trivially table-tested.
- **Level 1 — shape, no database** (`Validate()` on the decoded struct): required fields, enums, integrality, fixed ranges (g 1–300, % 0–100, Motiv 1–200 trimmed, buyer ≤ 120, login pattern), no duplicate sections in one body. Failures → `400 VALIDATION` with `issues[]`; the `message` of each issue is the §6 sentence when §6 has one (e.g. H1 "H1.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g." needs the section code, so the validator gets a resolver for codes from the bootstrap data the handler already loaded) and the `error.message` is the first issue's message.
- **Level 2 — business rules with database state** (in the service, inside the write transaction, after the farm lock): §6.1 counts, Z3/Z4/Z6, H2–H5, cycle rules, uniqueness (`ALREADY_ENTERED`), time windows. Failures → `422 REFUSED` with `refusals[]`/`warnings[]`, or the specific 403/409 code of §6/§7.0.
- **Messages**: every Romanian string of §6–§8 is a Go function in `internal/platform/messages/ro.go` with typed parameters (`messages.FeedOutOfRange(section string, sex dbq.Sex, g num.Num) string`); the formatters of §11.3 (`fmtx.Int`, `fmtx.Dec1`, `fmtx.Day`, `fmtx.SexWord`) live beside them. Tests snapshot each function's output against the exact strings quoted in this spec.
- The validator collects all issues (it doesn't stop at the first), in document order, so the client can mark every field at once.

### 7.1 Auth (A1, S1)
| Method · path | Body / response | Notes |
|---|---|---|
| `POST /api/auth/login` | `{ "login": "ceban", "password": "…" }` → `200 { user, farms, today }` + Set-Cookie | §6.16 errors; updates `last_login_at`; records `login_attempt` |
| `POST /api/auth/logout` | → `204` | revokes the session |
| `GET /api/auth/me` | → `{ "user": { "id", "name": "Dr. Ceban", "login": "ceban", "role": "vet" }, "farms": [{ "id", "name": "Hîrbovăț", "type": "laying" }], "today": "2026-11-18", "now": "2026-11-18T08:20:00+02:00", "startPage": "/acasa" }` | also the shell's "Cont" |

### 7.2 Shell bootstrap
`GET /api/bootstrap` → everything the shell, selectors and the voice context need, scoped to the caller:
```json
{ "farms": [{ "id": "…", "name": "Hîrbovăț", "type": "laying", "guide": { "id": "…", "name": "Ross 308 ouat" },
              "openCycle": { "id": "…", "seriesId": "…", "seriesName": "Seria 2026", "startDate": "2026-02-22", "saptStart": 20, "saptToday": 59, "ageDaysToday": 407 },
              "sections": [{ "id": "…", "code": "H5", "isHall": true, "children": [{ "id": "…", "code": "H5.1" }, { "id": "…", "code": "H5.2" }] }] }],
  "series": [{ "id": "…", "name": "Seria 2026", "status": "active", "saptSpan": [1, 59] }, { "id": "…", "name": "Seria 2025", "status": "closed", "saptSpan": [19, 65] }],
  "feedTypes": [{ "id": "…", "name": "Breeder 2", "forSex": "F", "active": true }],
  "settings": { "toleranceWeightPct": 5, "toleranceLayPts": 3, "toleranceFeedPct": 10, "stockDaysMin": 5 } }
```

### 7.3 Scope pages: series, farm, section headers (S3, S4, S5)
| Route | Returns |
|---|---|
| `GET /api/series/:id/page` | header (name, alive ♀ · ♂ across the user's farms, SAPT today), **cycle track** (`lanes[]` per farm: cycles with start/end, `startInferred`, open, SAPT span, transfer links `{date, fromFarm, toFarm, F, M}`, `today`), farm summaries (type, guide, cycle SAPT span, alive, key figures vs guide), tabs available |
| `GET /api/farms/:id/page?cycleId=` | header: name, type, guide tag (or "fără ghid"), alive ♀ · ♂, SAPT, "din 22.02.2026"; `cycles[]` (CycleSelector: every cycle of any series at this farm, newest first, with lamps deschis/închis); compact farm track; actions allowed for the caller (`move`, `sell`, `newCycle`, `changeGuide`); tabs (Producție only for laying cycles); empty-farm state ("Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21.") |
| `GET /api/sections/:id/page?cycleId=` | the same for a compartment (one group) or a hall ("Hîrbovăț › H5": compartments summed, Producție = hall eggs) |

Example (`GET /api/farms/hirbovat/page`, 18.11.2026):
```json
{ "farm": { "id": "…", "name": "Hîrbovăț", "type": "laying", "guide": { "id": "…", "name": "Ross 308 ouat" } },
  "cycle": { "id": "…", "series": "Seria 2026", "startDate": "2026-02-22", "endDate": null, "saptStart": 20, "saptToday": 59, "ageDaysToday": 407, "weeks": 40 },
  "alive": { "F": 12561, "M": 794 },
  "cycles": [ { "id": "…", "series": "Seria 2026", "label": "22.02.2026–azi · SAPT 20–59", "open": true },
              { "id": "…", "series": "Seria 2025", "label": "07.03.2025–28.01.2026 · SAPT 19–65", "open": false, "startInferred": true } ],
  "actions": { "move": true, "sell": true, "newCycle": true, "changeGuide": true }, "tabs": ["rezumat", "pasari", "hrana", "greutate", "productie"], "asOf": "2026-11-18T13:16:04+02:00" }
```

### 7.4 PeriodTable (D154, D159, D162, D165–D167, D190–D192)
`GET /api/table` — the one endpoint behind every table on S3, S4, S5, S11 Consum and S14.

| Param | Values | Default |
|---|---|---|
| `view` | `farm` · `section` · `series` · `report-production` · `report-feed` | – |
| `farmId` / `sectionId` / `seriesId` | ids per view (`report-*` take `farmId`) | – |
| `cycleId` | a farm cycle | the open cycle, else the latest |
| `tab` | `rezumat` · `pasari` · `hrana` · `greutate` · `productie` | per screen: farm and section views `rezumat`; series view (S3 has no Rezumat) `pasari`, its first tab; `report-*` ignore it |
| `period` | `cycle` · `weeks:4` · `weeks:8` · `weeks:13` · `interval:2026-11-01,2026-11-15` | `cycle` (series view: whole series, no param) |
| `detail` | `farm` · `halls` · `sections` (not on `rezumat`, not on section/series views) | `farm` |
| `types` | comma list of chips: `mortalitate,sacrificare,eroare_sexare,transfer,populare,vanzare,hrana,cantarire,oua` (phone sends one) | every type of the period |
| `grain` | `day` · `week` (week = no `days[]`) | `day` (D166) |
| `before` | cursor: the oldest SAPT already loaded (exclusive); absent = the newest 4 weeks | – |
| `count` | weeks per page | 4 |
| `sapt` | one week only: returns that week with its `days[]` (grain Săpt. opening one week, D166; voice F09) | – |
| `fromSapt` | load every week from `fromSapt` up to the newest not yet loaded (with `before`) — voice F08 "până la începutul ciclului" and F10 "arată ziua de 14 noiembrie" (the client sends the SAPT of that day) | – |

Response:
```json
{ "scope": { "view": "farm", "farmId": "…", "cycleId": "…", "seriesId": "…", "label": "Total fermă" },
  "period": { "kind": "cycle", "label": "Ciclul curent", "saptFrom": 20, "saptTo": 59, "from": "2026-02-22", "to": "2026-11-18", "weeksTotal": 40 },
  "chips": ["mortalitate", "sacrificare", "eroare_sexare", "transfer", "vanzare", "hrana", "cantarire", "oua"],
  "groups": [
    { "id": "vii", "record": "vii", "label": "În viață", "scope": "farm", "leaves": [ { "id": "alive.F", "label": "♀", "kind": "level" }, { "id": "alive.M", "label": "♂", "kind": "level" } ] },
    { "id": "morti", "record": "mortalitate", "label": "Morți", "scope": "farm", "leaves": [ { "id": "deaths.F", "kind": "count" }, { "id": "deaths.M", "kind": "count" } ] },
    { "id": "ouat", "record": "oua", "label": "Ouat", "scope": "farm", "leaves": [
        { "id": "egg.total", "label": "ouă" }, { "id": "egg.layPct", "label": "% ouat" }, { "id": "egg.guide", "label": "ghid" },
        { "id": "egg.inc", "label": "incub." }, { "id": "egg.incPct", "label": "% incub." }, { "id": "egg.jos", "label": "jos" }, { "id": "egg.josPct", "label": "% jos" } ] },
    { "id": "hranaKg", "record": "hrana", "label": "Hrană kg", "scope": "farm", "leaves": [ { "id": "feed.kg.F" }, { "id": "feed.kg.M" } ] },
    { "id": "vanzare", "record": "vanzare", "label": "Vânzare", "scope": "farm", "leaves": [ { "id": "sold.F" }, { "id": "sold.M" } ] } ],
  "weeks": [
    { "sapt": 59, "from": "2026-11-17", "to": "2026-11-18", "daysInRow": 2, "partial": true, "label": "SAPT 59 · 17.11–18.11 · 2 din 7 zile",
      "cells": { "alive.F": 12561, "alive.M": 794, "deaths.F": 10, "deaths.M": 0, "egg.total": 11806, "egg.layPct": 40.40, "egg.guide": 57.6,
                 "egg.inc": 10918, "egg.incPct": 92.48, "egg.jos": 311, "egg.josPct": 2.63, "feed.kg.F": 4675.3, "feed.kg.M": 329.7, "sold.F": 2255, "sold.M": 433 },
      "status": { "egg.layPct": "abatere" }, "reasons": {},
      "days": [
        { "date": "2026-11-17", "ageDays": 406, "sapt": 59, "weekday": "Ma",
          "cells": { "alive.F": 14395, "alive.M": 1083, "deaths.F": 7, "deaths.M": 0, "egg.total": 5892, "egg.layPct": 39.74, "egg.guide": 57.6, "egg.inc": 5610, "egg.incPct": 95.21, "egg.jos": 153, "egg.josPct": 2.60, "feed.kg.F": 2372.1, "feed.kg.M": 175.0, "sold.F": 424, "sold.M": 144 } },
        { "date": "2026-11-18", "ageDays": 407, "sapt": 59, "weekday": "Mi",
          "cells": { "alive.F": 12561, "alive.M": 794, "deaths.F": 3, "deaths.M": 0, "egg.total": 5914, "egg.layPct": 41.08, "egg.guide": 57.6, "egg.inc": 5308, "egg.incPct": 89.75, "egg.jos": 158, "egg.josPct": 2.67, "feed.kg.F": 2303.2, "feed.kg.M": 154.7, "sold.F": 1831, "sold.M": 289 } } ] },
    { "sapt": 58, "from": "2026-11-10", "to": "2026-11-16", "daysInRow": 7, "label": "SAPT 58 · 10.11–16.11",
      "cells": { "alive.F": 14826, "alive.M": 1227, "deaths.F": 36, "deaths.M": 8, "egg.total": 47512, "egg.layPct": 30.38, "egg.guide": 58.8, "egg.inc": 45745, "egg.incPct": 96.28, "egg.jos": 1221, "egg.josPct": 2.57, "feed.kg.F": 22230.9, "feed.kg.M": 1682.7, "sold.F": 10546, "sold.M": 749 }, "days": ["…7 day rows…"] },
    "…SAPT 57, 56…" ],
  "loaded": { "saptFrom": 56, "saptTo": 59, "weeks": 4 },
  "more": { "before": 56, "remaining": 36, "label": "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40" },
  "start": null,
  "totals": { "label": "Total SAPT 56–59", "cells": { "alive.F": 12561, "alive.M": 794, "deaths.F": 534, "deaths.M": 77, "egg.total": 226472, "egg.layPct": 41.55, "egg.guide": 59.9, "sold.F": 12801, "sold.M": 1182, "feed.kg.F": 78898.6, "feed.kg.M": 5996.6 } },
  "notes": [ "Morți include sacrificările, ca în registru." ], "asOf": "2026-11-18T13:16:04+02:00" }
```
Rules:
- **Weeks newest first; days inside a week by date** (D159). Every returned week has its `days[]` when `grain=day` (D166: every week open).
- **Load more** (D190–D192): the client calls again with `before = loaded.saptFrom`; the response contains only the next (older) ≤ 4 weeks **plus `totals` recomputed over every loaded week** (`saptFrom` = the new oldest, `saptTo` = the period's newest) and the new `more`. When the period has a set length (`weeks:n`, `interval`) the client widens it by 4 weeks ("4 săpt." → "8 săpt.", Interval start − 28 days) and sends the widened period with `before`. Fewer than 4 left → label "Încarcă încă 3 săptămâni"; at the first week `more = null` and `start = { "label": "Începutul ciclului · SAPT 20" }` (series view: "Începutul seriei · SAPT 1"). Live announcement text is client copy: "Au fost adăugate SAPT 52–55."
- **Groups per tab** (order: spine → `vii` first → record groups → `vanzare` last, D167):
  - `rezumat`: vii · morti (mortality + culls) · ouat (laying: ouă · % ouat · ghid · incub. · % incub. · jos · % jos) or growth: g/pasăre ♀ · ♂ (efectiv · ghid) + Greutate ♀ (weighing weeks) · hranaKg · vanzare (only when the period has sales). No `detail`.
  - `pasari`: vii (end of day) · mortalitate (morți ♀ · ♂) + `cumPct` group "% cumulat" ♀ · ♂ · sacrificare · eroare_sexare · transfer · populare · vanzare.
  - `hrana`: g/pasăre efectiv · ghid ♀ · ♂, kg ♀ · ♂; farm group adds Livrări kg per feed type, lei (owner, manager only; never viewer, D64) and Stoc kg (end of day).
  - `greutate`: per sex medie g · ghid (SAPT − 1) · față de ghid % · CV % · uniformitate %; weeks without a weighing → one day row with `reasons: { "*": "no_weighing" }` ("fără cântărire în această săptămână").
  - `productie` (laying only): găini (start) · ouă · % ouat · ghid · față de ghid (pct) · stare · % Cat. 1 · % Cat. 2 · % Cat. 3 · incubabile n · % · jos n · % · masă ou g (D183, D186).
  - `detail=halls|sections` adds one group per hall/compartment after "Total fermă" (scope = the section id); Producție with `sections` repeats the hall figures with `reasons: "hall_eggs"` ("ouăle se numără pe hală").
  - `series` view: one "Total serie" group + a text column `farm` per day ("Telița", "Hîrbovăț", both in transfer weeks SAPT 20–21; each bird counted once: in Telița until its batch arrives), every group of the series' types (Eroare sexare and Populare included, D45).
  - `report-production` (S14): per week the farm week row → its 7 farm day rows → the hall **week** rows (`halls[]` inside the week, no hall days); columns găini · ouă · Ouat (efectiv · ghid · față de ghid · stare) · % Cat. 1–3 · incubabile (n · %) · masă ou · mortalitate ♀ · ♂ · vândute ♀ (weeks with sales) · hrană g/găină (efectiv · ghid).
  - `report-feed` (S11 Consum): farm week row (kg ♀ · ♂, g vs guide) → 7 farm day rows → section week rows.
- **`cells` keys** are `"{leaf}"` for the scope group and `"{leaf}@{sectionCode}"` for detail groups (`"alive.F@H5"`). `null` = `–`; its reason in `reasons[key]`: `missing` ("lipsă"), `empty_since:2026-11-14` ("goală din 14.11"), `no_guide` ("fără date în ghid"), `no_weighing`, `not_in_register` ("– fără hrană în registru"), `hall_eggs`, `no_flock_count` (past cycle without counts). `status[key]` ∈ `in_norma|abatere` for guide comparisons.
- **Phone** sends a single `types` chip; the server still returns `vii` (and for Producție `găini`).
- **States**: no data in the period → `weeks: []`, `chips: []`, `empty: { "next": "Alege ciclul curent" }`; no guide → guide leaves omitted and `noGuide: true` (header Tag "fără ghid"); partial guide → null + `no_guide`.
- **Performance**: one SQL pass per page (deltas grouped by day for the page's weeks + one running-count seed at the page start), §11.6.

Go types (package `periodtable`; the query string is parsed by hand into `TableQuery`, unknown params → 400):
```go
type TableQuery struct {
	View      string     // farm | section | series | report-production | report-feed
	FarmID    *uuid.UUID // farmId
	SectionID *uuid.UUID // sectionId
	SeriesID  *uuid.UUID // seriesId
	CycleID   *uuid.UUID // cycleId (nil = open cycle, else the latest)
	Tab       string     // rezumat | pasari | hrana | greutate | productie ("" = per-view default)
	Period    Period     // parsed from "cycle" | "weeks:4" | "interval:2026-11-01,2026-11-15"
	Detail    string     // farm | halls | sections
	Types     []string   // chips; nil = every type of the period
	Grain     string     // day | week
	Before    *int       // oldest SAPT already loaded (exclusive)
	Count     int        // weeks per page, default 4
	Sapt      *int       // one week only
	FromSapt  *int       // load every week from here up to the newest not yet loaded
}
type Period struct {
	Kind     string    // cycle | weeks | interval
	Weeks    int       // weeks:n
	From, To timex.Day // interval
}

type TableResponse struct {
	Scope   Scope           `json:"scope"`
	Period  PeriodInfo      `json:"period"`
	Chips   []string        `json:"chips"`
	Groups  []Group         `json:"groups"`
	Weeks   []WeekRow       `json:"weeks"`
	Loaded  *Loaded         `json:"loaded,omitzero"` // absent when weeks is empty
	More    *More           `json:"more"`            // null at the first week of the period
	Start   *StartNote      `json:"start"`           // null until the first week is loaded
	Totals  *Totals         `json:"totals,omitzero"`
	Notes   []string        `json:"notes"`
	Empty   *EmptyState     `json:"empty,omitzero"` // { "next": "Alege ciclul curent" }
	NoGuide bool            `json:"noGuide,omitzero"`
	AsOf    timex.LocalTime `json:"asOf"`
}
type Scope struct {
	View      string     `json:"view"`
	FarmID    *uuid.UUID `json:"farmId,omitzero"`
	SectionID *uuid.UUID `json:"sectionId,omitzero"`
	CycleID   *uuid.UUID `json:"cycleId,omitzero"`
	SeriesID  *uuid.UUID `json:"seriesId,omitzero"`
	Label     string     `json:"label"`
}
type PeriodInfo struct {
	Kind       string    `json:"kind"`
	Label      string    `json:"label"`
	SaptFrom   int       `json:"saptFrom"`
	SaptTo     int       `json:"saptTo"`
	From       timex.Day `json:"from"`
	To         timex.Day `json:"to"`
	WeeksTotal int       `json:"weeksTotal"`
}
type Group struct {
	ID     string `json:"id"`
	Record string `json:"record"`
	Label  string `json:"label"`
	Scope  string `json:"scope"` // "farm" or a section id for detail groups
	Leaves []Leaf `json:"leaves"`
}
type Leaf struct {
	ID    string `json:"id"` // "alive.F", "egg.layPct", …
	Label string `json:"label,omitzero"`
	Kind  string `json:"kind,omitzero"` // count | level | rate | guide | status | text
}
type Cells map[string]*num.Num // nil value = "–"; its reason is in Reasons[key]

type WeekRow struct {
	Sapt      int               `json:"sapt"`
	From      timex.Day         `json:"from"`
	To        timex.Day         `json:"to"`
	DaysInRow int               `json:"daysInRow"`
	Partial   bool              `json:"partial,omitzero"`
	Label     string            `json:"label"`
	Cells     Cells             `json:"cells"`
	Status    map[string]string `json:"status,omitzero"`  // key → in_norma | abatere
	Reasons   map[string]string `json:"reasons,omitzero"` // key → missing | empty_since:2026-11-14 | no_guide | …
	Days      []DayRow          `json:"days,omitzero"`    // grain=day
	Halls     []WeekRow         `json:"halls,omitzero"`   // report-production: hall week rows
}
type DayRow struct {
	Date    timex.Day         `json:"date"`
	AgeDays int               `json:"ageDays"`
	Sapt    int               `json:"sapt"`
	Weekday string            `json:"weekday"`       // Lu Ma Mi Jo Vi Sâ Du
	Farm    string            `json:"farm,omitzero"` // series view text column
	Cells   Cells             `json:"cells"`
	Status  map[string]string `json:"status,omitzero"`
	Reasons map[string]string `json:"reasons,omitzero"`
}
type Loaded struct{ SaptFrom, SaptTo, Weeks int } // tags: saptFrom, saptTo, weeks
type More struct {
	Before, Remaining int
	Label             string
}                                     // tags: before, remaining, label
type StartNote struct{ Label string } // tag: label
type Totals struct {
	Label string
	Cells Cells
} // tags: label, cells (+ status/reasons like WeekRow)
```
(`Loaded`, `More`, `StartNote`, `Totals` are written on one line here; in code every field carries its camelCase tag like the others.) Cells hold full-precision `num.Num` values: counts are integral decimals (`12561`), rates keep ≤ 4 decimals (the example above shows 2 for brevity).

### 7.4b Tab charts and Stats (D154 "charts after the table", D191 Stats follow the period)
`GET /api/charts?<view, ids, cycleId, tab, period, detail>` — same scope params as `/api/table`, whole period (not paged) →
| Tab | `charts[]` (series per point: `{ sapt | date, value, guide? }`) | `stats[]` (follow the period; cycle by default) |
|---|---|---|
| rezumat | see §7.6 (readout + guide chart) | §7.6 stats |
| pasari | alive ♀ · ♂ per day; deaths per week ♀ · ♂; cumulative mortality % per week vs guide ♀ | "Morți · ciclu" ♀ · ♂, mortality cum. % vs guide, culls by reason, sexing errors, sold ♀ · ♂, transfers |
| hrana | g/pasăre per week ♀ · ♂ vs guide (±10 % band); kg/zi ♀ · ♂ per day (two lines, caption lamp vs guide g × birds, D58); stock per feed type per day + days left (farm) | "Consum · ciclu" kg ♀ · ♂, "Livrări · ciclu" kg (lei: owner, manager), g vs guide like-for-like (D46), days left |
| greutate | weight per weighing ♀ · ♂ vs guide SAPT − 1 (±5 % band); CV %, uniformitate % | "Cântăriri · ciclu" count, latest weight vs guide, CV, uniformity |
| productie | lay % per week vs guide (±3 pts band); egg weight per week; % incubabile per week | eggs in the period, lay % (hen-days), vs guide, peak (SAPT 30 90,3 %), egg weight (egg-weighted), % incub., % Cat. 1–3 |
Each chart also returns `table` rows for the screen-reader data table and `noGuide`/`partialGuide` flags (guide lines dropped or `null` points).

### 7.5 Event log · Jurnal (S4 Păsări, Rezumat latest records; D81, D82, D168, D178)
`GET /api/farms/:id/journal?sectionId=&types=&cursor=&limit=15`
- Page 1 = the last 7 days; `nextCursor` = "Mai vechi". Rows = one per **entry** (one save), newest first: `{ entryId, kind, at, date, sapt, author, summary: "Mortalitate · 4 secțiuni", figures: { F: -5, M: 0 }, sections: ["H5.1","H5.2","H6.1","H6.2"], corrected: { by: "Ion Rusu", at: "2026-11-18T10:12:00+02:00" } | null, actions: { edit: false, correct: true, delete: true } }`.
- Sale rows list their sections (`"Vânzare · H6.1, H6.2 · ♀ 1 831 · ♂ 289"`); cycle rows (open/close) carry `cycleMark: "start" | "end"` and no actions.
- `actions` is computed for the caller (§3.2): vet → none on others' rows (D71), `locked: true` ("blocat") on own past rows; viewer → none; owner → correct + delete on every row incl. today's vet entries (D81).

### 7.6 Rezumat summary (S4 b–f, S5)
`GET /api/farms/:id/summary?cycleId=&period=` (sections: `/api/sections/:id/summary`) →
```json
{ "readout": { "kind": "lay", "label": "SAPT 59 · efectiv 40,4% · ghid 57,6% · −17,2 pct", "value": 40.40, "guide": 57.6, "status": "abatere",
               "series": [ { "sapt": 20, "value": null, "guide": null }, "…", { "sapt": 59, "value": 40.40, "guide": 57.6 } ] },
  "stats": [ { "id": "alive", "F": 12561, "M": 794 },
             { "id": "cumMortality", "F": 5.66, "M": 34.66, "guideF": 7.25, "status": "in_norma", "note": "(ghid din intrarea în ouat)", "deaths": { "F": 1521, "M": 1048 } },
             { "id": "weight", "F": 3960, "M": 4870, "guide": null, "reason": "no_guide", "at": "2026-11-10", "section": "H5.1" },
             { "id": "feedToday", "F": 160, "M": 142.8, "guideF": 157, "guideM": 156, "status": "in_norma", "note": "H7 goală din 14.11" },
             { "id": "feedDaysLeft", "items": [ { "type": "Breeder 2", "days": 13.2 }, { "type": "Cocoși", "days": 10.4 } ] },
             { "id": "layToday", "value": 41.08, "guide": 57.6, "diffPts": -16.52, "status": "abatere", "eggs": 5914, "halls": 2 } ],
  "tiles": [ { "hall": "H5", "sections": [ { "code": "H5.1", "alive": { "F": 4263, "M": 342 }, "deathsWeek": { "F": 1, "M": 0 }, "gToday": { "F": 160, "M": 144 }, "layHall": 45.30, "lamps": [] } ] },
             { "hall": "H7", "emptySince": "2026-11-14", "label": "goală din 14.11.2026 · H7.1, H7.2" } ],
  "alerts": [ "…§5.12 items for this farm…" ],
  "latest": [ "…15 journal rows…" ] }
```
(The weight stat on Hîrbovăț uses the two *illustrative* 10.11.2026 weighings; the real workbook has none.)

### 7.7 Acasă (S2; D65, D86–D89, D169, D181)
`GET /api/home` → blocks by role:
| Role | Blocks (in order) |
|---|---|
| vet | `header` ("18.11.2026 · Hîrbovăț · SAPT 59"), `missing` (Lipsesc date rows, past days `locked: true`), `entryKeys` (Zilnic, Hrană zilnică, Cântărire), `today` (Introdus azi, §7.8) |
| manager | `header`, `alerts` (no missing/unpriced rows), `missing`, `feedDays` (per farm × type: kg, days, run-out, lamp), `entryKeys` (+ Transfer, Vânzare, Livrare hrană), `today` |
| owner | `track` (compact cycle track of the active series: lanes, blocks, SAPT readout, today line), `alerts` (incl. missing rows), `unpriced` ({ count: 4, kg: 18160, byType: [Breeder 2 11 000, Cocoși 7 160], from: "2026-02-17" }), `stock`, `farms` (FarmSummary per farm: type, guide, cycle SAPT span, alive, % lay, cumulative mortality vs guide) |
| viewer | `track`, `alerts` (incl. missing rows, read-only, no form links), `stock`, `farms` (no prices, no entry keys) |
Missing rows appear for today only from 12:00 (`app_setting missing.from_hour`). An empty farm: "Telița e goală din 28.02.2026; alertele apar cu ciclul nou."

### 7.8 Entry forms and saves (C1–C5, C7, C8; S6–S10, S13, S15)
**Form state** (read before editing; also feeds voice prefill):
| Route | Returns |
|---|---|
| `GET /api/forms/daily?farmId&date` | `sections[]` with birds (`aliveStart {F,M}`, `yesterday {F,M}`, `mean7 {F,M}`, hall code), `eggUnits[]` (hall, hens start, guide lay %), `existing[]` (entries already holding values for that day, with `editable`, `correctable`, author, time), `lockedReason` ("Pentru zilele trecute cere managerului." for a vet on a past day), `sapt`, `ageDays` |
| `GET /api/forms/feed-daily?farmId&date` | `rows[]` per section × sex with birds: `birdsStart`, `suggestion { g, feedTypeId }` (yesterday's), `guideG`, `existing`; `stock[]` per type: `now`, `avgDaily`, `daysLeft`; feed types allowed per sex |
| `GET /api/forms/weighing?farmId&date&sectionId&sex` | alive, sexes present, guide weight SAPT − 1, previous weighing, existing |
| `GET /api/forms/transfer?date` | sections of the caller's farms with current alive ♀ · ♂ (Din list), open cycles |
| `GET /api/forms/sale?farmId&date` | sections with current alive, farm total ("ouat · 6 secțiuni · 4 cu păsări · în viață ♀ 12 561 · ♂ 794") |
| `GET /api/forms/delivery?farmId` | active feed types, stock per type now |

**Creates** (all accept `clientRequestId`, `source: "manual" | "voice"`, `voiceTranscript?` + `voiceToken?`, `dryRun?`). `voiceTranscript` is stored only with a valid `voiceToken` (§8.9: HMAC over runId + userId + form + sha256(transcript), 10 min); otherwise it is ignored and the save is recorded as `manual`:
```http
POST /api/entries/daily
{ "farmId": "…", "date": "2026-11-18", "clientRequestId": "6f1c…", "source": "voice", "voiceTranscript": "pune 3 găini moarte la ha cinci punct unu", "voiceToken": "<from the prefill action>",
  "sections": [ { "sectionId": "…H5.1", "F": { "mortality": 1, "cull": null, "cullReason": null, "sexingError": null }, "M": { "mortality": 0 } },
                { "sectionId": "…H6.2", "noLosses": true } ],
  "eggs": [ { "sectionId": "…H5", "floor": 98, "cat1": 13, "cat2": 417, "cat3": 27, "hatching": 3415, "avgWeightG": 67.04 } ] }
→ 201 { "entry": { "id": "…", "kind": "daily", "version": "…" }, "warnings": [],
        "toast": "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0", "alive": { "F": 12561, "M": 794 }, "asOf": "…" }
```
Go types (package `entries`; the full handler is the worked example of §12.5):
```go
type DailyCreateRequest struct {
	httpx.WriteMeta                // clientRequestId, source, voiceTranscript, voiceToken, dryRun
	FarmID          uuid.UUID      `json:"farmId"`
	Date            timex.Day      `json:"date"`
	Sections        []DailySection `json:"sections"`
	Eggs            []DailyEggs    `json:"eggs"`
}
type DailySection struct {
	SectionID uuid.UUID `json:"sectionId"`
	NoLosses  bool      `json:"noLosses"` // "Fără pierderi azi": 0 rows for mortality and cull of the sexes present
	F         *DailySex `json:"F"`
	M         *DailySex `json:"M"`
}
type DailySex struct { // nil = field left empty → no row (empty is null, never 0)
	Mortality   *num.Num `json:"mortality"`
	Cull        *num.Num `json:"cull"`
	CullReason  *string  `json:"cullReason"`  // boala | schiopatare | altul
	SexingError *num.Num `json:"sexingError"` // signed
}
type DailyEggs struct { // sectionId = the egg unit (hall, or a section without parent and children)
	SectionID  uuid.UUID `json:"sectionId"`
	Floor      *num.Num  `json:"floor"`
	Cat1       *num.Num  `json:"cat1"`
	Cat2       *num.Num  `json:"cat2"`
	Cat3       *num.Num  `json:"cat3"`
	Hatching   *num.Num  `json:"hatching"`
	AvgWeightG *num.Num  `json:"avgWeightG"`
}
type EntryWriteResponse struct { // shared by every entry create / edit / correction / delete / undo / restore
	Entry    httpx.EntryRef      `json:"entry"`
	Warnings []httpx.Warning     `json:"warnings"`
	Toast    string              `json:"toast"`
	Alive    *httpx.SexPair[int] `json:"alive,omitzero"` // farm alive after the save (Zilnic, Transfer, Vânzare, corrections)
	AsOf     timex.LocalTime     `json:"asOf"`
}
```
`DailyCreateRequest.Validate` (level 1, §7.0b): date parses; ≥ 1 section or egg row; each `sectionId` once; counts integral and ≥ 0 except `sexingError` (Z2); `cullReason` in the enum; `noLosses` not combined with typed mortality/cull for the same section; egg counts integral ≥ 0, `avgWeightG` 20–120 (Z8). Everything else (Z1, Z3–Z7, Z9, §6.1) is level 2 inside the transaction.
```http
POST /api/entries/feed-daily
{ "farmId": "…", "date": "2026-11-18", "rows": [ { "sectionId": "…H5.1", "sex": "F", "feedTypeId": "…breeder2", "g": 160 }, "…8 rows…" ] }
→ 201 { "entry": {…}, "kgTotal": 2457.9, "byType": [ { "type": "Breeder 2", "kg": 2303.2, "stockAfter": 31664.9 }, { "type": "Cocoși", "kg": 154.7, "stockAfter": 1781.0 } ],
        "toast": "Hrană zilnică salvată · 8 rânduri · 2 457,9 kg · stoc Breeder 2: 31 665 kg", "warnings": [] }
```
```http
POST /api/entries/weighing
{ "farmId": "…telita", "date": "2026-02-24", "sectionId": "…H1", "sex": "F", "avgWeightG": 2553, "cvPct": 8.95, "uniformityPct": 74.1, "sampleSize": null }
→ 201 { "comparison": { "sapt": 21, "guideWeek": 20, "guideG": 2370, "diffPct": 7.72, "status": "abatere" },
        "warnings": [ { "code": "WEIGHT_GUIDE", "message": "H1 ♀: 2 553 g, cu 7,7% peste ghidul de 2 370 g de la sfârșitul SAPT 20 (prag ±5%). Verifică proba; poți salva." } ],
        "toast": "Cântărire salvată · H1 ♀ 2 553 g" }
```
```http
POST /api/entries/transfer
{ "date": "2026-02-28", "fromSectionId": "…telita-H2", "toSectionId": "…hirbovat-H7.2", "F": 4383, "M": 0, "dryRun": true }
→ 200 { "effects": [ { "farm": "Telița", "line": "Ciclu închis la Telița, sapt. 21", "sub": "ferma ajunge la 0 păsări", "cycle": "close" },
                     { "farm": "Hîrbovăț", "line": "Ciclul la Hîrbovăț continuă, sapt. 21", "cycle": "continue" } ],
        "sections": [ { "code": "H2", "farm": "Telița", "before": { "F": 4383, "M": 0 }, "after": { "F": 0, "M": 0 } } ], "warnings": [], "refusals": [] }
```
(The 28.02.2026 replay is the board example D83; with the real history Telița closes through the migration's register adjustment, §10.3.)
```http
POST /api/entries/sale
{ "farmId": "…", "date": "2026-11-18", "buyer": null, "note": null,
  "rows": [ { "sectionId": "…H6.1", "F": 676, "M": 201 }, { "sectionId": "…H6.2", "F": 1155, "M": 88 } ] }
→ 201 { "effects": [ { "line": "Ciclul la Hîrbovăț continuă, sapt. 59", "sub": "rămân ♀ 12 561 · ♂ 794" } ],
        "toast": "Vânzare înregistrată · ♀ 1 831 · ♂ 289 · Hîrbovăț · ciclul continuă" }
```
A sale that empties the farm without `"confirmEmptiesFarm": true` → `409 CONFIRM_REQUIRED` with the Dialog title/body (§6.6).
```http
POST /api/entries/delivery
{ "farmId": "…", "date": "2026-11-12", "feedTypeId": "…breeder2", "qtyKg": 9000, "note": null }
→ 201 { "toast": "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg · fără preț", "stockAfter": 25140 }
```

**Introdus azi** — `GET /api/entries/today` → the caller's entries created today, grouped by record type (D98): Mortalitate per section ♀ · ♂ (culls and sexing errors inside), Ouă per hall (with lay vs guide lamp), Hrană zilnică one row for the whole form (kg per type in the group head: "salvat la 08:44 · Breeder 2: 2 303,2 kg · Cocoși: 154,7 kg"), Cântărire per section × sex, Transfer, Vânzare, Livrare; each row `{ entryId, sectionId?, canEdit, canDelete, until: "23:59", modifiedAt? }`; head "Dr. Ceban · poți modifica până la 23:59". After 23:59 the keys come back disabled with the reason.

**Read, edit, correct, delete, undo, restore**
| Route | Who / rule | Body |
|---|---|---|
| `GET /api/entries/:id` | anyone who can see the farm | the whole save as its form (D172): every section × sex value, eggs, rows; `version`; author/time; `corrected`; `actions` for the caller |
| `PUT /api/entries/:id` | **edit** (author in the creation-day window, no reason, D70 A) or **correction** (manager/owner, `reason` required, D70 B, E1) — the server decides from caller + window; a vet outside the window → 403 | `{ ifVersion, values: <same shape as create>, reason?, dryRun?, source?, voiceTranscript?, voiceToken? }` (the transcript is kept only with a valid `voiceToken`, as on creates) |
| `DELETE /api/entries/:id` | author in the window (no reason) or manager/owner with `reason` | `{ reason?, sectionId?, dryRun? }` (`sectionId`: S15 per-section delete of Zilnic values, D98) |
| `POST /api/entries/:id/undo` | author in the window: undo the create (= delete, logged `undo`) | `{}` |
| `POST /api/entries/:id/restore` | the deleter in the same day (Toast Anulează after Șterge) | `{}` → §6.1 re-checked |
Correction dry-run response:
```json
{ "changes": [ { "key": "H5.2 · găini moarte", "old": 1, "new": 3 } ],
  "effects": [ "H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.", "Mortalitate · morți ♀ −5 → −7", "Hîrbovăț · ♀ azi 12 563 → 12 561" ],
  "refusals": [], "warnings": [ { "code": "MORTALITY_SPIKE", "message": "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva." } ] }
```
Saved correction Toast: "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281"; the record then reads "corectat · modificat de Ion Rusu · 18.11.2026, 10:12". Delete Toast: "Mortalitate ștearsă · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 568".
Go types for edit / Corectează (package `entries`):
```go
type EntryUpdateRequest struct {
	IfVersion       string          `json:"ifVersion"`
	Values          json.RawMessage `json:"values"` // decoded by the stored entry's kind into DailyCreateRequest, FeedDailyCreateRequest, …
	Reason          *string         `json:"reason"` // Motiv: required for a correction, ignored for a same-day author edit
	DryRun          bool            `json:"dryRun"`
	Source          string          `json:"source"`
	VoiceTranscript *string         `json:"voiceTranscript"`
	VoiceToken      *string         `json:"voiceToken"`
}
type CorrectionPreview struct { // dryRun response (200)
	Changes  []Change        `json:"changes"`
	Effects  []string        `json:"effects"`
	Refusals []httpx.Refusal `json:"refusals"`
	Warnings []httpx.Warning `json:"warnings"`
}
type Change struct { // also the element type of audit_log.changes (jsonb)
	Key     string   `json:"key"` // "H5.2 · găini moarte"
	Section string   `json:"section,omitzero"`
	Sex     string   `json:"sex,omitzero"`
	Field   string   `json:"field,omitzero"` // mortality, cull, cullReason, sexingError, cat1, g, …
	Old     *num.Num `json:"old"`            // null = was empty
	New     *num.Num `json:"new"`            // null = cleared
}
```
The handler loads the entry, decodes `Values` with `DisallowUnknownFields` into the create type of its kind (the date and farm in `values` must equal the stored ones: the date is not editable, D101), runs the same level-1 `Validate()`, then `entries.Service.Update(ctx, principal, id, req)` decides edit vs correction (§3.2), checks `IfVersion` (`409 STALE`), requires a trimmed `Reason` of 1–200 for a correction (§6.15 messages), computes the diff (`[]Change`, "Nu ai schimbat nicio cifră." when empty), and runs the same transaction as a create (§12.5) with `action = update`, `entry.corrected = true` for a correction. A saved edit/correction returns `EntryWriteResponse`.

### 7.9 Cycles (B4)
| Route | Body | Result |
|---|---|---|
| `POST /api/farms/:id/cycles` (manager, owner) | `{ date, guideId, seriesId?, reason, dryRun? }` — `seriesId` (D201): an active series, default the open cycle's; a closed series → 409 `SERIES_CLOSED`; a series that already has an open cycle at this farm → 422 *(proposed)* "Seria 2027 are deja un ciclu deschis la Telița." | closes the open cycle at `date − 1`, opens a manual one (§6.4 C3–C4); the dry-run names the series: "Ciclu nou la Telița · Seria 2027 · sapt. {n} de la {date}"; preview note "Ciclul curent (sapt. 59) se închide pe 18.11.2026. Săptămâna rămâne cea de vârstă a păsărilor." + "SAPT 59 · ciclu nou din 18.11.2026" |
| `PATCH /api/cycles/:id/guide` (manager, owner) | `{ guideId \| null }` | recomparison from the cycle start; audit `cycle_guide_change` |

### 7.10 Istoric modificări (E2, S17; D104–D106, D178)
`GET /api/audit?period=weeks:4|weeks:13|all|interval:a,b&userId=&farmId=&types=&entryId=&cursor=` (manager: only rows whose `farm_ids` overlap his farms; others 403 "Nu ai acces la Istoric modificări.").
- Period default "4 săpt." = the 28 days ending today (calendar, not age weeks: the list spans farms and series): 22.10–18.11.2026.
- Rows, newest first, one per save action (D178): `{ id, at, user: { name: "Ion Rusu", role: "manager" }, action, summary: "Mortalitate · 4 secțiuni", recordDate: "2026-11-17", sapt: 59, farm: "Hîrbovăț", changes: [ { key: "H5.2 · găini moarte", old: 1, new: 3 } ], reason: "Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.", source: "manual", voiceTranscript: null, link: "/ferme/hirbovat?zi=2026-11-17" }`.
- `chips` = record types with records (not only changes) in the period (D104, D178): a move between sections counts as Transfer, a priced delivery as Preț hrană.
- `users` = the Utilizator filter list ("Proprietar (tu)", "Manager (tu)" for the caller).
- `entryId` = the one-record view opened from "modificat de" (Banner "Vezi doar modificările înregistrării Mortalitate · 4 secțiuni · ziua 17.11.2026 · Hîrbovăț, din tot istoricul.").
- Interval end in the future → "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."

### 7.11 Rapoarte · Producție (D3, S14; D92–D94, D176, D183–D188)
`GET /api/reports/production?farmId&cycleId&period` → `{ filters (laying farms of the caller; none → state "fără fermă de ouat" with a link to Rapoarte › Hrană), table: <§7.4 with view=report-production>, readout: { "% ouat" weekly series vs guide }, stats: [ Vârf de ouat (SAPT 30, 90,3%), Masă medie ou în ciclu (egg-weighted), Ouă în ciclu, % incubabile în ciclu, Mortalitate cumulată ♀, Hrană g/găină ] }`. Seria 2025 (past cycle) keeps the same shape from its imported eggs (no hens, no lay %, no feed: those columns are omitted, D92; "· 4 din 43" weeks).

### 7.12 Rapoarte · Hrană (D4, S11; D49, D64, D163)
`GET /api/reports/feed?farmId=all|…&period` →
```json
{ "stock": [ { "farm": "Hîrbovăț", "type": "Breeder 2", "forSex": "F", "kg": 31664.9, "daysLeft": 13.2, "avgDaily": 2392.2, "window": ["2026-11-16", "2026-11-18"], "runOut": "2026-12-01", "lamp": "in_norma" },
             { "farm": "Hîrbovăț", "type": "Cocoși", "forSex": "M", "kg": 1781.0, "daysLeft": 10.4, "avgDaily": 170.9, "runOut": "2026-11-28", "lamp": "in_norma" },
             { "farm": "Telița", "lamp": "oprit", "reason": "fără ciclu deschis", "sub": "ultimul ciclu: Seria 2026, închis pe 28.02.2026" } ],
  "deliveries": { "total": 82, "unpriced": { "count": 4, "kg": 18160, "dates": ["2026-02-17", "2026-03-14", "2026-11-12"] },
                  "rows": [ { "id": "…", "date": "2026-11-13", "farm": "Hîrbovăț", "type": "Breeder 2", "forSex": "F", "kg": 25640, "pricePerKg": 7.7613, "totalLei": 199000, "pricedBy": "din registru", "actions": { "changePrice": true } },
                            { "id": "…", "date": "2026-11-12", "farm": "Hîrbovăț", "type": "Breeder 2", "kg": 9000, "pricePerKg": null, "totalLei": null, "tag": "fără preț" } ],
                  "nextCursor": "…" },
  "cost": { "asOf": "2026-11-18", "cycles": [
      { "label": "Telița · Seria 2026", "span": "07.10.2025–28.02.2026", "pricedCount": 13, "kg": 263740, "lei": 2031768, "leiPerBird": 64.32, "leiPerEgg": null, "eggReason": "creștere, fără ouă", "unpriced": null },
      { "label": "Hîrbovăț · Seria 2026", "span": "din 22.02.2026 · SAPT 59", "pricedCount": 65, "kg": 1174480, "lei": 9562437, "leiPerBird": 319.74, "leiPerEgg": 1.88,
        "unpriced": { "count": 4, "kg": 18160, "span": "17.02–12.11.2026" } } ],
    "total": { "label": "Total Seria 2026", "kg": 1438220, "lei": 11594205, "leiPerBird": 367.02, "leiPerEgg": 2.29, "excludedKg": 18160 } } }
```
- Viewer: `pricePerKg`, `totalLei`, `cost` omitted entirely (not nulls) — never sent (D64, D200). Vet: 403.
- Consum = `GET /api/table?view=report-feed&farmId&period` (§7.4). Its foot note: "g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre (♀ după găini, ♂ după cocoși). Toleranță ±10%."

### 7.13 Prețuri hrană (C6, S12)
| Route | Body | Result |
|---|---|---|
| `GET /api/deliveries?priced=false` (owner) | – | unpriced deliveries oldest first: date, farm, type, kg |
| `POST /api/deliveries/prices` (owner) | `{ items: [ { id, pricePerKg } \| { id, totalLei } ] }` | sets prices (partial batch ok); audit `price_set`; Toast "Prețuri salvate · 4 livrări · 18 160 kg intră în cost" + `undoToken` (same day) |
| `POST /api/deliveries/prices/undo` | `{ undoToken }` | back to "fără preț" (D122) |
| `PATCH /api/deliveries/:id/price` (owner) | `{ pricePerKg \| totalLei, reason }` | change with Motiv (D69); audit `price_change` |

### 7.14 Export .xlsx (D5, D191; §10.1)
`GET /api/export?<same params as /api/table without before/count>&format=xlsx` → `200 application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `Content-Disposition: attachment; filename="Hirbovat-Pasari-SAPT20-59.xlsx"`. Also `GET /api/export/production` and `/api/export/feed` for the report pages (one sheet per table/part). Viewer exports never contain lei.

### 7.15 Serii (B3, S18; D107–D109, D179)
| Route | Body / result |
|---|---|
| `GET /api/series` (owner) | series rows `{ name, year, status, ageAtPlacementDays, saptSpan: [1, 59], ageToday: 407, cycles: [ { farm, span: "07.10.2025–28.02.2026", saptSpan: [1, 21], open: false }, { farm: "Hîrbovăț", span: "din 22.02.2026", saptSpan: [20, 59], open: true } ], canClose, closeReason }` + the track |
| `POST /api/series` | `{ name, year, farmId, date, ageAtPlacementDays, rows: [ { sectionId, F, M } ], manualCycle?: { reason }, dryRun? }` → creates series + placement entry + auto cycle; dry-run returns the live line "La populare: vârstă 0 zile · SAPT 1." and "Ciclu nou la Telița, sapt. 1" / "ghid Ross 308 creștere" |
| `PATCH /api/series/:id/age` | `{ ageAtPlacementDays, reason, dryRun? }` → preview "Toate săptămânile Seriei 2026 se renumerotează (0 → 7 zile: azi SAPT 59 → SAPT 60) și se recompară cu ghidurile." |
| `POST /api/series/:id/close` | `{}` → only when no open cycle (§6.14 S8); afterwards read-only |

### 7.16 Ferme și secțiuni (B1, S19; D75, D110–D112)
| Route | Body |
|---|---|
| `GET /api/farms` (owner) | farms with type, guide, sections tree (active first, retired last), cycle lamp, `typeLocked` + reason |
| `POST /api/farms` | `{ name, type, guideId? }` |
| `PATCH /api/farms/:id` | `{ name?, type?, guideId? }` (guide → next cycle; Toast "Telița salvată · ghid Ross 308 creștere de la următorul ciclu al fermei") |
| `POST /api/farms/:id/sections` | `{ code, parentId? }` (each tree edit saves alone, D75) |
| `PATCH /api/sections/:id` | `{ code? }` rename (allowed with birds) or `{ moveBefore: sectionId \| null }` reorder among siblings |
| `POST /api/sections/:id/retire` | `{}` → §6.11 F8/F9; Dialog text "H5-2 nu mai apare în formulare și pe pagina fermei. Înregistrările ei din Seria 2026 rămân în rapoarte. O secțiune retrasă nu se mai poate reactiva." |

### 7.17 Ghiduri (B2, S20; D76, D113–D115, D180)
| Route | Body / result |
|---|---|
| `GET /api/guides` | `{ name, type, saptSpan: [19, 65], columns: [...], missing: ["greutatea", "masa oului"], usedBy: [ { farm, series, span, open } ] }` |
| `GET /api/guides/:id` | detail: used-by, per-indicator summary (SAPT range + gaps, first/max/last), weekly rows (♀/♂ pairs) |
| `POST /api/guides/preview` | multipart `file` (.xlsx/.csv) or `{ text }` (pasted TSV) + `type` + `replaceGuideId?` → `{ source: { name, format, type }, counts: { read: 51, valid: 48, refused: 3 }, rows: [ { line: 7, sapt: 4, sex: "F", values: {…}, status: "refused", errors: [ { cell: "body_weight_g", message: "Greutatea „#REF!” nu e număr. Scrie greutatea în grame." } ] } ], fileErrors: [], warnings: [ week-1 banner ], info: "Lipsurile nu blochează salvarea. …", canSave: false }` |
| `POST /api/guides` | `{ name, type, rows: [ { sapt, sex, feedG, bodyWeightG, mortalityCumPct, layPct, hatchingPct, eggWeightG } ], sourceFilename }` → re-validated server-side |
| `PUT /api/guides/:id/rows` | `{ rows, sourceFilename }` → replace in place (type fixed while used); Toast lists the cycles that compare against it |
| `DELETE /api/guides/:id` | unused only (§6.12 G15); Dialog "Ghidul și cele 94 de rânduri ale lui se șterg. Niciun ciclu nu îl folosește, deci nicio comparație nu se schimbă. …" |

### 7.18 Tipuri de hrană (S21; D116, D117)
Package `feedtypes` (§12.3). `GET /api/feed-types` (owner: all with `usedAt` farms, stock per farm; others: active only through bootstrap) · `POST /api/feed-types { name, forSex }` · `PATCH /api/feed-types/:id { name?, forSex?, active? }` (§6.13) · `DELETE /api/feed-types/:id` (unused only).

### 7.19 Utilizatori (A2, S22; D78, D118–D120)
| Route | Body / result |
|---|---|
| `GET /api/users` (owner) | `{ name, login, role, farms ("toate fermele" for owners), lastLoginAt, active, isSelf }`, active first; counts "4 utilizatori · 3 activi · 1 dezactivat" |
| `POST /api/users` | `{ name, login, role, farmIds }` → `201 { user, oneTimePassword: "mnpq6789rstu" }` (shown once, D78) |
| `PATCH /api/users/:id` | `{ name?, role?, farmIds? }` (login immutable; §6.10 U6) |
| `POST /api/users/:id/reset-password` | → `{ oneTimePassword }`; sessions revoked |
| `POST /api/users/:id/deactivate` · `/reactivate` | sessions revoked on deactivate; reactivate keeps the old password |
`Cache-Control: no-store` on every response that carries `oneTimePassword`; the value is never logged.

### 7.19b Same-day Anulează for setup saves (D108, D111, D116, D118, D122)
Every Toast of a setup save returns `undoToken` (HMAC-signed: audit row id + user id, valid until 23:59 of the save day). `POST /api/undo { undoToken }` (same user; owner or the manager who saved) reverts that one action and writes an audit row `undo`:
| Save | Undo does | Refused when *(proposed copy "Nu se mai poate anula: {obiect} are deja înregistrări.")* |
|---|---|---|
| farm create / edit | delete the new farm / restore name, type, guide | the farm has sections with records |
| section add / rename / reorder / retire | delete / old code / old order / un-retire (the one same-day exception to D75) | the new section has records |
| feed type add / edit / retire / reactivate | delete / old values / reactivate / retire again | the new type has deliveries or allocations; retiring again with stock > 0 |
| user add / edit / deactivate / reactivate | add → hard delete if the user never signed in and has no records, else deactivate; edit → old role/farms (sessions revoked again); deactivate ↔ reactivate | – |
| series create (D108) | delete series, its placement entry and auto cycle | any other record is dated in that cycle |
| manual new cycle | delete the manual cycle and reopen the previous one (`end_date = null`) | – (records follow by date) |
| prices (S12, D122) | §7.13 `prices/undo` | – |
| entries | §7.8 `undo` / `restore` | – |

### 7.20 Voice and live
- Voice: `POST /api/voice/stt` (audio), `GET /api/voice/stream` (WebSocket, optional live partials), `POST /api/voice/command` (transcript → action) — §8.
- Live: `GET /api/live` (SSE) — §9.

---

## 8. Voice commands (F1, D199 + notes 1–4; spec §8)

### 8.1 Architecture
```
browser (VoicePanel)                         Go API (internal/voice)                          providers
 mic → MediaRecorder (opus/webm; iOS mp4)  ──► POST /api/voice/stt  (or WS /api/voice/stream) ──► STT (ro-RO, streaming, EU)
                                           ◄── { transcript }  (partials over WS)
 transcript (editable) + screen context    ──► POST /api/voice/command ──► prompt + 81 intent tools ──► LLM (tool call, temp 0)
                                               validate: refs → role → farm → time window → never-list → existing saves
                                           ◄── action { navigate | answer | prefill | clarify | refuse | confirm | export | not_understood | client }
 client performs it through the same screen logic; data changes only through the normal endpoints (§7) when the USER presses Salvează
```
- **Server STT only.** The browser Web Speech API is not used: in Chrome it sends audio to a third party with no contract, which breaks the privacy rule ("not by the provider", open question 13), and it is missing on Firefox and unreliable in iOS home-screen apps. A typed command ("Scrie comanda") skips STT entirely.
- **The LLM never writes and never invents figures.** It only picks one intent and its parameters. Answers (Q-intents) are composed by the server from the same services the screens use (§8.7). Writes from voice don't exist: P-intents return a prefill; the user saves with the normal endpoint, which re-validates everything.

### 8.2 Endpoints
| Route | Request | Response |
|---|---|---|
| `POST /api/voice/stt` | `Content-Type: audio/webm;codecs=opus` (or `audio/mp4`, `audio/ogg`, `audio/wav`), body ≤ 1 MB, ≤ 20 s; header `X-Voice-Run: <uuid>` | `{ "transcript": "pune 3 găini moarte la H5.1", "durationSec": 3.8 }`; `503 VOICE_UNAVAILABLE` on provider failure |
| `GET /api/voice/stream` (WebSocket, optional; the upgrade is refused with 403 unless `Origin` equals the app origin and the session cookie is valid; one stream per user) | binary audio chunks (250 ms), then `{"type":"end"}` | `{"type":"partial","text":"pune trei găini"}` … `{"type":"final","text":"…"}` (the live transcript of the "Ascult…" state) |
| `POST /api/voice/command` | `{ runId, transcript, edited: bool, context }` or `{ runId, pending, choice }` (a clarification chip) | `{ runId, transcript, normalized, action }` (§8.5) |
Audio lives only in request memory (never on disk, never in logs) and is discarded after the STT call; the client keeps it in memory only to retry a failed transcription (spec §8 "Reîncearcă").

**Go realisation of the voice endpoints** (package `voice`):
- `POST /api/voice/stt`: `http.MaxBytesReader(w, r.Body, 1<<20)` and `io.ReadAll` into a `[]byte` that is passed to `STTProvider.Transcribe` and then dropped (no temp file, no logging of the body); `Content-Type` must be one of the four audio types; `X-Voice-Run` must parse as a UUID; the provider call runs under `context.WithTimeout(r.Context(), 6*time.Second)`; duration over 20 s → 400 (the provider's reported duration or the container header, verify per provider).
- `GET /api/voice/stream` (WebSocket, library `github.com/coder/websocket`): the auth middleware runs **before** the upgrade (no session → 401, no upgrade); then `websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{appHost}})` — the library refuses a cross-origin `Origin` unless it matches the patterns, which is the CSWSH check (§11.4); verify option names at implementation. One stream per user: a `sync.Map` of user id → active stream; a second upgrade while one is open → `409` before `Accept`. Limits: `conn.SetReadLimit(64 << 10)` per message (a 250 ms opus chunk is a few KiB), ≤ 1 MiB of audio in total and ≤ 20 s wall time per stream (`context.WithTimeout`), 25 s idle (each `Read` gets its own deadline context), then `conn.Close(websocket.StatusPolicyViolation, "limit")`. The handler clears the server's write deadline for this connection (`http.NewResponseController(w).SetWriteDeadline(time.Time{})` before `Accept`; §9 explains why). Messages out: `wsjson.Write(ctx, conn, partial)`; binary frames in go straight to `STTStream.Write`.
- `POST /api/voice/command`: `VoiceCommandRequest` → pipeline (§8.6) → `VoiceCommandResponse`; whole run under a 10 s context (LLM 8 s inside it).
```go
type VoiceCommandRequest struct {
	RunID      uuid.UUID     `json:"runId"`
	Transcript string        `json:"transcript"` // ≤ 500 characters (runes), else 400
	Edited     bool          `json:"edited"`
	Context    *VoiceContext `json:"context"` // required unless Pending is set
	Pending    *string       `json:"pending"` // signed clarify token (60 s)
	Choice     *string       `json:"choice"`  // the chip label chosen
}
type VoiceContext struct {
	Screen  string       `json:"screen"` // "S6"
	Tab     *string      `json:"tab"`
	Panel   *string      `json:"panel"`
	Sheet   *string      `json:"sheet"`
	Scope   VoiceScope   `json:"scope"`   // seriesId, farmId, cycleId, sectionId (all *uuid.UUID)
	Filters VoiceFilters `json:"filters"` // period, grain, types, detail, loadedSaptFrom
	Form    *VoiceForm   `json:"form"`    // name, date, mode, values map[string]json.RawMessage
	Client  VoiceClient  `json:"client"`  // device, now (informational only: the server uses its own clock)
}
type VoiceCommandResponse struct {
	RunID      uuid.UUID `json:"runId"`
	Transcript string    `json:"transcript"`
	Normalized string    `json:"normalized"`
	Action     Action    `json:"action"`
}

// Action is one of the concrete action structs of §8.5; each writes its own "type".
type Action interface{ actionType() string }
type NavigateAction struct {
	Type  string            `json:"type"` // "navigate"
	Route string            `json:"route"`
	Query map[string]string `json:"query,omitzero"`
	Toast string            `json:"toast"`
	Undo  bool              `json:"undo"`
}
type PrefillAction struct {
	Type         string            `json:"type"` // "prefill"
	Form         string            `json:"form"`
	Route        string            `json:"route"`
	Query        map[string]string `json:"query"`
	Mode         string            `json:"mode"`    // new | edit | correct
	EntryID      *uuid.UUID        `json:"entryId"` // null for mode new
	Fields       []PrefillField    `json:"fields"`
	Banner       Banner            `json:"banner"`
	Normalized   string            `json:"normalized"`
	Focus        string            `json:"focus"`
	ContinueWith *ContinueWith     `json:"continueWith,omitzero"`
	VoiceToken   string            `json:"voiceToken"`
}
type PrefillField struct {
	Path     string          `json:"path"`     // "sections.H5.1.F.mortality"
	Value    json.RawMessage `json:"value"`    // number, string or bool as the form field takes it
	Previous json.RawMessage `json:"previous"` // null when the field was empty
	Label    string          `json:"label"`
}

// AnswerAction, ClarifyAction, RefuseAction, ConfirmAction, ExportAction, NotUnderstoodAction, ClientAction follow the
// JSON of §8.5 field for field in the same way.
```

### 8.3 Context sent with every command (client → server)
```json
{ "screen": "S6", "tab": null, "panel": null, "sheet": null,
  "scope": { "seriesId": "…", "farmId": "…hirbovat", "cycleId": "…", "sectionId": null },
  "filters": { "period": "cycle", "grain": "day", "types": ["mortalitate"], "detail": "farm", "loadedSaptFrom": 56 },
  "form": { "name": "zilnic", "date": "2026-11-18", "mode": "new", "values": { "H5.1.F.mortality": 2 } },
  "client": { "device": "phone", "now": "2026-11-18T08:20:00+02:00" } }
```
The server adds (never trusting the client for these): user name and role, the user's farms with section codes (halls with compartments), today `2026-11-18`, today's SAPT per open cycle (59), the last 14 days with weekdays (for "ieri", "pe 13 noiembrie"), active feed types with for-sex, guides (owner), series names, user names (owner only), the role's allowed intents.

### 8.4 Prompt and tools
- **System prompt** (static part cached with prompt caching): what Impex Line is; "pick exactly one tool"; "never invent or compute a figure; copy spoken numbers as digits"; Romanian number words (§8.8); section normalisation ("ha cinci punct unu", "hala cinci compartimentul unu", "H cinci unu" → `H5.1`; "ha cinci liniuță unu" → `H5-1`; "hala cinci" / "H5" → the hall `H5`); farms ("Hirbovat" → Hîrbovăț, "Telita" → Telița); sex words (găini/găină/femele → F, cocoși/cocoș/masculi → M, păsări → both); dates relative to today; weeks ("săptămâna asta" = current SAPT, "săptămâna trecută" = SAPT − 1); record words (moarte/mortalitate/au murit; sacrificate; eroare de sexare; transfer/mută; vânzare/am vândut; hrană/grame/porție; livrare/am primit; cântărire/greutate/CV/uniformitate/probă; ouă/categoria unu/incubabile/jos/masa medie; preț/lei pe kilogram); "a phrase naming two forms: fill the first, put the second in `alsoAsked`"; "if a needed value is ambiguous call `clarify`; if nothing fits call `neinteles`". Dynamic part: the context above.
- **Tools**: one tool per intent (81, named by id, e.g. `P01_mortalitate`) plus `clarify` and `neinteles`; `tool_choice` = any (must call one); `temperature = 0`; strict JSON schemas (no extra properties). Enum values that depend on the user (farm names, section codes, feed types, guides, series, user names) are injected per request as `enum` lists so the model can't invent an entity.
- **Typing rule**: schemas carry types, not business ranges (grams 1–300, counts ≤ alive, age 0–140 are checked by the form after prefill), so a misheard number reaches the form and shows the form's own refusal instead of being silently "fixed" by the model.

**Shared definitions** (JSON Schema `$defs`; `?` = optional property, everything else required; `additionalProperties: false` everywhere):
```json
{ "Farm":     { "type": "string", "enum": ["Hîrbovăț", "Telița"] },
  "Section":  { "type": "string", "enum": ["H5", "H5.1", "H5.2", "H6", "H6.1", "H6.2", "H7", "H7.1", "H7.2", "H1", "H2", "H3", "H5-1", "H5-2"] },
  "Sex":      { "type": "string", "enum": ["F", "M", "both"] },
  "Date":     { "type": "string", "format": "date" },
  "Tab":      { "type": "string", "enum": ["rezumat", "pasari", "hrana", "greutate", "productie"] },
  "Period":   { "type": "object", "properties": { "kind": { "enum": ["cycle", "weeks", "interval", "all"] }, "weeks": { "type": "integer" }, "from": { "$ref": "#/$defs/Date" }, "to": { "$ref": "#/$defs/Date" } }, "required": ["kind"] },
  "Week":     { "type": "object", "properties": { "sapt": { "type": "integer" }, "relative": { "enum": ["current", "previous"] } } },
  "Cycle":    { "type": "object", "properties": { "kind": { "enum": ["current", "previous", "by_series"] }, "series": { "type": "string" } }, "required": ["kind"] },
  "Record":   { "type": "string", "enum": ["mortalitate", "sacrificare", "eroare_sexare", "transfer", "populare", "vanzare", "hrana", "livrare", "cantarire", "oua", "pret"] },
  "Form":     { "type": "string", "enum": ["zilnic", "hrana_zilnica", "cantarire", "transfer", "vanzare", "livrare"] },
  "Count":    { "type": "integer" },
  "Num":      { "type": "number" },
  "Reason":   { "type": "string", "enum": ["boala", "schiopatare", "altul"] },
  "Motiv":    { "type": "string" },
  "FeedType": { "type": "string", "enum": ["Breeder 2", "Cocoși"] },
  "Delivery": { "type": "object", "properties": { "date": { "$ref": "#/$defs/Date" }, "farm": { "$ref": "#/$defs/Farm" }, "feedType": { "$ref": "#/$defs/FeedType" }, "kg": { "$ref": "#/$defs/Num" } }, "required": ["date"] } }
```
Section codes are unique only **within** a farm (Telița H6 is a section, Hîrbovăț H6 a hall): the enum is the union of the caller's codes and the server resolves a code with the farm (named, or from context; a code that exists in two of the caller's farms and no farm given → `clarify` "La ce fermă?").

Full expansion of one intent (the rest follow the same mapping):
```json
{ "name": "P01_mortalitate", "description": "Pune mortalitatea în Zilnic (nu salvează).", "input_schema": { "type": "object", "additionalProperties": false,
  "properties": { "farm": { "$ref": "#/$defs/Farm" }, "date": { "$ref": "#/$defs/Date" },
    "items": { "type": "array", "minItems": 1, "items": { "type": "object", "additionalProperties": false,
               "properties": { "section": { "$ref": "#/$defs/Section" }, "sex": { "$ref": "#/$defs/Sex" }, "count": { "$ref": "#/$defs/Count" } }, "required": ["section", "sex", "count"] } },
    "alsoAsked": { "$ref": "#/$defs/Form" } },
  "required": ["items"] } }
```

### 8.5 The 81 intents: parameters, server action, roles
Mode: **I** imediat (navigate/filter + Toast with Anulează; export without) · **R** răspuns · **P** precompletează · **C** confirmă. Roles as spec §8 (vet · viewer · manager · owner; "viewer+" = viewer, manager, owner). Every parameter falls back to the context defaults (§8.6). `alsoAsked: Form?` is accepted on every P-intent.

| Id · tool | Parameters | Server builds | Mode · roles |
|---|---|---|---|
| N01 acasa | – | navigate `/acasa` | I · all |
| N02 inregistreaza | – | client `open_register_menu` | I · vet, manager, owner |
| N03 deschide_formular | form: Form, farm?, date? | navigate to the blank form | I · per form (§3) |
| N04 seria | series?: string, tab?: Tab | navigate S3 | I · viewer+ |
| N05 ferma | farm?, cycle?: Cycle, tab?: Tab | navigate S4 | I · viewer+ |
| N06 sectiune | section, farm?, tab? | navigate S5 (hall → hall page) | I · viewer+ |
| N07 tab | tab | navigate current scope + tab | I · viewer+ |
| N08 rapoarte_productie | farm?, period?: Period | navigate S14 | I · viewer+ |
| N09 rapoarte_hrana | part?: stoc\|livrari\|consum\|cost, farm? | navigate S11#part (cost: manager, owner) | I · viewer+ |
| N10 preturi | – | navigate S12 | I · owner |
| N11 introdus_azi | – | navigate S15 | I · vet, manager, owner |
| N12 lipsesc_date | farm?, date? | navigate S2 focus Lipsesc date (owner/viewer: De verificat) | I · all |
| N13 istoric | – | navigate S17 | I · manager, owner |
| N14 setari_pagina | page: utilizatori\|ghiduri\|tipuri_hrana\|ferme\|serii | navigate S18–S22 | I · owner |
| N15 setari_element | page, item: string | navigate to the item panel/detail | I · owner |
| N16 inapoi | what: back\|close_panel | client command | I · all |
| N17 cont | – | client `open_account` | I · all |
| F01 perioada | period: Period | navigate with `period` ("8 săpt." = weeks:8; on S3 → clarify Telița · Hîrbovăț) | I · viewer+ |
| F02 pas | grain: zi\|sapt | navigate with grain | I · viewer+ |
| F03 detaliu | detail: ferma\|hale\|compartimente, section? | navigate with detail | I · viewer+ |
| F04 tipuri | types: Record[], mode: only\|add\|remove | navigate with chips (phone: single) | I · viewer+ |
| F05 ciclu | cycle: Cycle, farm? | navigate S4 cycleId | I · viewer+ |
| F06 serie | series: string | navigate series selector | I · viewer+ |
| F07 ferma_rapoarte | farm | navigate S14/S11 farm filter | I · viewer+ |
| F08 incarca_saptamani | times?: integer, untilStart?: boolean | client `load_more` × n | I · viewer+ |
| F09 saptamana | sapt: integer, open: boolean | client `toggle_week` | I · viewer+ |
| F10 zi | date, section?, tab? | navigate with the day loaded, open, focused | I · viewer+ |
| F11 inregistrare | record: Record, date?, section? | navigate to the row (Jurnal / S11 Livrări) | I · viewer+ |
| F12 filtre_istoric | user?: string, farm?, types?: Record[], period? | navigate S17 filtered | I · manager, owner |
| F13 fara_pret | – | navigate S11 Livrări fără preț (owner: S12) | I · manager, owner |
| F14 reset_filtre | – | navigate Ciclul curent · Zi · Fermă · all chips | I · viewer+ |
| F15 mai_vechi | – | client `journal_older` | I · viewer+ |
| Q01 pasari_in_viata | farm?, section?, sex?, date? | answer + open S4/S5 Rezumat (vet: own sections today, opens Zilnic) | R · all |
| Q02 ouat_ghid | farm?, hall?: Section, week?: Week, date? | answer + open S14 | R · viewer+ |
| Q03 mortalitate | farm?, section?, period?, sex?, cumulative?: boolean | answer + S4 Păsări chip Mortalitate | R · viewer+ |
| Q04 hrana_data | farm?, section?, date?, sex?, measure?: grame\|kg | answer + S4 Hrană (vet: S7) | R · all |
| Q05 stoc | farm?, feedType?, measure?: kg\|zile | answer + S11 Stoc | R · viewer+ |
| Q06 greutate | section, farm?, sex? | answer + S5 Greutate | R · viewer+ |
| Q07 oua_categorii | farm?, hall?, date?, week?, category?: total\|incubabile\|jos\|cat1\|cat2\|cat3 | answer + S4 Producție, Detaliu Hale | R · viewer+ |
| Q08 cost | farm?, cycle?, measure?: total\|pe_pasare\|pe_ou | answer + S11 Cost | R · manager, owner |
| Q09 saptamana_varsta | farm? | answer, no navigation | R · all |
| Q10 ce_lipseste | farm?, date? | answer + S2 Lipsesc date / De verificat | R · all |
| Q11 de_verificat | farm? | answer + S2 / S4 De verificat | R · viewer+ |
| Q12 livrari_fara_pret | – | answer + S12 (manager: S11 Livrări) | R · manager, owner |
| Q13 ce_am_salvat | form?, date? | answer + S15 | R · vet, manager, owner |
| Q14 cine_a_modificat | record?: Record, date?, section?, farm? | answer + S17 filtered | R · manager, owner |
| Q15 ciclu_ghid | farm?, ask?: ghid\|inceput\|ambele | answer + S4 header | R · viewer+ |
| Q16 sacrificari_vanzari_transferuri | type: sacrificare\|vanzare\|transfer, farm?, section?, period?, sex? | answer + S4/S5 Păsări with that chip | R · viewer+ |
| Q17 livrari_perioada | farm?, feedType?, period? | answer + S11 Livrări (lei: manager, owner) | R · viewer+ |
| P01 mortalitate | farm?, date?, items: [{section, sex, count}] | prefill S6 | P · vet (today), manager, owner |
| P02 sacrificare | farm?, date?, items: [{section, sex, count, reason?: Reason}] | prefill S6 (+ clarify "Ce motiv are sacrificarea?" → Boală · Șchiopătare · Altul) | P · same |
| P03 eroare_sexare | farm?, date?, items: [{section, sex, signedCount: integer}] | prefill S6, Mai mult open | P · same |
| P04 oua | farm?, date?, hall: Section, floor?, cat1?, cat2?, cat3?, hatching?: Count, avgWeightG?: Num | prefill the hall's Ouă group | P · same |
| P05 fara_pierderi | farm?, date?, sections?: Section[], all?: boolean | prefill "Fără pierderi azi" | P · same |
| P06 grame | farm?, date?, rows: [{section?: Section, sex, g: Num}] (no section = every row of that sex) | prefill S7 rows (count as confirmed) | P · same |
| P07 copiaza_ieri | farm?, date? | prefill S7 with yesterday | P · same |
| P08 tip_hrana | farm?, date?, rows: [{section?, sex?, feedType}] | prefill S7 Select | P · same |
| P09 cantarire | section, farm?, date?, sex?, weightG?, cvPct?, uniformityPct?, sampleSize? | prefill S8 | P · same |
| P10 transfer | from: Section, to: Section, fromFarm?, toFarm?, date?, F?: Count, M?: Count, all?: boolean | prefill S9 + dry-run preview | P · manager, owner |
| P11 vanzare | farm?, date?, rows?: [{section, F?, M?, all?}], allSections?: boolean, totalF?, totalM?, buyer?, note? | prefill S10 (an unsplit total goes to the Banner: "Ai spus ♀ 1 831 · ♂ 289 în total; scrie-le pe secțiuni.") | P · manager, owner |
| P12 livrare | farm?, date?, feedType?, sex?, kg?: Num, note? | prefill S13 (never a price) | P · manager, owner |
| P13 pret_livrare | delivery: Delivery, pricePerKg?: Num, totalLei?: Num | prefill the S12 row | P · owner |
| P14 modifica_pret | delivery: Delivery, pricePerKg?, totalLei?, motiv?: Motiv | prefill the S11 Modifică row | P · owner |
| P15 corecteaza | record: {type: Record, date, section?, farm?}, changes: [{section?, sex?, field: CorrField, value: Num, previous?: Num}], motiv?; `CorrField` enum: `mortalitate`, `sacrificare`, `motiv_sacrificare`, `eroare_sexare`, `jos`, `cat1`, `cat2`, `cat3`, `incubabile`, `masa_medie_ou`, `g_pasare`, `tip_hrana`, `masa_medie`, `cv`, `uniformitate`, `proba`, `transfer_gaini`, `transfer_cocosi`, `vanzare_gaini`, `vanzare_cocosi`, `livrare_kg` (the server refuses a field that the record's kind doesn't have) | prefill S16 for the whole save (marked "modificat" + VoiceMark) | P · manager, owner |
| P16 editeaza_azi | form: Form, + the matching P01–P09 parameters | prefill the form in Editează (own save from today only; else → P15) | P · vet (own, until 23:59), manager (own), owner (own) |
| P17 ciclu_nou | farm?, date?, guide?: string, motiv? | open the B4 form Dialog prefilled | P · manager, owner |
| P18 schimba_ghid | farm?, guide? | open its form Dialog prefilled | P · manager, owner |
| P19 serie_noua | name?, year?: integer, farm?, date?, ageDays?: integer, rows?: [{section, F?, M?}] | prefill S18 form | P · owner |
| P20 varsta_populare | days: integer, series?, motiv? | open the age Dialog prefilled | P · owner |
| P21 utilizator_nou | name?, login?, role?: vet\|viewer\|manager\|owner, farms?: Farm[] | prefill the S22 add panel (never a password) | P · owner |
| P22 rol_ferme | user: string, role?, addFarms?: Farm[], removeFarms?: Farm[] | prefill the user panel | P · owner |
| P23 ferma_sectiuni | action: ferma_noua\|adauga_sectiune\|redenumeste, farm?, type?, guide?, code?, parent?: Section, name? | prefill S19 | P · owner |
| P24 ghid_nou | name?, type? | open Importă ghid prefilled (file/paste by hand) | P · owner |
| P25 tip_hrana_nou | name?, forSex?: Sex | prefill S21 panel | P · owner |
| P26 inchide_seria | series: string | open S18 with Închide seria focused (no Dialog) | P · owner |
| P27 editeaza_tip_hrana | feedType, name?, forSex? | prefill S21 panel (never retire) | P · owner |
| P28 inlocuieste_ghid | guide: string | open the guide's Înlocuiește page | P · owner |
| X01 exporta | – | export of the current screen's filters (whole period) | I (no Anulează) · viewer+ where Exportă exists |
| X02 iesi | – | confirm "Ieși din cont?" Ieși · Rămâi | C · all |
| X03 anuleaza | – | client `undo_last_voice` | I · all |
| X04 ajutor | – | answer with 4 examples for this screen and role | R · all |
| clarify | question: string, options: string[] (≤ 3) | clarify | – |
| neinteles | heard: string | not_understood with 2 examples for screen + role | – |

**Action shapes** returned by `/api/voice/command`:
```json
{ "type": "navigate", "route": "/ferme/hirbovat/sectiuni/h5", "query": { "tab": "pasari", "types": "mortalitate", "period": "weeks:8" },
  "toast": "Am deschis Hîrbovăț › H5 · Păsări: Mortalitate, 8 săpt.", "undo": true }
{ "type": "answer", "text": "Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6%, sub ghid cu 17,2 puncte.", "lamp": { "status": "warn", "word": "abatere" },
  "navigate": { "route": "/rapoarte/productie", "query": { "farmId": "…" } }, "opened": "Am deschis Rapoarte · Producție" }
{ "type": "prefill", "form": "zilnic", "route": "/inregistreaza/zilnic", "query": { "farmId": "…", "date": "2026-11-18" }, "mode": "new", "entryId": null,
  "fields": [ { "path": "sections.H5.1.F.mortality", "value": 3, "previous": null, "label": "H5.1 · ♀ Mortalitate 3" },
              { "path": "sections.H5.1.M.mortality", "value": 1, "previous": null, "label": "♂ Mortalitate 1" } ],
  "banner": { "title": "Am completat din ce ai spus. Verifică și salvează.", "summary": "2 câmpuri: H5.1 · ♀ Mortalitate 3 · ♂ Mortalitate 1" },
  "normalized": "Zilnic · Hîrbovăț · 18.11.2026 · H5.1 · ♀ Mortalitate 3", "focus": "sections.H5.1.F.mortality",
  "continueWith": { "form": "hrana_zilnica", "label": "Continuă cu Hrană zilnică" } }
{ "type": "clarify", "question": "La care secțiune din H5?", "options": [ { "label": "H5.1" }, { "label": "H5.2" } ], "other": null, "pending": "<signed token>" }
{ "type": "refuse", "message": "Pentru zilele trecute cere managerului.", "banner": "info", "chip": { "label": "Deschide Introdus azi", "route": "/inregistreaza/introdus-azi" } }
{ "type": "confirm", "question": "Ieși din cont?", "options": [ { "label": "Ieși", "client": "logout" }, { "label": "Rămâi" } ] }
{ "type": "export", "url": "/api/export?…", "toast": "Export .xlsx descărcat · Hîrbovăț · Păsări · sapt. 20–59" }
{ "type": "not_understood", "message": "Nu am înțeles „pune trei la hala”.", "examples": [ "pune 3 găini moarte la H5.1", "ce lipsește azi?" ] }
{ "type": "client", "command": "load_more", "times": 1, "toast": "Am încărcat sapt. 52–55" }
```
`prefill` also carries `voiceToken` = HMAC-SHA256(server secret, runId ‖ userId ‖ form ‖ sha256(transcript) ‖ exp), `exp` = now + 10 min. `prefill.mode`: `new`; `edit` (the caller's own save from today, D07); `correct` (a past save or someone else's today, manager/owner, D70 B; Motiv filled only if spoken); a vet on someone else's save → `refuse` (D71). A field that already held a typed value different from the spoken one is replaced and carries `previous` ("din voce · înainte 2"). `pending` is an HMAC-signed token (60 s) with the resolved intent so a chip choice resumes without another LLM call; "Alta…" (`other.selector`: `section`, `farm`, `feedType`) opens the screen's own selector.

**Signed tokens in Go** (package `platform/tokens`, standard library only: `crypto/hmac`, `crypto/sha256`, `crypto/hkdf`, `encoding/base64`). One helper signs every short-lived token of the API — `voiceToken` (10 min), clarify `pending` (60 s), setup `undoToken` (until 23:59, §7.19b), price `undoToken` (§7.13) — each with its own key derived from `APP_SECRET` (§11.10) by HKDF-SHA256 with a purpose label (`"il voice-token v1"`, `"il voice-pending v1"`, `"il undo v1"`), so a token of one kind never verifies as another.
```go
// voiceToken = base64url(payload) + "." + base64url(HMAC-SHA256(key, payload))
// payload    = runId "|" userId "|" form "|" hex(sha256(transcript)) "|" exp (unix seconds)
func SignVoice(key []byte, runID, userID uuid.UUID, form, transcript string, exp time.Time) string {
	sum := sha256.Sum256([]byte(transcript))
	payload := strings.Join([]string{runID.String(), userID.String(), form, hex.EncodeToString(sum[:]),
		strconv.FormatInt(exp.Unix(), 10)}, "|")
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return b64.EncodeToString([]byte(payload)) + "." + b64.EncodeToString(mac.Sum(nil)) // b64 = base64.RawURLEncoding
}

// VerifyVoice recomputes the MAC over the decoded payload and compares with hmac.Equal (constant time), then checks
// exp > now, userId = the caller, form = the endpoint's form, and the hash = sha256 of the voiceTranscript sent with the
// save. Any failure → the transcript is dropped and the save is recorded as manual (§8.9); never an error to the client.
```
The token carries a hash, not the transcript, so it reveals nothing; `transcript` is the text as finally sent (an edited transcript gets a new run and a new token).

### 8.6 Server validation pipeline (in this order)
1. **Defaults** from context: no farm → the scope's farm, else the user's only farm, else `clarify` "La ce fermă?" (Telița · Hîrbovăț); no date → today; "săptămâna asta" → the scope's current SAPT.
2. **Resolve references**: farm/section/feed type/guide/series/user against the caller's visible entities (diacritics-insensitive). A hall where a compartment is needed → `clarify` "La care secțiune din H5?" with its compartments (> 3 → the 3 with birds first + "Alta…"). A record (P15, F11, Q14) → the save that holds that value (type, date, section).
3. **Role** of the intent (table above) → `refuse` with the screen's sentence ("Nu ai acces la Istoric modificări.").
4. **Farm scope** → `refuse` "Nu ai acces la această fermă."
5. **Time window** (§3.2): vet + date ≠ today on a P-intent → `refuse` "Pentru zilele trecute cere managerului." + chip "Deschide Introdus azi"; vet own entry after 23:59 → the S15 reason.
6. **Never list** (§3.4) → `refuse` "Vocea nu salvează. Verifică formularul și apasă Salvează." / "Vocea nu șterge. Deschide înregistrarea și apasă Șterge." / "Parolele nu se spun cu voce." / prices for vet/viewer → the role refusal; a price with a manager's delivery → "Prețul îl adaugă proprietarul în Prețuri hrană."
7. **Existing data** (P01–P09): the day already saved → `edit` / `correct` / `refuse` as above.
8. **Required fields not given** (Motiv of a cull, a correction's Motiv) stay empty for the form's own refusal; a cull reason is asked first as `clarify` (3 choices).
9. Build the action. Figures for Q-intents come from the services of §5 with the caller's visibility (no lei for vet/viewer).

### 8.7 Answer templates (Q-intents)
Server-side Romanian templates; figures formatted per §11.3; status as Lamp + word.
| Intent | Template (examples on 18.11.2026) |
|---|---|
| Q01 | "La Hîrbovăț sunt azi ♀ 12 561 găini și ♂ 794 cocoși (18.11.2026, sapt. 59)." |
| Q02 | "Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6%, sub ghid cu 17,2 puncte." + abatere; day: "Ieri, 17.11.2026: ouat 39,7%, ghid 57,6%." *(proposed)* |
| Q03 | "Sapt. 59: au murit ♀ 10 găini și ♂ 0 cocoși. Mortalitate cumulată ♀ 5,7% (ghid 7,25%), ♂ 34,7%." *(proposed)* |
| Q04 | "Azi: ♀ 160 g (ghid 157 g) · ♂ 141–144 g (ghid 156 g); ♀ 2 303,2 kg · ♂ 154,7 kg." *(proposed)* |
| Q05 | "Stoc Hîrbovăț: Breeder 2 31 665 kg, 13,2 zile (până pe 01.12.2026) · Cocoși 1 781 kg, 10,4 zile." *(proposed)* |
| Q09 | "SAPT 59 · vârstă 407 zile · 18.11.2026" |
| Q16 | "Sapt. 58: vândute ♀ 10 546 · ♂ 749." |
| Q17 | "13.11.2026: 25 640 kg." |
| others | one sentence of the same shape *(proposed)*: subject · period · figures · guide · status |
Answers are written only (no speech synthesis); they use only figures the user may see on the screen that opens.

### 8.8 Romanian numbers and codes
The model converts; the server re-checks with a deterministic `voice.ParseRoNumber()` on the transcript span the model cites (`sourceText` optional property on numeric params) and, on disagreement, prefers the deterministic value and marks the field. Test table:
| Spoken | Value |
|---|---|
| "trei" / "trei găini" | 3 |
| "două sute cincizeci" | 250 |
| "o mie opt sute treizeci și unu" | 1 831 |
| "douăzeci și cinci de mii șase sute patruzeci de kilograme" | 25 640 kg |
| "doi virgulă patruzeci și unu" | 2,41 |
| "o sută șaizeci de grame" | 160 g |
| "nouă lei douăzeci și doi" | 9,22 lei |
| "zero" / "niciuna" / "nimic" | 0 (never empty) |
| "ha cinci punct unu", "hala cinci compartimentul unu", "H cinci unu" | H5.1 |
| "ha cinci liniuță unu" | H5-1 (Telița) |
| "hala cinci", "H5" | H5 (hall → clarify where a compartment is needed) |
| "săptămâna cincizeci și opt" / "SAPT 58" | SAPT 58 |
| "alaltăieri", "pe 13 noiembrie", "de pe 1 până pe 15 noiembrie" | 16.11.2026; 13.11.2026; interval 01.11–15.11.2026 |

### 8.9 Privacy, limits, cost
- Audio: memory only, discarded after STT; provider contract must guarantee no retention and no training (open question 13). Transcripts: not stored for navigate/filter/answer; for a save, the client sends `source: "voice"` + `voiceTranscript` + `voiceToken` only when ≥ 1 "din voce" mark is still on the form; the server verifies the token (signature, not expired, same user, same form, transcript hash) and only then stores the transcript in `audit_log` (`source = voice`) and nowhere else; an invalid or missing token → the transcript is dropped and the save is `manual` (no error). Server logs carry `runId`, intent id, timings — never transcript or audio. First-use notice (client): "Ce spui nu se păstrează. Dacă salvezi un formular completat din voce, textul comenzii rămâne în Istoric modificări."
- Limits (`app_setting`): ≤ 20 s audio, ≤ 1 MB, ≤ 500 characters per transcript, ≤ 10 commands per minute and ≤ 200 per day per user → `429` *(proposed)* "Ai folosit comenzile vocale de azi. Poți scrie sau apăsa în continuare." (open question 14).
- Cost: `voice_usage` per user/day (commands, audio seconds, tokens). Estimate at implementation from current provider prices: per command ≈ (STT price × ~5 s) + (LLM input ≈ 8–15 k cached tool tokens + 1 k context, output ≈ 150 tokens); prompt caching of the static tool list is required to keep it low.
- Timeouts: STT 6 s, LLM 8 s, whole run 10 s → the client shows the error state (spec §8 step 6).

### 8.10 Fallbacks
- STT failure → `503 VOICE_UNAVAILABLE`; the panel shows "Comanda nu s-a trimis: „…”." / "Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos." The typed path still works.
- LLM failure/timeout → a deterministic **fast path** handles the common navigation intents without the model (N01, N05–N09, N11–N14, F01, F02, F08, F14, X01–X04 via a keyword grammar: "deschide|arată|du-mă" + page words + farm/section codes + period words); anything else → `503`.
- The fast path also runs **before** the LLM for exact matches ("deschide Acasă"), saving cost and latency.

### 8.11 Provider-agnostic interfaces (Go), the generated tool list, and recommendation
```go
package voice

type STTOptions struct {
	Language string   // "ro-RO"
	Hints    []string // the caller's farm names, section codes, feed types (phrase boosting where supported)
}
type Transcript struct {
	Text        string
	DurationSec float64 // reporting only (voice_usage.audio_seconds), never money
}
type STTProvider interface {
	Name() string
	Transcribe(ctx context.Context, audio []byte, mime string, opts STTOptions) (Transcript, error)
}

// Optional: a provider that streams implements this too (checked with a type assertion).
type STTStreamer interface {
	Stream(ctx context.Context, mime string, opts STTOptions) (STTStream, error)
}
type STTStream interface {
	Write(chunk []byte) error // binary audio chunk from the WebSocket
	CloseSend() error         // client sent {"type":"end"}
	Events() <-chan STTEvent  // closed after the final event or on error
	Err() error
}
type STTEvent struct {
	Final bool // false = partial
	Text  string
}

type ToolDef struct {
	Name        string // "P01_mortalitate", "clarify", "neinteles"
	Description string
	InputSchema json.RawMessage // JSON Schema, additionalProperties false everywhere
}
type IntentInput struct {
	System     string // static part (cacheable) + the dynamic context of §8.3, as the provider needs it
	Context    json.RawMessage
	Tools      []ToolDef
	Transcript string
}
type IntentResult struct {
	Tool  string
	Args  json.RawMessage // validated against the tool's schema by the pipeline before use
	Usage Usage           // input/output/cached tokens → voice_usage
}
type IntentProvider interface {
	Name() string
	Parse(ctx context.Context, in IntentInput) (IntentResult, error)
}
```
Implementations live in `internal/voice/providers/<name>` and are chosen by `STT_PROVIDER` / `INTENT_PROVIDER` (§11.10); `providers/fake` returns scripted results for tests and CI. Errors and timeouts from either interface map to `503 VOICE_UNAVAILABLE` (STT) or to the fast path / 503 (intent, §8.10).

**The action catalogue and the per-user tool schemas.** The 81 intents of §8.5 plus `clarify` and `neinteles` are static Go data in `internal/voice/catalogue.go`: one `Intent{ID, Tool, Description, Mode, Roles, Params}` per row, where `Params` is built with a few schema helpers (`obj(req("items", arr(obj(req("section", ref("Section")), req("sex", ref("Sex")), req("count", ref("Count")))))), opt("farm", ref("Farm")), …)`) and `ref` points into the shared definitions of §8.4. Per command, `voice.BuildTools(principal, visible)`:
1. keeps only the intents allowed for the caller's role (§8.5 roles column), plus `clarify` and `neinteles`;
2. fills the user-dependent enums from the caller's visible entities: `Farm` (farm names), `Section` (union of section codes), `FeedType` (active types), guides, series names, user names (owner only); an enum that would be empty is replaced by a plain `string` so the model can still say what it heard and the server answers with `clarify` / `refuse`;
3. inlines the `$defs` into each tool's `input_schema` (no `$ref` left; verify whether the chosen provider accepts `$ref` before relying on it) and marshals each schema with sorted enum values and a fixed property order, so the same user gets byte-identical tools on every command (the provider's prompt cache keys on the exact prefix; static system prompt first, then the tools).
A golden test writes the generated tool list per role on the dev seed to `internal/voice/testdata/tools_<role>.json` and fails on any unreviewed change; another test checks every catalogue intent has a role list, a mode and an action builder.
**Recommendation** (verify model ids, regions and data terms at implementation; open question 13):
- **Intent**: Anthropic Claude, a small fast tier, through the Anthropic API or an EU-region deployment on AWS Bedrock / Google Vertex if EU processing is required; tool use with strict schemas, `tool_choice: any`, temperature 0, prompt caching for the tool list. One call per command. In Go: the official Anthropic Go SDK (`github.com/anthropics/anthropic-sdk-go`; verify its tool-use and cache-control API at implementation) behind `IntentProvider`.
- **STT**: a streaming ro-RO recogniser with an EU region and a no-retention option (e.g. Azure AI Speech or Google Cloud Speech-to-Text), chosen after a test in a hall with fans (open question 16); keep the adapter so it can be swapped. In Go: the vendor's Go SDK if it has one for streaming recognition, else its REST/WebSocket API with `net/http` / `coder/websocket` (verify at implementation).

---

## 9. Live updates (D197; MO03)

- **Transport**: Server-Sent Events, `GET /api/live` (`text/event-stream`), one connection per tab, cookie-authenticated. Heartbeat comment every 25 s. Proxy-safe headers (`Cache-Control: no-cache`, `X-Accel-Buffering: no`).
- **Source**: every write inserts `live_event` rows in its transaction and calls `pg_notify('live_event', id::text)` in the same transaction (delivered at commit, dropped on rollback); each Go API instance `LISTEN`s on one dedicated connection and fans out to its connections (§9.1). Replay: the browser's `Last-Event-ID` → events with `id > last` from `live_event` (24 h); older → `event: resync` (the client refetches the page).
- **Authorization filter** per connection: deliver an event only if its `farm_ids` intersect the user's farms (owner: all; empty `farm_ids` = global setup events, delivered to the owner, plus `user.changed` to the user concerned). Price payloads (`price.*`, lei fields) go only to owner and manager.
- **Event types and payloads** (ids only + the minimum to place the change; the client refetches the affected queries — never trust event figures for display beyond the Toast/"noi" bar):
| `event:` | `data` |
|---|---|
| `entry.created` / `entry.updated` / `entry.deleted` / `entry.restored` | `{ entryId, kind, farmIds, dates: ["2026-11-18"], sections: ["H6.1","H6.2"], recordTypes: ["vanzare"], author: "Ion Rusu", at, summary: "Vânzare · H6.1, H6.2 · ♀ 1 831 · ♂ 289", corrected: bool }` |
| `counts.changed` | `{ farmId, fromDate: "2026-11-18", sections: [...], alive: { F: 12561, M: 794 } }` (header counts swap, D197) |
| `cycle.opened` / `cycle.closed` / `cycle.reopened` / `cycle.guide_changed` | `{ cycleId, farmId, seriesId, date, sapt }` |
| `transfer.created` | `{ entryId, date, from: { farmId, section }, to: { farmId, section }, F, M }` (CycleTrack landing at the today line) |
| `delivery.created` | `{ entryId, farmId, feedType, kg, date }` |
| `price.set` / `price.changed` | `{ deliveryIds: [...], count, kg }` (owner, manager) |
| `setup.changed` | `{ entity: farm\|section\|guide\|feed_type\|series, id }` |
| `user.changed` | to that user only: `{ role, farms, active }` → the client re-reads `/api/auth/me` (deactivated → signed out) |
- **Batching**: the server coalesces per connection for 1 s (a Zilnic save of 4 sections + its counts change → one batch); the client inserts rows, tints, and announces one polite sentence per batch (D197); if the reader is scrolled away or typing, the client shows "N înregistrări noi · Arată".
- **"Actualizat hh:mm"** in the page head = the local time of the newest applied event (or the page fetch `asOf`). On reconnect the client shows the last value until replay finishes.
- **Reconnection**: `retry: 3000`; exponential back-off up to 30 s on the client; after 5 failures the page shows its normal offline/error Banner (online only, D08).

### 9.1 SSE in Go (package `live`)
**Handler** `GET /api/live` (cookie auth like every route; no request body):
```go
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{}) // lift http.Server.WriteTimeout for this response only (see caveat)
	hd := w.Header()
	hd.Set("Content-Type", "text/event-stream")
	hd.Set("Cache-Control", "no-cache")
	hd.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	sub := h.hub.Subscribe(p)    // subscribe BEFORE replay so nothing falls between the two
	defer h.hub.Unsubscribe(sub) // remove under the write lock; closes sub.done exactly once
	fmt.Fprint(w, "retry: 3000\n\n")
	last := parseLastEventID(r) // header Last-Event-ID (int64), 0 when absent
	if last > 0 {
		evs, tooOld := h.store.Since(r.Context(), last, 24*time.Hour) // id > last, at > now() − 24 h, ordered by id
		if tooOld {
			writeEvent(w, Event{Type: "resync"})
		} else {
			last = writeAll(w, p, evs)
		}
	}
	_ = rc.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	var batch []Event
	flushTimer := time.NewTimer(time.Hour)
	flushTimer.Stop()
	for {
		select {
		case <-r.Context().Done(): // client gone
			return
		case <-sub.Done(): // removed by the hub (shutdown); closed once via sync.Once
			return
		case <-sub.Kick(): // lagging: return, the deferred Unsubscribe removes and closes it
			return
		case ev := <-sub.C: // never closed
			if ev.ID <= last || !visible(p, ev) {
				continue
			} // dedupe against the replay; farm + price filter
			if len(batch) == 0 {
				flushTimer.Reset(time.Second)
			} // coalesce per connection for 1 s
			batch = append(batch, ev)
		case <-flushTimer.C:
			last = writeAll(w, p, batch)
			batch = batch[:0]
			if rc.Flush() != nil {
				return
			}
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n")
			if rc.Flush() != nil {
				return
			}
		}
	}
}
```
Each event is written as `id: <live_event.id>\nevent: <type>\ndata: <one-line JSON>\n\n`; the payload is filtered per caller before writing (price fields only for owner/manager, D200).

**Hub (per-farm fan-out)**: `live.Hub` holds `map[uuid.UUID]map[*Sub]struct{}` (farm id → subscribers), one set for owners (they get every farm), and `map[uuid.UUID]map[*Sub]struct{}` by user id for `user.changed`, all behind one `sync.RWMutex`. These maps change **only under the write lock** (`Subscribe`, `remove`). `Publish(ev)` takes the read lock, looks up the subscribers of `ev.farm_ids` (empty = global: owners, plus the user concerned for `user.changed`) and does a **non-blocking** send on each `Sub.C` (buffer 64); when the buffer is full it only marks the subscriber lagging (`sub.lagging.Store(true)`) and signals it (`sub.kick()`, a non-blocking send on a 1-slot channel) — it never closes anything under the read lock. Closing happens in one place: `remove(sub)` takes the write lock, deletes the subscriber from every map, and closes `sub.done` through `sub.closeOnce.Do(…)`, so a subscriber is closed exactly once whatever path gets there first (handler return, lagging kick, shutdown). The handler selects on `sub.C`, `sub.kick` and `sub.done`: on a kick it returns, its deferred `Unsubscribe` calls `remove`, the browser reconnects with `Last-Event-ID` and replays from the table — a slow client never blocks the hub or the listener. `Sub.C` itself is never closed (receivers stop on `done`), so a racing `Publish` can't send on a closed channel. When a user's role or farms change, their sessions are revoked (§4.3) and the open stream ends at its next auth check (the handler re-reads the session every 5 min and on `user.changed`).

**Postgres LISTEN/NOTIFY with pgx (multi-instance)**: each instance runs one goroutine `live.Listener` holding a dedicated `*pgx.Conn` (acquired outside the pool, or `pool.Acquire` + `Hijack`; verify at implementation) that runs `LISTEN live_event` and loops on `conn.WaitForNotification(ctx)`. A notification's payload is the event id; the listener reads that row (`select … from live_event where id = $1`) and publishes it. On a connection error it reconnects with back-off, re-runs `LISTEN`, and catches up with `where id > $lastSeen and at > now() − interval '24 hours' order by id`; ids already published are skipped. With one instance the same path runs (no in-process shortcut), so single- and multi-instance behave the same. Ids come from a sequence, so two transactions on different farms can commit out of id order: live delivery is unaffected (the listener fetches each notified id), while a replay after a disconnect (`id > Last-Event-ID`) can miss an event that committed during the disconnect with a lower id; the client therefore refetches the visible page queries after every reconnect (cheap, and it already shows the last values meanwhile).

**Purge**: every instance runs an hourly ticker on the maintenance pool (`MAINT_DATABASE_URL`, role `il_maint`, §11.9–§11.10); the app role has no DELETE (§11.4). Each tick: `conn := maintPool.Acquire(ctx)`; `select pg_try_advisory_lock(<fixed key>)` on that connection; if it returns false another instance is purging, release and stop; else, on the same connection, delete `live_event` older than 24 h, `login_attempt` older than 30 days and `request_replay` older than 7 days, then `select pg_advisory_unlock(<fixed key>)` and `conn.Release()` (session-level lock: lock, work and unlock must use the same connection, which is why it is acquired explicitly and not taken from `pool.Exec`).

**`http.Server` timeouts caveat**: `WriteTimeout` is a deadline for the whole response, so a server-wide `WriteTimeout: 60s` would cut every SSE stream (and the voice WebSocket) after 60 s. Keep `ReadHeaderTimeout` (10 s), `ReadTimeout` (30 s), `WriteTimeout` (60 s) and `IdleTimeout` (120 s) for ordinary requests, and in the two long-lived handlers clear the deadline per response with `http.NewResponseController(w).SetWriteDeadline(time.Time{})` (and the read deadline for the WebSocket). Also: no compression middleware and no request-timeout middleware on `/api/live` (both buffer or cancel the stream); the reverse proxy needs buffering off and a read timeout above the 25 s heartbeat (nginx: `proxy_buffering off; proxy_read_timeout 1h;`). On shutdown, `srv.RegisterOnShutdown(hub.Close)` ends every stream so `srv.Shutdown` doesn't wait for them; clients reconnect to another instance and replay.

---

## 10. Exports and imports

### 10.1 Export .xlsx (D5, D191; spec D1 › Export)
- Library: **excelize** (`github.com/xuri/excelize/v2`), package `export`. Each sheet is written with its `StreamWriter` (`f.NewStreamWriter(sheet)`; rows via `SetRow(cell, []any{excelize.Cell{StyleID: …, Value: …}})`, header group cells via `MergeCell`, then `Flush`), so a whole cycle of day rows stays small in memory; the file goes to the response with `f.Write(w)` after the headers are set, then `f.Close()`. Styles are created once per file (`f.NewStyle` with custom number formats `#,##0`, `#,##0.0`, `0.0`, `#,##0.00`, `dd.mm.yyyy`; bold for week rows). Data comes from the same `periodtable` builder as `/api/table` with "all weeks" (no paging) and the caller's role, so lei columns are absent for vet/viewer by construction. Verify the streaming API names at implementation.
- **What**: the whole chosen period (every week, not only the loaded ones): open weeks with their days, closed weeks as week rows (grain Săpt.), the column groups, chips, Detaliu level and order shown on screen; **one sheet per tab** of the page (S4: Rezumat · Păsări · Hrană · Greutate · Producție; report pages: one sheet per table/part).
- **Layout**: head lines (farm, cycle "Seria 2026 · Hîrbovăț · 22.02.2026–azi · SAPT 20–59", period, chips, Detaliu, "Exportat de {name} · 18.11.2026, 13:20"), then the two header rows (group, leaf), then rows: week rows bold with the "SAPT 58 · 10.11–16.11" label, day rows with Data (Excel date, `dd.mm.yyyy`) · Vârstă (zile) · SAPT; a foot "Total SAPT 20–59".
- **Values**: numbers as numbers (counts `#,##0`, kg `#,##0.0`, % as the number 41.1 with format `0.0` and "%" in the header, lei `#,##0.00`); a `–` cell exports empty and its reason goes to the sheet's note line under the table ("Celule goale: H7 goală din 14.11.2026; H7.1, H7.2 ♀ 13–14.11.2026 fără hrană în registru."). Viewer exports have no lei columns and no Cost sheet.
- File name: `{Farm}-{Tab}-SAPT{a}-{b}.xlsx` without diacritics (`textkey.ASCII`: NFD via `golang.org/x/text/unicode/norm`, drop combining marks, so "Hîrbovăț" → "Hirbovat"). Response within 3 s for a whole cycle (§11.6). Day cells are written as `time.Time` values at UTC midnight with the date style (Excel stores no zone), never as text.

### 10.2 Guide import
`.xlsx` / `.csv` / pasted table → preview → save (§6.12, §7.17). Parsing is server-side (same code for preview and save; the save re-parses the rows sent by the client, never trusts client validation). Go, package `guides`:
- Upload: `r.ParseMultipartForm(2 << 20)` behind `http.MaxBytesReader(w, r.Body, 2<<20+4096)`; the file is read into memory and never stored. The format is decided by content, not by the name only: an `.xlsx` must open with `excelize.OpenReader(bytes.NewReader(b))` (first sheet, `GetRows`), anything else is tried as text; neither → G1.
- CSV: `encoding/csv` with `Comma` = `;` when the header line has more `;` than `,`, else `,`; `LazyQuotes = true`; `FieldsPerRecord = -1`; a UTF-8 BOM is stripped. A pasted table (`{ text }`) uses `Comma = '\t'`.
- Header matching and name keys: `textkey.Key(s)` = NFD, drop combining marks (both ș/ş and ț/ţ fold to s/t), lower-case, drop spaces and unit words (`g`, `%`); the same function fills `name_key` for farms, guides, feed types and series (D111, D114, D116).
- Numbers: `guides.ParseNumber` implements §6.12 (decimal comma or point, thousands separators, trailing `%`/`g`) and returns `num.Num`; the G1–G15 fixtures of §11.8 are its table test.

### 10.3 One-time history migration (D09)
A developer command in Go, **`cmd/migrate-history`** (`cd backend && go run ./cmd/migrate-history --db "$DATABASE_URL" --extracts ../design/data/extracts [--replace] [--seed-dev]`), idempotent (it refuses to run on a database that already holds a `migration` entry unless `--replace`). It reads the **extracts** (already cleaned and checked by `data03_build.py` / `data03_check.py`; the workbook mapping, the sale split and the Cat. pairing are documented there), not the workbooks. One `entry(kind='migration', source='migration')` per farm; every row created points to it; `audit_log` gets one `create` row per farm with counts. The whole import is **one transaction** (bulk rows with pgx `CopyFrom` / sqlc `:copyfrom`), so a failed verification leaves the database untouched.

**Go or Python — decision: Go.** The two-stage pipeline stays: *workbooks → extracts* is the existing Python (`design/data/scripts/data01…data03_build.py`, openpyxl, already run; its JSON output is committed and is the migration's only input), and *extracts → database* is the Go command, which mirrors the mapping rules that `data03_build.py` documents and this section restates (transfer pairs by arrival date, link differences as `adjustment`, sales from 12.11, Telița feed types by date, eggs with Cat. 1–3, unpriced deliveries). Why Go for the second stage: (1) the loader is a library (`internal/migratehistory`, exposed by the command) that the Go test harness calls in-process to build the golden fixture (§11.8), so CI needs no Python and the fixture is always loaded by the same code as production; (2) its verification step calls the real services (`calc`, `periodtable`, `ledger.CheckCounts`) to reproduce every golden value of §11.7, which only Go can do; (3) it writes through the sqlc queries and the same enum/decimal/day types, so a schema change breaks it at compile time instead of at go-live; (4) one toolchain in the Docker image and CI. Re-implementing the openpyxl workbook parsing in Go is **not** planned: if the workbooks change, rerun the Python scripts, then this command.

| Target | From | Rule |
|---|---|---|
| `guide`, `guide_week` | `guide-growth.json` (SAPT 1–25: feed_g, body_weight_g), `guide-laying.json` (SAPT 19–65: feed_g ♀/♂, lay_pct ♀ = Productivitate col I, mortality_cum_pct ♀) | names illustrative (D22); the other guide-laying fields (eggs/hen, hatchability) are not imported (D184) |
| `farm`, `section` | `farms-sections.json` | Telița growth, guide creștere; Hîrbovăț laying, guide ouat; halls H5–H7 + compartments |
| `feed_type` | `feed-types.json` | Starter, Grower, Pre-ouat (for both sexes), Breeder 2 ♀, Cocoși ♂ (names illustrative, D30) |
| `series` | `cycles.json` | Seria 2026 (07.10.2025, age 0, active); Seria 2025 (01.11.2024 inferred, closed) |
| `cycle` | `cycles.json` | Telița 07.10.2025–28.02.2026 (auto, emptied); Hîrbovăț Seria 2026 from 22.02.2026 (open); Hîrbovăț Seria 2025 07.03.2025–28.01.2026 (`start_inferred`, migration) |
| Transfer pairs (definition) | `transfers-seria2026.json` `links_inferred_by_date` | One **pair per destination compartment and sex**: out row at the source section (−q), in row at the destination (+q), one `transfer` entry per link. The pair takes the **arrival date** (both rows dated the day the birds arrive); the departure date is kept on the entry (`note` = `"plecare {out_date}"`, e.g. Telița H2 out 27.02.2026 → in H7.1/H7.2 on 28.02.2026). A source that feeds 2–6 compartments (♂ H5-1 + H5-2 → H5.1…H7.2; ♀ H1 → H6.1, H6.2, H7.1) is split into one pair per destination with the destination's arrived quantity; the source's out total is the sum of its pairs. With two sources for one link (♂ H5-1 + H5-2 → six compartments) the destinations are filled from the sources in code order until each source's out quantity is used. The link difference (arrived − out: ♀ +2 at H2's link, ♂ +30) is an `adjustment` at the destination, note "diferență registru". On the source day before arrival (27.02 out, 28.02 in) the source keeps the birds one more day: accepted (D25/D63 reconciliation) |
| Telița birds | `placement-seria2026.json`, `mortality-telita-daily.json` (mortality, sexing_error signed, correction), `transfers-seria2026.json` (`internal_moves_telita`, `transfers_out`, `links_inferred_by_date`, `other_corrections_telita`) | placement rows; mortality per section-day (0 rows kept: "entered"); internal moves as transfer pairs; outs paired with Hîrbovăț arrivals by `links_inferred_by_date` as transfer pairs (the +2 ♀ H2 and +30 ♂ link differences = `adjustment` at the destination, note "diferență registru"); `other_corrections_telita` = adjustments; Telița's residual counts on 28.02.2026 (23/16/−9/0/454/244) = adjustments to 0, placed on the day a section would otherwise go below 0 and on 28.02.2026 for the rest (D25/D63, shown as "Diferență registru", never as Corecție), so the cycle closes on 28.02.2026 |
| Hîrbovăț birds | `hirbovat-daily.json` `section_day` | `mort` → mortality (incl. 0), `sold` → sale entries grouped per day × hall (DATA04: 7 sale records), `mv` → the 22.10.2026 H7.2 → H7.1 ♂ 19 transfer pair; arrivals come from the Telița links above |
| Feed | `feed-daily-telita.json` (section kg, g where present), `hirbovat-daily.json` (`g`, `kg`) | `feed_allocation` with `g_per_bird` (null allowed early at Telița) and `kg_recorded` = the typed kg; rows with g but no kg (H7 ♀ 13–14.11.2026) are **not** imported (they stay "– fără hrană în registru" / lipsă); Telița feed type by date: Starter to 13.11.2025, Grower 14.11.2025–19.01.2026, Pre-ouat from 20.01.2026 *(assumption from the delivery notes)*; Hîrbovăț ♀ → Breeder 2, ♂ → Cocoși |
| Deliveries | `feed-deliveries.json` rows | `feed_delivery` with `price_from_register = true` where priced (Who/when "din registru", D163); the 4 unpriced stay "fără preț"; `meta.hirbovat_adjustments` → `feed_stock_adjustment` |
| Eggs | `hirbovat-daily.json` `hall_day` (eggs, inc, jos, cat1–3, egg_g) | one `egg_production` per hall-day with eggs (first eggs 10.03.2026; every such hall-day has Cat. 1–3 from the form, so `total_eggs` = the register's eggs on all 757 hall-days, D18 check); hall-days without eggs → no row |
| Weighings | `weighings-weekly.json` | one `weighing` per row (sex from the section; redated rows as given) |
| Seria 2025 | `eggs-hirbovat-prev-daily.json` (old hall codes → H5–H7, D158c) | `egg_production` per hall-day; `legacy_mortality` per compartment-day (`emptying = true` on depopulation days) |
| Not migrated | `illustrative-events-hirbovat.json`, `sample-hirbovat-2026.json`, users Dr. Ceban / Ion Rusu | illustrative design data (D27, D152); production users are created by the owner (S22); the first owner by `go run ./cmd/create-owner` |
**Verification after migration** (inside the transaction, before commit; roll back and exit non-zero if any differs): every golden value of §11.7; `alive_end ≥ 0` everywhere; Hîrbovăț arrivals ♀ 26 883 · ♂ 3 024; deliveries Hîrbovăț 69 · 9 562 437 lei; stock end 18.11.2026 Breeder 2 31 664,9 · Cocoși 1 781; eggs Σ 5 073 155 to 18.11.2026.

---

## 11. Non-functional requirements

### 11.1 Time zone and day boundary
- Business time zone **Europe/Chisinau** (EET/EEST). All `date` columns are local days; timestamps are stored UTC (`timestamptz`). The DB runs in UTC; `app_today()` converts.
- A day starts at 00:00 local. "Until 23:59" = while `app_today()` equals the local date of `created_at`. "Missing from 12:00" = local noon. "Azi/Ieri" in forms are local.
- DST days have 23/25 hours; nothing is computed per hour, so no special case.
- **Go**: `timex.Chisinau, err = time.LoadLocation("Europe/Chisinau")` once at start-up (fail fast); `cmd/api` and the other commands `import _ "time/tzdata"` so the zone database is embedded in the binary (the distroless image has no `/usr/share/zoneinfo` guarantee). All "now" comes from an injected `timex.Clock` (`Now() time.Time`); `timex.Today(clock)` = the civil `Day` of `clock.Now().In(Chisinau)`. Handlers never call `time.Now()` directly (a lint rule or review check), so tests freeze time at 2026-11-18T12:00 local.
- **Go ↔ SQL day**: Go computes `today` and passes it as a query parameter wherever a query needs it; the only SQL that reads `app_today()` are the views (`v_section_sex_day` dense range, `v_cycle`). Production leaves `app.today` unset (both clocks are the real one); tests set it on every pooled connection (`pgxpool.Config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error { _, err := conn.Exec(ctx, "select set_config('app.today', $1, false)", "2026-11-18"); return err }`) together with the frozen `Clock`, so Go and SQL agree. The pool sets `timezone = 'UTC'` in `ConnConfig.RuntimeParams`.

### 11.2 Types
| Quantity | SQL type | API |
|---|---|---|
| bird counts, eggs | `integer` | integer |
| g/pasăre | `numeric(6,1)` | number |
| kg (deliveries, recorded) | `numeric(12,1)` / `numeric(10,1)` | number |
| derived kg | `numeric` exact (`g × birds / 1000`) | number (≤ 4 decimals) |
| lei total | `numeric(14,2)` | number |
| lei/kg | `numeric(12,4)` | number |
| % (guide) | `numeric(5,2)` | number |
| weights | `numeric(7,1)`, egg weight `numeric(5,2)` | number |
Never use `float64` for kg, lei, grams or percentages on the server. **Recommendation: `github.com/shopspring/decimal`** in Go code, registered with pgx through `github.com/jackc/pgx-shopspring-decimal` (`pgxdecimal.Register(conn.TypeMap())` in `AfterConnect`), and sqlc overrides mapping `numeric` → `decimal.Decimal` / nullable → `decimal.NullDecimal` (§12.4). Why not `pgtype.Numeric`: it is a transport type (no arithmetic, no rounding), so every formula of §5 would convert anyway. Sums stay in SQL `numeric` where the query can do them; Go does the ratios and the single final rounding (`Decimal.Round(n)`, half away from zero; verify the rounding mode with the D82 test cases of §5.15 at implementation). JSON goes through `num.Num` (§7.0a).

### 11.3 Formats (client-side; the server uses the same rules for Toasts, voice answers and xlsx head lines)
- Thousands: U+202F narrow no-break space ("14 080", "1 098"); decimal comma ("2,41 kg"); true minus U+2212 ("−82"); `+` on positive differences ("+7,7%").
- Percent without a space: "41,1%", "−1,4%" (D46); points: "−16,5 pct".
- Dates `dd.mm.yyyy`; date + time `dd.mm.yyyy, hh:mm` (D129); ranges with an en dash and no spaces, the start drops only a year shared with the end ("10.11–16.11.2026", "07.10.2025–28.02.2026"); in-context windows may drop both years ("16.11–18.11").
- Weeks: `SAPT 59` in labels, column heads, Mono context lines; `sapt. 59` in running text, Toasts, notes (D31, D115, D128). Day rows: "Mi 18.11 · 407 · 59".
- Unknown values `–` with a reason ("– lipsă", "– fără date în ghid", "goală din 14.11").
- Sex: ♀ găini / ♂ cocoși (icons with hidden words in the UI).

### 11.4 Auth and security
- Passwords: argon2id via `golang.org/x/crypto/argon2` — `argon2.IDKey(password, salt, 2 /*t*/, 19*1024 /*m KiB = 19 MiB*/, 1 /*p*/, 32)`, 16-byte salt from `crypto/rand`; stored as the PHC string `$argon2id$v=19$m=19456,t=2,p=1$<salt b64>$<hash b64>` (raw std base64), parameters read back from the string on verify so they can be raised later (re-hash on the next successful login when they differ); compare with `crypto/subtle.ConstantTimeCompare`. An unknown login still runs one `IDKey` against a fixed dummy hash, so timing doesn't reveal which logins exist. Hashing is bounded by a semaphore (4 at a time; 19 MiB each). Generated passwords: 12 chars from `abcdefghjkmnpqrstuvwxyz23456789` via `crypto/rand` (`rand.Int` per character, no modulo bias); never logged; responses carrying one are `Cache-Control: no-store`.
- Session: cookie `il_session` = 32 bytes from `crypto/rand`, `base64.RawURLEncoding`; set with `http.Cookie{Name: "il_session", Path: "/", MaxAge: 2592000, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}`; the DB stores `sha256(token)` (`user_session.token_hash`); the auth middleware looks it up joined to an **active** `app_user` with `revoked_at is null and expires_at > now()`, puts an `auth.Principal{UserID, Name, Role, FarmIDs (nil = owner, all farms)}` in the request context, and slides `expires_at` (+30 days) at most once per hour (`last_seen_at` older than 1 h). Revoked on logout, reset, deactivation, and any change of the user's role or farms (§4.3). Logout clears the cookie (`MaxAge: -1`).
- CSRF: `SameSite=Lax` plus, on every method other than GET/HEAD/OPTIONS: the `Origin` header must equal `APP_ORIGIN` (when `Origin` is absent, `Sec-Fetch-Site` must be `same-origin`; when both are absent → `403 FORBIDDEN`), and `Content-Type` must be `application/json` (or `multipart/form-data` for `/api/guides/preview`, an audio type for `/api/voice/stt`). Go 1.25 added `http.CrossOriginProtection` to `net/http`, which implements the `Sec-Fetch-Site`/`Origin` part; use it (with `APP_ORIGIN` as the trusted origin) if the chosen Go version has it, keep the Content-Type rule in our middleware (verify at implementation). No CSRF token is needed with these two checks.
- Login throttling (A1): counted in PostgreSQL (`login_attempt`), so it holds across instances and restarts: before checking the password, `count(*) where login = $1 and not success and at > now() − 15 min` ≥ 5 and the latest of them < 5 min ago → `429 RATE_LIMITED` with the A1 sentence and `retryAfterSeconds` (+ `Retry-After` header); the same per IP with 20. The client IP is `r.RemoteAddr`, or the right-most untrusted address of `X-Forwarded-For` when the request comes from a `TRUSTED_PROXIES` address (§11.10). Same message for unknown login and wrong password. Voice limits (§8.9) use the same pattern on `voice_usage` plus an in-memory per-user `golang.org/x/time/rate` limiter for the per-minute cap.
- Authorization in one place (package `auth`: `RequireRole(roles...)` middleware, `p.CanSeeFarm(id)`, `entries.CanEdit(p, entry, today)`) and tested by the §3.3 matrix. Prices are removed from DTOs for vet/viewer by type (§7.0a), not just hidden.
- WebSocket `/api/voice/stream`: `Origin` must equal the app origin (cross-site WebSocket hijacking; SameSite cookies don't cover it), session checked on upgrade, idle timeout 25 s.
- Headers: HSTS, CSP (self + the STT WebSocket origin), `X-Content-Type-Options`, `Referrer-Policy: same-origin`, `Permissions-Policy: microphone=(self)`.
- SQL only through sqlc-generated queries or pgx calls with `$n` parameters; never `fmt.Sprintf` of request values into SQL (the dynamic PeriodTable SQL of §12.4 assembles only whitelisted fragments). `audit_log` and `live_event`: no UPDATE/DELETE grant for the app role (a purge job runs as a separate role).
- Uploads: guide files ≤ 2 MB, parsed in memory, never stored; audio ≤ 1 MB, never stored.
- Backups: daily base backup + WAL (PITR 14 days); EU hosting.

### 11.5 Availability
Online only (D08). Forms keep drafts client-side; the API is idempotent on `clientRequestId`, so retries are safe. Target 99,5 % in working hours (06:00–20:00 local).

### 11.6 Performance targets (p95, production data ≤ 5 years, ≤ 20 users)
| Operation | Target |
|---|---|
| `GET /api/table` one page (4 weeks, Detaliu Compartimente) | ≤ 300 ms |
| load-more page | ≤ 200 ms |
| summary / home | ≤ 400 ms |
| any save incl. §6.1 check + reconcile + audit + live event | ≤ 300 ms |
| export whole cycle, all tabs | ≤ 3 s |
| SSE delivery after commit | ≤ 2 s |
| voice: STT ≤ 2 s after stop; command ≤ 3 s typical, 10 s hard cap | |

### 11.7 Golden values (tests must reproduce them)
Fixture: the dev seed (§11.9) loaded from the extracts, clock frozen at **2026-11-18T12:00 Europe/Chisinau**: the test DB session sets `set app.today = '2026-11-18'` (read by `app_today()`, schema.sql) and the app clock is injected with the same instant. Each value was verified by running Python against `design/data/extracts` (and rows 1–12, 19, 26–27 also by SQL against `schema.sql` views on the loaded Hîrbovăț data).
| # | Value | Query / service that must produce it |
|---|---|---|
| 1 | age 407, SAPT 59 on 18.11.2026 | `age_days(seria2026,'2026-11-18')`, `sapt(...)` |
| 2 | 24.02.2026 → age 140, SAPT 21; 22.02.2026 → 138, SAPT 20 | same |
| 3 | 20.10.2025 (13 days) SAPT 2; 21.10.2025 (14 days) SAPT 3 | same |
| 4 | SAPT 59 = 17.11–23.11.2026 (Tue–Mon) | `week_first_day/last_day(59)` |
| 5 | alive end 18.11 ♀ 12 561 · ♂ 794 | `v_farm_sex_day` alive_end, Hîrbovăț, 2026-11-18 |
| 6 | alive start 18.11 ♀ 14 395 · ♂ 1 083 | `v_farm_sex_day` alive_start |
| 7 | H5.1 4 263 · 342; H5.2 4 281 · 323; H6.1 3 542 · 124; H6.2 475 · 5; H7.1, H7.2 0 · 0 | `v_section_sex_day` alive_end |
| 8 | deaths 18.11 ♀ 3 · ♂ 0; sold 18.11 ♀ 1 831 · ♂ 289 (H6.1 676 · 201, H6.2 1 155 · 88) | `v_farm_sex_day` deaths, sold |
| 9 | arrivals ♀ 26 883 · ♂ 3 024 | base of §5.5 at the farm |
| 10 | cycle deaths ♀ 1 521 · ♂ 1 048 → cum. mortality 5,66 % · 34,66 % | §5.5 |
| 11 | sold 12–18.11 ♀ 12 801 · ♂ 1 182 | Σ sold in the cycle |
| 12 | lay 18.11 = 5 914 ÷ 14 395 = 41,1 % | `v_egg_day` Σ total_eggs ÷ Σ hens_start |
| 13 | guide lay SAPT 59 57,6 → −16,5 pct, abatere; guide SAPT 53 64,5; SAPT 56 61,1 | `guide_week` laying ♀ |
| 14 | H5 45,3 % (−12,3), H6 34,9 % (−22,7); H5 below guide since 01.11.2026 | §5.6 per hall; alert `lay.since` |
| 15 | eggs 18.11: 5 914 · incubabile 5 308 (89,8 %) · jos 158 (2,7 %) | §5.6 |
| 16 | Cat 18.11: H5 13 · 417 · 27 + 3 415 = 3 872; H6 8 · 110 · 31 + 1 893 = 2 042 | `egg_production.total_eggs` |
| 17 | egg weight 18.11 farm 66,4 g; SAPT 58 H7 66,3 g, farm 66,6 g | §5.6 egg-weighted |
| 18 | SAPT 59: eggs 11 806, hen-days 29 221, lay 40,4 %, guide 57,6 (−17,2), inc 92,5 %, jos 2,6 %, Cat 0,3 · 6,2 · 1,1 %, "2 din 7 zile" | table week row |
| 19 | SAPT 59 deaths ♀ 10 · ♂ 0, sold ♀ 2 255 · ♂ 433, kg ♀ 4 675,3 · ♂ 329,7 | table week row |
| 20 | SAPT 58: eggs 47 512, hen-days 156 418, lay 30,4 % (end-of-day basis 32,6 rejected), guide 58,8; inc 96,3 %; Cat 0,3 · 2,6 · 0,8 %; deaths 36 · 8; sold 10 546 · 749; alive end 14 826 · 1 227 | table week row |
| 21 | SAPT 58 kg ♀ 22 230,9 · ♂ 1 682,7; g ♀ 155,9 · ♂ 142,2 (kg ÷ bird-days) | Hrană week row |
| 22 | SAPT 57 eggs 49 225, lay 27,6 %, guide 59,9; SAPT 56 eggs 117 929, lay 65,2 %, guide 61,1, inc 98,3 % | table week rows |
| 23 | peak lay SAPT 30: 168 122 ÷ 186 088 = 90,3 % (not 90,4) | readout/stat "Vârf de ouat" |
| 24 | foot "Total SAPT 56–59": eggs 226 472, lay 41,6 %, guide 59,9 (hen-day weighted), deaths ♀ 534 · ♂ 77, sold ♀ 12 801 · ♂ 1 182, inc 97,7 %, kg ♀ 78 898,6 · ♂ 5 996,6 | table `totals` |
| 25 | load-more label "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40" (cycle SAPT 20–59) | table `more` |
| 26 | feed 18.11 ♀ 2 303,2 kg (14 395 × 160 g) · ♂ 154,7 kg; total 2 457,9 kg | `v_feed_consumption` |
| 27 | stock end 18.11 Breeder 2 31 664,9 kg · Cocoși 1 781 kg | `v_feed_stock` |
| 28 | days left Breeder 2 13,2 (avg 16–18.11 2 392 kg/zi), run-out 01.12.2026; Cocoși 10,4 (171 kg/zi), 28.11.2026 | §5.8 |
| 29 | Hîrbovăț deliveries 69 · ♀ 1 114 110 kg · ♂ 78 530 kg · 9 562 437 lei · 4 unpriced 18 160 kg (17.02 2 000 + 1 000, 14.03 6 160, 12.11 9 000) | `feed_delivery` |
| 30 | 13.11.2026 delivery 25 640 kg · 199 000 lei → 7,76 lei/kg | price derivation |
| 31 | cost Hîrbovăț 319,74 lei/pasăre (÷ 29 907), 1,88 lei/ou (÷ 5 073 155); Telița 2 031 768 lei, 64,32 lei/pasăre (÷ 31 590); total 11 594 205 lei, 367,02, 2,29 | §5.9 |
| 32 | weighing Telița H1 ♀ 24.02.2026 2 553 g vs guide wk 20 2 370 g = +7,7 % abatere | §5.10 |
| 33 | 3× warning: H6.1 ♀ 17.11.2026 4 deaths vs 7-day mean 0,9 | §6.2 Z6 |
| 34 | missing: H7.1, H7.2 ♀ Hrană zilnică 13.11.2026 | §5.11 (as of 18.11) |
| 35 | placement Seria 2026 ♀ 28 080 · ♂ 3 510; Telița cycle SAPT 1–21 | Telița migration |
| 36 | guide spans: ouat SAPT 19–65 (94 rows), creștere SAPT 1–25 (50 rows) | `guide_week` |
| 37 | correction 17.11 H5.2 ♀ 1 → 3 (illustrative) leaves H5.2 ♀ end 17.11 4 283, 18.11 4 281 | §6.1 check passes; audit diff "H5.2 · găini moarte 1 → 3" |
| 38 | H7 empty since 14.11.2026 ("goală din 14.11") | `empty_since` |

### 11.8 Testing (Go)
- **Tooling**: the standard `testing` package, table-driven tests (`for _, tc := range []struct{ name string; … }{…} { t.Run(tc.name, …) }`), `github.com/google/go-cmp/cmp` for diffs, golden files under `testdata/` with an `-update` flag. `go test -race ./...` in CI. No assertion framework is required.
- **Unit** (no database): §5 formulas in `calc` (age/week edges 13 → SAPT 2, 14 → 3; partial weeks; rounding half away from zero from full precision: 168 122 ÷ 186 088 → 90,3; guide weighting by hen-/bird-days), §6 message functions (exact strings, agreement găini/cocoși, U+202F and U+2212 in numbers), `voice.ParseRoNumber` (the §8.8 table), `guides.ParseNumber` and the guide parser (G1–G15 fixtures incl. the "645" paste and the week-1 laying file), `num.Num` JSON round trips, `timex.Day` parsing, token sign/verify (expired, wrong user, wrong form, wrong transcript, wrong purpose key).
- **Database for integration tests**: one PostgreSQL 16 per test run. Default: **testcontainers-go** (`github.com/testcontainers/testcontainers-go/modules/postgres`) started in `TestMain` of a shared `internal/testdb` package; when `TEST_DATABASE_URL` is set (CI service container or a local `docker compose up db`), that server is used instead and no container is started. `testdb` applies the goose migrations, runs `migratehistory.Load` on `design/data/extracts` (+ the dev users and `illustrative-events-hirbovat.json`), and keeps the result as a **template database** (`il_seed`); each test that writes gets its own `create database t_<n> template il_seed` (fast file copy), read-only tests share one copy. Every pool used by tests sets `app.today = '2026-11-18'` in `AfterConnect` and the services get `timex.FixedClock(2026-11-18T12:00 Europe/Chisinau)` (§11.1, §11.7).
- **Golden values**: `TestGolden` is one table with the 38 rows of §11.7 (`{id: 12, name: "lay 18.11", got: func(t) any {…}, want: …}`), each calling the real service (`periodtable.Build`, `calc.Stock`, `reports.Cost`, …) — never a hand-written SQL copy of the formula. PeriodTable responses are additionally compared as JSON golden files (`testdata/table_farm_rezumat_cycle.json`, …), which also guards §7.0a (no `null` arrays, camelCase, no lei for the viewer).
- **Rules**: never-below-0 with retroactive edits (raise a death count on 17.11 H5.2 until refused: bound 4 284; the inflow bound 4 383), cycle reconcile (a transfer that empties Telița closes it; a correction that raises it reopens it; a manual cycle; the C2 other-series refusal and its D201 pass), idempotent retries (same `clientRequestId` twice → one entry, second answer 200), soft delete + restore, `STALE` on a wrong `ifVersion`, audit rows per action (`changes` keys as §4.10).
- **Concurrency (transaction + advisory lock)**: (1) two goroutines each save a Vânzare of ♂ 4 from Hîrbovăț H6.2 on 18.11, where 5 cocoși are alive after the day's saved movements — they start together (a barrier channel) on separate pool connections; exactly one gets 201, the other 422 with the §6.1 Vânzare message ("cel mult 1"), and `alive_end ≥ 0` holds everywhere; (2) after a first transfer puts birds into Telița, two transfers in opposite directions (Hîrbovăț → Telița and Telița → Hîrbovăț) start at the same instant and both finish (ascending farm-id lock order: no deadlock; a 5 s `context.WithTimeout` fails the test on a hang); (3) the same two saves with the lock call removed (a test-only `ledger.LockFarms` stub) must be able to break the invariant, proving the test detects a missing lock — run under `-race`, repeated `-count=20` in CI's nightly job.
- **Permissions**: the §3.3 matrix as one table-driven test over every route registered in `cmd/api` (method, path, role → expected status), against an `httptest.Server` running the real handler chain; a route missing from the table fails the test.
- **HTTP contract**: decoder tests (unknown field → 400 with `issues[0].path`, a string where a number is expected, body over the limit → `413 PAYLOAD_TOO_LARGE`), CSRF/Origin middleware, cookie flags on login, CORS headers only for `APP_ORIGIN` when cross-origin mode is configured.
- **Voice**: an eval set of ≥ 150 transcripts (every catalogue example + misheard/ambiguous variants) → expected `action`; run against the chosen providers before release (open question 16) with `go test ./internal/voice -run TestEval -tags eval`; CI runs the server pipeline with `providers/fake`. WebSocket tests use `httptest.Server` + `coder/websocket.Dial`: wrong `Origin` refused, size and time limits close the stream.
- **Live**: SSE replay after disconnect (`Last-Event-ID`), farm filter, price events not delivered to viewers, a lagging subscriber dropped without blocking others, and two API instances in one test process sharing one database (event saved through instance A arrives on a stream of instance B via LISTEN/NOTIFY).

### 11.9 Migrations and seeding
- Migrations: **goose** (`github.com/pressly/goose/v3`), SQL files in `backend/db/migrations/`, embedded with `//go:embed` and applied by `api migrate up` (a subcommand of the `cmd/api` binary, connecting with `MIGRATE_DATABASE_URL`, the owner role; §11.10) as a release step **before** the new version starts — not on every instance start. `0001_baseline.sql` = `backend/schema.sql`, byte for byte (the file already carries the goose annotations; at M0 move it with `git mv backend/schema.sql backend/db/migrations/0001_baseline.sql` and from then on that file is what this spec calls `schema.sql`). Later changes are new numbered files (`0002_roles.sql`, …); views change with `create or replace view` (or `drop view … ; create view …` when columns change) in a new migration, never by editing `0001`. Verify at implementation how the goose version in use serialises concurrent runs.
- Roles: the database owner role runs migrations; the API connects as `il_app`, which gets `select, insert, update, delete` on the tables except `audit_log` and `live_event` (`select, insert` only, §4.10) and `usage` on sequences; `request_replay` gets `select, insert`; a maintenance role `il_maint` (`select, delete` on `live_event`, `login_attempt`, `request_replay`) runs the purge of §9.1 through `MAINT_DATABASE_URL`. Three DSNs, three roles: `MIGRATE_DATABASE_URL` (owner: `api migrate`), `DATABASE_URL` (`il_app`: `api serve`), `MAINT_DATABASE_URL` (`il_maint`: the purge loop inside `api serve`). Roles are created by infrastructure; their grants are migration `0002_roles.sql`.
- `go run ./cmd/migrate-history --seed-dev …`: runs the history migration (§10.3) on the extracts + creates illustrative users (Proprietar, Ion Rusu manager on both farms, Dr. Ceban vet on Hîrbovăț, an Observator) with dev passwords, and loads `illustrative-events-hirbovat.json` (culls, a sexing error, two weighings, the shared correction) for UI states — dev only: the command refuses `--seed-dev` when `APP_ENV=production`.
- `go run ./cmd/create-owner --name "…" --login …`: creates the first owner and prints a generated password once (§6.10 rules; refuses when an active owner exists unless `--force`).

### 11.10 Configuration (environment variables)
Read once at start-up by `internal/platform/config` with `os.LookupEnv`; a missing required variable stops the process with one log line listing all of them. Secrets are never logged (the config type's `LogValue()` redacts them).
| Variable | Example | Notes |
|---|---|---|
| `APP_ENV` | `production` · `staging` · `dev` | dev enables the seed command and text logs |
| `DATABASE_URL` | `postgres://il_app:…@db:5432/impex?sslmode=require` | required for `serve`; the app pool (role `il_app`) |
| `MAINT_DATABASE_URL` | `postgres://il_maint:…@db:5432/impex?sslmode=require` | required for `serve`; a 1-connection pool for the purge (§9.1) |
| `MIGRATE_DATABASE_URL` | `postgres://il_owner:…@db:5432/impex?sslmode=require` | required for `migrate` only; never given to `serve` |
| `DB_MAX_CONNS` | `10` | pool size per instance (+ 1 dedicated LISTEN connection) |
| `LISTEN_ADDR` | `:8080` | |
| `APP_ORIGIN` | `https://app.example` | required; Origin/CSRF/WebSocket checks and CORS (§1, §11.4) |
| `CORS_ENABLED` | `false` | `true` only in the cross-origin deployment of §1 |
| `TRUSTED_PROXIES` | `10.0.0.0/8,127.0.0.1/32` | CIDRs whose `X-Forwarded-For` / `X-Forwarded-Proto` are believed |
| `APP_SECRET` | 32+ random bytes, base64 | required; HKDF root for every signed token (§8.5) |
| `STT_PROVIDER`, `INTENT_PROVIDER` | `azure` · `google` · `anthropic` · `fake` | voice adapters (§8.11); `fake` refused in production |
| `ANTHROPIC_API_KEY`, `INTENT_MODEL`, `STT_*` | – | provider credentials, model id, region (EU) |
| `LOG_LEVEL`, `LOG_FORMAT` | `info`, `json` | §11.11 |
Thresholds and limits stay in `app_setting` (§4.11), not in the environment.

### 11.11 Logging (log/slog)
- `log/slog` with `slog.NewJSONHandler(os.Stdout, …)` in production (text handler in dev); one logger built in `main` and passed down (no package-level globals beyond `slog.SetDefault`).
- An access-log middleware writes one line per request: `request_id` (from `X-Request-ID` set by the proxy, else generated and returned in the response header), `method`, the matched route pattern (`r.Pattern`, e.g. `PUT /api/entries/{id}`), `status`, `duration_ms`, `user_id`, `bytes`. Errors (5xx) add the error chain; panics are recovered by a middleware, logged with the stack and answered with the 500 envelope.
- Never logged: request or response bodies, cookies, `Authorization`, passwords and one-time passwords, tokens, voice transcripts and audio (§8.9: voice lines carry `run_id`, intent id, provider, timings, token counts only), Motiv texts.

### 11.12 Health checks and shutdown
- `GET /healthz` (liveness): 200 `ok` while the process serves; no database call.
- `GET /readyz` (readiness): 200 when `pool.Ping` answers within 1 s, the goose version in the database equals the newest embedded migration, and the LISTEN connection is up; else 503 with the failing check names. Both paths are outside `/api`, unauthenticated, not logged at info level, and need not be exposed by the public proxy.
- Shutdown: `signal.NotifyContext(ctx, SIGINT, SIGTERM)`; readiness turns 503 first, then `srv.Shutdown` with a 20 s budget (SSE streams and the voice WebSocket are closed by `RegisterOnShutdown` hooks, §9.1), then the listener and the pool close.

### 11.13 Dockerfile
```dockerfile
# build: Go current stable (pin the exact version that go.mod's toolchain line names)
FROM golang:<go-version> AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate-history ./cmd/migrate-history \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/create-owner ./cmd/create-owner

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /usr/local/bin/
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
CMD ["serve"]
```
The binaries embed the migrations and tzdata, so the image needs nothing else (CA certificates for the provider APIs come with the distroless base). Release: run `api migrate up` once (a one-off container with the same image), then roll the `serve` containers. The extracts are not in the image: `migrate-history` runs once from a developer machine or a one-off job with the repo mounted. Verify the base image tag at implementation.

---

## 12. Recommended implementation (Go + PostgreSQL, D202)

### 12.1 Shape
- **Domain services** own every rule of §5–§6: `ledger` (running counts, the §6.1 bounds, `ReconcileCycles` §6.4, farm locks), `calc` (§5 formulas: age/week, lay, feed, stock, cost, weighing, missing, alerts), `entries` (create/edit/correct/delete/undo/restore with audit), `reports` (feed: consumption, stock, deliveries, cost; production), `guides` (parse, validate), `periodtable` (PeriodTable builder), `voice`. **Handlers are thin**: decode (`httpx.Decode`) → `Validate()` (§7.0b) → authorize (`auth`) → call a service → map to the DTO → `httpx.JSON`. Services take `context.Context`, an `auth.Principal` and a typed request; they never see `*http.Request`.
- **One transaction per save** (READ COMMITTED): `pg_advisory_xact_lock` on every touched farm in ascending id order (§6.1 step 0) → write rows → §6.1 check → `ReconcileCycles` → `audit_log` → `live_event` + `pg_notify` → commit (the notification is delivered at commit). Setup writes (farm, section, feed type, series, cycle) take the same farm lock when they touch a farm. The worked example is §12.5.
- **Heavy reads in SQL** (views §4.12 + window functions, one pass per page); presentation shaping (groups, labels, reasons, Romanian strings) in Go.
- No ORM, no web framework, no dependency-injection container: `cmd/api/main.go` builds the pool, the services and the handlers by hand and registers the routes.

### 12.2 Concrete stack
- **Go**: the current stable release at implementation time (verify; ≥ 1.24 is required for the `omitzero` tag option and `crypto/hkdf`, ≥ 1.22 for method/wildcard routing; 1.25 adds `http.CrossOriginProtection`). Pin it with the `go` and `toolchain` lines of `go.mod` and the same version in the Dockerfile and CI.
- **Router: the standard library `net/http.ServeMux`** with Go 1.22+ patterns (`mux.Handle("PUT /api/entries/{id}", …)`, `r.PathValue("id")`). Chosen over chi: the API is ~90 plain REST routes; method + wildcard patterns, precedence rules and conflict detection at registration (the mux panics on ambiguous patterns) cover it; middleware is ordinary `func(http.Handler) http.Handler` composition; `r.Pattern` gives the route template for logs. One dependency less, nothing to learn beyond the standard library.
- **PostgreSQL 16** (EU-hosted, managed, PITR), extension `btree_gist`.
- **pgx/v5** (`github.com/jackc/pgx/v5`, `pgxpool`) as driver and **sqlc** (`sqlc-dev/sqlc`, `sql_package: pgx/v5`) for every static query; hand-written pgx SQL only for the dynamic reads listed in §12.4. Why not GORM: the schema is SQL-first — views, partial unique indexes, an exclusion constraint, a generated column, window functions, advisory locks, LISTEN/NOTIFY, `numeric` money — which GORM can't model and would hide behind reflection; GORM's auto-migration would compete with goose and `schema.sql`; its query builder produces SQL you only see at run time. sqlc keeps the SQL readable in `db/queries/*.sql`, checks it against the schema at generate time, and emits plain typed Go with no runtime reflection; pgx adds what the app needs beyond that (`CopyFrom` for the migration, `WaitForNotification` for live updates, batches, native `pgtype`).
- **goose** (`github.com/pressly/goose/v3`) for migrations (§11.9); sqlc reads the same directory (it ignores `-- +goose Down` sections).
- **shopspring/decimal** + `jackc/pgx-shopspring-decimal` (§11.2); **google/uuid**; **golang.org/x/crypto/argon2** (§11.4); `crypto/hmac`, `crypto/sha256`, `crypto/hkdf`, `crypto/rand` from the standard library; **golang.org/x/text** (normalisation for name keys and file names); **golang.org/x/time/rate** (voice per-minute cap).
- **excelize** (`github.com/xuri/excelize/v2`) for `.xlsx` export and guide import; `encoding/csv` for CSV/TSV.
- **coder/websocket** (`github.com/coder/websocket`) for `/api/voice/stream`.
- **log/slog** (§11.11).
- Voice providers: the Anthropic Go SDK for the recommended intent provider; the STT vendor's SDK or HTTP API (§8.11).
- Tests: `testing`, testcontainers-go (postgres module), go-cmp (§11.8). Playwright end-to-end flows belong to the front-end repo and run against a real API.
- Lint: `go vet`, `staticcheck`, `gofmt` in CI (and `sqlc vet` / `sqlc diff` so generated code never drifts from `db/queries`).
Library names only: take the current release of each when the module is created, and verify any API marked "verify at implementation".

### 12.3 Module layout
The Go module lives in `backend/` (module path chosen at M0, e.g. `<git host>/<org>/impex-line/backend`).
```
backend/
  go.mod  go.sum  sqlc.yaml  Dockerfile  Makefile        (make generate | test | lint | run)
  SPEC.md                                                (this file)
  cmd/
    api/                main.go (serve | migrate up|status), routes.go (every §7 route → handler), wiring.go
    migrate-history/    main.go: flags, calls internal/migratehistory (§10.3; --replace, --seed-dev)
    create-owner/       main.go (§11.9)
  db/
    migrations/         0001_baseline.sql (= schema.sql, goose), 0002_roles.sql, …
    queries/            sqlc input, one file per module: auth.sql users.sql farms.sql sections.sql series.sql cycles.sql
                        entries.sql ledger.sql calc.sql reports.sql prices.sql feedtypes.sql guides.sql audit.sql live.sql replay.sql voice.sql settings.sql
    embed.go            //go:embed migrations/*.sql  → var Migrations embed.FS
  internal/
    db/                 pool.go (app pool + 1-conn maintenance pool; AfterConnect: decimal codec, timezone UTC, app.today in tests), tx.go (InTx, Tx),
                        errors.go (pgconn codes → typed errors); dbq/ = sqlc output (generated, never edited)
    platform/
      config/           env → Config (§11.10)
      httpx/            Decode, JSON, WriteError, APIError/Refusal/Warning/Issue, Optional[T], WriteMeta, middleware
                        (request id, access log, recover, auth, CSRF/Origin, CORS, security headers, body limits)
      timex/            Day, LocalTime, Clock, Chisinau, Today, week helpers
      num/              Num (decimal JSON), rounding helpers
      fmtx/             §11.3 formatters (U+202F, U+2212, "41,1%", dd.mm.yyyy, SAPT labels)
      messages/         ro.go: every Romanian string of §6–§8 as a typed function
      textkey/          Key (name_key), ASCII (file names)
      tokens/           HMAC sign/verify with HKDF purpose keys (§8.5)
    auth/               login/logout/me, sessions, argon2id, throttling, Principal, RequireRole, CanSeeFarm
    users/              §7.19 + one-time passwords
    farms/              §7.16 farms and sections, §7.2 bootstrap, §7.3 farm/section pages, §7.6 summary, §7.5 journal, §7.9 cycles routes
    series/             §7.15 series, placement entry, age change, close, series page + cycle track
    entries/            §7.8 form state, creates (daily, feed-daily, weighing, transfer, sale, delivery), edit/correct/delete/
                        undo/restore, Introdus azi; dto.go, validate.go, service.go, handler.go
    ledger/             LockFarms, CheckCounts (§6.1), ReconcileCycles (§6.4), alive/inflow lookups shared by entries, series, farms
    calc/               §5 formulas as pure functions over rows (no SQL), + the SQL readers they need; alerts (§5.12), missing (§5.11)
    periodtable/        §7.4 /api/table and §7.4b /api/charts builder (dynamic SQL, §12.4), groups, rows, totals, load-more
    reports/            §7.11 production, §7.12 feed (stock, deliveries, cost), §7.7 home (Acasă blocks per role)
    prices/             §7.13 Prețuri, price undo
    feedtypes/          §7.18 /api/feed-types (list with usedAt and stock per farm, create, edit, retire/reactivate, delete; §6.13 FT1–FT5)
    guides/             §7.17 + parser/validator (§6.12, §10.2)
    audit/              audit row builder (changes, record types, summary), §7.10 Istoric, §7.19b setup undo
    live/               Hub, Listener (LISTEN/NOTIFY), SSE handler, Emit (insert live_event + pg_notify), purge (§9)
    voice/              catalogue.go (§8.5), tools.go (BuildTools §8.11), prompt.go, pipeline.go (§8.6), answers.go (§8.7),
                        ronumbers.go (§8.8), fastpath.go (§8.10), stt.go / stream.go / command.go handlers, providers/{anthropic,<stt>,fake}
    export/             §7.14 / §10.1 xlsx writers per page (uses periodtable and reports)
    migratehistory/     §10.3 loader + verification (used by cmd/migrate-history and internal/testdb)
    testdb/             test-only: container or TEST_DATABASE_URL, template DB with the golden fixture (§11.8)
```
Rules: a module exposes `NewService(…)`, `NewHandler(…)` and `(*Handler).Routes(mux *http.ServeMux, mw httpx.Chain)`; modules call each other only through services (or the shared `ledger`/`calc`), never through handlers; `dbq` types stay inside services (DTOs are built in each module's `dto.go`). Import direction: `cmd` → modules → `ledger`/`calc`/`audit`/`live` → `db`/`platform`; no cycles (the compiler enforces it).

### 12.4 Data access: sqlc + pgx
`sqlc.yaml` (verify keys against the sqlc version in use):
```yaml
version: "2"
sql:
  - engine: postgresql
    schema: db/migrations
    queries: db/queries
    gen:
      go:
        package: dbq
        out: internal/db/dbq
        sql_package: pgx/v5
        emit_interface: true
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        overrides:
          - db_type: uuid
            go_type: github.com/google/uuid.UUID
          - db_type: uuid
            nullable: true
            go_type: { import: github.com/google/uuid, type: UUID, pointer: true }
          - db_type: pg_catalog.numeric
            go_type: github.com/shopspring/decimal.Decimal
          - db_type: pg_catalog.numeric
            nullable: true
            go_type: github.com/shopspring/decimal.NullDecimal
          - db_type: date
            go_type: <module>/internal/platform/timex.Day
          - db_type: date
            nullable: true
            go_type: { import: <module>/internal/platform/timex, type: Day, pointer: true }
```
`schema.sql` was checked in PostgreSQL 16 (it loads, the views return the golden values) but `sqlc generate` has not been run against it yet (no module download was possible when this was written): run it at M0 and fix any query-side typing per the table below. If sqlc's own type inference falls short on a query, enable its database-backed analysis against the migrated dev/CI database (`database:` block in `sqlc.yaml`; verify at implementation) rather than casting around it.

**What `schema.sql` does for sqlc**: no psql meta-commands and no `BEGIN/COMMIT`; function bodies dollar-quoted inside `-- +goose StatementBegin/End`; every view lists its columns (no `t.*`) and casts every computed column (`::integer`, `::numeric`, `::boolean`, `::date`), so `v_section_sex_day.alive_end` is `int32`, `v_feed_stock.stock_end` is `decimal.Decimal`, `v_egg_day.lay_pct` a nullable decimal, instead of `interface{}`. View columns are generated as nullable by default: wrap with `coalesce` in the query or read pointers.

**Constructs sqlc can't type or generate, and how to query them**
| Construct | Where | How |
|---|---|---|
| Dynamic column sets and filters | PeriodTable, charts, export (`tab` × `detail` × `types` × grain) | hand-written in `periodtable/sql.go`: a fixed per-tab SQL with every column of that tab, filters as parameters, the only variable fragments chosen from a whitelist (never request text); executed with `tx.Query(ctx, sql, args…)` and scanned with `pgx.CollectRows(rows, pgx.RowToStructByName[deltaRow])` into structs with `db:"…"` tags |
| `set app.today` | tests | `SET` takes no parameters: use `conn.Exec(ctx, "select set_config('app.today', $1, false)", day)` inside `AfterConnect(ctx, conn *pgx.Conn)` |
| `generate_series` over parameter dates | dense day lists for a page | cast the output: `select g::date as date from generate_series(sqlc.arg(from_date)::date, sqlc.arg(to_date)::date, interval '1 day') g` (sqlc then sees `date`) |
| Arrays of ids, overlap filters | farm scope, audit filter | `farm_id = any(sqlc.arg(farm_ids)::uuid[])`, `farm_ids && sqlc.arg(farm_ids)::uuid[]` (explicit casts so sqlc picks `[]uuid.UUID`) |
| Optional filters | journal, audit | `(sqlc.narg(section_id)::uuid is null or section_id = sqlc.narg(section_id))` |
| `jsonb` (`audit_log.changes`, `old_values`, `new_values`, `live_event.payload`, `app_setting.value`) | audit, live, settings | sqlc gives `[]byte`; marshal/unmarshal typed Go structs (`[]entries.Change`, live payload structs) in the service |
| Enum arrays / enum filters | record types, actions | cast in SQL (`sqlc.arg(kinds)::entry_kind[]`); sqlc's generated enum types have `Valid()` |
| Generated column `egg_production.total_eggs` | eggs insert | never in an INSERT/UPDATE column list; read it like any column |
| Advisory lock | every write | `-- name: LockFarm :exec` / `select pg_advisory_xact_lock(hashtextextended(sqlc.arg(farm_id)::uuid::text, 0));` |
| Notification | every write | `-- name: NotifyLive :exec` / `select pg_notify('live_event', sqlc.arg(id)::bigint::text);` |
| LISTEN / `WaitForNotification` | `live.Listener` | pgx connection API directly, not sqlc |
| Bulk inserts | `migratehistory` | sqlc `:copyfrom` or `pgx.CopyFrom` (COPY bypasses nothing: constraints and partial unique indexes still apply) |
| Constraint errors | every write | `var pgErr *pgconn.PgError; errors.As(err, &pgErr)` → `pgErr.Code` (`23505`, `23P01`, `23514`, `23503`) + `pgErr.ConstraintName` → the typed error of §7.0a (`bird_event_daily_uq`/`egg_production_uq`/`feed_allocation_uq`/`weighing_uq` → `ALREADY_ENTERED` with the owning entry id looked up by key) |

### 12.5 Worked example: the Zilnic save, end to end
`POST /api/entries/daily` (§7.8, rules §6.1–§6.2). Decode → validate → transaction → lock → insert → §6.1 check → `ReconcileCycles` → audit → `live_event` + NOTIFY → replay row → response. A dry run runs the same writes and checks and then rolls back (§7.0). Code is a sketch for structure; helper names are this spec's, signatures may differ.

```go
// internal/entries/handler.go
func (h *Handler) CreateDaily(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())                       // set by the auth middleware (401 before this)
	req, err := httpx.Decode[DailyCreateRequest](w, r, 64<<10) // 1. decode: 413 over the limit, unknown field → 400
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if issues := req.Validate(h.boot.Codes(r.Context(), req.FarmID)); len(issues) > 0 { // 2. level-1 validation (§7.0b)
		httpx.WriteError(w, httpx.Invalid(issues))
		return
	}
	res, err := h.svc.CreateDaily(r.Context(), p, req) // 3–12
	if err != nil {
		httpx.WriteError(w, err) // 403/409/422 envelope, or 500
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body) // the exact bytes request_replay stores
}
```
```go
// internal/entries/daily.go

// Result is an encoded response: 201 for a save, 200 for a dry run or a replay.
type Result struct {
	Status int
	Body   []byte
}

func (s *Service) CreateDaily(ctx context.Context, p auth.Principal, req DailyCreateRequest) (Result, error) {
	today := timex.Today(s.clock)
	// 3. authorization and the checks that need no farm state
	switch {
	case !p.Role.In(auth.Vet, auth.Manager, auth.Owner):
		return Result{}, httpx.Forbidden()
	case !p.CanSeeFarm(req.FarmID):
		return Result{}, httpx.FarmForbidden() // "Nu ai acces la această fermă."
	case req.Date.After(today):
		return Result{}, httpx.Refused(messages.FutureDate(req.Date, today))
	case p.Role == auth.Vet && req.Date != today:
		return Result{}, httpx.PastDayVet() // Z1 "Pentru zilele trecute cere managerului."
	}
	if req.ClientRequestID != nil && !req.DryRun { // idempotent retry: the stored bytes, unchanged
		if prev, ok, err := s.replay(ctx, p, *req.ClientRequestID); err != nil || ok {
			return prev, err
		}
	}
	src, transcript := s.tokens.VoiceSource(p, "zilnic", req.WriteMeta) // §8.9: manual unless the voiceToken verifies

	var out Result
	err := s.db.InTx(ctx, func(tx db.Tx) error { // 4. one READ COMMITTED transaction
		q := tx.Q                                                    // *dbq.Queries bound to tx
		if err := ledger.LockFarms(ctx, q, req.FarmID); err != nil { // 5. pg_advisory_xact_lock, ascending ids
			return err
		}
		st, err := s.loadDailyState(ctx, q, req.FarmID, req.Date) // under the lock: alive_start + inflows saved that day,
		if err != nil {                                           // egg units, live value keys, mean7, cycle + series status
			return err
		}
		if st.SeriesClosed {
			return httpx.SeriesClosed(st.SeriesName) // 409 SERIES_CLOSED
		}
		plan := planDaily(st, req)  // 6. level-2 rules: Z3 Z4 Z5 Z7 Z9 refusals, Z6 warnings, rows, toast, summary
		preview := plan.Preview()   // DryRunResponse{Effects, Warnings, Refusals}, slices never nil
		if len(plan.Refusals) > 0 { // refused rows are never inserted
			if req.DryRun {
				return answerDryRun(&out, preview) // 200 with the refusals in the body
			}
			return plan.RefusedError() // 422 REFUSED (409 ALREADY_ENTERED for Z9)
		}
		ins, err := q.InsertEntry(ctx, dbq.InsertEntryParams{ // 7. the save and its child rows
			Kind: dbq.EntryKindDaily, FarmID: req.FarmID, Date: req.Date,
			Source: src, ClientRequestID: req.ClientRequestID, CreatedBy: &p.UserID,
		})
		if err != nil {
			return s.mapInsertErr(ctx, err, req) // 23505 entry_client_request_uq → errReplay
		}
		for _, ev := range plan.BirdEvents { // mortality, cull + reason, sexing_error per section × sex
			if err := q.InsertBirdEvent(ctx, ev.Params(ins.ID)); err != nil {
				return s.mapInsertErr(ctx, err, req) // 23505 bird_event_daily_uq → 409 ALREADY_ENTERED + entryId
			}
		}
		for _, eg := range plan.Eggs { // one egg_production per hall (total_eggs is generated)
			if err := q.InsertEggProduction(ctx, eg.Params(ins.ID)); err != nil {
				return s.mapInsertErr(ctx, err, req)
			}
		}
		bad, err := ledger.CheckCounts(ctx, q, plan.Affected, req.Date) // 8. §6.1 over every later day, after the change
		if err != nil {
			return err
		}
		if len(bad) > 0 {
			refusals := refusalsFromBounds(bad, plan) // field + banner messages with max/min bounds
			if req.DryRun {
				preview.Refusals = append(preview.Refusals, refusals...)
				return answerDryRun(&out, preview)
			}
			return httpx.RefusedWith(refusals, plan.Warnings)
		}
		cyc, err := ledger.ReconcileCycles(ctx, q, req.FarmID, req.Date, today) // 9. §6.4 (a save can empty a farm)
		if err != nil {
			return err
		}
		if req.DryRun { // every write above is undone by the rollback
			preview.Effects = append(preview.Effects, cycleEffects(cyc)...)
			return answerDryRun(&out, preview)
		}
		if err := audit.Write(ctx, q, audit.Row{ // 10. one audit row per save action (§4.10)
			UserID: &p.UserID, Action: dbq.AuditActionCreate, Entity: "entry", EntityID: ins.ID,
			EntryID: &ins.ID, EntryKind: dbq.EntryKindDaily, RecordTypes: plan.RecordTypes,
			FarmIDs: []uuid.UUID{req.FarmID}, RecordDate: &req.Date, Summary: plan.Summary,
			NewValues: plan.Snapshot(), Source: src, VoiceTranscript: transcript,
		}); err != nil {
			return err
		}
		if err := audit.WriteCycleChanges(ctx, q, cyc); err != nil { // cycle_auto_close / cycle_reopen, user_id null
			return err
		}
		alive, err := q.FarmAliveEnd(ctx, dbq.FarmAliveEndParams{FarmID: req.FarmID, Date: today})
		if err != nil {
			return err
		}
		evs := append([]live.Event{ // 11. live_event rows + pg_notify (delivered at commit)
			live.EntryCreated(ins.ID, dbq.EntryKindDaily, plan.Live(), p),
			live.CountsChanged(req.FarmID, req.Date, plan.SectionCodes, alive.F, alive.M),
		}, live.CycleEvents(cyc)...)
		if err := live.Emit(ctx, q, evs...); err != nil {
			return err
		}
		body, err := httpx.Marshal(EntryWriteResponse{ // 12. encode once; the same bytes go to the client and the replay row
			Entry:    httpx.EntryRef{ID: ins.ID, Kind: "daily", Version: timex.Version(ins.VersionAt)},
			Warnings: plan.Warnings, // Z6 MORTALITY_SPIKE etc., never nil
			Toast:    plan.Toast(),  // "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0"
			Alive:    &httpx.SexPair[int]{F: int(alive.F), M: int(alive.M)},
			AsOf:     timex.LocalTime(s.clock.Now()),
		})
		if err != nil {
			return err
		}
		if req.ClientRequestID != nil {
			if err := q.InsertReplay(ctx, dbq.InsertReplayParams{
				ClientRequestID: *req.ClientRequestID, UserID: p.UserID,
				Route: "POST /api/entries/daily", Status: http.StatusCreated, Body: body,
			}); err != nil {
				return s.mapInsertErr(ctx, err, req) // 23505 request_replay_pkey → errReplay
			}
		}
		out = Result{Status: http.StatusCreated, Body: body}
		return nil
	})
	if errors.Is(err, errReplay) { // a concurrent duplicate committed first
		prev, _, rerr := s.replay(ctx, p, *req.ClientRequestID)
		return prev, rerr
	}
	return out, err
}

// answerDryRun encodes the preview as a 200 and asks InTx to roll back.
func answerDryRun(out *Result, preview DryRunResponse) error {
	body, err := httpx.Marshal(preview)
	if err != nil {
		return err
	}
	*out = Result{Status: http.StatusOK, Body: body}
	return db.ErrRollback
}

// replay returns the stored response of an earlier save with the same key (200, same bytes).
func (s *Service) replay(ctx context.Context, p auth.Principal, key uuid.UUID) (Result, bool, error) {
	row, err := s.q.GetReplay(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	if row.UserID != p.UserID {
		return Result{}, false, httpx.Stale() // a key is never shared between users
	}
	return Result{Status: http.StatusOK, Body: row.Body}, true, nil
}
```
```go
// internal/db/tx.go
var ErrRollback = errors.New("rollback")

func (s *Store) InTx(ctx context.Context, fn func(Tx) error) error {
	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(t pgx.Tx) error {
		return fn(Tx{Tx: t, Q: dbq.New(t)})
	})
	if errors.Is(err, ErrRollback) { // fn's error rolled the transaction back; for a dry run that is success
		return nil
	}
	return err
}

// internal/ledger/lock.go
func LockFarms(ctx context.Context, q *dbq.Queries, ids ...uuid.UUID) error {
	ids = slices.Clone(ids)
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) }) // = ascending text order
	for _, id := range slices.Compact(ids) {
		if err := q.LockFarm(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// internal/platform/timex/version.go
// Version formats an entry's coalesce(updated_at, created_at) as RFC 3339 with microseconds (PostgreSQL's precision),
// fixed width so equal instants give equal strings; ifVersion is parsed back and compared in SQL.
func Version(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000000Z07:00") }
```
```sql
-- db/queries/ledger.sql (sqlc)
-- name: LockFarm :exec
select pg_advisory_xact_lock(hashtextextended(sqlc.arg(farm_id)::uuid::text, 0));

-- name: NegativeCounts :many
-- §6.1 step 2: the (section, sex) pairs that would go below 0 on or after the earliest changed day.
select section_id, sex,
       min(date)::date         as first_negative_date,
       max(date)::date         as last_negative_date,
       min(alive_end)::integer as min_alive_end
from v_section_sex_day
where section_id = any(sqlc.arg(section_ids)::uuid[])
  and date >= sqlc.arg(from_date)::date
  and alive_end < 0
group by section_id, sex;

-- name: FarmAliveEnd :one
select coalesce(sum(alive_end) filter (where sex = 'F'), 0)::integer as f,
       coalesce(sum(alive_end) filter (where sex = 'M'), 0)::integer as m
from v_farm_sex_day
where farm_id = sqlc.arg(farm_id) and date = sqlc.arg(date)::date;

-- db/queries/entries.sql
-- name: InsertEntry :one
insert into entry (kind, farm_id, peer_farm_id, date, series_id, buyer, note, source, client_request_id, created_by)
values (sqlc.arg(kind), sqlc.arg(farm_id), sqlc.narg(peer_farm_id), sqlc.arg(date), sqlc.narg(series_id),
        sqlc.narg(buyer), sqlc.narg(note), sqlc.arg(source), sqlc.narg(client_request_id), sqlc.narg(created_by))
returning id, coalesce(updated_at, created_at)::timestamptz as version_at;

-- name: UpdateEntry :one
-- Edit / correction header; zero rows = the version changed since the client read it → 409 STALE.
update entry
set updated_by = sqlc.arg(updated_by), updated_at = now(), corrected = corrected or sqlc.arg(is_correction)::boolean
where id = sqlc.arg(id) and deleted_at is null
  and coalesce(updated_at, created_at) = sqlc.arg(if_version)::timestamptz
returning id, updated_at::timestamptz as version_at;

-- name: InsertBirdEvent :exec
insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason, peer_section_id, note)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- db/queries/replay.sql
-- name: InsertReplay :exec
insert into request_replay (client_request_id, user_id, route, status, body) values ($1, $2, $3, $4, $5);

-- name: GetReplay :one
select user_id, status, body from request_replay where client_request_id = $1;

-- db/queries/live.sql
-- name: InsertLiveEvent :one
insert into live_event (farm_ids, type, payload) values (sqlc.arg(farm_ids)::uuid[], sqlc.arg(type), sqlc.arg(payload)) returning id;

-- name: NotifyLive :exec
select pg_notify('live_event', sqlc.arg(id)::bigint::text);
```
Notes on the example:
- **Order matters**: the lock is taken before any count is read, so `loadDailyState`, the per-day bound (Z5), `CheckCounts` and `ReconcileCycles` all see a stable farm; a second save of the same farm waits at `LockFarm` until this one commits, then sees its rows (READ COMMITTED reads the newest committed data per statement).
- **Dry run** (§7.0): the same path up to `ReconcileCycles`, so the preview includes the §6.1 bounds over later days and the cycle effects; `answerDryRun` encodes `{effects, warnings, refusals}` as 200 and returns `db.ErrRollback`, which `InTx` turns into a rollback without an error. No audit, live event or replay row is written; a pending `pg_notify` would be dropped by the rollback anyway.
- **Idempotency** (§7.0): the response is encoded once (`httpx.Marshal`, the same encoder settings as `httpx.JSON`) and those bytes are both sent and stored in `request_replay` inside the save's transaction; a repeat with the same key gets them back unchanged with status 200. `entry_client_request_uq` and the `request_replay` primary key settle the race between two identical concurrent requests (the loser maps the 23505 to `errReplay`, rolls back and answers from the winner's row).
- **Version** (§7.0 optimistic concurrency): `InsertEntry`/`UpdateEntry` return `version_at` from the database, never the Go clock; `ifVersion` is compared in the `UPDATE … where coalesce(updated_at, created_at) = $if_version`, under the farm lock.
- **Errors inside `InTx`** roll the transaction back; nothing is half-saved, and no notification is sent (it would only be delivered at commit).
- **Edits and corrections** (`PUT /api/entries/{id}`) use the same steps with the stored entry's farm (both farms for a transfer, ascending ids), `UpdateEntry` for the header, soft-delete/update/insert of child rows by key, `action = update`, the diff as `changes`.
- The same skeleton serves every write: only `loadXState`, `planX` and the child inserts change per kind.

### 12.6 Build order (milestones follow the screens' build order, screens.md)
| M | Deliver | Done when |
|---|---|---|
| M0 | Go module in `backend/`, `go.mod` toolchain pin, `sqlc.yaml`, `git mv` of `schema.sql` to `db/migrations/0001_baseline.sql` + `0002_roles.sql`, `cmd/api` (`serve`, `migrate`), `platform` (config, slog, httpx decode/JSON/errors/middleware, timex, num, fmtx, tokens), `db` (pool, InTx, error mapping), auth (login/logout/me, argon2id, sessions, throttling, CSRF/Origin), `/healthz` `/readyz`, Dockerfile, CI (`gofmt`, `go vet`, `staticcheck`, `sqlc diff`, `go test -race` with testcontainers), `cmd/create-owner` | login works end to end through the reverse proxy; auth rows of the §3.3 matrix test pass; image builds and starts |
| M1 | `internal/migratehistory` + `cmd/migrate-history` (+ `--seed-dev`), `internal/testdb` template fixture, `calc` age/week/counts, golden test harness | §10.3 verification passes; goldens 1–12, 26–27 |
| M2 | `periodtable` (`/api/table`) for farm Rezumat + load-more + totals; farm page + summary (S4); JSON golden files | goldens 13–25 |
| M3 | Hrană zilnică form state + save (S7) on the §12.5 skeleton, stock/days left (S11 Stoc), feed report + Consum | goldens 26–28 |
| M4 | series page + track (S3); S11 Livrări, Cost; Prețuri (S12) + price undo tokens; Livrare (S13) | goldens 29–31; viewer JSON has no price keys |
| M5 | Zilnic (S6) incl. eggs/Cat. (§12.5), entries edit/correct/delete/undo/restore with audit, `ledger.CheckCounts` | goldens 33, 37; correction refusals; the concurrency tests of §11.8 |
| M6 | remaining tabs (Păsări + Jurnal, Greutate, Producție), section/hall pages (S5), `/api/charts` | |
| M7 | Transfer (S9), Vânzare (S10), `ledger.ReconcileCycles`, manual cycle (D201), guide change | cycle tests; opposite-direction transfer test |
| M8 | Acasă per role (S2), alerts, Lipsesc date | goldens 14, 34 |
| M9 | Producție report (S14), Cântărire (S8), Introdus azi (S15), Corectează (S16), Istoric (S17), `export` (.xlsx with excelize) | golden 32; export of a whole cycle ≤ 3 s |
| M10 | Setări: Serii (S18), Ferme (S19), Ghiduri + import (S20), Tipuri de hrană (S21, `internal/feedtypes`), Utilizatori (S22), setup undo tokens | G1–G15, FT1–FT5, U1–U9; full §3.3 matrix |
| M11 | `live`: Hub, Listener (LISTEN/NOTIFY), SSE handler, purge | replay, filter, lagging-subscriber and two-instance tests |
| M12 | `voice`: providers (STT, Anthropic intent, fake), `/stt`, WebSocket `/stream`, catalogue + `BuildTools`, pipeline, answers, fast path, tokens, eval set | eval ≥ 95 % intent accuracy on the set *(target, to agree)*; WebSocket limit tests |

---

## 13. Open questions and assumptions

**From `design/kit/open-questions.md`** (status for the backend):
1. One series per farm — decided (D201): a manual new cycle may pick another active series; birds of that series may then arrive although the farm still holds the old ones (§6.4 C2, C4; §6.14 S7). Remaining old birds are counted in the new cycle from its start.
2. History import — done by the CLI from the extracts (§10.3); from which date the app is the source of truth is the owner's call (go-live date).
3. Real guides — the extracts' guides are imported with illustrative names; replace them through Ghiduri when the real files arrive.
4. Egg total excludes Jos — implemented (D18).
5. Thresholds — in `app_setting` (D19 values).
6. Feed types and their sex — illustrative (D30); editable in S21.
7. Tone "tu" — all messages use "tu".
8. Incubation — out (D182). 9. Cat. 1–3 — counts per hall per day (D183). 10. Lay guide — Productivitate col I (D184).
11. Hîrbovăț feed products — names illustrative (Breeder 2 / Cocoși).
12. Workbook total-row errors — the backend recomputes from sections; the typed 460 kg rows stay as `kg_recorded`.
13. **Speech and AI provider, EU processing, no retention** — open; adapters ready (§8.11).
14. **Voice cost and daily cap** — default 200/day/user (`app_setting`), message proposed.
15. Wake word — not built.
16. Number accuracy in the hall — always prefill; test in a hall; the eval set (§11.8).
17. Russian words / no diacritics — the prompt accepts unaccented Romanian; Russian words → `neinteles` until decided.
18. Transcript in the audit and the notice — implemented as designed; legal consent unanswered.
19. Spoken answers — not built (written only).
20. Shortcut on ChromeOS / Windows layout switch — client only.
21. iPhone microphone permission per session — client only; server unaffected.

**Assumptions made in this spec (to review with the user)**
- A1. (decided, D200) Observator sees no prices anywhere: no lei, no Cost, no price endpoints (`403` on `/api/deliveries*` price routes).
- A2. Managers (and the owner) may create records for past days without a Motiv (filling missing days, D65); only changes to existing records need one.
- A3. Cumulative mortality counts deaths + culls (the register counts culls as deaths, D165); the base is the birds that entered the scope during the cycle.
- A4. A manual new cycle on date X ends the old one on X − 1 (records of X belong to the new cycle); the dialog copy "se închide pe {X}" means "from X".
- A5. A correction that would reopen an emptied cycle is refused when the farm has a later cycle (§6.4 step 4).
- A6. Feed cost of a cycle includes deliveries after the previous cycle's end and before the first birds (pre-stocking, e.g. Hîrbovăț 17.02.2026).
- A7. Days-left window = the last 3 complete days (today included once today's Hrană zilnică is complete).
- A8. Farm-level weight in a week with several section weighings = bird-weighted mean.
- A9. Producție foot "Găini" = hens at the start of the oldest loaded week.
- A10. Istoric lists every action except plain creates, plus creates that carry a voice transcript.
- A11. Migration: Telița feed type by date windows; Telița residuals, link differences and any day the register would take a section below 0 (H3 −9) become `adjustment` rows "diferență registru" on that day.
- A12. Proposed copy (marked *(proposed)* in §6–§8) has no board yet.
- A13. Weighing weight range 1–10 000 g, egg weight 20–120 g, sample size ≥ 1.
- A14. Sessions slide (30 days after last use); a password reset, deactivation, or any role/farm change signs the user out everywhere.
- A15. One feed type per section × sex × day.
- (decided, D201) A manual new cycle may pick the series; Serie nouă into a farm with birds can open that manual cycle in the same request (`manualCycle`).

**Illustrative data** (never treat as real): guide names, feed type names, "Seria 2025" name and its placement date, users (Dr. Ceban, Ion Rusu, Petrea, Observator exemplu) and every author/time, the culls, the sexing error, the Hîrbovăț weighings (10.11.2026) and the shared correction in `illustrative-events-hirbovat.json`, D39 state examples on the boards.
