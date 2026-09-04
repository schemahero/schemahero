create table t1 (
  c1 integer not null,
  c11 integer not null,
  primary key (c1, c11)
);

create table t2 (
  id integer primary key not null,
  c2 integer not null,
  c22 integer not null,
  constraint t2_c2_c22_fkey foreign key (c2, c22) references t1 (c1, c11) on delete cascade
);
