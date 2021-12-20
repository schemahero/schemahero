create table `table1` (
  `teamid` char (64) not null,
  `imageid` char (64) not null,
  `v2_blobsum` varchar (255) null,
  key idx_table1_imageid_teamid (imageid, teamid)
) default character set latin1;
