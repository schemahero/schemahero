create table other (
  id integer primary key not null,
  something varchar(255) not null
);

insert into other (id, something) values (1, 'one hundred');
insert into other (id, something) values (2, 'two hundred');

create table "users" ("id" integer, "email" character varying (255) not null, primary key ("id"));