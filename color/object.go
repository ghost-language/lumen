package color

import (
	"ghostlang.org/x/ghost/object"
)

const COLOR = "Color"

// Color objects consist of a nil value.
type Color struct {
	Red   uint8
	Green uint8
	Blue  uint8
	Alpha uint8
}

// String represents the color object's value as a string.
func (color *Color) String() string {
	return "color"
}

// Type returns the color object type.
func (color *Color) Type() object.Type {
	return COLOR
}

// Method defines the set of methods available on color objects.
func (color *Color) Method(method string, args []object.Object) (object.Object, bool) {
	return nil, false
}
