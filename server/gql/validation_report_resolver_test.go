package gql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestValidationReportResolver(t *testing.T) {
	fvsha1 := "96b67c0934b689d9085c52967365d8c233ea321d"
	q := `query($feed_version_sha1: String!, $where: ValidationReportFilter) {  feed_versions(where:{sha1:$feed_version_sha1}) {validation_reports(where:$where) {success failure_reason includes_static includes_rt validator validator_version errors { filename error_type error_code field count errors { filename error_type error_code entity_id field line value message geometry }} }} }`
	testcases := []testcase{
		// Saved validation reports
		{
			name:  "validation reports",
			query: q,
			vars:  hw{"feed_version_sha1": fvsha1},
			f: func(t *testing.T, jj string) {
				reports := gjson.Get(jj, "feed_versions.0.validation_reports")
				assert.Equal(t, 1, len(reports.Array()))
				report := reports.Get("0")
				assert.Equal(t, []string{"stops.txt", "stops.txt"}, astr(report.Get("errors.#.filename").Array()))
				assert.Equal(t, []string{"InvalidFieldError", "InvalidFieldError"}, astr(report.Get("errors.#.error_type").Array()))
				assert.Equal(t, []string{"1", "1"}, astr(report.Get("errors.#.count").Array()))
				var messages []string
				for _, a := range report.Get("errors").Array() {
					messages = append(messages, astr(a.Get("errors.#.message").Array())...)
				}
				expMessages := []string{
					"invalid value for field stop_lat '-200.000000': out of bounds, min -90.000000 max 90.000000",
					"invalid value for field stop_lon '-200.000000': out of bounds, min -180.000000 max 180.000000",
				}
				assert.ElementsMatch(t, expMessages, messages)
			},
		},
		{
			name:         "success=true",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"success": true}},
			selector:     "feed_versions.0.validation_reports.#.success",
			selectExpect: []string{"true"},
		},
		{
			name:         "success=false",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"success": false}},
			selector:     "feed_versions.0.validation_reports.#.success",
			selectExpect: []string{},
		},
		{
			name:         "includes_static=true",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"includes_static": true}},
			selector:     "feed_versions.0.validation_reports.#.includes_static",
			selectExpect: []string{"true"},
		},
		{
			name:         "includes_static=false",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"includes_static": false}},
			selector:     "feed_versions.0.validation_reports.#.includes_static",
			selectExpect: []string{},
		},

		{
			name:         "includes_rt=true",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"includes_rt": true}},
			selector:     "feed_versions.0.validation_reports.#.includes_rt",
			selectExpect: []string{},
		},
		{
			name:         "includes_rt=false",
			query:        q,
			vars:         hw{"feed_version_sha1": fvsha1, "where": hw{"includes_rt": false}},
			selector:     "feed_versions.0.validation_reports.#.includes_rt",
			selectExpect: []string{"false"},
		},
	}
	c, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}
