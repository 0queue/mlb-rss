package report2

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/0queue/mlb-rss/internal/mlb"
	"github.com/0queue/mlb-rss/ui"
	"golang.org/x/exp/slog"
)

type ReportGenerator struct {
	MyTeam   mlb.TeamFull
	AllTeams map[int]mlb.TeamFull
	Location time.Location
}

func NewReportGenerator(myTeam mlb.TeamFull, allTeams map[int]mlb.TeamFull, loc time.Location) ReportGenerator {
	return ReportGenerator{
		MyTeam:   myTeam,
		AllTeams: allTeams,
		Location: loc,
	}
}

// PastGame is used by past-game.html.tpl
type PastGame struct {
	PostponeReason string
	Venue          mlb.Venue
	IsWinnerHome   bool
	W              mlb.Team
	L              mlb.Team
}

type FutureGame struct {
	// GameTimeLocal is when the game is on in the rss feed's timezone
	GameTimeLocal string
	IsMyTeamHome  bool
	AgainstAbbr   string
}

type Yesterday struct {
	MyTeam          mlb.Team
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
}

// Do analysis of the games and generate the report. NO TEMPLATES
// assumes the first Date is yesterday, and the rest are the future
// ultimately, analysis consists of filtering
func (rg *ReportGenerator) GenerateReport(m mlb.Mlb, today time.Time) Report {

	// so we get an array of stuff which has a date attached
	// which means I should ignore time completely
	dates := filterMyTeam(m.Dates, rg.MyTeam.Id)
	pastGames := analyzePastGames(dates[0], rg.MyTeam.Id)
	futureGames := rg.analyzeFutureGames(today, dates[1:])

	yesterday := Yesterday{
		MyTeam:          mlb.Team{},
		PastGames:       pastGames,
		BaseballTheater: "",
	}

	tz, _ := today.Local().Zone()

	upcoming := Upcoming{
		FutureDays: futureGames,
		Timezone:   tz,
	}

	baseballTheaterDate := today.AddDate(0, 0, -1).Format("20060102")
	link := fmt.Sprintf("https://baseball.theater/games/%s", baseballTheaterDate)

	headline := rg.generateHeadline(pastGames, today)

	return Report{
		Yesterday: yesterday,
		Upcoming:  upcoming,
		Headline:  headline,
		Link:      link,
	}
}

// Render uses templates to render reports to html
func (r *Report) Render() string {
	t := template.Must(template.ParseFS(ui.ReportTemplates))
	var content bytes.Buffer
	t.ExecuteTemplate(&content, "report2", r)
	return content.String()
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

		var isWinnerHome = g.Teams.Home.IsWinner
		var winner mlb.Team
		var loser mlb.Team

		if isWinnerHome {
			winner = g.Teams.Home
			loser = g.Teams.Away
		} else {
			winner = g.Teams.Away
			loser = g.Teams.Away
		}

		p := PastGame{
			PostponeReason: g.Status.Reason,
			Venue:          g.Venue,
			IsWinnerHome:   isWinnerHome,
			W:              winner,
			L:              loser,
		}

		pastGames = append(pastGames, p)
	}

	return pastGames
}

func (rg *ReportGenerator) analyzeFutureGames(today time.Time, dates []mlb.Date) [8]FutureDay {
	var futureGames [8]FutureDay

	if len(dates) != 8 {
		slog.Warn("Number of future dates not as expected", slog.Int("expected", 8), slog.Int("actual", len(dates)))
	}

	for i, d := range dates {
		games := make([]FutureGame, 0)

		for _, g := range d.Games {

			isHome := g.Teams.Home.Team.Id == rg.MyTeam.Id
			var opponentTeam mlb.Team
			if isHome {
				opponentTeam = g.Teams.Away
			} else {
				opponentTeam = g.Teams.Home
			}

			futureGame := FutureGame{
				GameTimeLocal: g.GameDate.In(&rg.Location).Format("15:04"),
				IsMyTeamHome:  isHome,
				AgainstAbbr:   rg.AllTeams[opponentTeam.Team.Id].Abbreviation,
			}

			games = append(games, futureGame)
		}

		dayAbbr := today.AddDate(0, 0, i).Weekday().String()[:2]

		futureGames[i] = FutureDay{
			DayAbbr: dayAbbr,
			Games:   games,
		}
	}

	return futureGames
}

func (rg *ReportGenerator) generateHeadline(pastGames []PastGame, today time.Time) string {
	// 1. no games
	// 2. Postpone
	// 3. team wins! 1 of 1
	// 4. team loses :( 1 of 1
	// 5. Double header

	switch len(pastGames) {
	case 0:
		return fmt.Sprintf("Baseball report %s", today.Format("Monday 2006-01-02"))
	case 1:
		var headline string
		g := pastGames[0]
		if g.PostponeReason != "" {
			headline = fmt.Sprintf("Game was postponed due to %s", g.PostponeReason)
		} else if g.W.Team.Id == rg.MyTeam.Id {
			headline = fmt.Sprintf("The %s win! %d to %d", rg.MyTeam.Name, g.W.Score, g.L.Score)
		} else {
			headline = fmt.Sprintf("The %s lose, %d to %d", rg.MyTeam.Name, g.L.Score, g.W.Score)
		}
		return headline
	case 2:
		var winCount int
		var lossCount int
		for _, g := range pastGames {
			if g.W.Team.Id == rg.MyTeam.Id {
				winCount += 1
			} else {
				lossCount += 1
			}
		}
		return fmt.Sprintf("Doubleheader! The %s go %d - %d", rg.MyTeam.Name, winCount, lossCount)
	default:
		slog.Warn("Failed to generateHeadline, too many games played", slog.Int("pastGamesCount", len(pastGames)))
		return "error"
	}
}
