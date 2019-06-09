create table projects (
  id integer primary key not null,
  name varchar(255) not null
);

create table org (
  id integer primary key not null,
  project_id integer references projects(id)
);
