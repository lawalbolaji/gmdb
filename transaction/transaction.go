package transaction

import (
	"strings"

	"github.com/lawalbolaji/gmdb/commands"
	"github.com/lawalbolaji/gmdb/parser"
	"github.com/lawalbolaji/gmdb/store"
)

type commandQueue struct {
	commands [][]parser.Value
}

func (q *commandQueue) enqueue(commandWithArgs []parser.Value) {
	q.commands = append(q.commands, commandWithArgs)
}

func (q *commandQueue) dequeue() ([]parser.Value, bool) {
	if len(q.commands) == 0 {
		return []parser.Value{}, false
	}

	command := q.commands[0]
	q.commands = q.commands[1:]
	return command, true
}

func (q *commandQueue) size() int {
	return len(q.commands)
}

func (q *commandQueue) flush() {
	q.commands = [][]parser.Value{}
}

func NewCommandQueue() *commandQueue {
	return &commandQueue{}
}

func queueCommand(q *commandQueue, commandWithArgs []parser.Value) parser.Value {
	q.enqueue(commandWithArgs)

	return parser.Value{Typ: parser.BULK_STRING, Bulk: "<<Queued>>"}
}

// when exec command is received
func execCommandsInQueue(q *commandQueue) parser.Value {
	result := parser.Value{Typ: parser.ARRAY}
	for _, lock := range store.GetRequiredLocks(q.commands) {
		lock.Lock()
		defer lock.Unlock()
	}

	if q.size() == 0 {
		return parser.Value{Typ: parser.BULK_STRING, Bulk: "no changes to db"}
	}

	for range q.size() {
		commandWithArg, _ := q.dequeue() // this check is redundant since the for loop only runs for the length of the queue
		command := strings.ToUpper(commandWithArg[0].Bulk)
		args := commandWithArg[1:]

		handler, ok := commands.Handlers[command]
		if !ok {
			continue
		}

		result.Array = append(result.Array, handler(args))
	}

	exitTransaction(q, "")
	return result
}

func exitTransaction(q *commandQueue, msg string) parser.Value {
	q.flush()
	return parser.Value{Typ: parser.BULK_STRING, Bulk: msg}
}

func HandleCommandInTransactionMode(ast parser.Value, queue *commandQueue, command string) (bool, parser.Value) {
	switch command {
	case "WATCH":
		exitTransaction(queue, "")
		return false, parser.Value{Typ: parser.SIMPLE_STRING, Str: "ERR not supported yet, exiting transaction mode"}
	case "MULTI":
		exitTransaction(queue, "")
		return false, parser.Value{Typ: parser.SIMPLE_STRING, Str: "ERR multi commands cannot be nested, exiting transaction mode"}
	case "DISCARD":
		return false, exitTransaction(queue, "exiting transaction mode, no changes made to db")
	case "EXEC":
		return false, execCommandsInQueue(queue)
	default:
		return true, queueCommand(queue, ast.Array)
	}
}
