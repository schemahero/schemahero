alter table `user_projects` add column `project_id` int (11) not null;
alter table `user_projects` add constraint `user_projects_pkey` primary key (`user_id`, `project_id`);
