-- =====================================================================================================================
-- Impex Line · PostgreSQL schema (target: PostgreSQL 16)
-- Companion of backend/SPEC.md §4 (tables) and §5 (derived calculations). Decision ids (Dnn) refer to
-- design/kit/decisions-current.md; later decisions win over older spec text.
--
-- Principles
--   * Bird counts are NEVER stored as truth (D01, PRODUCT "Counts come from events"): alive = running SUM(bird_event.qty)
--     per section × sex. Views below derive every count; snapshots, if added later, are caches (SPEC §4.13).
--   * One form save = one `entry` row (D168, D172, D178); its figures live in child tables that point back to it.
--     Corrections act on the whole entry; every change is in `audit_log`.
--   * Soft delete: entry.deleted_at, mirrored on its child rows in the same transaction (partial unique indexes and views
--     only see live rows). Nothing a user entered is ever hard-deleted, except unused setup rows (guide, feed type)
--     that the screens let the owner delete (D114, D116).
--   * Empty is null, never 0 (spec C1): a 0 row means "entered 0" (e.g. "Fără pierderi azi"), no row means "not entered".
--   * Dates are local days in Europe/Chisinau (SPEC §11.1). Timestamps are timestamptz.
-- Sanity-checked with psql 16 (schema loads clean, views return the golden values of SPEC §11.7 on the Hîrbovăț extract).
-- =====================================================================================================================

begin;

create extension if not exists btree_gist;   -- exclusion constraint on cycle date ranges

-- ---------------------------------------------------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------------------------------------------------
create type farm_type        as enum ('growth', 'laying');                         -- Creștere / Ouat
create type user_role        as enum ('vet', 'viewer', 'manager', 'owner');        -- Veterinar / Observator / Manager / Proprietar
create type sex              as enum ('F', 'M');                                   -- ♀ găini / ♂ cocoși
create type bird_event_type  as enum ('placement', 'mortality', 'cull', 'sexing_error', 'transfer', 'sale', 'adjustment');
create type cull_reason      as enum ('boala', 'schiopatare', 'altul');            -- UI: boală · șchiopătare · altul (D36)
create type entry_kind       as enum ('daily',        -- Zilnic: mortality, culls, sexing errors (bird_event) + eggs (egg_production)
                                      'feed_daily',   -- Hrană zilnică (feed_allocation)
                                      'weighing',     -- Cântărire
                                      'transfer',     -- Transfer (2 bird_event rows per sex)
                                      'sale',         -- Vânzare (bird_event rows, one per section × sex)
                                      'delivery',     -- Livrare hrană (feed_delivery)
                                      'placement',    -- Serie nouă placement rows (bird_event)
                                      'migration');   -- one-time history import (SPEC §10.3); never created by the app
create type record_source    as enum ('manual', 'voice', 'migration');
create type cycle_start_kind as enum ('auto', 'manual', 'migration');
create type cycle_end_kind   as enum ('emptied', 'manual', 'migration');
create type series_status    as enum ('active', 'closed');
create type audit_action     as enum ('create', 'update', 'delete', 'undo', 'restore',
                                      'price_set', 'price_change',
                                      'cycle_start', 'cycle_guide_change', 'cycle_auto_open', 'cycle_auto_close', 'cycle_reopen',
                                      'series_create', 'series_age_change', 'series_close',
                                      'setup_change',            -- farm, section, guide, feed type
                                      'user_change', 'password_reset', 'user_deactivate', 'user_reactivate');

-- ---------------------------------------------------------------------------------------------------------------------
-- Helpers
-- ---------------------------------------------------------------------------------------------------------------------
-- The business day (SPEC §11.1). All "today", "same day until 23:59" and "missing from 12:00" rules use this clock.
create or replace function app_today() returns date
  language sql stable as $$ select (now() at time zone 'Europe/Chisinau')::date $$;

create or replace function app_now_local() returns timestamp
  language sql stable as $$ select now() at time zone 'Europe/Chisinau' $$;

-- ---------------------------------------------------------------------------------------------------------------------
-- Users, access, sessions (spec A1, A2; D78, D118, D119)
-- ---------------------------------------------------------------------------------------------------------------------
create table app_user (
  id                    uuid primary key default gen_random_uuid(),
  name                  text not null check (length(btrim(name)) between 1 and 60),
  login                 text not null check (login ~ '^[a-z0-9.]{3,40}$'),          -- A2; read-only after creation (D118)
  password_hash         text not null,                                              -- argon2id (SPEC §11.4); the plain password is shown once (D78)
  role                  user_role not null,
  active                boolean not null default true,
  password_generated_at timestamptz not null default now(),                         -- last create/reset (one-time display)
  password_generated_by uuid references app_user(id),
  last_login_at         timestamptz,                                                -- S22 "Ultima intrare"
  created_at            timestamptz not null default now(),
  created_by            uuid references app_user(id),
  updated_at            timestamptz not null default now(),
  constraint app_user_login_uq unique (login)                                       -- deactivated accounts keep their login (D118)
);
comment on table app_user is 'Users are never deleted (A2): their entries keep pointing to them. Rule "at least one active owner" is enforced in the service layer (SPEC §6.10).';

