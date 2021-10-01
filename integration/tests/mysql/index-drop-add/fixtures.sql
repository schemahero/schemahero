create table unique_items (
  primary_id varchar(20) not null,
  secondary_id varchar(20) not null,
  unique key uc_compound_ids (primary_id, secondary_id)
);

-- removing columns and indexes on empty tables always works, but with data sequencing is important
insert into unique_items values ("primary_a", "seconday_1");
insert into unique_items values ("primary_a", "seconday_2");