package fighter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSherdogServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fighter/97529" {
			t.Fatalf("path = %q, want /fighter/97529", r.URL.Path)
		}
		w.Write([]byte(fighterHTML))
	}))
	t.Cleanup(server.Close)

	service := NewSherdogServiceWithBaseURL(server.URL)
	got, err := service.Get(context.Background(), "97529")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if got.Name != "Jiri Prochazka" {
		t.Fatalf("Name = %q, want Jiri Prochazka", got.Name)
	}
	if got.Nickname != "BJP" || got.Origin != "Czech Republic" {
		t.Fatalf("unexpected title fields: %+v", got)
	}
	if got.Age != 31 || got.Height != "193.04" || got.Weight != "92.99" {
		t.Fatalf("unexpected physical fields: %+v", got)
	}
	if got.Wins.Total != 29 || got.Wins.KO != 25 || got.Losses.Total != 4 {
		t.Fatalf("unexpected outcomes: %+v %+v", got.Wins, got.Losses)
	}
	if len(got.Fights) != 1 {
		t.Fatalf("len(Fights) = %d, want 1", len(got.Fights))
	}
	if got.UpcomingFight != nil {
		t.Fatalf("UpcomingFight = %+v, want nil", got.UpcomingFight)
	}
	fight := got.Fights[0]
	if fight.Opponent.Name != "Alex Pereira" || fight.Event.Name != "UFC 295" || fight.Round != 2 {
		t.Fatalf("unexpected fight: %+v", fight)
	}
}

func TestSherdogServiceGetUpcomingFight(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fighter/101189" {
			t.Fatalf("path = %q, want /fighter/101189", r.URL.Path)
		}
		w.Write([]byte(fighterUpcomingHTML))
	}))
	t.Cleanup(server.Close)

	service := NewSherdogServiceWithBaseURL(server.URL)
	got, err := service.Get(context.Background(), "101189")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if got.UpcomingFight == nil {
		t.Fatal("UpcomingFight = nil, want upcoming fight")
	}
	if got.UpcomingFight.Opponent.Name != "Kyoji Horiguchi" {
		t.Fatalf("Opponent.Name = %q, want Kyoji Horiguchi", got.UpcomingFight.Opponent.Name)
	}
	if got.UpcomingFight.Event.Name != "UFC Fight Night 279" {
		t.Fatalf("Event.Name = %q, want UFC Fight Night 279", got.UpcomingFight.Event.Name)
	}
	if got.UpcomingFight.Event.URL != "/events/UFC-Fight-Night-279-Kape-vs-Horiguchi-2-112139" {
		t.Fatalf("Event.URL = %q, want event URL", got.UpcomingFight.Event.URL)
	}
	if got.UpcomingFight.Date != "2026-06-20T00:00:00+00:00" {
		t.Fatalf("Date = %q, want ISO start date", got.UpcomingFight.Date)
	}
	if got.UpcomingFight.Location != "UFC Apex, Las Vegas, Nevada, United States" {
		t.Fatalf("Location = %q, want location", got.UpcomingFight.Location)
	}
}

func TestCachedServiceUsesCacheHit(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	store.values["fighter:97529"] = `{"name":"Cached Fighter"}`
	service := NewCachedService(errorService{}, store)

	got, err := service.Get(context.Background(), "97529")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Name != "Cached Fighter" {
		t.Fatalf("Name = %q, want Cached Fighter", got.Name)
	}
}

func TestCachedServiceCachesMiss(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	service := NewCachedService(staticService{fighter: &Fighter{Name: "Parsed Fighter"}}, store)

	got, err := service.Get(context.Background(), "97529")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Name != "Parsed Fighter" {
		t.Fatalf("Name = %q, want Parsed Fighter", got.Name)
	}
	if store.values["fighter:97529"] == "" {
		t.Fatal("expected fighter to be cached")
	}
}

