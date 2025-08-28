alter table "orders" alter column "total" type numeric (10, 2);
alter table "orders" alter column "total_with_tax" type numeric (10, 2);
update "orders" set "total_with_tax" = "total" * 1.08;
