package main

import (
	"context"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/go-redis/redis/v8"
	"slices"
	"strings"
)

var commandMap map[string]*redis.CommandInfo
var lastResult = make([]string, 0)

type ConnectSubCommand struct{}

func (q ConnectSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Connect.Cmd.Happened()
}

func (q ConnectSubCommand) Execute(parameters *Parameters) (err error) {
	client := connectToRedis(parameters.Connect.Connect)
	if err = loadCompletion(client); err != nil {
		return err
	}
	err = showPrompt(client)
	return err
}

func loadCompletion(client *redis.Client) (err error) {
	commandMap, err = client.Command(context.Background()).Result()
	return err
}

func completer(d prompt.Document) []prompt.Suggest {
	items := strings.Split(d.Text, " ")
	s := make([]prompt.Suggest, 0)
	if len(items) > 1 && len(lastResult) > 0 {
		for _, r := range lastResult {
			s = append(s, prompt.Suggest{
				Text: r, Description: "",
			})
		}
	} else if len(items) == 1 {
		for _, info := range commandMap {
			s = append(s, prompt.Suggest{
				Text: info.Name, Description: fmt.Sprintf("%s command", info.Name),
			})
		}
	}
	slices.SortFunc(s, func(a, b prompt.Suggest) int {
		return strings.Compare(a.Text, b.Text)
	})
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func showPrompt(client *redis.Client) (err error) {
	history := make([]string, 0, 10)
	for {
		command := prompt.Input("redis> ", completer,
			prompt.OptionHistory(history),
			prompt.OptionCompletionOnDown(),
			prompt.OptionShowCompletionAtStart(),
		)
		history = append(history, command)
		var result interface{}
		if strings.ToLower(command) == "exit" {
			break
		}
		argsStr := strings.Split(command, " ")
		args := make([]interface{}, len(argsStr))
		for i, a := range argsStr {
			args[i] = a
		}
		if result, err = client.Do(context.Background(), args...).Result(); err != nil {
			printResult(err)
		} else {
			lastResult = printResult(result)
		}
	}
	return err
}
