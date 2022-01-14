package rtcache

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"google.golang.org/protobuf/proto"
)

var (
	pulsarConnectionTimeout = 30 * time.Second
	JobSchema               = `
	{
		"type": "record",
		"name": "Job",
		"namespace": "test",
		"fields": [
		{
			"name": "job_type",
			"type": "string"
		},
		{
			"name": "feed",
			"type": "string"
		}, {
			"name": "url",
			"type": "string"
		}]
	}`
)

type Cache interface {
	AddData(string, []byte) error
	Listen(string) (chan []byte, error)
	Close() error
}

type JobQueue interface {
	AddJob(Job) error
	Listen() (chan Job, error)
	Close() error
}

type Job struct {
	JobType string   `json:"job_type"`
	Feed    string   `json:"feed"`
	URL     string   `json:"url"`
	Args    []string `json:"args"`
}

//////

type listenChan struct {
	listener chan []byte
	done     chan bool
}

func newListenChan() *listenChan {
	return &listenChan{
		listener: make(chan []byte, 100),
		done:     make(chan bool),
	}
}

//////

func fetchFile(u string) ([]byte, error) {
	pbdata, err := ioutil.ReadFile(u)
	if err != nil {
		panic(err)
	}
	msg := pb.FeedMessage{}
	proto.Unmarshal(pbdata, &msg)
	fmt.Printf("fetched: '%s' %d bytes, entities %d\n", u, len(pbdata), len(msg.Entity))
	return pbdata, nil
}
