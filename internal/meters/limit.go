package meters

import (
	"fmt"
	"time"
)

func init() {
	var _ MeterProvider = &LimitMeterProvider{}
}

type LimitMeterProvider struct {
	MeterProvider
}

func NewLimitMeterProvider(provider MeterProvider) *LimitMeterProvider {
	return &LimitMeterProvider{
		MeterProvider: provider,
	}
}

func (c *LimitMeterProvider) NewMeter(u MeterUser) ApiMeter {
	userData, _ := u.GetExternalData("gatekeeper")
	return &LimitMeter{
		userId:   u.ID(),
		userData: userData,
		ApiMeter: c.MeterProvider.NewMeter(u),
	}
}

type LimitMeter struct {
	userId   string
	userData string
	ApiMeter
}

func (c *LimitMeter) GetLimit(meterName string, dimension string) (time.Duration, float64, bool) {
	fmt.Println("GET LIMIT:", meterName, dimension)
	return time.Second, 123, true
}

func (c *LimitMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	dimension := "testdim"
	d, lim, ok := c.GetLimit(meterName, dimension)
	_ = ok
	fmt.Println("GOT LIMIT:", lim)
	a, ok := c.ApiMeter.GetValue(meterName, d, extraDimensions)
	fmt.Println("GET VALUE:", a)

	return c.ApiMeter.Meter(meterName, value, extraDimensions)
}