create table user_farm (
  user_id uuid not null references app_user(id) on delete cascade,
  farm_id uuid not null,                                                            -- FK added after farm
  primary key (user_id, farm_id)
);
comment on table user_farm is 'Farm scope for vet/viewer/manager (D03). Owners have no rows: they see every farm, including later ones.';

create table user_session (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references app_user(id),
  token_hash    bytea not null unique,                                              -- sha256 of the 32-byte random cookie token
  created_at    timestamptz not null default now(),
  last_seen_at  timestamptz not null default now(),
  expires_at    timestamptz not null,                                               -- sliding 30 days (A1)
  revoked_at    timestamptz,                                                        -- logout, reset, deactivation
  user_agent    text,
  ip            inet
);
create index user_session_user_idx on user_session (user_id) where revoked_at is null;

create table login_attempt (
  id        bigserial primary key,
  login     text not null,                                                          -- as typed, lower-cased
  ip        inet,
  at        timestamptz not null default now(),
  success   boolean not null
);
create index login_attempt_login_at_idx on login_attempt (login, at desc);
create index login_attempt_ip_at_idx    on login_attempt (ip, at desc);
comment on table login_attempt is 'A1: 5 failed attempts → sign-in waits 5 minutes ("Prea multe încercări. Încearcă din nou peste 5 minute."). Retention 30 days.';

-- ---------------------------------------------------------------------------------------------------------------------
-- Guides (B2, S20; D151, D157, D180, D184, D185)
-- ---------------------------------------------------------------------------------------------------------------------
create table guide (
  id               uuid primary key default gen_random_uuid(),
  name             text not null check (length(btrim(name)) between 1 and 80),
  name_key         text not null,                                                   -- lower(unaccent(name)), computed by the app (D114)
  type             farm_type not null,
  source_filename  text,                                                            -- last imported file ("Tabel lipit" for paste)
  created_at       timestamptz not null default now(),
  created_by       uuid references app_user(id),
  rows_replaced_at timestamptz,                                                     -- Înlocuiește ghidul (D113): rows replaced in place
  constraint guide_name_uq unique (name_key),
  constraint guide_id_type_uq unique (id, type)                                     -- target of the farm/cycle type-match FKs
);

create table guide_week (
  guide_id          uuid not null references guide(id) on delete cascade,
  week              smallint not null check (week between 1 and 120),              -- the birds' age week SAPT (D151), no offset
  sex               sex not null,
  feed_g            numeric(6,1)  check (feed_g >= 0),                             -- Hrană g/pasăre/zi
  body_weight_g     numeric(7,1)  check (body_weight_g > 0),                       -- weight at the END of this week (week 1 = day 7, D157)
  mortality_cum_pct numeric(5,2)  check (mortality_cum_pct between 0 and 100),     -- cumulative % (laying: from start of lay, D161)
  lay_pct           numeric(5,2)  check (lay_pct between 0 and 100),               -- % ouat (laying ♀; Productivitate col I, D184)
  hatching_pct      numeric(5,2)  check (hatching_pct between 0 and 100),          -- % incubabile of all eggs, not hatchability (D153)
  egg_weight_g      numeric(5,1)  check (egg_weight_g > 0),
  primary key (guide_id, week, sex)                                                 -- "Duplicate week + sex rows are refused" (B2)
);
comment on table guide_week is 'A null cell = "– fără date în ghid" (D29). Rows outside the guide span are absent.';

-- ---------------------------------------------------------------------------------------------------------------------
-- Farms and sections (B1, S19; D75, D110, D111, D152)
-- ---------------------------------------------------------------------------------------------------------------------
create table farm (
  id          uuid primary key default gen_random_uuid(),
  name        text not null check (length(btrim(name)) between 1 and 60),
  name_key    text not null,                                                        -- unique ignoring case and diacritics (D111)
  type        farm_type not null,                                                   -- changeable only while the farm has no open cycle (D110)
  guide_id    uuid,                                                                 -- null = "fără ghid"; applies to the NEXT cycle (B1)
  created_at  timestamptz not null default now(),
  created_by  uuid references app_user(id),
  updated_at  timestamptz not null default now(),
  constraint farm_name_uq unique (name_key),
  constraint farm_id_type_uq unique (id, type),
  constraint farm_guide_type_fk foreign key (guide_id, type) references guide (id, type)  -- guide must have the farm's type (B1)
);
alter table user_farm add constraint user_farm_farm_fk foreign key (farm_id) references farm(id) on delete cascade;

