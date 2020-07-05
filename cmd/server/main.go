package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ww24/calendar-notifier/interface/config"
)

const (
	shutdownTimeout   = 30 * time.Second
	pollingInterval   = 1 * time.Minute
	calendarScanRange = 24 * time.Hour
	defaultPort       = "8080"
)

var (
	confFile = flag.String("config", "", "set path to config (required)")
	port     = os.Getenv("PORT")
)

func main() {
	flag.Parse()
	if *confFile == "" {
		fmt.Fprintln(os.Stderr, "-config flag is required")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf, err := config.Parse(*confFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config loading error: %+v\n", err)
		os.Exit(1)
	}
	log.Printf("Config loaded: v%s\n", conf.Version)
	log.Println("Running mode:", conf.Mode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := initialize(ctx, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialize Error: %+v\n", err)
		os.Exit(1)
	}
	srv := app.server(port)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	go func() {
		if err := app.worker(ctx); err != nil {
			panic(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-sigCh
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Shutdown Error:", err)
	}
}
