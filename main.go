package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"strings"
)

func main() {
	params := parseParameters()
	if *params.Env != "" {
		loadEnv(&params)
	}
	if *params.SetEnv != "" {
		saveEnv(params)
	} else if *params.Format != "" {
		format(params)
	} else if *params.Scan != "" {
		scan(params)
	} else {
		executeCommand(params)
	}
}

func executeCommand(params parameters) {
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
			if pipelineCount >= *params.PipelineSize {
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

func scan(params parameters) {
	client := connectToRedis(params)
	cursor := uint64(0)
	var err error
	var keys []string
	for {
		if keys, cursor, err = client.Scan(context.Background(), cursor, *params.Scan, int64(*params.ScanCount)).Result(); err != nil {
			log.Fatal(err)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
		if cursor == 0 {
			break
		}
	}
}

func format(params parameters) {
	scanner := bufio.NewScanner(os.Stdin)
	row := int64(1)
	for scanner.Scan() {
		output := *params.Format
		text := scanner.Text()
		output = strings.ReplaceAll(output, "{stdin}", text)
		output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", row))
		fmt.Println(output)
		row++
	}
}