create table section (
  id          uuid primary key default gen_random_uuid(),
  farm_id     uuid not null references farm(id),
  code        text not null check (code ~ '^[A-Za-z0-9][A-Za-z0-9.\-]{0,11}$'),   -- H1, H5-1, H5, H5.1
  parent_id   uuid,                                                                 -- hall of a compartment (one level only, D75)
  sort_order  smallint not null default 0,                                          -- reorder among siblings (D75)
  retired_at  timestamptz,                                                          -- Retrage: never deleted, no reactivation (D75)
  retired_by  uuid references app_user(id),
  created_at  timestamptz not null default now(),
  constraint section_id_farm_uq unique (id, farm_id),
  constraint section_parent_same_farm_fk foreign key (parent_id, farm_id) references section (id, farm_id),
  constraint section_not_own_parent check (parent_id is null or parent_id <> id)
);
create unique index section_code_active_uq on section (farm_id, upper(code)) where retired_at is null;  -- "H5.2 există deja în H5."
create index section_parent_idx on section (parent_id);
comment on table section is 'A section with children is a hall: it holds no birds, carries the hall''s eggs (egg_production) and groups its compartments. A section without children holds birds (bird_event, feed_allocation, weighing). Laying egg unit = the hall, or the section itself when it has no parent and no children (SPEC §4.4).';

-- ---------------------------------------------------------------------------------------------------------------------
-- Series and cycles (B3, B4; D151, D17, D54, D107–D109, D179)
-- ---------------------------------------------------------------------------------------------------------------------
create table series (
  id                     uuid primary key default gen_random_uuid(),
  name                   text not null check (length(btrim(name)) between 1 and 60),
  name_key               text not null,
  year                   smallint not null check (year between 2000 and 2100),    -- "Anul 2026 e deja folosit de Seria 2026."
  status                 series_status not null default 'active',
  placement_date         date not null,                                            -- date of the first placement (age reference, D151)
  placement_farm_id      uuid references farm(id),
  placement_inferred     boolean not null default false,                           -- Seria 2025: inferred 01.11.2024 (D36, DATA03)
  age_at_placement_days  smallint not null default 0 check (age_at_placement_days between 0 and 140),  -- B3 0–140
  closed_at              timestamptz,
  closed_by              uuid references app_user(id),
  created_at             timestamptz not null default now(),
  created_by             uuid references app_user(id),
  constraint series_name_uq unique (name_key),
  constraint series_year_uq unique (year),
  constraint series_closed_ck check ((status = 'closed') = (closed_at is not null))
);
comment on column series.age_at_placement_days is 'age_days(date) = age_at_placement_days + (date − placement_date); SAPT = floor(age_days/7)+1 (D151). A change renumbers every week of the series (owner, Motiv, audit; spec B3).';

create table cycle (
  id             uuid primary key default gen_random_uuid(),
  series_id      uuid not null references series(id),
  farm_id        uuid not null references farm(id),
  guide_id       uuid references guide(id),                                         -- snapshot of farm.guide_id at open; "Schimbă ghidul ciclului" (B4)
  start_date     date not null,
  end_date       date,                                                              -- null = open; inclusive last day of the cycle
  started_by     cycle_start_kind not null,
  start_inferred boolean not null default false,                                    -- "început dedus" (D54/D55)
  ended_by       cycle_end_kind,
  start_reason   text check (start_reason is null or length(start_reason) <= 200),  -- Motiv of Începe ciclu nou
  created_at     timestamptz not null default now(),
  created_by     uuid references app_user(id),                                      -- null for auto/migration
  constraint cycle_dates_ck check (end_date is null or end_date >= start_date),
  constraint cycle_end_kind_ck check ((end_date is null) = (ended_by is null)),
  constraint cycle_manual_reason_ck check (started_by <> 'manual' or start_reason is not null),
  constraint cycle_no_overlap exclude using gist (farm_id with =, daterange(start_date, coalesce(end_date, 'infinity'::date), '[]') with &&)
);
create unique index cycle_one_open_per_farm on cycle (farm_id) where end_date is null;   -- B4 "One open cycle per farm"
create index cycle_series_idx on cycle (series_id);
comment on table cycle is 'One stay of one series in one farm. Cycles of a farm never overlap; a dated record of the farm belongs to the cycle whose [start_date, end_date] contains its date. Auto open/close is recomputed by reconcileCycles (SPEC §6.4) after every write that moves birds.';

