package mlb

import "time"

type Content struct {
	Highlights struct {
		Highlights struct {
			Items []Highlight
		}
	}
}

type Highlight struct {
	Type            string
	State           string
	Date            time.Time
	Id              string
	Headline        string
	Slug            string
	Blurb           string
	KeywordsAll     []Keyword
	MediaPlaybackId string
	Title           string
	Description     string
	Duration        string // TODO parse?
	Playbacks       []Playback
}

type Keyword struct {
	Type        string
	Value       string
	DisplayName string
}

type Playback struct {
	Name string
	Url  string
}
