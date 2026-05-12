create materialized view "some_data_view" with (timescaledb.continuous) as select time_bucket('1 minute'::interval, created_at) as minute_bucket, id, sum(something) as total
from some_data
group by minute_bucket, id
 with data;
select add_continuous_aggregate_policy('some_data_view', start_offset => INTERVAL '1 day', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour');
