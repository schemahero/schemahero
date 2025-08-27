create table events (
  id integer primary key not null,
  event_name varchar(255) not null,
  created_at timestamp
);

-- Insert test data
insert into events (id, event_name, created_at) values 
  (1, 'user_signup', '2024-01-15 10:30:00'),
  (2, 'user_login', '2024-01-15 14:45:30'),
  (3, 'user_logout', '2024-01-15 18:20:15');