package meters

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type DefaultMeter struct {
	meterName string
	values    map[string]float64
	lock      sync.Mutex
}

func NewDefaultMeter() *DefaultMeter {
	return &DefaultMeter{}
}

func (m *DefaultMeter) Meter(u MeterUser, value float64, dims map[string]string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.values[u.Name()] += value
	log.Trace().
		Str("user", u.Name()).
		Str("meter", m.meterName).
		Float64("meter_value", value).
		Float64("total_value", m.values[m.meterName]).
		Msg("meter")
	return nil
}

func (m *DefaultMeter) GetValue(u MeterUser) (float64, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[u.Name()]
	return a, ok
}

func (m *DefaultMeter) Flush() error {
	return nil
}

func (m *DefaultMeter) Close() error {
	return nil
}

func (m *DefaultMeter) NewMeter(meterName string) ApiMeter {
	return &DefaultMeter{
		meterName: meterName,
		values:    map[string]float64{},
	}
}
