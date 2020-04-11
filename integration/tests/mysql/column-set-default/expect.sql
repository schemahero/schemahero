alter table `users` modify column `account_type` varchar (10) null default "trial";
update `users` set `account_type`="trial" where `account_type` is null;
alter table `users` modify column `account_type` varchar (10) not null default "trial";
alter table `users` modify column `num_seats` int (11) null default "5";
update `users` set `num_seats`="5" where `num_seats` is null;
alter table `users` modify column `num_seats` int (11) not null default "5";
