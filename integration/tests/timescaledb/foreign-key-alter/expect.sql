alter table issues drop constraint "issues_project_id_fkey";
alter table issues add constraint renamed_fkey foreign key ("project_id") references "projects" ("id");
