package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// mlb-rss serve <filename> -content-type="..." -listen="0.0.0.0:80"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mlb-rss",
		Short: "mlb-rss is a feed generator for the official MLB api",
		//Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("running")
		//},
	}

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
			fmt.Printf("serving... %v\n", args)
			fmt.Printf("Content-Type: %s\n", serveContentType)
		},
	}

	serveCmd.Flags().StringVarP(&serveContentType, "content-type", "", "application/rss+xml", "Content-Type to serve the file as")

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
