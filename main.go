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
		loadEnv(&params, *params.Command.EnvName)
	}
	if params.Scan.Cmd.Happened() && *params.Scan.EnvName != "" {
		loadEnv(&params, *params.Scan.EnvName)
	}
	if params.SetEnv.Cmd.Happened() {
		saveEnv(params)
	} else if params.DelEnv.Cmd.Happened() {
		delEnv(params)
	} else if params.ListEnv.Cmd.Happened() {
		listEnv()
	} else if params.DescribeEnv.Cmd.Happened() {
		describeEnv(params)
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
	rowNumber := int64(0)
	if waitingForPipeParams {
		pipeline := client.Pipeline()
		pipelineCount := 0
		stdin := ""
		for scanner.Scan() {
			doArgs := make([]interface{}, len(args))
			for i, arg := range args {
				if arg == "?" {
					text := scanner.Text()
					stdin += text
					doArgs[i] = text
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
						formatIfNeededAndPrint(&rowNumber, stdin, cmd.(*redis.Cmd).Val(), &params.Command.Format)
						stdin = ""
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
				formatIfNeededAndPrint(&rowNumber, "", cmd.(*redis.Cmd).Val(), &params.Command.Format)
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
	rowNumber := int64(0)
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
			if *params.Scan.Type != "" {
				if result, cursor, err = client.ScanType(context.Background(), cursor, *params.Scan.Pattern, int64(*params.Scan.Count), *params.Scan.Type).Result(); err != nil {
					PrintErrorAndExit(err)
				}
			} else {
				if result, cursor, err = client.Scan(context.Background(), cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
					PrintErrorAndExit(err)
				}
			}
		} else {
			PrintErrorAndExit(errors.New(fmt.Sprintf("Unable to scan key type: %s", keyType)))
		}
		for _, key = range result {
			formatIfNeededAndPrint(&rowNumber, "", key, &params.Scan.Format)
		}
		if cursor == 0 {
			break
		}
	}
}

func formatIfNeededAndPrint(row *int64, stdin string, result interface{}, params *FormatParameters) {
	if *params.Format == "" {
		fmt.Println(result)
	} else {
		output := *params.Format
		output = strings.ReplaceAll(output, "{stdin}", stdin)
		output = strings.ReplaceAll(output, "{result}", fmt.Sprintf("%v", result))
		output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", *row))
		*row = *row + 1
		fmt.Println(output)
	}
}
