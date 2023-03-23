package meters

import (
	"fmt"
	"sync"

	"github.com/interline-io/transitland-server/auth"
)

type DefaultMeter struct {
	meterName string
	values    map[string]int
	lock      sync.Mutex
}

func NewDefaultMeter() *DefaultMeter {
	return &DefaultMeter{}
}

func (m *DefaultMeter) Meter(u auth.User, value float64, dims map[string]string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.values[m.meterName] += 1
	fmt.Printf(
		"meter '%s': new val: %d\n",
		m.meterName,
		m.values[m.meterName],
	)
	return nil
}

func (m *DefaultMeter) Close() error {
	return nil
}

func (m *DefaultMeter) NewMeter(meterName string) ApiMeter {
	return &DefaultMeter{
		meterName: meterName,
		values:    map[string]int{},
	}
}
