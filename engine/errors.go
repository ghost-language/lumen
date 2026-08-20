package engine

import (
	"fmt"
	"os"

	"ghostlang.org/x/ghost/object"
)

// A runtime error inside update() or draw() repeats every frame, and sixty
// copies a second of the same message buries whatever came before it. Errors are
// printed once and then counted, so the console stays readable while still
// making clear the problem has not gone away.
type errorReporter struct {
	lastMessage string
	repeats     uint64
}

// reportError prints a runtime error raised inside a Ghost callback. Errors in
// update() and draw() are reported rather than fatal, so one bad frame does not
// take the whole game down.
func (engine *Engine) reportError(name string, err *object.Error) {
	message := fmt.Sprintf("lumen: error in %s(): %s", name, err.Message)

	if message == engine.errors.lastMessage {
		engine.errors.repeats++

		return
	}

	engine.flushRepeats()

	fmt.Fprintln(os.Stderr, message)

	engine.errors.lastMessage = message
	engine.errors.repeats = 0
}

// flushRepeats reports how many times the previous error recurred before a
// different one replaced it.
func (engine *Engine) flushRepeats() {
	if engine.errors.repeats > 0 {
		fmt.Fprintf(os.Stderr, "lumen: (repeated %d more times)\n", engine.errors.repeats)
	}

	engine.errors.repeats = 0
}
