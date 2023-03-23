package meters

import (
	"context"
	"net/http"
	"time"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

var meterCtxKey = &contextKey{"apiMeter"}

type contextKey struct {
	name string
}

type ApiMeter interface {
	Meter(auth.User, float64, map[string]string) error
}

type MeterProvider interface {
	NewMeter(string) ApiMeter
	Close() error
}

type MeterEvent struct {
	UserID     string
	MeterName  string
	MeterValue float64
	MeterTime  int64
	Dimensions map[string]string
}

func NewEvent() MeterEvent {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	return MeterEvent{
		MeterTime: t,
	}
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
