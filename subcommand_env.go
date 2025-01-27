package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SetEnvSubCommand struct{}

func (e SetEnvSubCommand) Accept(parameters *Parameters) bool {
	return parameters.SetEnv.Cmd.Happened()
}

func (e SetEnvSubCommand) Execute(parameters *Parameters) (err error) {
	file, _ := json.MarshalIndent(parameters.SetEnv.ConnectParameters, "", " ")
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	} else {
		if err = os.MkdirAll(fmt.Sprintf("%s/.redis-query", home), 0777); err != nil {
			return err
		}
		if err = os.WriteFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *parameters.SetEnv.Name), file, 0777); err != nil {
			return err
		}
	}
	return err
}

type LoadEnvSubCommand struct{}

func (l LoadEnvSubCommand) Accept(parameters *Parameters) bool {
	return (parameters.Scan.Cmd.Happened() && *parameters.Scan.EnvName != "") ||
		(parameters.Command.Cmd.Happened() && *parameters.Command.EnvName != "") ||
		(parameters.Query.Cmd.Happened() && *parameters.Query.EnvName != "") ||
		(parameters.Connect.Cmd.Happened() && *parameters.Connect.EnvName != "") ||
		(parameters.Migrate.Cmd.Happened() && *parameters.Migrate.SourceEnv != "" && *parameters.Migrate.TargetEnv != "")
}

func (l LoadEnvSubCommand) Execute(parameters *Parameters) (err error) {
	if parameters.Scan.Cmd.Happened() && *parameters.Scan.EnvName != "" {
		err = loadEnv(parameters.Scan.EnvName, &parameters.Scan.Connect)
	}
	if parameters.Command.Cmd.Happened() && *parameters.Command.EnvName != "" {
		err = loadEnv(parameters.Command.EnvName, &parameters.Command.Connect)
	}
	if parameters.Query.Cmd.Happened() && *parameters.Query.EnvName != "" {
		err = loadEnv(parameters.Query.EnvName, &parameters.Query.Connect)
	}
	if parameters.Connect.Cmd.Happened() && *parameters.Connect.EnvName != "" {
		err = loadEnv(parameters.Connect.EnvName, &parameters.Connect.Connect)
	}
	if parameters.Migrate.Cmd.Happened() && *parameters.Migrate.SourceEnv != "" && *parameters.Migrate.TargetEnv != "" {
		if err = loadEnv(parameters.Migrate.SourceEnv, &parameters.Migrate.SourceConnect); err != nil {
			return err
		}
		err = loadEnv(parameters.Migrate.TargetEnv, &parameters.Migrate.TargetConnect)
	}
	return err
}

func loadEnv(envName *string, connectParameters *ConnectParameters) (err error) {
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *envName)); err != nil {
			return
		} else {
			var loadedParams ConnectParameters
			if err = json.Unmarshal(file, &loadedParams); err != nil {
				return err
			}
			var connect ConnectParameters
			setIfNotDefault(&connect.Host, loadedParams.Host)
			setIfNotDefault(&connect.Db, loadedParams.Db)
			setIfNotDefault(&connect.Port, loadedParams.Port)
			setIfNotDefault(&connect.SentinelMaster, loadedParams.SentinelMaster)
			setIfNotDefault(&connect.User, loadedParams.User)
			setIfNotDefault(&connect.Password, loadedParams.Password)
			setIfNotDefault(&connect.SentinelAddrs, loadedParams.SentinelAddrs)

			*connectParameters = connect
		}
	}
	return err
}

type DelEnvSubCommand struct{}

func (d DelEnvSubCommand) Accept(parameters *Parameters) bool {
	return parameters.DelEnv.Cmd.Happened()
}

func (d DelEnvSubCommand) Execute(parameters *Parameters) (err error) {
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	} else {
		if err = os.Remove(fmt.Sprintf("%s/.redis-query/%s.json", home, *parameters.DelEnv.Name)); err != nil {
			return err
		}
	}
	return err
}

func setIfNotDefault[T comparable](param *T, loadedParameter T) {
	if loadedParameter != *new(T) {
		*param = loadedParameter
	}
}

type ListEnvSubCommand struct{}

func (l ListEnvSubCommand) Accept(parameters *Parameters) bool {
	return parameters.ListEnv.Cmd.Happened()
}

func (l ListEnvSubCommand) Execute(parameters *Parameters) (err error) {
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	} else {
		var dirEntries []os.DirEntry
		if dirEntries, err = os.ReadDir(fmt.Sprintf("%s/.redis-query", home)); err != nil {
			return err
		}
		for _, entry := range dirEntries {
			if strings.Contains(entry.Name(), ".json") {
				Print(strings.ReplaceAll(entry.Name(), ".json", ""))
			}
		}
	}
	return err
}

type DescribeEnvSubCommand struct{}

func (d DescribeEnvSubCommand) Accept(parameters *Parameters) bool {
	return parameters.DescribeEnv.Cmd.Happened()
}

func (d DescribeEnvSubCommand) Execute(parameters *Parameters) (err error) {
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *parameters.DescribeEnv.Name)); err != nil {
			return err
		} else {
			Print(string(file))
		}
	}
	return err
}
