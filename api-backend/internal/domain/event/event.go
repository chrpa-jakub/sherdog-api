package event

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const cacheTTL = time.Hour

var quotedNickname = regexp.MustCompile(`\s+'[^']+'\s*`)

type Service interface {
	Get(ctx context.Context, id string) (*Event, error)
}

type Database interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
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

type SherdogService struct {
	baseURL string
}

func NewSherdogService() *SherdogService {
	return &SherdogService{baseURL: "https://sherdog.com"}
}

func NewSherdogServiceWithBaseURL(baseURL string) *SherdogService {
	return &SherdogService{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *SherdogService) Get(_ context.Context, id string) (*Event, error) {
	event := &Event{}
	collector := colly.NewCollector(colly.UserAgent("Mozilla/5.0"))
	collector.AllowURLRevisit = true

	collector.OnHTML("div.left-module-pad-line", func(h *colly.HTMLElement) {
		event.Name = h.ChildText("h1 > span")
		event.Organization = h.ChildText("div > a > span")
		event.Date = h.ChildText("div.info > span:nth-child(1)")

		fight := Fight{
			WeightClass: h.ChildText("span.weight_class"),
			Fighters: []Fighter{
				{
					Name:    h.ChildText("div.fighter.left_side > h3 > a > span[itemprop=name]"),
					URL:     h.ChildAttr("div.fighter.left_side > h3 > a", "href"),
					Outcome: h.ChildText("div.fighter.left_side > span.final_result"),
				},
				{
					Name:    h.ChildText("div.fighter.right_side > h3 > a > span[itemprop=name]"),
					URL:     h.ChildAttr("div.fighter.right_side > h3 > a", "href"),
					Outcome: h.ChildText("div.fighter.right_side > span.final_result"),
				},
			},
			Way:   trimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(2)")),
			Round: toInt(trimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(4)"))),
			Time:  trimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(5)")),
		}
		if len(fight.Fighters) == 2 && fight.Fighters[0].Name != "" && fight.Fighters[1].Name != "" {
			event.Fights = append(event.Fights, fight)
		}
	})

	collector.OnHTML("tr", func(h *colly.HTMLElement) {
		fight := Fight{
			WeightClass: h.ChildText("span.weight_class"),
			Way:         h.ChildText("td.winby > b"),
			Round:       toInt(h.ChildText("td:nth-child(6)")),
			Time:        h.ChildText("td:last-child"),
			Fighters: []Fighter{
				{
					Name:    fighterListName(h, "left"),
					URL:     h.ChildAttr("div.fighter_list.left > div > a", "href"),
					Outcome: h.ChildText("div.fighter_list.left > div > span.final_result"),
				},
				{
					Name:    fighterListName(h, "right"),
					URL:     h.ChildAttr("div.fighter_list.right > div > a", "href"),
					Outcome: h.ChildText("div.fighter_list.right > div > span.final_result"),
				},
			},
		}
		if len(fight.Fighters) != 2 || fight.Fighters[0].Name == "" || fight.Fighters[1].Name == "" {
			return
		}

		event.Fights = append(event.Fights, fight)
	})

	if err := collector.Visit(s.resourceURL("events", id)); err != nil {
		return nil, err
	}

	return event, nil
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

func (s *CachedService) Get(ctx context.Context, id string) (*Event, error) {
	key := cacheKey(id)
	cached, err := s.database.Get(ctx, key)
	if err == nil {
		var event Event
		if err := json.Unmarshal([]byte(cached), &event); err == nil {
			return &event, nil
		}
	}

	event, err := s.next.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	_ = s.database.Set(ctx, key, string(payload), cacheTTL)

	return event, nil
}

func cacheKey(id string) string {
	return "event:" + id
}

func trimLabel(value string) string {
	_, after, found := strings.Cut(value, ":")
	if found {
		return strings.TrimSpace(after)
	}

	return strings.TrimSpace(value)
}

func toInt(numberString string) int {
	num, err := strconv.Atoi(numberString)
	if err != nil {
		return 0
	}

	return num
}

func fighterListName(h *colly.HTMLElement, side string) string {
	title := h.ChildAttr("div.fighter_list."+side+" > img", "title")
	if title != "" {
		return normalizeName(quotedNickname.ReplaceAllString(title, " "))
	}

	return normalizeName(h.ChildText("div.fighter_list." + side + " span[itemprop=name]"))
}

func normalizeName(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
