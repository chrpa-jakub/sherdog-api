package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chrpa-jakub/sherdog-api/internal/database/redis"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	eventservice "github.com/chrpa-jakub/sherdog-api/internal/service/event"
	fighterservice "github.com/chrpa-jakub/sherdog-api/internal/service/fighter"
	transporthttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	log.Print("starting sherdog api")

	fighterService := fighter.Service(fighterservice.NewSherdogService())
	eventService := event.Service(eventservice.NewSherdogService())

	if !cacheDisabled() {
		log.Print("cache enabled; connecting to redis")
		redisOptions, err := goredis.ParseURL(os.Getenv("DB_CONN"))
		if err != nil {
			log.Fatalf("parse redis connection URL: %v", err)
		}

		database := redis.New(goredis.NewClient(redisOptions))
		fighterService = fighterservice.NewCachedService(fighterService, database)
		eventService = eventservice.NewCachedService(eventService, database)
	} else {
		log.Print("cache disabled")
	}

	router := transporthttp.NewRouter(fighterService, eventService)
	log.Print("listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

func cacheDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CACHE_DISABLED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
