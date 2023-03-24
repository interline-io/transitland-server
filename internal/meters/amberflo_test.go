package meters

import (
	"os"
	"testing"
	"time"
)

type amberfloTestUser struct {
	name string
}

func (u *amberfloTestUser) Name() string {
	return u.name
}

func (u *amberfloTestUser) GetExternalID(eid string) (string, bool) {
	// must match key given in config below
	if eid == "amberflo" {
		return u.name, true
	}
	return "", false
}

func TestAmberFloMeter(t *testing.T) {
	checkKeys := []string{
		"TL_TEST_AMBERFLO_APIKEY",
		"TL_TEST_AMBERFLO_METER1",
		"TL_TEST_AMBERFLO_METER2",
		"TL_TEST_AMBERFLO_USER1",
		"TL_TEST_AMBERFLO_USER2",
		"TL_TEST_AMBERFLO_USER3",
	}
	for _, k := range checkKeys {
		v := os.Getenv(k)
		if v == "" {
			t.Skipf("key '%s' not set, skipping", k)
			return
		}
	}
	testConfig := testMeterConfig{
		testMeter1: os.Getenv("TL_TEST_AMBERFLO_METER1"),
		testMeter2: os.Getenv("TL_TEST_AMBERFLO_METER2"),
		user1:      &amberfloTestUser{name: os.Getenv("TL_TEST_AMBERFLO_USER1")},
		user2:      &amberfloTestUser{name: os.Getenv("TL_TEST_AMBERFLO_USER2")},
		user3:      &amberfloTestUser{name: os.Getenv("TL_TEST_AMBERFLO_USER3")},
	}
	mp := NewAmberFlo(os.Getenv("TL_TEST_AMBERFLO_APIKEY"), 1*time.Second, 1)
	mp.cfgs[testConfig.testMeter1] = amberFloConfig{Name: testConfig.testMeter1, ExternalIDKey: "amberflo"}
	mp.cfgs[testConfig.testMeter2] = amberFloConfig{Name: testConfig.testMeter2, ExternalIDKey: "amberflo"}
	testMeter(t, mp, testConfig)
}
