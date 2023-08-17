create table user_projects (
  manifest_id text not null,
  team_id text not null,
  project_id text not null,
  spec text not null,
  primary key (manifest_id, team_id, project_id)
);
