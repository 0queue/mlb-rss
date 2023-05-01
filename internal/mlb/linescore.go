package mlb

type Linescore struct {
	// if this is true, then the game ended with the home team
	// not playing the last inning because they won already
	IsTopInning bool
	Innings     []Inning
	Teams       struct {
		Home Stats
		Away Stats
	}
}

type Inning struct {
	Num  int
	Home Stats
	Away Stats
}

type Stats struct {
	Runs       int
	Hits       int
	Errors     int
	LeftOnBase int
}
