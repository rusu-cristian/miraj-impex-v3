-- Entry changes and the entry read (SPEC §7.8 "Read, edit, correct, delete, undo, restore", plan M5-D2, D3, D7, D10).
-- Owned by entries.Runner.Change (change.go), the kind adapters (kinds.go) and GET /api/entries/{id} (entryread.go).

-- name: EntryChangeHeader :one
-- An entry's kind, farms, date and whether it is deleted, read without a lock (plan M5-D2, M5 ARC-4): by the change
-- runner before its farm locks (farm_id and peer_farm_id never change, so the locks taken from this read are the
-- right ones; everything else is re-read under the locks, EntryChangeLock) and by PUT before it decodes `values`
-- (§7.8: the kind picks the create type, the farm and date are what `values` must repeat; kind and date never
-- change either). No row = no such entry.
select e.kind, e.farm_id, e.peer_farm_id, e.date, (e.deleted_at is not null)::boolean as deleted
from entry e
where e.id = sqlc.arg(id)::uuid;

-- name: EntryChangeLock :one
-- The entry row of a change, locked (select … for update) after the farm advisory locks (plan M5-D2): its version
-- (coalesce(updated_at, created_at), §12.5), author and creation instant (§3.2 windows), the corrected flag, whether
-- it is deleted, and the closed series whose cycle holds its date at either farm (closed_series_on, 0003; '' when
-- none: §3.1 → 409 SERIES_CLOSED).
select e.id, e.kind, e.farm_id, e.peer_farm_id, e.date, e.source, e.created_by, e.created_at,
       coalesce(e.updated_at, e.created_at)::timestamptz as version_at,
       e.corrected,
       (e.deleted_at is not null)::boolean as deleted,
       coalesce(closed_series_on(e.farm_id, e.date), closed_series_on(e.peer_farm_id, e.date), '')::text as closed_series
from entry e
where e.id = sqlc.arg(id)::uuid
for update of e;

-- name: UpdateEntryHeader :one
-- Moves an entry's version (plan M5-D3): updated_at = the injected clock's instant, but always at least 1 µs after
-- the current version, so two changes at the same frozen instant still give two versions and ifVersion keeps
-- guarding. The row changes only while its version is if_version: no row back = the version moved (409 STALE).
-- updated_by = the caller; corrected = the flag after the change (true after a Corectează, back to the old flag after
-- an undone update).
update entry
set updated_at = greatest(sqlc.arg(now)::timestamptz, coalesce(updated_at, created_at) + interval '1 microsecond'),
    updated_by = sqlc.arg(updated_by)::uuid,
    corrected  = sqlc.arg(corrected)::boolean
where id = sqlc.arg(id)::uuid
  and coalesce(updated_at, created_at) = sqlc.arg(if_version)::timestamptz
returning updated_at::timestamptz as version_at;

-- name: EntryLatestAudit :one
-- The latest audit row of an entry (ties on `at`, as under a frozen clock, go to the later row): undo and restore
-- act only on the caller's own latest action of the same day (M5 Q8), which the caller's current rights must still
-- allow (same_day_author: an author's action needs the author's window, else a correction right; M5 SEC-2).
select a.id, a.action, a.user_id, a.at, a.reason, a.same_day_author, a.changes, a.old_values, a.new_values
from audit_log a
where a.entry_id = sqlc.arg(entry_id)::uuid
order by a.at desc, a.id desc
limit 1;

-- name: EntryReadHead :one
-- The head of GET /api/entries/{id} (M5 Q6): the entry with its farm's name, the age parameters of the series whose
-- cycle holds the entry's date (null without a cycle; SAPT and age are calc's, §5.1), the author's name, the latest
-- correction's author and instant (lc below, M5 COR-1: null when none), the closed series of its date
-- (as EntryChangeLock) and its latest audit row as undo and restore read it (EntryLatestAudit: action, user, instant,
-- same_day_author, old_values; null without any audit row, M5 Q8). No row = no such entry.
select e.id, e.kind, e.farm_id, e.peer_farm_id, f.name as farm_name, e.date, e.source, e.created_by, e.created_at,
       coalesce(e.updated_at, e.created_at)::timestamptz as version_at,
       e.corrected,
       (e.deleted_at is not null)::boolean as deleted,
       au.name as author,
       cu.name as corrected_by_name,
       ca.at as corrected_at,
       s.placement_date, s.age_at_placement_days,
       coalesce(closed_series_on(e.farm_id, e.date), closed_series_on(e.peer_farm_id, e.date), '')::text as closed_series,
       la.action as latest_action,
       la.user_id as latest_user_id,
       la.at as latest_at,
       la.same_day_author as latest_same_day_author,
       la.old_values as latest_old_values
