package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

type parameters struct {
	Host             *string
	Db               *int
	Port             *int
	Scan             *string
	Format           *string
	ScanCount        *int
	SentinelMaster   *string
	User             *string
	Password         *string
	SentinelAddrs    *string
	SentinelUser     *string
	SentinelPassword *string
	PipelineSize     *int
	Env              *string `json:"-"`
	SetEnv           *string `json:"-"`
}

func parseParameters() parameters {
	parser := argparse.NewParser("rq", "redis query tool")
	parser.SetHelp("", "help")
	parser.ExitOnHelp(true)
	var params parameters

	params.Host = parser.String("h", "host", &argparse.Options{Required: false, Help: "redis host", Default: "127.0.0.1"})
	params.Port = parser.Int("p", "port", &argparse.Options{Required: false, Help: "redis port", Default: 6379})

	params.User = parser.String("u", "user", &argparse.Options{Required: false, Help: "redis user name"})
	params.Password = parser.String("w", "password", &argparse.Options{Required: false, Help: "redis password"})
	params.Db = parser.Int("d", "db", &argparse.Options{Required: false, Help: "redis db index", Default: 0})

	params.SentinelMaster = parser.String("", "sentinel-master", &argparse.Options{Required: false, Help: "redis sentinel master name"})
	params.SentinelUser = parser.String("", "sentinel-user", &argparse.Options{Required: false, Help: "redis sentinel user name"})
	params.SentinelPassword = parser.String("", "sentinel-password", &argparse.Options{Required: false, Help: "redis sentinel password"})
	params.SentinelAddrs = parser.String("", "sentinel-addrs", &argparse.Options{Required: false, Help: "redis sentinel addresses : <host1>:<port>,<host2>:<port>,...."})

	params.Scan = parser.String("s", "scan", &argparse.Options{Required: false, Help: "scan pattern"})
	params.ScanCount = parser.Int("l", "scan-len", &argparse.Options{Required: false, Help: "scan pattern", Default: 10})
	params.PipelineSize = parser.Int("P", "pipeline-len", &argparse.Options{Required: false, Help: "scan pattern", Default: 1})

	params.Format = parser.String("", "format", &argparse.Options{Required: false, Help: "format stdin with some variables : {stdin} = stdin value, {row} = row number from 1 to n"})

	params.Env = parser.String("e", "env", &argparse.Options{Required: false, Help: "environment to use"})
	params.SetEnv = parser.String("", "set-env", &argparse.Options{Required: false, Help: "environment to save with provided settings"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(nil))
		os.Exit(1)
	}
	return params
}
