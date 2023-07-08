package dbutil

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-redis/redis/v8"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func OpenDB(url string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		log.Error().Err(err).Msg("could not open database")
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		log.Error().Err(err).Msgf("could not connect to database")
		return nil, err
	}
	db.Mapper = reflectx.NewMapperFunc("db", toSnakeCase)
	return db.Unsafe(), nil
}

func LogDB(db *sqlx.DB) sqlx.Ext {
	return &tldb.QueryLogger{Ext: db}
}

// Select runs a query and reads results into dest.
func Select(ctx context.Context, db sqlx.Ext, q sq.SelectBuilder, dest interface{}) error {
	useStatement := false
	q = q.PlaceholderFormat(sq.Dollar)
	qstr, qargs, err := q.ToSql()
	if err == nil {
		if a, ok := db.(sqlx.PreparerContext); ok && useStatement {
			stmt, prepareErr := sqlx.PreparexContext(ctx, a, qstr)
			if prepareErr != nil {
				err = prepareErr
			} else {
				err = stmt.SelectContext(ctx, dest, qargs...)
			}
		} else if a, ok := db.(sqlx.QueryerContext); ok {
			err = sqlx.SelectContext(ctx, a, dest, qstr, qargs...)
		} else {
			err = sqlx.Select(db, dest, qstr, qargs...)
		}
	}
	if ctx.Err() == context.Canceled {
		log.Trace().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query canceled")
	} else if err != nil {
		log.Error().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query failed")
	}
	return err
}

// Select runs a query and reads results into dest.
func Get(ctx context.Context, db sqlx.Ext, q sq.SelectBuilder, dest interface{}) error {
	useStatement := false
	q = q.PlaceholderFormat(sq.Dollar)
	qstr, qargs, err := q.ToSql()
	if err == nil {
		if a, ok := db.(sqlx.PreparerContext); ok && useStatement {
			stmt, prepareErr := sqlx.PreparexContext(ctx, a, qstr)
			if prepareErr != nil {
				err = prepareErr
			} else {
				err = stmt.GetContext(ctx, dest, qargs...)
			}
		} else if a, ok := db.(sqlx.QueryerContext); ok {
			err = sqlx.GetContext(ctx, a, dest, qstr, qargs...)
		} else {
			err = sqlx.Get(db, dest, qstr, qargs...)
		}
	}
	if ctx.Err() == context.Canceled {
		log.Trace().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query canceled")
	} else if err != nil {
		log.Error().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query failed")
	}
	return err
}

// Test helpers

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
	dburl := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	db, err := OpenDB(dburl)
	if err != nil {
		log.Fatal().Err(err).Msgf("database error")
		return nil
	}
	return db
}

func MustOpenTestRedisClient() *redis.Client {
	redisUrl := os.Getenv("TL_TEST_REDIS_URL")
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	return client
}
