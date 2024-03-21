package fighter

import (
    "strings"

    "github.com/gocolly/colly"
    "github.com/chrpa-jakub/sherdog-api/converts"
)

type Fighter struct {
    Name string `json:"name"`
    Nickname string `json:"nickname"`
    Origin string  `json:"origin"`
    Age int `json:"age"`
    Height float64 `json:"height"`
    Weight float64 `json:"weight"`
    WeightClass string `json:"weightclass"`
    Gym string `json:"gym"`
    Wins Outcome `json:"win"`
    Loses Outcome `json:"loses"`
    Draws int `json:"draws"`
    NoContests int `json:"nocontests"`
    Fights []Fight `json:"fights"`

}

type Event struct {
    Name string `json:"name"`
    Url string `json:"url"`
}

type Outcome struct {
    Total int `json:"total"`
    Ko int `json:"ko"`
    Submission int `json:"submission"`
    Decision int `json:"decision"`
    Others int `json:"others"`
}

type Oponent struct {
    Name string `json:"name"`
    Url string `json:"url"`
}

type Fight struct {
    Outcome string `json:"outcome"`
    Oponent Oponent `json:"oponent"`
    Event Event `json:"event"`
    Way string `json:"way"`
    Round int `json:"round"`
    Time string `json:"time"`
}

func ParseFighter(id string, c *colly.Collector) (*Fighter, error) {
    originalFighter := &Fighter{}

    c.OnHTML("div.fighter-title", func(h *colly.HTMLElement) {
        fighter := originalFighter

        fighter.Name = h.ChildText("span.fn")

        fighter.Nickname = h.ChildText("span.nickname > em")

        fighter.Origin = h.ChildText("strong[itemprop=nationality]")
    })

    c.OnHTML("div.fighter-data", func(h *colly.HTMLElement) {
        fighter := originalFighter 

        fighter.Age = converts.ToInt(h.ChildText("tr:nth-child(1) > td > b"))

        heightRaw := strings.Split(h.ChildText("tr:nth-child(2) > td"), " ")
        fighter.Height = converts.ToFloat64(heightRaw[len(heightRaw)-2])

        weightRaw := strings.Split(h.ChildText("tr:nth-child(3) > td"), " ")
        fighter.Weight = converts.ToFloat64(weightRaw[len(weightRaw)-2])

        fighter.Gym = h.ChildText("div.association-class > span")

        fighter.WeightClass= h.ChildText("div.association-class > a")


        fighter.Wins = Outcome{
            Total: converts.ToInt(h.ChildText("div.win > span:nth-child(2)")),
            Ko: converts.ToInt(h.ChildText("div.wins > div.meter:nth-child(3) > div.pl")),
            Submission: converts.ToInt(h.ChildText("div.wins > div.meter:nth-child(5) > div.pl")),
            Decision: converts.ToInt(h.ChildText("div.wins > div.meter:nth-child(7) > div.pl")),
            Others: converts.ToInt(h.ChildText("div.wins > div.meter:nth-child(9) > div.pl")),
        }

        fighter.Loses = Outcome{
            Total: converts.ToInt(h.ChildText("div.lose > span:nth-child(2)")),
            Ko: converts.ToInt(h.ChildText("div.loses > div.meter:nth-child(3) > div.pl")),
            Submission: converts.ToInt(h.ChildText("div.loses > div.meter:nth-child(5) > div.pl")),
            Decision: converts.ToInt(h.ChildText("div.loses > div.meter:nth-child(7) > div.pl")),
            Others: converts.ToInt(h.ChildText("div.loses > div.meter:nth-child(9) > div.pl")),
        }

        fighter.Draws= converts.ToInt(h.ChildText("div.draws > span:nth-child(2)"))

        fighter.NoContests = converts.ToInt(h.ChildText("div.nc > span:nth-child(2)"))
    })

    c.OnHTML("tr", func(h *colly.HTMLElement) {
        fighter := originalFighter 

        fight := Fight {
            Outcome: h.ChildText("span.final_result"),
            Oponent: Oponent{
                Name: h.ChildText("td:nth-child(2) > a"),
                Url: h.ChildAttr("a", "href"),
            },
            Event: Event{
                Name: h.ChildText("td:nth-child(3) > a"),
                Url: h.ChildAttr("td:nth-child(3) > a", "href"),
            },
            Way: h.ChildText("td.winby > b"),
            Round: converts.ToInt(h.ChildText("td:nth-child(5)")),
            Time: h.ChildText("td:last-child"),
        }

        if len(fight.Event.Name) < 3  {
            return
        }

        fighter.AddFight(fight)
    })



    err := c.Visit("https://sherdog.com/fighter/"+id)
    if err != nil {
        return nil, err
    }

    return originalFighter, nil  
}

func (f *Fighter) AddFight(fight Fight){
    f.Fights = append(f.Fights, fight)
}
