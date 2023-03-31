package meters

import (
	"sync"

	"github.com/interline-io/transitland-lib/log"
)

type DefaultMeter struct {
	values map[string]map[string]float64
	lock   sync.Mutex
}

func NewDefaultMeter() *DefaultMeter {
	return &DefaultMeter{
		values: map[string]map[string]float64{},
	}
}

func (m *DefaultMeter) Flush() error {
	return nil
}

func (m *DefaultMeter) Close() error {
	return nil
}

func (m *DefaultMeter) NewMeter(user MeterUser) ApiMeter {
	return &defaultUserMeter{
		user: user,
		mp:   m,
	}
}

func (m *DefaultMeter) sendMeter(u MeterUser, meterName string, value float64, dims map[string]string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[meterName]
	if !ok {
		a = map[string]float64{}
		m.values[meterName] = a
	}
	a[u.Name()] += value
	log.Trace().
		Str("user", u.Name()).
		Str("meter", meterName).
		Float64("meter_value", value).
		Float64("total_value", a[u.Name()]).
		Msg("meter")
	return nil
}

func (m *DefaultMeter) getValue(u MeterUser, meterName string) (float64, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[meterName]
	if !ok {
		a = map[string]float64{}
		m.values[meterName] = a
	}
	v, ok := a[u.Name()]
	return v, ok
}

type defaultUserMeter struct {
	user MeterUser
	dims []string
	mp   *DefaultMeter
}

func (m *defaultUserMeter) Meter(meterName string, value float64, extraDimensions map[string]string) error {
	var dm2 map[string]string
	if len(extraDimensions) > 0 || len(m.dims) > 0 {
		dm2 = map[string]string{}
	}
	for i := 0; i < len(m.dims); i += 3 {
		a := m.dims[i]
		k := m.dims[i+1]
		v := m.dims[i+2]
		if a == meterName {
			dm2[k] = v
		}
	}
	for k, v := range extraDimensions {
		dm2[k] = v
	}
	return m.mp.sendMeter(m.user, meterName, value, dm2)
}

func (m *defaultUserMeter) AddDimension(meterName string, key string, value string) {
	m.dims = append(m.dims, meterName, key, value)
}

func (m *defaultUserMeter) GetValue(meterName string) (float64, bool) {
	return m.mp.getValue(m.user, meterName)
}
