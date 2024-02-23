package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	_simpleString = '+'
	_bulkString   = '$'
	_simpleError  = '-'
	_integer      = ':'
	_array        = '*'
	_null         = '_'
	_clrf         = "\r\n"
)

type Value struct {
	typ    rune
	str    string
	bulk   string
	array  []Value
	Expose string
}

func Respond(conn net.Conn) {
	writer := NewWriter(conn)
	writer.Write(Value{typ: '+', str: "Ok"})

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
	case _array:
		return resp.readArray()
	case _bulkString:
		return resp.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
func (resp *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = _array

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
	v.typ = _bulkString

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

func (val Value) Marshal() []byte {
	switch val.typ {
	case _simpleError:
		return val.marshallError()
	case _null:
		return val.marshallNull()
	case _simpleString:
		return val.marshalString()
	case _bulkString:
		return val.marshalBulkString()
	case _array:
		return val.marshalArray()
	default:
		return []byte{}
	}
}

func (val Value) marshalString() []byte {
	var bytes []byte

	bytes = append(bytes, _simpleString)
	bytes = append(bytes, val.str...)
	bytes = append(bytes, _clrf...)

	return bytes
}

func (val Value) marshalBulkString() []byte {
	var bytes []byte

	bytes = append(bytes, _bulkString)
	bytes = append(bytes, strconv.Itoa(len(val.bulk))...)
	bytes = append(bytes, _clrf...)
	bytes = append(bytes, val.bulk...)
	bytes = append(bytes, _clrf...)

	return bytes
}

func (val Value) marshalArray() []byte {
	len := len(val.array)
	var bytes []byte

	bytes = append(bytes, _array)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, _clrf...)

	for i := 0; i < len; i++ {
		bytes = append(bytes, val.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, _simpleError)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, _clrf...)

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}
