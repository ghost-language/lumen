package engine

import (
	"ghostlang.org/x/ghost/object"
	"github.com/shopspring/decimal"
	"github.com/veandco/go-sdl2/sdl"
)

const IMAGE = "Image"

type Image struct {
	Width     int32
	Height    int32
	XOffset   int32
	YOffset   int32
	Size      int32
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
	case "clip":
		return image.clip(args), true
	case "getWidth":
		return image.getWidth(args), true
	case "getHeight":
		return image.getHeight(args), true
	}

	return nil, false
}

// =============================================================================
// Object methods

func (image *Image) draw(args []object.Object) {
	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	x, y = Lumen.ApplyOffset(x, y)

	var (
		src *sdl.Rect
		dst *sdl.Rect
	)

	// If we have an X or Y offset and size, clip the image accordingly.
	if image.XOffset != 0 || image.YOffset != 0 || image.Size != 0 {
		src = &sdl.Rect{X: image.XOffset, Y: image.YOffset, W: image.Size, H: image.Size}
		dst = &sdl.Rect{X: x, Y: y, W: image.Size, H: image.Size}
	} else {
		src = &sdl.Rect{X: 0, Y: 0, W: int32(image.Width), H: int32(image.Height)}
		dst = &sdl.Rect{X: x, Y: y, W: int32(image.Width), H: int32(image.Height)}
	}

	image.Texture.SetBlendMode(image.BlendMode)
	Lumen.Renderer.Copy(image.Texture, src, dst)

	// After drawing, reset the X and Y offset.
	image.XOffset = 0
	image.YOffset = 0
	image.Size = 0
}

// clip applies an X and Y offset to the image.
func (image *Image) clip(args []object.Object) object.Object {
	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	size := int32(args[2].(*object.Number).Value.IntPart())

	image.XOffset = x
	image.YOffset = y
	image.Size = size

	return image
}

// getWidth returns the width of the image.
func (image *Image) getWidth(args []object.Object) object.Object {
	return &object.Number{Value: decimal.NewFromInt32(image.Width)}
}

// getHeight returns the height of the image.
func (image *Image) getHeight(args []object.Object) object.Object {
	return &object.Number{Value: decimal.NewFromInt32(image.Height)}
}
