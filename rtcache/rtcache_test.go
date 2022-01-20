package rtcache

import (
	"fmt"
	"testing"
	"time"
)

var (
	feeds = []string{"BA", "SF", "AC", "CT"}
)

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

// func testConsumers(t *testing.T, rtCache Cache, rtJobs workers.JobQueue) {
// fetchCycles      = 1
// fetchWorkers     = 4
// consumerWorkers  = 2
// consumerInterval = 10 * time.Second
// 	// Start consumers
// 	rtManager := NewRTFinder(rtCache, nil) // todo: DB
// 	var foundTrips []*pb.TripUpdate
// 	for i := 0; i < consumerWorkers; i++ {
// 		for _, feed := range feeds {
// 			go func(wid int, feed string) {
// 				for {
// 					fmt.Printf("reader '%s': waiting\n", feed)
// 					// peek and get the first trip...
// 					tids, err := rtManager.GetTripIDs(feed)
// 					if err != nil {
// 						fmt.Printf("reader '%s': error %s\n", feed, err.Error())
// 					}
// 					if len(tids) > 0 {
// 						trip, ok := rtManager.GetTrip(feed, tids[0])
// 						if ok {
// 							fmt.Printf("reader '%s': trip '%s' ok\n", feed, tids[0])
// 							foundTrips = append(foundTrips, trip)
// 						} else {
// 							fmt.Printf("reader '%s': trip '%s' not found\n", feed, tids[0])
// 							t.Error("trip not found")
// 						}
// 					}
// 					time.Sleep(consumerInterval)
// 				}
// 			}(i, feed)
// 		}
// 	}

// 	// Do fetching in go routines to test
// 	// blocking on waiting for first result
// 	go func() {
// 		for i := 0; i < fetchCycles; i++ {
// 			for _, feed := range feeds {
// 				url := fmt.Sprintf("test/%s.pb", feed)
// 				rtJobs.AddJob(workers.Job{Feed: feed, URL: url})
// 			}
// 		}
// 	}()

// 	// Start fetch workers
// 	for i := 0; i < fetchWorkers; i++ {
// 		go func(wid int) {
// 			fmt.Printf("worker %d: start\n", wid)
// 			jobfunc := func(job workers.Job) error { return nil }
// 			err := rtJobs.AddWorker(jobfunc, 1)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 		}(i)
// 	}
// 	time.Sleep(100 * time.Millisecond)
// 	rtJobs.Stop()
// 	rtCache.Close()
// 	fmt.Println("got trips:", len(foundTrips))
// }
