package fighter

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"time"

	domain "github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	"github.com/chrpa-jakub/sherdog-api/internal/util"
	"github.com/gocolly/colly"
)

const cacheTTL = time.Hour

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

func (s *SherdogService) Get(_ context.Context, id string) (*domain.Fighter, error) {
	fighter := &domain.Fighter{}
	collector := colly.NewCollector(colly.UserAgent("Mozilla/5.0"))
	collector.AllowURLRevisit = true
	collector.OnRequest(func(r *colly.Request) {
		log.Printf("fetching fighter source id=%s url=%s", id, r.URL.String())
	})
	collector.OnError(func(r *colly.Response, err error) {
		log.Printf("fighter source request failed id=%s url=%s status=%d err=%v", id, r.Request.URL.String(), r.StatusCode, err)
	})

	collector.OnHTML("div.fighter-title", func(h *colly.HTMLElement) {
		fighter.Name = h.ChildText("span.fn")
		fighter.Nickname = h.ChildText("span.nickname > em")
		fighter.Origin = h.ChildText("strong[itemprop=nationality]")
	})

	collector.OnHTML("div.fighter-data", func(h *colly.HTMLElement) {
		fighter.Age = util.ToInt(h.ChildText("tr:nth-child(1) > td > b"))
		fighter.Height = util.NextToLastField(h.ChildText("tr:nth-child(2) > td"))
		fighter.Weight = util.NextToLastField(h.ChildText("tr:nth-child(3) > td"))
		fighter.Gym = h.ChildText("div.association-class > span")
		fighter.WeightClass = h.ChildText("div.association-class > a")
		fighter.Wins = domain.Outcome{
			Total:      util.ToInt(h.ChildText("div.win > span:nth-child(2)")),
			KO:         util.ToInt(h.ChildText("div.wins > div.meter:nth-child(3) > div.pl")),
			Submission: util.ToInt(h.ChildText("div.wins > div.meter:nth-child(5) > div.pl")),
			Decision:   util.ToInt(h.ChildText("div.wins > div.meter:nth-child(7) > div.pl")),
			Others:     util.ToInt(h.ChildText("div.wins > div.meter:nth-child(9) > div.pl")),
		}
		fighter.Losses = domain.Outcome{
			Total:      util.ToInt(h.ChildText("div.lose > span:nth-child(2)")),
			KO:         util.ToInt(h.ChildText("div.loses > div.meter:nth-child(3) > div.pl")),
			Submission: util.ToInt(h.ChildText("div.loses > div.meter:nth-child(5) > div.pl")),
			Decision:   util.ToInt(h.ChildText("div.loses > div.meter:nth-child(7) > div.pl")),
			Others:     util.ToInt(h.ChildText("div.loses > div.meter:nth-child(9) > div.pl")),
		}
		fighter.Draws = util.ToInt(h.ChildText("div.draws > span:nth-child(2)"))
		fighter.NoContests = util.ToInt(h.ChildText("div.nc > span:nth-child(2)"))
	})

	collector.OnHTML("div.fight_card_preview", func(h *colly.HTMLElement) {
		if fighter.UpcomingFight != nil {
			return
		}

		fighter.UpcomingFight = &domain.UpcomingFight{
			Opponent: upcomingOpponent(h, id, fighter.Name),
			Event: domain.Event{
				Name: h.ChildText("h2[itemprop=name]"),
				URL:  h.ChildAttr("a.card_button", "href"),
			},
			Date:     h.ChildAttr("div.date_location meta[itemprop=startDate]", "content"),
			Location: strings.TrimSpace(h.ChildText("div.date_location em")),
		}
	})

	collector.OnHTML("tr", func(h *colly.HTMLElement) {
		fight := domain.Fight{
			Outcome: h.ChildText("span.final_result"),
			Opponent: domain.Opponent{
				Name: h.ChildText("td:nth-child(2) > a"),
				URL:  h.ChildAttr("td:nth-child(2) > a", "href"),
			},
			Event: domain.Event{
				Name: h.ChildText("td:nth-child(3) > a"),
				URL:  h.ChildAttr("td:nth-child(3) > a", "href"),
			},
			Way:   h.ChildText("td.winby > b"),
			Round: util.ToInt(h.ChildText("td:nth-child(5)")),
			Time:  h.ChildText("td:last-child"),
		}
		if len(fight.Event.Name) < 3 {
			return
		}

		fighter.Fights = append(fighter.Fights, fight)
	})

	if err := collector.Visit(s.resourceURL("fighter", id)); err != nil {
		log.Printf("fighter source visit failed id=%s err=%v", id, err)
		return nil, err
	}

	log.Printf("fighter parsed id=%s name=%q fights=%d has_upcoming=%t", id, fighter.Name, len(fighter.Fights), fighter.UpcomingFight != nil)
	return fighter, nil
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

func (s *CachedService) Get(ctx context.Context, id string) (*domain.Fighter, error) {
	key := cacheKey(id)
	cached, err := s.database.Get(ctx, key)
	if err != nil {
		log.Printf("fighter cache miss id=%s key=%s err=%v", id, key, err)
	}
	if err == nil {
		var fighter domain.Fighter
		unmarshalErr := json.Unmarshal([]byte(cached), &fighter)
		if unmarshalErr != nil {
			log.Printf("fighter cache payload invalid id=%s key=%s err=%v", id, key, unmarshalErr)
		}
		if unmarshalErr == nil {
			log.Printf("fighter cache hit id=%s key=%s", id, key)
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
	if err := s.database.Set(ctx, key, string(payload), cacheTTL); err != nil {
		log.Printf("fighter cache set failed id=%s key=%s err=%v", id, key, err)
		return fighter, nil
	}
	log.Printf("fighter cache set id=%s key=%s ttl=%s", id, key, cacheTTL)

	return fighter, nil
}

func cacheKey(id string) string {
	return "fighter:" + id
}

func upcomingOpponent(h *colly.HTMLElement, id, fighterName string) domain.Opponent {
	left := previewFighter(h, "left_side")
	right := previewFighter(h, "right_side")
	shortID := util.ShortenID(id)

	switch {
	case shortID != "" && util.ShortenID(left.URL) == shortID:
		return right
	case shortID != "" && util.ShortenID(right.URL) == shortID:
		return left
	case strings.EqualFold(left.Name, fighterName):
		return right
	default:
		return left
	}
}

func previewFighter(h *colly.HTMLElement, side string) domain.Opponent {
	return domain.Opponent{
		Name: h.ChildText("div.fighter." + side + " span[itemprop=name]"),
		URL:  h.ChildAttr("div.fighter."+side+" a[itemprop=url]", "href"),
	}
}
