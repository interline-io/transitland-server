package meters

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/log"
)

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

func (m *DefaultMeterProvider) sendMeter(u MeterUser, meterName string, value float64, dims []Dimension) error {
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
		time:  time.Now().In(time.UTC),
		dims:  dims,
	}
	a[userName] = append(a[userName], event)
	log.Trace().
		Str("user", userName).
		Str("meter", meterName).
		Float64("meter_value", value).
		Msg("meter")
	return nil
}

func (m *DefaultMeterProvider) getValue(u MeterUser, meterName string, startTime time.Time, endTime time.Time, checkDims Dimensions) (float64, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	a, ok := m.values[meterName]
	if !ok {
		return 0, false
	}
	total := 0.0
	for _, userEvent := range a[u.ID()] {
		match := true
		if userEvent.time.Equal(endTime) || userEvent.time.After(endTime) {
			// fmt.Println("not matched on end time", userEvent.time, endTime)
			match = false
		}
		if userEvent.time.Before(startTime) {
			// fmt.Println("not matched on start time", userEvent.time, startTime)
			match = false
		}
		if !dimsContainedIn(checkDims, userEvent.dims) {
			// fmt.Println("not matched on dims")
			match = false
		}
		if match {
			// fmt.Println("matched:", userEvent.value)
			total += userEvent.value
		}
	}
	return total, ok
}

type defaultUserMeter struct {
	user    MeterUser
	addDims []eventAddDim
	mp      *DefaultMeterProvider
}

func (m *defaultUserMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	// Copy in matching dimensions set through AddDimension
	var eventDims []Dimension
	for _, addDim := range m.addDims {
		if addDim.MeterName == meterName {
			eventDims = append(eventDims, Dimension{Key: addDim.Key, Value: addDim.Value})
		}
	}
	eventDims = append(eventDims, extraDimensions...)
	return m.mp.sendMeter(m.user, meterName, value, eventDims)
}

func (m *defaultUserMeter) AddDimension(meterName string, key string, value string) {
	m.addDims = append(m.addDims, eventAddDim{MeterName: meterName, Key: key, Value: value})
}

func (m *defaultUserMeter) GetValue(meterName string, startTime time.Time, endTime time.Time, dims Dimensions) (float64, bool) {
	return m.mp.getValue(m.user, meterName, startTime, endTime, dims)
}

///////////

type defaultMeterEvent struct {
	time  time.Time
	dims  []Dimension
	value float64
}

type defaultMeterUserEvents map[string][]defaultMeterEvent
