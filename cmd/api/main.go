package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	// import the pq driver so that it can register itelf with the databasse/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used
	_ "github.com/lib/pq"
	"github.com/ynrfin/greenlight/internal/data"
	"github.com/ynrfin/greenlight/internal/jsonlog"
	"github.com/ynrfin/greenlight/internal/mailer"
)

// Declare a string containing the application version number. later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

// Define config struct to hold all  the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environmoent for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command line flags when the application starts
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	// Add a new limiter struct containing fields for the request-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	cors struct {
		trustedOrigins []string
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middlewares. At the moment this only contains a copy of the config struct and a
// logger. But it will grow to include a lot more as our build progress.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// Declare an instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct. We
	// default using the port number 4000 and the environment "development" if no
	// corresponding flag are provide

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSNif no flag is provided.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connection")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max open connection")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max open connection")

	// Flag for rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "12f0eb8d562ae2", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "61f2b321892765", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.ynrfin.com", "SMTP sender")
	flag.Func("cors-trusted-origins", "Trusted CORS origins(space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	flag.Parse()

	// Initialize a new logger which writes message to the standard out stream,
	// prefixed with the current date and time.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call the openDB() helper function (see below) to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit
	// application immediately
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	// Defer a call to db.Close() so that the connection pool is closed before the
	//  main() function exists.
	defer db.Close()

	// Also log a message to say taht the connection pool has been successfully
	// established
	logger.PrintInfo("database connection pool established", nil)

	// Publish a new "version" variable in the expvar handler containing our application
	// version number.
	expvar.NewString("version").Set(version)

	// Publish the number of active goroutines
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Publish the databse connection pool statuses.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// Publish current unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModel(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()

	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// The openDB() function return a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5 second timeout deadiline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connectio to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this wil return an
	// error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}
