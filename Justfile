set positional-arguments

build:
	go build -o build/mlb-rss cmd/mlb-rss/main.go

run *args='': build
	build/mlb-rss $@