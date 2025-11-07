insert into users (id, login, name) values (1, 'test', 'test2') on duplicate key update id=1, login='test', name='test2';
insert into users (id, login, name) values (2, 'admin', 'Administrator') on duplicate key update id=2, login='admin', name='Administrator';
