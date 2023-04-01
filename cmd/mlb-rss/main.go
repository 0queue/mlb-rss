package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0queue/mlb-rss/internal/cache"
	"github.com/0queue/mlb-rss/internal/mlb"
	"github.com/0queue/mlb-rss/internal/report"
	"github.com/0queue/mlb-rss/internal/rss"
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

	mc, err := mlb.NewMlbClient()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	myTeam, ok := mc.FindTeam(c.MyTeam)
	if !ok {
		slog.Error("Failed to find team", slog.String("team", c.MyTeam))
		os.Exit(1)
	}
	rg := report.NewReportGenerator(myTeam.Id, mc, time.Local)

	// seed cache
	cache := cache.Cache[report.Report]{}
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
		//res, err := mc.FetchSchedule(now.AddDate(0, 0, -1), now.AddDate(0, 0, 7), rg.MyTeamId)
		//if err != nil {
		//	slog.Error("Failed to fetch latest information", slog.String("err", err.Error()))
		//}

		r, err := rg.GenerateReport(now)
		if err != nil {
			slog.Error("Failed to generate report", slog.String("err", err.Error()))
			return
		}

		cache.Set(r)
	})

	cron.StartAsync()

	// serve xml
	mux := http.NewServeMux()
	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		cachedReport, ok := cache.Get()
		if !ok {
			slog.Warn("Cache not populated yet")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rendered, err := rg.Render(cachedReport)
		if err != nil {
			slog.Error("Failed to render report", slog.String("err", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		feed := rss.Rss{
			Version: "2.0",
			Channel: rss.Channel{
				Title:       "MLB RSS",
				Link:        "https://baseball.theater",
				Description: "Feed generated from statsapi.mlb.com",
				Items: []rss.Item{
					{
						Title: cachedReport.Headline,
						Link:  cachedReport.Link,
						Description: &rss.Description{
							Text: rendered,
						},
						Guid:    "mlb-rss-" + cachedReport.When.Format(report.BaseballTheaterTimeFormat),
						PubDate: cachedReport.When.Format(time.RFC822),
					},
				},
			},
		}

		// should probably just cache the xml
		bytes, err := xml.MarshalIndent(feed, "", " ")
		if err != nil {
			slog.Error("Failed to marshal rss feed", slog.String("err", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/rss+xml")
		w.Write(bytes)
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
