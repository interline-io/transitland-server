#!/bin/bash
# export TL_LOG=debug
TL_TEST_GTFSDIR="."
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# older data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch -gtfsdir="$TL_TEST_GTFSDIR" -feed-url=test/data/external/bart-old.zip BA # old data
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" -gtfsdir="$TL_TEST_GTFSDIR"
# current data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch -gtfsdir="$TL_TEST_GTFSDIR"
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" -gtfsdir="$TL_TEST_GTFSDIR" -activate
# sync again
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
