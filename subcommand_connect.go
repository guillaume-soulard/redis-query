package main

import (
	"context"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/go-redis/redis/v8"
	"slices"
	"strings"
)

type FindKeySpec struct {
	LastKey int
	KeyStep int
	Limit   int
}

type FindKeySearchType string

const (
	FindKeySearchTypeRange = "range"
	FindKeySearchTypeKeynum = "keynum"
	FindKeySearchTypeUnknown = "unknown"
)

https://redis.io/docs/latest/develop/reference/key-specs/
type FindKey struct {
	Type FindKeySearchType
	RangeSpec FindKeyRangeSpec
	KeyNumSpec FindKeyKeyNumSpec
}

type BeginSearchIndexSpec struct {
	Index int
}
type BeginSearchKeywordSpec struct {
	KeyWord string
	StartFrom int
}

type BeginSearchType string

const (
	BeginSearchTypeIndex = "index"
	BeginSearchTypeKeyword = "keyword"
	BeginSearchTypeUnknown = "unknown"
)

type BeginSearchSpec struct {
	Type      BeginSearchType
	IndexSpec BeginSearchIndexSpec
	KeyWordSpec BeginSearchKeywordSpec
}

type KeySpecification struct {
	BeginSearch BeginSearchSpec
	FindKey     FindKey
	Notes       string
}

type CommandInfo struct {
	Name              string
	FirstKey          int
	LastKey           int
	StepKey           int
	Tips              string
	KeySpecifications []KeySpecification
	SubCommands       map[string]CommandInfo
}

var commands []interface{}
var commandMap map[string]CommandInfo
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
	var result interface{}
	if result, err = client.Do(context.Background(), "COMMAND").Result(); err != nil {
		return err
	}
	commands = result.([]interface{})
	return err
}

func completer(d prompt.Document) []prompt.Suggest {
	items := strings.Split(d.Text, " ")
	var commandName string
	if len(items) > 0 {
		commandName = items[0]
	}
	isCurrentArgKey, commandExists := commandMap[commandName]
	s := make([]prompt.Suggest, 0)
	if len(items) > 1 && len(lastResult) > 0 {
		for _, r := range lastResult {
			s = append(s, prompt.Suggest{
				Text: r, Description: "",
			})
		}
	} else if len(items) == 1 {
		for _, command := range commands {
			info := command.([]interface{})
			commandName := fmt.Sprintf("%v", info[0])
			s = append(s, prompt.Suggest{
				Text: commandName, Description: fmt.Sprintf("%s command", commandName),
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
