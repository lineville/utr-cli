package internal

// -----------------------------------------------------------------------------
// Defines the API Response Structures
// -----------------------------------------------------------------------------

type PlayerSearchResults struct {
	Players []Player `json:"hits"`
	Total   int      `json:"total"`
}

type Player struct {
	Source struct {
		Id          int    `json:"id"`
		DisplayName string `json:"displayName"`
		Gender      string `json:"gender"`
		AgeRange    string `json:"ageRange"`
		Location    struct {
			Display string `json:"display"`
		} `json:"location"`
	} `json:"source"`
}

type Profile struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Gender      string  `json:"gender"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Nationality string  `json:"nationality"`
	SinglesUTR  float64 `json:"singlesUtr"`
	DoublesUTR  float64 `json:"doublesUtr"`
}

type MatchResults struct {
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Events        []Event `json:"events"`
	WinLossString string  `json:"winLossString"`
}

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Draws     []Draw `json:"draws"`
}

type Draw struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	TeamType string  `json:"teamType"`
	Gender   string  `json:"gender"`
	Results  []Match `json:"results"`
}

type Match struct {
	Id       int               `json:"id"`
	Date     string            `json:"date"`
	Players  MatchParticipants `json:"players"`
	IsWinner bool              `json:"isWinner"`
	Score    Score             `json:"score"`
}

type MatchParticipants struct {
	Winner1 Profile `json:"winner1"`
	Winner2 Profile `json:"winner2"`
	Loser1  Profile `json:"loser1"`
	Loser2  Profile `json:"loser2"`
}

type Score struct {
	FirstSet  Set `json:"1"`
	SecondSet Set `json:"2"`
	ThirdSet  Set `json:"3"`
}

type Set struct {
	WinnerScore         int `json:"winner"`
	LoserScore          int `json:"loser"`
	LoserTieBreakScore  int `json:"tiebreak"`
	WinnerTieBreakScore int `json:"winnerTiebreak"`
}

type errorMessage struct {
	err error
}
