alter table `users` modify column `updated_at` datetime not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP;
alter table `users` add column `modified_at` timestamp on update CURRENT_TIMESTAMP;
