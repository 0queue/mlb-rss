package report2

import (
	"github.com/0queue/mlb-rss/internal/mlb"
	"golang.org/x/exp/slog"
)

type ReportGenerator struct {
	MyTeam mlb.TeamFull
}

func NewReportGenerator(myTeam mlb.TeamFull) ReportGenerator {
	return ReportGenerator{
		MyTeam: myTeam,
	}
}

// PastGame can be either postponed, or have a winner and loser
// represents a postponed game if PostponeReason is not ""
type PastGame struct {
	Venue          mlb.Venue
	PostponeReason string
	IsHome         bool
	// TODO change to winning and losing
	MyTeam       mlb.Team
	OpponentTeam mlb.Team
}

// PastGame2 is used by past-game.html.tpl
type PastGame2 struct {
	PostponeReason string
	Venue          mlb.Venue
	IsWinnerHome   bool
	W              mlb.Team
	L              mlb.Team
}

type Yesterday struct {
	MyTeam          mlb.Team
	PastGames       []PastGame2
	BaseballTheater string
}

type FutureDay struct {
	// Sun, Mon, Tue, etc
	DayAbbr string
	Games   []FutureGame2
}

type FutureGame2 struct {
	// GameTimeLocal is when the game is on in the rss feed's timezone
	GameTimeLocal string
	IsMyTeamHome  bool
	AgainstAbbr   string
}

type Upcoming struct {
	FutureDays [8]FutureDay
	// Timezone is the rss feed's timezone
	Timezone string
}

type FutureGame struct {
	IsHome       bool
	MyTeam       mlb.Team
	OpponentTeam mlb.Team
}

type Report struct {
	// team I'm interested in
	MyTeam mlb.TeamFull

	// Yesterday can contain 0..n games
	Yesterday []PastGame

	// Future is analyzed 8 days in advance including today
	// and each day may contain 0..n games
	Future [8][]FutureGame
}

type Report2 struct {
	Yesterday Yesterday
	Upcoming  Upcoming
	// TODO find a better way to do headlines
	Headline string
	Link     string
}

// Do analysis of the games and generate the report. NO RENDERING!
// assumes the first Date is yesterday, and the rest are the future
// ultimately, analysis consists of filtering and labelling interest
func (g *ReportGenerator) GenerateReport(m mlb.Mlb) Report {

	// so we get an array of stuff which has a date attached
	// which means I should ignore time completely
	dates := filterMyTeam(m.Dates, g.MyTeam.Id)
	pastGames := analyzePastGames(dates[0], g.MyTeam.Id)
	futureGames := analyzeFutureGames(dates[1:], g.MyTeam.Id)

	// TODO change the analyses to generate report structs directly

	return Report{
		MyTeam:    g.MyTeam,
		Yesterday: pastGames,
		Future:    futureGames,
	}
}

// keep games involving the team with the given id
func filterMyTeam(dates []mlb.Date, id int) []mlb.Date {
	newDates := make([]mlb.Date, 0)
	for _, d := range dates {
		newGames := make([]mlb.Game, 0)
		for _, g := range d.Games {
			if g.Teams.Home.Team.Id == id || g.Teams.Away.Team.Id == id {
				newGames = append(newGames, g)
			}
		}

		newDates = append(newDates, mlb.Date{
			Date:  d.Date,
			Games: newGames,
		})
	}

	return newDates
}

func analyzePastGames(date mlb.Date, id int) []PastGame {
	pastGames := make([]PastGame, 0)

	slog.Info("Analyzing yesterday's game", slog.String("date", date.Date))

	for _, g := range date.Games {

		isHome := g.Teams.Home.Team.Id == id
		var myTeam mlb.Team
		var opponentTeam mlb.Team

		if isHome {
			myTeam = g.Teams.Home
			opponentTeam = g.Teams.Away
		} else {
			myTeam = g.Teams.Away
			opponentTeam = g.Teams.Home
		}

		p := PastGame{
			Venue:          g.Venue,
			PostponeReason: g.Status.Reason,
			IsHome:         isHome,
			MyTeam:         myTeam,
			OpponentTeam:   opponentTeam,
		}

		pastGames = append(pastGames, p)
	}

	return pastGames
}

func analyzeFutureGames(dates []mlb.Date, id int) [8][]FutureGame {
	var futureGames [8][]FutureGame

	if len(dates) != 8 {
		slog.Warn("Number of future dates not as expected", slog.Int("expected", 8), slog.Int("actual", len(dates)))
	}

	for i, d := range dates {
		day := make([]FutureGame, 0)

		for _, g := range d.Games {

			isHome := g.Teams.Home.Team.Id == id
			var myTeam mlb.Team
			var opponentTeam mlb.Team
			if isHome {
				myTeam = g.Teams.Home
				opponentTeam = g.Teams.Away
			} else {
				myTeam = g.Teams.Away
				opponentTeam = g.Teams.Home
			}

			futureGame := FutureGame{
				IsHome:       isHome,
				MyTeam:       myTeam,
				OpponentTeam: opponentTeam,
			}

			day = append(day, futureGame)
		}

		futureGames[i] = day
	}

	return futureGames
}
