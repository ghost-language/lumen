package engine

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// A game draws in a coordinate space, and by default that space is the window:
// a point at (400, 300) lands 400 pixels from the left edge of the window and
// 300 down from the top. Resizing the window, or going fullscreen, therefore
// does not make anything bigger — it shows more of the world around the same
// sized sprites, and leaves every interface element stranded in the corner it
// was placed in.
//
// A logical size fixes that. The game declares the size of the space it draws
// in, and Lumen scales that space to fill as much of the window as it can while
// keeping its shape, centring what is left over behind black bars. A 320x180
// game looks the same in a 640x360 window, a 1280x720 one, and fullscreen on a
// 4K display: the same amount of world, the same text, four times the pixels.
//
// The scaling is applied as the bottom transform on the graphics stack rather
// than through SDL's own logical size, so everything downstream — the camera a
// game pushes on top of it, scissor rectangles, and mouse coordinates — agrees
// about where a point is without having to know that any scaling is happening.

// SetLogicalSize fixes the size of the coordinate space the game draws in.
func (engine *Engine) SetLogicalSize(width, height int32) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("logical size must be positive, got %dx%d", width, height)
	}

	engine.LogicalWidth = width
	engine.LogicalHeight = height

	engine.UpdateViewport()

	return nil
}

// ClearLogicalSize goes back to drawing in window pixels.
func (engine *Engine) ClearLogicalSize() {
	engine.LogicalWidth = 0
	engine.LogicalHeight = 0

	engine.UpdateViewport()
}

// HasLogicalSize reports whether the game has fixed its coordinate space.
func (engine *Engine) HasLogicalSize() bool {
	return engine.LogicalWidth > 0 && engine.LogicalHeight > 0
}

// SetPixelPerfect limits scaling to whole numbers, so that every pixel of the
// game covers exactly the same square of screen pixels. It is what keeps pixel
// art from developing uneven, shimmering edges, and it costs a little more of
// the screen to the bars: at 1.8 times, pixel-perfect scaling draws at 1.
func (engine *Engine) SetPixelPerfect(enabled bool) {
	engine.PixelPerfect = enabled

	engine.UpdateViewport()
}

// DrawWidth is the width of the coordinate space the game draws in.
func (engine *Engine) DrawWidth() int32 {
	if engine.HasLogicalSize() {
		return engine.LogicalWidth
	}

	width, _ := engine.OutputSize()

	return width
}

// DrawHeight is the height of the coordinate space the game draws in.
func (engine *Engine) DrawHeight() int32 {
	if engine.HasLogicalSize() {
		return engine.LogicalHeight
	}

	_, height := engine.OutputSize()

	return height
}

// ViewportScale is how many screen pixels one unit of the game's coordinate
// space covers.
func (engine *Engine) ViewportScale() float64 {
	if engine.viewScale <= 0 {
		return 1
	}

	return engine.viewScale
}

// OutputSize is the renderer's size in real pixels, which is not always the
// window's size in the coordinates the operating system reports: a window on a
// high-density display is often half the size of the pixels behind it.
func (engine *Engine) OutputSize() (int32, int32) {
	if engine.Renderer != nil {
		if width, height, err := engine.Renderer.GetOutputSize(); err == nil && width > 0 && height > 0 {
			return width, height
		}
	}

	if engine.Window != nil {
		return engine.Window.GetSize()
	}

	return engine.Width, engine.Height
}

// UpdateViewport recomputes where the game's coordinate space lands in the
// window. It runs whenever the window changes size or the logical size does.
func (engine *Engine) UpdateViewport() {
	width, height := engine.OutputSize()

	engine.OutputWidth = width
	engine.OutputHeight = height

	if !engine.HasLogicalSize() {
		engine.viewScale = 1
		engine.viewport = sdl.Rect{X: 0, Y: 0, W: width, H: height}

		if engine.Graphics != nil {
			engine.Graphics.Base = IdentityTransform()
		}

		return
	}

	// The smaller of the two ratios is the one that fits: scaling by the larger
	// would push the other axis off the edge of the window.
	scale := math.Min(
		float64(width)/float64(engine.LogicalWidth),
		float64(height)/float64(engine.LogicalHeight),
	)

	// A window smaller than the game still has to show all of it, so scaling
	// down stays fractional however pixel-perfect the game asked to be.
	if engine.PixelPerfect && scale > 1 {
		scale = math.Floor(scale)
	}

	drawWidth := math.Floor(float64(engine.LogicalWidth) * scale)
	drawHeight := math.Floor(float64(engine.LogicalHeight) * scale)

	// Whole-pixel offsets. Half a pixel of slack would blur the whole frame.
	offsetX := math.Floor((float64(width) - drawWidth) / 2)
	offsetY := math.Floor((float64(height) - drawHeight) / 2)

	engine.viewScale = scale
	engine.viewport = sdl.Rect{X: int32(offsetX), Y: int32(offsetY), W: int32(drawWidth), H: int32(drawHeight)}

	if engine.Graphics != nil && engine.Graphics.Target == nil {
		engine.Graphics.Base = engine.ViewportTransform()
	}
}

