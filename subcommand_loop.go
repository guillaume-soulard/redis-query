package main

import (
	"fmt"
	"math"
)

func loop(params Parameters) {
	from := 0
	to := math.MaxInt
	step := 1
	if params.Loop.LoopFrom != nil {
		from = *params.Loop.LoopFrom
	}
	if params.Loop.LoopTo != nil {
		to = *params.Loop.LoopTo
	}
	if params.Loop.LoopStep != nil {
		step = *params.Loop.LoopStep
	}
	row := 1
	for i := from; i <= to; i += step {
		formatIfNeededAndPrint(&row, "", fmt.Sprintf("%d", i), &params.Loop.Format)
	}
}
