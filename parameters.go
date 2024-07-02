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
	Sentinel         *bool
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
	parser.SetHelp("help", "help")
	parser.ExitOnHelp(true)
	var params parameters

	params.Host = parser.String("h", "host", &argparse.Options{Required: false, Help: "redis host", Default: "127.0.0.1"})
	params.Port = parser.Int("p", "port", &argparse.Options{Required: false, Help: "redis port", Default: 6379})

	params.User = parser.String("u", "user", &argparse.Options{Required: false, Help: "redis user name"})
	params.Password = parser.String("w", "password", &argparse.Options{Required: false, Help: "redis password"})
	params.Db = parser.Int("d", "db", &argparse.Options{Required: false, Help: "redis db index", Default: 0})

	//flag.BoolVar(&params.Sentinel, "sentinel", false, "sentinel connection")
	//flag.StringVar(&params.SentinelMaster, "sentinel-master", "mymaster", "redis sentinel master name")
	//flag.StringVar(&params.SentinelAddrs, "sentinel-addrs", "127.0.0.1:26379", "redis sentinel addresses : <host1>:<port>,<host2>:<port>,....")
	//flag.StringVar(&params.SentinelUser, "sentinel-user", "", "redis sentinel user")
	//flag.StringVar(&params.SentinelPassword, "sentinel-password", "", "redis sentinel password")

	params.Scan = parser.String("s", "scan", &argparse.Options{Required: false, Help: "scan pattern"})
	params.ScanCount = parser.Int("l", "scan-len", &argparse.Options{Required: false, Help: "scan pattern", Default: 10})
	params.PipelineSize = parser.Int("P", "pipeline-len", &argparse.Options{Required: false, Help: "scan pattern", Default: 1})

	//flag.StringVar(&params.Format, "format", "", "format output : special variables : {stdin} = stdin value, {row} = row number from 1 to n")
	//
	//flag.StringVar(&params.Env, "env", "local", "environment to use")
	//flag.StringVar(&params.SetEnv, "set-env", "", "environment to save with provided settings")
	//
	//flag.Parse()
	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(nil))
		os.Exit(1)
	}
	return params
}
