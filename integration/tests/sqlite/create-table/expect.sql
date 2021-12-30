begin transaction;
alter table `other` rename to `other_ffb36fbcac4bec40514212401762afb5a139e8371ddd8ca0f996b29bfd6ccc27`;
create table `other` (`id` int (11), `something` varchar (255) not null, primary key (`id`));
insert into other (id, something) select id, something from other_ffb36fbcac4bec40514212401762afb5a139e8371ddd8ca0f996b29bfd6ccc27;
drop table other_ffb36fbcac4bec40514212401762afb5a139e8371ddd8ca0f996b29bfd6ccc27;
commit;
replace into other (id, something) values (1, 'one');
replace into other (id, something) values (2, 'two');
create table `users` (`id` int (11), `login` varchar (255), `name` varchar (255) not null default 'ethan', primary key (`id`));
