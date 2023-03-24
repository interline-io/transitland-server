package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/dmfr/store"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/request"
	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/tidwall/gjson"
)

const latestFeedVersionQuery = `
query($feed_onestop_id: String!, $ids: [Int!]) {
	feeds(ids: $ids, where: { onestop_id: $feed_onestop_id }) {
	  onestop_id
	  license {
		redistribution_allowed
	  }
	  feed_versions(limit: 1) {
		sha1
	  }
	}
  }
`

// Query redirects user to download the given fv from S3 public URL
// assuming that redistribution is allowed for the feed.
func feedVersionDownloadLatestHandler(cfg restConfig, w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "feed_key")
	gvars := hw{}
	if key == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else if v, err := strconv.Atoi(key); err == nil {
		gvars["ids"] = []int{v}
	} else {
		gvars["feed_onestop_id"] = key
	}

	// Check if we're allowed to redistribute feed and look up latest feed version
	feedResponse, err := makeGraphQLRequest(r.Context(), cfg.srv, latestFeedVersionQuery, gvars)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	found := false
	allowed := false
	json, err := json.Marshal(feedResponse)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if gjson.Get(string(json), "feeds.0.feed_versions.0.sha1").Exists() {
		found = true
	}
	if gjson.Get(string(json), "feeds.0.license.redistribution_allowed").String() != "no" {
		allowed = true
	}
	fid := gjson.Get(string(json), "feeds.0.onestop_id").String()
	fvsha1 := gjson.Get(string(json), "feeds.0.feed_versions.0.sha1").String()
	if !found {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !allowed {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	// Send request to metering
	if apiMeter := meters.ForContext(r.Context()); apiMeter != nil {
		dims := map[string]string{
			"fv_sha1":                fvsha1,
			"feed_onestop_id":        fid,
			"is_latest_feed_version": "true",
		}
		apiMeter.Meter("feed-version-downloads", 1.0, dims)
	}

	serveFromStorage(w, r, cfg.Storage, fvsha1)
}

const feedVersionFileQuery = `
query($feed_version_sha1:String!, $ids: [Int!]) {
	feed_versions(limit:1, ids: $ids, where:{sha1:$feed_version_sha1}) {
	  sha1
	  feed {
		license {
			redistribution_allowed
		}
	  }
	}
  }
`

// Query redirects user to download the given fv from S3 public URL
// assuming that redistribution is allowed for the feed.
func feedVersionDownloadHandler(cfg restConfig, w http.ResponseWriter, r *http.Request) {
	gvars := hw{}
	key := chi.URLParam(r, "feed_version_key")
	if key == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else if v, err := strconv.Atoi(key); err == nil {
		gvars["ids"] = []int{v}
	} else {
		gvars["feed_version_sha1"] = key
	}
	// Check if we're allowed to redistribute feed
	checkfv, err := makeGraphQLRequest(r.Context(), cfg.srv, feedVersionFileQuery, gvars)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// todo: use gjson
	found := false
	allowed := false
	fid := ""
	fvsha1 := ""
	if v, ok := checkfv["feed_versions"].([]interface{}); len(v) > 0 && ok {
		if v2, ok := v[0].(hw); ok {
			fvsha1 = v2["sha1"].(string)
			if fvsha1 == key {
				found = true
			}
			if v3, ok := v2["feed"].(hw); ok {
				fid = v3["onestop_id"].(string)
				if v4, ok := v3["license"].(hw); ok {
					if v4["redistribution_allowed"] != "no" {
						allowed = true
					}
				}
			}
		}
	}
	if !found {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !allowed {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	// Send request to metering
	if apiMeter := meters.ForContext(r.Context()); apiMeter != nil {
		dims := map[string]string{
			"fv_sha1":                fvsha1,
			"feed_onestop_id":        fid,
			"is_latest_feed_version": "false",
		}
		apiMeter.Meter("feed-version-downloads", 1.0, dims)
	}

	serveFromStorage(w, r, cfg.Storage, fvsha1)
}

func serveFromStorage(w http.ResponseWriter, r *http.Request, storage string, fvsha1 string) {
	store, err := store.GetStore(storage)
	if err != nil {
		http.Error(w, "failed access file", http.StatusInternalServerError)
		return
	}
	fvkey := fmt.Sprintf("%s.zip", fvsha1)
	if v, ok := store.(request.Presigner); ok {
		signedUrl, err := v.CreateSignedUrl(r.Context(), fvkey, tl.Secret{})
		if err != nil {
			http.Error(w, "failed access file", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Location", signedUrl)
		w.WriteHeader(http.StatusFound)
	} else {
		rdr, _, err := store.Download(r.Context(), fvkey, tl.Secret{}, tl.FeedAuthorization{})
		if err != nil {
			http.Error(w, "failed access file", http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(w, rdr); err != nil {
			http.Error(w, "failed access file", http.StatusInternalServerError)
		}
	}
}
