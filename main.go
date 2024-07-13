package main

import (
	"fmt"
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
	} else {
		Print(params.Parser.Usage(nil))
	}
}

func formatIfNeededAndPrint(row *int, stdin string, result interface{}, params *FormatParameters) {
	if *params.Format == "" {
		Print(result)
	} else {
		output := *params.Format
		output = strings.ReplaceAll(output, "{stdin}", stdin)
		output = strings.ReplaceAll(output, "{result}", fmt.Sprintf("%v", result))
		output = strings.ReplaceAll(output, "{row}", fmt.Sprintf("%d", *row))
		Print(output)
	}
	*row = *row + 1
}
