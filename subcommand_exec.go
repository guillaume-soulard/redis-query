package main

import (
	"bufio"
	"os"
	"strings"
)

const fixArgumentPlaceHolder = "{?}"
const iteratorArgumentPlaceHolder = "{>}"

type ExecSubCommand struct{}

func (s ExecSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Command.Cmd.Happened()
}

func (s ExecSubCommand) Execute(parameters *Parameters) (err error) {
	client := connectToRedis(parameters.Command.Connect)
	result := make(chan Output, 10)
	executor := NewExecutor(client, result, *parameters.Command.Pipeline, *parameters.Command.NoOutput)
	go func() {
		rowNumber := 0
		for r := range result {
			formatIfNeededAndPrint(&rowNumber, r.Stdin, r.Out, &parameters.Command.Format)
			executor.Done()
		}
	}()
	if needAtLeastOneStdInArgument(&parameters.Command) {
		needToExecuteCommand := true
		scanner := bufio.NewScanner(os.Stdin)
		if err = scanner.Err(); err != nil {
			return err
		}
		for needToExecuteCommand {
			stdinArgs := make([]string, 0, 10)
			for _, command := range *parameters.Command.Commands {
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
							stdinArgs = append(stdinArgs, text)
						}
						doArgs[i] = strings.ReplaceAll(arg, fixArgumentPlaceHolder, *staticArg)
					} else if strings.Contains(arg, iteratorArgumentPlaceHolder) {
						if needToExecuteCommand = scanner.Scan(); !needToExecuteCommand {
							break
						}
						iteratorArg := scanner.Text()
						stdinArgs = append(stdinArgs, iteratorArg)
						doArgs[i] = strings.ReplaceAll(arg, iteratorArgumentPlaceHolder, iteratorArg)
					} else {
						doArgs[i] = arg
					}
				}
				if !needToExecuteCommand {
					break
				}
				executor.executePipeline(doArgs, stdinArgs)
			}
		}
	} else {
		for _, command := range *parameters.Command.Commands {
			args := ParseArguments(command)
			doArgs := make([]interface{}, len(args))
			for i, arg := range args {
				doArgs[i] = arg
			}
			executor.executePipeline(doArgs, nil)
		}
	}
	executor.executePipelineCommands()
	executor.Wait()
	return err
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
