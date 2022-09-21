alter table "issues" rename to "issues_36a17f93aaada01c0b438606c254940bbe48a88c1bc192feacfb363b0640107b";
create table "issues" ("id" integer not null, "project_id" integer, primary key ("id"), constraint renamed_fkey foreign key (project_id) references projects (id));
insert into issues (id, project_id) select id, project_id from issues_36a17f93aaada01c0b438606c254940bbe48a88c1bc192feacfb363b0640107b;
drop table issues_36a17f93aaada01c0b438606c254940bbe48a88c1bc192feacfb363b0640107b;
