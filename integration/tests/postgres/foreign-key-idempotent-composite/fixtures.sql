create table categories (
  category_id integer not null,
  account_id integer not null,
  primary key (category_id, account_id)
);

create table category_pages (
  id integer primary key not null,
  category_id integer not null,
  account_id integer not null,
  constraint category_pages_category_id_account_id_fkey foreign key (category_id, account_id) references categories (category_id, account_id) on delete cascade
);
