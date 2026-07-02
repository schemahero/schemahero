create index "idx_existing_table_name" on "myschema"."existing_table" (name);
create table "myschema"."new_table" ("id" integer, "name" character varying (255) not null, primary key ("id"));
create index "idx_new_table_name" on "myschema"."new_table" (name);
