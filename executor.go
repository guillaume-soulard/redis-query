package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strings"
	"sync"
)

type Executor struct {
	pipeline      redis.Pipeliner
	result        chan Output
	pipelineMax   int
	pipelineCount int
	wg            sync.WaitGroup
	noOutput      bool
	stdin         [][]string
}

type Output struct {
	Stdin string
	Out   interface{}
}

func NewExecutor(client *redis.Client, resultChan chan Output, pipelineMax int, noOutput bool) Executor {
	return Executor{
		pipeline:    client.Pipeline(),
		result:      resultChan,
		pipelineMax: pipelineMax,
		wg:          sync.WaitGroup{},
		noOutput:    noOutput,
		stdin:       make([][]string, 0, 100),
	}
}

func (e *Executor) executePipeline(args []interface{}, stdinArgs []string) {
	if _, err := e.pipeline.Do(context.Background(), args...).Result(); err != nil {
		PrintErrorAndExit(err)
	}
	e.stdin = append(e.stdin, stdinArgs)
	e.pipelineCount++
	if e.pipelineCount >= e.pipelineMax {
		e.executePipelineCommands()
	}
}

func (e *Executor) executePipelineCommands() {
	if cmds, err := e.pipeline.Exec(context.Background()); err != nil {
		PrintErrorAndExit(err)
	} else {
		if !e.noOutput {
			e.wg.Add(len(cmds))
			for i, cmd := range cmds {
				output := Output{
					Stdin: strings.Join(e.stdin[i], "\t"),
					Out:   cmd.(*redis.Cmd).Val(),
				}
				e.result <- output
			}
		}
	}
	e.stdin = e.stdin[:0]
	e.pipelineCount = 0
}

func (e *Executor) Wait() {
	e.wg.Wait()
}

func (e *Executor) Done() {
	e.wg.Done()
}
