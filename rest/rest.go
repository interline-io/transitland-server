package rest

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/meters"
)

// DEFAULTLIMIT is the default API limit
const DEFAULTLIMIT = 20

// MAXLIMIT is the API limit maximum
var MAXLIMIT = 1_000

// MAXRADIUS is the maximum point search radius
const MAXRADIUS = 100 * 1000.0

// restConfig holds the base config and the graphql handler
type restConfig struct {
	config.Config
	srv http.Handler
}

// NewServer .
func NewServer(cfg config.Config, srv http.Handler) (http.Handler, error) {
	restcfg := restConfig{Config: cfg, srv: srv}
	r := chi.NewRouter()

	feedHandler := makeHandler(restcfg, "feeds", func() apiHandler { return &FeedRequest{} })
	feedVersionHandler := makeHandler(restcfg, "feedVersions", func() apiHandler { return &FeedVersionRequest{} })
	agencyHandler := makeHandler(restcfg, "agencies", func() apiHandler { return &AgencyRequest{} })
	routeHandler := makeHandler(restcfg, "routes", func() apiHandler { return &RouteRequest{} })
	tripHandler := makeHandler(restcfg, "trips", func() apiHandler { return &TripRequest{} })
	stopHandler := makeHandler(restcfg, "stops", func() apiHandler { return &StopRequest{} })
	stopDepartureHandler := makeHandler(restcfg, "stopDepartures", func() apiHandler { return &StopDepartureRequest{} })
	operatorHandler := makeHandler(restcfg, "operators", func() apiHandler { return &OperatorRequest{} })

	r.HandleFunc("/feeds.{format}", feedHandler)
	r.HandleFunc("/feeds", feedHandler)
	r.HandleFunc("/feeds/{feed_key}.{format}", feedHandler)
	r.HandleFunc("/feeds/{feed_key}", feedHandler)
	r.HandleFunc("/feeds/{feed_key}/download_latest_feed_version", makeHandlerFunc(restcfg, "feedVersionDownloadLatest", feedVersionDownloadLatestHandler))

	r.HandleFunc("/feed_versions.{format}", feedVersionHandler)
	r.HandleFunc("/feed_versions", feedVersionHandler)
	r.HandleFunc("/feed_versions/{feed_version_key}.{format}", feedVersionHandler)
	r.HandleFunc("/feed_versions/{feed_version_key}", feedVersionHandler)
	r.HandleFunc("/feeds/{feed_key}/feed_versions", feedVersionHandler)
	r.Handle("/feed_versions/{feed_version_key}/download", auth.RoleRequired("tl_user_pro")(makeHandlerFunc(restcfg, "feedVersionDownload", feedVersionDownloadHandler)))

	r.HandleFunc("/agencies.{format}", agencyHandler)
	r.HandleFunc("/agencies", agencyHandler)
	r.HandleFunc("/agencies/{agency_key}.{format}", agencyHandler)
	r.HandleFunc("/agencies/{agency_key}", agencyHandler)

	r.HandleFunc("/routes.{format}", routeHandler)
	r.HandleFunc("/routes", routeHandler)
	r.HandleFunc("/routes/{route_key}.{format}", routeHandler)
	r.HandleFunc("/routes/{route_key}", routeHandler)
	r.HandleFunc("/agencies/{agency_key}/routes.{format}", routeHandler)
	r.HandleFunc("/agencies/{agency_key}/routes", routeHandler)

	r.HandleFunc("/routes/{route_key}/trips.{format}", tripHandler)
	r.HandleFunc("/routes/{route_key}/trips", tripHandler)
	r.HandleFunc("/routes/{route_key}/trips/{id}", tripHandler)
	r.HandleFunc("/routes/{route_key}/trips/{id}.{format}", tripHandler)

	r.HandleFunc("/stops.{format}", stopHandler)
	r.HandleFunc("/stops", stopHandler)
	r.HandleFunc("/stops/{stop_key}.{format}", stopHandler)
	r.HandleFunc("/stops/{stop_key}", stopHandler)

	r.HandleFunc("/stops/{stop_key}/departures", stopDepartureHandler)

	r.HandleFunc("/operators.{format}", operatorHandler)
	r.HandleFunc("/operators", operatorHandler)
	r.HandleFunc("/operators/{operator_key}.{format}", operatorHandler)
	r.HandleFunc("/operators/{operator_key}", operatorHandler)

	return r, nil
}

