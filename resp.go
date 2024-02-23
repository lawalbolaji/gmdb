package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	typ   rune
	str   string
	num   int
	bulk  string
	array []Value
}

/*
	{
		typ: array,
		array: [
			{
				typ: bulk,
				bulk: ""
			},
			{
				typ: integer,
				num: x
			}
		]
	}

*/

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (resp *Resp) ReadInteger() (int, error) {
	bSeq, err := resp.reader.ReadBytes('\r')
	if err != nil {
		if err == io.EOF {
			return -1, io.EOF
		}

		log.Fatal(err)
	}

	return strconv.Atoi(string(bSeq[:len(bSeq)-1]))
}

/*

	INITIAL COMMAND:
	*2\r\n$7\r\nCOMMAND\r\n$4\r\nDOCS\r\n

	SET name rasheed
	*3\r\n$3\r\nset\r\n$4\r\nname\r\n$7\r\nrasheed\r\n
*/

func (resp *Resp) Read() (Value, error) {
	_type, err := resp.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return resp.readArray()
	case BULK:
		return resp.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
func (resp *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = ARRAY

	len, err := resp.ReadInteger()
	if err != nil {
		log.Fatal(err)
	}

	// read clrf
	resp.reader.ReadBytes('\n')

	v.array = make([]Value, 0)
	for idx := 0; idx < len; idx++ {
		val, err := resp.Read()
		if err != nil {
			return v, err
		}

		v.array = append(v.array, val)
	}

	return v, nil
}

func (resp *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = BULK

	len, err := resp.ReadInteger()
	if err != nil {
		log.Fatal(err)
	}

	// read clrf
	resp.reader.ReadBytes('\n')

	bulk := make([]byte, len)
	resp.reader.Read(bulk)
	v.bulk = string(bulk)

	// read clrf
	resp.reader.ReadBytes('\n')

	return v, nil
}
