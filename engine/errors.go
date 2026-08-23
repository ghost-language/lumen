package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

// Lumen speaks Ghost's error vocabulary.
//
// Every failure a game can cause is a *fault.Fault carried back through the
// evaluator as an error object: a kind that says what sort of mistake it is, a
// sentence that describes it, the token it happened at, and — where Lumen is
// holding the answer — a line saying what to do about it. Nothing in Lumen
// formats a position into a message by hand, and nothing prints one where it
// happens: reporting is the loop's job, and it does it twice, once to the
// console and once into the window the player is actually looking at.
//
// The helpers below are Lumen's half of that. They are deliberately worded the
// same as Ghost's, in object/arguments.go, so that a bad argument to
// canvas.print() reads exactly like a bad argument to list.join(). They are not
// simply calls into Ghost's, because Ghost cannot name a Lumen value: its
// TypeName answers "unknown" for anything outside its own type enum, and
// "expects argument 1 to be a number, got unknown" is not a message worth
// printing. TypeName below fills that hole, and everything here goes through it.

// TypeName names the type of a value the way a message about it should read.
// Lumen's own types are named here; everything else is Ghost's to name.
func TypeName(value object.Object) string {
	if value == nil {
		return "null"
	}

	switch value.Type() {
	case IMAGE:
		return "image"
	case QUAD:
		return "quad"
	case FONT:
		return "font"
	case COLOR:
		return "color"
	case TARGET:
		return "target"
	case SOURCE:
		return "source"
	case TRANSFORM:
		return "transform"
	case SPRITESHEET:
		return "spritesheet"
	case ANIMATION:
		return "animation"
	}

	return object.TypeName(value)
}

// Signature renders a method's name the way it is written in a game, so that a
// message quotes something the reader can search their own source for.
func Signature(name string) string {
	return name + "()"
}

// Error reports a failure at the token the interpreter was looking at. The kind
// says what sort of failure it is and the message says what happened; neither
// should repeat the position, which the report adds for itself.
func Error(kind fault.Kind, tok token.Token, format string, arguments ...interface{}) *object.Error {
	return object.NewError(kind, tok, format, arguments...)
}

// =============================================================================
// Arguments

// Arity checks that a call received an exact number of arguments.
func Arity(name string, tok token.Token, args []object.Object, want int) *object.Error {
	if len(args) == want {
		return nil
	}

	return Error(fault.Argument, tok, "`%s` expects %s, got %d", Signature(name), plural(want), len(args))
}

// ArityRange checks that a call received a number of arguments within bounds,
// which is what an optional trailing argument amounts to.
func ArityRange(name string, tok token.Token, args []object.Object, low int, high int) *object.Error {
	if len(args) >= low && len(args) <= high {
		return nil
	}

	return Error(fault.Argument, tok, "`%s` expects between %d and %d arguments, got %d", Signature(name), low, high, len(args))
}

// ArityAtLeast checks that a call received no fewer than a number of arguments,
// which suits the drawing methods that read a whole run of coordinates.
func ArityAtLeast(name string, tok token.Token, args []object.Object, low int) *object.Error {
	if len(args) >= low {
		return nil
	}

	return Error(fault.Argument, tok, "`%s` expects at least %s, got %d", Signature(name), plural(low), len(args))
}

// Number reads an argument that has to be a number.
func Number(name string, tok token.Token, args []object.Object, index int) (float64, *object.Error) {
	value, err := Argument(name, tok, args, index, "a number", object.NUMBER)

	if err != nil {
		return 0, err
	}

	return value.(*object.Number).Float64(), nil
}

