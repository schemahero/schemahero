create table user_projects (
  user_id integer not null,
  project_id integer not null,
  primary key (user_id, project_id)
);
