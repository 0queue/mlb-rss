package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0queue/mlb-rss/mlb"
	"github.com/0queue/mlb-rss/report"
	"github.com/spf13/cobra"
)

//go:embed teams.json
var embeddedTeamJson []byte

var now time.Time = time.Now()

func readEmbeddedTeams() map[int]mlb.TeamFull {
	type root struct {
		Teams []mlb.TeamFull
	}
	var embeddedTeams root
	err := json.Unmarshal(embeddedTeamJson, &embeddedTeams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	teams := make(map[int]mlb.TeamFull)
	for _, t := range embeddedTeams.Teams {
		t := t
		teams[t.Id] = t
	}

	return teams
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

func findTeam(teams map[int]mlb.TeamFull, teamFragment string) mlb.TeamFull {
	var found *mlb.TeamFull
	for _, t := range teams {
		t := t
		if strings.Contains(strings.ToLower(t.Name), strings.ToLower(teamFragment)) {
			found = &t
			break
		}
	}

	if found == nil {
		fmt.Printf("Failed to find team with fragment '%s'\n", teamFragment)
		os.Exit(1)
	}

	return *found
}

func getMlb(endpoint string) mlb.Mlb {
	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	yesterday := now.AddDate(0, 0, -1)
	nextWeek := now.AddDate(0, 0, 7)
	startDate := yesterday.Format("2006-01-02")
	endDate := nextWeek.Format("2006-01-02")

	q := u.Query()
	q.Set("sportId", "1")
	q.Set("startDate", startDate)
	q.Set("endDate", endDate)
	u.RawQuery = q.Encode()

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

	var m mlb.Mlb
	// golang was poorly designed
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return m
}

func generate(path string, endpoint string, teamFragment string) {

	m := getMlb(endpoint)
	teams := readEmbeddedTeams()
	myTeam := findTeam(teams, teamFragment)
	r := report.MakeReport(teams, myTeam, m, now)
	fmt.Printf("headline: %s\n", r.Headline)
	fmt.Printf("content: %s\n", r.Content)

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
	var generateTeamFragment string
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
			generate(abspath, generateEndpoint, generateTeamFragment)
		},
	}

	generateCmd.Flags().StringVarP(&generateEndpoint, "endpoint", "", "https://statsapi.mlb.com/api/v1/schedule/games", "Endpoint to make requests to")
	generateCmd.Flags().StringVarP(&generateTeamFragment, "team", "", "orioles", "The team you root for")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
