package event

import (
	"context"
)

type Service interface {
	Get(ctx context.Context, id string) (*Event, error)
}

type Event struct {
	Name         string  `json:"name"`
	Date         string  `json:"date"`
	Organization string  `json:"organization"`
	Fights       []Fight `json:"fights"`
}

type Fight struct {
	Fighters    []Fighter `json:"fighters"`
	Way         string    `json:"way"`
	Round       int       `json:"round"`
	Time        string    `json:"time"`
	WeightClass string    `json:"weightclass"`
}

type Fighter struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Outcome string `json:"outcome"`
}
