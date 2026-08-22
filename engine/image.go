package engine

import (
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
)

// Image is a texture loaded from disk, ready to be drawn. Views produced by
// clip() share the underlying texture with the image they came from and never
// own it, so freeing resources stays the loader's responsibility.
type Image struct {
	Width   int32
	Height  int32
	Path    string
	Quad    *Quad
	Surface *sdl.Surface
	Texture *sdl.Texture
	owned   bool

	// parent is the image a view was clipped from. A view reads its texture
	// through the parent rather than copying the pointer, so a texture rebuilt
	// by setFilter() is picked up by every view of it instead of leaving them
	// pointing at freed memory.
	parent *Image

	// views caches the sub-region images clip() hands out. Spritesheet code
	// clips the same handful of regions out of the same sheet on every frame,
	// so without this a tilemap allocates a fresh image and quad per tile per
	// frame purely to throw them away again.
	views map[viewKey]*Image
}

// viewKey identifies a clipped region of an image.
type viewKey struct {
	x      int32
	y      int32
	width  int32
	height int32
}

// viewCacheMaxEntries bounds the clip cache. A tileset has a fixed, small set
// of regions, so the bound is only ever reached by a game clipping arbitrary
// rectangles — which is exactly the case that should not be cached at all.
const viewCacheMaxEntries = 4096

// NewImage loads an image from disk and uploads it to the GPU. Loading a path
// that is already in memory hands back the image already loaded rather than
// decoding the file and uploading a second copy of it to the GPU.
//
// The image is shared, not copied, so a game that loads the same sheet from two
// places gets one texture and one set of pixels — and a setFilter() call from
// either place is seen by both, which is the same thing that happens when a
// game passes one loaded image around itself.
func NewImage(path string) (*Image, error) {
	if cached, ok := Lumen.images[path]; ok {
		return cached, nil
	}

	image := &Image{Path: path, owned: true}

	surface, err := loadSurface(path)

	if err != nil {
		return nil, err
	}

	texture, err := Lumen.Renderer.CreateTextureFromSurface(surface)

	if err != nil {
		surface.Free()

		return nil, err
	}

	image.Surface = surface
	image.Texture = texture
	image.Width = surface.W
	image.Height = surface.H

	if Lumen.images == nil {
		Lumen.images = make(map[string]*Image)
	}

	Lumen.images[path] = image

	Lumen.RegisterResource(image)

	return image, nil
}

// View returns a non-owning image that draws a sub-region of this image.
func (image *Image) View(quad *Quad) *Image {
	owner := image

	if image.parent != nil {
		owner = image.parent
	}

	return &Image{
		Width:   quad.Width,
		Height:  quad.Height,
		Path:    image.Path,
		Quad:    quad,
		Surface: image.Surface,
		parent:  owner,
	}
}

