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
		redirectUrl = fmt.Sprintf("/rest/feeds/%s", onestop_id)
		// redirect to feeds/
	} else if strings.HasPrefix(onestop_id, "o-") {
		redirectUrl = fmt.Sprintf("/rest/operators/%s", onestop_id)
	} else if strings.HasPrefix(onestop_id, "s-") {
		redirectUrl = fmt.Sprintf("/rest/stops/%s", onestop_id)
	} else if strings.HasPrefix(onestop_id, "r-") {
		redirectUrl = fmt.Sprintf("/rest/routes/%s", onestop_id)
	}
	if redirectUrl != "" {
		w.Header().Add("Location", redirectUrl)
		w.WriteHeader(http.StatusFound)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
}
