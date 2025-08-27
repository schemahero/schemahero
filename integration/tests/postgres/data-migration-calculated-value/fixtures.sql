create table orders (
  id integer primary key not null,
  customer_id integer not null,
  total numeric(10, 2),
  total_with_tax numeric(10, 2)
);

-- Insert test data
insert into orders (id, customer_id, total, total_with_tax) values 
  (1, 100, 100.00, null),
  (2, 101, 250.50, null),
  (3, 102, 75.25, null);