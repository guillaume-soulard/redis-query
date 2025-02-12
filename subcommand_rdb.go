package main

import (
	"fmt"
	"os"
	"strconv"
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

func (q *RdbConsumer) Next(amount int) (b []byte, eof bool, err error) {
	b = make([]byte, amount)
	for i := 0; i < amount; i++ {
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
		b[i] = q.buffer[q.bufferIndex]
		q.bufferIndex++
	}
	return b, eof, err
}

func parseRdb(file os.File) (err error) {
	consumer := NewRdbConsumer(file)
	var eof bool
	var b []byte
	if b, eof, err = consumer.Next(5); err != nil || eof {
		return err
	}
	fmt.Println(string(b))
	if b, eof, err = consumer.Next(4); err != nil || eof {
		return err
	}
	fmt.Println(strconv.Atoi(string(b)))
	return err
}
