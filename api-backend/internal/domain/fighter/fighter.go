package fighter

import (
	"context"
)

type Service interface {
	Get(ctx context.Context, id string) (*Fighter, error)
}

type Fighter struct {
	Name          string         `json:"name"`
	Nickname      string         `json:"nickname"`
	Origin        string         `json:"origin"`
	Age           int            `json:"age"`
	Height        string         `json:"height"`
	Weight        string         `json:"weight"`
	WeightClass   string         `json:"weightclass"`
	Gym           string         `json:"gym"`
	Wins          Outcome        `json:"win"`
	Losses        Outcome        `json:"loses"`
	Draws         int            `json:"draws"`
	NoContests    int            `json:"nocontests"`
	UpcomingFight *UpcomingFight `json:"upcomingfight"`
	Fights        []Fight        `json:"fights"`
}

type Event struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Outcome struct {
	Total      int `json:"total"`
	KO         int `json:"ko"`
	Submission int `json:"submission"`
	Decision   int `json:"decision"`
	Others     int `json:"others"`
}

type Opponent struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type UpcomingFight struct {
	Opponent Opponent `json:"opponent"`
	Event    Event    `json:"event"`
	Date     string   `json:"date"`
	Location string   `json:"location"`
}

type Fight struct {
	Outcome  string   `json:"outcome"`
	Opponent Opponent `json:"oponent"`
	Event    Event    `json:"event"`
	Way      string   `json:"way"`
	Round    int      `json:"round"`
	Time     string   `json:"time"`
}
