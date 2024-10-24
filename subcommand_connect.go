package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/go-redis/redis/v8"
	"os"
	"os/exec"
	"slices"
	"strconv"
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

var allCommandDocMap map[string]CommandDoc
var lastResult = make([]string, 0)
var actualRedisVersion int

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
	if err = json.Unmarshal([]byte(commandDocJson), &allCommandDocMap); err != nil {
		return err
	}
	if actualRedisVersion, err = getActualRedisVersion(client); err != nil {
		return err
	}
	return err
}

func filterCommandsByVersion() (filteredCommands map[string]CommandDoc, err error) {
	filteredCommands = make(map[string]CommandDoc)
	var commandSince int
	for commandName, commandDoc := range allCommandDocMap {
		if commandSince, err = getRedisVersionInt(commandDoc.Since); err != nil {
			return filteredCommands, err
		}
		if actualRedisVersion >= commandSince {
			filteredCommands[commandName] = commandDoc
		}
	}
	return filteredCommands, err
}

func getActualRedisVersion(client *redis.Client) (version int, err error) {
	var infos string
	if infos, err = client.Info(context.Background()).Result(); err != nil {
		return version, err
	}
	for _, info := range strings.Split(infos, "\r\n") {
		infoSplit := strings.Split(info, ":")
		if len(infoSplit) >= 2 && infoSplit[0] == "redis_version" {
			if version, err = getRedisVersionInt(infoSplit[1]); err != nil {
				return version, err
			}
		}
	}
	return version, err
}

func getRedisVersionInt(versionStr string) (version int, err error) {
	version = 0
	version, err = strconv.Atoi(strings.ReplaceAll(versionStr, ".", ""))
	return version, err
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
		var filteredCommands map[string]CommandDoc
		var err error
		if filteredCommands, err = filterCommandsByVersion(); err != nil {
			PrintErrorAndExit(err)
		}
		for name, command := range filteredCommands {
			s = append(s, prompt.Suggest{
				Text:        name,
				Description: getCommandDescription(name, command),
			})
		}
	}
	slices.SortFunc(s, func(a, b prompt.Suggest) int {
		return strings.Compare(a.Text, b.Text)
	})
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func getCommandDescription(commandName string, command CommandDoc) string {
	return fmt.Sprintf(
		"%s %s. %s Complexity : %s",
		strings.ToUpper(commandName),
		strings.Join(getArgumentsString(command.Arguments), " "),
		command.Summary,
		command.Complexity,
	)
}

func getArgumentsString(arguments []CommandDocArguments) []string {
	parts := make([]string, len(arguments))
	for i, argument := range arguments {
		if argument.Type == "pure-token" {
			parts[i] = strings.ToUpper(argument.Token)
		} else if argument.Type == "oneof" {
			parts[i] = strings.Join(getArgumentsString(argument.Arguments), " | ")
		} else {
			parts[i] = argument.Name
		}
		if argument.Type != "pure-token" && argument.Token != "" {
			parts[i] = fmt.Sprintf("%s %s", argument.Token, parts[i])
		}
		if flagsContains(argument.Flags, "optional") {
			parts[i] = fmt.Sprintf("[%s]", parts[i])
		}
	}
	return parts
}

func flagsContains(flags []string, value string) bool {
	for _, flag := range flags {
		if flag == value {
			return true
		}
	}
	return false
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
		if result, err = client.Do(context.Background(), args...).Result(); err != nil {
			printResult(err)
		} else {
			lastResult = printResult(result)
		}
	}
	return err
}
