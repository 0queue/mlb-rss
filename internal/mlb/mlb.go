package mlb

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

var defaultEndpoint = "https://statsapi.mlb.com/api/v1/schedule/games"

//go:embed teams.json
var teamInfoEmbed []byte

type MlbClient struct {
	Teams map[int]TeamFull
	// endpoint is a string so we don't have to clone *url.URL just to send a request
	endpoint string
	client   http.Client
}

func NewMlbClient(endpoint string) (MlbClient, error) {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	_, err := url.Parse(endpoint)
	if err != nil {
		return MlbClient{}, err
	}

	var teamFullSlice []TeamFull
	err = json.Unmarshal(teamInfoEmbed, &teamFullSlice)
	if err != nil {
		return MlbClient{}, nil
	}

	teams := make(map[int]TeamFull)
	for _, t := range teamFullSlice {
		t := t
		teams[t.Id] = t
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	return MlbClient{
		Teams:    teams,
		endpoint: endpoint,
		client:   client,
	}, nil
}

// Download raw json
// mostly used to fetch test data
func (c *MlbClient) FetchRaw(start, end time.Time) ([]byte, error) {
	startDate := start.Format(time.DateOnly)
	endDate := end.Format(time.DateOnly)

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("sportId", "1")
	q.Set("startDate", startDate)
	q.Set("endDate", endDate)
	url.RawQuery = q.Encode()

	resp, err := c.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *MlbClient) Fetch(start, end time.Time) (Mlb, error) {
	raw, err := c.FetchRaw(start, end)
	if err != nil {
		return Mlb{}, err
	}

	var m Mlb
	err = json.Unmarshal(raw, &m)
	if err != nil {
		return Mlb{}, err
	}

	return m, nil
}

// TODO find out where I got the data, and make a function to download it
func (c *MlbClient) FetchTeamFull() {
	panic("not implemented yet")
}

type Mlb struct {
	Dates []Date
}

type Date struct {
	Date  string
	Games []Game
}

type Game struct {
	GamePk            int
	GameType          string
	Season            string
	GameDate          time.Time
	Status            Status
	Teams             Teams
	Venue             Venue
	GameNumber        int
	DoubleHeader      string
	DayNight          string
	Description       string
	GamesInSeries     int
	SeriesGameNumber  int
	SeriesDescription string
}

type Status struct {
	CodedGameState string
	DetailedState  string
	StartTimeTBD   bool
	Reason         string
}

type Teams struct {
	Away Team
	Home Team
}

type Team struct {
	LeagueRecord LeagueRecord
	Team         TeamInfo
	SeriesNumber int
	Score        int
	IsWinner     bool
}

type LeagueRecord struct {
	Wins   int
	Losses int
	Pct    string
}

type TeamInfo struct {
	Id   int
	Name string
	Link string
}

type Venue struct {
	Id   int
	Name string
	Link string
}

// not actually full...
type TeamFull struct {
	Id           int
	Name         string
	Link         string
	Venue        Venue
	Abbreviation string
	TeamName     string
	LocationName string
	ShortName    string
}
