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
	gamePk            string
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