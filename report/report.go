package report

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"github.com/0queue/mlb-rss/mlb"
)

//go:embed content.gohtml
var embeddedTemplate string

type Report struct {
	Headline string
	Link     string
	Content  string
}

// Warning: this is a poorly defined function that happens
// to work when running in the morning in Germany
func daysBetween(a time.Time, b time.Time) int {
	return int(a.Sub(b).Hours() / 24)
}

func involvesTeam(g mlb.Game, teamId int) bool {
	return g.Teams.Away.Team.Id == teamId || g.Teams.Home.Team.Id == teamId
}

type upcomingInfo struct {
	IsHome      bool
	AgainstAbbr string
}

type postponeInfo struct {
	Where  string
	Reason string
}

type yesterdayInfo struct {
	Where       string
	WinningTeam mlb.Team
	LosingTeam  mlb.Team
}

// would be nice to use an either type but I'm lazy
type pastGame struct {
	PostponeInfo  *postponeInfo
	YesterdayInfo *yesterdayInfo
}

type futureGame struct {
	UpcomingInfo upcomingInfo
	RenderedTime string
}

type futureDay struct {
	DayAbbr     string
	FutureGames []futureGame
}

type render struct {
	Team             mlb.TeamFull
	BaseballTheater  string
	Yesterday        []pastGame
	Upcoming         [8]futureDay
	UpcomingTimezone string
}

func analyzePastGames(yesterdaysGames []mlb.Game) []pastGame {
	var analyzed []pastGame

	for _, g := range yesterdaysGames {
		var yesterday pastGame
		if g.Status.DetailedState == "Postponed" {
			yesterday.PostponeInfo = &postponeInfo{
				Where:  g.Venue.Name,
				Reason: g.Status.Reason,
			}
		} else {
			var winningTeam mlb.Team
			var losingTeam mlb.Team
			var where string
			if g.Teams.Away.IsWinner {
				where = "on the road"
				winningTeam = g.Teams.Away
				losingTeam = g.Teams.Home
			} else {
				where = "at home"
				winningTeam = g.Teams.Home
				losingTeam = g.Teams.Away
			}

			yesterday.YesterdayInfo = &yesterdayInfo{
				Where:       where,
				WinningTeam: winningTeam,
				LosingTeam:  losingTeam,
			}
		}
		analyzed = append(analyzed, yesterday)
	}

	return analyzed
}

func analyzeFutureGames(
	future [8][]mlb.Game,
	today time.Time,
	teams map[int]mlb.TeamFull,
	myTeam mlb.TeamFull,
) [8]futureDay {
	var futureDays [8]futureDay
	for i := 0; i < 8; i++ {
		gs := future[i]

		var futureGames []futureGame
		for _, g := range gs {
			var against mlb.Team
			if g.Teams.Home.Team.Id != myTeam.Id {
				against = g.Teams.Home
			} else {
				against = g.Teams.Away
			}

			upcomingInfo := upcomingInfo{
				IsHome:      g.Teams.Home.Team.Id == myTeam.Id,
				AgainstAbbr: teams[against.Team.Id].Abbreviation,
			}

			upcomingTime := g.GameDate.Local().Format("15:04")

			futureGames = append(futureGames, futureGame{
				UpcomingInfo: upcomingInfo,
				RenderedTime: upcomingTime,
			})
		}

		dayAbbr := today.AddDate(0, 0, i).Weekday().String()[:2]
		futureDays[i] = futureDay{
			DayAbbr:     dayAbbr,
			FutureGames: futureGames,
		}
	}
	return futureDays
}

func makeHeadline(yesterday []pastGame, today time.Time, myTeam mlb.TeamFull) string {

	var headline string
	if len(yesterday) == 0 {
		headline = fmt.Sprintf("Baseball report %s", today.Format("Monday 2006-01-02"))
	} else if len(yesterday) == 1 {
		y := yesterday[0]
		if y.PostponeInfo != nil {
			headline = fmt.Sprintf("Game was postponed due to %s", y.PostponeInfo.Reason)
		} else if y.YesterdayInfo != nil {
			g := y.YesterdayInfo
			if g.WinningTeam.Team.Id == myTeam.Id {
				headline = fmt.Sprintf("The %s win! %d to %d", myTeam.Name, g.WinningTeam.Score, g.LosingTeam.Score)
			} else {
				headline = fmt.Sprintf("The %s lose, %d to %d", myTeam.Name, g.LosingTeam.Score, g.WinningTeam.Score)
			}
		}
	} else {
		var winCount int
		var lossCount int
		for _, y := range yesterday {
			if y.YesterdayInfo != nil {
				if y.YesterdayInfo.WinningTeam.Team.Id == myTeam.Id {
					winCount += 1
				} else {
					lossCount += 1
				}
			}
		}

		headline = fmt.Sprintf("Doubleheader! The %s go %d - %d", myTeam.Name, winCount, lossCount)
	}

	return headline
}

func MakeReport(teams map[int]mlb.TeamFull, myTeam mlb.TeamFull, m mlb.Mlb, today time.Time) Report {
	yesterday := today.AddDate(0, 0, -1)
	var yesterdaysGame []mlb.Game
	var upcoming [8][]mlb.Game

	for _, date := range m.Dates {
		for _, game := range date.Games {
			if involvesTeam(game, myTeam.Id) {
				game := game

				if daysBetween(game.GameDate, yesterday) == 0 {
					yesterdaysGame = append(yesterdaysGame, game)
				} else if days := daysBetween(game.GameDate, today); days >= 0 && days < len(upcoming) {
					upcoming[days] = append(upcoming[days], game)
				}
			}
		}
	}

	pastGames := analyzePastGames(yesterdaysGame)
	futureDays := analyzeFutureGames(upcoming, today, teams, myTeam)
	baseballTheaterDate := yesterday.Format("20060102")
	upcomingTimezone, _ := today.Local().Zone()

	r := render{
		Team:             myTeam,
		BaseballTheater:  fmt.Sprintf("https://baseball.theater/games/%s", baseballTheaterDate),
		Yesterday:        pastGames,
		Upcoming:         futureDays,
		UpcomingTimezone: upcomingTimezone,
	}

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	t := template.Must(template.New("content").Funcs(funcMap).Parse(embeddedTemplate))

	var content bytes.Buffer
	err := t.Execute(&content, r)
	if err != nil {
		panic(err)
	}

	return Report{
		Headline: makeHeadline(pastGames, today, myTeam),
		Link:     r.BaseballTheater,
		Content:  content.String(),
	}
}
