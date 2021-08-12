#!/bin/bash
tlserver sync -dburl=$1 test/data/server/server-test.dmfr.json
tlserver fetch -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR
tlserver import -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR -activate
tlserver sync -dburl=$1 test/data/server/server-test.dmfr.json
