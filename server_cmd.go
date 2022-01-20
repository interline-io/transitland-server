package server

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/interline-io/transitland-server/rest"
	"github.com/interline-io/transitland-server/rtcache"
	"github.com/interline-io/transitland-server/workers"
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
	UseAuth          string
	auth.AuthConfig
	config.Config
}

func (cmd *Command) Parse(args []string) error {
	fl := flag.NewFlagSet("sync", flag.ExitOnError)
	fl.Usage = func() {
		log.Print("Usage: server")
		fl.PrintDefaults()
	}
	fl.StringVar(&cmd.DB.DBURL, "dburl", "", "Database URL (default: $TL_DATABASE_URL)")
	fl.StringVar(&cmd.RT.RedisURL, "redisurl", "localhost:6379", "Redis URL")
	fl.IntVar(&cmd.Timeout, "timeout", 60, "")
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.JwtAudience, "jwt-audience", "", "JWT Audience")
	fl.StringVar(&cmd.JwtIssuer, "jwt-issuer", "", "JWT Issuer")
	fl.StringVar(&cmd.JwtPublicKeyFile, "jwt-public-key-file", "", "Path to JWT public key file")
	fl.StringVar(&cmd.UseAuth, "auth", "", "")
	fl.StringVar(&cmd.GtfsDir, "gtfsdir", "", "Directory to store GTFS files")
	fl.StringVar(&cmd.GtfsS3Bucket, "s3", "", "S3 bucket for GTFS files")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.BoolVar(&cmd.DisableImage, "disable-image", false, "Disable image generation")
	fl.BoolVar(&cmd.DisableGraphql, "disable-graphql", false, "Disable GraphQL endpoint")
	fl.BoolVar(&cmd.DisableRest, "disable-rest", false, "Disable REST endpoint")
	fl.BoolVar(&cmd.EnablePlayground, "enable-playground", false, "Enable GraphQL playground")
	fl.BoolVar(&cmd.EnableProfiler, "enable-profile", false, "Enable profiling")
	fl.BoolVar(&cmd.EnableJobsApi, "enable-jobs-api", false, "Enable job api")
	fl.BoolVar(&cmd.EnableWorkers, "enable-workers", false, "Enable workers")
	fl.Parse(args)
	if cmd.DB.DBURL == "" {
		cmd.DB.DBURL = os.Getenv("TL_DATABASE_URL")
	}
	return nil
}

func (cmd *Command) Run() error {
	queueName := "tlv2-default"

	// Open database
	cfg := cmd.Config
	dbx := find.MustOpenDB(cfg.DB.DBURL)

	// Default finders and job queue
	var dbFinder model.Finder
	var rtFinder model.RTFinder
	var jobQueue workers.JobQueue
	dbFinder = find.NewDBFinder(dbx)
	rtFinder = rtcache.NewRTFinder(rtcache.NewLocalCache(), dbx)
	jobQueue = workers.NewLocalJobs()

	if cmd.Config.RT.RedisURL != "" {
		redisClient := redis.NewClient(&redis.Options{Addr: cfg.RT.RedisURL})
		rtFinder = rtcache.NewRTFinder((rtcache.NewRedisCache(redisClient)), dbx)
		workers.NewRedisJobs(redisClient, queueName)
	}

	// Setup CORS and logging
	root := mux.NewRouter()
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type", "apikey", "authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	root.Use(cors)
	root.Use(loggingMiddleware)

	// Setup user middleware
	userMiddleware, err := auth.GetUserMiddleware(cmd.UseAuth, cmd.AuthConfig)
	if err != nil {
		return err
	}
	root.Use(userMiddleware)

	// Workers
	if cmd.EnableJobsApi || cmd.EnableWorkers {
		// Open Redis
		jobWorkers := 1
		if cmd.EnableWorkers {
			jr, _ := workers.NewJobRunner(cfg, dbFinder, rtFinder, jobQueue, queueName, jobWorkers)
			go jr.RunWorkers()
		}
		if cmd.EnableJobsApi {
			jobServer, err := workers.NewServer(cfg, dbFinder, rtFinder, jobQueue, queueName, jobWorkers)
			if err != nil {
				return err
			}
			// Mount with admin permissions required
			mount(root, "/jobs", auth.AdminRequired(jobServer))
		}
	}

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
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
		mount(root, "/rest", restServer)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", "0.0.0.0", cmd.Port)
	fmt.Println("listening on:", addr)
	timeOut := time.Duration(cmd.Timeout)
	srv := &http.Server{
		Handler:      root,
		Addr:         addr,
		WriteTimeout: timeOut * time.Second,
		ReadTimeout:  timeOut * time.Second,
	}
	return srv.ListenAndServe()

}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting /query rewrite to /query/ to match subrouter's "/"
		if r.URL.Path == path {
			r.URL.Path = r.URL.Path + "/"
		}
		// Remove path prefix
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		handler.ServeHTTP(w, r)
	}))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
