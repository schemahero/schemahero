alter table `unique_items` drop index `uc_compound_ids`;
alter table `unique_items` drop column `secondary_id`;
alter table `unique_items` add column `item_type` varchar (20);
create index uc_item_type on unique_items (item_type);
