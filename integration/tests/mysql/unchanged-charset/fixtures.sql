create table `table1` (
  `channel_id` varchar (255) not null,
  `created_at` datetime not null,
  `updated_at` datetime null,
  `version_label` varchar (255) null,
  `release_notes` text character set utf8mb4 null,
  `release_sequence` bigint (20) not null,
  `channel_sequence` bigint (20) not null,
  `airgap_locked_at` datetime null,
  `force_airgap_build` tinyint (1) not null default '0',
  primary key (`channel_id`, `channel_sequence`),
  key idx_channel_id_release_sequence (channel_id, release_sequence)
) default character set latin1;
