create table users (
  id integer primary key not null,
  login varchar(255),
  name varchar(255) not null
);

create table other (
  id integer primary key not null,
  something varchar(255) not null
);
