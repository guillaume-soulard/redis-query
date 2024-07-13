package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func loadEnv(params *Parameters, envName string) {
	if home, err := os.UserHomeDir(); err != nil {
		PrintErrorAndExit(err)
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, envName)); err != nil {
			return
		} else {
			var loadedParams ConnectParameters
			if err = json.Unmarshal(file, &loadedParams); err != nil {
				PrintErrorAndExit(err)
			}
			var connect ConnectParameters
			setIfNotDefault(&connect.Host, loadedParams.Host)
			setIfNotDefault(&connect.Db, loadedParams.Db)
			setIfNotDefault(&connect.Port, loadedParams.Port)
			setIfNotDefault(&connect.SentinelMaster, loadedParams.SentinelMaster)
			setIfNotDefault(&connect.User, loadedParams.User)
			setIfNotDefault(&connect.Password, loadedParams.Password)
			setIfNotDefault(&connect.SentinelAddrs, loadedParams.SentinelAddrs)
			setIfNotDefault(&connect.SentinelUser, loadedParams.SentinelUser)
			setIfNotDefault(&connect.SentinelPassword, loadedParams.SentinelPassword)

			params.Command.Connect = connect
			params.Scan.Connect = connect
		}
	}
}

func delEnv(params Parameters) {
	if home, err := os.UserHomeDir(); err != nil {
		PrintErrorAndExit(err)
	} else {
		if err = os.Remove(fmt.Sprintf("%s/.redis-query/%s.json", home, *params.DelEnv.Name)); err != nil {
			PrintErrorAndExit(err)
		}
	}
}

func setIfNotDefault[T comparable](param *T, loadedParameter T) {
	if loadedParameter != *new(T) {
		*param = loadedParameter
	}
}

func saveEnv(params Parameters) {
	file, _ := json.MarshalIndent(params.SetEnv.ConnectParameters, "", " ")
	if home, err := os.UserHomeDir(); err != nil {
		PrintErrorAndExit(err)
	} else {
		if err = os.MkdirAll(fmt.Sprintf("%s/.redis-query", home), 0777); err != nil {
			PrintErrorAndExit(err)
		}
		if err = os.WriteFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *params.SetEnv.Name), file, 0777); err != nil {
			PrintErrorAndExit(err)
		}
	}
}

func listEnv() {
	if home, err := os.UserHomeDir(); err != nil {
		PrintErrorAndExit(err)
	} else {
		var dirEntries []os.DirEntry
		if dirEntries, err = os.ReadDir(fmt.Sprintf("%s/.redis-query", home)); err != nil {
			PrintErrorAndExit(err)
		}
		for _, entry := range dirEntries {
			if strings.Contains(entry.Name(), ".json") {
				Print(strings.ReplaceAll(entry.Name(), ".json", ""))
			}
		}
	}
}

func describeEnv(params Parameters) {
	if home, err := os.UserHomeDir(); err != nil {
		PrintErrorAndExit(err)
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *params.DescribeEnv.Name)); err != nil {
			return
		} else {
			Print(string(file))
		}
	}
}
