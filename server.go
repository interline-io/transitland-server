package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting /query rewrite to /query/ to match subrouter's "/"
		if r.URL.Path == path {
			r.URL.Path = r.URL.Path + "/"
		}
		// Remove path prefix
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		handler.ServeHTTP(w, r)
	}))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		user := auth.ForContext(r.Context())
		if user == nil {
			user = &auth.User{IsAnon: true}
		}
		wr := wrapResponseWriter(w)
		next.ServeHTTP(wr, r)
		log.Info().
			Int64("duration_ms", (time.Now().UnixNano()-t1.UnixNano())/1e6).
			Str("method", r.Method).
			Str("path", r.URL.EscapedPath()).
			Str("query", r.URL.Query().Encode()).
			Str("user", user.Name).
			Int("status", wr.status).
			Msg("request")
	})
}

func getRedisOpts(v string) (*redis.Options, error) {
	a, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	if a.Scheme != "redis" {
		return nil, errors.New("redis URL must begin with redis://")
	}
	port := a.Port()
	if port == "" {
		port = "6379"
	}
	addr := fmt.Sprintf("%s:%s", a.Hostname(), port)
	dbNo := 0
	if len(a.Path) > 0 {
		var err error
		f := a.Path[1:len(a.Path)]
		dbNo, err = strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
	}
	return &redis.Options{Addr: addr, DB: dbNo}, nil
}

// https://blog.questionable.services/article/guide-logging-middleware-go/
// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}
