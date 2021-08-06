#!/bin/bash
tlserver dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
tlserver dmfr fetch -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR
tlserver dmfr import -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR -activate
tlserver dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
