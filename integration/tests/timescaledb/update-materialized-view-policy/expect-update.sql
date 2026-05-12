select remove_continuous_aggregate_policy('some_data_view');
select add_continuous_aggregate_policy('some_data_view', start_offset => INTERVAL '2 days', end_offset => INTERVAL '2 hours', schedule_interval => INTERVAL '2 hours');
