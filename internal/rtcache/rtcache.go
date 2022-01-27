package rtcache

import "fmt"

// Cache provides a method for looking up and listening for changed RT data
type Cache interface {
	AddData(string, []byte) error
	Listen(string) (chan []byte, error)
	Close() error
}

func getTopicKey(topic string, t string) string {
	return fmt.Sprintf("rtdata:%s:%s", topic, t)
}
