alter table "projects" rename to "projects_0342a1ab69fe0887ec4f83d0a5c1a3678e4b4e2c9d631b82549f55cbaea90b75";
create table "projects" ("id" integer not null, "email" text not null, primary key ("id"));
create unique index idx_projects_email on projects (email);
insert into projects (id, email) select id, email from projects_0342a1ab69fe0887ec4f83d0a5c1a3678e4b4e2c9d631b82549f55cbaea90b75;
drop table projects_0342a1ab69fe0887ec4f83d0a5c1a3678e4b4e2c9d631b82549f55cbaea90b75;
