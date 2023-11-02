package meters

import (
	"context"
	"net/http"
	"time"

	"github.com/interline-io/transitland-server/auth/authn"
)

var meterCtxKey = struct{ name string }{"apiMeter"}

type Dimensions map[string]string

type ApiMeter interface {
	Meter(string, float64, Dimensions) error
	AddDimension(string, string, string)
	GetValue(string, time.Duration, Dimensions) (float64, bool)
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
			next.ServeHTTP(w, r)
			ctxMeter.Meter(meterName, meterValue, dims)
		})
	}
}

func ForContext(ctx context.Context) ApiMeter {
	raw, _ := ctx.Value(meterCtxKey).(ApiMeter)
	return raw
}
