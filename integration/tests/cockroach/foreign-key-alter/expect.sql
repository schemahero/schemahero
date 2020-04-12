create table "issues" ("id" integer, "project_id" integer, primary key ("id"), constraint renamed_fkey foreign key (project_id) references projects (id));
