package meters

import (
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/transitland-lib/log"
	"github.com/tidwall/gjson"
)

func init() {
	var _ MeterProvider = &LimitMeterProvider{}
}

type LimitMeterProvider struct {
	Enabled       bool
	DefaultLimits []UserMeterLimit
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

func (c *LimitMeter) GetLimits(meterName string, checkDims Dimensions) []UserMeterLimit {
	// The limit matches the event dimensions if all of the LIMIT dimensions are contained in event
	var lims []UserMeterLimit
	for _, userLimit := range parseGkUserLimits(c.userData) {
		if userLimit.MeterName == meterName && dimsContainedIn(userLimit.Dims, checkDims) {
			lims = append(lims, userLimit)
		}
	}
	for _, defaultLimit := range c.provider.DefaultLimits {
		if defaultLimit.MeterName == meterName && dimsContainedIn(defaultLimit.Dims, checkDims) {
			lims = append(lims, defaultLimit)
		}
	}
	return lims
}

func (c *LimitMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	if c.provider.Enabled {
		for _, lim := range c.GetLimits(meterName, extraDimensions) {
			d1, d2 := lim.Span()
			currentValue, _ := c.GetValue(meterName, d1, d2, lim.Dims)
			if currentValue+value > lim.Limit {
				log.Info().Str("meter", meterName).Str("user", c.userId).Float64("limit", lim.Limit).Float64("current", currentValue).Float64("add", value).Str("dims", fmt.Sprintf("%v", lim.Dims)).Msg("rate limited")
				return errors.New("rate check: limited")
			} else {
				log.Info().Str("meter", meterName).Str("user", c.userId).Float64("limit", lim.Limit).Float64("current", currentValue).Float64("add", value).Str("dims", fmt.Sprintf("%v", lim.Dims)).Msg("rate check: ok")
			}
		}
	}
	return c.ApiMeter.Meter(meterName, value, extraDimensions)
}

type UserMeterLimit struct {
	User      string
	MeterName string
	Dims      Dimensions
	Period    string
	Limit     float64
}

func (lim *UserMeterLimit) Span() (time.Time, time.Time) {
	now := time.Now().In(time.UTC)
	d1 := now
	d2 := now
	if lim.Period == "hourly" {
		d1 = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		d2 = d1.Add(3600 * time.Second)
	} else if lim.Period == "daily" {
		d1 = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 0, 1)
	} else if lim.Period == "monthly" {
		d1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		d2 = d1.AddDate(0, 1, 0)
	} else if lim.Period == "yearly" {
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

func parseGkUserLimits(v string) []UserMeterLimit {
	var lims []UserMeterLimit
	for _, productLimit := range gjson.Get(v, "product_limits").Map() {
		for _, plim := range productLimit.Array() {
			lim := UserMeterLimit{
				MeterName: plim.Get("amberflo_meter").String(),
				Limit:     plim.Get("limit_value").Float(),
				Period:    plim.Get("time_period").String(),
			}
			if dim := plim.Get("amberflo_dimension").String(); dim != "" {
				lim.Dims = append(lim.Dims, Dimension{
					Key:   dim,
					Value: plim.Get("amberflo_dimension_value").String(),
				})
			}
			lims = append(lims, lim)
		}
	}
	return lims
}
