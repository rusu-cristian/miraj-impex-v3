-- +goose Up
-- closed_series_on(farm, day): the name of the closed series whose cycle at the farm holds the day (start and end
-- included; an open cycle has no end), null when there is none (SPEC §3.1: a closed series is read-only). The one
-- definition of "a date inside a cycle of a closed series" the price writes and list, the Livrare save and the
-- Livrări actions read (M4 ARC-2). Cycles of one farm never overlap, so at most one cycle matches; the oldest wins
-- if they ever did. The body is a SELECT without FROM, so the planner may inline it into the calling query.
-- Functions are executable by PUBLIC (0002_roles.sql revokes nothing on them), so il_app needs no grant.

-- +goose StatementBegin
create function closed_series_on(p_farm uuid, p_date date) returns text
  language sql stable as $$
    select (select s.name
            from cycle c
            join series s on s.id = c.series_id
            where c.farm_id = p_farm
              and c.start_date <= p_date
              and (c.end_date is null or c.end_date >= p_date)
              and s.status = 'closed'
            order by c.start_date
            limit 1) $$;
-- +goose StatementEnd

-- +goose Down
drop function closed_series_on(uuid, date);
