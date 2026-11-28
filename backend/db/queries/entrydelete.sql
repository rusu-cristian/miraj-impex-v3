-- Delete, undo of a create, restore (SPEC §7.8, §6.15; plan M5-D7, Q8, Q9).
-- Owned by entries/delete.go and entries/restore.go. Soft delete is an UPDATE (il_app has no DELETE on these tables).

-- name: EntryDeleteContext :one
-- What the delete Dialog and the Toasts name (§6.15, M5 Q9): the farm's name, the author's name, and who made the
-- entry's latest correction and when ("… și corectată de Ion Rusu pe 18.11.2026"; as EntryReadHead, M5 COR-1: an
-- update not by the author in the creation-day window and not undone; null when none). No row = no such entry.
select f.name as farm_name, au.name as author, cu.name as corrected_by_name, ca.at as corrected_at
from entry e
join farm f on f.id = e.farm_id
left join app_user au on au.id = e.created_by
left join lateral (
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
where e.id = sqlc.arg(id)::uuid;

-- name: EntrySoftDeleteRows :one
-- Soft-deletes the asked live child rows of one entry (plan M5-D7): deleted_at = the injected clock's instant on each
-- row whose id is listed for its table. One statement over the four child tables; returns how many rows changed (the
-- caller checks it is every row it asked for, so a row deleted meanwhile is not silently skipped).
with b as (
  update bird_event set deleted_at = sqlc.arg(now)::timestamptz
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(bird_ids)::uuid[]) and deleted_at is null
  returning 1
), g as (
  update egg_production set deleted_at = sqlc.arg(now)::timestamptz
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(egg_ids)::uuid[]) and deleted_at is null
  returning 1
), fa as (
  update feed_allocation set deleted_at = sqlc.arg(now)::timestamptz
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(feed_ids)::uuid[]) and deleted_at is null
  returning 1
), fd as (
  update feed_delivery set deleted_at = sqlc.arg(now)::timestamptz
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(delivery_ids)::uuid[]) and deleted_at is null
  returning 1
)
select ((select count(*) from b) + (select count(*) from g) + (select count(*) from fa)
        + (select count(*) from fd))::integer as changed;

-- name: EntryRestoreRows :one
-- Un-deletes exactly the listed child rows of one entry (a delete's audit snapshot, plan M5-D7): rows soft-deleted
-- earlier by an edit are not listed, so they stay deleted. A row whose key another entry took meanwhile fails the
-- partial unique index (23505), mapped after the transaction to 409 ALREADY_ENTERED (EntryRestoreHolder). Returns how
-- many rows changed.
with b as (
  update bird_event set deleted_at = null
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(bird_ids)::uuid[]) and deleted_at is not null
  returning 1
), g as (
  update egg_production set deleted_at = null
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(egg_ids)::uuid[]) and deleted_at is not null
  returning 1
), fa as (
  update feed_allocation set deleted_at = null
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(feed_ids)::uuid[]) and deleted_at is not null
  returning 1
), fd as (
  update feed_delivery set deleted_at = null
  where entry_id = sqlc.arg(entry_id)::uuid and id = any(sqlc.arg(delivery_ids)::uuid[]) and deleted_at is not null
  returning 1
)
select ((select count(*) from b) + (select count(*) from g) + (select count(*) from fa)
        + (select count(*) from fd))::integer as changed;

-- name: EntryDeleteHeader :execrows
-- Marks a live entry deleted (§4.7 soft delete): deleted_at = the injected clock's instant, deleted_by = the caller,
-- delete_reason = the Motiv (null for an author's same-day delete or an undo). 0 rows = already deleted.
update entry
set deleted_at    = sqlc.arg(now)::timestamptz,
    deleted_by    = sqlc.arg(deleted_by)::uuid,
    delete_reason = sqlc.narg(delete_reason)::text
where id = sqlc.arg(id)::uuid
  and deleted_at is null;

-- name: EntryRestoreHeader :execrows
-- Un-deletes an entry (restore of a whole-save delete, M5 Q8). 0 rows = it was not deleted.
update entry
set deleted_at = null, deleted_by = null, delete_reason = null
where id = sqlc.arg(id)::uuid
  and deleted_at is not null;

-- name: EntryRestoreHolder :one
-- After a restore failed on a unique key (M3 D2, plan M5-D2): the first listed row of the entry whose key another
-- live row holds now, with that row's entry, its creation time, author and source, the section's code and the row's
-- sex ('' for an egg row) and table (bird | eggs | feed), in that table order. No row = nothing collides any more.
select x.holder_entry_id, x.code, x.sex, x.tbl, n.created_at, n.created_by, n.source
from (
  select o.entry_id as holder_entry_id, s.code, b.sex::text as sex, 'bird'::text as tbl, 1 as ord
  from bird_event b
  join bird_event o on o.section_id = b.section_id and o.date = b.date and o.sex = b.sex and o.type = b.type
       and o.deleted_at is null and o.id <> b.id
  join section s on s.id = b.section_id
  where b.entry_id = sqlc.arg(entry_id)::uuid and b.id = any(sqlc.arg(bird_ids)::uuid[])
    and b.type in ('mortality', 'cull', 'sexing_error')
  union all
  select o.entry_id, s.code, ''::text, 'eggs'::text, 2
  from egg_production g
  join egg_production o on o.section_id = g.section_id and o.date = g.date and o.deleted_at is null and o.id <> g.id
  join section s on s.id = g.section_id
  where g.entry_id = sqlc.arg(entry_id)::uuid and g.id = any(sqlc.arg(egg_ids)::uuid[])
  union all
  select o.entry_id, s.code, fa.sex::text, 'feed'::text, 3
  from feed_allocation fa
  join feed_allocation o on o.section_id = fa.section_id and o.date = fa.date and o.sex = fa.sex
       and o.deleted_at is null and o.id <> fa.id
  join section s on s.id = fa.section_id
  where fa.entry_id = sqlc.arg(entry_id)::uuid and fa.id = any(sqlc.arg(feed_ids)::uuid[])
) x
join entry n on n.id = x.holder_entry_id
order by x.ord, x.code, x.sex
limit 1;
