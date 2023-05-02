package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/gbfsfinder"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/interline-io/transitland-server/internal/metrics"
	"github.com/interline-io/transitland-server/internal/playground"
	"github.com/interline-io/transitland-server/internal/rtfinder"
	"github.com/interline-io/transitland-server/internal/workers"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/interline-io/transitland-server/rest"
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
	AuthMiddlewares   ArrayFlags
	DefaultQueue      string
	SecretsFile       string
	metersConfig
	metricsConfig
	auth.AuthConfig
	config.Config
}

type metricsConfig struct {
	EnableMetrics   bool
	MetricsProvider string
}

type metersConfig struct {
	EnableMetering         bool
	MeteringProvider       string
	MeteringAmberfloConfig string
}

func (cmd *Command) Parse(args []string) error {
	fl := flag.NewFlagSet("sync", flag.ExitOnError)
	fl.Usage = func() {
		log.Print("Usage: server")
		fl.PrintDefaults()
	}
	fl.StringVar(&cmd.DBURL, "dburl", "", "Database URL (default: $TL_DATABASE_URL)")
	fl.StringVar(&cmd.RedisURL, "redisurl", "", "Redis URL (default: $TL_REDIS_URL)")
	fl.IntVar(&cmd.Timeout, "timeout", 60, "")
	fl.IntVar(&cmd.LongQueryDuration, "long-query", 1000, "Log queries over this duration (ms)")
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.Storage, "storage", "", "Static storage backend")
	fl.StringVar(&cmd.RTStorage, "rt-storage", "", "RT storage backend")
	fl.StringVar(&cmd.SecretsFile, "secrets", "", "DMFR file containing secrets")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.StringVar(&cmd.DefaultQueue, "queue", "tlv2-default", "Job queue name")

	fl.Var(&cmd.AuthMiddlewares, "auth", "Add one or more auth middlewares")
	fl.StringVar(&cmd.AuthConfig.DefaultUsername, "default-username", "", "Default user name (for --auth=admin)")
	fl.StringVar(&cmd.AuthConfig.JwtAudience, "jwt-audience", "", "JWT Audience (use with -auth=jwt)")
	fl.StringVar(&cmd.AuthConfig.JwtIssuer, "jwt-issuer", "", "JWT Issuer (use with -auth=jwt)")
	fl.StringVar(&cmd.AuthConfig.JwtPublicKeyFile, "jwt-public-key-file", "", "Path to JWT public key file (use with -auth=jwt)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperEndpoint, "gatekeeper-endpoint", "", "Gatekeeper endpoint (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperRoleSelector, "gatekeeper-selector", "", "Gatekeeper role selector (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperExternalIDSelector, "gatekeeper-eid-selector", "", "Gatekeeper External ID selector (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.GatekeeperParam, "gatekeeper-param", "", "Gatekeeper param (use with -auth=gatekeeper)")
	fl.BoolVar(&cmd.AuthConfig.GatekeeperAllowError, "gatekeeper-allow-error", false, "Gatekeeper ignore errors (use with -auth=gatekeeper)")
	fl.StringVar(&cmd.AuthConfig.UserHeader, "user-header", "", "Header to check for username (use with -auth=header)")

	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.BoolVar(&cmd.DisableImage, "disable-image", false, "Disable image generation")
	fl.BoolVar(&cmd.DisableGraphql, "disable-graphql", false, "Disable GraphQL endpoint")
	fl.BoolVar(&cmd.DisableRest, "disable-rest", false, "Disable REST endpoint")
	fl.BoolVar(&cmd.EnablePlayground, "enable-playground", false, "Enable GraphQL playground")
	fl.BoolVar(&cmd.EnableProfiler, "enable-profile", false, "Enable profiling")

	// Metrics
	// fl.BoolVar(&cmd.EnableMetrics, "enable-metrics", false, "Enable metrics")
	fl.StringVar(&cmd.MetricsProvider, "metrics-provider", "", "Specify metrics provider")

	// Metering
	// fl.BoolVar(&cmd.EnableMetering, "enable-metering", false, "Enable metering")
	fl.StringVar(&cmd.MeteringProvider, "metering-provider", "", "Use metering provider")
	fl.StringVar(&cmd.MeteringAmberfloConfig, "metering-amberflo-config", "", "Use provided config for AmberFlo metering")

	// Jobs
	fl.BoolVar(&cmd.EnableJobsApi, "enable-jobs-api", false, "Enable job api")
	fl.BoolVar(&cmd.EnableWorkers, "enable-workers", false, "Enable workers")

	// Admin
	fl.BoolVar(&cmd.EnableAdminApi, "enable-admin-api", false, "Enable admin api")

	fl.Parse(args)
	if cmd.MetricsProvider != "" {
		cmd.EnableMetrics = true
	}
	if cmd.MeteringProvider != "" {
		cmd.EnableMetering = true
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
	// Default finders and job queue
	cfg := cmd.Config

	var dbFinder model.Finder
	var rtFinder model.RTFinder
	var gbfsFinder model.GbfsFinder
	var jobQueue jobs.JobQueue

	// Open metrics
	var metricProvider metrics.MetricProvider
	metricProvider = metrics.NewDefaultMetric()
	if cmd.EnableMetrics {
		if cmd.MetricsProvider == "prometheus" {
			metricProvider = metrics.NewPromMetrics()
		}
	}

	// Open metering
	var meterProvider meters.MeterProvider
	meterProvider = meters.NewDefaultMeter()
	if cmd.EnableMetering {
		if cmd.MeteringProvider == "amberflo" {
			a := meters.NewAmberFlo(os.Getenv("AMBERFLO_APIKEY"), 30*time.Second, 100)
			if cmd.MeteringAmberfloConfig != "" {
				if err := a.LoadConfig(cmd.MeteringAmberfloConfig); err != nil {
					return err
				}
			}
			meterProvider = a
		}
		defer meterProvider.Close()
	}

	// Open database
	var db sqlx.Ext
	dbx := find.MustOpenDB(cfg.DBURL)
	db = dbx
	if log.Logger.GetLevel() == zerolog.TraceLevel {
		db = find.LogDB(dbx)
	}

	// Open redis
	var redisClient *redis.Client
	if cmd.RedisURL != "" {
		// Redis backed RTFinder
		rOpts, err := getRedisOpts(cfg.RedisURL)
		if err != nil {
			return err
		}
		redisClient = redis.NewClient(rOpts)
	}

	// Create Finder
	dbFinder = find.NewDBFinder(db)

	// Create RTFinder
	if cmd.RedisURL != "" {
		// Use redis backed finders
		rtFinder = rtfinder.NewFinder(rtfinder.NewRedisCache(redisClient), db)
		gbfsFinder = gbfsfinder.NewFinder(redisClient)
		jobQueue = jobs.NewRedisJobs(redisClient, cmd.DefaultQueue)
	} else {
		// Default to in-memory cache
		rtFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
		gbfsFinder = gbfsfinder.NewFinder(nil)
		jobQueue = jobs.NewLocalJobs()
	}

	// Setup router
	root := chi.NewRouter()
	root.Use(middleware.RequestID)
	root.Use(middleware.RealIP)
	root.Use(middleware.Recoverer)
	root.Use(middleware.StripSlashes)

	// Setup user middleware
	for _, k := range cmd.AuthMiddlewares {
		if userMiddleware, err := auth.GetUserMiddleware(k, cmd.AuthConfig, redisClient); err != nil {
			return err
		} else {
			root.Use(userMiddleware)
		}
	}

	// Add logging middleware - must be after auth
	root.Use(loggingMiddleware(cmd.LongQueryDuration))

	// Setup CORS
	root.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"content-type", "apikey", "authorization"},
		AllowCredentials: true,
	}))

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	// Metrics
	if cmd.EnableMetrics {
		root.Handle("/metrics", metricProvider.MetricsHandler())
	}

	// GraphQL API
	graphqlServer, err := resolvers.NewServer(cfg, dbFinder, rtFinder, gbfsFinder)
	if err != nil {
		return err
	}
	if !cmd.DisableGraphql {
		// Mount with user permissions required
		r := chi.NewRouter()
		r.Use(metrics.WithMetric(metricProvider.NewApiMetric("graphql")))
		r.Use(meters.WithMeter(meterProvider, "graphql", 1.0, nil))
		r.Use(auth.UserRequired)
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
		r.Use(auth.UserRequired)
		r.Mount("/", restServer)
		root.Mount("/rest", r)
	}

	// Admin API
	if cmd.EnableAdminApi {
		adminServer, err := authz.NewServer(dbFinder)
		if err != nil {
			return err
		}
		r := chi.NewRouter()
		r.Use(auth.UserRequired)
		r.Mount("/", adminServer)
		root.Mount("/admin", r)
	}

	// Workers
	if cmd.EnableJobsApi || cmd.EnableWorkers {
		// Start workers/api
		jobWorkers := 10
		jobOptions := jobs.JobOptions{
			Logger:     log.Logger,
			JobQueue:   jobQueue,
			Finder:     dbFinder,
			RTFinder:   rtFinder,
			GbfsFinder: gbfsFinder,
			Config:     cfg,
		}
		// Add metrics
		jobQueue.Use(metrics.NewJobMiddleware("", metricProvider.NewJobMetric("default")))
		if cmd.EnableWorkers {
			log.Infof("Enabling job workers")
			jobQueue.AddWorker(workers.GetWorker, jobOptions, jobWorkers)
			go jobQueue.Run()
		}
		if cmd.EnableJobsApi {
			log.Infof("Enabling job api")
			jobServer, err := workers.NewServer(cfg, cmd.DefaultQueue, jobWorkers, jobOptions)
			if err != nil {
				return err
			}
			// Mount with admin permissions required
			r := chi.NewRouter()
			r.Use(auth.AdminRequired)
			r.Mount("/", jobServer)
			root.Mount("/jobs", r)
		}
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

// ArrayFlags allow repeatable command line options.
// https://stackoverflow.com/questions/28322997/how-to-get-a-list-of-values-into-a-flag-in-golang/28323276#28323276
type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