type staticService struct {
	fighter *Fighter
}

func (s staticService) Get(context.Context, string) (*Fighter, error) {
	return s.fighter, nil
}

type errorService struct{}

func (errorService) Get(context.Context, string) (*Fighter, error) {
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

const fighterHTML = `
<html>
	<body>
		<div class="fighter-title">
			<span class="fn">Jiri Prochazka</span>
			<span class="nickname"><em>BJP</em></span>
			<strong itemprop="nationality">Czech Republic</strong>
		</div>
		<div class="fighter-data">
			<table>
				<tr><td><b>31</b></td></tr>
				<tr><td>6'4&quot; 193.04 cm</td></tr>
				<tr><td>205 lbs 92.99 kg</td></tr>
			</table>
			<div class="association-class"><span>Jetsaam Gym Brno</span><a>Light Heavyweight</a></div>
			<div class="win"><span></span><span>29</span></div>
			<div class="wins">
				<div></div><div></div><div class="meter"><div class="pl">25</div></div>
				<div></div><div class="meter"><div class="pl">3</div></div>
				<div></div><div class="meter"><div class="pl">1</div></div>
				<div></div><div class="meter"><div class="pl">0</div></div>
			</div>
			<div class="lose"><span></span><span>4</span></div>
			<div class="loses">
				<div></div><div></div><div class="meter"><div class="pl">3</div></div>
				<div></div><div class="meter"><div class="pl">1</div></div>
				<div></div><div class="meter"><div class="pl">0</div></div>
				<div></div><div class="meter"><div class="pl">0</div></div>
			</div>
			<div class="draws"><span></span><span>1</span></div>
			<div class="nc"><span></span><span>0</span></div>
		</div>
		<table>
			<tr>
				<td><span class="final_result">loss</span></td>
				<td><a href="/fighter/Alex-Pereira-224511">Alex Pereira</a></td>
				<td><a href="/events/UFC-295-98494">UFC 295</a></td>
				<td class="winby"><b>TKO</b></td>
				<td>2</td>
				<td>4:08</td>
			</tr>
		</table>
	</body>
</html>`

const fighterUpcomingHTML = `
<html>
	<body>
		<div class="fighter-title">
			<span class="fn">Manel Kape</span>
			<span class="nickname"><em>Starboy</em></span>
			<strong itemprop="nationality">Angola</strong>
		</div>
		<div class="fighter-data">
			<table>
				<tr><td><b>32</b></td></tr>
				<tr><td>5'6&quot; 167.64 cm</td></tr>
				<tr><td>125 lbs 56.7 kg</td></tr>
			</table>
			<div class="association-class"><span>Xtreme Couture</span><a>Flyweight</a></div>
		</div>
		<div class="fight_card_preview" itemprop="performerIn">
			<h2 itemprop="name">UFC Fight Night 279</h2>
			<div class="date_location">
				<meta itemprop="startDate" content="2026-06-20T00:00:00+00:00"/>
				June 20, 2026
				<em itemprop="location"><span itemprop="name">UFC Apex</span>, <span itemprop="address">Las Vegas, Nevada, United States</span></em>
			</div>
			<div class="fight">
				<div class="fighter left_side" itemprop="performer">
					<a itemprop="url" href="/fighter/Manel-Kape-101189"></a>
					<h3><a href="/fighter/Manel-Kape-101189"><span itemprop="name">Manel Kape</span></a></h3>
				</div>
				<div class="fighter right_side" itemprop="performer">
					<a itemprop="url" href="/fighter/Kyoji-Horiguchi-64413"></a>
					<h3><a href="/fighter/Kyoji-Horiguchi-64413"><span itemprop="name">Kyoji Horiguchi</span></a></h3>
				</div>
			</div>
			<a href="/events/UFC-Fight-Night-279-Kape-vs-Horiguchi-2-112139" class="card_button">See entire fight card</a>
		</div>
	</body>
</html>`
