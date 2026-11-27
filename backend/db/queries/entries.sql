-- name: InsertEntry :one
-- The header of one form save (§4.7, §12.5). created_at comes from the injected clock (M3 D3), so "today = local
-- date of created_at" (§3.2) holds under a frozen clock; the version is read back from the row, never formatted
-- from the Go clock. client_request_id is null for a dry run and for a save without a key.
insert into entry (kind, farm_id, peer_farm_id, date, series_id, buyer, note, source, client_request_id, created_by,
                   created_at)
values (sqlc.arg(kind), sqlc.arg(farm_id), sqlc.narg(peer_farm_id), sqlc.arg(date), sqlc.narg(series_id),
        sqlc.narg(buyer), sqlc.narg(note), sqlc.arg(source), sqlc.narg(client_request_id), sqlc.narg(created_by),
        sqlc.arg(created_at))
returning id, coalesce(updated_at, created_at)::timestamptz as version_at;

-- name: EntriesFarmSections :many
-- The sections of a farm (retired ones and halls too) with their codes, for the level-1 checks and messages of a
-- create (§7.0b: "sectionId not a section of this farm", "H1.1 ♀: …"). One row with a null section when the farm has
-- none; no row when the farm doesn't exist.
select f.id as farm_id, s.id as section_id, s.code
from farm f
left join section s on s.farm_id = f.id
where f.id = sqlc.arg(farm_id)::uuid
order by s.code;

-- name: EntriesKnownFeedTypes :many
-- The asked feed type ids that exist (active or not): an unknown id is a level-1 issue, an inactive one is H2.
select ft.id
from feed_type ft
where ft.id = any(sqlc.arg(feed_type_ids)::uuid[]);

-- name: EntriesFeedSaveContext :many
-- The farm context of a Hrană zilnică save (§6.3, M3 D5/D6), read under the farm lock: the farm's name, the cycle
-- whose [start_date, end_date] contains the day with its series (closed → 409 SERIES_CLOSED, §3.1), the guide's
-- feed_g per sex for SAPT(day) (null: no guide, no guide row or an empty cell, so no H4), tolerance.feed_pct, and one
-- row per asked feed type that exists (name, for_sex, active: H2), ordered by name_key. The context columns repeat on
-- every row; a farm without any asked type gives one row with null type columns; no row = no such farm.
select f.name as farm_name,
       s.name as series_name,
       coalesce(s.status = 'closed', false)::boolean as series_closed,
       gf.feed_g as guide_feed_g_f,
       gm.feed_g as guide_feed_g_m,
       (select a.value #>> '{}' from app_setting a where a.key = 'tolerance.feed_pct')::numeric as tolerance_feed_pct,
       ft.id as feed_type_id,
       ft.name as feed_type_name,
       ft.for_sex as feed_type_for_sex,
       ft.active as feed_type_active
from farm f
left join cycle cy on cy.farm_id = f.id
     and cy.start_date <= sqlc.arg(day)::date
     and (cy.end_date is null or cy.end_date >= sqlc.arg(day)::date)
left join series s on s.id = cy.series_id
left join guide_week gf on gf.guide_id = cy.guide_id and gf.sex = 'F' and gf.week = sapt(cy.series_id, sqlc.arg(day)::date)
left join guide_week gm on gm.guide_id = cy.guide_id and gm.sex = 'M' and gm.week = sapt(cy.series_id, sqlc.arg(day)::date)
left join feed_type ft on ft.id = any(sqlc.arg(feed_type_ids)::uuid[])
where f.id = sqlc.arg(farm_id)::uuid
order by ft.name_key, ft.id;

-- name: EntriesFeedDaySlots :many
-- Every section × sex of a farm on one day for a Hrană zilnică save (§6.3, M3 D5): whether the section is a hall or
-- retired, the birds alive at the start of the day (Σ deltas before it: a running sum over v_section_sex_delta, no
-- dense view, ARC-6), and the allocation holding the (section, day, sex) key with its entry and that entry's
-- creation time, author and source (the feed_allocation_uq rule: a row not soft-deleted). Read under the farm lock: H3, ALREADY_ENTERED
-- and the "{n} din {m} rânduri" count; read again on the pool to name the holder of a feed_allocation_uq violation.
with alive as (
  select x.section_id, x.sex, sum(x.delta)::integer as alive_start
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date < sqlc.arg(day)::date
  group by x.section_id, x.sex
)
select s.id as section_id, s.code, sx.sex::sex as sex,
       exists (select 1 from section c where c.parent_id = s.id)::boolean as is_hall,
       (s.retired_at is not null)::boolean as retired,
       coalesce(a.alive_start, 0)::integer as alive_start,
       fa.entry_id as holder_entry_id,
       n.created_at as holder_created_at,
       n.created_by as holder_created_by,
       n.source as holder_source
from section s
cross join (values ('F'::sex), ('M'::sex)) as sx(sex)
left join alive a on a.section_id = s.id and a.sex = sx.sex
left join feed_allocation fa on fa.section_id = s.id and fa.sex = sx.sex and fa.date = sqlc.arg(day)::date
     and fa.deleted_at is null
left join entry n on n.id = fa.entry_id
where s.farm_id = sqlc.arg(farm_id)::uuid
order by s.id, sx.sex;

-- name: EntriesInsertFeedAllocation :exec
-- One Hrană zilnică row (§4.8): kg is never stored for app data, it is g × birds at the start of the day (DATA03).
insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
values (sqlc.arg(entry_id), sqlc.arg(farm_id), sqlc.arg(section_id), sqlc.arg(date), sqlc.arg(sex),
        sqlc.arg(feed_type_id), sqlc.arg(g_per_bird));
