package httpcache

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/tinylru"
)

type cacheResponse struct {
	Headers    map[string][]string
	Body       []byte
	StatusCode int
}

func newCacheResponse(res *http.Response) (*cacheResponse, error) {
	// Save and restore body
	var bodyB []byte
	if res.Body != nil {
		bodyB, _ = ioutil.ReadAll(res.Body)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyB))
	}

	c := cacheResponse{}
	c.Body = bodyB
	c.Headers = map[string][]string{}
	for k, v := range res.Header {
		c.Headers[k] = v
	}
	c.StatusCode = res.StatusCode
	return &c, nil
}

func fromCacheResponse(a *cacheResponse) (*http.Response, error) {
	rr := http.Response{}
	rr.Body = io.NopCloser(bytes.NewReader(a.Body))
	rr.ContentLength = int64(len(a.Body))
	rr.StatusCode = a.StatusCode
	rr.Header = http.Header{}
	for k, v := range a.Headers {
		for _, vv := range v {
			rr.Header.Add(k, vv)
		}
	}
	return &rr, nil
}

type HTTPKey func(*http.Request) string

func DefaultKey(req *http.Request) string {
	// Save and restore body
	var bodyB []byte
	if req.Body != nil {
		bodyB, _ = ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyB))
	}

	// Key
	s := sha1.New()
	s.Write([]byte(req.Method))
	s.Write([]byte(req.URL.String()))
	s.Write(bodyB)
	for k, v := range req.Header {
		s.Write([]byte(k))
		for _, vv := range v {
			s.Write([]byte(vv))
		}
	}
	return fmt.Sprintf("%x", s.Sum(nil))
}

type Cache interface {
	Get(interface{}) (interface{}, bool)
	Set(interface{}, interface{}) (interface{}, bool)
	Len() int
}

type HTTPCache struct {
	Key          HTTPKey
	RoundTripper http.RoundTripper
	cache        Cache
}

func NewHTTPCache(rt http.RoundTripper, key HTTPKey) *HTTPCache {
	if key == nil {
		key = DefaultKey
	}
	if rt == nil {
		rt = http.DefaultTransport
	}
	lrucache := tinylru.LRU{}
	lrucache.Resize(16 * 1024)
	return &HTTPCache{
		RoundTripper: rt,
		Key:          key,
		cache:        &lrucache,
	}
}

func (h *HTTPCache) makeRequest(req *http.Request, key string) (*http.Response, error) {
	// Make request
	res, err := h.RoundTripper.RoundTrip(req)
	if err != nil {
		return res, err
	}
	// Save response
	rr, err := newCacheResponse(res)
	if err != nil {
		return nil, err
	}
	h.cache.Set(key, rr)
	// _, _, evictedKey, _, evicted := h.lrucache.Set(key, rr)
	// if evicted {
	// 	fmt.Println("lru cache evicted:", evictedKey)
	// }
	// fmt.Println("roundtrip: saved value for ", key)
	// Return
	return res, nil
}

func (h *HTTPCache) check(key string) (*http.Response, error) {
	if a, ok := h.cache.Get(key); ok {
		// fmt.Println("roundtrip: got cached value for ", key)
		v, ok := a.(*cacheResponse)
		if ok {
			return fromCacheResponse(v)
		}
	}
	return nil, nil
}

func (h *HTTPCache) RoundTrip(req *http.Request) (*http.Response, error) {
	// fmt.Println("roundtrip:", req.URL)
	key := h.Key(req)
	if a, err := h.check(key); a != nil {
		return a, err
	}
	return h.makeRequest(req, key)
}
