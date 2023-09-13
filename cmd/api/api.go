package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"webapp/internal/driver"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting API server in %s mode on port %d", app.config.env, app.config.port)
	return srv.ListenAndServe()
}

func main() {
	var err error
	var cfg config

	if err = godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment (development|production|maitenance)")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")
	cfg.db.dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&tls=false",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Println("Database connection successful")
	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}
	err = app.serve()
	errorLog.Fatal(err)
}
