package httpcache

import (
	"net/http"
)

type Cache interface {
	Get(string) (interface{}, bool)
	Set(string, interface{}) error
	Len() int
	Close() error
}

type HTTPCache struct {
	key          HTTPKey
	roundTripper http.RoundTripper
	cache        Cache
}

func NewHTTPCache(rt http.RoundTripper, key HTTPKey, cache Cache) *HTTPCache {
	if key == nil {
		key = DefaultKey
	}
	if rt == nil {
		rt = http.DefaultTransport
	}
	if cache == nil {
		cache = NewLRUCache(16 * 1024)
	}
	return &HTTPCache{
		roundTripper: rt,
		key:          key,
		cache:        cache,
	}
}

func (h *HTTPCache) makeRequest(req *http.Request, key string) (*http.Response, error) {
	// Make request
	res, err := h.roundTripper.RoundTrip(req)
	if err != nil {
		return res, err
	}
	// Save response
	rr, err := newCacheResponse(res)
	if err != nil {
		return nil, err
	}
	h.cache.Set(key, rr)
	return res, nil
}

func (h *HTTPCache) check(key string) (*http.Response, error) {
	if a, ok := h.cache.Get(key); ok {
		v, ok := a.(*cacheResponse)
		if ok {
			return fromCacheResponse(v)
		}
	}
	return nil, nil
}

func (h *HTTPCache) RoundTrip(req *http.Request) (*http.Response, error) {
	key := h.key(req)
	if a, err := h.check(key); a != nil {
		// fmt.Println("httpcache: got cached:", key)
		return a, err
	}
	rr, err := h.makeRequest(req, key)
	// fmt.Println("httpcache: saved to cache:", key)
	return rr, err
}