-- ---------------------------------------------------------------------------------------------------------------------
-- Entries: one form save = one record (D168, D172, D178)
-- ---------------------------------------------------------------------------------------------------------------------
create table entry (
  id                 uuid primary key default gen_random_uuid(),
  kind               entry_kind not null,
  farm_id            uuid not null references farm(id),                             -- transfer: the source farm
  peer_farm_id       uuid references farm(id),                                      -- transfer: the destination farm (may equal farm_id)
  date               date not null,                                                 -- the record's day (read-only in corrections, D101)
  series_id          uuid references series(id),                                    -- placement entries
  buyer              text check (buyer is null or length(buyer) <= 120),            -- Vânzare (C4)
  note               text check (note is null or length(note) <= 500),
  source             record_source not null default 'manual',                       -- 'voice' when the save still carried a "din voce" mark (D199)
  client_request_id  uuid,                                                          -- idempotency key from the form draft (SPEC §7.0)
  created_by         uuid references app_user(id),                                  -- null only for migration
  created_at         timestamptz not null default now(),
  updated_by         uuid references app_user(id),
  updated_at         timestamptz,
  corrected          boolean not null default false,                                -- true once changed by Corectează (Tag "corectat", D58)
  deleted_at         timestamptz,
  deleted_by         uuid references app_user(id),
  delete_reason      text check (delete_reason is null or length(delete_reason) <= 200),
  constraint entry_peer_ck check ((kind = 'transfer') = (peer_farm_id is not null)),
  constraint entry_buyer_ck check (buyer is null or kind = 'sale'),
  constraint entry_series_ck check ((kind = 'placement') = (series_id is not null)),
  constraint entry_deleted_ck check ((deleted_at is null) = (deleted_by is null))
);
create unique index entry_client_request_uq on entry (client_request_id) where client_request_id is not null;
create index entry_farm_date_idx   on entry (farm_id, date desc) where deleted_at is null;
create index entry_peer_date_idx   on entry (peer_farm_id, date desc) where deleted_at is null and peer_farm_id is not null;
create index entry_author_day_idx  on entry (created_by, created_at desc);
comment on column entry.date is 'Local day (Europe/Chisinau). Vet edits allowed while app_today() = date and created_by = the vet (D07).';

-- ---------------------------------------------------------------------------------------------------------------------
-- The bird ledger (D01; data-model BIRD_EVENT)
-- ---------------------------------------------------------------------------------------------------------------------
create table bird_event (
  id               uuid primary key default gen_random_uuid(),
  entry_id         uuid not null references entry(id),
  farm_id          uuid not null,
  section_id       uuid not null,
  date             date not null,
  sex              sex not null,
  type             bird_event_type not null,
  qty              integer not null,                                                -- signed delta to birds alive
  cull_reason      cull_reason,
  peer_section_id  uuid references section(id),                                     -- transfer: the other side
  note             text check (note is null or length(note) <= 200),                -- e.g. adjustment "diferență registru"
  deleted_at       timestamptz,                                                     -- mirrors entry.deleted_at / superseded by an edit
  constraint bird_event_section_fk foreign key (section_id, farm_id) references section (id, farm_id),
  constraint bird_event_sign_ck check (
       (type = 'placement'    and qty >= 0)
    or (type in ('mortality', 'cull', 'sale') and qty <= 0)
    or (type in ('sexing_error', 'transfer', 'adjustment'))),
  constraint bird_event_cull_reason_ck check ((type = 'cull' and (qty = 0 or cull_reason is not null)) or (type <> 'cull' and cull_reason is null)),
  constraint bird_event_peer_ck check (type = 'transfer' or peer_section_id is null)  -- app: required on every non-migration transfer row
);
-- Zilnic values: one live value per section × day × sex × kind (spec C1 "a second save replaces the earlier entry")
create unique index bird_event_daily_uq on bird_event (section_id, date, sex, type)
  where deleted_at is null and type in ('mortality', 'cull', 'sexing_error');
create index bird_event_section_sex_date_idx on bird_event (section_id, sex, date) where deleted_at is null;
create index bird_event_farm_date_idx        on bird_event (farm_id, date) where deleted_at is null;
create index bird_event_entry_idx            on bird_event (entry_id);
comment on column bird_event.qty is 'Signed: placement +; mortality, cull, sale −; transfer − out / + in (two rows per sex, peer_section_id = the other side); sexing_error as recorded (± per sex, spec D1 Păsări); adjustment ± (migration only: "Diferență registru", D63). 0 = entered as zero.';

-- ---------------------------------------------------------------------------------------------------------------------
-- Eggs (C1 Ouă; D18, D183)
-- ---------------------------------------------------------------------------------------------------------------------
create table egg_production (
  id               uuid primary key default gen_random_uuid(),
  entry_id         uuid not null references entry(id),
  farm_id          uuid not null,
  section_id       uuid not null,                                                   -- the egg unit: the hall (parent) or a parentless section
  date             date not null,
  floor_eggs       integer check (floor_eggs >= 0),                                 -- Jos (kept apart from the total, D18)
  cat1             integer check (cat1 >= 0),
  cat2             integer check (cat2 >= 0),
  cat3             integer check (cat3 >= 0),
  hatching_eggs    integer check (hatching_eggs >= 0),                              -- Incubabile
  avg_egg_weight_g numeric(5,2) check (avg_egg_weight_g between 20 and 120),        -- Masa medie
  total_eggs       integer generated always as (cat1 + cat2 + cat3 + hatching_eggs) stored,  -- null until all four are filled (D53)
  deleted_at       timestamptz,
  constraint egg_section_fk foreign key (section_id, farm_id) references section (id, farm_id)
);
create unique index egg_production_uq on egg_production (section_id, date) where deleted_at is null;
create index egg_production_farm_date_idx on egg_production (farm_id, date) where deleted_at is null;

