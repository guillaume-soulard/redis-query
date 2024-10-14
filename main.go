package main

import (
	"fmt"
	"github.com/lucasjones/reggen"
	"strings"
)

func main() {
	params := parseParameters()
	Run(&params)
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
