package workers

import (
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/model"
)

type Command struct {
	Config Config
}

func (cmd *Command) Parse(args []string) error {
	cmd.Config.QueueName = "tlv2-default"
	return nil
}

func (cmd *Command) Run() error {
	db := model.DB
	client := redis.NewClient(&redis.Options{Addr: cmd.Config.RedisURL})
	jr, err := NewJobRunner(client, db, cmd.Config)
	if err != nil {
		return err
	}
	return jr.RunWorkers()
}
