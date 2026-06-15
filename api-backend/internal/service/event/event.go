package event

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	domain "github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/util"
	"github.com/gocolly/colly"
)

const cacheTTL = time.Hour

var quotedNickname = regexp.MustCompile(`\s+'[^']+'\s*`)

type Database interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
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

func (s *SherdogService) Get(_ context.Context, id string) (*domain.Event, error) {
	event := &domain.Event{}
	collector := colly.NewCollector(colly.UserAgent("Mozilla/5.0"))
	collector.AllowURLRevisit = true
	collector.OnRequest(func(r *colly.Request) {
		log.Printf("fetching event source id=%s url=%s", id, r.URL.String())
	})
	collector.OnError(func(r *colly.Response, err error) {
		log.Printf("event source request failed id=%s url=%s status=%d err=%v", id, r.Request.URL.String(), r.StatusCode, err)
	})

	collector.OnHTML("div.left-module-pad-line", func(h *colly.HTMLElement) {
		event.Name = h.ChildText("h1 > span")
		event.Organization = h.ChildText("div > a > span")
		event.Date = h.ChildText("div.info > span:nth-child(1)")

		fight := domain.Fight{
			WeightClass: h.ChildText("span.weight_class"),
			Fighters: []domain.Fighter{
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
			Way:   util.TrimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(2)")),
			Round: util.ToInt(util.TrimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(4)"))),
			Time:  util.TrimLabel(h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(5)")),
		}
		if len(fight.Fighters) == 2 && fight.Fighters[0].Name != "" && fight.Fighters[1].Name != "" {
			event.Fights = append(event.Fights, fight)
		}
	})

	collector.OnHTML("tr", func(h *colly.HTMLElement) {
		fight := domain.Fight{
			WeightClass: h.ChildText("span.weight_class"),
			Way:         h.ChildText("td.winby > b"),
			Round:       util.ToInt(h.ChildText("td:nth-child(6)")),
			Time:        h.ChildText("td:last-child"),
			Fighters: []domain.Fighter{
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
		log.Printf("event source visit failed id=%s err=%v", id, err)
		return nil, err
	}

	log.Printf("event parsed id=%s name=%q fights=%d", id, event.Name, len(event.Fights))
	return event, nil
}

func (s *SherdogService) resourceURL(resource, id string) string {
	escapedID := url.PathEscape(id)
	return s.baseURL + "/" + resource + "/" + escapedID
}

type CachedService struct {
	next     domain.Service
	database Database
}

func NewCachedService(next domain.Service, database Database) *CachedService {
	return &CachedService{next: next, database: database}
}

func (s *CachedService) Get(ctx context.Context, id string) (*domain.Event, error) {
	key := cacheKey(id)
	cached, err := s.database.Get(ctx, key)
	if err == nil {
		var event domain.Event
		if err := json.Unmarshal([]byte(cached), &event); err == nil {
			log.Printf("event cache hit id=%s key=%s", id, key)
			return &event, nil
		} else {
			log.Printf("event cache payload invalid id=%s key=%s err=%v", id, key, err)
		}
	} else {
		log.Printf("event cache miss id=%s key=%s err=%v", id, key, err)
	}

	event, err := s.next.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	if err := s.database.Set(ctx, key, string(payload), cacheTTL); err != nil {
		log.Printf("event cache set failed id=%s key=%s err=%v", id, key, err)
	} else {
		log.Printf("event cache set id=%s key=%s ttl=%s", id, key, cacheTTL)
	}

	return event, nil
}

func cacheKey(id string) string {
	return "event:" + id
}

func fighterListName(h *colly.HTMLElement, side string) string {
	title := h.ChildAttr("div.fighter_list."+side+" > img", "title")
	if title != "" {
		return util.NormalizeName(quotedNickname.ReplaceAllString(title, " "))
	}

	return util.NormalizeName(h.ChildText("div.fighter_list." + side + " span[itemprop=name]"))
}
