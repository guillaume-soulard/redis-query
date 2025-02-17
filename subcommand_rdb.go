package main

import (
	"errors"
	"fmt"
	"github.com/pektezol/bitreader"
	"math"
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
	reader      *bitreader.Reader
	buffer      []byte
	bufferIndex int
	maxToRead   int
}

func NewRdbConsumer(file os.File) *RdbConsumer {
	return &RdbConsumer{
		reader:      bitreader.NewReader(&file, true),
		buffer:      make([]byte, 4096),
		bufferIndex: -1,
		maxToRead:   0,
	}
}

func (q *RdbConsumer) NextBytes(amount uint64) (b []byte, err error) {
	b, err = q.reader.ReadBytesToSlice(amount)
	return b, err
}

func (q *RdbConsumer) NextBits(amount uint64) (b []byte, err error) {
	var raw []byte
	if raw, err = q.reader.ReadBitsToSlice(amount); err != nil {
		return b, err
	}
	b = make([]byte, len(raw)*8)
	for i, byteItem := range raw {
		for j := 0; j < 8; j++ {
			b[i*8+j] = byteItem & (1 << j) >> j
		}
	}
	b = b[:amount]
	return b, err
}

var byteField = map[byte]func(*RdbConsumer) error{
	'\372': auxFieldConsumer,
}

func parseRdb(file os.File) (err error) {
	consumer := NewRdbConsumer(file)
	var b []byte
	if b, err = consumer.NextBytes(5); err != nil {
		return err
	}
	fmt.Println(string(b))
	if b, err = consumer.NextBytes(4); err != nil {
		return err
	}
	var rdbVersion int
	if rdbVersion, err = strconv.Atoi(string(b)); err != nil {
		return err
	}
	fmt.Println(rdbVersion)
	for {
		if b, err = consumer.NextBytes(1); err != nil {
			return err
		}
		if err = byteField[b[0]](consumer); err != nil {
			return err
		}
		if b[0] == '\377' {
			break
		}
	}
	return err
}

func auxFieldConsumer(consumer *RdbConsumer) (err error) {
	var key, value string
	if key, err = stringConsumer(consumer); err != nil {
		return err
	}
	if value, err = stringConsumer(consumer); err != nil {
		return err
	}
	_ = key
	_ = value
	return err
}

func stringConsumer(consumer *RdbConsumer) (str string, err error) {
	var length int64
	if length, err = lengthConsumer(consumer); err != nil {
		return str, err
	}
	_ = length
	return str, err
}

func lengthConsumer(consumer *RdbConsumer) (length int64, err error) {
	var bits []byte
	if bits, err = consumer.NextBits(2); err != nil {
		return length, err
	}
	if bits[0] == 0 && bits[1] == 0 {
		if bits, err = consumer.NextBits(6); err != nil {
			return length, err
		}
		length = bitsToInteger(bits)
	} else if bits[0] == 0 && bits[1] == 1 {
		if bits, err = consumer.NextBits(14); err != nil {
			return length, err
		}
		length = bitsToInteger(bits)
	} else if bits[0] == 1 && bits[1] == 0 {
		if _, err = consumer.NextBits(6); err != nil {
			return length, err
		}
		if bits, err = consumer.NextBits(4 * 8); err != nil {
			return length, err
		}
		length = bitsToInteger(bits)
	} else if bits[0] == 1 && bits[1] == 1 {
		var format []byte
		if format, err = consumer.NextBits(6); err != nil {
			return length, err
		}
		if bits, err = consumer.NextBytes(4); err != nil {
			return length, err
		}
		_ = format
		err = errors.New("unimplemented length encoding format")
	}
	return length, err
}

func bitsToInteger(bits []byte) (result int64) {
	result = 0
	pow := 0
	for i := len(bits) - 1; i >= 0; i-- {
		result += int64(bits[i]) * int64(math.Pow(float64(2), float64(pow)))
		pow++
	}
	return result
}

func bitsToByte(bits []byte) (result []byte) {
	result = make([]byte, len(bits)/8)
	const bitMask = 1
	bitCounter := 0
	for b := 0; b < len(bits)/8; b++ {
		for bit := 0; bit < 8; bit++ {
			result[b] |= (bits[bitCounter] & bitMask) << (7 - bit)
			bitCounter++
		}
	}
	return result
}
