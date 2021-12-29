create table `users` (`id` int (11), `login` varchar (255), `address` text, primary key (`id`));
insert into users (id, login, address) values (1, 'test', CONCAT_WS(CHAR(10 using utf8), '123 Main St', 'Los Angeles, CA 90015')) on duplicate key update id=1, login='test', address=CONCAT_WS(CHAR(10 using utf8), '123 Main St', 'Los Angeles, CA 90015');
