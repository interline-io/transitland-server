package meters

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitMeter(t *testing.T) {
	meterName := "testmeter"
	user := testUser{name: "testuser"}
	mp := NewDefaultMeterProvider()
	cmp := NewLimitMeterProvider(mp)
	cmp.Enabled = true
	testLimitMeter(t, cmp, meterName, user)
}

func TestLimitMeter_Amberflo(t *testing.T) {
	mp, testConfig, err := getTestAmberFloMeter()
	if err != nil {
		t.Skip(err.Error())
		return
	}
	cmp := NewLimitMeterProvider(mp)
	cmp.Enabled = true
	testLimitMeter(t, cmp, testConfig.testMeter1, testUser{name: testConfig.user1.ID()})
}

func testLimitMeter(t *testing.T, cmp *LimitMeterProvider, meterName string, user testUser) {
	m := cmp.NewMeter(user)
	testKey := 1 // time.Now().In(time.UTC).Unix()
	lims := []userMeterLimit{
		// foo tests
		{
			MeterName: meterName,
			Period:    "hour",
			Limit:     5.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "day",
			Limit:     8.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "month",
			Limit:     11.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		// bar tests
		{
			MeterName: meterName,
			Period:    "hour",
			Limit:     14.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "day",
			Limit:     17.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "month",
			Limit:     20.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
	}

	incr := 3.0
	for _, lim := range lims {
		t.Run(fmt.Sprintf("%v", lim), func(t *testing.T) {
			startTime, endTime := lim.Span()
			base, _ := m.GetValue(meterName, startTime, endTime, lim.Dims)
			lim.Limit += base
			cmp.UserLimits[user.name] = []userMeterLimit{lim}

			successCount := 0.0
			for i := 0; i < 10; i++ {
				err := m.Meter(meterName, incr, lim.Dims)
				if err == nil {
					successCount += 1
				}
				cmp.Flush()
			}
			expectCount := math.Floor((lim.Limit - base) / incr)
			// fmt.Println("successCount:", successCount, "expectCount:", expectCount)
			assert.Equal(t, expectCount, successCount)
			total, _ := m.GetValue(meterName, startTime, endTime, lim.Dims)
			total = total - base
			expectTotal := successCount * incr
			// fmt.Println("total:", total, "expectTotal:", expectTotal)
			assert.Equal(t, expectTotal, total)
		})
	}

}
