package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"
	"strconv"
	"strings"
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
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-mw/auth/ancheck"
	"github.com/interline-io/transitland-mw/auth/azcheck"
	"github.com/interline-io/transitland-mw/lmw"
	"github.com/interline-io/transitland-mw/meters"
	"github.com/interline-io/transitland-mw/metrics"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/playground"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/interline-io/transitland-server/server/rest"
	"github.com/interline-io/transitland-server/workers"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Command struct {
	Timeout           int
	Port              string
	LongQueryDuration int
	DisableGraphql    bool
	DisableRest       bool
	EnablePlayground  bool
	EnableAdminApi    bool
	EnableJobsApi     bool
	EnableWorkers     bool
	EnableProfiler    bool
	EnableRateLimits  bool
	LoadAdmins        bool
	QueuePrefix       string
	SecretsFile       string
	AuthMiddlewares   arrayFlags
	metersConfig      meters.Config
	metricsConfig     metrics.Config
	AuthConfig        ancheck.AuthConfig
	CheckerConfig     azcheck.CheckerConfig
	config.Config
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
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
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

	// Auth config
	fl.Var(&cmd.AuthMiddlewares, "auth", "Add one or more auth middlewares")
	fl.StringVar(&cmd.AuthConfig.DefaultUsername, "default-username", "", "Default user name (for --auth=admin)")
	fl.StringVar(&cmd.AuthConfig.JwtAudience, "jwt-audience", "", "JWT Audience (use with -auth=jwt)")
	fl.StringVar(&cmd.AuthConfig.JwtIssuer, "jwt-issuer", "", "JWT Issuer (use with -auth=jwt)")
	fl.StringVar(&cmd.AuthConfig.JwtPublicKeyFile, "jwt-public-key-file", "", "Path to JWT public key file (use with -auth=jwt)")
	fl.BoolVar(&cmd.AuthConfig.JwtUseEmailAsId, "jwt-use-email-as-id", false, "JWT use claim email as user id")
	fl.StringVar(&cmd.AuthConfig.GatekeeperEndpoint, "gatekeeper-endpoint", "", "Gatekeeper endpoint (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperRoleSelector, "gatekeeper-selector", "", "Gatekeeper role selector (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperExternalIDSelector, "gatekeeper-eid-selector", "", "Gatekeeper External ID selector (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperParam, "gatekeeper-param", "", "Gatekeeper param (use with -auth=gatekeeper)")
	fl.BoolVar(&cmd.AuthConfig.GatekeeperAllowError, "gatekeeper-allow-error", false, "Gatekeeper ignore errors (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.UserHeader, "user-header", "", "Header to check for username (use with -auth=header)")

	// Admin api
	fl.StringVar(&cmd.CheckerConfig.GlobalAdmin, "global-admin", "", "Global admin user")
	fl.StringVar(&cmd.CheckerConfig.Auth0ClientID, "auth0-client-id", "", "Auth0 client ID")
	fl.StringVar(&cmd.CheckerConfig.Auth0ClientSecret, "auth0-client-secret", "", "Auth0 client secret")
	fl.StringVar(&cmd.CheckerConfig.Auth0Domain, "auth0-domain", "", "Auth0 domain")
	fl.StringVar(&cmd.CheckerConfig.Auth0Connection, "auth0-connection", "", "Auth0 Connection")
	fl.StringVar(&cmd.CheckerConfig.FGAEndpoint, "fga-endpoint", "", "FGA endpoint")
	fl.StringVar(&cmd.CheckerConfig.FGAStoreID, "fga-store-id", "", "FGA store")
	fl.StringVar(&cmd.CheckerConfig.FGAModelID, "fga-model-id", "", "FGA model")
	fl.StringVar(&cmd.CheckerConfig.FGALoadModelFile, "fga-load-model-file", "", "")

	// Metrics
	// fl.BoolVar(&cmd.EnableMetrics, "enable-metrics", false, "Enable metrics")
	fl.StringVar(&cmd.metricsConfig.MetricsProvider, "metrics-provider", "", "Specify metrics provider")

	// Metering
	// fl.BoolVar(&cmd.EnableMetering, "enable-metering", false, "Enable metering")
	fl.BoolVar(&cmd.EnableRateLimits, "enable-rate-limits", false, "Enable rate limits")
	fl.StringVar(&cmd.metersConfig.MeteringProvider, "metering-provider", "", "Use metering provider")
	fl.StringVar(&cmd.metersConfig.MeteringAmberfloConfig, "metering-amberflo-config", "", "Use provided config for Amberflo metering")

	// Jobs
	fl.BoolVar(&cmd.EnableJobsApi, "enable-jobs-api", false, "Enable job api")
	fl.BoolVar(&cmd.EnableWorkers, "enable-workers", false, "Enable workers")

	// Admin
	fl.BoolVar(&cmd.EnableAdminApi, "enable-admin-api", false, "Enable admin api")

	fl.Parse(args)
	if cmd.metricsConfig.MetricsProvider != "" {
		cmd.metricsConfig.EnableMetrics = true
	}
	if cmd.metersConfig.MeteringProvider != "" {
		cmd.metersConfig.EnableMetering = true
	}

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
	cmd.Config.Secrets = secrets
	return nil
}

func (cmd *Command) Run() error {
	cfg := cmd.Config

	// Open database
	var db sqlx.Ext
	dbx, err := dbutil.OpenDB(cfg.DBURL)
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
		rOpts, err := getRedisOpts(cfg.RedisURL)
		if err != nil {
			return err
		}
		redisClient = redis.NewClient(rOpts)
	}

	// Setup authorization checker
	var checker model.Checker
	checker, err = azcheck.NewCheckerFromConfig(cmd.CheckerConfig, db)
	if err != nil {
		return err
	}

	// Create Finder
	var dbFinder model.Finder
	f := dbfinder.NewFinder(db, checker)
	if cmd.LoadAdmins {
		f.LoadAdmins()
	}
	dbFinder = f

	// Create RTFinder, GBFSFinder
	var rtFinder model.RTFinder
	var gbfsFinder model.GbfsFinder
	var jobQueue jobs.JobQueue
	if redisClient != nil {
		// Use redis backed finders
		rtFinder = rtfinder.NewFinder(rtfinder.NewRedisCache(redisClient), db)
		gbfsFinder = gbfsfinder.NewFinder(redisClient)
		jobQueue = jobs.NewRedisJobs(redisClient, cmd.QueuePrefix)
	} else {
		// Default to in-memory cache
		rtFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
		gbfsFinder = gbfsfinder.NewFinder(nil)
		jobQueue = jobs.NewLocalJobs()
	}

	// Setup metrics
	metricProvider, err := metrics.GetProvider(cmd.metricsConfig)
	if err != nil {
		return err
	}

	// Setup metering
	meterProvider, err := meters.GetProvider(cmd.metersConfig)
	if err != nil {
		return err
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

	// Setup user middleware
	for _, k := range cmd.AuthMiddlewares {
		if userMiddleware, err := ancheck.GetUserMiddleware(k, cmd.AuthConfig, redisClient); err != nil {
			return err
		} else {
			root.Use(userMiddleware)
		}
	}

	// Add logging middleware - must be after auth
	root.Use(lmw.LoggingMiddleware(cmd.LongQueryDuration))

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	// Metrics
	if cmd.metricsConfig.EnableMetrics {
		root.Handle("/metrics", metricProvider.MetricsHandler())
	}

	// GraphQL API
	graphqlServer, err := gql.NewServer(cfg, dbFinder, rtFinder, gbfsFinder, checker)
	if err != nil {
		return err
	}
	if !cmd.DisableGraphql {
		// Mount with user permissions required
		r := chi.NewRouter()
		r.Use(metrics.WithMetric(metricProvider.NewApiMetric("graphql")))
		r.Use(meters.WithMeter(meterProvider, "graphql", 1.0, nil))
		r.Use(ancheck.UserRequired)
		r.Mount("/", graphqlServer)
		root.Mount("/query", r)
	}

	// REST API
	if !cmd.DisableRest {
		restServer, err := rest.NewServer(cfg, graphqlServer)
		if err != nil {
			return err
		}
		r := chi.NewRouter()
		r.Use(metrics.WithMetric(metricProvider.NewApiMetric("rest")))
		r.Use(meters.WithMeter(meterProvider, "rest", 1.0, nil))
		r.Use(ancheck.UserRequired)
		r.Mount("/", restServer)
		root.Mount("/rest", r)
	}

	// GraphQL Playground
	if cmd.EnablePlayground && !cmd.DisableGraphql {
		root.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	// Admin API
	if cmd.EnableAdminApi {
		adminServer, err := azcheck.NewServer(checker)
		if err != nil {
			return err
		}
		r := chi.NewRouter()
		r.Use(ancheck.UserRequired)
		r.Mount("/", adminServer)
		root.Mount("/admin", r)
	}

	// Workers
	if cmd.EnableJobsApi || cmd.EnableWorkers {
		// Start workers/api
		jobWorkers := 8
		jobOptions := jobs.JobOptions{
			Logger:     log.Logger,
			JobQueue:   jobQueue,
			Finder:     dbFinder,
			RTFinder:   rtFinder,
			GbfsFinder: gbfsFinder,
			Config:     cfg,
		}
		// Add metrics
		// jobQueue.Use(metrics.NewJobMiddleware("", metricProvider.NewJobMetric("default")))
		if cmd.EnableWorkers {
			log.Infof("Enabling job workers")
			jobQueue.AddWorker("default", workers.GetWorker, jobOptions, jobWorkers)
			jobQueue.AddWorker("rt-fetch", workers.GetWorker, jobOptions, jobWorkers)
			jobQueue.AddWorker("static-fetch", workers.GetWorker, jobOptions, jobWorkers)
			jobQueue.AddWorker("gbfs-fetch", workers.GetWorker, jobOptions, jobWorkers)
			go jobQueue.Run()
		}
		if cmd.EnableJobsApi {
			log.Infof("Enabling job api")
			jobServer, err := workers.NewServer(cfg, "", jobWorkers, jobOptions)
			if err != nil {
				return err
			}
			// Mount with admin permissions required
			r := chi.NewRouter()
			r.Use(ancheck.AdminRequired)
			r.Mount("/", jobServer)
			root.Mount("/jobs", r)
		}
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

// arrayFlags allow repeatable command line options.
// https://stackoverflow.com/questions/28322997/how-to-get-a-list-of-values-into-a-flag-in-golang/28323276#28323276
type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
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
