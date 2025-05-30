#!/bin/sh
# Set up a new postgres database and import Natural Earth data.
SCRIPTDIR=$(dirname "$0")
TL_TEST_STORAGE=$(dirname "$0")/tmp
mkdir -p "${TL_TEST_STORAGE}"; rm ${TL_TEST_STORAGE}/*.zip

set -e

# Rebuild binary
(cd cmd/tlserver && go install .)

# Wait for database to accept connections
${SCRIPTDIR}/wait-for-it.sh "${PGHOST}:${PGPORT}"

# Fail if db already exists -- this is a bootstrap script after all!
createdb "${PGDATABASE}"

# Run migrations
tlserver dbmigrate --dburl="$TL_TEST_SERVER_DATABASE_URL" up

# Load Natural Earth data
tlserver dbmigrate --dburl="$TL_TEST_SERVER_DATABASE_URL" natural-earth

# Remove import files
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