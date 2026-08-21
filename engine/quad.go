package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
)

// Quad is a rectangular region of a texture. It is what lets one tileset or
// character sheet be drawn as many separate sprites without slicing the source
// image into individual files.
type Quad struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

// NewQuad builds a quad covering the given region.
func NewQuad(x, y, width, height int32) *Quad {
	return &Quad{X: x, Y: y, Width: width, Height: height}
}

// Rect returns the quad as an SDL source rectangle.
func (quad *Quad) Rect() *sdl.Rect {
	return &sdl.Rect{X: quad.X, Y: quad.Y, W: quad.Width, H: quad.Height}
}

// String represents the quad object's value as a string.
func (quad *Quad) String() string {
	return fmt.Sprintf("Quad: %d, %d, %d, %d", quad.X, quad.Y, quad.Width, quad.Height)
}

// Type returns the quad object type.
func (quad *Quad) Type() object.Type {
	return QUAD
}

// Method defines the set of methods available on quad objects.
func (quad *Quad) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "getX":
		return object.NewInt(int64(quad.X)), true
	case "getY":
		return object.NewInt(int64(quad.Y)), true
	case "getWidth":
		return object.NewInt(int64(quad.Width)), true
	case "getHeight":
		return object.NewInt(int64(quad.Height)), true
	case "setViewport":
		return quad.setViewport(args)
	case "toString":
		return &object.String{Value: quad.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// setViewport moves the quad's region without allocating a new quad, which keeps
// per-frame animation from churning objects.
func (quad *Quad) setViewport(args []object.Object) (object.Object, bool) {
	if len(args) != 4 {
		return object.NewError("quad.setViewport() expects 4 arguments. got=%d", len(args)), true
	}

	for index, arg := range args {
		number, ok := arg.(*object.Number)

		if !ok {
			return object.NewError("quad.setViewport() expects number arguments. argument %d is %s", index+1, arg.Type()), true
		}

		switch index {
		case 0:
			quad.X = int32(number.Int64())
		case 1:
			quad.Y = int32(number.Int64())
		case 2:
			quad.Width = int32(number.Int64())
		case 3:
			quad.Height = int32(number.Int64())
		}
	}

	return value.NULL, true
}
