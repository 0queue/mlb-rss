alias b := build
alias r := run

v := "1.3.1"

export KO_DOCKER_REPO := "ghcr.io/0queue/mlb-rss"

build:
	go build -o bin/mlb-rss cmd/mlb-rss/main.go

run: build
	bin/mlb-rss

# visit localhost:8000 to use a nice json viewer
serve-test-data:
	python3 -m http.server -d test/data/

fetch-team-data:
	curl 'https://statsapi.mlb.com/api/v1/teams?sportId=1' | jq -r > internal/mlb/teams.json

ko:
	ko build --bare --tags={{v}} ./cmd/mlb-rss

install-ko:
	go install github.com/google/ko@latest