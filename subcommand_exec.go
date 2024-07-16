package main

import (
	"bufio"
	"os"
	"strings"
)

const fixArgumentPlaceHolder = "{?}"
const iteratorArgumentPlaceHolder = "{>}"

func executeCommands(params Parameters) {
	client := connectToRedis(params.Command.Connect)
	result := make(chan interface{}, 10)
	executor := NewExecutor(client, result, *params.Command.Pipeline, *params.Command.NoOutput)
	go func() {
		rowNumber := 0
		for r := range result {
			formatIfNeededAndPrint(&rowNumber, "", r, &params.Command.Format)
			executor.Done()
		}
	}()
	if needAtLeastOneStdInArgument(&params.Command) {
		needToExecuteCommand := true
		scanner := bufio.NewScanner(os.Stdin)
		if err := scanner.Err(); err != nil {
			PrintErrorAndExit(err)
		}
		for needToExecuteCommand {
			for _, command := range *params.Command.Commands {
				args := ParseArguments(command)
				doArgs := make([]interface{}, len(args))
				var staticArg *string
				for i, arg := range args {
					if strings.Contains(arg, fixArgumentPlaceHolder) {
						if staticArg == nil {
							if needToExecuteCommand = scanner.Scan(); !needToExecuteCommand {
								break
							}
							text := scanner.Text()
							staticArg = &text
						}
						doArgs[i] = strings.ReplaceAll(arg, fixArgumentPlaceHolder, *staticArg)
					} else if strings.Contains(arg, iteratorArgumentPlaceHolder) {
						if needToExecuteCommand = scanner.Scan(); !needToExecuteCommand {
							break
						}
						doArgs[i] = strings.ReplaceAll(arg, iteratorArgumentPlaceHolder, scanner.Text())
					} else {
						doArgs[i] = arg
					}
				}
				if !needToExecuteCommand {
					break
				}
				executor.executePipeline(doArgs)
			}
		}
	} else {
		for _, command := range *params.Command.Commands {
			args := ParseArguments(command)
			doArgs := make([]interface{}, len(args))
			for i, arg := range args {
				doArgs[i] = arg
			}
			executor.executePipeline(doArgs)
		}
	}
	executor.executePipelineCommands()
	executor.Wait()
}

func needAtLeastOneStdInArgument(commands *CommandCommand) bool {
	waitingForPipeParams := false
	for _, command := range *commands.Commands {
		args := ParseArguments(command)
		for _, arg := range args {
			if strings.Contains(arg, fixArgumentPlaceHolder) || strings.Contains(arg, iteratorArgumentPlaceHolder) {
				waitingForPipeParams = true
				break
			}
		}
	}
	return waitingForPipeParams
}
