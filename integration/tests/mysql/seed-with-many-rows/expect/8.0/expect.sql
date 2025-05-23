create table `users` (`id` int (11), `login` varchar (255), `name` varchar (255) not null, primary key (`id`));
insert into users (id, login, name) values (1, 'test', 'test2') on duplicate key update id=1, login='test', name='test2';
insert into users (id, login, name) values (2, 'other', 'test2') on duplicate key update id=2, login='other', name='test2';
insert into users (id, login, name) values (3, 'yet', 'someone') on duplicate key update id=3, login='yet', name='someone';
insert into users (id, login, name) values (4, 'another', 'more') on duplicate key update id=4, login='another', name='more';
