package modules

import (
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

// The module layer converts Ghost values into Go values a great many times, and
// doing it inline turns every method into a wall of type assertions. These
// helpers keep the conversions in one place.
//
// They are thin wrappers over the checks in engine/errors.go rather than a
// second set of rules. A bad argument to a module method, to a method on an
// image, and to a method on a Ghost list are all described by the same sentence,
// because all three of them end up in the same function.

// arity checks that a method received an exact number of arguments.
func arity(name string, tok token.Token, args []object.Object, want int) *object.Error {
	return engine.Arity(name, tok, args, want)
}

// arityRange checks that a method received a number of arguments within bounds.
func arityRange(name string, tok token.Token, args []object.Object, low, high int) *object.Error {
	return engine.ArityRange(name, tok, args, low, high)
}

// arityAtLeast checks that a method received no fewer than a number of
// arguments, which suits the drawing methods that read a run of coordinates.
func arityAtLeast(name string, tok token.Token, args []object.Object, low int) *object.Error {
	return engine.ArityAtLeast(name, tok, args, low)
}

// number reads a numeric argument.
func number(name string, tok token.Token, args []object.Object, index int) (float64, *object.Error) {
	return engine.Number(name, tok, args, index)
}

// numbers reads every argument as a number, which suits the geometry methods
// that take a run of coordinates.
func numbers(name string, tok token.Token, args []object.Object) ([]float64, *object.Error) {
	return engine.Numbers(name, tok, args)
}

// integer reads a numeric argument as a whole number.
func integer(name string, tok token.Token, args []object.Object, index int) (int64, *object.Error) {
	return engine.Integer(name, tok, args, index)
}

// text reads a string argument.
func text(name string, tok token.Token, args []object.Object, index int) (string, *object.Error) {
	return engine.Text(name, tok, args, index)
}

// boolean reads a boolean argument.
func boolean(name string, tok token.Token, args []object.Object, index int) (bool, *object.Error) {
	return engine.Boolean(name, tok, args, index)
}

// list builds a Ghost list from Go floats, for methods that return several
// numbers at once.
func list(values ...float64) *object.List {
	elements := make([]object.Object, 0, len(values))

	for _, value := range values {
		elements = append(elements, object.NewFloat(value))
	}

	return &object.List{Elements: elements}
}

// integerList builds a Ghost list from whole numbers.
func integerList(values ...int64) *object.List {
	elements := make([]object.Object, 0, len(values))

	for _, value := range values {
		elements = append(elements, object.NewInt(value))
	}

	return &object.List{Elements: elements}
}
