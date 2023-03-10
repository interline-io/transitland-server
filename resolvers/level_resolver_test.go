package resolvers

import (
	"testing"
)

func TestLevelResolver(t *testing.T) {
	te := newTestEnv(t)
	testcases := []testcase{
		// TODO: level by stop
		// TODO: stops by level
	}
	queryTestcases(t, te.client, testcases)
}
