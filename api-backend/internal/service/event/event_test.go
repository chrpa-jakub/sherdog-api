package event

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "github.com/chrpa-jakub/sherdog-api/internal/domain/event"
)

func TestSherdogServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/UFC-295-98494" {
			t.Fatalf("path = %q, want /events/UFC-295-98494", r.URL.Path)
		}
		w.Write([]byte(eventHTML))
	}))
	t.Cleanup(server.Close)

	service := NewSherdogServiceWithBaseURL(server.URL)
	got, err := service.Get(context.Background(), "UFC-295-98494")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if got.Name != "UFC 295" || got.Organization != "Ultimate Fighting Championship" {
		t.Fatalf("unexpected event fields: %+v", got)
	}
	if len(got.Fights) != 2 {
		t.Fatalf("len(Fights) = %d, want 2", len(got.Fights))
	}
	main := got.Fights[0]
	if main.Way != "TKO" || main.Round != 2 || main.Time != "4:08" {
		t.Fatalf("unexpected main event fight: %+v", main)
	}
	prelim := got.Fights[1]
	if prelim.Fighters[0].Name != "Fighter One" || prelim.Fighters[1].Outcome != "loss" {
		t.Fatalf("unexpected undercard fight: %+v", prelim)
	}
}

func TestSherdogServiceGetUpcomingEvent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/UFC-Fight-Night-279-112139" {
			t.Fatalf("path = %q, want /events/UFC-Fight-Night-279-112139", r.URL.Path)
		}
		w.Write([]byte(upcomingEventHTML))
	}))
	t.Cleanup(server.Close)

	service := NewSherdogServiceWithBaseURL(server.URL)
	got, err := service.Get(context.Background(), "UFC-Fight-Night-279-112139")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if got.Name != "UFC Fight Night 279 - Kape vs. Horiguchi 2" {
		t.Fatalf("Name = %q, want upcoming event name", got.Name)
	}
	if len(got.Fights) != 2 {
		t.Fatalf("len(Fights) = %d, want 2", len(got.Fights))
	}

	main := got.Fights[0]
	if main.Fighters[0].Outcome != "yet to come" || main.Fighters[1].Outcome != "yet to come" {
		t.Fatalf("unexpected main event outcomes: %+v", main.Fighters)
	}
	if main.Way != "" || main.Round != 0 || main.Time != "" {
		t.Fatalf("upcoming main event should not have result data: %+v", main)
	}

	prelim := got.Fights[1]
	if prelim.Fighters[0].Name != "Andre Fili" || prelim.Fighters[1].Name != "Vinicius Oliveira" {
		t.Fatalf("unexpected upcoming fight names: %+v", prelim.Fighters)
	}
	if prelim.WeightClass != "Featherweight" || prelim.Way != "" || prelim.Round != 0 || prelim.Time != "" {
		t.Fatalf("unexpected upcoming fight: %+v", prelim)
	}
}

func TestCachedServiceUsesCacheHit(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	store.values["event:98494"] = `{"name":"Cached Event"}`
	service := NewCachedService(errorService{}, store)

	got, err := service.Get(context.Background(), "98494")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Name != "Cached Event" {
		t.Fatalf("Name = %q, want Cached Event", got.Name)
	}
}

func TestCachedServiceCachesMiss(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	service := NewCachedService(staticService{event: &domain.Event{Name: "Parsed Event"}}, store)

	got, err := service.Get(context.Background(), "98494")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Name != "Parsed Event" {
		t.Fatalf("Name = %q, want Parsed Event", got.Name)
	}
	if store.values["event:98494"] == "" {
		t.Fatal("expected event to be cached")
	}
}

type staticService struct {
	event *domain.Event
}

func (s staticService) Get(context.Context, string) (*domain.Event, error) {
	return s.event, nil
}

type errorService struct{}

func (errorService) Get(context.Context, string) (*domain.Event, error) {
	return nil, errors.New("unexpected service call")
}

type memoryStore struct {
	values map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{values: make(map[string]string)}
}

