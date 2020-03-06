package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ww24/calendar-notifier"
	"github.com/ww24/calendar-notifier/action"
	"github.com/ww24/calendar-notifier/config"
	"golang.org/x/sync/errgroup"
)

const (
	shutdownTimeout   = 30 * time.Second
	pollingInterval   = 1 * time.Minute
	calendarScanRange = 24 * time.Hour
)

var (
	confFile = flag.String("config", "", "set path to config (required)")
	cache    = newItemsCache()
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

	actionHandler, err := newActionHander(conf.Action)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
						c.Cancel()
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

				if ok := item.Register(now, newEventHandler(conf.Handler, actionHandler)); ok {
					log.Println("Register:", item.ID)
				}
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	srv := &http.Server{Addr: ":8080"}
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

func newEventHandler(handler map[string]config.EventHandler, actionHandler map[config.ActionName]action.Action) func(context.Context, *calendar.EventItem) {
	return func(ctx context.Context, e *calendar.EventItem) {
		name := strings.TrimSpace(e.Summary)
		eh, ok := handler[name]
		if !ok {
			log.Printf("skipped because handler not exists: %s\n", name)
			return
		}
		var ans []config.ActionName
		switch e.EventType {
		case calendar.Start:
			ans = eh.Start
		case calendar.End:
			ans = eh.End
		default:
			return
		}
		var g errgroup.Group
		defer g.Wait()
		for _, an := range ans {
			an := an
			g.Go(func() error {
				if err := actionHandler[an].Exec(ctx, e); err != nil {
					log.Printf("Exec error[%v]: %+v\n", an, err)
					return err
				}
				return nil
			})
		}
	}
}

func newActionHander(actions map[config.ActionName]config.Action) (map[config.ActionName]action.Action, error) {
	actionHandler := make(map[config.ActionName]action.Action, len(actions))
	for k, a := range actions {
		switch a.Type {
		case config.ActionHTTP:
			actionHandler[k] = action.NewHTTPAction(a.Header, a.Method, a.URL, a.Payload)
		case config.ActionPubSub:
			act, err := action.NewPubSubAction(a.Topic, a.Payload)
			if err != nil {
				return nil, err
			}
			actionHandler[k] = act
		}
	}
	return actionHandler, nil
}
