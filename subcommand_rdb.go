package main

import (
	"errors"
	rdbParser "github.com/hdt3213/rdb/parser"
	"os"
)

type RdbObjectHandler interface {
	Accept(command RdbCommand) bool
	Begin(command RdbCommand) error
	Execute(object rdbParser.RedisObject, command RdbCommand) error
	End(command RdbCommand) error
}

var objectHandlers = []RdbObjectHandler{
	RdbSubCommandUpdateTtl{},
}

type RdbSubCommand struct{}

func (q RdbSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Rdb.Cmd.Happened()
}

func (q RdbSubCommand) Execute(parameters *Parameters) (err error) {
	var objectHandler RdbObjectHandler
	for _, handler := range objectHandlers {
		if handler.Accept(parameters.Rdb) {
			objectHandler = handler
			break
		}
	}
	if objectHandler == nil {
		err = errors.New("no rdb handler found")
		return err
	}
	if err = objectHandler.Begin(parameters.Rdb); err != nil {
		return err
	}
	if parameters.Rdb.InputFiles != nil {
		for _, file := range *parameters.Rdb.InputFiles {
			var inputFile *os.File
			if inputFile, err = os.Open(file); err != nil {
				return err
			}
			if err = processRdb(inputFile, objectHandler, parameters.Rdb); err != nil {
				return err
			}
		}
	}
	if err = objectHandler.End(parameters.Rdb); err != nil {
		return err
	}
	return err
}

// https://github.com/HDT3213/rdb

func processRdb(file *os.File, objectHandler RdbObjectHandler, rdbCommand RdbCommand) (err error) {
	decoder := rdbParser.NewDecoder(file)
	if err = objectHandler.Begin(rdbCommand); err != nil {
		return err
	}
	if err = decoder.Parse(func(o rdbParser.RedisObject) bool {
		if err = objectHandler.Execute(o, rdbCommand); err != nil {
			PrintErrorAndExit(err)
			return false
		}
		return true
	}); err != nil {
		return err
	}
	if err = objectHandler.End(rdbCommand); err != nil {
		return err
	}
	return err
}
