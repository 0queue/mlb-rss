package report

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/0queue/mlb-rss/internal/mlb"
	"github.com/0queue/mlb-rss/ui"
	"golang.org/x/exp/slog"
)

const BaseballTheaterTimeFormat = "20060102"

type ReportGenerator struct {
	MyTeamId int
	mc       *mlb.MlbClient
	Location *time.Location
	t        *template.Template
}

func NewReportGenerator(myTeamId int, mc *mlb.MlbClient, loc *time.Location) ReportGenerator {
	funcs := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	return ReportGenerator{
		MyTeamId: myTeamId,
		mc:       mc,
		Location: loc,
		t:        template.Must(template.New("").Funcs(funcs).ParseFS(ui.ReportTemplates, "*.html.tpl")),
	}
}

// Do analysis of the games and generate the report. NO TEMPLATES
// assumes the first Date is yesterday, and the rest are the future
// ultimately, analysis consists of filtering
func (rg *ReportGenerator) GenerateReport(today time.Time) (Report, error) {

	s, err := rg.mc.FetchSchedule(today.AddDate(0, 0, -1), today.AddDate(0, 0, 7), rg.MyTeamId)
	if err != nil {
		return Report{}, nil
	}

	pastGames := rg.analyzePastGames(s.Dates, rg.MyTeamId, today)
	futureGames := rg.analyzeFutureGames(today, s.Dates)

	baseballTheaterDate := today.AddDate(0, 0, -1).Format(BaseballTheaterTimeFormat)
	link := fmt.Sprintf("https://baseball.theater/games/%s", baseballTheaterDate)

	yesterday := Yesterday{
		MyTeamName:      rg.mc.AllTeams[rg.MyTeamId].Name,
		PastGames:       pastGames,
		BaseballTheater: link,
	}

	tz, _ := today.Local().Zone()

	upcoming := Upcoming{
		FutureDays: futureGames,
		Timezone:   tz,
	}

	headline := rg.generateHeadline(pastGames, today)

	return Report{
		Yesterday: yesterday,
		Upcoming:  upcoming,
		Headline:  headline,
		Link:      link,
		When:      today,
	}, nil
}

