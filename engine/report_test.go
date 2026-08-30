package engine

import (
	"bytes"
	"strings"
	"testing"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

// reporting is an engine with nowhere to draw and somewhere to write, which is
// everything a report needs that is not the window.
func reporting() (*Engine, *bytes.Buffer) {
	engine := &Engine{Graphics: NewGraphics()}
	written := &bytes.Buffer{}

	engine.SetReportWriter(written)

	return engine, written
}

func raise(line int) *fault.Fault {
	return fault.At(fault.Type, token.Token{File: "main.gs", Line: line, Column: 1, Length: 3, Lexeme: "sum"},
		"cannot use `+` between number and string")
}

func TestAFailureStopsTheGameAndIsWrittenOut(t *testing.T) {
	engine, written := reporting()

	engine.Raise("draw()", true, raise(4))

	if !engine.Halted() {
		t.Fatal("expected the game to stop on the failure")
	}

	if !strings.Contains(written.String(), "cannot use `+` between number and string") {
		t.Errorf("got %q, want the failure written to the console", written.String())
	}

	if !strings.Contains(written.String(), "main.gs:4:1") {
		t.Errorf("got %q, want the position in the report", written.String())
	}
}

func TestTheCallbackIsRecordedAsTheOutermostFrame(t *testing.T) {
	engine, written := reporting()

	engine.Raise("draw()", true, raise(4))

	if !strings.Contains(written.String(), "in draw()") {
		t.Errorf("got %q, want the callback named in the report", written.String())
	}
}

func TestARepeatedFailureIsCountedRatherThanReprinted(t *testing.T) {
	engine, written := reporting()

	engine.Raise("update()", true, raise(4))
	engine.Resume()

	for count := 0; count < 5; count++ {
		engine.Raise("update()", true, raise(4))
	}

	engine.flushRepeats()

	if got := strings.Count(written.String(), "cannot use `+`"); got != 1 {
		t.Errorf("wrote the same failure %d times, want once", got)
	}

	if !strings.Contains(written.String(), "repeated 5 more times") {
		t.Errorf("got %q, want the repeats counted", written.String())
	}
}

func TestCarryingOnPastAFailureDoesNotStopForItAgain(t *testing.T) {
	engine, _ := reporting()

	engine.Raise("update()", true, raise(4))
	engine.Resume()

	if engine.Halted() {
		t.Fatal("expected the game to carry on")
	}

	engine.Raise("update()", true, raise(4))

	if engine.Halted() {
		t.Error("stopped again for a failure that was waved through")
	}

	// A different failure is a different decision, and stops the game.
	engine.Raise("update()", true, raise(9))

	if !engine.Halted() {
		t.Error("expected a new failure to stop the game")
	}
}

func TestAFailureThatCannotBeCarriedOnFromIsNotResumable(t *testing.T) {
	engine, _ := reporting()

	engine.Raise("load()", false, raise(2))
	engine.Resume()

	if !engine.Halted() {
		t.Error("expected load() to stay stopped: there is nothing to carry on with")
	}

	if !engine.Failed() {
		t.Error("expected the run to be marked as failed")
	}
}

func TestTheFirstFailureIsTheOneOnScreen(t *testing.T) {
	engine, _ := reporting()

	engine.Raise("update()", true, raise(4))
	engine.Raise("draw()", true, raise(9))

	if engine.report.Fault.Position.Line != 4 {
		t.Errorf("got the failure at line %d, want the one the game stopped on", engine.report.Fault.Position.Line)
	}
}

func TestShowDoesNotReprintWhatGhostAlreadyWrote(t *testing.T) {
	engine, written := reporting()

	engine.Show("the game's source", raise(1))

	if written.Len() != 0 {
		t.Errorf("got %q, want nothing: the console already has it", written.String())
	}

	if !engine.Halted() || !engine.Failed() {
		t.Error("expected the window to show it and the run to have failed")
	}
}

func TestAPanicBecomesAReportAboutLumen(t *testing.T) {
	raised := internalFault("draw()", "runtime error: index out of range")

	if raised.Kind != fault.Internal {
		t.Errorf("got kind %s, want an internal error", raised.Kind)
	}

	if !strings.Contains(raised.Message, "draw()") {
		t.Errorf("got %q, want the callback named", raised.Message)
	}

	if !strings.Contains(raised.Help, "bug in Lumen") {
		t.Errorf("got help %q, want it to say whose bug this is", raised.Help)
	}
}

func TestAPanicComesBackAsAnErrorObject(t *testing.T) {
	// The panic here is induced rather than realistic — calling into an
	// interpreter that is not there. What is being tested is the guard around
	// game code, which is what stands between a bug in Lumen and a Go traceback
	// landing in front of someone who was writing a game.
	engine, _ := reporting()

	result := engine.call("draw", nil)

	raised, ok := result.(*object.Error)

	if !ok {
		t.Fatalf("got %T, want an error object", result)
	}

	if raised.Fault.Kind != fault.Internal {
		t.Errorf("got kind %s, want an internal error", raised.Fault.Kind)
	}
}
