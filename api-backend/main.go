package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"

	"github.com/chrpa-jakub/sherdog-api/caching"
	"github.com/chrpa-jakub/sherdog-api/event"
	"github.com/chrpa-jakub/sherdog-api/fighter"
	"github.com/redis/go-redis/v9"
)

type API struct {
    Collector *colly.Collector
    Router *gin.Engine
    Database *redis.Client
}

func (api *API) Serve(port string){
    api.Router.GET("/fighter/:id", api.GetFighter) 
    api.Router.GET("/events/:id", api.GetEvent) 
    fmt.Println("Running on port", port)
    api.Router.Run(port)
}

func (api *API) GetFighter(c *gin.Context) {
        id := caching.ShortenFighterId(c.Param("id"))
        fighterCached, err := caching.GetCachedFighter(id, api.Database)

        if err == nil {
            c.JSON(200, string(fighterCached))
            return
        }

        fighter, err := fighter.ParseFighter(id, api.Collector)

        if err != nil {
            c.JSON(400, "Rate limit!")
            return
        }

        fighterJson, err := json.Marshal(fighter)

        if err != nil {
            c.JSON(500, "Could not parse fighter.")
            return
        }

        fighterJsonString := string(fighterJson)
        caching.CacheFighter(id, fighterJsonString, api.Database)
        c.JSON(200, fighterJsonString)
}

func (api *API) GetEvent(c *gin.Context) {
        id := c.Param("id")
        eventCached, err := caching.GetCachedEvent(id, api.Database)

        if err == nil {
            c.JSON(200, string(eventCached))
            return
        }

        event, err := event.ParseEvent(id, api.Collector)

        if err != nil {
            c.JSON(400, "Rate limit!")
            return
        }

        eventJson , err := json.Marshal(event)

        if err != nil {
            c.JSON(500, "Could not parse event.")
            return
        }

        eventJsonString := string(eventJson)
        caching.CacheEvent(id, eventJsonString, api.Database)
        c.JSON(200, eventJsonString)
}

func main() {
    gin.SetMode(gin.ReleaseMode)
    redisConn, err := redis.ParseURL(os.Getenv("DB_CONN"))
    
    if err != nil {
        panic(err)
    }

    api := API{
        Collector: colly.NewCollector(),
        Router: gin.New(),
        Database: redis.NewClient(redisConn),
    }

    api.Serve(":8080")
}
