package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"

	"github.com/chrpa-jakub/sherdog-api/event"
	"github.com/chrpa-jakub/sherdog-api/fighter"
)

type API struct {
    Collector *colly.Collector
    Router *gin.Engine
}

func (api *API) Serve(port string){
    api.Router.GET("/fighter/:id", api.GetFighter) 
    api.Router.GET("/events/:id", api.GetEvent) 
    api.Router.Run(port)
}

func (api *API) GetFighter(c *gin.Context) {
        fighter, err := fighter.ParseFighter(c.Param("id"), api.Collector)

        if err != nil {
            c.JSON(400, "Rate limit!")
            return
        }

        fighterJson, err := json.Marshal(fighter)

        if err != nil {
            c.JSON(500, "Could not parse fighter.")
            return
        }

        c.JSON(200, string(fighterJson))
}

func (api *API) GetEvent(c *gin.Context) {
        event, err := event.ParseEvent(c.Param("id"), api.Collector)

        if err != nil {
            c.JSON(400, "Rate limit!")
            return
        }

        eventJson , err := json.Marshal(event)

        if err != nil {
            c.JSON(500, "Could not parse event.")
            return
        }

        c.JSON(200, string(eventJson))
}

func main() {
    gin.SetMode(gin.ReleaseMode)

    api := API{
        Collector: colly.NewCollector(),
        Router: gin.New(),
    }

    api.Serve(":8080")
}
