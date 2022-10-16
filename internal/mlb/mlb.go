package mlb

import "time"

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
