package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/auth/ancheck"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/interline-io/transitland-server/server/playground"
	"github.com/interline-io/transitland-server/server/rest"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Command struct {
	Timeout            int
	LongQueryDuration  int
	Port               string
	RestPrefix         string
	LoadAdmins         bool
	ValidateLargeFiles bool
	SecretsFile        string
	Storage            string
	RTStorage          string
	DBURL              string
	RedisURL           string
	secrets            []tl.Secret
}

func (cmd *Command) Parse(args []string) error {
	fl := flag.NewFlagSet("sync", flag.ExitOnError)
	fl.Usage = func() {
		log.Print("Usage: server")
		fl.PrintDefaults()
	}
	fl.StringVar(&cmd.DBURL, "dburl", "", "Database URL (default: $TL_DATABASE_URL)")
	fl.StringVar(&cmd.RedisURL, "redisurl", "", "Redis URL (default: $TL_REDIS_URL)")
	fl.StringVar(&cmd.Storage, "storage", "", "Static storage backend")
	fl.StringVar(&cmd.RTStorage, "rt-storage", "", "RT storage backend")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.SecretsFile, "secrets", "", "DMFR file containing secrets")
	fl.IntVar(&cmd.Timeout, "timeout", 60, "")
	fl.IntVar(&cmd.LongQueryDuration, "long-query", 1000, "Log queries over this duration (ms)")
	fl.BoolVar(&cmd.LoadAdmins, "load-admins", false, "Load admin polygons from database into memory")
	fl.Parse(args)

	// DB
	if cmd.DBURL == "" {
		cmd.DBURL = os.Getenv("TL_DATABASE_URL")
	}
	if cmd.RedisURL == "" {
		cmd.RedisURL = os.Getenv("TL_REDIS_URL")
	}

	// Load secrets
	var secrets []tl.Secret
	if v := cmd.SecretsFile; v != "" {
		rr, err := dmfr.LoadAndParseRegistry(v)
		if err != nil {
			return errors.New("unable to load secrets file")
		}
		secrets = rr.Secrets
	}
	cmd.secrets = secrets
	return nil
}

func (cmd *Command) Run() error {
	// Open database
	var db sqlx.Ext
	dbx, err := dbutil.OpenDB(cmd.DBURL)
	if err != nil {
		return err
	}
	db = dbx
	if log.Logger.GetLevel() == zerolog.TraceLevel {
		db = &tldb.QueryLogger{Ext: dbx}
	}

	// Open redis
	var redisClient *redis.Client
	if cmd.RedisURL != "" {
		redisClient, err = dbutil.OpenRedis(cmd.RedisURL)
		if err != nil {
			return err
		}
	}

	// Create Finder
	dbFinder := dbfinder.NewFinder(db)
	if cmd.LoadAdmins {
		dbFinder.LoadAdmins()
	}

	// Create RTFinder, GBFSFinder
	var rtFinder model.RTFinder
	var gbfsFinder model.GbfsFinder
	if redisClient != nil {
		// Use redis backed finders
		rtFinder = rtfinder.NewFinder(rtfinder.NewRedisCache(redisClient), db)
		gbfsFinder = gbfsfinder.NewFinder(redisClient)
	} else {
		// Default to in-memory cache
		rtFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
		gbfsFinder = gbfsfinder.NewFinder(nil)
	}

	// Setup config
	cfg := model.Config{
		Finder:             dbFinder,
		RTFinder:           rtFinder,
		GbfsFinder:         gbfsFinder,
		Secrets:            cmd.secrets,
		Storage:            cmd.Storage,
		RTStorage:          cmd.RTStorage,
		ValidateLargeFiles: cmd.ValidateLargeFiles,
		RestPrefix:         cmd.RestPrefix,
	}

	// Setup router
	root := chi.NewRouter()
	root.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"content-type", "apikey", "authorization"},
		AllowCredentials: true,
	}))

	// Finders config
	root.Use(model.AddConfig(cfg))

	// This server only supports admin access
	root.Use(ancheck.AdminDefaultMiddleware("admin"))

	// Add logging middleware - must be after auth
	root.Use(log.LoggingMiddleware(cmd.LongQueryDuration, func(ctx context.Context) string {
		if user := authn.ForContext(ctx); user != nil {
			return user.Name()
		}
		return ""
	}))

	// Profiling
	root.HandleFunc("/debug/pprof/", pprof.Index)
	root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	root.HandleFunc("/debug/pprof/profile", pprof.Profile)
	root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// GraphQL API
	graphqlServer, err := gql.NewServer()
	if err != nil {
		return err
	}
	if true {
		r := chi.NewRouter()
		r.Mount("/", graphqlServer)
		root.Mount("/query", r)
	}

	// REST API
	restServer, err := rest.NewServer(graphqlServer)
	if err != nil {
		return err
	}
	if true {
		r := chi.NewRouter()
		r.Mount("/", restServer)
		root.Mount("/rest", r)
	}

	// GraphQL Playground
	root.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// Start server
	timeOut := time.Duration(cmd.Timeout) * time.Second
	addr := fmt.Sprintf("%s:%s", "0.0.0.0", cmd.Port)
	log.Infof("Listening on: %s", addr)
	srv := &http.Server{
		Handler:      http.TimeoutHandler(root, timeOut, "timeout"),
		Addr:         addr,
		WriteTimeout: 2 * timeOut,
		ReadTimeout:  2 * timeOut,
	}
	return srv.ListenAndServe()
}
