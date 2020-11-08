begin transaction;
alter table `projects` rename to `projects_dbb570ff3221477e0d90f5b378ea3402bb752984d035a84ba65589f5a96aeda6`;
create table `projects` (`id` int (11) not null, `name` varchar (255) not null, `icon_uri` varchar (255), primary key (`id`));
insert into projects (id, name, icon_uri) select id, name, icon_uri from projects_dbb570ff3221477e0d90f5b378ea3402bb752984d035a84ba65589f5a96aeda6;
drop table projects_dbb570ff3221477e0d90f5b378ea3402bb752984d035a84ba65589f5a96aeda6;
commit;
