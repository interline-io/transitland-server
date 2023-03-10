insert into ext_plus_route_attributes(route_id,category,subcategory,running_way) values (
    (select r.id from gtfs_routes r join feed_states fs using(feed_version_id) join current_feeds cf on cf.id = fs.feed_id where cf.onestop_id = 'BA' and route_id = '01'),
    2,
    201,
    1
);