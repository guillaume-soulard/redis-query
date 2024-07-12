package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"sync"
)

type Executor struct {
	pipeline      redis.Pipeliner
	result        chan interface{}
	pipelineMax   int
	pipelineCount int
	wg            sync.WaitGroup
	noOutput      bool
}

func NewExecutor(client *redis.Client, resultChan chan interface{}, pipelineMax int, noOutput bool) Executor {
	return Executor{
		pipeline:    client.Pipeline(),
		result:      resultChan,
		pipelineMax: pipelineMax,
		wg:          sync.WaitGroup{},
		noOutput:    noOutput,
	}
}

func (e *Executor) executePipeline(args []interface{}) {
	if _, err := e.pipeline.Do(context.Background(), args...).Result(); err != nil {
		PrintErrorAndExit(err)
	}
	e.pipelineCount++
	if e.pipelineCount >= e.pipelineMax {
		if cmds, err := e.pipeline.Exec(context.Background()); err != nil {
			PrintErrorAndExit(err)
		} else {
			if !e.noOutput {
				e.wg.Add(len(cmds))
				for _, cmd := range cmds {
					e.result <- cmd.(*redis.Cmd).Val()
				}
			}
		}
		e.pipelineCount = 0
	}
}

func (e *Executor) Wait() {
	e.wg.Wait()
}

func (e *Executor) Done() {
	e.wg.Done()
}
