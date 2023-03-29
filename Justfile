alias b := build
alias r := run

v := "1.0.1"

build:
	go build -o bin/mlb-rss cmd/mlb-rss/main.go
	go build -o bin/fetch-mlb-data cmd/fetch-mlb-data/main.go

run: build
	bin/mlb-rss

# visit localhost:8000 to use a nice json viewer
serve-test-data:
	python3 -m http.server -d test/data/

fetch-team-data:
	curl 'https://statsapi.mlb.com/api/v1/teams?sportId=1' | jq -r > internal/mlb/teams.json

build-container-image:
	docker build -t ghcr.io/0queue/mlb-rss:{{v}} .

push-container-image: build-container-image
	docker push ghcr.io/0queue/mlb-rss:{{v}}