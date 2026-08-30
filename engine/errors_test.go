package engine

import (
	"strings"
	"testing"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

// The wording of these messages is the point of them, so the tests check the
// sentences rather than the kinds: a reader who has read one argument error
// should be able to predict every other one, and that is only true for as long
// as they are all written by the helpers here.

func at(line int, column int) token.Token {
	return token.Token{File: "main.gs", Line: line, Column: column, Length: 4, Lexeme: "draw"}
}

func TestArityReadsAsASentence(t *testing.T) {
	tests := []struct {
		name  string
		args  []object.Object
		want  int
		error string
	}{
		{"canvas.rotate", nil, 1, "`canvas.rotate()` expects 1 argument, got 0"},
		{"canvas.translate", []object.Object{object.NewInt(1)}, 2, "`canvas.translate()` expects 2 arguments, got 1"},
	}

	for _, test := range tests {
		raised := Arity(test.name, at(1, 1), test.args, test.want)

		if raised == nil {
			t.Fatalf("%s: expected an error", test.name)
		}

		if raised.Message() != test.error {
			t.Errorf("%s: got %q, want %q", test.name, raised.Message(), test.error)
		}
	}
}

func TestArityAcceptsWhatItShould(t *testing.T) {
	args := []object.Object{object.NewInt(1), object.NewInt(2)}

	if raised := Arity("canvas.translate", at(1, 1), args, 2); raised != nil {
		t.Errorf("expected no error, got %q", raised.Message())
	}

	if raised := ArityRange("canvas.scale", at(1, 1), args, 1, 2); raised != nil {
		t.Errorf("expected no error, got %q", raised.Message())
	}

	if raised := ArityAtLeast("canvas.point", at(1, 1), args, 2); raised != nil {
		t.Errorf("expected no error, got %q", raised.Message())
	}
}

func TestArgumentNamesLumenTypes(t *testing.T) {
	// Ghost has no name for a Lumen value and calls it "unknown". A message
	// that says "got unknown" is not worth printing, which is what TypeName and
	// these wrappers exist to prevent.
	args := []object.Object{NewQuad(0, 0, 16, 16)}

	_, raised := Number("image.draw", at(3, 5), args, 0)

	if raised == nil {
		t.Fatal("expected an error")
	}

	want := "`image.draw()` expects argument 1 to be a number, got quad"

	if raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}
}

func TestMissingArgumentIsItsOwnMistake(t *testing.T) {
	_, raised := Text("font.load", at(2, 2), nil, 0)

	if raised == nil {
		t.Fatal("expected an error")
	}

	if want := "`font.load()` is missing argument 1"; raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}
}

func TestErrorsCarryThePositionTheyHappenedAt(t *testing.T) {
	raised := Arity("canvas.rotate", at(7, 11), nil, 1)

	if raised.Fault.Position.Line != 7 || raised.Fault.Position.Column != 11 {
		t.Errorf("got %s, want main.gs:7:11", raised.Fault.Position)
	}

	if raised.Fault.Position.Length != 4 {
		t.Errorf("got a width of %d, want the width of the lexeme", raised.Fault.Position.Length)
	}
}

func TestChoiceListsWhatIsAcceptedAndSuggestsTheNearest(t *testing.T) {
	raised := Choice("animation.setMode", at(1, 1), "pignpong", AnimationModeNames...)

	want := "`animation.setMode()` expects `loop`, `once`, or `pingpong`, got `pignpong`"

	if raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}

	if raised.Fault.Help != "did you mean `pingpong`?" {
		t.Errorf("got help %q, want a suggestion of pingpong", raised.Fault.Help)
	}

	if raised.Fault.Kind != fault.Value {
		t.Errorf("got kind %s, want a value error", raised.Fault.Kind)
	}
}

func TestChoiceKeepsQuietWhenNothingIsClose(t *testing.T) {
	raised := Choice("canvas.setBlendMode", at(1, 1), "shatter", BlendModeNames...)

	if raised.Fault.Help != "" {
		t.Errorf("got help %q, want none: nothing is close enough to guess at", raised.Fault.Help)
	}
}