// Render uses templates to render reports to html
func (rg *ReportGenerator) Render(r Report) (string, error) {
	var content bytes.Buffer
	err := rg.t.ExecuteTemplate(&content, "report.html.tpl", r)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func (rg *ReportGenerator) RenderWeb(r Report) (string, error) {

	body, err := rg.Render(r)
	if err != nil {
		return "", err
	}

	var content bytes.Buffer
	err = rg.t.ExecuteTemplate(&content, "web.html.tpl", struct {
		Title string
		H2    string
		Body  string
	}{
		Title: r.Headline,
		H2:    r.Headline,
	})
	if err != nil {
		return "", err
	}

	// hack until I make a separate template for web...
	return content.String() + body + `</body></html>`, nil
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

func (rg *ReportGenerator) analyzePastGames(dates []mlb.Date, id int, today time.Time) []PastGame {

	yesterdayDateFormatted := today.AddDate(0, 0, -1).Format(time.DateOnly)

	var date mlb.Date
	var found bool
	for _, d := range dates {
		d := d
		if d.Date == yesterdayDateFormatted {
			date = d
			found = true
		}
	}

	if !found {
		slog.Info("No game played yesterday", slog.Time("yesterday", today.AddDate(0, 0, -1)))
	}

	pastGames := make([]PastGame, 0)

	slog.Info("Analyzing yesterday's game", slog.String("date", date.Date))

	for _, g := range date.Games {

		// TODO doesn't really handle ties well
		var isWinnerHome = g.Teams.Home.IsWinner
		var winner mlb.GameTeam
		var loser mlb.GameTeam

		if isWinnerHome {
			winner = g.Teams.Home
			loser = g.Teams.Away
		} else {
			winner = g.Teams.Away
			loser = g.Teams.Home
		}

		var postponeReason string
		if g.Status.DetailedState == "Postponed" {
			postponeReason = g.Status.Reason
		}

		u, err := rg.fetchCondensedGame(g.GamePk)
		if err != nil {
			slog.Warn(
				"Failed to fetch condensed game",
				slog.Int("gamePk", g.GamePk),
				slog.String("err", err.Error()),
			)
		}

		l, err := rg.fetchLinescore(
			g.GamePk,
			g.Teams.Home.Team.Id,
			g.Teams.Away.Team.Id,
		)
		hasLinescore := err == nil
		if err != nil {
			slog.Warn(
				"Failed to fetch linescore",
				slog.Int("gamePk", g.GamePk),
				slog.String("err", err.Error()),
			)
		}

		p := PastGame{
			PostponeReason:   postponeReason,
			Venue:            g.Venue,
			IsWinnerHome:     isWinnerHome,
			W:                winner,
			L:                loser,
			CondensedGameUrl: u,
			HasLinescore:     hasLinescore,
			Linescore:        l,
		}

		pastGames = append(pastGames, p)
	}

	return pastGames
}

func (rg *ReportGenerator) analyzeFutureGames(today time.Time, dates []mlb.Date) [8]FutureDay {
	var futureGames [8]FutureDay

	// if a Date has no games then it will not be there
	m := make(map[string]mlb.Date)
	for _, d := range dates {
		m[d.Date] = d
	}

	daysWithGames := 0

	for i := 0; i < 8; i += 1 {
		k := today.AddDate(0, 0, i).Format("2006-01-02")
		d, ok := m[k]
		var gs []mlb.Game
		if !ok {
			gs = make([]mlb.Game, 0)
		} else {
			daysWithGames += 1
			gs = d.Games
		}

		games := make([]FutureGame, 0)
		for _, g := range gs {

			isHome := g.Teams.Home.Team.Id == rg.MyTeamId
			var opponentTeam mlb.GameTeam
			if isHome {
				opponentTeam = g.Teams.Away
			} else {
				opponentTeam = g.Teams.Home
			}

			futureGame := FutureGame{
				GameTimeLocal: g.GameDate.In(rg.Location).Format("15:04"),
				IsMyTeamHome:  isHome,
				AgainstAbbr:   rg.mc.AllTeams[opponentTeam.Team.Id].Abbreviation,
			}

			games = append(games, futureGame)
		}

		dayAbbr := today.AddDate(0, 0, i).Weekday().String()[:2]

		futureGames[i] = FutureDay{
			DayAbbr: dayAbbr,
			Games:   games,
		}
	}

	slog.Info("Upcoming games analyzed", slog.Int("daysWithGames", daysWithGames))

	return futureGames
}

func (rg *ReportGenerator) generateHeadline(pastGames []PastGame, today time.Time) string {
	// 1. no games
	// 2. Postpone
	// 3. team wins! 1 of 1
	// 4. team loses :( 1 of 1
	// 5. team ties? guess so
	// 6. Double header

	myTeamName := rg.mc.AllTeams[rg.MyTeamId].Name

	switch len(pastGames) {
	case 0:
		return fmt.Sprintf("Baseball Report %s", today.Format("Monday 2006-01-02"))
	case 1:
		var headline string
		g := pastGames[0]
		if g.PostponeReason != "" {
			headline = fmt.Sprintf("Game was postponed due to %s", g.PostponeReason)
		} else if g.W.Score == g.L.Score {
			headline = fmt.Sprintf("The %s tie, %d to %d", myTeamName, g.W.Score, g.L.Score)
		} else if g.W.Team.Id == rg.MyTeamId {
			headline = fmt.Sprintf("The %s win! %d to %d", myTeamName, g.W.Score, g.L.Score)
		} else {
			headline = fmt.Sprintf("The %s lose, %d to %d", myTeamName, g.L.Score, g.W.Score)
		}
		return headline
	case 2:
		var winCount int
		var lossCount int
		for _, g := range pastGames {
			if g.W.Team.Id == rg.MyTeamId {
				winCount += 1
			} else {
				lossCount += 1
			}
		}
		return fmt.Sprintf("Doubleheader! The %s go %d - %d", myTeamName, winCount, lossCount)
	default:
		slog.Warn("Failed to generateHeadline, too many games played", slog.Int("pastGamesCount", len(pastGames)))
		return "error"
	}
}

// logs errors
func (rg *ReportGenerator) fetchCondensedGame(gamePk int) (string, error) {

	c, err := rg.mc.FetchContent(gamePk)
	if err != nil {
		return "", err
	}

	h, found := c.FindByTypeAndValue("mlbtax", "condensed_game")
	if !found {
		return "", errors.New("Failed to find condensed game")
	}

	p, found := h.FindPlaybackByName("highBit")
	if !found {
		return "", errors.New("Failed to find high bit rate condensed game")
	}

	return p.Url, nil
}

func (rg *ReportGenerator) fetchLinescore(gamePk, homeId, awayId int) (Linescore, error) {

	l, err := rg.mc.FetchLinescore(gamePk)
	if err != nil {
		return Linescore{}, err
	}

	homeLinescore := LinescoreTeam{
		Abbr:    rg.mc.AllTeams[homeId].Abbreviation,
		Innings: []int{},
		Runs:    l.Teams.Home.Runs,
		Hits:    l.Teams.Home.Hits,
		Errors:  l.Teams.Home.Errors,
	}

	awayLinescore := LinescoreTeam{
		Abbr:    rg.mc.AllTeams[awayId].Abbreviation,
		Innings: []int{},
		Runs:    l.Teams.Away.Runs,
		Hits:    l.Teams.Away.Hits,
		Errors:  l.Teams.Away.Errors,
	}

	for _, i := range l.Innings {
		homeLinescore.Innings = append(homeLinescore.Innings, i.Home.Runs)
		awayLinescore.Innings = append(awayLinescore.Innings, i.Away.Runs)
	}

	if l.IsTopInning {
		// negative numbers are rendered as x
		homeLinescore.Innings[len(homeLinescore.Innings)-1] = -1
	}

	return Linescore{
		Home: homeLinescore,
		Away: awayLinescore,
	}, nil
}
