package server

import (
	"errors"
	"flag"
	"fmt"

	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/playground"
	"github.com/interline-io/transitland-server/internal/rtcache"
	"github.com/interline-io/transitland-server/internal/workers"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/interline-io/transitland-server/rest"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Command struct {
	Timeout          int
	Port             string
	DisableGraphql   bool
	DisableRest      bool
	EnablePlayground bool
	EnableJobsApi    bool
	EnableWorkers    bool
	EnableProfiler   bool
	EnableMetrics    bool
	UseAuth          string
	DefaultQueue     string
	SecretsFile      string
	auth.AuthConfig
	config.Config
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
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.JwtAudience, "jwt-audience", "", "JWT Audience")
	fl.StringVar(&cmd.JwtIssuer, "jwt-issuer", "", "JWT Issuer")
	fl.StringVar(&cmd.JwtPublicKeyFile, "jwt-public-key-file", "", "Path to JWT public key file")
	fl.StringVar(&cmd.UseAuth, "auth", "anon", "")
	fl.StringVar(&cmd.GtfsDir, "gtfsdir", "", "Directory to store GTFS files")
	fl.StringVar(&cmd.GtfsS3Bucket, "s3", "", "S3 bucket for GTFS files")
	fl.StringVar(&cmd.SecretsFile, "secrets", "", "DMFR file containing secrets")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.StringVar(&cmd.DefaultQueue, "queue", "tlv2-default", "Job queue name")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.BoolVar(&cmd.DisableImage, "disable-image", false, "Disable image generation")
	fl.BoolVar(&cmd.DisableGraphql, "disable-graphql", false, "Disable GraphQL endpoint")
	fl.BoolVar(&cmd.DisableRest, "disable-rest", false, "Disable REST endpoint")
	fl.BoolVar(&cmd.EnablePlayground, "enable-playground", false, "Enable GraphQL playground")
	fl.BoolVar(&cmd.EnableProfiler, "enable-profile", false, "Enable profiling")
	fl.BoolVar(&cmd.EnableMetrics, "enable-metrics", true, "Enable metrics endpoint for Prometheus")
	fl.BoolVar(&cmd.EnableJobsApi, "enable-jobs-api", false, "Enable job api")
	fl.BoolVar(&cmd.EnableWorkers, "enable-workers", false, "Enable workers")
	fl.Parse(args)
	if cmd.DBURL == "" {
		cmd.DBURL = os.Getenv("TL_DATABASE_URL")
	}
	if cmd.RedisURL == "" {
		cmd.RedisURL = os.Getenv("TL_REDIS_URL")
	}
	return nil
}

func (cmd *Command) Run() error {
	// Default finders and job queue
	var dbFinder model.Finder
	var rtFinder model.RTFinder
	var jobQueue jobs.JobQueue

	// Create Finder
	cfg := cmd.Config
	dbx := find.MustOpenDB(cfg.DBURL)
	dbFinder = find.NewDBFinder(dbx)

	// Create RTFinder
	var redisClient *redis.Client
	if cmd.RedisURL != "" {
		// Redis backed RTFinder
		rOpts, err := getRedisOpts(cfg.RedisURL)
		if err != nil {
			return err
		}
		redisClient = redis.NewClient(rOpts)
		// Replace RTFinder; use redis backed cache now
		rtFinder = rtcache.NewRTFinder(rtcache.NewRedisCache(redisClient), dbx)
		jobQueue = jobs.NewRedisJobs(redisClient, cmd.DefaultQueue)
	} else {
		// Default to in-memory cache
		rtFinder = rtcache.NewRTFinder(rtcache.NewLocalCache(), dbx)
		jobQueue = jobs.NewLocalJobs()
	}

	root := mux.NewRouter()

	// Setup user middleware
	userMiddleware, err := auth.GetUserMiddleware(cmd.UseAuth, cmd.AuthConfig)
	if err != nil {
		return err
	}
	root.Use(userMiddleware)

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	if cmd.EnableMetrics {
		// TODO: turn on when meaningful metrics added
		// metrics.RecordPromMetrics()
		root.Handle("/metrics", promhttp.Handler())
	}

	// GraphQL API
	graphqlServer, err := resolvers.NewServer(cfg, dbFinder, rtFinder)
	if err != nil {
		return err
	}
	if !cmd.DisableGraphql {
		// Mount with user permissions required
		mount(root, "/query", auth.UserRequired(graphqlServer))
	}

	// GraphQL Playground
	if cmd.EnablePlayground && !cmd.DisableGraphql {
		root.Handle("/", playground.Handler("GraphQL playground", "/query/"))
	}

	// REST API
	if !cmd.DisableRest {
		restServer, err := rest.NewServer(cfg, graphqlServer)
		if err != nil {
			return err
		}
		mount(root, "/rest", auth.UserRequired(restServer))
	}

	// Workers
	if cmd.EnableJobsApi || cmd.EnableWorkers {
		// Load secrets
		var secrets []tl.Secret
		if v := cmd.SecretsFile; v != "" {
			rr, err := dmfr.LoadAndParseRegistry(v)
			if err != nil {
				return errors.New("unable to load secrets file")
			}
			secrets = rr.Secrets
		}
		// Start workers/api
		jobWorkers := 10
		jobOptions := jobs.JobOptions{
			Logger:   log.Logger,
			JobQueue: jobQueue,
			Finder:   dbFinder,
			RTFinder: rtFinder,
			Secrets:  secrets,
		}
		if cmd.EnableWorkers {
			log.Print("enabling workers")
			jobQueue.AddWorker(workers.GetWorker, jobOptions, jobWorkers)
			go jobQueue.Run()
		}
		if cmd.EnableJobsApi {
			log.Print("enabling jobs api")
			jobServer, err := workers.NewServer(cfg, cmd.DefaultQueue, jobWorkers, jobOptions)
			if err != nil {
				return err
			}
			// Mount with admin permissions required
			mount(root, "/jobs", auth.AdminRequired(jobServer))
		}
	}

	// Setup CORS and logging
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type", "apikey", "authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	root.Use(cors)
	root.Use(loggingMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%s", "0.0.0.0", cmd.Port)
	log.Infof("listening on: %s", addr)
	timeOut := time.Duration(cmd.Timeout)
	srv := &http.Server{
		Handler:      root,
		Addr:         addr,
		WriteTimeout: timeOut * time.Second,
		ReadTimeout:  timeOut * time.Second,
	}
	return srv.ListenAndServe()
}
