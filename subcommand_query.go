package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/go-redis/redis/v8"
	"strconv"
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
		Clients:    make(map[string]*redis.Client),
		Parameters: parameters,
	}); err != nil {
		return err
	}
	printResult(result.Result)
	return err
}

type QueryContext struct {
	Clients         map[string]*redis.Client
	QueryParameters map[string]interface{}
	Env             *string
	Parameters      *Parameters
}

func (c *QueryContext) GetClient() (client *redis.Client, err error) {
	var exists bool
	var envToUse string
	if c.Env == nil {
		envToUse = "default"
	} else {
		envToUse = *c.Env
	}
	if client, exists = c.Clients[envToUse]; exists {
		return client, err
	}
	if envToUse == "default" {
		client = connectToRedis(c.Parameters.Query.Connect)
	} else {
		var connect ConnectParameters
		if err = loadEnv(&envToUse, &connect); err != nil {
			return client, err
		}
		client = connectToRedis(connect)
	}
	c.Clients[envToUse] = client
	return client, err
}

func (c *QueryContext) Copy() QueryContext {
	return QueryContext{
		Clients:         c.Clients,
		Parameters:      c.Parameters,
		QueryParameters: c.QueryParameters,
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
	BlockArgs       *BlockArgs        `[@@ "-" ">"]`
	Commands        []Command         `"{" (@@)* "}"`
	BlockProperties []BlockProperties `[@@ ("," @@)* ]`
	Db              *int
}

type BlockProperties struct {
	Db  *int    `("Db" @Number) |`
	Env *string `("ENV" @String)`
}

func (b *Block) Execute(executableContext QueryContext) (executableResult ExecutableResult, err error) {
	var commandResult ExecutableResult
	result := make([]interface{}, len(b.Commands))
	blockContext := executableContext.Copy()
	blockContext.Env = b.GetEnv()
	if err = b.SetDb(blockContext); err != nil {
		return commandResult, err
	}
	for i, command := range b.Commands {
		if commandResult, err = command.Execute(blockContext); err != nil {
			return executableResult, err
		}
		result[i] = commandResult.Result
	}
	if err = b.RestoreDb(blockContext); err != nil {
		return commandResult, err
	}
	executableResult = ExecutableResult{
		Result: result,
	}
	return executableResult, err
}

func (b *Block) GetEnv() *string {
	for _, property := range b.BlockProperties {
		if property.Env != nil {
			return property.Env
		}
	}
	return nil
}

func (b *Block) SetDb(executableContext QueryContext) (err error) {
	dbToSwitch := b.GetDbToSwitch()
	if dbToSwitch != nil {
		var client *redis.Client
		if client, err = executableContext.GetClient(); err != nil {
			return err
		}
		var infos interface{}
		if infos, err = client.Do(context.Background(), "CLIENT", "INFO").Result(); err != nil {
			return err
		}
		for _, v := range strings.Split(infos.(string), " ") {
			item := strings.Split(v, "=")
			if item[0] == "db" {
				var originalDb int
				if originalDb, err = strconv.Atoi(item[1]); err != nil {
					return err
				}
				b.Db = &originalDb
			}
		}
		err = client.Do(context.Background(), "SELECT", *dbToSwitch).Err()
	}
	return err
}

func (b *Block) RestoreDb(queryContext QueryContext) (err error) {
	if b.Db != nil {
		var client *redis.Client
		if client, err = queryContext.GetClient(); err != nil {
			return err
		}
		err = client.Do(context.Background(), "SELECT", *b.Db).Err()
	}
	return err
}

func (b *Block) GetDbToSwitch() *int {
	if b.BlockProperties != nil {
		for _, property := range b.BlockProperties {
			if property.Db != nil {
				return property.Db
			}
		}
	}
	return nil
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
	var cmdResult interface{}
	var client *redis.Client
	if client, err = executableContext.GetClient(); err != nil {
		return executableResult, err
	}
	if cmdResult, err = client.Do(context.Background(), args...).Result(); err != nil {
		return executableResult, err
	}
	if c.Block != nil {
		if c.Block.BlockArgs != nil && c.Block.BlockArgs.Args != nil && len(c.Block.BlockArgs.Args) > 0 {
			blockContext := executableContext.Copy()
			if array, isArray := cmdResult.([]interface{}); isArray {
				if blockContext.QueryParameters == nil {
					blockContext.QueryParameters = make(map[string]interface{})
				}
				paramIndex := 0
				resultArray := make([]interface{}, 0, len(array)/len(c.Block.BlockArgs.Args))
				readParams := 0
				for {
					if len(array) == 0 {
						break
					}
					for _, arg := range c.Block.BlockArgs.Args {
						blockContext.QueryParameters[arg] = array[paramIndex]
						paramIndex++
						readParams++
						if paramIndex >= len(blockContext.QueryParameters) {
							var subResult ExecutableResult
							if subResult, err = c.Block.Execute(blockContext); err != nil {
								return executableResult, err
							}
							resultArray = append(resultArray, subResult.Result)
						}
						delete(blockContext.QueryParameters, arg)
					}
					if readParams >= len(array) {
						break
					}
				}
				executableResult.Result = resultArray
			} else {
				blockContext.QueryParameters = map[string]interface{}{c.Block.BlockArgs.Args[0]: cmdResult}
				executableResult, err = c.Block.Execute(blockContext)
			}
		} else {
			executableResult, err = c.Block.Execute(executableContext)
		}
	} else {
		executableResult = ExecutableResult{
			Result: cmdResult,
		}
	}
	return executableResult, err
}

type Variable struct {
	String   *string `(@String`
	Literal  *string `| @Ident`
	Variable *string `| ("#" @Ident))`
}

func (c *Variable) Execute(queryContext QueryContext) (executableResult ExecutableResult, err error) {
	if c.String != nil {
		value := *c.String
		if strings.Contains(value, "#") {
			for k, v := range queryContext.QueryParameters {
				value = strings.ReplaceAll(value, fmt.Sprintf("#%s", k), fmt.Sprintf("%v", v))
			}
		}
		executableResult = ExecutableResult{
			Result: value,
		}
	} else if c.Literal != nil {
		executableResult = ExecutableResult{
			Result: *c.Literal,
		}
	} else if c.Variable != nil {
		if value, hasValue := queryContext.QueryParameters[*c.Variable]; hasValue {
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
		{"Punct", `[,#{}>-]`},
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

func printResult(result interface{}) (output []string) {
	output = make([]string, 0)
	if array, isArray := result.([]interface{}); isArray {
		for _, item := range array {
			output = append(output, printResult(item)...)
		}
	} else {
		Print(result)
		output = append(output, fmt.Sprintf("%v", result))
	}
	return output
}