-- ---------------------------------------------------------------------------------------------------------------------
-- Feed (C2, C5, C6, S21; D05, D06, D13, D40, D116, D117)
-- ---------------------------------------------------------------------------------------------------------------------
create table feed_type (
  id          uuid primary key default gen_random_uuid(),
  name        text not null check (length(btrim(name)) between 1 and 40),
  name_key    text not null,
  for_sex     sex,                                                                  -- null = ambele sexe; fixed once used (D116)
  active      boolean not null default true,                                        -- retire refused while any farm has it in stock (D116)
  created_at  timestamptz not null default now(),
  created_by  uuid references app_user(id),
  constraint feed_type_name_uq unique (name_key)
);
comment on table feed_type is 'Global list (no farm, D116). A used type (any delivery or allocation) is never deleted, only retired.';

create table feed_delivery (
  id            uuid primary key default gen_random_uuid(),
  entry_id      uuid not null references entry(id),
  farm_id       uuid not null references farm(id),
  feed_type_id  uuid not null references feed_type(id),
  date          date not null,
  qty_kg        numeric(12,1) not null check (qty_kg > 0),
  note          text check (note is null or length(note) <= 200),
  price_per_kg  numeric(12,4) check (price_per_kg > 0),                             -- lei/kg (owner only, D05)
  total_lei     numeric(14,2) check (total_lei > 0),                                -- invoice total; the other value is derived (C6)
  price_input   text check (price_input in ('per_kg', 'total')),                      -- which one the owner typed
  priced_by     uuid references app_user(id),
  priced_at     timestamptz,
  price_from_register boolean not null default false,                              -- migrated price: "din registru" (D163)
  deleted_at    timestamptz,
  constraint feed_delivery_price_ck check (
    (price_per_kg is null and total_lei is null and priced_at is null and price_input is null)
    or (price_per_kg is not null and total_lei is not null and (priced_at is not null or price_from_register)))
);
create index feed_delivery_farm_type_date_idx on feed_delivery (farm_id, feed_type_id, date) where deleted_at is null;
create index feed_delivery_unpriced_idx on feed_delivery (date) where deleted_at is null and total_lei is null;

create table feed_stock_adjustment (                                                -- migration only (workbook stock moves, SPEC §10.3)
  id            uuid primary key default gen_random_uuid(),
  entry_id      uuid not null references entry(id),
  farm_id       uuid not null references farm(id),
  feed_type_id  uuid not null references feed_type(id),
  date          date not null,
  qty_kg        numeric(12,1) not null check (qty_kg <> 0),                         -- signed
  note          text not null
);

create table feed_allocation (
  id            uuid primary key default gen_random_uuid(),
  entry_id      uuid not null references entry(id),
  farm_id       uuid not null,
  section_id    uuid not null,
  date          date not null,
  sex           sex not null,
  feed_type_id  uuid not null references feed_type(id),
  g_per_bird    numeric(6,1) check (g_per_bird between 1 and 300),                  -- C2 "g: decimal from 1 to 300"; app rows always set it
  kg_recorded   numeric(10,1) check (kg_recorded >= 0),                             -- migration only: the register's typed kg (D163, open q. 12)
  deleted_at    timestamptz,
  constraint feed_allocation_value_ck check (g_per_bird is not null or kg_recorded is not null),  -- early Telița days: kg without g
  constraint feed_allocation_section_fk foreign key (section_id, farm_id) references section (id, farm_id)
);
create unique index feed_allocation_uq on feed_allocation (section_id, date, sex) where deleted_at is null;
create index feed_allocation_farm_date_idx on feed_allocation (farm_id, date) where deleted_at is null;
comment on column feed_allocation.kg_recorded is 'null for app data: kg = g_per_bird × birds alive at the START of the day / 1000 (spec C2 sample 14 395 × 160 g = 2 303,2 kg; DATA03).';

