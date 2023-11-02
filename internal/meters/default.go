package meters

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/log"
)

type defaultMeterEvent struct {
	time  time.Time
	dims  []string
	value float64
}

type defaultMeterUserEvents map[string][]defaultMeterEvent

type DefaultMeterProvider struct {
	values map[string]defaultMeterUserEvents
	lock   sync.Mutex
}

func NewDefaultMeterProvider() *DefaultMeterProvider {
	return &DefaultMeterProvider{
		values: map[string]defaultMeterUserEvents{},
	}
}

func (m *DefaultMeterProvider) Flush() error {
	return nil
}

func (m *DefaultMeterProvider) Close() error {
	return nil
}

func (m *DefaultMeterProvider) NewMeter(user MeterUser) ApiMeter {
	return &defaultUserMeter{
		user: user,
		mp:   m,
	}
}

func (m *DefaultMeterProvider) sendMeter(u MeterUser, meterName string, value float64, dims map[string]string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[meterName]
	if !ok {
		a = defaultMeterUserEvents{}
		m.values[meterName] = a
	}
	userName := ""
	if u != nil {
		userName = u.ID()
	}
	event := defaultMeterEvent{
		value: value,
		time:  time.Now(),
	}
	for k, v := range dims {
		event.dims = append(event.dims, meterName, k, v)
	}
	a[userName] = append(a[userName], event)
	log.Trace().
		Str("user", userName).
		Str("meter", meterName).
		Float64("meter_value", value).
		Msg("meter")
	return nil
}

func (m *DefaultMeterProvider) getValue(u MeterUser, meterName string, d time.Duration, dims Dimensions) (float64, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[meterName]
	if !ok {
		a = map[string][]defaultMeterEvent{}
		m.values[meterName] = a
	}
	v, ok := a[u.ID()]

	total := 0.0
	minTime := time.Now().Add(-d)
	for _, e := range v {
		match := true
		if e.time.Before(minTime) {
			match = false
		}

		if match {
			total += e.value
		}
	}

	return total, ok
}

type defaultUserMeter struct {
	user MeterUser
	dims []string
	mp   *DefaultMeterProvider
}

func (m *defaultUserMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	var dm2 Dimensions
	if len(extraDimensions) > 0 || len(m.dims) > 0 {
		dm2 = Dimensions{}
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

func (m *defaultUserMeter) GetValue(meterName string, d time.Duration, dims Dimensions) (float64, bool) {
	return m.mp.getValue(m.user, meterName, d, dims)
}
