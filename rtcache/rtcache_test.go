package rtcache

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"google.golang.org/protobuf/proto"
)

var (
	feeds            = []string{"BA", "SF", "AC", "CT"}
	fetchCycles      = 1
	fetchWorkers     = 4
	consumerWorkers  = 2
	consumerInterval = 10 * time.Second
)

func testJobs(t *testing.T, rtJobs JobQueue) {
	for _, feed := range feeds {
		url := fmt.Sprintf("test/%s.pb", feed)
		rtJobs.AddJob(Job{Feed: feed, URL: url})
	}
	var foundJobs []Job
	go func() {
		// Get first item and then return
		ch, err := rtJobs.Listen()
		if err != nil {
			t.Error(err)
		}
		for job := range ch {
			t.Log("got job:", job)
			foundJobs = append(foundJobs, job)
		}
	}()
	time.Sleep(200 * time.Millisecond)
	rtJobs.Close()
	if len(foundJobs) != len(feeds) {
		t.Errorf("got %d jobs, expected %d", len(foundJobs), len(feeds))
	}
}

func testCache(t *testing.T, rtCache Cache) {
	var topics []string
	for _, feed := range feeds {
		topic := fmt.Sprintf("%s-%d", feed, time.Now().UnixNano())
		rtdata := []byte(fmt.Sprintf("test-%s-%d", feed, time.Now().UnixNano()))
		if err := rtCache.AddData(topic, rtdata); err != nil {
			t.Fatal(err)
		}
		topics = append(topics, topic)
	}
	found := [][]byte{}
	for _, topic := range topics {
		go func(fid string) {
			a, err := rtCache.Listen(fid)
			if err != nil {
				t.Error(err)
			}
			for data := range a {
				found = append(found, data)
			}
		}(topic)
	}
	time.Sleep(200 * time.Millisecond)
	rtCache.Close()
	if len(found) != len(feeds) {
		t.Errorf("got %d items, expected %d", len(found), len(feeds))
	}
}

func testConsumers(t *testing.T, rtCache Cache, rtJobs JobQueue) {
	// Start consumers
	rtManager := NewRTConsumerManager(rtCache, nil)
	var foundTrips []*pb.TripUpdate
	for i := 0; i < consumerWorkers; i++ {
		for _, feed := range feeds {
			go func(wid int, feed string) {
				for {
					fmt.Printf("reader '%s': waiting\n", feed)
					// peek and get the first trip...
					tids, err := rtManager.GetTripIDs(feed)
					if err != nil {
						fmt.Printf("reader '%s': error %s\n", feed, err.Error())
					}
					if len(tids) > 0 {
						trip, ok := rtManager.GetTrip(feed, tids[0])
						if ok {
							fmt.Printf("reader '%s': trip '%s' ok\n", feed, tids[0])
							foundTrips = append(foundTrips, trip)
						} else {
							fmt.Printf("reader '%s': trip '%s' not found\n", feed, tids[0])
							t.Error("trip not found")
						}
					}
					time.Sleep(consumerInterval)
				}
			}(i, feed)
		}
	}

	// Do fetching in go routines to test
	// blocking on waiting for first result
	go func() {
		for i := 0; i < fetchCycles; i++ {
			for _, feed := range feeds {
				url := fmt.Sprintf("test/%s.pb", feed)
				rtJobs.AddJob(Job{Feed: feed, URL: url})
			}
		}
	}()

	// Start fetch workers
	for i := 0; i < fetchWorkers; i++ {
		go func(wid int) {
			fmt.Printf("worker %d: start\n", wid)
			jobQueue, err := rtJobs.Listen()
			if err != nil {
				t.Error(err)
			}
			for job := range jobQueue {
				fmt.Printf("worker %d: received job: %s\n", wid, job.Feed)
				rtdata, err := fetchFile(job.URL)
				if err != nil {
					t.Error(err)
					break
				}
				if err := rtCache.AddData(job.Feed, rtdata); err != nil {
					t.Error(err)
					break
				}
			}
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	rtJobs.Close()
	rtCache.Close()
	fmt.Println("got trips:", len(foundTrips))
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
