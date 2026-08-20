package engine

import "ghostlang.org/x/ghost/object"

// DrawArguments carries the positional arguments every drawable accepts, laid
// out the same way as LOVE's love.graphics.draw: position, rotation in radians,
// per-axis scale, and an origin offset applied before rotation and scale.
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
func ParseDrawArguments(name string, args []object.Object, offset int) (DrawArguments, *object.Error) {
	arguments := DrawArguments{ScaleX: 1, ScaleY: 1}

	values := make([]float64, 0, 7)

	for index := offset; index < len(args); index++ {
		number, ok := args[index].(*object.Number)

		if !ok {
			return arguments, object.NewError("%s() expects number arguments. argument %d is %s", name, index+1, args[index].Type())
		}

		values = append(values, number.Float64())
	}

	if len(values) < 2 {
		return arguments, object.NewError("%s() expects at least an x and y position. got=%d", name, len(values))
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

// Float reads a number argument, reporting a useful error when the caller passed
// something else.
func Float(name string, args []object.Object, index int) (float64, *object.Error) {
	if index >= len(args) {
		return 0, object.NewError("%s() is missing argument %d", name, index+1)
	}

	number, ok := args[index].(*object.Number)

	if !ok {
		return 0, object.NewError("%s() expects argument %d to be a number. got=%s", name, index+1, args[index].Type())
	}

	return number.Float64(), nil
}
