#!/bin/bash
# export TL_LOG=debug
tlserver sync test/data/server/server-test.dmfr.json
tlserver fetch -gtfsdir=$TL_TEST_GTFSDIR
tlserver import -gtfsdir=$TL_TEST_GTFSDIR -activate
tlserver sync test/data/server/server-test.dmfr.json
