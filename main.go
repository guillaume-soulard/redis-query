package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type parameters struct {
	Host             string
	Db               int
	Port             int
	Scan             string
	Format           string
	ScanCount        int64
	Sentinel         bool
	SentinelMaster   string
	User             string
	Password         string
	SentinelAddrs    string
	SentinelUser     string
	SentinelPassword string
	PipelineSize     int
	Env              string `json:"-"`
	SetEnv           string `json:"-"`
}

func main() {
	params := parseParameters()
	if params.Env != "" {
		loadEnv(&params)
	}
	if params.SetEnv != "" {
		saveEnv(params)
	} else if params.Format != "" {
		scanner := bufio.NewScanner(os.Stdin)
		row := int64(1)
		for scanner.Scan() {
			output := params.Format
			text := scanner.Text()
			output = strings.ReplaceAll(output, "{stdin}", text)
			output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", row))
			fmt.Println(output)
			row++
		}
	} else if params.Scan != "" {
		client := connectToRedis(params)
		cursor := uint64(0)
		var err error
		var keys []string
		for {
			if keys, cursor, err = client.Scan(context.Background(), cursor, params.Scan, params.ScanCount).Result(); err != nil {
				log.Fatal(err)
			}
			for _, key := range keys {
				fmt.Println(key)
			}
			if cursor == 0 {
				break
			}
		}
	} else {
		client := connectToRedis(params)
		scanner := bufio.NewScanner(os.Stdin)
		args := flag.Args()
		waitingForPipeParams := false
		for _, arg := range args {
			if arg == "?" {
				waitingForPipeParams = true
				break
			}
		}
		if waitingForPipeParams {
			pipeline := client.Pipeline()
			pipelineCount := 0
			for scanner.Scan() {
				doArgs := make([]interface{}, len(args))
				for i, arg := range args {
					if arg == "?" {
						doArgs[i] = scanner.Text()
					} else {
						doArgs[i] = arg
					}
				}
				if _, err := pipeline.Do(context.Background(), doArgs...).Result(); err != nil {
					panic(err)
				}
				pipelineCount++
				if pipelineCount >= params.PipelineSize {
					if cmds, err := pipeline.Exec(context.Background()); err != nil {
						panic(err)
					} else {
						for _, cmd := range cmds {
							fmt.Println(cmd.(*redis.Cmd).Val())
						}
					}
				}
			}
		} else {
			pipeline := client.Pipeline()
			doArgs := make([]interface{}, len(args))
			for i, arg := range args {
				doArgs[i] = arg
			}
			if _, err := pipeline.Do(context.Background(), doArgs...).Result(); err != nil {
				panic(err)
			}
			if cmds, err := pipeline.Exec(context.Background()); err != nil {
				panic(err)
			} else {
				for _, cmd := range cmds {
					fmt.Println(cmd.(*redis.Cmd).Val())
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}
}

func loadEnv(params *parameters) {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, params.Env)); err != nil {
			return
		} else {
			var loadedParams parameters
			if err = json.Unmarshal(file, &loadedParams); err != nil {
				panic(err)
			}
			setIfNotDefault(&params.Host, loadedParams.Host)
			setIfNotDefault(&params.Db, loadedParams.Db)
			setIfNotDefault(&params.Port, loadedParams.Port)
			setIfNotDefault(&params.Scan, loadedParams.Scan)
			setIfNotDefault(&params.ScanCount, loadedParams.ScanCount)
			setIfNotDefault(&params.Sentinel, loadedParams.Sentinel)
			setIfNotDefault(&params.SentinelMaster, loadedParams.SentinelMaster)
			setIfNotDefault(&params.User, loadedParams.User)
			setIfNotDefault(&params.Password, loadedParams.Password)
			setIfNotDefault(&params.SentinelAddrs, loadedParams.SentinelAddrs)
			setIfNotDefault(&params.SentinelUser, loadedParams.SentinelUser)
			setIfNotDefault(&params.SentinelPassword, loadedParams.SentinelPassword)
		}
	}
}

func setIfNotDefault[T comparable](param *T, loadedParameter T) {
	if loadedParameter != *new(T) {
		*param = loadedParameter
	}
}

func saveEnv(params parameters) {
	file, _ := json.MarshalIndent(params, "", " ")
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		if err = os.MkdirAll(fmt.Sprintf("%s/.redis-query", home), 0777); err != nil {
			panic(err)
		}
		if err = os.WriteFile(fmt.Sprintf("%s/.redis-query/%s.json", home, params.SetEnv), file, 0777); err != nil {
			panic(err)
		}
	}
}

func parseParameters() parameters {
	var params parameters
	flag.StringVar(&params.Host, "host", "127.0.0.1", "redis host")
	flag.IntVar(&params.Port, "port", 6379, "redis Port")

	flag.StringVar(&params.User, "user", "", "user name")
	flag.StringVar(&params.Password, "password", "", "password for user")
	flag.IntVar(&params.Db, "db", 0, "database to use default 0")

	flag.BoolVar(&params.Sentinel, "sentinel", false, "sentinel connection")
	flag.StringVar(&params.SentinelMaster, "sentinel-master", "mymaster", "redis sentinel master name")
	flag.StringVar(&params.SentinelAddrs, "sentinel-addrs", "127.0.0.1:26379", "redis sentinel addresses : <host1>:<port>,<host2>:<port>,....")
	flag.StringVar(&params.SentinelUser, "sentinel-user", "", "redis sentinel user")
	flag.StringVar(&params.SentinelPassword, "sentinel-password", "", "redis sentinel password")

	flag.StringVar(&params.Scan, "scan", "", "pattern to Scan")
	flag.Int64Var(&params.ScanCount, "scan-count", 10, "pattern to Scan")
	flag.IntVar(&params.PipelineSize, "pipeline-size", 1, "pipeline size to send to redis")

	flag.StringVar(&params.Format, "format", "", "format output : special variables : {stdin} = stdin value, {row} = row number from 1 to n")

	flag.StringVar(&params.Env, "env", "local", "environment to use")
	flag.StringVar(&params.SetEnv, "set-env", "", "environment to save with provided settings")

	flag.Parse()
	return params
}

func connectToRedis(params parameters) (client *redis.Client) {

	if params.Sentinel {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    strings.Split(params.SentinelAddrs, ","),
			SentinelUsername: params.SentinelUser,
			SentinelPassword: params.SentinelPassword,
			MasterName:       params.SentinelMaster,
			Username:         params.User,
			Password:         params.Password,
			DB:               params.Db,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", params.Host, params.Port),
			Username: params.User,
			Password: params.Password,
			DB:       params.Db,
		})
	}
	var err error
	pong, err := client.Ping(context.Background()).Result()
	if err != nil || pong != "PONG" {
		log.Fatal(err)
	}
	return client
}
