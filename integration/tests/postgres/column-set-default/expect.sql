alter table "users" alter column "account_type" set default 'trial';
alter table "users" alter column "num_seats" set default '5';
alter table "users" alter column "created_on" set default current_timestamp;
