create table projects (id integer primary key not null, name text not null);

create table org (id integer primary key not null, project_id integer references projects(id));
