package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadEnv(params *parameters) {
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		var file []byte
		if file, err = os.ReadFile(fmt.Sprintf("%s/.redis-query/%s.json", home, params.Env)); err != nil {
			return
		} else {
			var loadedParams parameters
			if err = json.Unmarshal(file, &loadedParams); err != nil {
				panic(err)
			}
			setIfNotDefault(&params.Host, loadedParams.Host)
			setIfNotDefault(&params.Db, loadedParams.Db)
			setIfNotDefault(&params.Port, loadedParams.Port)
			setIfNotDefault(&params.Scan, loadedParams.Scan)
			setIfNotDefault(&params.ScanCount, loadedParams.ScanCount)
			setIfNotDefault(&params.SentinelMaster, loadedParams.SentinelMaster)
			setIfNotDefault(&params.User, loadedParams.User)
			setIfNotDefault(&params.Password, loadedParams.Password)
			setIfNotDefault(&params.SentinelAddrs, loadedParams.SentinelAddrs)
			setIfNotDefault(&params.SentinelUser, loadedParams.SentinelUser)
			setIfNotDefault(&params.SentinelPassword, loadedParams.SentinelPassword)
		}
	}
}

func setIfNotDefault[T comparable](param *T, loadedParameter T) {
	if loadedParameter != *new(T) {
		*param = loadedParameter
	}
}

func saveEnv(params parameters) {
	file, _ := json.MarshalIndent(params, "", " ")
	if home, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		if err = os.MkdirAll(fmt.Sprintf("%s/.redis-query", home), 0777); err != nil {
			panic(err)
		}
		if err = os.WriteFile(fmt.Sprintf("%s/.redis-query/%s.json", home, *params.SetEnv), file, 0777); err != nil {
			panic(err)
		}
	}
}
