create table users (
  id integer primary key not null,
  email varchar(255) not null
);

create table projects (
  id integer primary key not null,
  name varchar(255) not null
);

create table issues (
  id integer primary key not null,
  project_id integer references users(id)
);

create table other (
  id integer primary key not null,
  project_id integer references users(id) on delete cascade
);
