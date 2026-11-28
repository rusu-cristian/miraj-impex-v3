-- Edit / Corectează PUT /api/entries/{id} and the undo of an update (SPEC §7.8, §6.15; plan M5-D5, D6, Q8).
-- Owned by entries/update.go, entries/daily_update.go, entries/feed_update.go, entries/delivery_update.go and
-- entries/undo.go.
-- Soft delete and un-delete of child rows reuse EntrySoftDeleteRows / EntryRestoreRows (entrydelete.sql); new
-- Zilnic rows reuse the create's inserts (daily.sql).

-- name: EntryUpdateBirdRows :execrows
-- Changes live bird_event rows of one entry in place (plan M5-D6: a changed value keeps its row, so the partial
-- unique index entry stays): rows is a JSON array of {id, qty, cull_reason} (qty the signed delta, cull_reason null
-- except on a cull). Returns how many rows changed (the caller checks it is every row it asked for).
update bird_event b
set qty = r.qty, cull_reason = r.cull_reason
from jsonb_to_recordset(sqlc.arg(rows)::jsonb) as r(id uuid, qty integer, cull_reason cull_reason)
where b.id = r.id
  and b.entry_id = sqlc.arg(entry_id)::uuid
  and b.deleted_at is null;

-- name: EntryUpdateEggRows :execrows
-- Changes live egg_production rows of one entry in place: rows is a JSON array of {id, floor, cat1, cat2, cat3,
-- hatching, avg_weight_g}, null for a field left empty. total_eggs is generated and never set. Returns how many rows
-- changed.
update egg_production g
set floor_eggs       = r.floor,
    cat1             = r.cat1,
    cat2             = r.cat2,
    cat3             = r.cat3,
    hatching_eggs    = r.hatching,
    avg_egg_weight_g = r.avg_weight_g
from jsonb_to_recordset(sqlc.arg(rows)::jsonb)
     as r(id uuid, floor integer, cat1 integer, cat2 integer, cat3 integer, hatching integer, avg_weight_g numeric)
where g.id = r.id
  and g.entry_id = sqlc.arg(entry_id)::uuid
  and g.deleted_at is null;

-- name: EntryUpdateFeedRows :execrows
-- Changes live feed_allocation rows of one entry in place (plan M5-D6, Q11: the row keeps its (section, day, sex)
-- key and its unique index entry): rows is a JSON array of {id, feed_type_id, g, kg_recorded}. kg_recorded is null
-- unless the row is put back as it was (the undo of an update): a row whose g is corrected is charged g × birds.
-- Returns how many rows changed.
update feed_allocation fa
set feed_type_id = r.feed_type_id,
    g_per_bird   = r.g,
    kg_recorded  = r.kg_recorded
from jsonb_to_recordset(sqlc.arg(rows)::jsonb) as r(id uuid, feed_type_id uuid, g numeric, kg_recorded numeric)
where fa.id = r.id
  and fa.entry_id = sqlc.arg(entry_id)::uuid
  and fa.deleted_at is null;

-- name: EntryInsertFeedRows :exec
-- The rows a Hrană zilnică correction adds (§4.8) in one statement: rows is a JSON array of
-- {section_id, sex, feed_type_id, g}. A key already live in another entry violates feed_allocation_uq (23505, mapped
-- after the transaction to 409 ALREADY_ENTERED, M3 D2).
insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
select sqlc.arg(entry_id)::uuid, sqlc.arg(farm_id)::uuid, r.section_id, sqlc.arg(date)::date, r.sex, r.feed_type_id,
       r.g
from jsonb_to_recordset(sqlc.arg(rows)::jsonb) as r(section_id uuid, sex sex, feed_type_id uuid, g numeric);

-- name: EntryUpdateDeliveryRow :execrows
-- Changes the live feed_delivery row of one entry in place (M5 Q11): type, kg, note and, for a priced delivery, the
-- price pair re-derived from the typed input (calc.DerivePrice; both null for an unpriced one). priced_by, priced_at,
-- price_input and price_from_register are not touched. Returns how many rows changed (0: the row is gone).
update feed_delivery
set feed_type_id = sqlc.arg(feed_type_id)::uuid,
    qty_kg       = sqlc.arg(qty_kg)::numeric,
    note         = sqlc.narg(note)::text,
    price_per_kg = sqlc.narg(price_per_kg)::numeric,
    total_lei    = sqlc.narg(total_lei)::numeric
where id = sqlc.arg(id)::uuid
  and entry_id = sqlc.arg(entry_id)::uuid
  and deleted_at is null;
