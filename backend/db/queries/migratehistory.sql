-- History migration (SPEC §10.3): bulk COPY of the plan, the per-farm audit rows and the checks of Verify.
-- The column lists are explicit so a schema change breaks the generated code at compile time. The generated
-- column egg_production.total_eggs is never listed.

-- name: CopyUsers :copyfrom
-- --seed-dev users (M1-D7).
insert into app_user (id, name, login, password_hash, role, active, password_generated_at, created_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CopyUserFarms :copyfrom
insert into user_farm (user_id, farm_id) values ($1, $2);

-- name: CopyGuides :copyfrom
insert into guide (id, name, name_key, type, created_at) values ($1, $2, $3, $4, $5);

-- name: CopyGuideWeeks :copyfrom
insert into guide_week (guide_id, week, sex, feed_g, body_weight_g, mortality_cum_pct, lay_pct, hatching_pct, egg_weight_g)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CopyFarms :copyfrom
insert into farm (id, name, name_key, type, guide_id, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7);

-- name: CopySections :copyfrom
insert into section (id, farm_id, code, parent_id, sort_order, created_at) values ($1, $2, $3, $4, $5, $6);

-- name: CopyFeedTypes :copyfrom
insert into feed_type (id, name, name_key, for_sex, active, created_at) values ($1, $2, $3, $4, $5, $6);

-- name: CopySeries :copyfrom
insert into series (id, name, name_key, year, status, placement_date, placement_farm_id, placement_inferred,
                    age_at_placement_days, closed_at, created_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CopyCycles :copyfrom
insert into cycle (id, series_id, farm_id, guide_id, start_date, end_date, started_by, start_inferred, ended_by,
                   created_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: CopyEntries :copyfrom
insert into entry (id, kind, farm_id, peer_farm_id, date, series_id, buyer, note, source, created_by, created_at,
                   updated_by, updated_at, corrected)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: CopyBirdEvents :copyfrom
insert into bird_event (id, entry_id, farm_id, section_id, date, sex, type, qty, cull_reason, peer_section_id, note)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CopyEggProductions :copyfrom
insert into egg_production (id, entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs,
                            avg_egg_weight_g)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CopyFeedAllocations :copyfrom
insert into feed_allocation (id, entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird, kg_recorded)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CopyFeedDeliveries :copyfrom
insert into feed_delivery (id, entry_id, farm_id, feed_type_id, date, qty_kg, note, price_per_kg, total_lei,
                           price_input, priced_by, priced_at, price_from_register)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: CopyFeedStockAdjustments :copyfrom
insert into feed_stock_adjustment (id, entry_id, farm_id, feed_type_id, date, qty_kg, note)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: CopyWeighings :copyfrom
insert into weighing (id, entry_id, farm_id, section_id, date, sex, avg_weight_g, cv_pct, uniformity_pct, sample_size)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: CopyLegacyMortality :copyfrom
insert into legacy_mortality (id, entry_id, farm_id, section_id, date, sex, deaths, emptying)
values ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: MigrationEntryExists :one
-- §10.3: the import runs once; a database that holds a migration entry is refused (unless --replace).
select exists (select 1 from entry where kind = 'migration') as exists;

-- name: InsertEntryAudit :exec
-- An audit row about an entry (SPEC §4.10). `at` comes from the injected clock (M1-D2); user_id null = migration.
insert into audit_log (at, user_id, action, entity, entity_id, entry_id, entry_kind, record_types, farm_ids,
                       record_date, reason, summary, changes, old_values, new_values, source)
values (sqlc.arg(at), sqlc.narg(user_id), sqlc.arg(action), 'entry', sqlc.arg(entry_id), sqlc.arg(entry_id),
        sqlc.arg(entry_kind), sqlc.arg(record_types)::text[], sqlc.arg(farm_ids)::uuid[], sqlc.narg(record_date),
        sqlc.narg(reason), sqlc.narg(summary), coalesce(sqlc.narg(changes)::jsonb, '[]'::jsonb),
        sqlc.narg(old_values), sqlc.narg(new_values), sqlc.arg(source));

-- name: VerifyFarmDeliveries :one
-- §10.3 check: live deliveries of a farm, their Σ lei and how many have no price.
select count(*)::integer                                    as deliveries,
       coalesce(sum(total_lei), 0)::numeric                 as total_lei,
       (count(*) filter (where total_lei is null))::integer as unpriced
from feed_delivery
where farm_id = sqlc.arg(farm_id)::uuid
  and deleted_at is null;

-- name: VerifyEggTotal :one
-- §10.3 check: Σ total_eggs of a farm over [from_day, to_day].
select coalesce(sum(total_eggs), 0)::bigint as eggs
from egg_production
where farm_id = sqlc.arg(farm_id)::uuid
  and deleted_at is null
  and date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date;

-- name: VerifyNegativeAlive :many
-- §6.1 / M1-D4: the section × sex days that end below 0 (none expected), first ones first.
select s.code, d.sex, d.date, d.alive_end
from v_section_sex_day d
join section s on s.id = d.section_id
where d.alive_end < 0
order by d.date, s.code, d.sex
limit 5;
