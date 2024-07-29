package main

import (
	"fmt"
	"github.com/lucasjones/reggen"
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
		executeCommands(params)
	} else if params.Query.Cmd.Happened() {
		executeQuery(params)
	} else {
		Print(params.Parser.Usage(nil))
	}
}

var generator, _ = reggen.NewGenerator("[0-9a-zA-Z]{5,50}")

func init() {
	generator.SetSeed(1)
}

func formatIfNeededAndPrint(row *int, stdin string, result interface{}, params *FormatParameters) {
	if items, ok := result.([]interface{}); ok {
		for _, item := range items {
			formatIfNeededAndPrint(row, stdin, item, params)
		}
	} else {
		if *params.Format == "" {
			Print(result)
		} else {
			output := *params.Format
			output = strings.ReplaceAll(output, "{stdin}", stdin)
			output = strings.ReplaceAll(output, "{result}", fmt.Sprintf("%v", result))
			output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", *row))
			if strings.Contains(*params.Format, "{random}") {
				output = strings.ReplaceAll(output, "{random}", generator.Generate(9999))
			}
			Print(output)
		}
		*row = *row + 1
	}
}
