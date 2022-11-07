create table foo (
  id integer primary key,
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) strict;

create trigger foo_updated_timestamp after update on foo begin
  update foo set updated = (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')) where id = old.id;
end;
