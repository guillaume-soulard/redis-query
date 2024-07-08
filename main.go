package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"math"
	"os"
	"strings"
)

func main() {
	params := parseParameters()
	if params.Command.Cmd.Happened() && *params.Command.EnvName != "" {
		loadEnv(&params)
	}
	if params.SetEnv.Cmd.Happened() {
		saveEnv(params)
	} else if params.DelEnv.Cmd.Happened() {
		delEnv(params)
	} else if params.ListEnv.Cmd.Happened() {
		listEnv()
	} else if params.DescribeEnv.Cmd.Happened() {
		describeEnv(params)
	} else if params.Format.Cmd.Happened() {
		format(params)
	} else if params.Scan.Cmd.Happened() {
		scan(params)
	} else if params.Loop.Cmd.Happened() {
		loop(params)
	} else if params.Command.Cmd.Happened() {
		executeCommand(params)
	} else {
		fmt.Println(params.Parser.Usage(nil))
	}
}

func loop(params Parameters) {
	from := 0
	to := math.MaxInt
	step := 1
	if params.Loop.LoopFrom != nil {
		from = *params.Loop.LoopFrom
	}
	if params.Loop.LoopTo != nil {
		to = *params.Loop.LoopTo
	}
	if params.Loop.LoopStep != nil {
		step = *params.Loop.LoopStep
	}
	for i := from; i <= to; i += step {
		fmt.Println(fmt.Sprintf("%d", i))
	}
}

func executeCommand(params Parameters) {
	client := connectToRedis(params.Command.Connect)
	scanner := bufio.NewScanner(os.Stdin)
	args := strings.Split(*params.Command.Command, " ")
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
				PrintErrorAndExit(err)
			}
			pipelineCount++
			if pipelineCount >= *params.Command.Pipeline {
				if cmds, err := pipeline.Exec(context.Background()); err != nil {
					PrintErrorAndExit(err)
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
			PrintErrorAndExit(err)
		}
		if cmds, err := pipeline.Exec(context.Background()); err != nil {
			PrintErrorAndExit(err)
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

func scan(params Parameters) {
	client := connectToRedis(params.Scan.Connect)
	cursor := uint64(0)
	var err error
	var result []string
	var keyType string
	key := *params.Scan.KeyToScan
	if key != "" {
		if keyType, err = client.Type(context.Background(), key).Result(); err != nil {
			PrintErrorAndExit(err)
		}
	}
	for {
		if keyType == "set" {
			if result, cursor, err = client.SScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "zset" {
			if result, cursor, err = client.ZScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "hash" {
			if result, cursor, err = client.HScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "" {
			if result, cursor, err = client.Scan(context.Background(), cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else {
			PrintErrorAndExit(errors.New(fmt.Sprintf("Unable to scan key type: %s", keyType)))
		}
		for _, key := range result {
			fmt.Println(key)
		}
		if cursor == 0 {
			break
		}
	}
}

func format(params Parameters) {
	scanner := bufio.NewScanner(os.Stdin)
	row := int64(1)
	for scanner.Scan() {
		output := *params.Format.Format
		text := scanner.Text()
		output = strings.ReplaceAll(output, "{stdin}", text)
		output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", row))
		fmt.Println(output)
		row++
	}
}
