alter table user_projects drop constraint "user_projects_pkey";
alter table user_projects add constraint user_projects_pkey primary key (team_id, project_id, manifest_id);
