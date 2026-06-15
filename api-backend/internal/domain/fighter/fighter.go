package fighter

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const cacheTTL = time.Hour

type Service interface {
	Get(ctx context.Context, id string) (*Fighter, error)
}

type Database interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
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

type SherdogService struct {
	baseURL string
}

func NewSherdogService() *SherdogService {
	return &SherdogService{baseURL: "https://sherdog.com"}
}

func NewSherdogServiceWithBaseURL(baseURL string) *SherdogService {
	return &SherdogService{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *SherdogService) Get(_ context.Context, id string) (*Fighter, error) {
	fighter := &Fighter{}
	collector := colly.NewCollector(colly.UserAgent("Mozilla/5.0"))
	collector.AllowURLRevisit = true

	collector.OnHTML("div.fighter-title", func(h *colly.HTMLElement) {
		fighter.Name = h.ChildText("span.fn")
		fighter.Nickname = h.ChildText("span.nickname > em")
		fighter.Origin = h.ChildText("strong[itemprop=nationality]")
	})

	collector.OnHTML("div.fighter-data", func(h *colly.HTMLElement) {
		fighter.Age = toInt(h.ChildText("tr:nth-child(1) > td > b"))
		fighter.Height = nextToLastField(h.ChildText("tr:nth-child(2) > td"))
		fighter.Weight = nextToLastField(h.ChildText("tr:nth-child(3) > td"))
		fighter.Gym = h.ChildText("div.association-class > span")
		fighter.WeightClass = h.ChildText("div.association-class > a")
		fighter.Wins = Outcome{
			Total:      toInt(h.ChildText("div.win > span:nth-child(2)")),
			KO:         toInt(h.ChildText("div.wins > div.meter:nth-child(3) > div.pl")),
			Submission: toInt(h.ChildText("div.wins > div.meter:nth-child(5) > div.pl")),
			Decision:   toInt(h.ChildText("div.wins > div.meter:nth-child(7) > div.pl")),
			Others:     toInt(h.ChildText("div.wins > div.meter:nth-child(9) > div.pl")),
		}
		fighter.Losses = Outcome{
			Total:      toInt(h.ChildText("div.lose > span:nth-child(2)")),
			KO:         toInt(h.ChildText("div.loses > div.meter:nth-child(3) > div.pl")),
			Submission: toInt(h.ChildText("div.loses > div.meter:nth-child(5) > div.pl")),
			Decision:   toInt(h.ChildText("div.loses > div.meter:nth-child(7) > div.pl")),
			Others:     toInt(h.ChildText("div.loses > div.meter:nth-child(9) > div.pl")),
		}
		fighter.Draws = toInt(h.ChildText("div.draws > span:nth-child(2)"))
		fighter.NoContests = toInt(h.ChildText("div.nc > span:nth-child(2)"))
	})

	collector.OnHTML("div.fight_card_preview", func(h *colly.HTMLElement) {
		if fighter.UpcomingFight != nil {
			return
		}

		fighter.UpcomingFight = &UpcomingFight{
			Opponent: upcomingOpponent(h, id, fighter.Name),
			Event: Event{
				Name: h.ChildText("h2[itemprop=name]"),
				URL:  h.ChildAttr("a.card_button", "href"),
			},
			Date:     h.ChildAttr("div.date_location meta[itemprop=startDate]", "content"),
			Location: strings.TrimSpace(h.ChildText("div.date_location em")),
		}
	})

	collector.OnHTML("tr", func(h *colly.HTMLElement) {
		fight := Fight{
			Outcome: h.ChildText("span.final_result"),
			Opponent: Opponent{
				Name: h.ChildText("td:nth-child(2) > a"),
				URL:  h.ChildAttr("td:nth-child(2) > a", "href"),
			},
			Event: Event{
				Name: h.ChildText("td:nth-child(3) > a"),
				URL:  h.ChildAttr("td:nth-child(3) > a", "href"),
			},
			Way:   h.ChildText("td.winby > b"),
			Round: toInt(h.ChildText("td:nth-child(5)")),
			Time:  h.ChildText("td:last-child"),
		}
		if len(fight.Event.Name) < 3 {
			return
		}

		fighter.Fights = append(fighter.Fights, fight)
	})

	if err := collector.Visit(s.resourceURL("fighter", id)); err != nil {
		return nil, err
	}

	return fighter, nil
}

func (s *SherdogService) resourceURL(resource, id string) string {
	escapedID := url.PathEscape(id)
	return s.baseURL + "/" + resource + "/" + escapedID
}

type CachedService struct {
	next     Service
	database Database
}

func NewCachedService(next Service, database Database) *CachedService {
	return &CachedService{next: next, database: database}
}

func (s *CachedService) Get(ctx context.Context, id string) (*Fighter, error) {
	key := cacheKey(id)
	cached, err := s.database.Get(ctx, key)
	if err == nil {
		var fighter Fighter
		if err := json.Unmarshal([]byte(cached), &fighter); err == nil {
			return &fighter, nil
		}
	}

	fighter, err := s.next.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(fighter)
	if err != nil {
		return nil, err
	}
	_ = s.database.Set(ctx, key, string(payload), cacheTTL)

	return fighter, nil
}

func cacheKey(id string) string {
	return "fighter:" + id
}

func ShortenID(id string) string {
	parts := strings.Split(id, "-")
	return parts[len(parts)-1]
}

func toInt(numberString string) int {
	num, err := strconv.Atoi(numberString)
	if err != nil {
		return 0
	}

	return num
}

func nextToLastField(value string) string {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return ""
	}

	return fields[len(fields)-2]
}

func upcomingOpponent(h *colly.HTMLElement, id, fighterName string) Opponent {
	left := previewFighter(h, "left_side")
	right := previewFighter(h, "right_side")
	shortID := shortenID(id)

	switch {
	case shortID != "" && shortenID(left.URL) == shortID:
		return right
	case shortID != "" && shortenID(right.URL) == shortID:
		return left
	case strings.EqualFold(left.Name, fighterName):
		return right
	default:
		return left
	}
}

func previewFighter(h *colly.HTMLElement, side string) Opponent {
	return Opponent{
		Name: h.ChildText("div.fighter." + side + " span[itemprop=name]"),
		URL:  h.ChildAttr("div.fighter."+side+" a[itemprop=url]", "href"),
	}
}

func shortenID(id string) string {
	parts := strings.Split(strings.Trim(id, "/"), "-")
	return parts[len(parts)-1]
}
