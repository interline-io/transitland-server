package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-mw/auth/authn"
)

// log request and duration
func LoggingMiddleware(longQueryDuration int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			t1 := time.Now()
			userName := ""
			if user := authn.ForContext(ctx); user != nil {
				userName = user.ID()
			}
			// Get request body for logging if request is json and length under 20kb
			var body []byte
			if r.Header.Get("content-type") == "application/json" && r.ContentLength < 1024*20 {
				body, _ = ioutil.ReadAll(r.Body)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
			// Wrap context to get error code and errors
			wr := wrapResponseWriter(w)
			next.ServeHTTP(wr, r)
			// Extra logging of request body if duration > 1s
			durationMs := (time.Now().UnixNano() - t1.UnixNano()) / 1e6
			msg := log.Info().
				Int64("duration_ms", durationMs).
				Str("method", r.Method).
				Str("path", r.URL.EscapedPath()).
				Str("query", r.URL.Query().Encode()).
				Str("user", userName).
				Int("status", wr.status)
			// Add duration info
			if durationMs > int64(longQueryDuration) {
				// Verify it's valid json
				msg = msg.Bool("long_query", true)
				var x interface{}
				if err := json.Unmarshal(body, &x); err == nil {
					msg = msg.RawJSON("body", body)
				}
			}
			// Get any GraphQL errors. We need to log these because the response
			// code will always be 200.
			// var gqlErrs []string
			// if len(gqlErrs) > 0 {
			// 	msg = msg.Strs("gql_errors", gqlErrs)
			// }
			msg.Msg("request")
		})
	}
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
	status      int
	wroteHeader bool
	http.ResponseWriter
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(response []byte) (int, error) {
	if !rw.wroteHeader {
		rw.status = http.StatusOK
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(response)
}
