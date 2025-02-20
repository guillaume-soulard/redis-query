package main

import (
	"errors"
	"github.com/hdt3213/rdb/encoder"
	rdbParser "github.com/hdt3213/rdb/parser"
	"os"
)

type RdbSubCommandUpdateTtl struct {
	encoder *encoder.Encoder
}

func (q RdbSubCommandUpdateTtl) Accept(command RdbCommand) bool {
	return command.Update.Cmd.Happened()
}

func (q RdbSubCommandUpdateTtl) Begin(command RdbCommand) (err error) {
	var outputFile *os.File
	if _, err = os.Stat(*command.OutputFile); err == nil {
		if outputFile, err = os.Open(*command.OutputFile); err != nil {
			return err
		}
		if err = os.Truncate(*command.OutputFile, 0); err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if outputFile, err = os.Create(*command.OutputFile); err != nil {
			return err
		}
	} else {
		return err
	}
	q.encoder = encoder.NewEncoder(outputFile)
	err = q.encoder.WriteHeader()
	return err
}

func (q RdbSubCommandUpdateTtl) Execute(object rdbParser.RedisObject, command RdbCommand) (err error) {
	switch object.GetType() {
	case rdbParser.DBSizeType:
		dbSize := object.(*rdbParser.DBSizeObject)
		println(dbSize.Key, dbSize.Size)
	case rdbParser.AuxType:
		aux := object.(*rdbParser.AuxObject)
		err = q.encoder.WriteAux(aux.Key, aux.Value)
		println(aux.Key, aux.Value)
	case rdbParser.StringType:
		str := object.(*rdbParser.StringObject)
		err = q.encoder.WriteStringObject(str.Key, str.Value)
		println(str.Key, str.Value)
	case rdbParser.ListType:
		list := object.(*rdbParser.ListObject)
		err = q.encoder.WriteListObject(list.Key, list.Values)
		println(list.Key, list.Values)
	case rdbParser.HashType:
		hash := object.(*rdbParser.HashObject)
		err = q.encoder.WriteHashMapObject(hash.Key, hash.Hash)
		println(hash.Key, hash.Hash)
	case rdbParser.ZSetType:
		zSet := object.(*rdbParser.ZSetObject)
		err = q.encoder.WriteZSetObject(zSet.Key, zSet.Entries)
		println(zSet.Key, zSet.Entries)
	case rdbParser.SetType:
		set := object.(*rdbParser.SetObject)
		err = q.encoder.WriteSetObject(set.Key, set.Members)
		println(set.Key, set.Members)
	case rdbParser.StreamType:
		stream := object.(*rdbParser.StreamObject)
		err = errors.New("unsupported stream type")
		println(stream.Entries, stream.Groups)
	}
	return err
}

func (q RdbSubCommandUpdateTtl) End(_ RdbCommand) (err error) {
	err = q.encoder.WriteEnd()
	return err
}
