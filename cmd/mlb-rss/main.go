package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0queue/mlb-rss/internal/cache"
	"github.com/0queue/mlb-rss/internal/mlb"
	"github.com/0queue/mlb-rss/internal/report2"
	"github.com/caarlos0/env/v7"
	"github.com/go-co-op/gocron"
	"golang.org/x/exp/slog"
)

type config struct {
	JsonLog bool   `envDefault:"false"`
	Addr    string `envDefault:":8080"`
	Cron    string `envDefault:"0 7 * * *"`
}

func main() {
	// read config
	var c config
	opts := env.Options{
		UseFieldNameByDefault: true,
	}
	if err := env.Parse(&c, opts); err != nil {
		fmt.Printf("Failed to parse config: %s", err)
		os.Exit(1)
	}

	var handler slog.Handler
	if c.JsonLog {
		handler = slog.NewJSONHandler(os.Stdout)
	} else {
		handler = slog.NewTextHandler(os.Stdout)
	}
	slog.SetDefault(slog.New(handler))

	m, err := mlb.NewMlbClient("")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	// TODO finish instantiating the report generator
	//      which means a config for team abbr or name
	//      which searches all teams
	//      then generate a report
	//      then serve it as html!
	rg := report2.NewReportGenerator(m.Teams, time.Local)

	// seed cache
	cache := cache.Cache[report2.Report]{}
	cache.Set(report2.GenerateReport())

	// prepare shutdown channel
	// this signalCtx goes to the report generator
	// not the http server though, because it is already cancelled
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	// TODO figure out implications of local time
	// start refresh cron job
	cron := gocron.NewScheduler(time.Local)
	cron.Cron(c.Cron).Do(func() {
		fmt.Println("Doing the thing")
	})

	cron.StartAsync()

	// serve xml
	mux := http.NewServeMux()
	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := http.Server{
		Addr:    c.Addr,
		Handler: mux,
	}

	slog.Info("Starting http server", slog.String("addr", c.Addr))

	go func() {
		server.ListenAndServe()
	}()

	slog.Info("mlb-rss ready")
	<-signalCtx.Done()

	slog.Info("Shutting down")

	timeout, timeoutCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer timeoutCancel()

	cron.Stop()
	server.Shutdown(timeout)

	slog.Info("Shutdown finished")
}
