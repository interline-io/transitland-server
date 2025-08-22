-- Check if pg_stat_statements is enabled and monitor query patterns
-- Run this in your PostgreSQL database

-- 1. Verify pg_stat_statements extension is available
SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';

-- 2. Check if it's installed
SELECT * FROM pg_extension WHERE extname = 'pg_stat_statements';

-- 3. Monitor your specific query pattern
SELECT 
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    plans,
    total_plan_time,
    mean_plan_time
FROM pg_stat_statements 
WHERE query ILIKE '%active_services%' 
   OR query ILIKE '%route_lookup%'
   OR query ILIKE '%gtfs_stop_times%'
ORDER BY calls DESC;

-- 4. Check for prepared statements specifically
SELECT 
    query,
    calls,
    total_exec_time / calls as avg_exec_time,
    total_plan_time / calls as avg_plan_time
FROM pg_stat_statements 
WHERE query LIKE 'EXECUTE%' -- Prepared statements show as EXECUTE
ORDER BY calls DESC;
