package meters

import (
	"fmt"
	"sync"
)

type DefaultMeter struct {
	handlerName string
	values      map[string]int
	lock        sync.Mutex
}

func NewDefaultMeter() *DefaultMeter {
	return &DefaultMeter{}
}

func (m *DefaultMeter) Meter(e MeterEvent) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.values[e.MeterName] += 1
	fmt.Printf(
		"meter '%s' handling meter event: %#v new val: %d\n",
		m.handlerName,
		e,
		m.values[e.MeterName],
	)
	return nil
}

func (m *DefaultMeter) Close() error {
	return nil
}

func (m *DefaultMeter) NewMeter(handlerName string) ApiMeter {
	return &DefaultMeter{
		handlerName: handlerName,
		values:      map[string]int{},
	}
}
