alter table projects convert to character set utf16 collate utf16_german2_ci;
alter table `users` modify column `email` varchar (255) character set utf8mb4 not null;
