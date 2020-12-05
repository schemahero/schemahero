begin transaction;
alter table `projects` rename to `projects_4af4d804ad6edd0ca84e50bf32fa5413e3542e2a26d6076851ef1fee7318aadf`;
create table `projects` (`id` int (11), `name` varchar (255) not null default 'name me', `icon_uri` varchar (255), primary key (`id`));
insert into projects (id, name, icon_uri) select id, name, icon_uri from projects_4af4d804ad6edd0ca84e50bf32fa5413e3542e2a26d6076851ef1fee7318aadf;
drop table projects_4af4d804ad6edd0ca84e50bf32fa5413e3542e2a26d6076851ef1fee7318aadf;
commit;