from entry e
join farm f on f.id = e.farm_id
left join cycle cy on cy.farm_id = e.farm_id
     and cy.start_date <= e.date
     and (cy.end_date is null or cy.end_date >= e.date)
left join series s on s.id = cy.series_id
left join app_user au on au.id = e.created_by
left join lateral (
  -- The entry's latest correction (M5 COR-1): an update not by the author in the creation-day window, and not
  -- undone (an undo always follows the update it undoes directly, Q8).
  select x.id
  from (
    select a.id, a.user_id, a.at, a.action, a.same_day_author,
           lead(a.action) over (order by a.at, a.id) as next_action
    from audit_log a
    where a.entry_id = e.id
  ) x
  where x.action = 'update' and not x.same_day_author and x.next_action is distinct from 'undo'
  order by x.at desc, x.id desc
  limit 1
) lc on true
left join audit_log ca on ca.id = lc.id
left join app_user cu on cu.id = ca.user_id
left join audit_log la on la.id = (
  select a.id
  from audit_log a
  where a.entry_id = e.id
  order by a.at desc, a.id desc
  limit 1
)
where e.id = sqlc.arg(id)::uuid;

-- name: EntryBirdRows :many
-- The live bird_event rows of an entry (§4.7, §4.8) with their section's code, hall code (its own for a parentless
-- section) and retired flag, in tree order (hall, then compartment), ♀ before ♂, then by type. qty is the signed
-- stored delta (mortality and cull ≤ 0).
select b.id, b.section_id, s.code, coalesce(p.code, s.code)::text as hall_code,
       (s.retired_at is not null)::boolean as retired,
       b.sex, b.type, b.qty, b.cull_reason
from bird_event b
join section s on s.id = b.section_id
left join section p on p.id = s.parent_id
where b.entry_id = sqlc.arg(entry_id)::uuid
  and b.deleted_at is null
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code, b.sex, b.type;

-- name: EntryEggRows :many
-- The live egg_production rows of an entry (one per egg unit), with the unit's code and retired flag, in tree order.
select g.id, g.section_id, s.code, (s.retired_at is not null)::boolean as retired,
       g.floor_eggs, g.cat1, g.cat2, g.cat3, g.hatching_eggs, g.avg_egg_weight_g
from egg_production g
join section s on s.id = g.section_id
where g.entry_id = sqlc.arg(entry_id)::uuid
  and g.deleted_at is null
order by s.sort_order, s.code;

-- name: EntryFeedRows :many
-- The live feed_allocation rows of an entry with their section (code, hall code, retired) and feed type (name,
-- for_sex, active), in tree order, ♀ before ♂. kg_recorded is set only on imported rows (D163).
select fa.id, fa.section_id, s.code, coalesce(p.code, s.code)::text as hall_code,
       (s.retired_at is not null)::boolean as retired,
       fa.sex, fa.feed_type_id, ft.name as feed_type_name, ft.for_sex as feed_type_for_sex,
       ft.active as feed_type_active, fa.g_per_bird, fa.kg_recorded
from feed_allocation fa
join section s on s.id = fa.section_id
left join section p on p.id = s.parent_id
join feed_type ft on ft.id = fa.feed_type_id
where fa.entry_id = sqlc.arg(entry_id)::uuid
  and fa.deleted_at is null
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code, fa.sex;

-- name: EntryDeliveryRows :many
-- The live feed_delivery row of an entry (one per delivery entry) with its feed type and its price columns (a
-- delete's snapshot keeps the price, M5 Q9).
select d.id, d.feed_type_id, ft.name as feed_type_name, ft.for_sex as feed_type_for_sex,
       ft.active as feed_type_active, d.qty_kg, d.note, d.price_per_kg, d.total_lei, d.price_input, d.priced_by,
       d.priced_at, d.price_from_register
from feed_delivery d
join feed_type ft on ft.id = d.feed_type_id
where d.entry_id = sqlc.arg(entry_id)::uuid
  and d.deleted_at is null
order by d.id;
