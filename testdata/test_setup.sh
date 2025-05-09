#!/bin/bash
# Remove import files
TL_TEST_STORAGE=$(dirname "$0")/tmp
mkdir -p "${TL_TEST_STORAGE}"; rm ${TL_TEST_STORAGE}/*.zip
# export TL_LOG=debug
(cd cmd/tlserver && go install .)
tlserver sync --dburl="$TL_TEST_SERVER_DATABASE_URL" testdata/server/server-test.dmfr.json
# older data and forced error
tlserver fetch --dburl="$TL_TEST_SERVER_DATABASE_URL" --storage="$TL_TEST_STORAGE" --validation-report --validation-report-storage="$TL_TEST_STORAGE" --allow-local-fetch --feed-url=testdata/gtfs/bart-errors.zip BA # error data
tlserver fetch --dburl="$TL_TEST_SERVER_DATABASE_URL" --storage="$TL_TEST_STORAGE" --validation-report --validation-report-storage="$TL_TEST_STORAGE" --allow-local-fetch --feed-url=testdata/gtfs/bart-old.zip BA # old data
tlserver import --dburl="$TL_TEST_SERVER_DATABASE_URL" --storage="$TL_TEST_STORAGE" 
# current data
tlserver fetch --dburl="$TL_TEST_SERVER_DATABASE_URL" --storage="$TL_TEST_STORAGE" --validation-report --validation-report-storage="$TL_TEST_STORAGE" --allow-local-fetch 
tlserver import --dburl="$TL_TEST_SERVER_DATABASE_URL" --storage="$TL_TEST_STORAGE" --activate
# sync again
tlserver sync --dburl="$TL_TEST_SERVER_DATABASE_URL" testdata/server/server-test.dmfr.json
# supplemental data
psql $TL_TEST_SERVER_DATABASE_URL -f $(dirname "$0")/test_supplement.pgsql
# load census data
psql $TL_TEST_SERVER_DATABASE_URL -f $(dirname "$0")/census/census.pgsql