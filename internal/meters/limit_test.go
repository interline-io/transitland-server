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
	m := cmp.NewMeter(user)
	testDims := Dimensions{{Key: "ok", Value: "test"}}
	testDims2 := Dimensions{{Key: "ok", Value: "bar"}}
	cmp.UserLimits[user.name] = append(cmp.UserLimits[user.name],
		userMeterLimit{
			MeterName: meterName,
			Period:    "month",
			Limit:     100.0,
			Dims:      testDims,
		},
		userMeterLimit{
			MeterName: meterName,
			Period:    "day",
			Limit:     500.0,
			Dims:      testDims2,
		},
	)

	successCount1 := 0.0
	for i := 0; i < 10; i++ {
		err := m.Meter(meterName, 15.0, testDims)
		if err == nil {
			successCount1 += 1
		}
	}
	assert.Equal(t, successCount1, math.Floor(100.0/15.0))

	successCount2 := 0.0
	for i := 0; i < 50; i++ {
		err := m.Meter(meterName, 15.0, testDims2)
		if err == nil {
			successCount2 += 1
		}
	}
	assert.Equal(t, successCount2, math.Floor(500.0/15.0))

	v1, _ := m.GetValue(meterName, time.Unix(0, 0), time.Now(), testDims)
	assert.Equal(t, successCount1*15.0, v1)

	v2, _ := m.GetValue(meterName, time.Unix(0, 0), time.Now(), testDims2)
	assert.Equal(t, successCount2*15.0, v2)

}
