import os
import csv
import sys
import collections
import json
import psycopg2
import psycopg2.extras

QUERY = """
with bt as (
    SELECT TIMESTAMP %s AT TIME ZONE %s as ts
)
select
    t.trip_id,
    r.route_id,
    s.stop_id,
    st.stop_sequence,
    st.arrival_time,
    st.departure_time,
    extract(epoch from bt.ts + make_interval(secs:=st.arrival_time::float))::integer + %s as rt_arrival,
    extract(epoch from bt.ts + make_interval(secs:=st.departure_time::float))::integer + %s as rt_departure
from gtfs_trips t
join bt on true
join gtfs_routes r on r.id = t.route_id
join gtfs_stop_times st on st.trip_id = t.id
join gtfs_stops s on s.id = st.stop_id
where t.trip_id = ANY (%s);
"""

BASETIME = "2018-05-30 00:00:00"
TIMEZONE = "America/Los_Angeles"
DELAY = 30
UNCERTAINTY = 30
TRIPID = [
    "261",
    "263",
    "375",
    "365",
    "277",
    "267",
    "371",
    "269",
    "273"
]
# TRIPID = [
#     "2291633WKDY",
#     "2271618WKDY",
#     "2251603WKDY",
#     "2231548WKDY",
#     "2211533WKDY",
#     "2351718WKDY",
#     "2331703WKDY",
#     "2311648WKDY",
#     "1031645WKDY",
#     "1151545WKDY",
#     "1171600WKDY",
#     "1191615WKDY",
#     "1011630WKDY",
#     "1071715WKDY",
#     "1051700WKDY",
#     "1131530WKDY",
#     "5051637WKDY",
#     "5031622WKDY",
#     "5011607WKDY",
#     "5191552WKDY",
#     "5171537WKDY",
#     "5011728WKDY",
#     "5191713WKDY",
#     "5171658WKDY",
#     "5071543WKDY",
#     "5091558WKDY",
#     "5111613WKDY",
#     "5131628WKDY",
#     "5151643WKDY",
#     "5071652WKDY",
#     "5091707WKDY",
#     "5111722WKDY",
#     "2311535WKDY",
#     "2331550WKDY",
#     "2351605WKDY",
#     "2371620WKDY",
#     "2391635WKDY",
#     "1171712WKDY",
#     "1111627WKDY",
#     "1131642WKDY",
#     "1151657WKDY",
#     "1031527WKDY",
#     "1051542WKDY",
#     "1071557WKDY",
#     "1091612WKDY",
#     "2241705WKDY",
#     "2221650WKDY",
#     "2261720WKDY",
# ]


by_trip = collections.defaultdict(list)
min_ts = None
with psycopg2.connect(os.environ['TL_DATABASE_URL']) as conn:
  with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
    cur.execute(QUERY, (BASETIME, TIMEZONE, DELAY, DELAY, TRIPID,))
    for row in cur.fetchall():
        by_trip[row.get('trip_id')].append(dict(row))
        rta = row.get("rt_arrival", 0)
        if min_ts is None:
            min_ts = rta
        if row.get("rt_arrival", 0) < min_ts:
            min_ts = rta



rt_msg = {
    "header": {
        "gtfs_realtime_version": "1.0",
        "incrementality": 0,
        "timestamp": min_ts
    },
    "entity": []
}

for k,v in by_trip.items():
    if len(v) == 0:
        continue
    tv = v[0]
    tu = {
        "trip": {
            "trip_id": tv.get("trip_id"),
            "route_id": tv.get("route_id"),
            "schedule_relationship": 0
        },
     }
    sts = sorted(v, key=lambda x:x.get('stop_sequence', 0))
    stu = []
    for st in sts:
        stu.append({
            "stop_sequence": st.get("stop_sequence"),
            "stop_id": st.get("stop_id"),
            "arrival": {
                "delay": DELAY,
                "uncertainty": UNCERTAINTY,
                "time": st.get("rt_arrival")
            },
            "departure": {
                "delay": DELAY,
                "uncertainty": UNCERTAINTY,
                "time": st.get("rt_departure")
            }
        })
    tu["stop_time_update"] = stu
    rt_msg["entity"].append({
        "id": tv.get("trip_id"),
        "trip_update": tu
    })

print(json.dumps(rt_msg))