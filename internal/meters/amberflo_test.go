package meters

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/testutil"
)

type amberfloTestUser struct {
	name string
}

func (u *amberfloTestUser) ID() string {
	return u.name
}

func (u *amberfloTestUser) GetExternalData(eid string) (string, bool) {
	// must match key given in config below
	if eid == "amberflo" {
		return u.name, true
	}
	return "", false
}

func TestAmberFloMeter(t *testing.T) {
	mp, testConfig, err := getTestAmberFloMeter()
	if err != nil {
		t.Skip(err.Error())
		return
	}
	testMeter(t, mp, testConfig)
}

func getTestAmberFloMeter() (*AmberFlo, testMeterConfig, error) {
	checkKeys := []string{
		"TL_TEST_AMBERFLO_APIKEY",
		"TL_TEST_AMBERFLO_METER1",
		"TL_TEST_AMBERFLO_METER2",
		"TL_TEST_AMBERFLO_USER1",
		"TL_TEST_AMBERFLO_USER2",
		"TL_TEST_AMBERFLO_USER3",
	}
	for _, k := range checkKeys {
		_, a, ok := testutil.CheckEnv(k)
		if !ok {
			return nil, testMeterConfig{}, errors.New(a)
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
	return mp, testConfig, nil
}
