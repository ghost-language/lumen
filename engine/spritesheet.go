package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
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
func (sheet *Spritesheet) Method(method string, tok token.Token, args []object.Object) (object.Object, bool) {
	switch method {
	case "draw":
		return sheet.draw(tok, args)
	case "newAnimation":
		return sheet.newAnimation(tok, args)
	case "getQuad":
		return sheet.getQuad(tok, args)
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
func (sheet *Spritesheet) draw(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := ArityAtLeast("spritesheet.draw", tok, args, 3); err != nil {
		return err, true
	}

	frame, err := Number("spritesheet.draw", tok, args, 0)

	if err != nil {
		return err, true
	}

	if sheet.Quad(int32(frame)) == nil {
		return sheet.outsideSheet("spritesheet.draw", tok, int32(frame)), true
	}

	arguments, parseErr := ParseDrawArguments("spritesheet.draw", tok, args, 1)

	if parseErr != nil {
		return parseErr, true
	}

	sheet.Draw(int32(frame), arguments)

	return value.NULL, true
}

// outsideSheet reports a frame number the sheet does not have. Drawing one used
// to do nothing at all, which is the worst of the three possible answers: the
// game carries on with a hole where a sprite should be, and the reason is
// somewhere in whatever arithmetic produced the number.
func (sheet *Spritesheet) outsideSheet(name string, tok token.Token, frame int32) *object.Error {
	return Error(fault.Index, tok, "`%s` was given frame %d, and the sheet has %d", Signature(name), frame, sheet.Count()).
		WithHelp("frames are numbered from 0 to %d, left to right and top to bottom", sheet.Count()-1)
}

// getQuad returns the region of the image a frame occupies.
func (sheet *Spritesheet) getQuad(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("spritesheet.getQuad", tok, args, 1); err != nil {
		return err, true
	}

	frame, err := Number("spritesheet.getQuad", tok, args, 0)

	if err != nil {
		return err, true
	}

	quad := sheet.Quad(int32(frame))

	if quad == nil {
		return sheet.outsideSheet("spritesheet.getQuad", tok, int32(frame)), true
	}

	return quad, true
}

// newAnimation builds an animation over a list of this sheet's frames:
// sheet.newAnimation([0, 1, 2], 0.1, 'loop').
func (sheet *Spritesheet) newAnimation(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := ArityRange("spritesheet.newAnimation", tok, args, 2, 3); err != nil {
		return err, true
	}

	list, err := List("spritesheet.newAnimation", tok, args, 0)

	if err != nil {
		return err, true
	}

	frames := make([]int32, 0, len(list.Elements))

	for index, element := range list.Elements {
		number, ok := element.(*object.Number)

		if !ok {
			return Error(fault.Argument, tok, "`%s` expects a list of frame numbers, and element %d is %s", Signature("spritesheet.newAnimation"), index+1, TypeName(element)), true
		}

		frame := int32(number.Int64())

		if frame < 0 || frame >= sheet.Count() {
			return sheet.outsideSheet("spritesheet.newAnimation", tok, frame), true
		}

		frames = append(frames, frame)
	}

	if len(frames) == 0 {
		return Value("spritesheet.newAnimation", tok, "was given no frames to play").
			WithHelp("an animation needs at least one frame, as in `sheet.newAnimation([0, 1, 2], 0.1)`"), true
	}

	duration, err := Number("spritesheet.newAnimation", tok, args, 1)

	if err != nil {
		return err, true
	}

	// Zero is deliberately allowed: an animation of one frame held forever is
	// how a still pose is written, and Frame() reads a duration of zero as
	// exactly that. A negative one is not a pose, it is a mistake.
	if duration < 0 {
		return Value("spritesheet.newAnimation", tok, "was given a frame duration of %g", duration).
			WithHelp("a frame is held for a number of seconds; use 0 for a pose that never advances"), true
	}

	mode := AnimationLoop

	if len(args) == 3 {
		name, err := Text("spritesheet.newAnimation", tok, args, 2)

		if err != nil {
			return err, true
		}

		parsed, valid := AnimationModeFromName(name)

		if !valid {
			return Choice("spritesheet.newAnimation", tok, name, AnimationModeNames...), true
		}

		mode = parsed
	}

	return NewAnimation(sheet, frames, duration, mode), true
}
