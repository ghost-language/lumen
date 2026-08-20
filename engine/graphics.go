package engine

import (
	"math"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// Graphics holds the mutable drawing state the canvas module manipulates. It is
// the Lumen equivalent of LOVE's graphics state: a current color, stroke width,
// blend mode, scissor box, render target, and a stack of affine transforms.
type Graphics struct {
	Color           *Color
	BackgroundColor *Color
	LineWidth       float64
	BlendMode       sdl.BlendMode
	PointSize       float64
	Transforms      []Transform
	Target          *Target
	Scissor         *sdl.Rect
	stateStack      []graphicsState
}

// graphicsState is the snapshot canvas.push('all') saves and canvas.pop()
// restores.
type graphicsState struct {
	color     *Color
	lineWidth float64
	blendMode sdl.BlendMode
	pointSize float64
	scissor   *sdl.Rect
	full      bool
}

// NewGraphics returns the default drawing state: opaque white, hairline strokes,
// alpha blending, and an identity transform.
func NewGraphics() *Graphics {
	return &Graphics{
		Color:           NewColor(255, 255, 255, 255),
		BackgroundColor: NewColor(0, 0, 0, 255),
		LineWidth:       1,
		PointSize:       1,
		BlendMode:       sdl.BLENDMODE_BLEND,
		Transforms:      []Transform{IdentityTransform()},
	}
}

// Transform returns the transform currently on top of the stack.
func (graphics *Graphics) Transform() Transform {
	return graphics.Transforms[len(graphics.Transforms)-1]
}

// SetTransform replaces the transform on top of the stack.
func (graphics *Graphics) SetTransform(transform Transform) {
	graphics.Transforms[len(graphics.Transforms)-1] = transform
}

// Push saves the current transform, and optionally the rest of the draw state.
func (graphics *Graphics) Push(all bool) {
	graphics.Transforms = append(graphics.Transforms, graphics.Transform())

	state := graphicsState{full: all}

	if all {
		state.color = graphics.Color
		state.lineWidth = graphics.LineWidth
		state.blendMode = graphics.BlendMode
		state.pointSize = graphics.PointSize
		state.scissor = graphics.Scissor
	}

	graphics.stateStack = append(graphics.stateStack, state)
}

// Pop restores the transform (and draw state, if it was pushed with 'all')
// saved by the matching Push. Popping the last transform is a no-op so a
// mismatched pop cannot corrupt the stack.
func (graphics *Graphics) Pop() {
	if len(graphics.Transforms) <= 1 {
		return
	}

	graphics.Transforms = graphics.Transforms[:len(graphics.Transforms)-1]

	if len(graphics.stateStack) == 0 {
		return
	}

	state := graphics.stateStack[len(graphics.stateStack)-1]
	graphics.stateStack = graphics.stateStack[:len(graphics.stateStack)-1]

	if !state.full {
		return
	}

	graphics.Color = state.color
	graphics.LineWidth = state.lineWidth
	graphics.BlendMode = state.blendMode
	graphics.PointSize = state.pointSize
	graphics.Scissor = state.scissor
}

// Reset returns the draw state to its frame-start defaults, so every frame draws
// the same way regardless of what the previous one left behind. LOVE only resets
// the transform and carries color, stroke width, and blend mode across frames,
// which is a well-worn source of "why is the whole world tinted" bugs: a fade
// drawn at the end of one frame silently tints the next. Lumen resets those too.
//
// The current font is deliberately left alone. Choosing a font is a decision a
// game makes once, usually in load(), not something it re-states every frame.
func (graphics *Graphics) Reset() {
	graphics.Transforms = graphics.Transforms[:0]
	graphics.Transforms = append(graphics.Transforms, IdentityTransform())
	graphics.stateStack = graphics.stateStack[:0]

	graphics.Color = NewColor(255, 255, 255, 255)
	graphics.LineWidth = 1
	graphics.PointSize = 1
	graphics.BlendMode = sdl.BLENDMODE_BLEND
	graphics.Scissor = nil
}

// =============================================================================
// Engine drawing helpers

// applyDrawState pushes the current color, blend mode, and scissor box down into
// SDL. Every draw call goes through it so state changes made from Ghost take
// effect without the module layer having to touch SDL directly.
func (engine *Engine) applyDrawState() {
	color := engine.Graphics.Color

	engine.Renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	engine.Renderer.SetDrawBlendMode(engine.Graphics.BlendMode)
	engine.Renderer.SetClipRect(engine.Graphics.Scissor)
}

// transformPoints maps a flat list of x/y pairs through the current transform.
func (engine *Engine) transformPoints(points []float64) []sdl.FPoint {
	transform := engine.Graphics.Transform()
	transformed := make([]sdl.FPoint, 0, len(points)/2)

	for index := 0; index+1 < len(points); index += 2 {
		x, y := transform.Apply(points[index], points[index+1])
		transformed = append(transformed, sdl.FPoint{X: float32(x), Y: float32(y)})
	}

	return transformed
}

// DrawPoints draws each x/y pair as a point, honouring the current point size.
func (engine *Engine) DrawPoints(points []float64) {
	engine.applyDrawState()

	transformed := engine.transformPoints(points)
	size := engine.Graphics.PointSize * engine.Graphics.Transform().ApproximateScale()

	if size <= 1.5 {
		engine.Renderer.DrawPointsF(transformed)

		return
	}

	half := float32(size / 2)

	for _, point := range transformed {
		engine.Renderer.FillRectF(&sdl.FRect{X: point.X - half, Y: point.Y - half, W: float32(size), H: float32(size)})
	}
}

// DrawPolyline strokes a path through the given x/y pairs. Strokes wider than a
// hairline are expanded into quads so the width survives scaling and rotation.
func (engine *Engine) DrawPolyline(points []float64, closed bool) {
	if len(points) < 4 {
		return
	}

	engine.applyDrawState()

	transformed := engine.transformPoints(points)

	if closed && len(transformed) > 2 {
		transformed = append(transformed, transformed[0])
	}

	width := engine.Graphics.LineWidth * engine.Graphics.Transform().ApproximateScale()

	if width <= 1.5 {
		engine.Renderer.DrawLinesF(transformed)

		return
	}

	half := width / 2

	for index := 0; index+1 < len(transformed); index++ {
		start := transformed[index]
		end := transformed[index+1]

		dx := float64(end.X - start.X)
		dy := float64(end.Y - start.Y)
		length := math.Hypot(dx, dy)

		if length == 0 {
			continue
		}

		// Offset perpendicular to the segment to give the stroke its width.
		nx := float32(-dy / length * half)
		ny := float32(dx / length * half)

		engine.fillTriangleFan([]sdl.FPoint{
			{X: start.X + nx, Y: start.Y + ny},
			{X: end.X + nx, Y: end.Y + ny},
			{X: end.X - nx, Y: end.Y - ny},
			{X: start.X - nx, Y: start.Y - ny},
		})
	}

	// Square off the joints so wide strokes do not show gaps at corners.
	if len(transformed) > 2 {
		joints := transformed[1 : len(transformed)-1]

		for _, joint := range joints {
			engine.fillTriangleFan([]sdl.FPoint{
				{X: joint.X - float32(half), Y: joint.Y - float32(half)},
				{X: joint.X + float32(half), Y: joint.Y - float32(half)},
				{X: joint.X + float32(half), Y: joint.Y + float32(half)},
				{X: joint.X - float32(half), Y: joint.Y + float32(half)},
			})
		}
	}
}

// FillPolygon fills the convex polygon described by the given x/y pairs.
func (engine *Engine) FillPolygon(points []float64) {
	if len(points) < 6 {
		return
	}

	engine.applyDrawState()
	engine.fillTriangleFan(engine.transformPoints(points))
}

// fillTriangleFan rasterises an already-transformed convex outline as a fan of
// triangles tinted with the current draw color.
func (engine *Engine) fillTriangleFan(points []sdl.FPoint) {
	if len(points) < 3 {
		return
	}

	color := engine.Graphics.Color.SDL()
	vertices := make([]sdl.Vertex, len(points))

	for index, point := range points {
		vertices[index] = sdl.Vertex{Position: point, Color: color}
	}

	indices := make([]int32, 0, (len(points)-2)*3)

	for index := 1; index < len(points)-1; index++ {
		indices = append(indices, 0, int32(index), int32(index+1))
	}

	engine.Renderer.RenderGeometry(nil, vertices, indices)
}

// EllipsePoints tessellates an ellipse into a closed outline. The segment count
// scales with the on-screen radius so circles stay smooth when zoomed in.
func (engine *Engine) EllipsePoints(x, y, radiusX, radiusY float64, segments int) []float64 {
	if segments <= 0 {
		scale := engine.Graphics.Transform().ApproximateScale()
		radius := math.Max(math.Abs(radiusX), math.Abs(radiusY)) * scale
		segments = int(math.Ceil(4 * math.Sqrt(math.Max(radius, 1))))
	}

	if segments < 8 {
		segments = 8
	}

	if segments > 512 {
		segments = 512
	}

	points := make([]float64, 0, segments*2)

	for index := 0; index < segments; index++ {
		angle := 2 * math.Pi * float64(index) / float64(segments)
		sin, cos := math.Sincos(angle)

		points = append(points, x+cos*radiusX, y+sin*radiusY)
	}

	return points
}

// ArcPoints tessellates a circular arc between two angles. When pie is true the
// centre point is included so the arc closes into a wedge.
func (engine *Engine) ArcPoints(x, y, radius, start, end float64, segments int, pie bool) []float64 {
	if segments <= 0 {
		scale := engine.Graphics.Transform().ApproximateScale()
		sweep := math.Abs(end-start) / (2 * math.Pi)
		segments = int(math.Ceil(4 * math.Sqrt(math.Max(radius*scale, 1)) * math.Max(sweep, 0.05)))
	}

	if segments < 3 {
		segments = 3
	}

	if segments > 512 {
		segments = 512
	}

	points := make([]float64, 0, (segments+2)*2)

	if pie {
		points = append(points, x, y)
	}

	for index := 0; index <= segments; index++ {
		angle := start + (end-start)*float64(index)/float64(segments)
		sin, cos := math.Sincos(angle)

		points = append(points, x+cos*radius, y+sin*radius)
	}

	return points
}

// DrawTexture draws a texture through the current transform with LOVE's draw
// arguments: position, rotation in radians, per-axis scale, and an origin offset
// applied before rotation and scaling. Passing the corners through the transform
// (rather than leaning on SDL's rotated blit) keeps camera transforms, sprite
// rotation, and negative scale flips composing correctly.
func (engine *Engine) DrawTexture(texture *sdl.Texture, source *sdl.Rect, textureWidth, textureHeight int32, x, y, rotation, scaleX, scaleY, originX, originY float64) {
	if texture == nil || textureWidth == 0 || textureHeight == 0 {
		return
	}

	width := float64(textureWidth)
	height := float64(textureHeight)

	if source != nil {
		width = float64(source.W)
		height = float64(source.H)
	}

	local := IdentityTransform().
		Translate(x, y).
		Rotate(rotation).
		Scale(scaleX, scaleY).
		Translate(-originX, -originY)

	transform := engine.Graphics.Transform().Multiply(local)

	corners := [4][2]float64{{0, 0}, {width, 0}, {width, height}, {0, height}}

	// Normalised texture coordinates for the source rectangle.
	u0, v0, u1, v1 := float32(0), float32(0), float32(1), float32(1)

	if source != nil {
		u0 = float32(source.X) / float32(textureWidth)
		v0 = float32(source.Y) / float32(textureHeight)
		u1 = float32(source.X+source.W) / float32(textureWidth)
		v1 = float32(source.Y+source.H) / float32(textureHeight)
	}

	coordinates := [4]sdl.FPoint{{X: u0, Y: v0}, {X: u1, Y: v0}, {X: u1, Y: v1}, {X: u0, Y: v1}}

	color := engine.Graphics.Color.SDL()
	vertices := make([]sdl.Vertex, 4)

	for index, corner := range corners {
		px, py := transform.Apply(corner[0], corner[1])

		vertices[index] = sdl.Vertex{
			Position: sdl.FPoint{X: float32(px), Y: float32(py)},
			Color:    color,
			TexCoord: coordinates[index],
		}
	}

	engine.Renderer.SetDrawBlendMode(engine.Graphics.BlendMode)
	engine.Renderer.SetClipRect(engine.Graphics.Scissor)
	texture.SetBlendMode(engine.Graphics.BlendMode)

	engine.Renderer.RenderGeometry(texture, vertices, []int32{0, 1, 2, 0, 2, 3})
}

// SetScissor limits drawing to the given rectangle, expressed in the coordinate
// space of the current transform. SDL clips against an axis-aligned rectangle,
// so a rotated scissor falls back to the bounding box of its transformed corners.
func (engine *Engine) SetScissor(x, y, width, height float64) {
	transform := engine.Graphics.Transform()

	corners := [4][2]float64{{x, y}, {x + width, y}, {x + width, y + height}, {x, y + height}}

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for _, corner := range corners {
		px, py := transform.Apply(corner[0], corner[1])

		minX = math.Min(minX, px)
		minY = math.Min(minY, py)
		maxX = math.Max(maxX, px)
		maxY = math.Max(maxY, py)
	}

	engine.Graphics.Scissor = &sdl.Rect{
		X: int32(math.Floor(minX)),
		Y: int32(math.Floor(minY)),
		W: int32(math.Ceil(maxX - minX)),
		H: int32(math.Ceil(maxY - minY)),
	}
}

// ClearScissor removes any active scissor rectangle.
func (engine *Engine) ClearScissor() {
	engine.Graphics.Scissor = nil

	engine.Renderer.SetClipRect(nil)
}

// BlendModeFromName maps LOVE's blend mode names onto SDL blend modes.
func BlendModeFromName(name string) (sdl.BlendMode, bool) {
	switch name {
	case "alpha":
		return sdl.BLENDMODE_BLEND, true
	case "add", "additive":
		return sdl.BLENDMODE_ADD, true
	case "multiply":
		return sdl.BLENDMODE_MOD, true
	case "none", "replace":
		return sdl.BLENDMODE_NONE, true
	}

	return sdl.BLENDMODE_BLEND, false
}

// Screenshot writes the current contents of the window to a PNG file. It reads
// back from the renderer, so it captures exactly what the player sees.
func (engine *Engine) Screenshot(path string) error {
	width, height, err := engine.Renderer.GetOutputSize()

	if err != nil {
		return err
	}

	surface, err := sdl.CreateRGBSurfaceWithFormat(0, width, height, 32, sdl.PIXELFORMAT_ARGB8888)

	if err != nil {
		return err
	}

	defer surface.Free()

	if err := engine.Renderer.ReadPixels(nil, sdl.PIXELFORMAT_ARGB8888, surface.Data(), int(surface.Pitch)); err != nil {
		return err
	}

	return img.SavePNG(surface, path)
}
