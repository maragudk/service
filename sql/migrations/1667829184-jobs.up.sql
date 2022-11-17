create table jobs (
  id integer primary key,
  name text not null,
  payload text not null,
  timeout int not null,
  run text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  received text,
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ'))
) strict;

create trigger jobs_updated_timestamp after update on jobs begin
  update jobs set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create index jobs_created_idx on jobs (created);
create index jobs_run_idx on jobs (run);
