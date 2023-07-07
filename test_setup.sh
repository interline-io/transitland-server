#!/bin/bash
# Remove import files
rm *.zip
# export TL_LOG=debug
(cd cmd/tlserver && go install .)
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# older data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch -feed-url=test/data/external/bart-old.zip BA # old data
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" 
# current data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -allow-local-fetch 
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL"  -activate
# sync again
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# supplemental data
psql -f test_supplement.pgsql
rm *.zip
