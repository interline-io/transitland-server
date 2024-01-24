package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"net/http"
	"net/http/pprof"
	"net/url"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/dbutil"
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
	DisableImage       bool
	DisableGraphql     bool
	DisableRest        bool
	EnablePlayground   bool
	EnableAdminApi     bool
	EnableJobsApi      bool
	EnableWorkers      bool
	EnableProfiler     bool
	EnableRateLimits   bool
	LoadAdmins         bool
	ValidateLargeFiles bool
	QueuePrefix        string
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

	// Base config
	fl.StringVar(&cmd.DBURL, "dburl", "", "Database URL (default: $TL_DATABASE_URL)")
	fl.StringVar(&cmd.RedisURL, "redisurl", "", "Redis URL (default: $TL_REDIS_URL)")
	fl.StringVar(&cmd.Storage, "storage", "", "Static storage backend")
	fl.StringVar(&cmd.RTStorage, "rt-storage", "", "RT storage backend")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.BoolVar(&cmd.DisableImage, "disable-image", false, "Disable image generation")

	// Server config
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.SecretsFile, "secrets", "", "DMFR file containing secrets")
	fl.StringVar(&cmd.QueuePrefix, "queue", "", "Job name prefix")
	fl.IntVar(&cmd.Timeout, "timeout", 60, "")
	fl.IntVar(&cmd.LongQueryDuration, "long-query", 1000, "Log queries over this duration (ms)")
	fl.BoolVar(&cmd.DisableGraphql, "disable-graphql", false, "Disable GraphQL endpoint")
	fl.BoolVar(&cmd.DisableRest, "disable-rest", false, "Disable REST endpoint")
	fl.BoolVar(&cmd.EnablePlayground, "enable-playground", false, "Enable GraphQL playground")
	fl.BoolVar(&cmd.EnableProfiler, "enable-profile", false, "Enable profiling")
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
		db = dbutil.LogDB(dbx)
	}

	// Open redis
	var redisClient *redis.Client
	if cmd.RedisURL != "" {
		rOpts, err := getRedisOpts(cmd.RedisURL)
		if err != nil {
			return err
		}
		redisClient = redis.NewClient(rOpts)
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
		DisableImage:       cmd.DisableImage,
		RestPrefix:         cmd.RestPrefix,
	}

	// Setup router
	root := chi.NewRouter()
	root.Use(middleware.RequestID)
	root.Use(middleware.RealIP)
	root.Use(middleware.Recoverer)
	root.Use(middleware.StripSlashes)
	root.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"content-type", "apikey", "authorization"},
		AllowCredentials: true,
	}))

	// Finders config
	root.Use(model.AddConfig(cfg))

	// Add logging middleware - must be after auth
	root.Use(log.LoggingMiddleware(cmd.LongQueryDuration, func(ctx context.Context) string {
		if user := authn.ForContext(ctx); user != nil {
			return user.Name()
		}
		return ""
	}))

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	// GraphQL API
	graphqlServer, err := gql.NewServer()
	if err != nil {
		return err
	}
	if !cmd.DisableGraphql {
		// Mount with user permissions required
		r := chi.NewRouter()
		r.Mount("/", graphqlServer)
		root.Mount("/query", r)
	}

	// REST API
	if !cmd.DisableRest {
		restServer, err := rest.NewServer(graphqlServer)
		if err != nil {
			return err
		}
		r := chi.NewRouter()
		r.Mount("/", restServer)
		root.Mount("/rest", r)
	}

	// GraphQL Playground
	if cmd.EnablePlayground && !cmd.DisableGraphql {
		root.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

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
	go func() {
		srv.ListenAndServe()
	}()

	// Listen for shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	<-signalChan
	// Start http server shutdown with 5 second timeout
	// Run this in main thread so we block for shutdown to succeed
	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	return srv.Shutdown(gracefullCtx)
}

func getRedisOpts(v string) (*redis.Options, error) {
	a, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	if a.Scheme != "redis" {
		return nil, errors.New("redis URL must begin with redis://")
	}
	port := a.Port()
	if port == "" {
		port = "6379"
	}
	addr := fmt.Sprintf("%s:%s", a.Hostname(), port)
	dbNo := 0
	if len(a.Path) > 0 {
		var err error
		f := a.Path[1:len(a.Path)]
		dbNo, err = strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
	}
	return &redis.Options{Addr: addr, DB: dbNo}, nil
}
