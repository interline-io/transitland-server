package find

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-lib/log"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// MAXLIMIT .
const MAXLIMIT = 1000

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func MustOpenDB(url string) sqlx.Ext {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		log.Fatal().Err(err).Msg("could not open database")
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msgf("could not connect to database")
	}
	db.Mapper = reflectx.NewMapperFunc("db", toSnakeCase)
	return db.Unsafe()
	// return &tldb.QueryLogger{Ext: db.Unsafe()}
}

// MustSelect runs a query or panics.
func MustSelect(ctx context.Context, db sqlx.Ext, q sq.SelectBuilder, dest interface{}) {
	useStatement := false
	q = q.PlaceholderFormat(sq.Dollar)
	qstr, qargs := q.MustSql()
	if a, ok := db.(sqlx.PreparerContext); ok && useStatement {
		stmt, err := sqlx.PreparexContext(ctx, a, qstr)
		if err != nil {
			panic(err)
		}
		if err := stmt.SelectContext(ctx, dest, qargs...); err != nil {
			log.Error().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query failed")
			panic(err)
		}
	} else if a, ok := db.(sqlx.QueryerContext); ok {
		if err := sqlx.SelectContext(ctx, a, dest, qstr, qargs...); err != nil {
			log.Error().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query failed")
			panic(err)
		}
	} else {
		if err := sqlx.Select(db, dest, qstr, qargs...); err != nil {
			log.Error().Err(err).Str("query", qstr).Interface("args", qargs).Msg("query failed")
			panic(err)
		}
	}
}

// helpers

func checkLimit(limit *int) uint64 {
	return checkRange(limit, 0, MAXLIMIT)
}

func checkRange(limit *int, min, max int) uint64 {
	if limit == nil {
		return uint64(max)
	} else if *limit >= max {
		return uint64(max)
	} else if *limit < min {
		return uint64(min)
	}
	return uint64(*limit)
}

func checkFloat(v *float64, min float64, max float64) float64 {
	if v == nil || *v < min {
		return min
	} else if *v > max {
		return max
	}
	return *v
}

func atoi(v string) int {
	a, _ := strconv.Atoi(v)
	return a
}

// unicode aware remove all non-alphanumeric characters
// this is not for escaping sql; just for preparing to_tsquery
func alphanumeric(v string) string {
	ret := []rune{}
	for _, ch := range v {
		if unicode.IsSpace(ch) {
			ret = append(ret, ' ')
		} else if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
			ret = append(ret, ch)
		}
	}
	return string(ret)
}

// az09 removes any character that is not a a-z or 0-9 or _
func az09(v string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return reg.ReplaceAllString(v, "")
}

func escapeWordsWithSuffix(v string, sfx string) []string {
	var ret []string
	for _, s := range strings.Fields(v) {
		aa := alphanumeric(s)
		// Minimum length 2 characters
		if len(aa) > 1 {
			ret = append(ret, aa+sfx)
		}
	}
	return ret
}

func tsQuery(s string) (rank sq.Sqlizer, wc sq.Sqlizer) {
	s = strings.TrimSpace(s)
	words := append([]string{}, escapeWordsWithSuffix(s, ":*")...)
	wordstsq := strings.Join(words, " & ")
	rank = sq.Expr("ts_rank_cd(t.textsearch,to_tsquery('tl',?)) as search_rank", wordstsq)
	wc = sq.Expr("t.textsearch @@ to_tsquery('tl',?)", wordstsq)
	return rank, wc
}

func lateralWrap(q sq.SelectBuilder, parent string, pkey string, ckey string, pids []int) sq.SelectBuilder {
	parent = az09(parent)
	pkey = az09(pkey)
	ckey = az09(ckey)
	q = q.Where(fmt.Sprintf("%s = parent.%s", ckey, pkey))
	q2 := sq.StatementBuilder.
		Select("t.*").
		From(parent + " parent").
		JoinClause(q.Prefix("JOIN LATERAL (").Suffix(") t on true")).
		Where(sq.Eq{"parent." + pkey: pids})
	return q2
}

func quickSelect(table string, limit *int, after *int, ids []int) sq.SelectBuilder {
	return quickSelectOrder(table, limit, after, ids, "id")
}

func quickSelectOrder(table string, limit *int, after *int, ids []int, order string) sq.SelectBuilder {
	table = az09(table)
	order = az09(order)
	q := sq.StatementBuilder.
		Select("t.*").
		From(table + " t").
		Limit(checkLimit(limit))
	if order != "" {
		q = q.OrderBy("t." + order)
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	return q
}
