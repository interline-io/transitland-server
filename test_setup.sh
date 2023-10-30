#!/bin/bash
# Remove import files
export TL_TEST_STORAGE="${PWD}/tmp"
mkdir -p "${TL_TEST_STORAGE}"; rm ${TL_TEST_STORAGE}/*.zip
# export TL_LOG=debug
(cd cmd/tlserver && go install .)
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# older data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -storage="$TL_TEST_STORAGE" -allow-local-fetch -feed-url=test/data/external/bart-old.zip BA # old data
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" -storage="$TL_TEST_STORAGE" 
# current data
tlserver fetch -dburl="$TL_TEST_SERVER_DATABASE_URL" -storage="$TL_TEST_STORAGE" -allow-local-fetch 
tlserver import -dburl="$TL_TEST_SERVER_DATABASE_URL" -storage="$TL_TEST_STORAGE" -activate
# sync again
tlserver sync -dburl="$TL_TEST_SERVER_DATABASE_URL" test/data/server/server-test.dmfr.json
# supplemental data
psql -f test_supplement.pgsql
