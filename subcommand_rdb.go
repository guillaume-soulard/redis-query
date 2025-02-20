package main

import (
	rdbParser "github.com/hdt3213/rdb/parser"
	"os"
)

type RdbSubCommand struct{}

func (q RdbSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Rdb.Cmd.Happened()
}

func (q RdbSubCommand) Execute(parameters *Parameters) (err error) {
	if parameters.Rdb.InputFiles != nil {
		for _, file := range *parameters.Rdb.InputFiles {
			if err = parseRdb(&file); err != nil {
				return err
			}
		}
	}
	return err
}

// https://github.com/HDT3213/rdb

func parseRdb(file *os.File) (err error) {
	decoder := rdbParser.NewDecoder(file)
	err = decoder.Parse(func(o rdbParser.RedisObject) bool {
		switch o.GetType() {
		case rdbParser.DBSizeType:
			dbSize := o.(*rdbParser.DBSizeObject)
			println(dbSize.Key, dbSize.Size)
		case rdbParser.AuxType:
			aux := o.(*rdbParser.AuxObject)
			println(aux.Key, aux.Value)
		case rdbParser.StringType:
			str := o.(*rdbParser.StringObject)
			println(str.Key, str.Value)
		case rdbParser.ListType:
			list := o.(*rdbParser.ListObject)
			println(list.Key, list.Values)
		case rdbParser.HashType:
			hash := o.(*rdbParser.HashObject)
			println(hash.Key, hash.Hash)
		case rdbParser.ZSetType:
			zSet := o.(*rdbParser.ZSetObject)
			println(zSet.Key, zSet.Entries)
		case rdbParser.SetType:
			set := o.(*rdbParser.SetObject)
			println(set.Key, set.Members)
		case rdbParser.StreamType:
			stream := o.(*rdbParser.StreamObject)
			println(stream.Entries, stream.Groups)
		}
		return true
	})
	return err
}
