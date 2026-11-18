-- Two schemas with identically named FK constraints.
-- Joining information_schema.referential_constraints on constraint_name alone
-- duplicates rows for public.t2 and falsely reports FK drift on re-plan.
create schema other;

create table other.t1 (
  c1 integer primary key not null
);

create table other.t2 (
  id integer primary key not null,
  c2 integer not null,
  constraint t2_c2_fkey foreign key (c2) references other.t1 (c1) on delete cascade
);

create table t1 (
  c1 integer primary key not null
);

create table t2 (
  id integer primary key not null,
  c2 integer not null,
  constraint t2_c2_fkey foreign key (c2) references t1 (c1) on delete cascade
);
