package meters

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func init() {
	var _ MeterProvider = &LimitMeterProvider{}
}

type LimitMeterProvider struct {
	Enabled    bool
	UserLimits map[string][]userMeterLimit
	MeterProvider
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

func (c *LimitMeter) GetLimit(meterName string, checkDims Dimensions) (userMeterLimit, bool) {
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
		return lim, false
	}
	return lim, true
}

func (c *LimitMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	lim, foundLimit := c.GetLimit(meterName, extraDimensions)
	d1, d2 := lim.Span()
	if c.provider.Enabled && foundLimit {
		currentValue, _ := c.GetValue(meterName, d1, d2, extraDimensions)
		if foundLimit && currentValue+value > lim.Limit {
			log.Info().Str("meter", meterName).Str("user", c.userId).Float64("current", currentValue).Float64("add", value).Str("dims", fmt.Sprintf("%v", extraDimensions)).Msg("rate limited")
			return errors.New("rate limited")
		} else {
			log.Info().Str("meter", meterName).Str("user", c.userId).Float64("current", currentValue).Float64("add", value).Str("dims", fmt.Sprintf("%v", extraDimensions)).Msg("rate check")
		}
	}
	return c.ApiMeter.Meter(meterName, value, extraDimensions)
}

type userMeterLimit struct {
	User      string
	MeterName string
	Dims      Dimensions
	Period    string
	Limit     float64
}

func (lim *userMeterLimit) Span() (time.Time, time.Time) {
	now := time.Now().In(time.UTC)
	d1 := now
	d2 := now
	if lim.Period == "hour" {
		d1 = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		d2 = d1.Add(3600 * time.Second)
	} else if lim.Period == "day" {
		d1 = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 0, 1)
	} else if lim.Period == "month" {
		d1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 1, 0)
	} else if lim.Period == "year" {
		d1 = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(1, 0, 0)
	} else if lim.Period == "total" {
		d1 = time.Unix(0, 0)
		d2 = time.Unix(1<<63-1, 0)
	} else {
		return now, now
	}
	return d1, d2
}
