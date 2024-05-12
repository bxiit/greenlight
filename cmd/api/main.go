package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/bxiit/greenlight/internal/jsonlog"
	"github.com/bxiit/greenlight/internal/mailer"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bxiit/greenlight/internal/data"
	// underscore (alias) is used to avoid go compiler complaining or erasing this
	// library.
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// Add a Db struct field to hold the configuration settings for our database connection
// pool. For now this only holds the DSN, which we will read in from a command-line flag.
type Config struct {
	Port int
	Env  string
	Db   struct {
		Dsn          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	Limiter struct {
		Enabled bool
		Rps     float64
		Burst   int
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

type application struct {
	config Config
	logger *jsonlog.Logger
	models data.Models // hold new models in App
	mailer mailer.Mailer
	wg     sync.WaitGroup
	gormDB *gorm.DB
}

type ApplicationX struct {
	mailer mailer.Mailer
	wg     sync.WaitGroup
	db     *sql.DB
	models data.Models // hold new models in App
}

func main() {

	viper.SetConfigFile("./config.json")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading Config file, %s", err)
		return
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
		return
	}

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	// Db will be closed before main function is completed.
	defer db.Close()

	gormDB, err := OpenGDB(db)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // data.NewModels() function to initialize a Models struct
		mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
		gormDB: gormDB,
	}

	go app.checkAndResendActivation()

	// Use the httprouter instance returned by App.routes() as the server handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"Env":  cfg.Env,
	})
	// reuse defined variable err

	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.Db.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Db.MaxOpenConns)

	duration, err := time.ParseDuration(cfg.Db.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// optional lifetime limit, to use this, uncomment Db substruct field and corresponding flag stringvar
	// lifetime, err := time.ParseDuration(cfg.Db.MaxIdleTime)
	// if err != nil {
	// 	return nil, err
	// }
	// Db.SetConnMaxLifetime(lifetime)
	//context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx) // create a connection and verify that everything is set up correctly.

	if err != nil {
		return nil, err
	}

	return db, nil
}

func OpenGDB(db *sql.DB) (*gorm.DB, error) {
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return gdb, nil
}
