package main

import (
	"errors"
	"fmt"
	"net/http"
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

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
