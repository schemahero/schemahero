create table `complex_data` (`id` int (11), `complex_data` json not null, primary key (`id`));
alter table `other` add column `complex_data` json not null;
