package main

import (
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

type Parameters struct {
	SetEnv      EnvCommand
	DelEnv      EnvCommand
	ListEnv     EnvCommand
	DescribeEnv EnvCommand
	Parser      *argparse.Parser
	Loop        LoopCommand
	Scan        ScanCommand
	Command     CommandCommand
	Query       QueryCommand
	Connect     ConnectCommand
	Migrate     MigrateCommand
	Rdb         RdbCommand
}

type RdbCommand struct {
	Cmd        *argparse.Command
	InputFiles *[]string
	OutputFile *string
	KeyPattern *string
	Update     RdbCommandUpdateTtl
}

type RdbCommandUpdateTtl struct {
	Ttl *int
	Cmd *argparse.Command
}

type MigrateCommand struct {
	SourceEnv     *string
	SourceConnect ConnectParameters
	TargetEnv     *string
	TargetConnect ConnectParameters
	SourcePattern *string
	Count         *int
	Ttl           *int
	Replace       *bool
	Cmd           *argparse.Command
	Limit         *int
}

type EnvCommand struct {
	Name              *string
	ConnectParameters ConnectParameters
	Cmd               *argparse.Command
}

type ConnectParameters struct {
	Host           *string
	Db             *int
	Port           *int
	SentinelMaster *string
	User           *string
	Password       *string
	SentinelAddrs  *string
}

type LoopCommand struct {
	LoopFrom *int
	LoopTo   *int
	LoopStep *int
	Format   FormatParameters
	Cmd      *argparse.Command
}

type ScanCommand struct {
	EnvName   *string
	Pattern   *string
	Count     *int
	Limit     *int
	Type      *string
	Connect   ConnectParameters
	Format    FormatParameters
	Cmd       *argparse.Command
	KeyToScan *string
}

type CommandCommand struct {
	EnvName  *string
	Pipeline *int
	Commands *[]string
	Connect  ConnectParameters
	Format   FormatParameters
	Cmd      *argparse.Command
	NoOutput *bool
}

type FormatParameters struct {
	Format *string
}

type QueryCommand struct {
	EnvName *string
	Query   *string
	Connect ConnectParameters
	Format  FormatParameters
	Cmd     *argparse.Command
}

type ConnectCommand struct {
	EnvName *string
	Connect ConnectParameters
	Cmd     *argparse.Command
}

func parseParameters() Parameters {
	parser := argparse.NewParser("rq", "redis command line query tool")
	parser.SetHelp("", "help")
	parser.ExitOnHelp(true)
	var params Parameters

	params.Parser = parser

	configCommand := parser.NewCommand("env", "configure environment")
	configSetCommand := configCommand.NewCommand("set", "configure the environment")
	params.SetEnv.Name = configSetCommand.String("", "name", &argparse.Options{Required: true, Help: "env name"})
	params.SetEnv.Cmd = configSetCommand
	setConnect(&params.SetEnv.ConnectParameters, configSetCommand)

	configDelCommand := configCommand.NewCommand("remove", "remove the environment")
	params.DelEnv.Name = configDelCommand.String("", "name", &argparse.Options{Required: true, Help: "env name"})
	params.DelEnv.Cmd = configDelCommand

	configListCommand := configCommand.NewCommand("list", "list environments")
	params.ListEnv.Cmd = configListCommand

	configDescribeCommand := configCommand.NewCommand("describe", "describe the environment")
	params.DescribeEnv.Name = configDescribeCommand.String("", "name", &argparse.Options{Required: true, Help: "env name"})
	params.DescribeEnv.Cmd = configDescribeCommand

	loopCommand := parser.NewCommand("loop", "loop from integers")
	params.Loop.LoopFrom = loopCommand.Int("", "from", &argparse.Options{Required: false, Default: nil, Help: "loop from the provided number"})
	params.Loop.LoopTo = loopCommand.Int("", "to", &argparse.Options{Required: false, Help: "loop to the provided number"})
	params.Loop.LoopStep = loopCommand.Int("", "step", &argparse.Options{Required: false, Default: 1, Help: "loop step if loop from is provided"})
	setFormat(&params.Loop.Format, loopCommand)
	params.Loop.Cmd = loopCommand

	scanCommand := parser.NewCommand("scan", "scan the redis instance or a scanable key iterativly")
	params.Scan.Pattern = scanCommand.String("", "pattern", &argparse.Options{Required: true, Help: "scan pattern"})
	params.Scan.Count = scanCommand.Int("c", "count", &argparse.Options{Required: false, Help: "scan count argument for scan command", Default: 10})
	params.Scan.Limit = scanCommand.Int("l", "limit", &argparse.Options{Required: false, Help: "limit the number of keys to return", Default: -1})
	params.Scan.Type = scanCommand.String("t", "type", &argparse.Options{Required: false, Help: "type of key to scan : string, list, set, zset, hash and stream"})
	params.Scan.EnvName = scanCommand.String("e", "env", &argparse.Options{Required: false, Help: "environment name to use"})
	params.Scan.KeyToScan = scanCommand.String("k", "key", &argparse.Options{Required: false, Help: "key to scan (set, hash or sorted set)"})
	setFormat(&params.Scan.Format, scanCommand)
	params.Scan.Cmd = scanCommand
	setConnect(&params.Scan.Connect, scanCommand)

	commandCommand := parser.NewCommand("exec", "execute a redis command")
	params.Command.Commands = commandCommand.StringList("c", "command", &argparse.Options{Required: false, Help: "command to run on redis instance"})
	params.Command.Pipeline = commandCommand.Int("P", "pipeline", &argparse.Options{Required: false, Help: "pipeline len to use for server interaction", Default: 1})
	params.Command.EnvName = commandCommand.String("e", "env", &argparse.Options{Required: false, Help: "environment name to use"})
	params.Command.NoOutput = commandCommand.Flag("", "no-output", &argparse.Options{Required: false, Help: "no output for command"})
	setFormat(&params.Command.Format, commandCommand)
	setConnect(&params.Command.Connect, commandCommand)
	params.Command.Cmd = commandCommand

	queryCommand := parser.NewCommand("query", "execute a rql query")
	params.Query.Query = queryCommand.String("q", "query", &argparse.Options{Required: true, Help: "the rql query to execute"})
	params.Query.EnvName = queryCommand.String("e", "env", &argparse.Options{Required: false, Help: "environment name to use"})
	setFormat(&params.Query.Format, queryCommand)
	setConnect(&params.Query.Connect, queryCommand)
	params.Query.Cmd = queryCommand

	connectCommand := parser.NewCommand("connect", "connect to the right redis instance")
	params.Connect.EnvName = connectCommand.String("e", "env", &argparse.Options{Required: false, Help: "environment name to use"})
	setConnect(&params.Connect.Connect, connectCommand)
	params.Connect.Cmd = connectCommand

	migrateCommand := parser.NewCommand("migrate", "migrate keys from one redis to another redis instance")
	params.Migrate.SourceEnv = migrateCommand.String("s", "source-env", &argparse.Options{Required: true, Help: "source environment name to use"})
	params.Migrate.TargetEnv = migrateCommand.String("t", "target-env", &argparse.Options{Required: true, Help: "target environment name to use"})
	params.Migrate.SourcePattern = migrateCommand.String("p", "source-pattern", &argparse.Options{Required: true, Help: "source keys pattern to dump"})
	params.Migrate.Count = migrateCommand.Int("c", "count", &argparse.Options{Required: false, Help: "source scan command count arg"})
	params.Migrate.Limit = migrateCommand.Int("l", "limit", &argparse.Options{Required: false, Help: "maximum number of keys to migrate"})
	params.Migrate.Ttl = migrateCommand.Int("", "ttl", &argparse.Options{Required: false, Help: "the ttl to use for restore (by default use the source ttl)"})
	params.Migrate.Replace = migrateCommand.Flag("r", "replace", &argparse.Options{Required: false, Help: "replace the key on target if exists on target. If not specify the 'Target key name is busy' error is ignored"})
	params.Migrate.Cmd = migrateCommand

	rdbCommand := parser.NewCommand("rdb", "tool for rdb analyse and manipulations")
	rdbSetTtlCommand := rdbCommand.NewCommand("set-ttl", "Update rdb data")
	params.Rdb.InputFiles = rdbSetTtlCommand.StringList("i", "input-file", &argparse.Options{Required: true, Help: "file to use for the rdb tool"})
	params.Rdb.OutputFile = rdbSetTtlCommand.String("o", "output-file", &argparse.Options{Required: true, Help: "file to write"})
	params.Rdb.KeyPattern = rdbSetTtlCommand.String("k", "key-pattern", &argparse.Options{Required: false, Help: "redis key pattern to process"})
	params.Rdb.Update.Ttl = rdbSetTtlCommand.Int("t", "ttl", &argparse.Options{Required: true, Help: "the ttl in seconds to set to selected keys (-1 if no expire)"})
	params.Rdb.Update.Cmd = rdbSetTtlCommand
	params.Rdb.Cmd = rdbCommand

	if err := parser.Parse(os.Args); err != nil {
		PrintErrorAndExit(errors.New(fmt.Sprintf("%s \n\n %s", err.Error(), parser.Usage(nil))))
	}
	return params
}

func setConnect(parameters *ConnectParameters, command *argparse.Command) {
	parameters.Host = command.String("h", "host", &argparse.Options{Required: false, Help: "redis host", Default: "127.0.0.1"})
	parameters.Port = command.Int("p", "port", &argparse.Options{Required: false, Help: "redis port", Default: 6379})
	parameters.User = command.String("u", "user", &argparse.Options{Required: false, Help: "redis user name"})
	parameters.Password = command.String("w", "password", &argparse.Options{Required: false, Help: "redis password"})
	parameters.Db = command.Int("d", "db", &argparse.Options{Required: false, Help: "redis db index", Default: 0})
	parameters.SentinelMaster = command.String("", "sentinel-master", &argparse.Options{Required: false, Help: "redis sentinel master name"})
	parameters.SentinelAddrs = command.String("", "sentinel-addrs", &argparse.Options{Required: false, Help: "redis sentinel addresses : <host1>:<port>,<host2>:<port>,...."})
}

func setFormat(parameters *FormatParameters, command *argparse.Command) {
	parameters.Format = command.String("", "format", &argparse.Options{Required: false, Help: "format stdin with some variables : {stdin} = stdin value, {row} = row number from 1 to n, {result} = the result of the command, {random} = a random value"})
}
