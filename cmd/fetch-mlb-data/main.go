package main

import (
	"fmt"
	"time"

	"github.com/0queue/mlb-rss/internal/mlb"
)

func main() {
	m, err := mlb.NewMlbClient()
	if err != nil {
		panic(err)
	}

	now := time.Now()

	raw, err := m.FetchRaw(now, now)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(raw))
}
