#!/bin/bash
transitland dmfr sync -dburl=$TL_TEST_SERVER_DATABASE_URL test/data/server/server-test.dmfr.json
transitland dmfr fetch -dburl=$TL_TEST_SERVER_DATABASE_URL -gtfsdir=$TL_TEST_GTFSDIR
transitland dmfr import -dburl=$TL_TEST_SERVER_DATABASE_URL -gtfsdir=$TL_TEST_GTFSDIR -activate
