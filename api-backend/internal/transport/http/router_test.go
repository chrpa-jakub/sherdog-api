package http_test

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	transporthttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http"
)

func TestGetFighter(t *testing.T) {
	router := transporthttp.NewRouter(
		fighterStub{fighter: &fighter.Fighter{Name: "Jiri Prochazka"}},
		eventStub{},
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(stdhttp.MethodGet, "/fighter/Jiri-Prochazka-97529", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
	}

	var got fighter.Fighter
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if got.Name != "Jiri Prochazka" {
		t.Fatalf("Name = %q, want Jiri Prochazka", got.Name)
	}
}

func TestGetEvent(t *testing.T) {
	router := transporthttp.NewRouter(
		fighterStub{},
		eventStub{event: &event.Event{Name: "UFC 295"}},
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(stdhttp.MethodGet, "/events/UFC-295-98494", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
	}

	var got event.Event
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if got.Name != "UFC 295" {
		t.Fatalf("Name = %q, want UFC 295", got.Name)
	}
}

func TestHandlerServiceError(t *testing.T) {
	router := transporthttp.NewRouter(fighterStub{err: errors.New("boom")}, eventStub{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(stdhttp.MethodGet, "/fighter/97529", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusBadRequest)
	}
}

type fighterStub struct {
	fighter *fighter.Fighter
	err     error
}

func (s fighterStub) Get(context.Context, string) (*fighter.Fighter, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.fighter, nil
}

type eventStub struct {
	event *event.Event
	err   error
}

func (s eventStub) Get(context.Context, string) (*event.Event, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.event, nil
}