func getKey(value string) string {
	h := sha1.New()
	h.Write([]byte(value))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// A type that can generate a GraphQL query and variables.
type apiHandler interface {
	Query() (string, map[string]interface{})
}

// A type that can generate a GeoJSON response.
type canProcessGeoJSON interface {
	ProcessGeoJSON(map[string]interface{}) error
}

// A type that defines if meta should be included or not
type canIncludeNext interface {
	IncludeNext() bool
}

// A type that specifies a JSON response key.
type hasResponseKey interface {
	ResponseKey() string
}

// Alias for map string interface
type hw = map[string]interface{}

func commaSplit(v string) []string {
	var ret []string
	for _, i := range strings.Split(v, ",") {
		b := strings.TrimSpace(i)
		if b != "" {
			ret = append(ret, b)
		}
	}
	return ret
}

// checkIds returns a id as a []int{id} slice if >0, otherwise nil.
func checkIds(id int) []int {
	if id > 0 {
		return []int{id}
	}
	return nil
}

// checkAfter checks the value is positive.
func checkAfter(after int) int {
	if after < 0 {
		return 0
	}
	return after
}

// checkLimit checks the limit is positive and below the maximum limit.
func checkLimit(limit int) int {
	if limit <= 0 {
		return DEFAULTLIMIT
	}
	if limit > MAXLIMIT {
		return MAXLIMIT
	}
	return limit
}

// queryToMap converts url.Values to map[string]string
func queryToMap(vars url.Values) map[string]string {
	m := map[string]string{}
	for k := range vars {
		if b := vars.Get(k); b != "" {
			m[k] = vars.Get(k)
		}
	}
	return m
}

// makeHandler wraps an apiHandler into an HandlerFunc and performs common checks.
func makeHandler(cfg restConfig, handlerName string, f func() apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ent := f()
		opts := queryToMap(r.URL.Query())

		// Extract URL params from request
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for _, k := range rctx.URLParams.Keys {
				if k == "*" {
					continue
				}
				opts[k] = rctx.URLParam(k)
			}
		}

		// Metrics
		if apiMeter := meters.ForContext(r.Context()); apiMeter != nil {
			apiMeter.Meter("rest", 1.0, map[string]string{"handler": handlerName})
		}

		format := opts["format"]
		if format == "png" && cfg.DisableImage {
			http.Error(w, "image generation disabled", http.StatusInternalServerError)
			return
		}

		// If this is a image request, check the local cache
		urlkey := getKey(r.URL.Path + "/" + r.URL.RawQuery)
		if format == "png" && localFileCache != nil {
			if ok, _ := localFileCache.Has(urlkey); ok {
				w.WriteHeader(http.StatusOK)
				err := localFileCache.Get(w, urlkey)
				if err != nil {
					log.Error().Err(err).Msg("file cache error")
				}
				return
			}
		}

		// Use json marshal/unmarshal to convert string params to correct types
		s, err := json.Marshal(opts)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal request params")
			http.Error(w, "parameter error", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(s, ent); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal request params")
			http.Error(w, "parameter error", http.StatusInternalServerError)
			return
		}

		// Make the request
		response, err := makeRequest(r.Context(), cfg, ent, format, r.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the output data
		if format == "png" {
			w.Header().Add("Content-Type", "image/png")
		} else {
			w.Header().Add("Content-Type", "application/json")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)

		// Cache image response
		if format == "png" && localFileCache != nil {
			if err := localFileCache.Put(urlkey, bytes.NewReader(response)); err != nil {
				log.Error().Err(err).Msgf("file cache error")
			}
		}
	}
}

