#!/bin/bash
# export TL_LOG=debug
TL_TEST_GTFSDIR="."
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch -gtfsdir="$TL_TEST_GTFSDIR"
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" -gtfsdir="$TL_TEST_GTFSDIR" -activate
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
