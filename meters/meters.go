package meters

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/interline-io/transitland-server/auth/authn"
)

var meterCtxKey = struct{ name string }{"apiMeter"}

type ApiMeter interface {
	Meter(string, float64, Dimensions) error
	AddDimension(string, string, string)
	GetValue(string, time.Time, time.Time, Dimensions) (float64, bool)
}

type MeterProvider interface {
	NewMeter(MeterUser) ApiMeter
	Close() error
	Flush() error
}

type MeterUser interface {
	ID() string
	GetExternalData(string) (string, bool)
}

func WithMeter(apiMeter MeterProvider, meterName string, meterValue float64, dims Dimensions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make ctxMeter available in context
			ctx := r.Context()
			ctxMeter := apiMeter.NewMeter(authn.ForContext(ctx))
			r = r.WithContext(context.WithValue(ctx, meterCtxKey, ctxMeter))
			if err := ctxMeter.Meter(meterName, meterValue, dims); err != nil {
				http.Error(w, "429", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) ApiMeter {
	raw, _ := ctx.Value(meterCtxKey).(ApiMeter)
	return raw
}

type Dimension struct {
	Key   string
	Value string
}

type Dimensions []Dimension

type eventAddDim struct {
	MeterName string
	Key       string
	Value     string
}

func dimsContainedIn(checkDims Dimensions, eventDims Dimensions) bool {
	for _, matchDim := range checkDims {
		match := false
		for _, ed := range eventDims {
			if ed.Key == matchDim.Key && ed.Value == matchDim.Value {
				match = true
			}
		}
		if !match {
			return false
		}
	}
	return true
}

//////

type Config struct {
	EnableMetering         bool
	EnableRateLimits       bool
	MeteringProvider       string
	MeteringAmberfloConfig string
}

func GetProvider(cfg Config) (MeterProvider, error) {
	var meterProvider MeterProvider
	meterProvider = NewDefaultMeterProvider()
	if cfg.MeteringProvider == "amberflo" {
		a := NewAmberfloMeterProvider(os.Getenv("AMBERFLO_APIKEY"), 30*time.Second, 100)
		if cfg.MeteringAmberfloConfig != "" {
			if err := a.LoadConfig(cfg.MeteringAmberfloConfig); err != nil {
				return nil, err
			}
		}
		meterProvider = a
	}
	if cfg.EnableRateLimits {
		mp := NewLimitMeterProvider(meterProvider)
		mp.Enabled = true
		// mp.DefaultLimits = append(mp.DefaultLimits, meters.UserMeterLimit{Limit: 10, Period: "monthly", MeterName: "rest"})
		meterProvider = mp
	}
	return meterProvider, nil
}
