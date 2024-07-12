package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
	result := make(chan interface{}, 10)
	executor := NewExecutor(client, result, *params.Command.Pipeline, *params.Command.NoOutput)
	go func() {
		rowNumber := 0
		for r := range result {
			formatIfNeededAndPrint(&rowNumber, "", r, &params.Command.Format)
			executor.Done()
		}
	}()
	if needAtLeastOneStdInArgument(args) {
		needToExecuteCommand := true
		for needToExecuteCommand {
			doArgs := make([]interface{}, len(args))
			var staticArg *string
			for i, arg := range args {
				if strings.Contains(arg, "{?}") {
					if staticArg == nil {
						if needToExecuteCommand = scanner.Scan(); !needToExecuteCommand {
							break
						}
						text := scanner.Text()
						staticArg = &text
					}
					doArgs[i] = strings.ReplaceAll(arg, "{?}", *staticArg)
				} else if strings.Contains(arg, "{>}") {
					if needToExecuteCommand = scanner.Scan(); !needToExecuteCommand {
						break
					}
					doArgs[i] = strings.ReplaceAll(arg, "{>}", scanner.Text())
				} else {
					doArgs[i] = arg
				}
			}
			if !needToExecuteCommand {
				break
			}
			executor.executePipeline(doArgs)
		}
	} else {
		doArgs := make([]interface{}, len(args))
		for i, arg := range args {
			doArgs[i] = arg
		}
		executor.executePipeline(doArgs)
	}
	executor.Wait()
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func needAtLeastOneStdInArgument(args []string) bool {
	waitingForPipeParams := false
	for _, arg := range args {
		if strings.Contains(arg, "{?}") || strings.Contains(arg, "{>}") {
			waitingForPipeParams = true
			break
		}
	}
	return waitingForPipeParams
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
	rowNumber := 0
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
			if rowNumber >= *params.Scan.Limit {
				break
			}
			formatIfNeededAndPrint(&rowNumber, "", key, &params.Scan.Format)
		}
		if cursor == 0 || rowNumber >= *params.Scan.Limit {
			break
		}
	}
}

func formatIfNeededAndPrint(row *int, stdin string, result interface{}, params *FormatParameters) {
	if *params.Format == "" {
		fmt.Println(result)
	} else {
		output := *params.Format
		output = strings.ReplaceAll(output, "{stdin}", stdin)
		output = strings.ReplaceAll(output, "{result}", fmt.Sprintf("%v", result))
		output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", *row))
		fmt.Println(output)
	}
	*row = *row + 1
}
