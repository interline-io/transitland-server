package meters

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimitMeter(t *testing.T) {
	meterName := "testmeter"
	user := testUser{name: "testuser"}
	mp := NewDefaultMeterProvider()
	cmp := NewLimitMeterProvider(mp)
	testLimitMeter(t, cmp, meterName, user)
}

func TestLimitMeter_Amberflo(t *testing.T) {
	mp, testConfig, err := getTestAmberFloMeter()
	if err != nil {
		t.Skip(err.Error())
		return
	}
	cmp := NewLimitMeterProvider(mp)
	testLimitMeter(t, cmp, testConfig.testMeter1, testUser{name: testConfig.user1.ID()})
}

func testLimitMeter(t *testing.T, cmp *LimitMeterProvider, meterName string, user testUser) {
	m := cmp.NewMeter(user)
	testDims1 := Dimensions{{Key: "ok", Value: "test"}}
	testDims2 := Dimensions{{Key: "ok", Value: "bar"}}

	lim1 := 10.0
	lim2 := 11.0
	incr := 3.0

	cmp.UserLimits[user.name] = append(cmp.UserLimits[user.name],
		userMeterLimit{
			MeterName: meterName,
			Period:    "month",
			Limit:     lim1,
			Dims:      testDims1,
		},
		userMeterLimit{
			MeterName: meterName,
			Period:    "day",
			Limit:     lim2,
			Dims:      testDims2,
		},
	)

	// 1
	successCount1 := 0.0
	for i := 0; i < 10; i++ {
		err := m.Meter(meterName, incr, testDims1)
		if err == nil {
			successCount1 += 1
		}
		cmp.Flush()
	}
	assert.Equal(t, successCount1, math.Floor(lim1/incr))

	// 2
	successCount2 := 0.0
	for i := 0; i < 10; i++ {
		err := m.Meter(meterName, incr, testDims2)
		if err == nil {
			successCount2 += 1
		}
		cmp.Flush()
	}
	assert.Equal(t, successCount2, math.Floor(lim2/incr))

	// total 1
	v1, _ := m.GetValue(meterName, time.Unix(0, 0), time.Now(), testDims1)
	assert.Equal(t, successCount1*incr, v1)

	// total 2
	v2, _ := m.GetValue(meterName, time.Unix(0, 0), time.Now(), testDims2)
	assert.Equal(t, successCount2*incr, v2)

}
