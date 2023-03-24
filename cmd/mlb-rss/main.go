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
	MyTeam  string `envDefault:"BAL"`
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

	m, err := mlb.NewMlbClient()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	myTeam, ok := m.SearchTeams(c.MyTeam)
	if !ok {
		slog.Error("Failed to find team", slog.String("team", c.MyTeam))
		os.Exit(1)
	}
	rg := report2.NewReportGenerator(myTeam, m.Teams, time.Local)

	// seed cache
	cache := cache.Cache[report2.Report]{}
	// prepare shutdown channel
	// this signalCtx goes to the report generator
	// not the http server though, because it is already cancelled
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	// start refresh cron job
	cron := gocron.NewScheduler(time.Local)
	_, _ = cron.Cron(c.Cron).StartImmediately().Do(func() {
		now := time.Now()
		slog.Info("Updating cache", slog.Time("now", now))
		res, err := m.Fetch(now.AddDate(0, 0, -1), now.AddDate(0, 0, 7))
		if err != nil {
			slog.Error("Failed to fetch latest information", slog.String("err", err.Error()))
		}
		cache.Set(rg.GenerateReport(res, now))
	})

	cron.StartAsync()

	// serve xml
	mux := http.NewServeMux()
	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		report, ok := cache.Get()
		if !ok {
			slog.Warn("Cache not populated yet")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rendered, err := rg.Render(report)
		if err != nil {
			slog.Error("Failed to render report", slog.String("err", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><head><meta charset="utf-8"></head><html><body>`))
		w.Write([]byte(fmt.Sprintf(`<h2>%s</h2>`, report.Headline)))
		w.Write([]byte(rendered))
		w.Write([]byte(`</body></html>`))
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
