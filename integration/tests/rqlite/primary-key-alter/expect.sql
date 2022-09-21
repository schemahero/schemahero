alter table "user_projects" rename to "user_projects_a0870d4dbd1f49995bcf4142bed53756cfb850fa68e9ae0dfdd29980264e9793";
create table "user_projects" ("user_id" integer not null, "project_id" integer not null, primary key ("user_id", "project_id"));
insert into user_projects (user_id, project_id) select user_id, project_id from user_projects_a0870d4dbd1f49995bcf4142bed53756cfb850fa68e9ae0dfdd29980264e9793;
drop table user_projects_a0870d4dbd1f49995bcf4142bed53756cfb850fa68e9ae0dfdd29980264e9793;
