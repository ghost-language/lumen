package engine

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"ghostlang.org/x/ghost/color"
	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"github.com/veandco/go-sdl2/sdl"
)

// A game is not a program someone is watching a terminal for.
//
// It runs in its own window, usually fullscreen, often on a machine that was
// never started from a shell — and when it goes wrong, the person looking at it
// sees a frozen frame, a black screen, or a sprite that stopped moving. The
// message explaining why is somewhere behind the window, in a console nobody
// opened. That is the problem this file exists to solve: every failure is
// reported twice, once to the console for whoever is developing the game, and
// once into the window for whoever is looking at it.
//
// Reporting is also the only thing that stops. Nothing below here prints, and
// nothing below here decides whether the game carries on: a failing method
// hands back an error object, the loop notices it here, and the game holds
// still on an error screen until the reader says what to do about it. A game
// that scribbles the same failure into the console sixty times a second, and
// carries on drawing with half its state missing, is a game whose real problem
// is three thousand lines further up the output.

// Report is a failure the game has stopped for.
type Report struct {
	// Fault is what went wrong, described rather than formatted. It carries the
	// position, the source it happened on, and the help, and it is what both
	// the console report and the on-screen one are rendered from.
	Fault *fault.Fault

	// Where names the callback the failure surfaced in — "draw()", "load()" —
	// or the phase, for a failure that happened before the loop started.
	Where string

	// Resumable says whether carrying on is worth offering. A game whose
	// update() failed once can usually be watched a little longer; a game whose
	// load() failed never built the state every later frame reads, so there is
	// nothing to carry on with.
	Resumable bool
}

// errorReporter holds what the console needs to stay readable and what the
// window needs to stay quiet about failures the reader has already waved
// through.
type errorReporter struct {
	// lastMessage and repeats collapse a failure that recurs. A game that has
	// been resumed past a bad frame raises the same error on the next one, and
	// sixty copies a second of it buries whatever came before.
	lastMessage string
	repeats     uint64

	// dismissed holds the failures the reader has chosen to carry on past.
	// They are still counted in the console; they no longer stop the game.
	dismissed map[string]bool

	// failed records that the run is not going to end well, so the process can
	// exit with something other than success even if the window is closed
	// tidily afterwards.
	failed bool

	// writer is where console reports go. It is standard error unless something
	// embedding Lumen says otherwise, which keeps a game's own output clean
	// when it is piped somewhere.
	writer io.Writer
}

// SetReportWriter chooses where console reports are written. The window is
// reported to either way: the error screen is not something an embedder can
// turn off, because it is the only report a player will ever see.
func (engine *Engine) SetReportWriter(writer io.Writer) {
	engine.errors.writer = writer
}

// reportWriter is where a console report goes.
func (engine *Engine) reportWriter() io.Writer {
	if engine.errors.writer == nil {
		return os.Stderr
	}

	return engine.errors.writer
}

// Raise reports a fault and stops the game on it.
//
// This is the door every failure leaves by. Whatever noticed the problem
// describes it; this decides that the console hears about it, that the window
// shows it, and that the loop stops calling game code until the reader has
// dealt with it.
func (engine *Engine) Raise(where string, resumable bool, raised *fault.Fault) {
	if raised == nil {
		return
	}

	engine.note(where, raised)
	engine.write(raised)
	engine.halt(Report{Fault: raised, Where: where, Resumable: resumable})
}

// Show stops the game on a fault that has already been written to the console.
//
// Ghost reports what it finds while scanning and parsing a game's source, and
// it reports all of it, not just the first one; re-printing that here would
// only say it twice. The window has not seen it, though, and the window is
// where a player — or a developer who opened the game by double-clicking it —
// is looking.
func (engine *Engine) Show(where string, raised *fault.Fault) {
	if raised == nil {
		return
	}

	engine.halt(Report{Fault: raised, Where: where})
}

