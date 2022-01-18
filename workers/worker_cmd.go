package workers

import (
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
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
	cmd.Config.DB = config.DBConfig{
		DB: model.MustOpenDB(cmd.DB.DBURL),
	}
	cmd.Config.RT = config.RTConfig{
		Redis: redis.NewClient(&redis.Options{Addr: cmd.RT.RedisURL}),
	}
	jr, err := NewJobRunner(cmd.Config, cmd.QueueName, cmd.Workers)
	if err != nil {
		return err
	}
	return jr.RunWorkers()
}
