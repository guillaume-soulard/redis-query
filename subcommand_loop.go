package main

import (
	"fmt"
	"math"
)

type LoopSubCommand struct{}

func (l LoopSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Loop.Cmd.Happened()
}

func (l LoopSubCommand) Execute(parameters *Parameters) (err error) {
	from := 0
	to := math.MaxInt
	step := 1
	if parameters.Loop.LoopFrom != nil {
		from = *parameters.Loop.LoopFrom
	}
	if parameters.Loop.LoopTo != nil {
		to = *parameters.Loop.LoopTo
	}
	if parameters.Loop.LoopStep != nil {
		step = *parameters.Loop.LoopStep
	}
	row := 1
	for i := from; i <= to; i += step {
		formatIfNeededAndPrint(&row, "", fmt.Sprintf("%d", i), &parameters.Loop.Format)
	}
	return err
}
