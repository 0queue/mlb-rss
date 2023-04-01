package mlb

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

var apiEndpoint = "https://statsapi.mlb.com/api/v1"

//go:embed teams.json
var teamInfoEmbed []byte

type MlbClient struct {
	AllTeams map[int]Team
	client   http.Client
}

func NewMlbClient() (*MlbClient, error) {
	var teamFullSlice struct {
		Teams []Team
	}
	err := json.Unmarshal(teamInfoEmbed, &teamFullSlice)
	if err != nil {
		return nil, err
	}

	teams := make(map[int]Team)
	for _, t := range teamFullSlice.Teams {
		t := t
		teams[t.Id] = t
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	return &MlbClient{
		AllTeams: teams,
		client:   client,
	}, nil
}

// Download raw json
// mostly used to fetch test data
// if start (date) == end (date), only fetches data for that day
func (mc *MlbClient) FetchScheduleRaw(start, end time.Time, teamId int) ([]byte, error) {
	startDate := start.Format(time.DateOnly)
	endDate := end.Format(time.DateOnly)

	u, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "schedule")

	q := u.Query()
	q.Set("sportId", "1")
	q.Set("sportId", strconv.Itoa(teamId))
	q.Set("startDate", startDate)
	q.Set("endDate", endDate)
	u.RawQuery = q.Encode()

	resp, err := mc.client.Get(u.String())
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

func (mc *MlbClient) FetchSchedule(start, end time.Time, teamId int) (Schedule, error) {
	raw, err := mc.FetchScheduleRaw(start, end, teamId)
	if err != nil {
		return Schedule{}, err
	}

	var m Schedule
	err = json.Unmarshal(raw, &m)
	if err != nil {
		return Schedule{}, err
	}

	return m, nil
}

// TODO find out where I got the data, and make a function to download it
// https://statsapi.mlb.com/api/v1/teams?sportId=1
func (mc *MlbClient) FetchTeamFull() {
	panic("not implemented yet")
}

// FindTeam searches for a team based on the abbreviation if q is three letters,
// or as a substring of the full name
func (mc *MlbClient) FindTeam(q string) (Team, bool) {
	var found Team
	var ok bool
	for _, t := range mc.AllTeams {
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
