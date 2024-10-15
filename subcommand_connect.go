package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"strings"
)

type ConnectSubCommand struct{}

func (q ConnectSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Connect.Cmd.Happened()
}

func (q ConnectSubCommand) Execute(parameters *Parameters) (err error) {
	client := connectToRedis(parameters.Connect.Connect)
	err = showPrompt(parameters, client)
	return err
}

func showPrompt(parameters *Parameters, client *redis.Client) (err error) {
	r := bufio.NewReader(os.Stdin)
	for {
		var s string
		for {
			if _, err = fmt.Fprint(os.Stderr, fmt.Sprintf("%s> ", "redis")); err != nil {
				return err
			}
			s, _ = r.ReadString('\n')
			if s != "" {
				break
			}
		}
		// TODO history with arrow up and down
		// TODO completion
		var result interface{}
		cleanCommand := strings.TrimSuffix(s, "\n")
		if strings.ToLower(cleanCommand) == "exit" {
			break
		}
		argsStr := strings.Split(cleanCommand, " ")
		args := make([]interface{}, len(argsStr))
		for i, a := range argsStr {
			args[i] = a
		}
		if result, err = client.Do(context.Background(), args...).Result(); err != nil {
			printResult(err)
		} else {
			printResult(result)
		}
	}
	return err
}
