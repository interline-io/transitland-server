package gql

import (
	"fmt"
	"testing"
)

func TestValidationReportResolver(t *testing.T) {
	vars := hw{"feed_version_sha1": "32506751d043b4f4675409ba7cb230f89f7114e4"}
	testcases := []testcase{
		// Saved validation reports
		{
			name:         "validation reports",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {validation_reports{success failure_reason errors { filename error_type error_code message field count limit errors { filename error_type error_code entity_id field line value message geometry }} }} }`,
			vars:         vars,
			selector:     "feed_versions.0.validation_reports.0.success",
			selectExpect: []string{"true"},
			f: func(t *testing.T, jj string) {
				fmt.Println(jj)
			},
		},
	}
	c, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}
