package main

import (
	"context"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/go-redis/redis/v8"
	"os"
	"os/exec"
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

var history []string
var globalClient *redis.Client

func executor(command string) {
	if len(history) == 0 || history[len(history)-1] != command {
		history = append(history, command)
	}
	var result interface{}
	if strings.ToLower(command) == "exit" {
		rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
		rawModeOff.Stdin = os.Stdin
		_ = rawModeOff.Run()
		if err := rawModeOff.Wait(); err != nil {
			PrintErrorAndExit(err)
		}
		os.Exit(0)
	}
	argsStr := strings.Split(command, " ")
	args := make([]interface{}, len(argsStr))
	for i, a := range argsStr {
		args[i] = a
	}
	var err error
	if result, err = globalClient.Do(context.Background(), args...).Result(); err != nil {
		printResult(err)
	} else {
		lastResult = printResult(result)
	}
}

func showPrompt(client *redis.Client) (err error) {
	history = make([]string, 0, 10)
	globalClient = client
	p := prompt.New(
		executor,
		completer,
		prompt.OptionHistory(history),
		prompt.OptionCompletionOnDown(),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionPrefix("redis>"),
	)
	p.Run()
	return err
}
