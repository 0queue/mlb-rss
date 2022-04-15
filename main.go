package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

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

func generate(path string, endpoint string) {

	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	q := u.Query()
	q.Set("sportId", "1")
	q.Set("startDate", "2022-04-15")
	q.Set("endDate", "2022-04-22")
	u.RawQuery = q.Encode()

	fmt.Printf("Calling %s\n", u)

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

	fmt.Printf("Writing to %s\n", path)
	os.WriteFile(path, bytes, 0666)
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
				return errors.New("file does not exist")
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
