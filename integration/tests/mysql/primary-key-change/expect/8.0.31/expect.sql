alter table `table1` drop primary key;
alter table `table1` add constraint `table1_pkey` primary key (`team_id`, `namespace`, `manifest_id`);
