package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var TimerMethods = map[string]*object.LibraryFunction{}
var TimerProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(TimerMethods, "sleep", timerSleepMethod)
	modules.RegisterMethod(TimerMethods, "getDelta", timerGetDeltaMethod)
	modules.RegisterMethod(TimerMethods, "getFps", timerGetFpsMethod)
	modules.RegisterMethod(TimerMethods, "getTime", timerGetTimeMethod)

	modules.RegisterProperty(TimerProperties, "delta", timerDeltaProperty)
	modules.RegisterProperty(TimerProperties, "fps", timerFpsProperty)
	modules.RegisterProperty(TimerProperties, "time", timerTimeProperty)
	modules.RegisterProperty(TimerProperties, "averageDelta", timerAverageDeltaProperty)
	modules.RegisterProperty(TimerProperties, "frame", timerFrameProperty)
}

// timerSleepMethod blocks for a number of seconds. It is a blunt instrument that
// stalls the whole game, and exists for scripts and tools rather than gameplay.
func timerSleepMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("timer.sleep", tok, args, 1); err != nil {
		return err
	}

	seconds, err := number("timer.sleep", tok, args, 0)

	if err != nil {
		return err
	}

	sdl.Delay(uint32(seconds * 1000))

	return value.NULL
}

func timerGetDeltaMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return timerDeltaProperty(scope, tok)
}

func timerGetFpsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return timerFpsProperty(scope, tok)
}

func timerGetTimeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return timerTimeProperty(scope, tok)
}

// timerDeltaProperty is the seconds the previous frame took. Multiplying
// movement by it keeps a game running at the same speed on any machine, and it
// is the same value handed to update().
func timerDeltaProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewFloat(engine.Lumen.Delta)
}

func timerFpsProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(engine.Lumen.CurrentFps))
}

// timerTimeProperty is the seconds elapsed since the game started, useful for
// driving anything that should oscillate or cycle over time.
func timerTimeProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewFloat(engine.Lumen.ElapsedTime)
}

// timerAverageDeltaProperty is a smoothed delta, which is what should be shown
// to a player: the raw delta jitters too much to read.
func timerAverageDeltaProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewFloat(engine.Lumen.AverageDelta)
}

func timerFrameProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(engine.Lumen.FrameCount))
}
