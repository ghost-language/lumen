package engine

import (
	"ghostlang.org/x/ghost/object"
	"github.com/veandco/go-sdl2/sdl"
)

const IMAGE = "Image"

// Image objects consist of a nil value.
type Image struct {
	Width     int32
	Height    int32
	Path      string
	BlendMode sdl.BlendMode
	Surface   *sdl.Surface
	Texture   *sdl.Texture
}

// String represents the image object's value as a string.
func (image *Image) String() string {
	return "Image: " + image.Path
}

// Type returns the image object type.
func (image *Image) Type() object.Type {
	return IMAGE
}

// Method defines the set of methods available on image objects.
func (image *Image) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "draw":
		image.draw(args)
	}

	return nil, false
}

// =============================================================================
// Object methods

func (image *Image) draw(args []object.Object) {
	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	x, y = Lumen.ApplyOffset(x, y)

	src := &sdl.Rect{X: 0, Y: 0, W: int32(image.Width), H: int32(image.Height)}
	dst := &sdl.Rect{X: x, Y: y, W: int32(image.Width), H: int32(image.Height)}

	image.Texture.SetBlendMode(image.BlendMode)
	Lumen.Renderer.Copy(image.Texture, src, dst)
}
