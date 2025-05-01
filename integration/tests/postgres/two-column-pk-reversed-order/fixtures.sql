create table "users" (
  "col_two" bytea not null,
  "col_one" bytea not null,
  primary key ("col_one", "col_two")
);