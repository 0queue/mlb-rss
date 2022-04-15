package report

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/0queue/mlb-rss/mlb"
)

//go:embed content.tpl
var embeddedTemplate string

type Report struct {
	Headline string
	Content  string
}

func daysBetween(a time.Time, b time.Time) int {
	return int(a.Sub(b).Hours() / 24)
}

func involvesTeam(g mlb.Game, teamId int) bool {
	return g.Teams.Away.Team.Id == teamId || g.Teams.Home.Team.Id == teamId
}

func MakeReport(teams map[int]mlb.TeamFull, myTeam mlb.TeamFull, m mlb.Mlb, today time.Time) Report {
	yesterday := today.AddDate(0, 0, -1)
	var yesterdaysGame *mlb.Game
	var upcoming [8]*mlb.Game

	for _, date := range m.Dates {
		for _, game := range date.Games {
			if involvesTeam(game, myTeam.Id) {
				game := game

				if daysBetween(game.GameDate, yesterday) == 0 {
					yesterdaysGame = &game
				}

				days := daysBetween(game.GameDate, today)

				if days >= 0 && days < len(upcoming) {
					upcoming[days] = &game
				}
			}
		}
	}

	// generate renderable info

	type yesterdayInfo struct {
		Where       string
		WinningTeam mlb.Team
		LosingTeam  mlb.Team
	}

	var yesterdayGameInfo *yesterdayInfo
	if yesterdaysGame != nil {

		var winningTeam mlb.Team
		var losingTeam mlb.Team
		var where string
		if yesterdaysGame.Teams.Away.IsWinner {
			where = "on the road"
			winningTeam = yesterdaysGame.Teams.Away
			losingTeam = yesterdaysGame.Teams.Home
		} else {
			where = "at home"
			winningTeam = yesterdaysGame.Teams.Home
			losingTeam = yesterdaysGame.Teams.Away
		}

		yesterdayGameInfo = &yesterdayInfo{
			Where:       where,
			WinningTeam: winningTeam,
			LosingTeam:  losingTeam,
		}
	}

	type render struct {
		Team            mlb.TeamFull
		BaseballTheater string
		Yesterday       *yesterdayInfo
		Upcoming        [8]*mlb.Game
	}

	baseballTheaterDate := yesterday.Format("20060102")

	r := render{
		Team:            myTeam,
		BaseballTheater: fmt.Sprintf("https://baseballtheater.com/games/%s", baseballTheaterDate),
		Yesterday:       yesterdayGameInfo,
		Upcoming:        upcoming,
	}

	t := template.Must(template.New("content").Parse(embeddedTemplate))

	err := t.Execute(os.Stdout, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return Report{
		Headline: "todo",
		Content:  "content",
	}
}
