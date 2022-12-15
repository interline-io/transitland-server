#!/bin/bash
# export TL_LOG=debug
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# older data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch -feed-url=test/data/external/bart-old.zip BA # old data
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" 
# current data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch 
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL"  -activate
# sync again
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
