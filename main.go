package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/0queue/mlb-rss/mlb"
	"github.com/spf13/cobra"
)

//go:embed report.html.gotpl
var tpl string

var now time.Time = time.Now()

type Report struct {
	MyTeam    string
	Yesterday *Yesterday
}

type Yesterday struct {
	Outcome        string
	MyTeamScore    int
	OtherTeamScore int
}

func DateEqual(a time.Time, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()

	return ya == yb && ma == mb && da == db
}

func serve(addr string, path string, contentType string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		bytes, err := os.ReadFile(path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", contentType)
		w.Write(bytes)
	})

	fmt.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}

func generateReport(game *mlb.Game) Report {

	var yesterday *Yesterday = nil

	if game != nil {

		var myTeam mlb.Team
		var otherTeam mlb.Team
		var outcome string
		if game.Teams.Away.Team.Id == 133 {
			myTeam = game.Teams.Away
			otherTeam = game.Teams.Home
			if game.Teams.Away.IsWinner {
				outcome = "won"
			} else {
				outcome = "lost"
			}
		} else {
			myTeam = game.Teams.Home
			otherTeam = game.Teams.Away
			if game.Teams.Home.IsWinner {
				outcome = "won"
			} else {
				outcome = "lost"
			}
		}

		yesterday = &Yesterday{
			Outcome:        outcome,
			MyTeamScore:    myTeam.Score,
			OtherTeamScore: otherTeam.Score,
		}
	}

	return Report{
		MyTeam:    "team",
		Yesterday: yesterday,
	}
}

func process(bytes []byte) string {
	type JsonObject map[string]any

	var m mlb.Mlb
	json.Unmarshal(bytes, &m)

	yesterday := now.AddDate(0, 0, -1)

	var yesterdaysGame *mlb.Game = nil

	for _, date := range m.Dates {
		for _, game := range date.Games {

			if DateEqual(yesterday, game.GameDate) {
				if game.Teams.Away.Team.Id == 133 || game.Teams.Home.Team.Id == 133 {
					report := generateReport(&game)
					t, err := template.New("report").Parse(tpl)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					t.Execute(os.Stdout, report)
				}
				yesterdaysGame = &game
			}
		}
	}

	if yesterdaysGame != nil {
		fmt.Printf("Found yesterday's game: %v\n", *yesterdaysGame)
	}

	return "" //string(s)
}

func generate(path string, endpoint string) {

	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	yesterday := now.AddDate(0, 0, -1)
	nextWeek := now.AddDate(0, 0, 7)
	baseballTheaterDate := yesterday.Format("20060102")
	startDate := yesterday.Format("2006-01-02")
	endDate := nextWeek.Format("2006-01-02")

	q := u.Query()
	q.Set("sportId", "1")
	q.Set("startDate", startDate)
	q.Set("endDate", endDate)
	u.RawQuery = q.Encode()

	fmt.Printf("Calling %s\n", u)

	fmt.Printf("Yesterday's game: https://baseball.theater/games/%s\n", baseballTheaterDate)

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(process(bytes))
	//fmt.Printf("Writing to %s\n", path)
	//os.WriteFile(path, bytes, 0666)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mlb-rss",
		Short: "mlb-rss is a feed generator for the official MLB api",
	}

	var serveAddr string
	var serveContentType string
	var serveCmd = &cobra.Command{
		Use:   "serve [FILE]",
		Short: "serve a single file over HTTP",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a path to a file")
			}

			if _, err := os.Stat(args[0]); errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Warning: %s does not exist\n", args[0])
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			abspath, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			serve(serveAddr, abspath, serveContentType)
		},
	}

	serveCmd.Flags().StringVarP(&serveContentType, "content-type", "", "application/rss+xml", "Content-Type to serve the file as")
	serveCmd.Flags().StringVarP(&serveAddr, "addr", "", ":8080", "Interface and port to listen on")

	var generateEndpoint string
	var generateCmd = &cobra.Command{
		Use:   "generate [FILE]",
		Short: "Update the RSS feed in FILE or overwrite with a new feed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			abspath, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			generate(abspath, generateEndpoint)
		},
	}

	generateCmd.Flags().StringVarP(&generateEndpoint, "endpoint", "", "https://statsapi.mlb.com/api/v1/schedule/games", "Endpoint to make requests to")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
