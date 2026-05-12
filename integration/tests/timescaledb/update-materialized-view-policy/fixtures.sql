create table some_data (
  created_at timestamp without time zone not null,
  id integer not null,
  something integer not null,
  primary key ("created_at", "id")
);

select create_hypertable('some_data', 'created_at');

insert into some_data (created_at, id, something) values ('2023-02-03 23:51:39.045', 100, 1);
insert into some_data (created_at, id, something) values ('2023-02-03 23:50:59.845', 100, 2);
insert into some_data (created_at, id, something) values ('2023-02-03 23:49:48.844', 200, 10);
