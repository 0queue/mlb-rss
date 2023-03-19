alias b := build
alias r := run

build:
	go build -o bin/mlb-rss cmd/mlb-rss/main.go
	go build -o bin/fetch-mlb-data cmd/fetch-mlb-data/main.go

run: build
  bin/mlb-rss
