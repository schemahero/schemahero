create table users (id integer primary key not null, email text not null);

create table projects (id integer primary key not null, name text not null);

create table issues (id integer primary key not null, project_id integer references users(id));

create table other (id integer primary key not null, project_id integer references users(id) on delete cascade);
