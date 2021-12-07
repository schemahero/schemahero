create table `users` (`id` int (11), `login` varchar (255), `name` varchar (255) not null, primary key (`id`));
insert ignore into users (id, login, name) values (1, 'test', 'test2');
insert ignore into users (id, login, name) values (2, 'other', 'test2');
insert ignore into users (id, login, name) values (3, 'yet', 'someone');
insert ignore into users (id, login, name) values (4, 'another', 'more');
