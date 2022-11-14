create table foo (
  id integer primary key,
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ'))
) strict;

create trigger foo_updated_timestamp after update on foo begin
  update foo set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;