// makeGraphQLRequest issues the graphql request and unpacks the response.
func makeGraphQLRequest(ctx context.Context, srv http.Handler, query string, vars map[string]interface{}) (map[string]interface{}, error) {
	gqlData := map[string]any{
		"query":     query,
		"variables": vars,
	}
	gqlBody, err := json.Marshal(gqlData)
	if err != nil {
		return nil, err
	}
	gqlRequest, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewReader(gqlBody))
	gqlRequest.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	wr := httptest.NewRecorder()
	srv.ServeHTTP(wr, gqlRequest)
	response := map[string]any{}
	if err := json.Unmarshal(wr.Body.Bytes(), &response); err != nil {
		return nil, err
	}
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil, err
	}
	return data, nil
}

// makeRequest prepares an apiHandler and makes the request.
func makeRequest(ctx context.Context, cfg restConfig, ent apiHandler, format string, u *url.URL) ([]byte, error) {
	query, vars := ent.Query()
	response, err := makeGraphQLRequest(ctx, cfg.srv, query, vars)
	if err != nil {
		vjson, _ := json.Marshal(vars)
		log.Error().Err(err).Str("query", query).Str("vars", string(vjson)).Msgf("graphql request failed")
		return nil, errors.New("request error")
	}

	// Add meta
	addMeta := true
	if v, ok := ent.(canIncludeNext); ok {
		addMeta = v.IncludeNext()
	}
	if addMeta {
		if lastId, err := getAfterID(ent, response); err != nil {
			log.Error().Err(err).Msg("pagination failed to get max entity id")
		} else if lastId > 0 {
			meta := hw{"after": lastId}
			if u != nil {
				newUrl, err := url.Parse(u.String())
				if err != nil {
					panic(err)
				}
				rq := newUrl.Query()
				rq.Set("after", strconv.Itoa(lastId))
				newUrl.RawQuery = rq.Encode()
				meta["next"] = cfg.RestPrefix + newUrl.String()
			}
			response["meta"] = meta
		}
	}

	if format == "geojson" || format == "png" {
		// TODO: Don't process response in-place.
		if v, ok := ent.(canProcessGeoJSON); ok {
			if err := v.ProcessGeoJSON(response); err != nil {
				return nil, err
			}
		} else {
			if err := processGeoJSON(ent, response); err != nil {
				return nil, err
			}
		}
		if format == "png" {
			b, err := json.Marshal(response)
			if err != nil {
				return nil, err
			}
			return renderMap(b, 800, 800)
		}
	}
	return json.Marshal(response)
}

func makeHandlerFunc(cfg restConfig, handlerName string, f func(restConfig, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(cfg, w, r)
		// Metrics
		if apiMeter := meters.ForContext(r.Context()); apiMeter != nil {
			apiMeter.Meter("rest", 1.0, map[string]string{"handler": handlerName})
		}
	}
}

func getAfterID(ent apiHandler, response map[string]interface{}) (int, error) {
	maxid := 0
	fkey := ""
	if v, ok := ent.(hasResponseKey); ok {
		fkey = v.ResponseKey()
	} else {
		return 0, errors.New("pagination: response key missing")
	}
	entities, ok := response[fkey].([]interface{})
	if !ok {
		return 0, errors.New("pagination: unknown response key value")
	}
	if len(entities) == 0 {
		return 0, nil
	}
	lastEnt, ok := entities[len(entities)-1].(map[string]interface{})
	if !ok {
		return 0, errors.New("pagination: last entity not map[string]interface{}")
	}
	switch id := lastEnt["id"].(type) {
	case int:
		maxid = id
	case float64:
		maxid = int(id)
	case int64:
		maxid = int(id)
	default:
		return 0, errors.New("pagination: last entity id not numeric")
	}
	return maxid, nil
}
