package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
)

// Target is an off-screen render target, LOVE's Canvas. Drawing into one and
// then drawing the result as a texture is how effects like screen fades, light
// maps, and fixed-resolution pixel-art scaling are built.
type Target struct {
	Width   int32
	Height  int32
	Texture *sdl.Texture
}

// NewTarget creates an off-screen surface that can be drawn into.
func NewTarget(width, height int32) (*Target, error) {
	texture, err := Lumen.Renderer.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET,
		width,
		height,
	)

	if err != nil {
		return nil, err
	}

	texture.SetBlendMode(sdl.BLENDMODE_BLEND)

	target := &Target{Width: width, Height: height, Texture: texture}

	Lumen.RegisterResource(target)

	return target, nil
}

// String represents the target object's value as a string.
func (target *Target) String() string {
	return fmt.Sprintf("Target: %dx%d", target.Width, target.Height)
}

// Type returns the target object type.
func (target *Target) Type() object.Type {
	return TARGET
}

// Method defines the set of methods available on target objects.
func (target *Target) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "draw":
		return target.draw(args)
	case "getWidth":
		return object.NewInt(int64(target.Width)), true
	case "getHeight":
		return object.NewInt(int64(target.Height)), true
	case "getDimensions":
		return &object.List{Elements: []object.Object{
			object.NewInt(int64(target.Width)),
			object.NewInt(int64(target.Height)),
		}}, true
	case "toString":
		return &object.String{Value: target.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// draw renders the target's contents like any other texture.
func (target *Target) draw(args []object.Object) (object.Object, bool) {
	arguments, err := ParseDrawArguments("target.draw", args, 0)

	if err != nil {
		return err, true
	}

	Lumen.DrawTexture(target.Texture, nil, target.Width, target.Height,
		arguments.X, arguments.Y, arguments.Rotation,
		arguments.ScaleX, arguments.ScaleY, arguments.OriginX, arguments.OriginY)

	return value.NULL, true
}

// Release frees the target's texture.
func (target *Target) Release() {
	if target.Texture != nil {
		target.Texture.Destroy()
		target.Texture = nil
	}
}

// =============================================================================
// Engine helpers

// SetTarget routes subsequent drawing into an off-screen target. Passing nil
// restores drawing to the window.
func (engine *Engine) SetTarget(target *Target) error {
	engine.Graphics.Target = target

	if target == nil {
		return engine.Renderer.SetRenderTarget(nil)
	}

	return engine.Renderer.SetRenderTarget(target.Texture)
}
