package report

import (
	"bytes"
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
	Link     string
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

	type upcomingInfo struct {
		IsHome      bool
		AgainstAbbr string
	}

	type render struct {
		Team             mlb.TeamFull
		BaseballTheater  string
		Yesterday        *yesterdayInfo
		UpcomingDayAbbr  [8]string
		UpcomingInfos    [8]*upcomingInfo
		UpcomingTimes    [8]*string
		UpcomingTimezone string
	}

	var upcomingDayAbbr [8]string
	for i := 0; i < 8; i++ {
		upcomingDayAbbr[i] = today.AddDate(0, 0, i).Weekday().String()[:2]
	}

	var upcomingInfos [8]*upcomingInfo
	for i := 0; i < 8; i++ {
		g := upcoming[i]
		if g != nil {

			var against mlb.Team
			if g.Teams.Home.Team.Id != myTeam.Id {
				against = g.Teams.Home
			} else {
				against = g.Teams.Away
			}

			upcomingInfos[i] = &upcomingInfo{
				IsHome:      g.Teams.Home.Team.Id == myTeam.Id,
				AgainstAbbr: teams[against.Team.Id].Abbreviation,
			}
		}
	}

	var upcomingTimes [8]*string
	for i := 0; i < 8; i++ {
		g := upcoming[i]
		if g != nil {
			s := g.GameDate.Local().Format("15:04")
			upcomingTimes[i] = &s
		}
	}

	baseballTheaterDate := yesterday.Format("20060102")

	upcomingTimezone, _ := today.Local().Zone()

	r := render{
		Team:             myTeam,
		BaseballTheater:  fmt.Sprintf("https://baseball.theater/games/%s", baseballTheaterDate),
		Yesterday:        yesterdayGameInfo,
		UpcomingDayAbbr:  upcomingDayAbbr,
		UpcomingInfos:    upcomingInfos,
		UpcomingTimes:    upcomingTimes,
		UpcomingTimezone: upcomingTimezone,
	}

	t := template.Must(template.New("content").Parse(embeddedTemplate))

	var content bytes.Buffer
	err := t.Execute(&content, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var headline string
	if yesterdayGameInfo != nil {
		if yesterdayGameInfo.WinningTeam.Team.Id == myTeam.Id {
			headline = fmt.Sprintf("The %s win! %d to %d", myTeam.Name, yesterdayGameInfo.WinningTeam.Score, yesterdayGameInfo.LosingTeam.Score)
		} else {
			headline = fmt.Sprintf("The %s lose, %d to %d", myTeam.Name, yesterdayGameInfo.LosingTeam.Score, yesterdayGameInfo.WinningTeam.Score)
		}
	} else {
		headline = fmt.Sprintf("Baseball report %s %s", today.Weekday().String(), today.Format("2006-01-02"))
	}

	return Report{
		Headline: headline,
		Link:     r.BaseballTheater,
		Content:  content.String(),
	}
}
