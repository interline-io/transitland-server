package rt

import (
	"fmt"
	"os"

	"github.com/interline-io/transitland-lib/rt"
	"github.com/interline-io/transitland-lib/rt/pb"
)

type msgKey struct {
	feed string
	key  string
}

type MsgCache struct {
	cache map[msgKey]*pb.FeedMessage
}

func (m *MsgCache) Get(feed, key string) (*pb.FeedMessage, bool) {
	a, ok := m.cache[msgKey{feed, key}]
	return a, ok
}

func (m *MsgCache) Set(feed string, key string, msg *pb.FeedMessage) {
	m.cache[msgKey{feed: feed, key: key}] = msg
}

var MC = MsgCache{
	cache: map[msgKey]*pb.FeedMessage{},
}

func (m *MsgCache) StartDebug() {
	fmt.Println("init rt fetch")
	apiKey := os.Getenv("SFBAY511APIKEY")
	fetchRt("f-9q9-actransit", "alerts", fmt.Sprintf("http://api.511.org/transit/serviceAlerts?api_key=%s&agency=AC", apiKey))
	fetchRt("f-9q9-actransit", "trip_updates", fmt.Sprintf("http://api.511.org/transit/TripUpdates?api_key=%s&agency=AC", apiKey))
	fmt.Println("init rt fetch done")
}

func fetchRt(feed string, key string, url string) error {
	m, err := rt.ReadURL(url)
	if err != nil {
		panic(err)
	}
	MC.Set(feed, key, m)
	return nil
}
