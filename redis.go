package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
)

type NoLog struct{}

func (NoLog) Printf(_ context.Context, _ string, _ ...interface{}) {}

func connectToRedis(params ConnectParameters) (client *redis.Client) {
	redis.SetLogger(NoLog{})
	if *params.SentinelAddrs != "" {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: strings.Split(*params.SentinelAddrs, ","),
			MasterName:    *params.SentinelMaster,
			Username:      *params.User,
			Password:      *params.Password,
			DB:            *params.Db,
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
		PrintErrorAndExit(err)
	}
	return client
}
