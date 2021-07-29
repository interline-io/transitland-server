#!/bin/bash
transitland_server dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
transitland_server dmfr fetch -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR
transitland_server dmfr import -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR -activate
transitland_server dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
