package persistence

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/lawalbolaji/gmdb/parser"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mtx  sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// Start a goroutine to sync AOF to disk every 1 second
	go func() {
		for {
			aof.mtx.Lock()

			aof.file.Sync()

			aof.mtx.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value parser.Value) error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(fn func(value parser.Value)) error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()

	aof.file.Seek(0, io.SeekStart)

	reader := parser.NewResp(aof.file)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fn(value)
	}

	return nil
}