-- ---------------------------------------------------------------------------------------------------------------------
-- Weighings (C7; D95–D97, D157)
-- ---------------------------------------------------------------------------------------------------------------------
create table weighing (
  id              uuid primary key default gen_random_uuid(),
  entry_id        uuid not null references entry(id),
  farm_id         uuid not null,
  section_id      uuid not null,
  date            date not null,
  sex             sex not null,
  avg_weight_g    numeric(7,1) not null check (avg_weight_g between 1 and 10000),
  cv_pct          numeric(5,2) check (cv_pct between 0 and 100),
  uniformity_pct  numeric(5,2) check (uniformity_pct between 0 and 100),
  sample_size     integer check (sample_size between 1 and 100000),
  deleted_at      timestamptz,
  constraint weighing_section_fk foreign key (section_id, farm_id) references section (id, farm_id)
);
create unique index weighing_uq on weighing (section_id, date, sex) where deleted_at is null;
create index weighing_farm_date_idx on weighing (farm_id, date) where deleted_at is null;

-- Past cycles without a flock count (Seria 2025 at Hîrbovăț, D52/D54): deaths per compartment and day, display only.
-- They cannot be bird_events: with no placement/arrival the running count would go below 0.
create table legacy_mortality (
  id          uuid primary key default gen_random_uuid(),
  entry_id    uuid not null references entry(id),                                   -- the migration entry
  farm_id     uuid not null,
  section_id  uuid not null,
  date        date not null,
  sex         sex not null,
  deaths      integer not null check (deaths >= 0),
  emptying    boolean not null default false,                                       -- removals on emptying days: never shown as mortality (D52, D54, D92)
  constraint legacy_mortality_section_fk foreign key (section_id, farm_id) references section (id, farm_id),
  constraint legacy_mortality_uq unique (section_id, date, sex)
);

-- ---------------------------------------------------------------------------------------------------------------------
-- Audit (E1, E2; D07, D70, D104, D105, D172, D178, D199 note)
-- ---------------------------------------------------------------------------------------------------------------------
create table audit_log (
  id                bigserial primary key,
  at                timestamptz not null default now(),
  user_id           uuid references app_user(id),                                   -- null = system (auto cycle) or migration
  action            audit_action not null,
  entity            text not null check (entity in ('entry', 'feed_delivery', 'cycle', 'series', 'farm', 'section',
                                                    'guide', 'feed_type', 'app_user')),
  entity_id         uuid not null,
  entry_id          uuid references entry(id),
  entry_kind        entry_kind,
  record_types      text[] not null default '{}',                                   -- S17 chips: mortalitate, sacrificare, eroare_sexare, transfer, populare, vanzare, hrana, livrare, pret, cantarire, oua, ciclu, setari
  farm_ids          uuid[] not null default '{}',                                   -- farm scope (manager sees rows touching his farms)
  record_date       date,                                                           -- the day of the changed record ("ziua 17.11.2026")
  reason            text check (reason is null or length(reason) <= 200),           -- Motiv; null only for same-day author edits (D70 A) and system rows
  same_day_author   boolean not null default false,                                 -- "editat"/"șters" by the author before 23:59, no reason (D70 A)
  summary           text,                                                           -- "Mortalitate · 4 secțiuni" (D178)
  changes           jsonb not null default '[]'::jsonb,                             -- [{key:"H5.2 · găini moarte", section:"H5.2", sex:"F", field:"mortality", old:1, new:3}]
  old_values        jsonb,                                                          -- full snapshot of the save before
  new_values        jsonb,                                                          -- full snapshot after (null on delete)
  source            record_source not null default 'manual',
  voice_transcript  text                                                            -- only when the save still carried a "din voce" mark (D199)
);
create index audit_log_at_idx       on audit_log (at desc);
create index audit_log_user_at_idx  on audit_log (user_id, at desc);
create index audit_log_entity_idx   on audit_log (entity, entity_id, at desc);
create index audit_log_entry_idx    on audit_log (entry_id);
create index audit_log_farms_idx    on audit_log using gin (farm_ids);
comment on table audit_log is 'Append-only (no update/delete grants for the app role). One row per save action (D178).';

-- ---------------------------------------------------------------------------------------------------------------------
-- Voice usage (spec §8 Privacy: "A usage count per user and day (no text) is kept for cost.")
-- ---------------------------------------------------------------------------------------------------------------------
create table voice_usage (
  user_id         uuid not null references app_user(id),
  day             date not null,
  commands        integer not null default 0,
  audio_seconds   numeric(8,1) not null default 0,
  llm_tokens_in   integer not null default 0,
  llm_tokens_out  integer not null default 0,
  primary key (user_id, day)
);

-- ---------------------------------------------------------------------------------------------------------------------
-- Live update outbox (D197; SPEC §9): short-retention event log for SSE replay after reconnect
-- ---------------------------------------------------------------------------------------------------------------------
create table live_event (
  id        bigserial primary key,
  at        timestamptz not null default now(),
  farm_ids  uuid[] not null,                                                        -- empty = global (setup, users)
  type      text not null,
  payload   jsonb not null
);
create index live_event_at_idx on live_event (at);
comment on table live_event is 'Written in the same transaction as the change; NOTIFY live_event after commit. Rows older than 24 h are purged.';

