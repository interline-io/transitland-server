package model

import (
	"log"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// TODO: replace with middleware or configuration

func MustOpenDB(url string) sqlx.Ext {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.Mapper = reflectx.NewMapperFunc("db", toSnakeCase)
	return db.Unsafe()
}

func Sqrl(db sqlx.Ext) sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
