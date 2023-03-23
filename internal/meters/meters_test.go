package meters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	name string
}

func (u testUser) Name() string {
	return u.name
}

func (u testUser) GetExternalID(string) (string, bool) {
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
	t.Run("Meter", func(t *testing.T) {
		m := mp.NewMeter(cfg.testMeter1)
		v, _ := m.GetValue(cfg.user1)

		m.Meter(cfg.user1, 1, nil)
		mp.Flush()

		a, _ := m.GetValue(cfg.user1)
		assert.Equal(t, 1.0, a-v)

		m.Meter(cfg.user1, 1, nil)
		mp.Flush()

		b, _ := m.GetValue(cfg.user1)
		assert.Equal(t, 2.0, b-v)
	})
	t.Run("NewMeter", func(t *testing.T) {
		m1 := mp.NewMeter(cfg.testMeter1)
		m2 := mp.NewMeter(cfg.testMeter2)

		v1, _ := m1.GetValue(cfg.user1)
		v2, _ := m2.GetValue(cfg.user1)

		m1.Meter(cfg.user1, 1, nil)
		m2.Meter(cfg.user1, 2, nil)
		mp.Flush()

		a, _ := m1.GetValue(cfg.user1)
		assert.Equal(t, 1.0, a-v1)
		b, _ := m2.GetValue(cfg.user1)
		assert.Equal(t, 2.0, b-v2)
	})
	t.Run("GetValue", func(t *testing.T) {
		m := mp.NewMeter(cfg.testMeter1)
		v1, _ := m.GetValue(cfg.user1)
		v2, _ := m.GetValue(cfg.user2)
		v3, _ := m.GetValue(cfg.user3)

		m.Meter(cfg.user1, 1, nil)
		m.Meter(cfg.user2, 2.0, nil)
		mp.Flush()

		a, ok := m.GetValue(cfg.user1)
		assert.Equal(t, 1.0, a-v1)
		assert.Equal(t, true, ok)

		a, ok = m.GetValue(cfg.user2)
		assert.Equal(t, 2.0, a-v2)
		assert.Equal(t, true, ok)

		a, ok = m.GetValue(cfg.user3)
		assert.Equal(t, 0.0, a-v3)
	})
}
