package main

import (
	"context"
	"encoding/json"
	"github.com/c-bata/go-prompt"
	"github.com/go-redis/redis/v8"
	"slices"
	"strings"
)

// redis-cli --json command docs | jq > commands.json

type CommandDoc struct {
	Summary         string                `json:"summary"`
	Since           string                `json:"since"`
	Group           string                `json:"group"`
	Complexity      string                `json:"complexity"`
	History         [][]string            `json:"history"`
	Arguments       []CommandDocArguments `json:"arguments"`
	DocFlags        []string              `json:"doc_flags"`
	DeprecatedSince string                `json:"deprecated_since"`
	Replacedby      string                `json:"replaced_by"`
}

type CommandDocArguments struct {
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	DisplayText     string                `json:"display_text"`
	KeySpecIndex    int                   `json:"key_spec_index"`
	Flags           []string              `json:"flags"`
	Since           string                `json:"since"`
	Arguments       []CommandDocArguments `json:"arguments"`
	Token           string                `json:"token"`
	Summary         string                `json:"summary"`
	DeprecatedSince string                `json:"deprecated_since"`
	Value           string                `json:"value"`
}

var commandDocMap map[string]CommandDoc
var commands []interface{}
var lastResult = make([]string, 0)

type ConnectSubCommand struct{}

func (q ConnectSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Connect.Cmd.Happened()
}

func (q ConnectSubCommand) Execute(parameters *Parameters) (err error) {
	client := connectToRedis(parameters.Connect.Connect)
	if err = loadCompletion(); err != nil {
		return err
	}
	err = showPrompt(client)
	return err
}

func loadCompletion() (err error) {
	if err = json.Unmarshal([]byte(commandDocJson), &commandDocMap); err != nil {
		return err
	}
	return err
}

func completer(d prompt.Document) []prompt.Suggest {
	items := strings.Split(d.Text, " ")
	//var commandName string
	//if len(items) > 0 {
	//	commandName = items[0]
	//}
	//isCurrentArgKey, commandExists := commandDocMap[commandName]
	s := make([]prompt.Suggest, 0)
	if len(items) > 1 && len(lastResult) > 0 {
		for _, r := range lastResult {
			s = append(s, prompt.Suggest{
				Text: r, Description: "",
			})
		}
	} else if len(items) == 1 {
		for name, command := range commandDocMap {
			s = append(s, prompt.Suggest{
				Text: name, Description: command.Summary,
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
