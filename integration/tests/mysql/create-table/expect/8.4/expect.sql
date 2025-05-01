insert into other (id, something) values (1, 'one') on duplicate key update id=1, something='one';
insert into other (id, something) values (2, 'two') on duplicate key update id=2, something='two';
create table `users` (`id` int (11), `login` varchar (255), `name` varchar (255) not null default 'ethan', `email` varchar (255) not null, primary key (`id`), key idx_users_name (name), unique key idx_users_email (email));
