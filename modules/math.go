package modules

import (
	"math"
	"math/rand"

	ghostmodules "ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
)

// Ghost's own math module covers trigonometry and a few predicates, but a game
// needs square roots, rounding, angles between points, and random numbers in a
// range on nearly every frame. Rather than introduce a second, competing module,
// Lumen adds these to the math module games already reach for. Ghost's existing
// methods are untouched.

// generator is Lumen's random source. It is kept separate from the global one so
// that seeding it for reproducible level generation cannot be disturbed by
// anything else in the process.
var generator = rand.New(rand.NewSource(1))

func registerMathExtensions() {
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "floor", mathFloorMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "ceil", mathCeilMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "round", mathRoundMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "sqrt", mathSqrtMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "pow", mathPowMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "exp", mathExpMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "log", mathLogMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "sign", mathSignMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "asin", mathAsinMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "acos", mathAcosMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "atan", mathAtanMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "atan2", mathAtan2Method)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "degrees", mathDegreesMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "radians", mathRadiansMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "clamp", mathClampMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "lerp", mathLerpMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "distance", mathDistanceMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "angle", mathAngleMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "random", mathRandomMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "randomSeed", mathRandomSeedMethod)
	ghostmodules.RegisterMethod(ghostmodules.MathMethods, "noise", mathNoiseMethod)
}

func mathFloorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unaryInteger("math.floor", tok, args, math.Floor)
}

func mathCeilMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unaryInteger("math.ceil", tok, args, math.Ceil)
}

func mathRoundMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unaryInteger("math.round", tok, args, math.Round)
}

func mathSqrtMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.sqrt", tok, args, math.Sqrt)
}

func mathExpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.exp", tok, args, math.Exp)
}

func mathLogMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.log", tok, args, math.Log)
}

func mathAsinMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.asin", tok, args, math.Asin)
}

func mathAcosMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.acos", tok, args, math.Acos)
}

func mathAtanMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.atan", tok, args, math.Atan)
}

func mathDegreesMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.degrees", tok, args, func(radians float64) float64 {
		return radians * 180 / math.Pi
	})
}

func mathRadiansMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return unary("math.radians", tok, args, func(degrees float64) float64 {
		return degrees * math.Pi / 180
	})
}

// mathSignMethod returns -1, 0, or 1, which turns "which way is the enemy" into
// a single multiplication.
func mathSignMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	given, err := number("math.sign", tok, args, 0)

	if err != nil {
		return err
	}

	if given > 0 {
		return object.NewInt(1)
	}

	if given < 0 {
		return object.NewInt(-1)
	}

	return object.NewInt(0)
}

func mathPowMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.pow", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("math.pow", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(math.Pow(values[0], values[1]))
}

// mathAtan2Method returns the angle of a vector, handling every quadrant. It is
// how a game points a sprite or a projectile at a target.
func mathAtan2Method(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.atan2", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("math.atan2", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(math.Atan2(values[0], values[1]))
}

// mathClampMethod keeps a value inside a range, which is most of what keeping a
// camera inside a map or a health bar inside its bounds amounts to.
func mathClampMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.clamp", tok, args, 3); err != nil {
		return err
	}

	values, err := numbers("math.clamp", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(math.Max(values[1], math.Min(values[2], values[0])))
}

// mathLerpMethod blends between two values. Camera smoothing, fades, and health
// bars that slide rather than jump are all one lerp per frame.
func mathLerpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.lerp", tok, args, 3); err != nil {
		return err
	}

	values, err := numbers("math.lerp", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(values[0] + (values[1]-values[0])*values[2])
}

func mathDistanceMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.distance", tok, args, 4); err != nil {
		return err
	}

	values, err := numbers("math.distance", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(math.Hypot(values[2]-values[0], values[3]-values[1]))
}

// mathAngleMethod returns the angle in radians from one point to another.
func mathAngleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.angle", tok, args, 4); err != nil {
		return err
	}

	values, err := numbers("math.angle", tok, args)

	if err != nil {
		return err
	}

	return object.NewFloat(math.Atan2(values[3]-values[1], values[2]-values[0]))
}

