create table "users" (
  "col_one" bytea not null,
  "col_two" bytea not null,
  primary key ("col_one", "col_two")
);