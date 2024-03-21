package caching

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func ShortenFighterId(id string) string {
    if strings.Contains(id, "-") {
        idSplit := strings.Split(id, "-")
        id = idSplit[len(idSplit)-1]
    }

    return id
}


func CacheFighter(id string, json string, client *redis.Client) {
    err := client.Set(context.Background(), "fighter:"+id, json, time.Hour).Err()

    fmt.Println("Fighter", id, "cached.")

    if err != nil {
        panic(err)
    }
}

func GetCachedFighter(id string, client *redis.Client) (string, error) {
    json, err :=  client.Get(context.Background(), "fighter:"+id).Result()

    if len(json) == 0 {
        return "", errors.New("Not found")
    }

    if err != nil {
        return "", err
    }

    return json, nil
}

func CacheEvent(id string, json string, client *redis.Client) {
    err := client.Set(context.Background(), "event:"+id, json, time.Hour).Err()

    fmt.Println("Event", id, "cached.")

    if err != nil {
        panic(err)
    }
}

func GetCachedEvent(id string, client *redis.Client) (string, error) {
    json, err :=  client.Get(context.Background(), "event:"+id).Result()

    if len(json) == 0 {
        return "", errors.New("Not found")
    }

    if err != nil {
        return "", err
    }

    return json, nil
}
