package main

import (
	"log"
	"net/http"

	"github.com/chrpa-jakub/sherdog-api/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	fighterService := fighter.Service(fighterservice.NewSherdogService())
	eventService := event.Service(eventservice.NewSherdogService())

	if !cfg.CacheDisabled {
		log.Print("cache enabled; connecting to redis")
		redisOptions, err := goredis.ParseURL(cfg.DBConn)
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
