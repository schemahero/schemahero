create table `table1` (
  `manifest_id` varchar (255) not null,
  `team_id` char (36) not null,
  `namespace` varchar (255) not null,
  `image_name` varchar (255) not null,
  `image_tag` varchar (255) not null,
  `signed_json` mediumtext not null,
  `content_type` varchar (255) null,
  primary key (`manifest_id`, `team_id`, `namespace`)
) default character set latin1;
