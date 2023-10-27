package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Config struct {
	port int
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	frontendURL string
}

type application struct {
	config   Config
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

	app.infoLog.Printf("Starting Invoice microserice running on port %d", app.config.port)
	return srv.ListenAndServe()
}

func main() {
	var err error
	var cfg Config

	if err = godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	flag.IntVar(&cfg.port, "port", 5000, "Server port to listen on")
	flag.Parse()

	cfg.smtp.host = os.Getenv("SMTP_HOST")
	cfg.smtp.port = 587
	cfg.smtp.username = os.Getenv("SMTP_USERNAME")
	cfg.smtp.password = os.Getenv("SMTP_PASSWORD")
	cfg.frontendURL = os.Getenv("FRONTEND_URL")

	app := &application{
		config:   cfg,
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		version:  version,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Fatal(err)
	}
}
