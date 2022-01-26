package resolvers

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Additional tests for RT data on StopResolver

// rtFetchJson fetches test protobuf in JSON format
// URL is relative to project root
func rtFetchJson(feed string, ftype string, url string, rtfinder model.RTFinder) error {
	var msg pb.FeedMessage
	jdata, err := ioutil.ReadFile(RelPath(url))
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(jdata, &msg); err != nil {
		return err
	}
	rtdata, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return rtfinder.AddData(key, rtdata)
}

func TestStopRTResolver(t *testing.T) {
	baseQuery := `query($stop_id:String!, $stf:StopTimeFilter!) {
		stops(where: { stop_id: $stop_id }) {
		  id
		  stop_id
		  stop_name
		  stop_times(where:$stf) {
			trip {
			  trip_id
			}
			arrival {
				scheduled
				estimated
				estimated_utc
				stop_timezone
				delay
				uncertainty
			}
			departure {
				scheduled
				estimated
				estimated_utc
				stop_timezone
				delay
				uncertainty

			}
		  }
		}
	  }
	`

	baseSTVars := hw{"service_date": "2018-05-30", "start_time": 57600, "end_time": 64800}
	testcases := []struct {
		name  string
		query string
		vars  map[string]interface{}
		cb    func(t *testing.T, jj string)
	}{
		{
			"basic",
			baseQuery,
			hw{"stop_id": "FTVL", "stf": baseSTVars},
			func(t *testing.T, jj string) { fmt.Println("jj:", jj) },
		},
	}

	cfg := config.Config{}
	srv, _ := NewServer(cfg, TestDBFinder, TestRTFinder)
	c := client.New(srv)

	// Load RT data for test
	if err := rtFetchJson("BA", "trip_updates", "test/data/rt/BA.json", TestRTFinder); err != nil {
		t.Fatal(err)
	}
	if err := rtFetchJson("CT", "trip_updates", "test/data/rt/CT.json", TestRTFinder); err != nil {
		t.Fatal(err)
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var resp map[string]interface{}
			opts := []client.Option{}
			for k, v := range tc.vars {
				opts = append(opts, client.Var(k, v))
			}
			c.MustPost(tc.query, &resp, opts...)
			jj := toJson(resp)
			if tc.cb != nil {
				tc.cb(t, jj)
			}
		})
	}
}
