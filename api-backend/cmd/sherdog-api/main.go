package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chrpa-jakub/sherdog-api/internal/database/redis"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	transporthttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	fighterService := fighter.Service(fighter.NewSherdogService())
	eventService := event.Service(event.NewSherdogService())

	if !cacheDisabled() {
		redisOptions, err := goredis.ParseURL(os.Getenv("DB_CONN"))
		if err != nil {
			log.Fatal(err)
		}

		database := redis.New(goredis.NewClient(redisOptions))
		fighterService = fighter.NewCachedService(fighterService, database)
		eventService = event.NewCachedService(eventService, database)
	}

	router := transporthttp.NewRouter(fighterService, eventService)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
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
