#!/bin/bash
transitland dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
transitland dmfr fetch -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR
transitland dmfr import -dburl=$1 -gtfsdir=$TL_TEST_GTFSDIR -activate
transitland dmfr sync -dburl=$1 test/data/server/server-test.dmfr.json