// ViewportTransform maps the game's coordinate space onto the window: the scale
// that fits it, and the offset that centres it.
func (engine *Engine) ViewportTransform() Transform {
	if !engine.HasLogicalSize() {
		return IdentityTransform()
	}

	scale := engine.ViewportScale()

	return IdentityTransform().
		Translate(float64(engine.viewport.X), float64(engine.viewport.Y)).
		Scale(scale, scale)
}

// ToLogical maps a point in window coordinates — where SDL reports the mouse —
// into the coordinate space the game draws in.
func (engine *Engine) ToLogical(x, y int32) (float64, float64) {
	pixelX, pixelY := engine.toPixels(float64(x), float64(y))

	if !engine.HasLogicalSize() {
		return pixelX, pixelY
	}

	scale := engine.ViewportScale()

	return (pixelX - float64(engine.viewport.X)) / scale, (pixelY - float64(engine.viewport.Y)) / scale
}

// ToWindow maps a point in the game's coordinate space back to the window
// coordinates SDL expects when it is asked to move the pointer.
func (engine *Engine) ToWindow(x, y float64) (int32, int32) {
	if engine.HasLogicalSize() {
		scale := engine.ViewportScale()

		x = x*scale + float64(engine.viewport.X)
		y = y*scale + float64(engine.viewport.Y)
	}

	windowWidth, windowHeight := engine.Window.GetSize()

	if engine.OutputWidth > 0 && engine.OutputHeight > 0 {
		x = x * float64(windowWidth) / float64(engine.OutputWidth)
		y = y * float64(windowHeight) / float64(engine.OutputHeight)
	}

	return int32(math.Round(x)), int32(math.Round(y))
}

// MousePosition is the pointer in the game's coordinate space.
func (engine *Engine) MousePosition() (float64, float64) {
	return engine.ToLogical(engine.MouseX, engine.MouseY)
}

// MousePixels is the pointer in the renderer's pixels, which is the space the
// transform stack maps into and so the space an inverse transform starts from.
func (engine *Engine) MousePixels() (float64, float64) {
	return engine.toPixels(float64(engine.MouseX), float64(engine.MouseY))
}

// toPixels converts window coordinates to renderer pixels.
func (engine *Engine) toPixels(x, y float64) (float64, float64) {
	if engine.Window == nil {
		return x, y
	}

	windowWidth, windowHeight := engine.Window.GetSize()

	if windowWidth <= 0 || windowHeight <= 0 {
		return x, y
	}

	return x * float64(engine.OutputWidth) / float64(windowWidth),
		y * float64(engine.OutputHeight) / float64(windowHeight)
}

// drawLetterbox paints the bars around the game's coordinate space. A game is
// free to draw outside its own bounds — a fade covering the screen is drawn as
// one rectangle, and rounding leaves the odd half pixel at the edges — so the
// bars go down after the frame rather than being clipped out of it.
func (engine *Engine) drawLetterbox() {
	if !engine.HasLogicalSize() {
		return
	}

	view := engine.viewport
	width, height := engine.OutputWidth, engine.OutputHeight

	if view.X <= 0 && view.Y <= 0 && view.W >= width && view.H >= height {
		return
	}

	engine.Renderer.SetClipRect(nil)
	engine.Renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
	engine.Renderer.SetDrawColor(0, 0, 0, 255)

	right := view.X + view.W
	bottom := view.Y + view.H

	bars := []sdl.Rect{
		{X: 0, Y: 0, W: width, H: view.Y},
		{X: 0, Y: bottom, W: width, H: height - bottom},
		{X: 0, Y: view.Y, W: view.X, H: view.H},
		{X: right, Y: view.Y, W: width - right, H: view.H},
	}

	for index := range bars {
		if bars[index].W <= 0 || bars[index].H <= 0 {
			continue
		}

		engine.Renderer.FillRect(&bars[index])
	}
}
