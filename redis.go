package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
)

func connectToRedis(params parameters) (client *redis.Client) {

	if *params.SentinelAddrs != "" {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    strings.Split(*params.SentinelAddrs, ","),
			SentinelUsername: *params.SentinelUser,
			SentinelPassword: *params.SentinelPassword,
			MasterName:       *params.SentinelMaster,
			Username:         *params.User,
			Password:         *params.Password,
			DB:               *params.Db,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", *params.Host, *params.Port),
			Username: *params.User,
			Password: *params.Password,
			DB:       *params.Db,
		})
	}
	var err error
	pong, err := client.Ping(context.Background()).Result()
	if err != nil || pong != "PONG" {
		log.Fatal(err)
	}
	return client
}
