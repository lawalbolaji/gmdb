package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
)

const (
	SIMPLE_STRING = '+'
	BULK_STRING   = '$'
	SIMPLE_ERROR  = '-'
	INTEGER       = ':'
	ARRAY         = '*'
	NULL          = '_'
	CLRF          = "\r\n"
)

type Value struct {
	Typ   rune
	Str   string
	Bulk  string
	Array []Value
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

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	return err
}

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
	e.g input:

	$-> INITIAL COMMAND:
	*2\r\n$7\r\nCOMMAND\r\n$4\r\nDOCS\r\n

	$-> SET name rasheed
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
	case BULK_STRING:
		return resp.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
func (resp *Resp) readArray() (Value, error) {
	v := Value{}
	v.Typ = ARRAY

	len, err := resp.ReadInteger()
	if err != nil {
		log.Fatal(err)
	}

	// read clrf
	resp.reader.ReadBytes('\n')

	v.Array = make([]Value, 0)
	for idx := 0; idx < len; idx++ {
		val, err := resp.Read()
		if err != nil {
			return v, err
		}

		v.Array = append(v.Array, val)
	}

	return v, nil
}

func (resp *Resp) readBulk() (Value, error) {
	v := Value{}
	v.Typ = BULK_STRING

	len, err := resp.ReadInteger()
	if err != nil {
		log.Fatal(err)
	}

	// read clrf
	resp.reader.ReadBytes('\n')

	bulk := make([]byte, len)
	resp.reader.Read(bulk)
	v.Bulk = string(bulk)

	// read clrf
	resp.reader.ReadBytes('\n')

	return v, nil
}

func (val Value) Marshal() []byte {
	switch val.Typ {
	case SIMPLE_ERROR:
		return val.marshallError()
	case NULL:
		return val.marshallNull()
	case SIMPLE_STRING:
		return val.marshalString()
	case BULK_STRING:
		return val.marshalBulkString()
	case ARRAY:
		return val.marshalArray()
	default:
		return []byte{}
	}
}

func (val Value) marshalString() []byte {
	var bytes []byte

	bytes = append(bytes, SIMPLE_STRING)
	bytes = append(bytes, val.Str...)
	bytes = append(bytes, CLRF...)

	return bytes
}

func (val Value) marshalBulkString() []byte {
	var bytes []byte

	bytes = append(bytes, BULK_STRING)
	bytes = append(bytes, strconv.Itoa(len(val.Bulk))...)
	bytes = append(bytes, CLRF...)
	bytes = append(bytes, val.Bulk...)
	bytes = append(bytes, CLRF...)

	return bytes
}

func (val Value) marshalArray() []byte {
	len := len(val.Array)
	var bytes []byte

	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, CLRF...)

	for i := 0; i < len; i++ {
		bytes = append(bytes, val.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, SIMPLE_ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, CLRF...)

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}
