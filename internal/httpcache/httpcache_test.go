package httpcache

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/tinylru"
)

func TestHTTPCache(t *testing.T) {
	requests := 0
	ts200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello!"))
		requests += 1
	}))

	testClient := func(t *testing.T, client *http.Client, count int, a int) {
		requests = 0
		var req *http.Request
		// first
		for i := 0; i < count; i++ {
			req, _ = http.NewRequest("GET", ts200.URL, nil)
			res, _ := client.Do(req)
			res.Body.Close()
		}
		assert.Equal(t, a, requests)
		// second
		for i := 0; i < count; i++ {
			req, _ = http.NewRequest("POST", ts200.URL, nil)
			res, _ := client.Do(req)
			res.Body.Close()
		}
		assert.Equal(t, a*2, requests)
		// third
		for i := 0; i < count; i++ {
			req, _ = http.NewRequest("POST", ts200.URL, nil)
			req.Header.Add("foo", "bar")
			res, _ := client.Do(req)
			res.Body.Close()
		}
		assert.Equal(t, a*3, requests)
		// fourth
		for i := 0; i < count; i++ {
			req, _ = http.NewRequest("POST", ts200.URL, bytes.NewReader([]byte(`{"test":"ok"}`)))
			res, _ := client.Do(req)
			res.Body.Close()
		}
		assert.Equal(t, a*4, requests)
		// fifth
		for i := 0; i < count; i++ {
			req, _ = http.NewRequest("GET", ts200.URL+"?a=1", nil)
			res, _ := client.Do(req)
			res.Body.Close()
		}
		assert.Equal(t, a*5, requests)

	}

	t.Run("no cache", func(t *testing.T) {
		testClient(t, &http.Client{}, 100, 100)
	})

	t.Run("with cache", func(t *testing.T) {
		c := NewHTTPCache(nil, nil)
		testClient(t, &http.Client{Transport: c}, 100, 1)
	})

	t.Run("with cache test lru", func(t *testing.T) {
		csize := 10
		c := NewHTTPCache(nil, nil)
		// manually resize...
		cache := tinylru.LRU{}
		cache.Resize(csize)
		c.cache = &cache
		// test
		client := &http.Client{Transport: c}
		// First pass to fill up cache
		for i := 0; i < csize; i++ {
			assert.Equal(t, i, cache.Len())
			req, _ := http.NewRequest("POST", ts200.URL, nil)
			req.Header.Add("foo", fmt.Sprintf("%d", i))
			res, _ := client.Do(req)
			res.Body.Close()
		}
		// Second cache to evict and stay same size
		for i := 0; i < csize*2; i++ {
			assert.Equal(t, 10, cache.Len())
			req, _ := http.NewRequest("POST", ts200.URL, nil)
			req.Header.Add("foo", fmt.Sprintf("%d", i))
			res, _ := client.Do(req)
			res.Body.Close()
		}
	})
}
