package testutil

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/jmoiron/sqlx"
)

// Test helpers

var db *sqlx.DB

func CheckEnv(key string) (string, string, bool) {
	g := os.Getenv(key)
	if g == "" {
		return "", fmt.Sprintf("%s is not set, skipping", key), false
	}
	return g, "", true
}

func CheckTestDB() (string, bool) {
	_, a, ok := CheckEnv("TL_TEST_SERVER_DATABASE_URL")
	return a, ok
}

func CheckTestRedisClient() (string, bool) {
	_, a, ok := CheckEnv("TL_TEST_REDIS_URL")
	return a, ok
}

func MustOpenTestDB() *sqlx.DB {
	if db != nil {
		return db
	}
	dburl := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	var err error
	db, err = dbutil.OpenDB(dburl)
	if err != nil {
		log.Fatal().Err(err).Msgf("database error")
		return nil
	}
	return db
}

func MustOpenTestRedisClient() *redis.Client {
	redisClient, err := dbutil.OpenRedis(os.Getenv("TL_TEST_REDIS_URL"))
	if err != nil {
		log.Fatal().Err(err).Msgf("redis error")
		return nil
	}
	return redisClient
}