func (s *memoryStore) Get(_ context.Context, key string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return "", errors.New("cache entry not found")
	}

	return value, nil
}

func (s *memoryStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.values[key] = value
	return nil
}

const eventHTML = `
<html>
	<body>
		<div class="left-module-pad-line">
			<h1><span>UFC 295</span></h1>
			<div><a><span>Ultimate Fighting Championship</span></a></div>
			<div class="info"><span>Nov 11, 2023</span></div>
			<span class="weight_class">Light Heavyweight</span>
			<div class="fighter left_side">
				<h3><a href="/fighter/Jiri-Prochazka-97529"><span itemprop="name">Jiri Prochazka</span></a></h3>
				<span class="final_result">loss</span>
			</div>
			<div class="fighter right_side">
				<h3><a href="/fighter/Alex-Pereira-224511"><span itemprop="name">Alex Pereira</span></a></h3>
				<span class="final_result">win</span>
			</div>
			<table class="fight_card_resume">
				<tbody>
					<tr>
						<td></td>
						<td>Method: TKO</td>
						<td></td>
						<td>Round: 2</td>
						<td>Time: 4:08</td>
					</tr>
				</tbody>
			</table>
		</div>
		<table>
			<tr>
				<td><span class="weight_class">Middleweight</span></td>
				<td>
					<div class="fighter_list left">
						<img title="Fighter One">
						<div><a href="/fighter/one">Fighter One</a><span class="final_result">win</span></div>
					</div>
				</td>
				<td>
					<div class="fighter_list right">
						<img title="Fighter Two">
						<div><a href="/fighter/two">Fighter Two</a><span class="final_result">loss</span></div>
					</div>
				</td>
				<td class="winby"><b>Decision</b></td>
				<td></td>
				<td>3</td>
				<td>5:00</td>
			</tr>
		</table>
	</body>
</html>`

const upcomingEventHTML = `
<html>
	<body>
		<div class="left-module-pad-line">
			<div class="event_detail">
				<h1><span itemprop="name">UFC Fight Night 279 - Kape vs. Horiguchi 2</span></h1>
				<div class="organization"><a><span itemprop="name">Ultimate Fighting Championship (UFC)</span></a></div>
				<div class="info"><span>Jun 20, 2026</span></div>
			</div>
			<div class="fight_card">
				<div class="fighter left_side">
					<h3><a href="/fighter/Manel-Kape-101189"><span itemprop="name">Manel Kape</span></a></h3>
					<span class="final_result yet_to_come">yet to come</span>
				</div>
				<div class="versus">
					<span class="weight_class">Flyweight</span>
				</div>
				<div class="fighter right_side">
					<h3><a href="/fighter/Kyoji-Horiguchi-64413"><span itemprop="name">Kyoji Horiguchi</span></a></h3>
					<span class="final_result yet_to_come">yet to come</span>
				</div>
			</div>
		</div>
		<table class="new_table upcoming">
			<tbody>
				<tr class="table_head"><td>Match</td><td></td><td>Fighters</td><td></td><td></td></tr>
				<tr itemprop="subEvent">
					<td>11</td>
					<td class="text_right col_fc_upcoming">
						<div class="fighter_list left">
							<img title="Andre 'Touchy' Fili">
							<div class="fighter_result_data">
								<a href="/fighter/Andre-Fili-58385"><span itemprop="name">Andre<br />Fili</span></a>
							</div>
						</div>
					</td>
					<td class="text_center">
						<span class="weight_class">Featherweight</span>
					</td>
					<td class="text_left col_fc_upcoming">
						<div class="fighter_list right">
							<img title="Vinicius 'LokDog' Oliveira">
							<div class="fighter_result_data">
								<a href="/fighter/Vinicius-Oliveira-139517"><span itemprop="name">Vinicius<br />Oliveira</span></a>
							</div>
						</div>
					</td>
					<td></td>
				</tr>
			</tbody>
		</table>
	</body>
</html>`
