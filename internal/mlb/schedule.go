package mlb

import "time"

type Schedule struct {
	Dates []Date
}

type Date struct {
	// TODO maybe custom time.Time unmarshaller?
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
	Away GameTeam
	Home GameTeam
}

type GameTeam struct {
	LeagueRecord LeagueRecord
	Team         TeamSummary
	SeriesNumber int
	Score        int
	IsWinner     bool
}

type LeagueRecord struct {
	Wins   int
	Losses int
	Pct    string
}

type TeamSummary struct {
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
type Team struct {
	Id           int
	Name         string
	Link         string
	Venue        Venue
	Abbreviation string
	TeamName     string
	LocationName string
	ShortName    string
}