// Numbers reads every argument as a number, which suits the geometry methods
// that take a run of coordinates.
func Numbers(name string, tok token.Token, args []object.Object) ([]float64, *object.Error) {
	values := make([]float64, 0, len(args))

	for index := range args {
		value, err := Number(name, tok, args, index)

		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// Integer reads an argument that has to be a number, as a whole number.
func Integer(name string, tok token.Token, args []object.Object, index int) (int64, *object.Error) {
	value, err := Number(name, tok, args, index)

	if err != nil {
		return 0, err
	}

	return int64(value), nil
}

// Text reads an argument that has to be a string.
func Text(name string, tok token.Token, args []object.Object, index int) (string, *object.Error) {
	value, err := Argument(name, tok, args, index, "a string", object.STRING)

	if err != nil {
		return "", err
	}

	return value.(*object.String).Value, nil
}

// Boolean reads an argument that has to be a boolean.
func Boolean(name string, tok token.Token, args []object.Object, index int) (bool, *object.Error) {
	value, err := Argument(name, tok, args, index, "a boolean", object.BOOLEAN)

	if err != nil {
		return false, err
	}

	return value.(*object.Boolean).Value, nil
}

// List reads an argument that has to be a list.
func List(name string, tok token.Token, args []object.Object, index int) (*object.List, *object.Error) {
	value, err := Argument(name, tok, args, index, "a list", object.LIST)

	if err != nil {
		return nil, err
	}

	return value.(*object.List), nil
}

// ImageArgument reads an argument that has to be an image.
func ImageArgument(name string, tok token.Token, args []object.Object, index int) (*Image, *object.Error) {
	value, err := Argument(name, tok, args, index, "an image", IMAGE)

	if err != nil {
		return nil, err
	}

	return value.(*Image), nil
}

// QuadArgument reads an argument that has to be a quad.
func QuadArgument(name string, tok token.Token, args []object.Object, index int) (*Quad, *object.Error) {
	value, err := Argument(name, tok, args, index, "a quad", QUAD)

	if err != nil {
		return nil, err
	}

	return value.(*Quad), nil
}

// FontArgument reads an argument that has to be a font.
func FontArgument(name string, tok token.Token, args []object.Object, index int) (*Font, *object.Error) {
	value, err := Argument(name, tok, args, index, "a font", FONT)

	if err != nil {
		return nil, err
	}

	return value.(*Font), nil
}

// ColorArgument reads an argument that has to be a color.
func ColorArgument(name string, tok token.Token, args []object.Object, index int) (*Color, *object.Error) {
	value, err := Argument(name, tok, args, index, "a color", COLOR)

	if err != nil {
		return nil, err
	}

	return value.(*Color), nil
}

// TargetArgument reads an argument that has to be a render target.
func TargetArgument(name string, tok token.Token, args []object.Object, index int) (*Target, *object.Error) {
	value, err := Argument(name, tok, args, index, "a target", TARGET)

	if err != nil {
		return nil, err
	}

	return value.(*Target), nil
}

// SourceArgument reads an argument that has to be an audio source.
func SourceArgument(name string, tok token.Token, args []object.Object, index int) (*Source, *object.Error) {
	value, err := Argument(name, tok, args, index, "a source", SOURCE)

	if err != nil {
		return nil, err
	}

	return value.(*Source), nil
}

// SpritesheetArgument reads an argument that has to be a spritesheet.
func SpritesheetArgument(name string, tok token.Token, args []object.Object, index int) (*Spritesheet, *object.Error) {
	value, err := Argument(name, tok, args, index, "a spritesheet", SPRITESHEET)

	if err != nil {
		return nil, err
	}

	return value.(*Spritesheet), nil
}

// Argument reads and type-checks one argument. Everything above is a thin
// wrapper over it, which is what keeps the two messages it can produce the only
// two messages in Lumen for a bad argument.
//
// The article is the phrase the type reads as in a sentence — "a number", "an
// image" — because "expects argument 1 to be image" is not English and a
// message a reader trips over is a message they have to read twice.
func Argument(name string, tok token.Token, args []object.Object, index int, article string, want object.Type) (object.Object, *object.Error) {
	if index >= len(args) {
		return nil, Missing(name, tok, index)
	}

	value := args[index]

	if value == nil || value.Type() != want {
		return nil, Mistyped(name, tok, index, article, value)
	}

	return value, nil
}

// Mistyped reports an argument of the wrong type. Argument is the way to reach
// it; this is here for the few methods that accept more than one type in a
// position and so have to do the checking themselves, and should still word the
// result the way everything else does.
func Mistyped(name string, tok token.Token, index int, article string, value object.Object) *object.Error {
	return Error(fault.Argument, tok, "`%s` expects argument %d to be %s, got %s", Signature(name), index+1, article, TypeName(value))
}

// Missing reports an argument that was never passed. It is kept apart from a
// wrong-typed one because the fix is different: one is a value to change, the
// other is a value to add.
func Missing(name string, tok token.Token, index int) *object.Error {
	return Error(fault.Argument, tok, "`%s` is missing argument %d", Signature(name), index+1)
}

// plural writes an argument count as the phrase it belongs in, so a message
// says "1 argument" rather than "1 arguments".
func plural(count int) string {
	if count == 1 {
		return "1 argument"
	}

	return fmt.Sprintf("%d arguments", count)
}

// =============================================================================
// Values

// maxChoices is how many accepted names are worth printing. Past a handful the
// list stops being an answer and starts being something to scroll past, and the
// suggestion underneath it was doing the work anyway.
const maxChoices = 8

// Choice reports a string argument that is not one of the names a method
// accepts. Lumen has a good many of these — blend modes, animation modes,
// alignments, source kinds — and a reader who has misremembered one is helped
// far more by the list of the real ones than by being told theirs is wrong.
func Choice(name string, tok token.Token, given string, valid ...string) *object.Error {
	if len(valid) > maxChoices {
		return Unknown(name, tok, "", given, valid...)
	}

	raised := Error(fault.Value, tok, "`%s` expects %s, got `%s`", Signature(name), oneOf(valid), given)

	if suggestion, ok := NearestName(given, valid); ok {
		raised.WithHelp("did you mean `%s`?", suggestion)
	}

	return raised
}

// Unknown reports a name drawn from a set too large to print: every key on the
// keyboard, every button a controller mapping knows. The message says only that
// this one is not among them, and the suggestion says which one was meant.
//
// The noun is what the name names — "key", "button" — and may be left empty
// where the sentence reads better without it.
func Unknown(name string, tok token.Token, noun string, given string, known ...string) *object.Error {
	subject := "`" + given + "`"

	if noun != "" {
		subject = "the " + noun + " `" + given + "`"
	}

	raised := Error(fault.Value, tok, "`%s` does not recognise %s", Signature(name), subject)

	if suggestion, ok := NearestName(given, known); ok {
		raised.WithHelp("did you mean `%s`?", suggestion)
	}

	return raised
}

// Value reports a value of the right type that a method cannot accept: a
// negative size, a frame that is not in the sheet, a transform with no inverse.
func Value(name string, tok token.Token, format string, arguments ...interface{}) *object.Error {
	return Error(fault.Value, tok, "`%s` %s", Signature(name), fmt.Sprintf(format, arguments...))
}

// State reports a method called on a value that cannot answer it — panning a
// streamed source, reading pixels from a canvas that has none. The value is the
// right type and the arguments are fine; it is the thing itself that cannot do
// this, which is what makes it a type error rather than an argument one.
func State(name string, tok token.Token, format string, arguments ...interface{}) *object.Error {
	return Error(fault.Type, tok, "`%s` %s", Signature(name), fmt.Sprintf(format, arguments...))
}

// oneOf renders a list of accepted names as the phrase it reads as.
func oneOf(names []string) string {
	quoted := make([]string, 0, len(names))

	for _, name := range names {
		quoted = append(quoted, "`"+name+"`")
	}

	switch len(quoted) {
	case 0:
		return "a different value"
	case 1:
		return quoted[0]
	case 2:
		return quoted[0] + " or " + quoted[1]
	}

	return strings.Join(quoted[:len(quoted)-1], ", ") + ", or " + quoted[len(quoted)-1]
}

// =============================================================================
// The world outside the game

// SystemFailure reports the machine refusing to cooperate: a window that will
// not go fullscreen, a save file that will not open, a clipboard that is not
// there. The underlying error is quoted as-is, because what it says — "no such
// file or directory", "permission denied" — is already the clearest description
// available.
func SystemFailure(name string, tok token.Token, failure error) *object.Error {
	return Error(fault.System, tok, "`%s` failed: %s", Signature(name), failure)
}

// AssetFailure reports an asset that could not be loaded.
//
// Nearly every one of these is a path that does not exist, and nearly every one
// of those is a typo or a file in the folder next door. Lumen is standing in
// the directory the game asked for and can see what is actually in it, so it
// says: the nearest name gets offered outright, and a path that is simply
// missing gets told where Lumen looked, which is the other half of the mistake
// — an asset path is relative to the game's own folder, not to wherever the
// player happened to run it from.
func AssetFailure(name string, tok token.Token, path string, resolved string, failure error) *object.Error {
	raised := Error(fault.System, tok, "`%s` could not load `%s`: %s", Signature(name), path, reason(failure, resolved))

	if suggestion, ok := nearestFile(resolved); ok {
		return raised.WithHelp("did you mean `%s`?", suggestion)
	}

	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		return raised.WithHelp("Lumen looked in %s; asset paths are relative to the game's folder", filepath.Dir(resolved))
	}

	return raised
}

// reason trims the noise off a failure that is about to be quoted.
//
// Both halves of Lumen's asset loading answer with the path folded into the
// message — Go writes "open /a/long/path: no such file or directory" and SDL
// writes "Couldn't open /a/long/path: No such file or directory" — and the path
// is already in the sentence this is being folded into. Where the path is in
// there, whatever follows the last colon is the part that says something new.
func reason(failure error, resolved string) string {
	if failure == nil {
		return "unknown error"
	}

	message := failure.Error()

	if resolved == "" || !strings.Contains(message, resolved) {
		return message
	}

	if index := strings.LastIndex(message, ": "); index >= 0 {
		return message[index+2:]
	}

	return message
}

// nearestFile looks for a file next to the one that was asked for, whose name
// is close enough to the one that was asked for to be worth mentioning.
func nearestFile(resolved string) (string, bool) {
	entries, err := os.ReadDir(filepath.Dir(resolved))

	if err != nil {
		return "", false
	}

	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return NearestName(filepath.Base(resolved), names)
}

// =============================================================================
// Suggestions

// NearestName picks the candidate closest to a misspelling, if one is close
// enough to be worth mentioning. The threshold scales with the length of the
// name: a two-letter name has to match almost exactly, while a longer one can
// survive a couple of slips. Guessing wrongly is worse than not guessing, so
// the bar is deliberately high.
//
// This is Ghost's rule, applied to the names Lumen knows about: the modes a
// method accepts, and the files sitting next to the one a game asked for. A
// reader who has been offered `did you mean` once should be offered it
// everywhere the answer is knowable.
func NearestName(name string, candidates []string) (string, bool) {
	limit := len(name) / 3

	if limit < 1 {
		limit = 1
	}

	if limit > 3 {
		limit = 3
	}

	best := ""
	bestDistance := limit + 1

	// Sorting keeps the answer stable: a directory listing and a map of modes
	// both arrive in an order nothing promises, and two candidates the same
	// distance away would otherwise alternate between runs.
	sorted := append([]string(nil), candidates...)
	sort.Strings(sorted)

	for _, candidate := range sorted {
		if candidate == name {
			continue
		}

		distance := editDistance(strings.ToLower(name), strings.ToLower(candidate))

		if distance > bestDistance {
			continue
		}

		if distance < bestDistance || sharesMore(name, candidate, best) {
			best = candidate
			bestDistance = distance
		}
	}

	return best, best != ""
}

// sharesMore breaks a tie between two candidates the same distance from a
// misspelling. The one that agrees with it for longer wins, so `playr.png` is
// answered with `player.png` rather than with `plate.png`, which is equally
// close by the arithmetic and obviously further away to a reader.
func sharesMore(name string, candidate string, best string) bool {
	if best == "" {
		return true
	}

	return commonPrefix(name, candidate) > commonPrefix(name, best)
}

// commonPrefix counts the characters two names begin with in common.
func commonPrefix(left string, right string) int {
	shared := 0

	for shared < len(left) && shared < len(right) && left[shared] == right[shared] {
		shared++
	}

	return shared
}

// editDistance is the Damerau-Levenshtein distance between two strings: the
// number of single-character insertions, deletions, substitutions, or
// transpositions that turn one into the other.
//
// Transpositions are counted because they are what typing actually produces.
// Plain Levenshtein scores `pignpong` against `pingpong` as two changes and
// would put it out of reach of any threshold tight enough to be trusted;
// counting the swap as the one slip it was brings the suggestion back.
func editDistance(from string, to string) int {
	source := []rune(from)
	target := []rune(to)

	if len(source) == 0 {
		return len(target)
	}

	if len(target) == 0 {
		return len(source)
	}

	// A transposition needs the row before last, so three rows are kept rather
	// than the whole grid.
	beforeLast := make([]int, len(target)+1)
	previous := make([]int, len(target)+1)
	current := make([]int, len(target)+1)

	for column := range previous {
		previous[column] = column
	}

	for row := 1; row <= len(source); row++ {
		current[0] = row

		for column := 1; column <= len(target); column++ {
			cost := 1

			if source[row-1] == target[column-1] {
				cost = 0
			}

			current[column] = smallest(
				current[column-1]+1,
				previous[column]+1,
				previous[column-1]+cost,
			)

			if row > 1 && column > 1 && source[row-1] == target[column-2] && source[row-2] == target[column-1] {
				current[column] = smallest(current[column], beforeLast[column-2]+1)
			}
		}

		beforeLast, previous, current = previous, current, beforeLast
	}

	return previous[len(target)]
}

func smallest(values ...int) int {
	least := values[0]

	for _, value := range values[1:] {
		if value < least {
			least = value
		}
	}

	return least
}
