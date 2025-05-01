create table `users` (`id` int (11), `login` varchar (255), `name` varchar (255) not null default 'ethan', `email` varchar (255) not null, primary key (`id`), key idx_users_name (name), unique key idx_users_email (email));
insert into users (id, login, name, email) values (1, 'test', 'test2', 'email@mail.com') on duplicate key update id=1, login='test', name='test2', email='email@mail.com';
