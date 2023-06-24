-- route_attributes.txt
insert into ext_plus_route_attributes(route_id,feed_version_id,category,subcategory,running_way) values (
    (select r.id from gtfs_routes r join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and route_id = '01'),
    (select r.feed_version_id from gtfs_routes r join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and route_id = '01'),
    2,
    201,
    1
);

-- working stop ref
insert into tl_stop_external_references(id,feed_version_id,target_feed_onestop_id,target_stop_id) values (
    (select s.id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'FTVL'),
    (select s.feed_version_id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'FTVL'),    
    'CT',
    '70041'
);

-- broken stop ref
insert into tl_stop_external_references(id,feed_version_id,target_feed_onestop_id,target_stop_id) values (
    (select s.id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'POWL'),
    (select s.feed_version_id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'POWL'),
    'CT',
    'missing'
);

-- stop obs 1
insert into ext_performance_stop_observations(id,feed_version_id,source,trip_start_date,from_stop_id,to_stop_id,trip_id,route_id,observed_arrival_time,observed_departure_time) values (
    (select s.id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'FTVL'),
    (select s.feed_version_id from gtfs_stops s join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and stop_id = 'FTVL'),
    'TripUpdate',
    '2023-03-09'::date,
    'LAKE',
    'FTVL',
    'test',
    '03',
    36000,
    36010
);

-- unactivate feed
 update feed_states set feed_version_id = null where feed_id = (select id from current_feeds where onestop_id = 'EX');


insert into tl_tenants(tenant_name) values ('tl-tenant');
insert into tl_tenants(tenant_name) values ('restricted-tenant');
insert into tl_tenants(tenant_name) values ('all-users-tenant');

insert into tl_groups(group_name) values ('CT-group');
insert into tl_groups(group_name) values ('BA-group');
insert into tl_groups(group_name) values ('HA-group');
insert into tl_groups(group_name) values ('EX-group');
insert into tl_groups(group_name) values ('test-group');
