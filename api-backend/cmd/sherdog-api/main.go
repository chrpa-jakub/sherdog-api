package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if err := run(); err != nil {
		log.Fatalf("sherdog api: %v", err)
	}
}

func run() error {
	log.Print("starting sherdog api")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	fighterService := fighter.Service(fighterservice.NewSherdogService())
	eventService := event.Service(eventservice.NewSherdogService())

	if !cfg.CacheDisabled {
		log.Print("cache enabled; connecting to redis")
		redisOptions, err := goredis.ParseURL(cfg.DBConn)
		if err != nil {
			return fmt.Errorf("parse redis connection URL: %w", err)
		}

		database := redis.New(goredis.NewClient(redisOptions))
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("close redis: %v", err)
			}
		}()
		fighterService = fighterservice.NewCachedService(fighterService, database)
		eventService = eventservice.NewCachedService(eventService, database)
	} else {
		log.Print("cache disabled")
	}

	router := transporthttp.NewRouter(fighterService, eventService)
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Print("listening on :8080")
		serverErr <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
		stop()
		log.Print("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	log.Print("server stopped")

	return nil
}