// view returns a cached non-owning image for a sub-region, creating it on first
// use. Views hold no resources of their own and are never mutated after they
// are built, so handing the same one out repeatedly is safe.
func (image *Image) view(x, y, width, height int32) *Image {
	key := viewKey{x: x, y: y, width: width, height: height}

	if cached, ok := image.views[key]; ok {
		return cached
	}

	view := image.View(NewQuad(x, y, width, height))

	if image.views == nil {
		image.views = make(map[viewKey]*Image)
	}

	if len(image.views) < viewCacheMaxEntries {
		image.views[key] = view
	}

	return view
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
		return image.draw(args)
	case "drawQuad":
		return image.drawQuad(args)
	case "clip":
		return image.clip(args)
	case "getWidth":
		return object.NewInt(int64(image.Width)), true
	case "getHeight":
		return object.NewInt(int64(image.Height)), true
	case "getDimensions":
		return &object.List{Elements: []object.Object{
			object.NewInt(int64(image.Width)),
			object.NewInt(int64(image.Height)),
		}}, true
	case "getPixel":
		return image.getPixel(args)
	case "setFilter":
		return image.setFilter(args)
	case "toString":
		return &object.String{Value: image.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// draw renders the image at a position, optionally rotated, scaled, and offset
// about an origin: image.draw(x, y, rotation, scaleX, scaleY, originX, originY).
func (image *Image) draw(args []object.Object) (object.Object, bool) {
	arguments, err := ParseDrawArguments("image.draw", args, 0)

	if err != nil {
		return err, true
	}

	source := image.source()

	Lumen.DrawTexture(image.texture(), source, image.textureWidth(), image.textureHeight(),
		arguments.X, arguments.Y, arguments.Rotation,
		arguments.ScaleX, arguments.ScaleY, arguments.OriginX, arguments.OriginY)

	return value.NULL, true
}

// drawQuad renders a single region of the image, which is how spritesheets and
// tilesets are drawn: image.drawQuad(quad, x, y, rotation, sx, sy, ox, oy).
func (image *Image) drawQuad(args []object.Object) (object.Object, bool) {
	if len(args) < 3 {
		return object.NewError("image.drawQuad() expects at least a quad, x, and y. got=%d", len(args)), true
	}

	quad, ok := args[0].(*Quad)

	if !ok {
		return object.NewError("image.drawQuad() expects a quad as its first argument. got=%s", args[0].Type()), true
	}

	arguments, err := ParseDrawArguments("image.drawQuad", args, 1)

	if err != nil {
		return err, true
	}

	Lumen.DrawTexture(image.texture(), quad.Rect(), image.textureWidth(), image.textureHeight(),
		arguments.X, arguments.Y, arguments.Rotation,
		arguments.ScaleX, arguments.ScaleY, arguments.OriginX, arguments.OriginY)

	return value.NULL, true
}

// clip returns a view onto a rectangular region of the image. A third argument
// alone gives a square region, which keeps the original tile-sized form working:
// image.clip(x, y, size) and image.clip(x, y, width, height) are both valid.
func (image *Image) clip(args []object.Object) (object.Object, bool) {
	if len(args) != 3 && len(args) != 4 {
		return object.NewError("image.clip() expects 3 or 4 arguments. got=%d", len(args)), true
	}

	values := make([]int32, 0, 4)

	for index, arg := range args {
		number, ok := arg.(*object.Number)

		if !ok {
			return object.NewError("image.clip() expects number arguments. argument %d is %s", index+1, arg.Type()), true
		}

		values = append(values, int32(number.Int64()))
	}

	width := values[2]
	height := values[2]

	if len(values) == 4 {
		height = values[3]
	}

	return image.view(values[0], values[1], width, height), true
}

// getPixel reads the color of a single pixel from the image's source surface.
// Games use it to bake collision or spawn data straight into a map image.
func (image *Image) getPixel(args []object.Object) (object.Object, bool) {
	if len(args) != 2 {
		return object.NewError("image.getPixel() expects 2 arguments. got=%d", len(args)), true
	}

	x, err := Float("image.getPixel", args, 0)

	if err != nil {
		return err, true
	}

	y, err := Float("image.getPixel", args, 1)

	if err != nil {
		return err, true
	}

	if image.Surface == nil {
		return object.NewError("image.getPixel() is not available for this image"), true
	}

	offsetX := int32(x)
	offsetY := int32(y)

	if image.Quad != nil {
		offsetX += image.Quad.X
		offsetY += image.Quad.Y
	}

	if offsetX < 0 || offsetY < 0 || offsetX >= image.Surface.W || offsetY >= image.Surface.H {
		return object.NewError("image.getPixel() coordinates are outside the image"), true
	}

	pixel := image.Surface.At(int(offsetX), int(offsetY))
	red, green, blue, alpha := pixel.RGBA()

	return NewColor(uint8(red>>8), uint8(green>>8), uint8(blue>>8), uint8(alpha>>8)), true
}

// setFilter chooses between nearest-neighbour and linear sampling. Pixel art
// wants 'nearest', which is also Lumen's default.
func (image *Image) setFilter(args []object.Object) (object.Object, bool) {
	if len(args) != 1 {
		return object.NewError("image.setFilter() expects 1 argument. got=%d", len(args)), true
	}

	name, ok := args[0].(*object.String)

	if !ok {
		return object.NewError("image.setFilter() expects a string. got=%s", args[0].Type()), true
	}

	var quality string

	switch name.Value {
	case "nearest":
		quality = "0"
	case "linear":
		quality = "1"
	default:
		return object.NewError("image.setFilter() expects 'nearest' or 'linear'. got=%s", name.Value), true
	}

	// The scaling hint is read when a texture is created, so the texture has to
	// be rebuilt from its surface for a filter change to take effect.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, quality)

	if image.Surface != nil && image.owned {
		Lumen.FlushTexture(image.Texture)

		texture, err := Lumen.Renderer.CreateTextureFromSurface(image.Surface)

		if err == nil {
			image.Texture.Destroy()
			image.Texture = texture
		}
	}

	return value.NULL, true
}

// =============================================================================
// Helper methods

// texture returns the texture this image draws from, which for a clipped view
// is the texture of the image it was clipped from.
func (image *Image) texture() *sdl.Texture {
	if image.parent != nil {
		return image.parent.Texture
	}

	return image.Texture
}

// source returns the sub-rectangle this image draws, or nil for a whole texture.
func (image *Image) source() *sdl.Rect {
	if image.Quad == nil {
		return nil
	}

	return image.Quad.Rect()
}

// textureWidth returns the width of the backing texture, which differs from the
// image's own width when the image is a clipped view.
func (image *Image) textureWidth() int32 {
	if image.Surface != nil {
		return image.Surface.W
	}

	return image.Width
}

// textureHeight returns the height of the backing texture.
func (image *Image) textureHeight() int32 {
	if image.Surface != nil {
		return image.Surface.H
	}

	return image.Height
}

// Release frees the image's GPU and CPU memory. Views share their parent's
// texture and are left alone.
func (image *Image) Release() {
	if !image.owned {
		return
	}

	Lumen.FlushTexture(image.Texture)

	if image.Texture != nil {
		image.Texture.Destroy()
		image.Texture = nil
	}

	if image.Surface != nil {
		image.Surface.Free()
		image.Surface = nil
	}
}
