package meters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	name string
}

func (u testUser) ID() string {
	return u.name
}

func (u testUser) GetExternalData(string) (string, bool) {
	return "test", true
}

type testMeterConfig struct {
	testMeter1 string
	testMeter2 string
	user1      MeterUser
	user2      MeterUser
	user3      MeterUser
}

func testMeter(t *testing.T, mp MeterProvider, cfg testMeterConfig) {
	dim := Dimensions{"test": "ok"}
	d := time.Duration(0)
	t.Run("Meter", func(t *testing.T) {
		m := mp.NewMeter(cfg.user1)
		v, _ := m.GetValue(cfg.testMeter1, d, nil)

		m.Meter(cfg.testMeter1, 1, nil)
		mp.Flush()

		a, _ := m.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 1.0, a-v)

		m.Meter(cfg.testMeter1, 1, nil)
		mp.Flush()

		b, _ := m.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 2.0, b-v)
	})
	t.Run("NewMeter", func(t *testing.T) {
		m1 := mp.NewMeter(cfg.user1)

		v1, _ := m1.GetValue(cfg.testMeter1, d, dim)
		v2, _ := m1.GetValue(cfg.testMeter2, d, dim)

		m1.Meter(cfg.testMeter1, 1, nil)
		m1.Meter(cfg.testMeter2, 2, nil)
		mp.Flush()

		va1, _ := m1.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 1.0, va1-v1)
		va2, _ := m1.GetValue(cfg.testMeter2, d, dim)
		assert.Equal(t, 2.0, va2-v2)
	})
	t.Run("GetValue", func(t *testing.T) {
		m1 := mp.NewMeter(cfg.user1)
		m2 := mp.NewMeter(cfg.user2)
		m3 := mp.NewMeter(cfg.user3)
		v1, _ := m1.GetValue(cfg.testMeter1, d, dim)
		v2, _ := m2.GetValue(cfg.testMeter1, d, dim)
		v3, _ := m3.GetValue(cfg.testMeter1, d, dim)

		m1.Meter(cfg.testMeter1, 1, nil)
		m2.Meter(cfg.testMeter1, 2.0, nil)
		mp.Flush()

		a, ok := m1.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 1.0, a-v1)
		assert.Equal(t, true, ok)

		a, ok = m2.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 2.0, a-v2)
		assert.Equal(t, true, ok)

		a, _ = m3.GetValue(cfg.testMeter1, d, dim)
		assert.Equal(t, 0.0, a-v3)
	})
}
