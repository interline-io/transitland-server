package meters

import (
	"errors"
	"time"
)

func init() {
	var _ MeterProvider = &LimitMeterProvider{}
}

type userMeterLimit struct {
	User      string
	MeterName string
	Dims      Dimensions
	Period    string
	Limit     float64
}

type LimitMeterProvider struct {
	MeterProvider
	UserLimits map[string][]userMeterLimit
}

func NewLimitMeterProvider(provider MeterProvider) *LimitMeterProvider {
	return &LimitMeterProvider{
		MeterProvider: provider,
		UserLimits:    map[string][]userMeterLimit{},
	}
}

func (c *LimitMeterProvider) NewMeter(u MeterUser) ApiMeter {
	userData, _ := u.GetExternalData("gatekeeper")
	return &LimitMeter{
		userId:   u.ID(),
		userData: userData,
		provider: c,
		ApiMeter: c.MeterProvider.NewMeter(u),
	}
}

type LimitMeter struct {
	userId   string
	userData string
	provider *LimitMeterProvider
	ApiMeter
}

func (c *LimitMeter) GetLimit(meterName string, checkDims Dimensions) (time.Time, time.Time, float64, bool) {
	var lim userMeterLimit
	found := false
	for _, checkLim := range c.provider.UserLimits[c.userId] {
		if checkLim.MeterName == meterName && matchDims(checkDims, checkLim.Dims) {
			found = true
			lim = checkLim
			break
		}
	}
	if !found {
		return time.Now(), time.Now(), 0, false
	}
	now := time.Now()
	d1 := now
	d2 := now
	if lim.Period == "month" {
		d1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 1, 0)
	} else if lim.Period == "day" {
		d1 = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 0, 1)
	} else if lim.Period == "year" {
		d1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(1, 0, 0)
	} else if lim.Period == "total" {
		d1 = time.Unix(0, 0)
		d2 = time.Unix(1<<63-1, 0)
	} else {
		return time.Now(), time.Now(), 0, false
	}
	return d1, d2, lim.Limit, true
}

func (c *LimitMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	d1, d2, lim, foundLimit := c.GetLimit(meterName, extraDimensions)
	a, _ := c.ApiMeter.GetValue(meterName, d1, d2, extraDimensions)
	if foundLimit && a+value > lim {
		return errors.New("rate limited")
	}
	return c.ApiMeter.Meter(meterName, value, extraDimensions)
}
