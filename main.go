package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	configPath := flag.String("config", "config.json", "configuration file")
	listen := flag.String("listen", "0.0.0.0:8000", "HTTP listen address")
	flag.Parse()
	c, err := loadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", c.dsn())
	if err != nil {
		log.Fatal("could not configure PostgreSQL")
	}
	defer db.Close()
	db.SetMaxOpenConns(c.DBMaxConnections)
	db.SetMaxIdleConns(c.DBMinCached)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		log.Fatal("could not connect to PostgreSQL; check database settings")
	}
	a := newAPI(&postgresStore{db}, c)
	a.e.Server.ReadHeaderTimeout = 10 * time.Second
	stop, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-stop.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.e.Shutdown(ctx); err != nil {
			log.Print("HTTP shutdown timed out")
		}
	}()
	log.Printf("http_mail listening on %s", *listen)
	if err := a.e.Start(*listen); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-shutdownDone
}
