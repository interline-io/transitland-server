package meters

import (
	"context"
	"net/http"

	"github.com/interline-io/transitland-server/auth"
)

var meterCtxKey = &contextKey{"apiMeter"}

type contextKey struct {
	name string
}

type ApiMeter interface {
	Meter(string, float64, map[string]string) error
	AddDimension(string, string, string)
	GetValue(string) (float64, bool)
}

type MeterProvider interface {
	NewMeter(MeterUser) ApiMeter
	Close() error
	Flush() error
}

type MeterUser interface {
	Name() string
	GetExternalID(string) (string, bool)
}

func NewHttpMiddleware(apiMeter MeterProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make ctxMeter available in context
			ctx := r.Context()
			ctxMeter := apiMeter.NewMeter(auth.ForContext(ctx))
			r = r.WithContext(context.WithValue(ctx, meterCtxKey, ctxMeter))
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) ApiMeter {
	raw, _ := ctx.Value(meterCtxKey).(ApiMeter)
	return raw
}
