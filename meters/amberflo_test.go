package meters

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/testutil"
)

func TestAmberfloMeter(t *testing.T) {
	mp, testConfig, err := getTestAmberfloMeter()
	if err != nil {
		t.Skip(err.Error())
		return
	}
	testMeter(t, mp, testConfig)
}

func getTestAmberfloMeter() (*AmberfloMeterProvider, testMeterConfig, error) {
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
	eidKey := "amberflo"
	testConfig := testMeterConfig{
		testMeter1: os.Getenv("TL_TEST_AMBERFLO_METER1"),
		testMeter2: os.Getenv("TL_TEST_AMBERFLO_METER2"),
		user1: &testUser{
			name: os.Getenv("TL_TEST_AMBERFLO_USER1"),
			data: map[string]string{eidKey: os.Getenv("TL_TEST_AMBERFLO_USER1")},
		},
		user2: &testUser{
			name: os.Getenv("TL_TEST_AMBERFLO_USER2"),
			data: map[string]string{eidKey: os.Getenv("TL_TEST_AMBERFLO_USER2")},
		},
		user3: &testUser{
			name: os.Getenv("TL_TEST_AMBERFLO_USER3"),
			data: map[string]string{eidKey: os.Getenv("TL_TEST_AMBERFLO_USER3")},
		},
	}
	mp := NewAmberfloMeterProvider(os.Getenv("TL_TEST_AMBERFLO_APIKEY"), 1*time.Second, 1)
	mp.cfgs[testConfig.testMeter1] = amberFloConfig{Name: testConfig.testMeter1, ExternalIDKey: eidKey}
	mp.cfgs[testConfig.testMeter2] = amberFloConfig{Name: testConfig.testMeter2, ExternalIDKey: eidKey}
	return mp, testConfig, nil
}
