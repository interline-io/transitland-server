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
	Meter(MeterEvent) error
	Close() error
}

type MeterEvent struct {
	UserID     string
	MeterName  string
	MeterValue float64
	MeterTime  int64
	Dimensions map[string]string
}

type MeterProvider interface {
	NewMeter(string) ApiMeter
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
			r = r.WithContext(context.WithValue(r.Context(), meterCtxKey, apiMeter))
			next.ServeHTTP(w, r)
			m := MeterEvent{
				MeterValue: 1.0,
			}
			if user := auth.ForContext(r.Context()); user != nil {
				m.UserID = user.Name
			}
			if err := apiMeter.Meter(m); err != nil {
				log.Error().Err(err).Msg("amberflo error")
			}
		})
	}
}

func ForContext(ctx context.Context) ApiMeter {
	raw, _ := ctx.Value(meterCtxKey).(ApiMeter)
	return raw
}
