alter table "projects" alter column "name" set default 'unnamed';
update "projects" set "name"='unnamed' where "name" is null;
alter table "projects" alter column "name" set not null;