// Halted reports whether the game is stopped on a failure. While it is, no game
// code runs: not update(), not draw(), and not the input callbacks either.
func (engine *Engine) Halted() bool {
	return engine.report != nil
}

// Failed reports whether the run hit something it could not carry on from,
// which is what the process exits non-zero for.
func (engine *Engine) Failed() bool {
	return engine.errors.failed
}

// Resume carries on past the failure on screen, and remembers not to stop for
// that same failure again. A game whose draw() fails on one particular enemy
// would otherwise stop on the very next frame, and the reader who asked to
// carry on would have to ask again sixty times a second.
func (engine *Engine) Resume() {
	if engine.report == nil || !engine.report.Resumable {
		return
	}

	if engine.errors.dismissed == nil {
		engine.errors.dismissed = make(map[string]bool)
	}

	engine.errors.dismissed[engine.report.Fault.String()] = true
	engine.report = nil
}

// Copy puts the report on the clipboard, which is the shortest path from a
// window that cannot be selected from to an issue, a chat message, or a search.
func (engine *Engine) Copy() bool {
	if engine.report == nil {
		return false
	}

	return sdl.SetClipboardText(engine.report.Fault.Render(color.Plain)) == nil
}

// note records which callback a fault came out of, as the outermost frame of
// its trace. It is the same "in draw()" line Ghost writes for a call inside the
// game's own code, because from the reader's side it is the same fact: this is
// where the failing code was called from.
func (engine *Engine) note(where string, raised *fault.Fault) {
	if where == "" {
		return
	}

	raised.Trace = append(raised.Trace, fault.Frame{Name: where})
}

// write reports a fault on the console, collapsing one that repeats.
func (engine *Engine) write(raised *fault.Fault) {
	message := raised.String()

	if message == engine.errors.lastMessage {
		engine.errors.repeats++

		return
	}

	engine.flushRepeats()

	writer := engine.reportWriter()

	fmt.Fprintln(writer, raised.Render(color.Detect(writer)))

	engine.errors.lastMessage = message
	engine.errors.repeats = 0
}

// flushRepeats reports how many times the previous failure recurred before a
// different one replaced it, or before the game shut down.
func (engine *Engine) flushRepeats() {
	if engine.errors.repeats > 0 {
		fmt.Fprintf(engine.reportWriter(), "lumen: (repeated %d more times)\n", engine.errors.repeats)
	}

	engine.errors.repeats = 0
}

// halt puts a report on screen, unless the reader has already waved this exact
// failure through or is still looking at an earlier one.
func (engine *Engine) halt(report Report) {
	if !report.Resumable {
		engine.errors.failed = true
	}

	if engine.report != nil {
		return
	}

	if engine.errors.dismissed[report.Fault.String()] {
		return
	}

	engine.report = &report
}

// internalFault describes a panic that escaped into the game loop.
//
// Nothing should panic: every failure in Lumen's own module layer comes back as
// an error object, and Ghost turns a panic inside the interpreter into one of
// these itself. This is here because "should" is not "does" — a nil texture, a
// closed font, a bug in Lumen — and a Go traceback naming files a game's author
// has never heard of tells them nothing about their game and nothing about what
// to do next.
func internalFault(where string, recovered interface{}) *fault.Fault {
	raised := fault.New(fault.Internal, "Lumen stopped unexpectedly in %s: %v", where, recovered)

	if os.Getenv("LUMEN_DEBUG") != "" {
		return raised.WithHelp("this is a bug in Lumen; please report it\n\n%s", debug.Stack())
	}

	return raised.WithHelp("this is a bug in Lumen, not in your game; please report it, and set LUMEN_DEBUG=1 for the details to include")
}

// RaiseError reports an error object raised by game code. It is the shape the
// loop actually holds a failure in, since everything Ghost hands back is an
// object.
func (engine *Engine) RaiseError(where string, resumable bool, err *object.Error) {
	if err == nil {
		return
	}

	engine.Raise(where, resumable, err.Fault)
}