// mathRandomMethod mirrors LOVE's love.math.random: no arguments give a float in
// [0, 1), one argument gives a whole number in [1, n], and two give a whole
// number in [low, high].
func mathRandomMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("math.random", tok, args, 0, 2); err != nil {
		return err
	}

	if len(args) == 0 {
		return object.NewFloat(generator.Float64())
	}

	values, err := numbers("math.random", tok, args)

	if err != nil {
		return err
	}

	low := int64(1)
	high := int64(values[0])

	if len(values) == 2 {
		low = int64(values[0])
		high = int64(values[1])
	}

	if high < low {
		return object.NewError("%d:%d: runtime error: math.random() expects the upper bound to be at least the lower bound", tok.Line, tok.Column)
	}

	return object.NewInt(low + generator.Int63n(high-low+1))
}

// mathRandomSeedMethod seeds the generator so a run can be reproduced, which is
// what procedural maps need to be shareable.
func mathRandomSeedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("math.randomSeed", tok, args, 1); err != nil {
		return err
	}

	seed, err := integer("math.randomSeed", tok, args, 0)

	if err != nil {
		return err
	}

	generator = rand.New(rand.NewSource(seed))

	return value.NULL
}

// mathNoiseMethod is value noise smoothed with a cubic curve: continuous,
// repeatable for the same input, and in the 0-1 range. It suits terrain heights,
// torch flicker, and drifting movement, where truly random values would jitter.
func mathNoiseMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("math.noise", tok, args, 1, 2); err != nil {
		return err
	}

	values, err := numbers("math.noise", tok, args)

	if err != nil {
		return err
	}

	y := 0.0

	if len(values) == 2 {
		y = values[1]
	}

	return object.NewFloat(noise2D(values[0], y))
}

func unary(name string, tok token.Token, args []object.Object, operation func(float64) float64) object.Object {
	if err := arity(name, tok, args, 1); err != nil {
		return err
	}

	given, err := number(name, tok, args, 0)

	if err != nil {
		return err
	}

	return object.NewFloat(operation(given))
}

// unaryInteger is for the rounding operations, whose results should come back as
// whole numbers so they can index a list without a further conversion.
func unaryInteger(name string, tok token.Token, args []object.Object, operation func(float64) float64) object.Object {
	if err := arity(name, tok, args, 1); err != nil {
		return err
	}

	given, err := number(name, tok, args, 0)

	if err != nil {
		return err
	}

	return object.NewInt(int64(operation(given)))
}

// noise2D samples smoothed value noise at a point.
func noise2D(x, y float64) float64 {
	cellX := math.Floor(x)
	cellY := math.Floor(y)

	fractionX := smoothstep(x - cellX)
	fractionY := smoothstep(y - cellY)

	topLeft := latticeValue(int64(cellX), int64(cellY))
	topRight := latticeValue(int64(cellX)+1, int64(cellY))
	bottomLeft := latticeValue(int64(cellX), int64(cellY)+1)
	bottomRight := latticeValue(int64(cellX)+1, int64(cellY)+1)

	top := topLeft + (topRight-topLeft)*fractionX
	bottom := bottomLeft + (bottomRight-bottomLeft)*fractionX

	return top + (bottom-top)*fractionY
}

// smoothstep eases the interpolation between lattice points so the noise has no
// visible creases along cell boundaries.
func smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}

// latticeValue hashes integer coordinates into a repeatable value in [0, 1).
func latticeValue(x, y int64) float64 {
	hash := uint64(x)*0x9E3779B97F4A7C15 ^ uint64(y)*0xC2B2AE3D27D4EB4F

	hash ^= hash >> 33
	hash *= 0xFF51AFD7ED558CCD
	hash ^= hash >> 33

	return float64(hash>>11) / float64(uint64(1)<<53)
}
