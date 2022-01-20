package workers

import (
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/rtcache"
)

type Command struct {
	QueueName string
	Workers   int
	config.Config
}

func (cmd *Command) Parse(args []string) error {
	cmd.QueueName = "tlv2-default"
	cmd.Workers = 1
	return nil
}

func (cmd *Command) Run() error {
	redisClient := redis.NewClient(&redis.Options{Addr: cmd.RT.RedisURL})
	finder := find.NewDBFinder(find.MustOpenDB(cmd.DB.DBURL))
	rtFinder := rtcache.NewRTFinder(rtcache.NewRedisCache(redisClient), finder.DBX())
	jq := NewRedisJobs(redisClient, cmd.QueueName)
	jr, err := NewJobRunner(cmd.Config, finder, rtFinder, jq, cmd.QueueName, cmd.Workers)
	if err != nil {
		return err
	}
	return jr.RunWorkers()
}
