alter table `projects` modify column `name` varchar (255) null default "name me";
update `projects` set `name`="name me" where `name` is null;
alter table `projects` modify column `name` varchar (255) not null default "name me";
