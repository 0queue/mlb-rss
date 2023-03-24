package mlb

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var scheduleEndpoint = "https://statsapi.mlb.com/api/v1/schedule/games"

//go:embed teams.json
var teamInfoEmbed []byte

type MlbClient struct {
	Teams  map[int]TeamFull
	client http.Client
}

func NewMlbClient() (*MlbClient, error) {
	var teamFullSlice struct {
		Teams []TeamFull
	}
	err := json.Unmarshal(teamInfoEmbed, &teamFullSlice)
	if err != nil {
		return nil, err
	}

	teams := make(map[int]TeamFull)
	for _, t := range teamFullSlice.Teams {
		t := t
		teams[t.Id] = t
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	return &MlbClient{
		Teams:  teams,
		client: client,
	}, nil
}

// Download raw json
// mostly used to fetch test data
// if start (date) == end (date), only fetches data for that day
func (c *MlbClient) FetchRaw(start, end time.Time) ([]byte, error) {
	startDate := start.Format(time.DateOnly)
	endDate := end.Format(time.DateOnly)

	u, err := url.Parse(scheduleEndpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("sportId", "1")
	q.Set("startDate", startDate)
	q.Set("endDate", endDate)
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
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
// https://statsapi.mlb.com/api/v1/teams?sportId=1
func (c *MlbClient) FetchTeamFull() {
	panic("not implemented yet")
}

// SearchTeams searches for a team based on the abbreviation if q is three letters,
// or as a substring of the full name
func (c *MlbClient) SearchTeams(q string) (TeamFull, bool) {
	var found TeamFull
	var ok bool
	for _, t := range c.Teams {
		t := t
		if len(q) == 3 && strings.ToLower(t.Abbreviation) == strings.ToLower(q) {
			found = t
			ok = true
			break
		} else if strings.Contains(strings.ToLower(t.Name), strings.ToLower(q)) {
			found = t
			ok = true
			break
		}
	}

	return found, ok
}

type Mlb struct {
	Dates []Date
}

type Date struct {
	Date  string
	Games []Game
}

// TODO hydrate videos and pass all the way up
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
