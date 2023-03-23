package meters

import (
	"context"
	"net/http"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

var meterCtxKey = &contextKey{"apiMeter"}

type contextKey struct {
	name string
}

type ApiMeter interface {
	Meter(MeterUser, float64, map[string]string) error
	GetValue(MeterUser) (float64, bool)
}

type MeterProvider interface {
	NewMeter(string) ApiMeter
	Close() error
	Flush() error
}

type MeterUser interface {
	Name() string
	GetExternalID(string) (string, bool)
}

func NewHttpMiddleware(apiMeter ApiMeter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Make ApiMeter available in context
			r = r.WithContext(context.WithValue(ctx, meterCtxKey, apiMeter))
			// Wrap
			next.ServeHTTP(w, r)
			// On successful HTTP, log event
			user := auth.ForContext(ctx)
			if err := apiMeter.Meter(user, 1.0, nil); err != nil {
				log.Error().Err(err).Msg("metering error")
			}
		})
	}
}

func ForContext(ctx context.Context) ApiMeter {
	raw, _ := ctx.Value(meterCtxKey).(ApiMeter)
	return raw
}
