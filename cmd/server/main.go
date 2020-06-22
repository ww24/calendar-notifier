package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ww24/calendar-notifier"
	"github.com/ww24/calendar-notifier/config"
)

const (
	shutdownTimeout   = 30 * time.Second
	pollingInterval   = 1 * time.Minute
	calendarScanRange = 24 * time.Hour
	defaultPort       = "8080"
)

var (
	confFile = flag.String("config", "", "set path to config (required)")
	cache    = newItemsCache()
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

	var launchHandler func() (map[string]interface{}, error)
	switch conf.Mode {
	case config.ModeResident:
		launchHandler = launchResident(ctx, conf)
	case config.ModeOnDemand:
		launchHandler = launchOnDemand(ctx, conf)
	}

	handler := newHandler(conf, launchHandler)
	if port == "" {
		port = defaultPort
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler.Handler(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println("Server Error:", err)
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

func launchResident(ctx context.Context, conf config.Config) func() (map[string]interface{}, error) {
	actionHandler, err := newActionHander(conf.Action)
	if err != nil {
		panic(err)
	}
	go func() {
		cal := calendar.NewCalendar(conf.CalendarID)
		for range calendar.ImmediateTick(ctx, pollingInterval) {
			now := time.Now()
			items, err := cal.Events(ctx, now, now.Add(calendarScanRange))
			if err != nil {
				panic(err)
			}

			filteredItems := make([]*calendar.EventItem, 0, len(items))
			m := cache.Get()
			for _, e := range items {
				if c, ok := m[e.ID]; ok {
					if e.UpdatedAt.After(c.UpdatedAt) {
						// cached item is outdated
						c.CancelSchedule()
						filteredItems = append(filteredItems, e)
					} else {
						// cached item is up to date
						filteredItems = append(filteredItems, c)
					}
					continue
				}
				// item is not cached
				filteredItems = append(filteredItems, e)
			}
			cache.SetFromList(filteredItems)

			for _, item := range filteredItems {
				log.Printf("ID: %s, Summary: %s, Start: %s, End: %s\n",
					item.ID, item.Summary, item.StartAt, item.EndAt)

				item.Exec(now, newExecutorHandler(conf.Handler, actionHandler))
				if ok := item.Schedule(now, newRegistratorHandler(conf.Handler, actionHandler)); ok {
					log.Println("Scheduled:", item.ID)
				}
			}
			// TODO: cleanup
		}
	}()
	return func() (map[string]interface{}, error) {
		return nil, errors.New("launch handler is unavailable if running mode is resident")
	}
}

func launchOnDemand(ctx context.Context, conf config.Config) func() (map[string]interface{}, error) {
	actionHandler, err := newActionHander(conf.Action)
	if err != nil {
		panic(err)
	}
	cal := calendar.NewCalendar(conf.CalendarID)
	return func() (map[string]interface{}, error) {
		now := time.Now()
		items, err := cal.Events(ctx, now, now.Add(calendarScanRange))
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			log.Printf("ID: %s, Summary: %s, Start: %s, End: %s\n",
				item.ID, item.Summary, item.StartAt, item.EndAt)

			item.Exec(now, newExecutorHandler(conf.Handler, actionHandler))
		}
		// TODO: cleanup

		return map[string]interface{}{"status": "maybe ok"}, nil
	}
}
