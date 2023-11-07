package meters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitMeter(t *testing.T) {
	meterName := "testmeter"
	user := testUser{name: "testuser"}
	// cmp.DefaultLimits = testLims(meterName)
	// for _, lim := range cmp.DefaultLimits {
	for _, lim := range testLims(meterName) {
		t.Run("", func(t *testing.T) {
			mp := NewDefaultMeterProvider()
			cmp := NewLimitMeterProvider(mp)
			cmp.Enabled = true
			cmp.DefaultLimits = []UserMeterLimit{lim}
			testLimitMeter(t,
				cmp,
				lim.MeterName,
				user,
				lim,
			)
		})
	}
}

func TestLimitMeter_Amberflo(t *testing.T) {
	mp, testConfig, err := getTestAmberfloMeter()
	if err != nil {
		t.Skip(err.Error())
		return
	}
	user := testUser{
		name: testConfig.user1.ID(),
		data: map[string]string{"amberflo": "amberflo"},
	}
	for _, lim := range testLims(testConfig.testMeter1) {
		t.Run("", func(t *testing.T) {
			cmp := NewLimitMeterProvider(mp)
			cmp.Enabled = true
			cmp.DefaultLimits = []UserMeterLimit{lim}
			testLimitMeter(t,
				cmp,
				lim.MeterName,
				user,
				lim,
			)
		})
	}
}

func TestLimitMeter_Gatekeeper(t *testing.T) {
	// JSON blob
	gkData := `	
	{
		"product_limits": {
			"tlv2_api": [
				{
					"amberflo_dimension": "fv",
					"amberflo_dimension_value": true,
					"amberflo_meter": "testmeter",
					"limit_value": 100,
					"time_period": "monthly"
				},
				{
					"amberflo_dimension": "fv",
					"amberflo_dimension_value": false,
					"amberflo_meter": "testmeter",
					"limit_value": 500,
					"time_period": "monthly"
				}
			]
		},
	}`
	user := testUser{name: "testuser"}
	user.data = map[string]string{"gatekeeper": gkData}
	lims := parseGkUserLimits(gkData)
	for _, lim := range lims {
		t.Run("", func(t *testing.T) {
			mp := NewDefaultMeterProvider()
			cmp := NewLimitMeterProvider(mp)
			cmp.Enabled = true
			testLimitMeter(t,
				cmp,
				lim.MeterName,
				user,
				lim,
			)
		})
	}
}

func testLims(meterName string) []UserMeterLimit {
	testKey := 1 // time.Now().In(time.UTC).Unix()
	lims := []UserMeterLimit{
		// foo tests
		{
			MeterName: meterName,
			Period:    "hourly",
			Limit:     50.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "daily",
			Limit:     80.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "monthly",
			Limit:     110.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("foo:%d", testKey)}},
		},
		// bar tests
		{
			MeterName: meterName,
			Period:    "hourly",
			Limit:     140.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "daily",
			Limit:     170.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
		{
			MeterName: meterName,
			Period:    "monthly",
			Limit:     200.0,
			Dims:      Dimensions{{Key: "ok", Value: fmt.Sprintf("bar:%d", testKey)}},
		},
	}
	return lims
}

func testLimitMeter(t *testing.T, cmp *LimitMeterProvider, meterName string, user testUser, lim UserMeterLimit) {
	incr := 1.0
	m := cmp.NewMeter(user)
	startTime, endTime := lim.Span()
	base, _ := m.GetValue(meterName, startTime, endTime, lim.Dims)

	// Probably ok
	if err := m.Meter(meterName, incr, lim.Dims); err != nil {
		t.Error(err)
	}
	cmp.MeterProvider.Flush()

	// push past limit
	if err := m.Meter(meterName, incr+lim.Limit, lim.Dims); err == nil {
		t.Error("expected error, got none")
	}

	// Check updated value
	total, _ := m.GetValue(meterName, startTime, endTime, lim.Dims)
	assert.Equal(t, base+incr, total, "expected total")
}