func TestOneOfReadsAsEnglish(t *testing.T) {
	tests := []struct {
		names []string
		want  string
	}{
		{nil, "a different value"},
		{[]string{"static"}, "`static`"},
		{[]string{"static", "stream"}, "`static` or `stream`"},
		{[]string{"loop", "once", "pingpong"}, "`loop`, `once`, or `pingpong`"},
	}

	for _, test := range tests {
		if got := oneOf(test.names); got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

func TestNearestNameGuessesOnlyWhenItIsWorthIt(t *testing.T) {
	files := []string{"player.png", "tileset.png", "music.ogg"}

	if suggestion, ok := NearestName("playr.png", files); !ok || suggestion != "player.png" {
		t.Errorf("got %q, want player.png", suggestion)
	}

	// A transposition is what typing actually produces, and is counted as the
	// one slip it was.
	if suggestion, ok := NearestName("tilseet.png", files); !ok || suggestion != "tileset.png" {
		t.Errorf("got %q, want tileset.png", suggestion)
	}

	if suggestion, ok := NearestName("dragon.png", files); ok {
		t.Errorf("got %q, want no guess at all", suggestion)
	}
}

func TestNearestNameBreaksTiesOnTheSharedPrefix(t *testing.T) {
	// Both are one edit away by the arithmetic, and only one of them is close
	// to a reader.
	if suggestion, _ := NearestName("pii", []string{"phi", "pi"}); suggestion != "pi" {
		t.Errorf("got %q, want pi", suggestion)
	}
}

func TestSystemFailureQuotesTheReasonAsGiven(t *testing.T) {
	raised := SystemFailure("filesystem.write", at(1, 1), errorString("permission denied"))

	if want := "`filesystem.write()` failed: permission denied"; raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}

	if raised.Fault.Kind != fault.System {
		t.Errorf("got kind %s, want a system error", raised.Fault.Kind)
	}
}

func TestAssetFailureDropsThePathFromTheQuotedReason(t *testing.T) {
	// SDL and Go both fold the path into their message, and the path is already
	// in the sentence it is being folded into.
	resolved := "/games/mine/playr.png"

	raised := AssetFailure("image.load", at(1, 1), "playr.png", resolved,
		errorString("Couldn't open "+resolved+": No such file or directory"))

	want := "`image.load()` could not load `playr.png`: No such file or directory"

	if raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}
}

func TestAssetFailureKeepsAReasonThatIsNotAboutAPath(t *testing.T) {
	raised := AssetFailure("image.load", at(1, 1), "sheet.png", "/games/mine/sheet.png",
		errorString("Unsupported image format"))

	if !strings.HasSuffix(raised.Message(), "Unsupported image format") {
		t.Errorf("got %q, want the reason kept whole", raised.Message())
	}
}

func TestAssetFailureSuggestsTheFileNextToIt(t *testing.T) {
	directory := t.TempDir()

	if err := writeFile(directory+"/player.png", "not really a png"); err != nil {
		t.Fatal(err)
	}

	raised := AssetFailure("image.load", at(1, 1), "playr.png", directory+"/playr.png",
		errorString("no such file or directory"))

	if raised.Fault.Help != "did you mean `player.png`?" {
		t.Errorf("got help %q, want a suggestion of player.png", raised.Fault.Help)
	}
}

func TestAssetFailureSaysWhereItLookedWhenItCannotGuess(t *testing.T) {
	directory := t.TempDir()

	raised := AssetFailure("image.load", at(1, 1), "player.png", directory+"/player.png",
		errorString("no such file or directory"))

	if !strings.Contains(raised.Fault.Help, directory) {
		t.Errorf("got help %q, want it to name the folder Lumen looked in", raised.Fault.Help)
	}
}

func TestTypeNameNamesBothLanguages(t *testing.T) {
	tests := []struct {
		value object.Object
		want  string
	}{
		{object.NewInt(1), "number"},
		{&object.String{Value: "x"}, "string"},
		{NewQuad(0, 0, 1, 1), "quad"},
		{NewColor(0, 0, 0, 255), "color"},
		{nil, "null"},
	}

	for _, test := range tests {
		if got := TypeName(test.value); got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

func TestALongListOfNamesIsNotPrinted(t *testing.T) {
	// Every key on the keyboard is not an answer to a question about one of
	// them, and the suggestion underneath was doing the work anyway.
	raised := Choice("keyboard.isDown", at(1, 1), "spcae", KeyNames()...)

	if strings.Contains(raised.Message(), "Backspace") {
		t.Errorf("got %q, want a message that does not list the whole keyboard", raised.Message())
	}

	if raised.Fault.Help != "did you mean `Space`?" {
		t.Errorf("got help %q, want a suggestion of Space", raised.Fault.Help)
	}
}

func TestUnknownNamesWhatItDidNotRecognise(t *testing.T) {
	raised := Unknown("joystick.isDown", at(1, 1), "button", "stat", "start", "back", "guide")

	if want := "`joystick.isDown()` does not recognise the button `stat`"; raised.Message() != want {
		t.Errorf("got %q, want %q", raised.Message(), want)
	}

	if raised.Fault.Help != "did you mean `start`?" {
		t.Errorf("got help %q, want a suggestion of start", raised.Fault.Help)
	}
}
