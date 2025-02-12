package main

import (
	"fmt"
	"os"
)

type RdbSubCommand struct{}

func (q RdbSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Rdb.Cmd.Happened()
}

func (q RdbSubCommand) Execute(parameters *Parameters) (err error) {
	if parameters.Rdb.InputFiles != nil {
		for _, file := range *parameters.Rdb.InputFiles {
			if err = parseRdb(file); err != nil {
				return err
			}
		}
	}
	return err
}

type RdbConsumer struct {
	file        os.File
	buffer      []byte
	bufferIndex int
	maxToRead   int
}

func NewRdbConsumer(file os.File) *RdbConsumer {
	return &RdbConsumer{
		file:        file,
		buffer:      make([]byte, 4096),
		bufferIndex: -1,
		maxToRead:   0,
	}
}

func (q *RdbConsumer) Next() (b byte, eof bool, err error) {
	if q.bufferIndex >= len(q.buffer) || q.bufferIndex >= q.maxToRead || q.bufferIndex < 0 {
		if q.maxToRead, err = q.file.Read(q.buffer); err != nil {
			return b, eof, err
		}
		if q.maxToRead == 0 {
			eof = true
			return b, eof, err
		}
		q.bufferIndex = 0
	}
	b = q.buffer[q.bufferIndex]
	q.bufferIndex++
	return b, eof, err
}

func parseRdb(file os.File) (err error) {
	consumer := NewRdbConsumer(file)
	var eof bool
	var b byte
	for b, eof, err = consumer.Next(); err == nil && !eof; b, eof, err = consumer.Next() {
		fmt.Println(string(b))
	}
	return err
}
