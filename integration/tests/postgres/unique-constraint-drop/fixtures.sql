create table projects (
  id integer primary key not null,
  name varchar(255) not null,
  CONSTRAINT ukey_projects_name UNIQUE (name)
);
