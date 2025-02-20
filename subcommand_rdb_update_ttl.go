package main

// https://pkg.go.dev/github.com/yaofaqian/rdb#Encoder.EncodeExpiry
import (
	"errors"
	"github.com/hdt3213/rdb/encoder"
	rdbParser "github.com/hdt3213/rdb/parser"
	"os"
)

type RdbSubCommandUpdateTtl struct {
	encoder *encoder.Encoder
}

func (q *RdbSubCommandUpdateTtl) Accept(command RdbCommand) bool {
	return command.Update.Cmd.Happened()
}

func (q *RdbSubCommandUpdateTtl) Begin(command RdbCommand) (err error) {
	var outputFile *os.File
	if _, err = os.Stat(*command.OutputFile); err == nil {
		if err = os.Remove(*command.OutputFile); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if outputFile, err = os.Create(*command.OutputFile); err != nil {
		return err
	}
	q.encoder = encoder.NewEncoder(outputFile)
	if err = q.encoder.WriteHeader(); err != nil {
		return err
	}
	auxMap := map[string]string{
		"redis-ver":  "7.4.2",
		"redis-bits": "64",
	}
	for k, v := range auxMap {
		if err = q.encoder.WriteAux(k, v); err != nil {
			return err
		}
	}
	err = q.encoder.WriteDBHeader(0, 5, 0)
	return err
}

func (q *RdbSubCommandUpdateTtl) Execute(object rdbParser.RedisObject, command RdbCommand) (err error) {
	ttl := encoder.WithTTL(uint64(*command.Update.Ttl))
	switch object.GetType() {
	case rdbParser.DBSizeType:
		dbSize := object.(*rdbParser.DBSizeObject)
		err = errors.New("unsupported dbSize type")
		println(dbSize.Key, dbSize.Size)
	case rdbParser.AuxType:
		aux := object.(*rdbParser.AuxObject)
		err = q.encoder.WriteAux(aux.Key, aux.Value)
		println(aux.Key, aux.Value)
	case rdbParser.StringType:
		str := object.(*rdbParser.StringObject)
		err = q.encoder.WriteStringObject(str.Key, str.Value, ttl)
		println(str.Key, str.Value)
	case rdbParser.ListType:
		list := object.(*rdbParser.ListObject)
		err = q.encoder.WriteListObject(list.Key, list.Values, ttl)
		println(list.Key, list.Values)
	case rdbParser.HashType:
		hash := object.(*rdbParser.HashObject)
		err = q.encoder.WriteHashMapObject(hash.Key, hash.Hash, ttl)
		println(hash.Key, hash.Hash)
	case rdbParser.ZSetType:
		zSet := object.(*rdbParser.ZSetObject)
		err = q.encoder.WriteZSetObject(zSet.Key, zSet.Entries, ttl)
		println(zSet.Key, zSet.Entries)
	case rdbParser.SetType:
		set := object.(*rdbParser.SetObject)
		err = q.encoder.WriteSetObject(set.Key, set.Members, ttl)
		println(set.Key, set.Members)
	case rdbParser.StreamType:
		stream := object.(*rdbParser.StreamObject)
		err = errors.New("unsupported stream type")
		println(stream.Entries, stream.Groups)
	default:
		err = errors.New("unsupported object type")
	}
	return err
}

func (q *RdbSubCommandUpdateTtl) End(_ RdbCommand) (err error) {
	err = q.encoder.WriteEnd()
	return err
}
