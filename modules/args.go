package modules

import (
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

// The module layer converts Ghost values into Go values a great many times, and
// doing it inline turns every method into a wall of type assertions. These
// helpers keep the conversions in one place and give consistent error messages.

// arity checks that a method received an exact number of arguments.
func arity(name string, tok token.Token, args []object.Object, want int) *object.Error {
	if len(args) != want {
		return object.NewError("%d:%d: runtime error: %s() expects %d argument(s). got=%d", tok.Line, tok.Column, name, want, len(args))
	}

	return nil
}

// arityRange checks that a method received a number of arguments within bounds.
func arityRange(name string, tok token.Token, args []object.Object, low, high int) *object.Error {
	if len(args) < low || len(args) > high {
		return object.NewError("%d:%d: runtime error: %s() expects between %d and %d arguments. got=%d", tok.Line, tok.Column, name, low, high, len(args))
	}

	return nil
}

// number reads a numeric argument.
func number(name string, tok token.Token, args []object.Object, index int) (float64, *object.Error) {
	if index >= len(args) {
		return 0, object.NewError("%d:%d: runtime error: %s() is missing argument %d", tok.Line, tok.Column, name, index+1)
	}

	value, ok := args[index].(*object.Number)

	if !ok {
		return 0, object.NewError("%d:%d: runtime error: %s() expects argument %d to be a number. got=%s", tok.Line, tok.Column, name, index+1, args[index].Type())
	}

	return value.Float64(), nil
}

// numbers reads every argument as a number, which suits the geometry methods
// that take a run of coordinates.
func numbers(name string, tok token.Token, args []object.Object) ([]float64, *object.Error) {
	values := make([]float64, 0, len(args))

	for index := range args {
		value, err := number(name, tok, args, index)

		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// integer reads a numeric argument as a whole number.
func integer(name string, tok token.Token, args []object.Object, index int) (int64, *object.Error) {
	value, err := number(name, tok, args, index)

	if err != nil {
		return 0, err
	}

	return int64(value), nil
}

// text reads a string argument.
func text(name string, tok token.Token, args []object.Object, index int) (string, *object.Error) {
	if index >= len(args) {
		return "", object.NewError("%d:%d: runtime error: %s() is missing argument %d", tok.Line, tok.Column, name, index+1)
	}

	value, ok := args[index].(*object.String)

	if !ok {
		return "", object.NewError("%d:%d: runtime error: %s() expects argument %d to be a string. got=%s", tok.Line, tok.Column, name, index+1, args[index].Type())
	}

	return value.Value, nil
}

// boolean reads a boolean argument.
func boolean(name string, tok token.Token, args []object.Object, index int) (bool, *object.Error) {
	if index >= len(args) {
		return false, object.NewError("%d:%d: runtime error: %s() is missing argument %d", tok.Line, tok.Column, name, index+1)
	}

	value, ok := args[index].(*object.Boolean)

	if !ok {
		return false, object.NewError("%d:%d: runtime error: %s() expects argument %d to be a boolean. got=%s", tok.Line, tok.Column, name, index+1, args[index].Type())
	}

	return value.Value, nil
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
