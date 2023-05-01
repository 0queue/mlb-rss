package report

import (
	"time"

	"github.com/0queue/mlb-rss/internal/mlb"
)

// PastGame is used by past-game.html.tpl
type PastGame struct {
	PostponeReason   string
	Venue            mlb.Venue
	IsWinnerHome     bool
	W                mlb.GameTeam
	L                mlb.GameTeam
	CondensedGameUrl string
	HasLinescore     bool
	Linescore        Linescore
}

type Linescore struct {
	Home LinescoreTeam
	Away LinescoreTeam
}

type LinescoreTeam struct {
	Abbr    string
	Innings []int
	Runs    int
	Hits    int
	Errors  int
}

type FutureGame struct {
	// GameTimeLocal is when the game is on in the rss feed's timezone
	GameTimeLocal string
	IsMyTeamHome  bool
	AgainstAbbr   string
}

type Yesterday struct {
	MyTeamName      string
	PastGames       []PastGame
	BaseballTheater string
}

type FutureDay struct {
	// Sun, Mon, Tue, etc
	DayAbbr string
	Games   []FutureGame
}

type Upcoming struct {
	FutureDays [8]FutureDay
	// Timezone is the rss feed's timezone
	Timezone string
}

type Report struct {
	Yesterday Yesterday
	Upcoming  Upcoming
	Headline  string
	Link      string
	When      time.Time
}
