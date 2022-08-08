alter table "org" alter column "project_id" set not null;
alter table org drop constraint "org_project_id_fkey";
