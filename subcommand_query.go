package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/go-redis/redis/v8"
	"strings"
)

type QuerySubCommand struct{}

func (q QuerySubCommand) Accept(parameters *Parameters) bool {
	return parameters.Query.Cmd.Happened()
}

func (q QuerySubCommand) Execute(parameters *Parameters) (err error) {
	var executable Executable
	if executable, err = Parse(*parameters.Query.Query); err != nil {
		return err
	}
	var result ExecutableResult
	if result, err = executable.Execute(QueryContext{
		Client: connectToRedis(parameters.Query.Connect),
	}); err != nil {
		return err
	}
	printResult(result.Result)
	return err
}

type QueryContext struct {
	Client     *redis.Client
	Parameters map[string]interface{}
}

func (c QueryContext) Copy() QueryContext {
	return QueryContext{
		Client: c.Client,
	}
}

type ExecutableResult struct {
	Result interface{}
}

type Executable interface {
	Execute(executableContext QueryContext) (executableResult ExecutableResult, err error)
}

type Query struct {
	Block Block `@@`
}

func (q *Query) Execute(executableContext QueryContext) (executableResult ExecutableResult, err error) {
	return q.Block.Execute(executableContext)
}

type Block struct {
	BlockArgs *BlockArgs `[@@ "-" ">"]`
	Commands  []Command  `"{" (@@)* "}"`
}

func (b *Block) Execute(executableContext QueryContext) (executableResult ExecutableResult, err error) {
	var commandResult ExecutableResult
	result := make([]interface{}, len(b.Commands))
	for i, command := range b.Commands {
		if commandResult, err = command.Execute(executableContext); err != nil {
			return executableResult, err
		}
		result[i] = commandResult.Result
	}
	executableResult = ExecutableResult{
		Result: result,
	}
	return executableResult, err
}

type BlockArgs struct {
	Args []string `@Ident ("," @Ident)*`
}

type Command struct {
	Name  *string    `@Ident`
	Args  []Variable `@@*`
	Block *Block     `["-" ">" @@]`
}

func (c *Command) Execute(executableContext QueryContext) (executableResult ExecutableResult, err error) {
	args := make([]interface{}, len(c.Args)+1)
	args[0] = *c.Name
	var argResult ExecutableResult
	for i, arg := range c.Args {
		if argResult, err = arg.Execute(executableContext); err != nil {
			return executableResult, err
		}
		if argResult.Result != nil {
			args[i+1] = argResult.Result
		}
	}
	var cmdresult interface{}
	if cmdresult, err = executableContext.Client.Do(context.Background(), args...).Result(); err != nil {
		return executableResult, err
	}
	if c.Block != nil {
		if c.Block.BlockArgs != nil && c.Block.BlockArgs.Args != nil && len(c.Block.BlockArgs.Args) > 0 {
			blockContext := executableContext.Copy()
			if array, isArray := cmdresult.([]interface{}); isArray {
				blockContext.Parameters = make(map[string]interface{})
				paramIndex := 0
				resultArray := make([]interface{}, 0, len(array)/len(c.Block.BlockArgs.Args))
				readParams := 0
				for {
					for _, arg := range c.Block.BlockArgs.Args {
						blockContext.Parameters[arg] = array[paramIndex]
						paramIndex++
						readParams++
						if paramIndex >= len(blockContext.Parameters) {
							var subResult ExecutableResult
							if subResult, err = c.Block.Execute(blockContext); err != nil {
								return executableResult, err
							}
							resultArray = append(resultArray, subResult.Result)
							paramIndex = 0
						}
					}
					if paramIndex >= len(blockContext.Parameters) {
						var subResult ExecutableResult
						if subResult, err = c.Block.Execute(blockContext); err != nil {
							return executableResult, err
						}
						resultArray = append(resultArray, subResult.Result)
						paramIndex = 0
					}
					if readParams >= len(array) {
						break
					}
				}
				executableResult.Result = resultArray
			} else {
				blockContext.Parameters = map[string]interface{}{"": cmdresult}
				executableResult, err = c.Block.Execute(blockContext)
			}
		} else {
			executableResult, err = c.Block.Execute(executableContext)
		}
	} else {
		executableResult = ExecutableResult{
			Result: cmdresult,
		}
	}
	return executableResult, err
}

type Variable struct {
	String   *string `(@String`
	Variable *string `| ("$" @Ident))`
}

func (c *Variable) Execute(queryContext QueryContext) (executableResult ExecutableResult, err error) {
	if c.String != nil {
		value := *c.String
		if strings.Contains(value, "$") {
			for k, v := range queryContext.Parameters {
				value = strings.ReplaceAll(value, fmt.Sprintf("$%s", k), fmt.Sprintf("%v", v))
			}
		}
		executableResult = ExecutableResult{
			Result: value,
		}
	} else if c.Variable != nil {
		if value, hasValue := queryContext.Parameters[*c.Variable]; hasValue {
			executableResult = ExecutableResult{
				Result: value,
			}
		} else {
			err = errors.New(fmt.Sprintf("undefined variable %s", *c.Variable))
		}
	} else {
		err = errors.New("illegal state")
	}
	return executableResult, err
}

var (
	rqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Command`, `[a-zA-Z][a-zA-Z0-9]*`},
		{"Punct", `[${}>-]`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{"whitespace", `\s+`},
	})
	parser = participle.MustBuild[Query](
		participle.Lexer(rqlLexer),
		participle.Unquote("String"),
		participle.UseLookahead(2),
		participle.CaseInsensitive("Ident"),
	)
)

func Parse(query string) (block *Query, err error) {
	block, err = parser.ParseString("", query)
	return block, err
}

func printResult(result interface{}) {
	if array, isArray := result.([]interface{}); isArray {
		for _, item := range array {
			printResult(item)
		}
	} else {
		fmt.Println(result)
	}
}
