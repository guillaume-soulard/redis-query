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
	for i := from; i <= to; i += step {
		Print(fmt.Sprintf("%d", i))
	}
}
