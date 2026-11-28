-- Prețuri hrană (SPEC §6.8, §7.13, S12; M4 D3, D4, Q10–Q12). Package internal/prices.

-- name: PricesUnpriced :many
-- The S12 list (§7.13, M4 Q10): every live delivery without a price, oldest first (then farm, type, id), except
-- those dated inside a cycle of a closed series (closed_series_on, 0003; §3.1: read-only; they still count in Cost
-- `unpriced`).
select fd.id, fd.date, fd.qty_kg, f.name as farm_name, f.name_key as farm_name_key, ft.name as feed_type_name,
       ft.for_sex
from feed_delivery fd
join entry n on n.id = fd.entry_id and n.deleted_at is null
join farm f on f.id = fd.farm_id
join feed_type ft on ft.id = fd.feed_type_id
where fd.deleted_at is null
  and fd.total_lei is null
  and closed_series_on(fd.farm_id, fd.date) is null
order by fd.date, f.name_key, ft.name_key, fd.id;

-- name: PricesDeliveryRefs :many
-- What the level-1 checks of a price write name (§7.0b, M4 Q10, Q12), read before the transaction: each asked
-- delivery's date, kg, type name and current price, deleted ones too (a deleted delivery is 409 STALE in a set and
-- 404 in a change, not an unknown id). An id without a row is not in the result.
select fd.id, fd.farm_id, fd.date, fd.qty_kg, ft.name as feed_type_name, fd.price_per_kg, fd.total_lei,
       (fd.deleted_at is not null or n.deleted_at is not null)::boolean as deleted
from feed_delivery fd
join entry n on n.id = fd.entry_id
join feed_type ft on ft.id = fd.feed_type_id
where fd.id = any(sqlc.arg(ids)::uuid[]);

-- name: PricesDeliveryFarms :many
-- The farms of the asked deliveries, for the farm locks taken before the rows are locked (M4 D3).
select distinct fd.farm_id
from feed_delivery fd
where fd.id = any(sqlc.arg(ids)::uuid[]);

-- name: PricesLockDeliveries :many
-- The asked deliveries locked for the write (select … for update, after the farm locks; M4 D3): kg, date, type,
-- the five price columns, whether the delivery (or its entry) is deleted, and the closed series whose cycle holds
-- its date (closed_series_on, 0003; '' when none, §3.1).
select fd.id, fd.farm_id, fd.date, fd.qty_kg, ft.name as feed_type_name,
       fd.price_per_kg, fd.total_lei, fd.price_input, fd.priced_by, fd.priced_at, fd.price_from_register,
       (fd.deleted_at is not null or n.deleted_at is not null)::boolean as deleted,
       coalesce(closed_series_on(fd.farm_id, fd.date), '')::text as closed_series
from feed_delivery fd
join entry n on n.id = fd.entry_id
join feed_type ft on ft.id = fd.feed_type_id
where fd.id = any(sqlc.arg(ids)::uuid[])
order by fd.id
for update of fd;

-- name: PricesSet :execrows
-- Prices a batch of unpriced deliveries in one statement (§6.8, M4 D3): the parallel arrays hold each delivery's
-- price per kg, total and typed input; priced_by is the caller, priced_at the save's instant (= its audit row's at,
-- the price undo matches on it, D4). Only rows still unpriced and live are updated; the caller compares the count.
update feed_delivery fd
set price_per_kg = u.price_per_kg,
    total_lei = u.total_lei,
    price_input = u.price_input,
    priced_by = sqlc.arg(priced_by)::uuid,
    priced_at = sqlc.arg(priced_at)::timestamptz,
    price_from_register = false
from (select (sqlc.arg(ids)::uuid[])[g.i] as id,
             (sqlc.arg(prices_per_kg)::numeric[])[g.i] as price_per_kg,
             (sqlc.arg(totals_lei)::numeric[])[g.i] as total_lei,
             (sqlc.arg(price_inputs)::text[])[g.i] as price_input
      from generate_subscripts(sqlc.arg(ids)::uuid[], 1) as g(i)) as u
where fd.id = u.id
  and fd.deleted_at is null
  and fd.total_lei is null;

-- name: PricesUndo :execrows
-- Clears the prices of a batch saved at priced_at by priced_by (the price undo, §7.13, M4 D4): the five price
-- columns back to null, only on rows that still hold that save's prices.
update feed_delivery
set price_per_kg = null,
    total_lei = null,
    price_input = null,
    priced_by = null,
    priced_at = null
where id = any(sqlc.arg(ids)::uuid[])
  and deleted_at is null
  and priced_at = sqlc.arg(priced_at)::timestamptz
  and priced_by = sqlc.arg(priced_by)::uuid;

-- name: PricesChange :execrows
-- Changes a priced delivery's price (§6.8 Modifică, M4 Q12): the new pair and typed input, the caller and the
-- instant; a register price becomes the caller's (price_from_register false).
update feed_delivery
set price_per_kg = sqlc.arg(price_per_kg)::numeric,
    total_lei = sqlc.arg(total_lei)::numeric,
    price_input = sqlc.arg(price_input)::text,
    priced_by = sqlc.arg(priced_by)::uuid,
    priced_at = sqlc.arg(priced_at)::timestamptz,
    price_from_register = false
where id = sqlc.arg(id)::uuid
  and deleted_at is null
  and total_lei is not null;

-- name: PricesAuditRow :one
-- The price_set audit row a price undo token names (M4 D4): its instant, author, action and new_values (the
-- deliveries of the batch). No row = no such audit row.
select at, user_id, action, new_values
from audit_log
where id = sqlc.arg(id)::bigint;
