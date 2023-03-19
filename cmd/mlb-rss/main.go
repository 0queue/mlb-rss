package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0queue/mlb-rss/internal/cache"
	"github.com/0queue/mlb-rss/internal/report2"
	"github.com/caarlos0/env/v7"
	"github.com/go-co-op/gocron"
	"golang.org/x/exp/slog"
)

type config struct {
	Json bool   `envDefault:"false"`
	Addr string `envDefault:":8080"`
	Cron string `envDefault:"0 7 * * *"`
}

func main() {
	// read config
	var c config
	opts := env.Options{
		UseFieldNameByDefault: true,
	}
	if err := env.Parse(&c, opts); err != nil {
		fmt.Printf("Failed to parse config: %w", err)
		os.Exit(1)
	}

	var handler slog.Handler
	if c.Json {
		handler = slog.NewJSONHandler(os.Stdout)
	} else {
		handler = slog.NewTextHandler(os.Stdout)
	}
	slog.SetDefault(slog.New(handler))

	// seed cache
	cache := cache.Cache[report2.Report2]{}
	cache.Set(report2.GenerateReport())

	// TODO figure out implications of local time
	// start refresh cron job
	s := gocron.NewScheduler(time.Local)
	s.Cron(c.Cron).Do(func() {
		fmt.Println("Doing the thing")
	})

	s.StartAsync()

	// serve xml
	http.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	slog.Info("Starting http server", slog.String("addr", c.Addr))
	http.ListenAndServe(c.Addr, nil)
}