-- ---------------------------------------------------------------------------------------------------------------------
-- Settings (D19 thresholds; editable only by a developer/SQL for now)
-- ---------------------------------------------------------------------------------------------------------------------
create table app_setting (
  key    text primary key,
  value  jsonb not null,
  note   text
);
insert into app_setting (key, value, note) values
  ('tolerance.weight_pct',      '5',   'Weighing vs guide (SAPT − 1): |Δ| > 5 % → abatere (D19, D157)'),
  ('tolerance.lay_pts',         '3',   'Lay % vs guide: |Δ| > 3 points → abatere (D19)'),
  ('tolerance.feed_pct',        '10',  'g/pasăre vs guide: |Δ| > 10 % → abatere (D19, C2)'),
  ('alert.stock_days_min',      '5',   'Stock days left < 5 → abatere (D19)'),
  ('alert.stock_avg_days',      '3',   'Average consumption window for days left (data-model)'),
  ('alert.mortality_factor',    '3',   'Zilnic warning: deaths > 3 × the section''s 7-day mean for that sex (D19, D53)'),
  ('missing.from_hour',         '12',  'Lipsesc date shows today''s gaps from 12:00 (C9)'),
  ('missing.days_back',         '7',   'Lipsesc date window: today + 7 days back (C9)'),
  ('session.days',              '30',  'Sliding session length (A1)'),
  ('login.max_failures',        '5',   'A1'),
  ('login.lock_minutes',        '5',   'A1'),
  ('voice.daily_cap_per_user',  '200', 'Assumption (open q. 14): commands per user per day'),
  ('series.max_age_at_placement','140','B3');

-- =====================================================================================================================
-- Derived views (SPEC §5). Counts are computed, never stored. All views read live rows only.
-- =====================================================================================================================

-- Age and week (D151)
create or replace function age_days(p_series uuid, p_date date) returns integer
  language sql stable as $$
    select s.age_at_placement_days + (p_date - s.placement_date) from series s where s.id = p_series $$;
create or replace function sapt(p_series uuid, p_date date) returns integer
  language sql stable as $$ select floor(age_days(p_series, p_date) / 7.0)::int + 1 $$;
comment on function sapt(uuid, date) is 'SAPT = floor(age_days / 7) + 1: 13 days → 2, 14 days → 3; Seria 2026 18.11.2026 → age 407, SAPT 59.';

-- The cycle a farm-day belongs to
create or replace view v_cycle as
select c.*, s.name as series_name, s.placement_date, s.age_at_placement_days,
       age_days(c.series_id, c.start_date)                     as age_days_start,
       sapt(c.series_id, c.start_date)                         as sapt_start,
       sapt(c.series_id, coalesce(c.end_date, app_today()))    as sapt_end_or_today,
       (c.end_date is null)                                    as is_open
from cycle c join series s on s.id = c.series_id;

-- Live bird events with display counts
create or replace view v_bird_event as
select e.*, n.kind as entry_kind, n.created_by, n.created_at,
       case when e.type in ('mortality', 'cull', 'sale') then -e.qty end as count_removed
from bird_event e
join entry n on n.id = e.entry_id
where e.deleted_at is null and n.deleted_at is null;

-- Per section × sex × day deltas (only days with events)
create or replace view v_section_sex_delta as
select section_id, farm_id, sex, date,
       coalesce(sum(qty) filter (where type = 'placement'), 0)                as placed,
       coalesce(-sum(qty) filter (where type = 'mortality'), 0)               as deaths,
       coalesce(-sum(qty) filter (where type = 'cull'), 0)                    as culls,
       coalesce(sum(qty) filter (where type = 'sexing_error'), 0)             as sexing_error,
       coalesce(sum(qty) filter (where type = 'transfer' and qty > 0), 0)     as transfer_in,
       coalesce(-sum(qty) filter (where type = 'transfer' and qty < 0), 0)    as transfer_out,
       coalesce(-sum(qty) filter (where type = 'sale'), 0)                    as sold,
       coalesce(sum(qty) filter (where type = 'adjustment'), 0)               as adjustment,
       sum(qty)                                                               as delta,
       bool_or(type = 'mortality')                                            as mortality_entered
from v_bird_event
group by section_id, farm_id, sex, date;

-- Dense daily alive per section × sex: from the first event to max(last event, today). alive_start = end of the day before.
create or replace view v_section_sex_day as
with b as (
  select section_id, farm_id, sex, min(date) as d0, greatest(max(date), app_today()) as d1
  from v_section_sex_delta group by section_id, farm_id, sex
), days as (
  select b.section_id, b.farm_id, b.sex, g::date as date
  from b, generate_series(b.d0, b.d1, interval '1 day') g
)
select d.section_id, d.farm_id, d.sex, d.date,
       coalesce(x.placed, 0) as placed, coalesce(x.deaths, 0) as deaths, coalesce(x.culls, 0) as culls,
       coalesce(x.sexing_error, 0) as sexing_error, coalesce(x.transfer_in, 0) as transfer_in,
       coalesce(x.transfer_out, 0) as transfer_out, coalesce(x.sold, 0) as sold, coalesce(x.adjustment, 0) as adjustment,
       coalesce(x.mortality_entered, false) as mortality_entered,
       sum(coalesce(x.delta, 0)) over w - coalesce(x.delta, 0) as alive_start,
       sum(coalesce(x.delta, 0)) over w                         as alive_end
