package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Query returns a GraphQL query string and variables.
func onestopIdEntityRedirectHandler(cfg restConfig, w http.ResponseWriter, r *http.Request) {
	onestop_id := mux.Vars(r)["onestop_id"]
	var redirectUrl string
	if strings.HasPrefix(onestop_id, "f-") {
		redirectUrl = fmt.Sprintf("%s/feeds/%s", cfg.RestPrefix, onestop_id)
		// redirect to feeds/
	} else if strings.HasPrefix(onestop_id, "o-") {
		redirectUrl = fmt.Sprintf("%s/operators/%s", cfg.RestPrefix, onestop_id)
	} else if strings.HasPrefix(onestop_id, "s-") {
		redirectUrl = fmt.Sprintf("%s/stops/%s", cfg.RestPrefix, onestop_id)
	} else if strings.HasPrefix(onestop_id, "r-") {
		redirectUrl = fmt.Sprintf("%s/routes/%s", cfg.RestPrefix, onestop_id)
	}
	if redirectUrl != "" {
		w.Header().Add("Location", redirectUrl)
		w.WriteHeader(http.StatusFound)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
}
