package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
)

// Spritesheet slices one image into a grid of equally sized frames, numbered
// left to right and top to bottom starting at zero.
//
// The quads are built once, at construction, and reused for the life of the
// sheet. Building them per draw is the shape this replaces: three of Lumen's
// own examples each grew their own version of this class, and two of them
// allocated a quad per tile per frame to do it.
type Spritesheet struct {
	Image       *Image
	FrameWidth  int32
	FrameHeight int32
	Columns     int32
	Rows        int32

	quads []*Quad
}

// NewSpritesheet slices an image into frames of the given size. A sheet whose
// image is not a whole number of frames across keeps only the complete ones.
func NewSpritesheet(image *Image, frameWidth, frameHeight int32) (*Spritesheet, error) {
	if image == nil {
		return nil, fmt.Errorf("a spritesheet needs an image")
	}

	if frameWidth <= 0 || frameHeight <= 0 {
		return nil, fmt.Errorf("frame size must be positive, got %dx%d", frameWidth, frameHeight)
	}

	columns := image.Width / frameWidth
	rows := image.Height / frameHeight

	if columns <= 0 || rows <= 0 {
		return nil, fmt.Errorf("a %dx%d frame does not fit in a %dx%d image", frameWidth, frameHeight, image.Width, image.Height)
	}

	sheet := &Spritesheet{
		Image:       image,
		FrameWidth:  frameWidth,
		FrameHeight: frameHeight,
		Columns:     columns,
		Rows:        rows,
		quads:       make([]*Quad, 0, columns*rows),
	}

	for index := int32(0); index < columns*rows; index++ {
		x := (index % columns) * frameWidth
		y := (index / columns) * frameHeight

		sheet.quads = append(sheet.quads, NewQuad(x, y, frameWidth, frameHeight))
	}

	return sheet, nil
}

// Count returns how many frames the sheet holds.
func (sheet *Spritesheet) Count() int32 {
	return int32(len(sheet.quads))
}

// Quad returns the region of the image a frame occupies, or nil when the frame
// is outside the sheet.
func (sheet *Spritesheet) Quad(frame int32) *Quad {
	if frame < 0 || frame >= sheet.Count() {
		return nil
	}

	return sheet.quads[frame]
}

// Draw renders one frame through the current transform.
func (sheet *Spritesheet) Draw(frame int32, arguments DrawArguments) {
	quad := sheet.Quad(frame)

	if quad == nil {
		return
	}

	Lumen.DrawTexture(sheet.Image.texture(), quad.Rect(), sheet.Image.textureWidth(), sheet.Image.textureHeight(),
		arguments.X, arguments.Y, arguments.Rotation,
		arguments.ScaleX, arguments.ScaleY, arguments.OriginX, arguments.OriginY)
}

// String represents the spritesheet object's value as a string.
func (sheet *Spritesheet) String() string {
	return fmt.Sprintf("Spritesheet: %s, %dx%d frames of %dx%d", sheet.Image.Path, sheet.Columns, sheet.Rows, sheet.FrameWidth, sheet.FrameHeight)
}

// Type returns the spritesheet object type.
func (sheet *Spritesheet) Type() object.Type {
	return SPRITESHEET
}

// Method defines the set of methods available on spritesheet objects.
func (sheet *Spritesheet) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "draw":
		return sheet.draw(args)
	case "newAnimation":
		return sheet.newAnimation(args)
	case "getQuad":
		return sheet.getQuad(args)
	case "getImage":
		return sheet.Image, true
	case "getCount":
		return object.NewInt(int64(sheet.Count())), true
	case "getColumns":
		return object.NewInt(int64(sheet.Columns)), true
	case "getRows":
		return object.NewInt(int64(sheet.Rows)), true
	case "getFrameWidth":
		return object.NewInt(int64(sheet.FrameWidth)), true
	case "getFrameHeight":
		return object.NewInt(int64(sheet.FrameHeight)), true
	case "getFrameDimensions":
		return &object.List{Elements: []object.Object{
			object.NewInt(int64(sheet.FrameWidth)),
			object.NewInt(int64(sheet.FrameHeight)),
		}}, true
	case "toString":
		return &object.String{Value: sheet.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// draw renders a frame by index: sheet.draw(frame, x, y, rotation, sx, sy, ox, oy).
func (sheet *Spritesheet) draw(args []object.Object) (object.Object, bool) {
	if len(args) < 3 {
		return object.NewError("spritesheet.draw() expects at least a frame, x, and y. got=%d", len(args)), true
	}

	frame, err := Float("spritesheet.draw", args, 0)

	if err != nil {
		return err, true
	}

	arguments, parseErr := ParseDrawArguments("spritesheet.draw", args, 1)

	if parseErr != nil {
		return parseErr, true
	}

	sheet.Draw(int32(frame), arguments)

	return value.NULL, true
}

// getQuad returns the region of the image a frame occupies.
func (sheet *Spritesheet) getQuad(args []object.Object) (object.Object, bool) {
	if len(args) != 1 {
		return object.NewError("spritesheet.getQuad() expects 1 argument. got=%d", len(args)), true
	}

	frame, err := Float("spritesheet.getQuad", args, 0)

	if err != nil {
		return err, true
	}

	quad := sheet.Quad(int32(frame))

	if quad == nil {
		return object.NewError("spritesheet.getQuad() frame %d is outside a sheet of %d frames", int32(frame), sheet.Count()), true
	}

	return quad, true
}

// newAnimation builds an animation over a list of this sheet's frames:
// sheet.newAnimation([0, 1, 2], 0.1, 'loop').
func (sheet *Spritesheet) newAnimation(args []object.Object) (object.Object, bool) {
	if len(args) < 2 || len(args) > 3 {
		return object.NewError("spritesheet.newAnimation() expects 2 or 3 arguments. got=%d", len(args)), true
	}

	list, ok := args[0].(*object.List)

	if !ok {
		return object.NewError("spritesheet.newAnimation() expects a list of frames as its first argument. got=%s", args[0].Type()), true
	}

	frames := make([]int32, 0, len(list.Elements))

	for index, element := range list.Elements {
		number, ok := element.(*object.Number)

		if !ok {
			return object.NewError("spritesheet.newAnimation() expects a list of frame numbers. element %d is %s", index+1, element.Type()), true
		}

		frame := int32(number.Int64())

		if frame < 0 || frame >= sheet.Count() {
			return object.NewError("spritesheet.newAnimation() frame %d is outside a sheet of %d frames", frame, sheet.Count()), true
		}

		frames = append(frames, frame)
	}

	if len(frames) == 0 {
		return object.NewError("spritesheet.newAnimation() expects at least one frame"), true
	}

	duration, err := Float("spritesheet.newAnimation", args, 1)

	if err != nil {
		return err, true
	}

	mode := AnimationLoop

	if len(args) == 3 {
		name, ok := args[2].(*object.String)

		if !ok {
			return object.NewError("spritesheet.newAnimation() expects a mode name as its third argument. got=%s", args[2].Type()), true
		}

		parsed, valid := AnimationModeFromName(name.Value)

		if !valid {
			return object.NewError("spritesheet.newAnimation() expects 'loop', 'once', or 'pingpong'. got=%s", name.Value), true
		}

		mode = parsed
	}

	return NewAnimation(sheet, frames, duration, mode), true
}
