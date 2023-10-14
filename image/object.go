package image

import (
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/lumen/renderer"
	"github.com/veandco/go-sdl2/sdl"
)

const IMAGE = "Image"

// Image objects consist of a nil value.
type Image struct {
	width     int32
	height    int32
	blendmode sdl.BlendMode
	surface   *sdl.Surface
	texture   *sdl.Texture
}

// String represents the image object's value as a string.
func (image *Image) String() string {
	return "image"
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

	src := &sdl.Rect{X: 0, Y: 0, W: int32(image.width), H: int32(image.height)}
	dst := &sdl.Rect{X: x, Y: y, W: int32(image.width), H: int32(image.height)}

	image.texture.SetBlendMode(image.blendmode)
	renderer.Renderer.Copy(image.texture, src, dst)
}
