create table users (
  id integer primary key not null,
  email varchar(255) not null,
  status varchar(50) default 'active'
);

-- Insert test data
insert into users (id, email, status) values 
  (1, 'user1@example.com', 'active'),
  (2, 'user2@example.com', 'inactive'),
  (3, 'user3@example.com', 'pending');