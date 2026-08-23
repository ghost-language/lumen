package engine

import (
	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

// DrawArguments carries the positional arguments every drawable accepts, in the
// order they are always written: position, rotation in radians, per-axis scale,
// and an origin offset applied before rotation and scale.
type DrawArguments struct {
	X        float64
	Y        float64
	Rotation float64
	ScaleX   float64
	ScaleY   float64
	OriginX  float64
	OriginY  float64
}

// ParseDrawArguments reads draw arguments starting at the given offset. Missing
// trailing arguments fall back to the identity values, so drawing at a position
// only needs two numbers.
func ParseDrawArguments(name string, tok token.Token, args []object.Object, offset int) (DrawArguments, *object.Error) {
	arguments := DrawArguments{ScaleX: 1, ScaleY: 1}

	values := make([]float64, 0, 7)

	for index := offset; index < len(args); index++ {
		number, ok := args[index].(*object.Number)

		if !ok {
			return arguments, Error(fault.Argument, tok, "`%s` expects argument %d to be a number, got %s", Signature(name), index+1, TypeName(args[index])).
				WithHelp("after the position, %s takes rotation, scale, and origin, all of them numbers", Signature(name))
		}

		values = append(values, number.Float64())
	}

	if len(values) < 2 {
		return arguments, Error(fault.Argument, tok, "`%s` expects at least an x and y position, got %d", Signature(name), len(values))
	}

	arguments.X = values[0]
	arguments.Y = values[1]

	if len(values) > 2 {
		arguments.Rotation = values[2]
	}

	if len(values) > 3 {
		arguments.ScaleX = values[3]
		arguments.ScaleY = values[3]
	}

	if len(values) > 4 {
		arguments.ScaleY = values[4]
	}

	if len(values) > 5 {
		arguments.OriginX = values[5]
	}

	if len(values) > 6 {
		arguments.OriginY = values[6]
	}

	return arguments, nil
}
