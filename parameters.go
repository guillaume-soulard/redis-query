package main

import (
	"errors"
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
}

type EnvCommand struct {
	Name              *string
	ConnectParameters ConnectParameters
	Cmd               *argparse.Command
}

type ConnectParameters struct {
	Host             *string
	Db               *int
	Port             *int
	SentinelMaster   *string
	User             *string
	Password         *string
	SentinelAddrs    *string
	SentinelUser     *string
	SentinelPassword *string
}

type LoopCommand struct {
	LoopFrom *int
	LoopTo   *int
	LoopStep *int
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

	if err := parser.Parse(os.Args); err != nil {
		PrintErrorAndExit(errors.New(parser.Usage(nil)))
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
	parameters.SentinelUser = command.String("", "sentinel-user", &argparse.Options{Required: false, Help: "redis sentinel user name"})
	parameters.SentinelPassword = command.String("", "sentinel-password", &argparse.Options{Required: false, Help: "redis sentinel password"})
	parameters.SentinelAddrs = command.String("", "sentinel-addrs", &argparse.Options{Required: false, Help: "redis sentinel addresses : <host1>:<port>,<host2>:<port>,...."})
}

func setFormat(parameters *FormatParameters, command *argparse.Command) {
	parameters.Format = command.String("", "format", &argparse.Options{Required: false, Help: "format stdin with some variables : {stdin} = stdin value, {row} = row number from 1 to n, {result} = the result of the command"})
}
