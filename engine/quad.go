package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
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

	// rect backs Rect(). Handing out a pointer to a field rather than to a
	// fresh value keeps drawing a tile free of allocations.
	rect sdl.Rect
}

// NewQuad builds a quad covering the given region.
func NewQuad(x, y, width, height int32) *Quad {
	return &Quad{X: x, Y: y, Width: width, Height: height}
}

// Rect returns the quad as an SDL source rectangle. The rectangle is owned by
// the quad and is rewritten on every call, so callers should read it before
// asking another quad for its own — which is all any of them do.
func (quad *Quad) Rect() *sdl.Rect {
	quad.rect = sdl.Rect{X: quad.X, Y: quad.Y, W: quad.Width, H: quad.Height}

	return &quad.rect
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
func (quad *Quad) Method(method string, tok token.Token, args []object.Object) (object.Object, bool) {
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
		return quad.setViewport(tok, args)
	case "toString":
		return &object.String{Value: quad.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// setViewport moves the quad's region without allocating a new quad, which keeps
// per-frame animation from churning objects.
func (quad *Quad) setViewport(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("quad.setViewport", tok, args, 4); err != nil {
		return err, true
	}

	values, err := Numbers("quad.setViewport", tok, args)

	if err != nil {
		return err, true
	}

	quad.X = int32(values[0])
	quad.Y = int32(values[1])
	quad.Width = int32(values[2])
	quad.Height = int32(values[3])

	return value.NULL, true
}
