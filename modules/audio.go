package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/mix"
)

var AudioMethods = map[string]*object.LibraryFunction{}
var AudioProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(AudioMethods, "newSource", audioNewSourceMethod)
	modules.RegisterMethod(AudioMethods, "play", audioPlayMethod)
	modules.RegisterMethod(AudioMethods, "stop", audioStopMethod)
	modules.RegisterMethod(AudioMethods, "pause", audioPauseMethod)
	modules.RegisterMethod(AudioMethods, "resume", audioResumeMethod)
	modules.RegisterMethod(AudioMethods, "setVolume", audioSetVolumeMethod)
	modules.RegisterMethod(AudioMethods, "getVolume", audioGetVolumeMethod)
}

// audioNewSourceMethod loads a sound. The second argument picks how it is
// decoded: 'static' holds the whole sound in memory and suits short effects that
// need to overlap, while 'stream' decodes as it plays and suits music. Static is
// the default because most sounds in a game are effects.
func audioNewSourceMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("audio.newSource", tok, args, 1, 2); err != nil {
		return err
	}

	path, err := text("audio.newSource", tok, args, 0)

	if err != nil {
		return err
	}

	streaming := false

	if len(args) == 2 {
		kind, err := text("audio.newSource", tok, args, 1)

		if err != nil {
			return err
		}

		parsed, ok := engine.SourceKindFromName(kind)

		if !ok {
			return engine.Choice("audio.newSource", tok, kind, engine.SourceKindNames...)
		}

		streaming = parsed
	}

	resolved := resolvePath(path)

	source, loadErr := engine.NewSource(resolved, streaming)

	if loadErr != nil {
		return engine.AssetFailure("audio.newSource", tok, path, resolved, loadErr)
	}

	return source
}

// audioPlayMethod is a shorthand for source.play(), so a one-off sound effect
// reads as audio.play(sound).
func audioPlayMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("audio.play", tok, args, 1); err != nil {
		return err
	}

	source, err := engine.SourceArgument("audio.play", tok, args, 0)

	if err != nil {
		return err
	}

	result, _ := source.Method("play", tok, nil)

	return result
}

// audioStopMethod stops one source, or everything when called with no arguments.
func audioStopMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) == 0 {
		mix.HaltMusic()
		mix.HaltChannel(-1)

		return value.NULL
	}

	source, err := engine.SourceArgument("audio.stop", tok, args, 0)

	if err != nil {
		return err
	}

	result, _ := source.Method("stop", tok, nil)

	return result
}

func audioPauseMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	mix.PauseMusic()
	mix.Pause(-1)

	return value.NULL
}

func audioResumeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	mix.ResumeMusic()
	mix.Resume(-1)

	return value.NULL
}

// audioSetVolumeMethod sets the master volume between 0 and 1. Every source's
// own volume is scaled by it, which is what a settings screen should adjust.
func audioSetVolumeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("audio.setVolume", tok, args, 1); err != nil {
		return err
	}

	volume, err := number("audio.setVolume", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.MasterVolume = clamp(volume, 0, 1)
	engine.Lumen.ApplyMasterVolume()

	return value.NULL
}

func audioGetVolumeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return object.NewFloat(engine.Lumen.MasterVolume)
}
