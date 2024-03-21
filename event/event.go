package event

import (
    "github.com/gocolly/colly"
    "github.com/chrpa-jakub/sherdog-api/converts"
)


type Event struct { 
    Name string `json:"name"`
    Date string `json:"date"`
    Organization string `json:"organization"`
    Fights []Fight `json:"fights"`
}

type Fight struct {
    Fighters []Fighter `json:"fighters"`
    Way string `json:"way"`
    Round int `json:"round"`
    Time string `json:"time"`
    WeightClass string `json:"weightclass"`
}

type Fighter struct {
    Name string `json:"name"`
    Url string `json:"url"`
    Outcome string `json:"outcome"`
}

func ParseEvent(id string, c *colly.Collector) (*Event, error) {
    originalEvent := &Event{}

    c.OnHTML("div.left-module-pad-line", func(h *colly.HTMLElement) {
        event := originalEvent

        event.Name = h.ChildText("h1 > span")
        event.Organization = h.ChildText("div > a > span")
        event.Date = h.ChildText("div.info > span:nth-child(1)")

        wayRaw := h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(2)")
        roundRaw := h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(4)")
        timeRaw := h.ChildText("table.fight_card_resume > tbody > tr > td:nth-child(5)")

        fight := Fight {
            WeightClass: h.ChildText("span.weight_class"),
            Fighters: []Fighter{
                {
                    Name: h.ChildText("div.fighter.left_side > h3 > a > span[itemprop=name]"),
                    Url: h.ChildAttr("div.fighter.left_side > h3 > a", "href"),
                    Outcome: h.ChildText("div.fighter.left_side > span.final_result"),
                },
                {
                    Name: h.ChildText("div.fighter.right_side > h3 > a > span[itemprop=name]"),
                    Url: h.ChildAttr("div.fighter.right_side > h3 > a", "href"),
                    Outcome: h.ChildText("div.fighter.right_side > span.final_result"),
                },
            },
            Way: wayRaw[7:],
            Round: converts.ToInt(roundRaw[6:]),
            Time:  timeRaw[5:],
        }
        event.AddFight(fight)
    })

    c.OnHTML("tr", func(h *colly.HTMLElement) {
        event := originalEvent
        fight := Fight{
            WeightClass: h.ChildText("span.weight_class"),
            Way: h.ChildText("td.winby > b"),
            Round: converts.ToInt(h.ChildText("td:nth-child(6)")),
            Time: h.ChildText("td:last-child"),

            Fighters: []Fighter{
                {
                    Name: h.ChildAttr("div.fighter_list.left > img", "title"),
                    Url: h.ChildAttr("div.fighter_list.left > div > a", "href"),
                    Outcome: h.ChildText("div.fighter_list.left > div > span.final_result"),
                },
                {
                    Name: h.ChildAttr("div.fighter_list.right > img", "title"),
                    Url: h.ChildAttr("div.fighter_list.right > div > a", "href"),
                    Outcome: h.ChildText("div.fighter_list.right > div > span.final_result"),
                },
            },
        }

        if len(fight.Fighters[0].Outcome) == 0 {
            return
        }

        event.AddFight(fight)
    })


    err := c.Visit("https://sherdog.com/events/"+id)
    if err != nil {
        return nil, err
    }
    return originalEvent, nil
}

func (event *Event) AddFight(fight Fight){
   event.Fights = append(event.Fights, fight) 
}

