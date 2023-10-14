package main

import (
	"context"
	"encoding/xml"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/0queue/mlb-rss/internal/cache"
	"github.com/0queue/mlb-rss/internal/mlb"
	"github.com/0queue/mlb-rss/internal/report"
	"github.com/0queue/mlb-rss/internal/rss"
	"github.com/0queue/mlb-rss/internal/tinycron"
	"github.com/0queue/mlb-rss/ui"
)

type config struct {
	JsonLog     bool
	Addr        string
	CheckAtHour int
	MyTeam      string
	Offseason   bool
}

func readConfigFromEnv() config {
	jsonLog := strings.ToLower(os.Getenv("JSON_LOG")) == "true"

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	checkAtHourRaw := os.Getenv("CHECK_AT_HOUR")
	if checkAtHourRaw == "" {
		checkAtHourRaw = "7"
	}
	checkAtHour, err := strconv.Atoi(checkAtHourRaw)
	if err != nil {
		checkAtHour = 7
	}
	if checkAtHour < 0 || checkAtHour > 23 {
		checkAtHour = 7
	}

	myTeam := os.Getenv("MY_TEAM")
	if myTeam == "" {
		myTeam = "BAL"
	}

	offseason := strings.ToLower(os.Getenv("OFFSEASON")) == "true"

	return config{
		JsonLog:     jsonLog,
		Addr:        addr,
		CheckAtHour: checkAtHour,
		MyTeam:      myTeam,
		Offseason:   offseason,
	}
}

func main() {
	// read config
	c := readConfigFromEnv()

	var handler slog.Handler
	if c.JsonLog {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	slog.Info("configuration successful", slog.Int("CHECK_AT_HOUR", c.CheckAtHour))

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
	tinycron.EveryDay(signalCtx, c.CheckAtHour, func() {
		if c.Offseason {
			slog.Info("No more baseball, go to sleep!")
			return
		}

		now := time.Now()
		slog.Info("Updating cache", slog.Time("now", now))

		r, err := rg.GenerateReport(now)
		if err != nil {
			slog.Error("Failed to generate report", slog.String("err", err.Error()))
			return
		}

		cache.Set(r)
	})

	// serve xml
	mux := http.NewServeMux()
	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		var items []rss.Item

		cachedReport, ok := cache.Get()
		if ok {
			rendered, err := rg.Render(cachedReport)
			if err != nil {
				slog.Error("Failed to render report", slog.String("err", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			items = append(items, rss.Item{
				Title: cachedReport.Headline,
				Link:  cachedReport.Link,
				Description: &rss.Description{
					Text: rendered,
				},
				// TODO investigate setting the Guid as prefix + hash(date + content)
				//      to enable iteration on content, and seeing results immediately
				//      after deploying and refreshing in miniflux
				Guid:    "mlb-rss-" + cachedReport.When.Format(report.BaseballTheaterTimeFormat),
				PubDate: cachedReport.When.Format(time.RFC822),
			})
		}

		feed := rss.Rss{
			Version: "2.0",
			Channel: rss.Channel{
				Title:       "MLB RSS",
				Link:        "https://baseball.theater",
				Description: "Feed generated from statsapi.mlb.com",
				Items:       items,
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if c.Offseason {
			w.Write([]byte("Offseason! 💤"))
			return
		}

		cachedReport, ok := cache.Get()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rendered, err := rg.RenderWeb(cachedReport)
		if err != nil {
			slog.Error("Failed to render web", slog.String("err", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "text/html")
		w.Write([]byte(rendered))
	})
	mux.HandleFunc("/favicon-32x32.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "image/png")
		w.Write(ui.Favicon)
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

	server.Shutdown(timeout)

	slog.Info("Shutdown finished")
}
