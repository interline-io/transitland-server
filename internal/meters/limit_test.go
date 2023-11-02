package meters

import (
	"testing"
)

func TestLimitMeter(t *testing.T) {
	mp := NewDefaultMeterProvider()
	cmp := NewLimitMeterProvider(mp)
	testConfig := testMeterConfig{
		testMeter1: "test1",
		testMeter2: "test2",
		user1:      &testUser{name: "test1"},
		user2:      &testUser{name: "test2"},
		user3:      &testUser{name: "test3"},
	}
	testMeter(t, cmp, testConfig)
}
