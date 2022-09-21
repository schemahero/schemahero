create table "user_project" ("user_id" integer, "project_id" integer, "misc_id" text not null, primary key ("user_id", "project_id"), constraint user_project_user_id_fkey foreign key (user_id) references users (id), constraint user_project_project_id_fkey foreign key (project_id) references projects (id), constraint misc_named_fk foreign key (misc_id) references misc (pk) on delete cascade);

