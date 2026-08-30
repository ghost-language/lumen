package engine

import (
	"os"
	"strings"
	"testing"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/source"
	"ghostlang.org/x/ghost/token"
)

func TestMarkedSpanFindsTheLexemeInBytes(t *testing.T) {
	// Columns are counted in characters and text is sliced in bytes, and the
	// two are only the same until a game writes a word with an accent in it.
	text := `    canvas.print(café, 10, 10)`

	start, end := markedSpan(text, fault.Position{Column: 18, Length: 4})

	if text[start:end] != "café" {
		t.Errorf("marked %q, want café", text[start:end])
	}
}

func TestMarkedSpanStaysInsideTheLine(t *testing.T) {
	text := "x = 1"

	start, end := markedSpan(text, fault.Position{Column: 5, Length: 40})

	if start > len(text) || end > len(text) || start > end {
		t.Errorf("marked %d..%d, which is not inside a %d byte line", start, end, len(text))
	}
}

func TestMarkedSpanMarksSomethingWhenNothingWasMeasured(t *testing.T) {
	start, end := markedSpan("x = 1", fault.Position{Column: 1, Length: 0})

	if end <= start {
		t.Errorf("marked %d..%d, want at least one character", start, end)
	}
}

func TestFirstParagraphLeavesTheStackOnTheConsole(t *testing.T) {
	help := "this is a bug in Lumen; please report it\n\ngoroutine 1 [running]:\nengine.draw(...)"

	if got := firstParagraph(help); got != "this is a bug in Lumen; please report it" {
		t.Errorf("got %q, want the first line alone", got)
	}
}

func TestDescribeFrameReadsTheSameAsTheConsole(t *testing.T) {
	frame := fault.Frame{Name: "sum()", Position: fault.Position{File: "main.gs", Line: 9, Column: 1}}

	if got := describeFrame(frame); got != "in sum(), called at main.gs:9:1" {
		t.Errorf("got %q", got)
	}

	if got := describeFrame(fault.Frame{Name: "draw()"}); got != "in draw()" {
		t.Errorf("got %q", got)
	}
}

// TestTheErrorScreenDraws runs the whole of the on-screen report against a real
// renderer, headlessly. It asserts nothing about the pixels: what it is here to
// catch is the drawing path failing on a report — a nil texture, a font that
// was never loaded, a measurement that runs off the end of a line — which is
// the one failure a game's author has no way to see past.
func TestTheErrorScreenDraws(t *testing.T) {
	os.Setenv("SDL_VIDEODRIVER", "dummy")
	os.Setenv("SDL_AUDIODRIVER", "dummy")

	engine := New("Lumen")
	defer engine.shutdown()

	engine.SetReportWriter(&strings.Builder{})

	source.Register("main.gs", "function draw() {\n\tcanvas.print(score, 10, 10)\n}\n")

	reports := []*fault.Fault{
		fault.At(fault.Argument, token.Token{File: "main.gs", Line: 2, Column: 15, Length: 5},
			"`canvas.print()` expects argument 1 to be a string, got number").
			WithHelp("did you mean `text(score)`?"),

		// A failure with no source to quote, and one whose position is past the
		// end of the line it names.
		fault.New(fault.Internal, "Lumen stopped unexpectedly"),
		fault.At(fault.Type, token.Token{File: "main.gs", Line: 2, Column: 400, Length: 9}, "a long way off the end"),
		fault.At(fault.Value, token.Token{File: "nowhere.gs", Line: 40, Column: 1, Length: 1}, "a file that was never scanned"),
		fault.New(fault.Syntax, strings.Repeat("a very long message that has to wrap ", 20)),
	}

	for _, raised := range reports {
		engine.report = &Report{Fault: raised, Where: "draw()", Resumable: true}

		engine.drawReport()
	}
}