from days d
left join v_section_sex_delta x using (section_id, farm_id, sex, date)
window w as (partition by d.section_id, d.sex order by d.date rows between unbounded preceding and current row);
comment on view v_section_sex_day is 'Invariant (PRODUCT, spec §1): alive_end >= 0 on every row. Writes are refused when they would break it (SPEC §6.1).';

-- Farm × sex × day totals (= Σ sections, D156: never a register Total row)
create or replace view v_farm_sex_day as
select farm_id, sex, date,
       sum(alive_start) as alive_start, sum(alive_end) as alive_end,
       sum(placed) as placed, sum(deaths) as deaths, sum(culls) as culls, sum(sexing_error) as sexing_error,
       sum(transfer_in) as transfer_in, sum(transfer_out) as transfer_out, sum(sold) as sold, sum(adjustment) as adjustment
from v_section_sex_day
group by farm_id, sex, date;

-- Egg unit per section: the hall (parent) or the section itself
create or replace view v_egg_unit as
select s.id as section_id, s.farm_id, coalesce(s.parent_id, s.id) as egg_unit_id
from section s
where not exists (select 1 from section c where c.parent_id = s.id);

-- Eggs per unit and day with hens at the start of the day (D156) and the D18/D183 parts
create or replace view v_egg_day as
select ep.farm_id, ep.section_id as egg_unit_id, ep.date,
       ep.total_eggs, ep.hatching_eggs, ep.floor_eggs, ep.cat1, ep.cat2, ep.cat3, ep.avg_egg_weight_g,
       h.hens_start,
       case when h.hens_start > 0 and ep.total_eggs is not null then 100.0 * ep.total_eggs / h.hens_start end as lay_pct
from egg_production ep
join entry n on n.id = ep.entry_id and n.deleted_at is null
left join lateral (
  select sum(d.alive_start) as hens_start
  from v_egg_unit u join v_section_sex_day d on d.section_id = u.section_id and d.sex = 'F' and d.date = ep.date
  where u.egg_unit_id = ep.section_id
) h on true
where ep.deleted_at is null;

-- Daily feed consumption (kg) per allocation: g × birds at the start of the day (DATA03), or the migrated typed kg
create or replace view v_feed_consumption as
select fa.id, fa.entry_id, fa.farm_id, fa.section_id, fa.date, fa.sex, fa.feed_type_id, fa.g_per_bird,
       coalesce(d.alive_start, 0) as birds_start,
       coalesce(fa.kg_recorded, fa.g_per_bird * coalesce(d.alive_start, 0) / 1000.0) as kg   -- exact; round only for display
from feed_allocation fa
join entry n on n.id = fa.entry_id and n.deleted_at is null
left join v_section_sex_day d on d.section_id = fa.section_id and d.sex = fa.sex and d.date = fa.date
where fa.deleted_at is null;

-- Stock movements per farm × feed type × day, and the running stock at the end of each movement day
create or replace view v_feed_stock as
with m as (
  select fd.farm_id, fd.feed_type_id, fd.date, fd.qty_kg as delivered, 0::numeric as consumed, 0::numeric as adjusted
  from feed_delivery fd join entry n on n.id = fd.entry_id and n.deleted_at is null where fd.deleted_at is null
  union all
  select farm_id, feed_type_id, date, 0, kg, 0 from v_feed_consumption
  union all
  select a.farm_id, a.feed_type_id, a.date, 0, 0, a.qty_kg from feed_stock_adjustment a
), d as (
  select farm_id, feed_type_id, date, sum(delivered) as delivered, sum(consumed) as consumed, sum(adjusted) as adjusted
  from m group by farm_id, feed_type_id, date
)
select d.*, sum(delivered - consumed + adjusted) over (partition by farm_id, feed_type_id order by date) as stock_end
from d;
comment on view v_feed_stock is 'stock_end(date) = Σ deliveries − Σ consumption + Σ adjustments up to that day. Hîrbovăț 18.11.2026: Breeder 2 31 664,9 kg · Cocoși 1 781 kg.';

-- Section codes with their hall (labels for tables, audit keys)
create or replace view v_section_label as
select s.id, s.farm_id, s.code, s.parent_id, p.code as hall_code,
       exists (select 1 from section c where c.parent_id = s.id) as is_hall,
       s.retired_at is not null as retired
from section s left join section p on p.id = s.parent_id;

commit;
