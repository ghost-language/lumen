package engine

import (
	"fmt"
	"math"
	"strings"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/mix"
)

// Source is a playable sound. Sources come in two flavours: 'static' sources
// are decoded into memory up front and are what sound effects should use, while
// 'stream' sources are decoded as they play and are meant for music.
// SDL_mixer plays one stream at a time, so starting a second one replaces the
// first, which matches how background music is normally used anyway.
type Source struct {
	Path      string
	Streaming bool
	Looping   bool
	Volume    float64
	Chunk     *mix.Chunk
	Music     *mix.Music

	channel int
	spatial *spatial
}

// spatial is where a source sits in the stereo field. SDL_mixer applies these
// effects to a channel rather than to a sound, and a static source is only
// assigned a channel when it starts playing, so the settings are kept here and
// re-applied on every play.
type spatial struct {
	// panning is set by setPanning; positional is set by setPosition. Only one
	// of the two is in force at a time, because each replaces the other's effect
	// on the channel.
	positional bool

	left     uint8
	right    uint8
	angle    int16
	distance uint8
}

// audioMaxVolume is SDL_mixer's full-scale volume.
const audioMaxVolume = 128

// currentStream tracks which streaming source owns the music channel, so
// isPlaying() cannot claim a source is playing after another replaced it.
var currentStream *Source

// NewSource loads a sound from disk. Streaming sources back onto SDL_mixer's
// music channel; static sources become chunks that can overlap freely.
func NewSource(path string, streaming bool) (*Source, error) {
	source := &Source{Path: path, Streaming: streaming, Volume: 1, channel: -1}

	if streaming {
		music, err := mix.LoadMUS(path)

		if err != nil {
			return nil, err
		}

		source.Music = music
	} else {
		chunk, err := mix.LoadWAV(path)

		if err != nil {
			return nil, err
		}

		source.Chunk = chunk
	}

	Lumen.RegisterResource(source)

	return source, nil
}

// String represents the source object's value as a string.
func (source *Source) String() string {
	kind := "static"

	if source.Streaming {
		kind = "stream"
	}

	return fmt.Sprintf("Source: %s (%s)", source.Path, kind)
}

// Type returns the source object type.
func (source *Source) Type() object.Type {
	return SOURCE
}

// Method defines the set of methods available on source objects.
func (source *Source) Method(method string, tok token.Token, args []object.Object) (object.Object, bool) {
	switch method {
	case "play":
		return source.play(tok, args)
	case "stop":
		return source.stop()
	case "pause":
		return source.pause()
	case "resume":
		return source.resume()
	case "isPlaying":
		return &object.Boolean{Value: source.isPlaying()}, true
	case "isPaused":
		return &object.Boolean{Value: source.isPaused()}, true
	case "setLooping":
		return source.setLooping(tok, args)
	case "isLooping":
		return &object.Boolean{Value: source.Looping}, true
	case "setVolume":
		return source.setVolume(tok, args)
	case "getVolume":
		return object.NewFloat(source.Volume), true
	case "fadeIn":
		return source.fadeIn(tok, args)
	case "fadeOut":
		return source.fadeOut(tok, args)
	case "setPanning":
		return source.setPanning(tok, args)
	case "setPosition":
		return source.setPosition(tok, args)
	case "clearEffects":
		return source.clearEffects()
	case "clone":
		return source.clone(tok)
	case "toString":
		return &object.String{Value: source.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// play starts the source. Static sources start a new overlapping playback each
// time; streaming sources restart from the beginning.
func (source *Source) play(tok token.Token, args []object.Object) (object.Object, bool) {
	loops := 0

	if source.Looping {
		loops = -1
	}

	if source.Streaming {
		if err := source.Music.Play(loops); err != nil {
			return SystemFailure("source.play", tok, err), true
		}

		mix.VolumeMusic(source.mixVolume())
		currentStream = source

		return value.NULL, true
	}

	source.Chunk.Volume(source.mixVolume())

	// With every channel busy there is nowhere to put this sound. Dropping it is
	// the right answer: a game in the middle of a loud moment should lose the
	// least important effect, not have the frame that triggered it fail.
	if mix.GroupAvailable(-1) < 0 {
		return value.NULL, true
	}

	channel, err := source.Chunk.Play(-1, loops)

	if err != nil {
		return SystemFailure("source.play", tok, err), true
	}

	source.channel = channel

	source.applySpatial()

	return value.NULL, true
}

// stop halts playback.
func (source *Source) stop() (object.Object, bool) {
	if source.Streaming {
		if currentStream == source {
			mix.HaltMusic()

			currentStream = nil
		}

		return value.NULL, true
	}

	if source.channel >= 0 {
		mix.HaltChannel(source.channel)

		source.channel = -1
	}

	return value.NULL, true
}

// pause suspends playback so it can be resumed from the same position.
func (source *Source) pause() (object.Object, bool) {
	if source.Streaming {
		if currentStream == source {
			mix.PauseMusic()
		}

		return value.NULL, true
	}

	if source.channel >= 0 {
		mix.Pause(source.channel)
	}

	return value.NULL, true
}

// resume continues paused playback.
func (source *Source) resume() (object.Object, bool) {
	if source.Streaming {
		if currentStream == source {
			mix.ResumeMusic()
		}

		return value.NULL, true
	}

	if source.channel >= 0 {
		mix.Resume(source.channel)
	}

	return value.NULL, true
}

// setLooping controls whether playback repeats. It takes effect the next time
// the source is played.
func (source *Source) setLooping(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.setLooping", tok, args, 1); err != nil {
		return err, true
	}

	looping, err := Boolean("source.setLooping", tok, args, 0)

	if err != nil {
		return err, true
	}

	source.Looping = looping

	return value.NULL, true
}

// setVolume sets this source's volume between 0 and 1.
func (source *Source) setVolume(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.setVolume", tok, args, 1); err != nil {
		return err, true
	}

	volume, err := Number("source.setVolume", tok, args, 0)

	if err != nil {
		return err, true
	}

	source.Volume = clamp(volume, 0, 1)

	if source.Streaming {
		if currentStream == source {
			mix.VolumeMusic(source.mixVolume())
		}
	} else if source.Chunk != nil {
		source.Chunk.Volume(source.mixVolume())
	}

	return value.NULL, true
}

// fadeIn starts playback, ramping the volume up over the given seconds.
func (source *Source) fadeIn(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.fadeIn", tok, args, 1); err != nil {
		return err, true
	}

	seconds, err := Number("source.fadeIn", tok, args, 0)

	if err != nil {
		return err, true
	}

	loops := 0

	if source.Looping {
		loops = -1
	}

	milliseconds := int(seconds * 1000)

	if source.Streaming {
		if err := source.Music.FadeIn(loops, milliseconds); err != nil {
			return SystemFailure("source.fadeIn", tok, err), true
		}

		mix.VolumeMusic(source.mixVolume())
		currentStream = source

		return value.NULL, true
	}

	source.Chunk.Volume(source.mixVolume())

	channel, chunkErr := source.Chunk.FadeIn(-1, loops, milliseconds)

	if chunkErr != nil {
		return SystemFailure("source.fadeIn", tok, chunkErr), true
	}

	source.channel = channel

	source.applySpatial()

	return value.NULL, true
}

// fadeOut ramps the volume down over the given seconds and then stops.
func (source *Source) fadeOut(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.fadeOut", tok, args, 1); err != nil {
		return err, true
	}

	seconds, err := Number("source.fadeOut", tok, args, 0)

	if err != nil {
		return err, true
	}

	milliseconds := int(seconds * 1000)

	if source.Streaming {
		if currentStream == source {
			mix.FadeOutMusic(milliseconds)
		}

		return value.NULL, true
	}

	if source.channel >= 0 {
		mix.FadeOutChannel(source.channel, milliseconds)
	}

	return value.NULL, true
}

// setPanning places the sound in the stereo field directly: two volumes between
// 0 and 1, one per speaker. source.setPanning(1, 0) is hard left.
func (source *Source) setPanning(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.setPanning", tok, args, 2); err != nil {
		return err, true
	}

	if source.Streaming {
		return streamingOnly("source.setPanning", tok), true
	}

	left, err := Number("source.setPanning", tok, args, 0)

	if err != nil {
		return err, true
	}

	right, err := Number("source.setPanning", tok, args, 1)

	if err != nil {
		return err, true
	}

	source.spatial = &spatial{
		left:  uint8(clamp(left, 0, 1) * 255),
		right: uint8(clamp(right, 0, 1) * 255),
	}

	source.applySpatial()

	return value.NULL, true
}

// setPosition places the sound around the listener: an angle in degrees, where
// 0 is straight ahead and 90 is to the right, and a distance from 0 (at the
// listener) to 1 (as far away as the mix allows). It is the convenient form for
// world sounds, where a game knows where a thing is but not what that means in
// terms of speaker volumes.
func (source *Source) setPosition(tok token.Token, args []object.Object) (object.Object, bool) {
	if err := Arity("source.setPosition", tok, args, 2); err != nil {
		return err, true
	}

	if source.Streaming {
		return streamingOnly("source.setPosition", tok), true
	}

	angle, err := Number("source.setPosition", tok, args, 0)

	if err != nil {
		return err, true
	}

	distance, err := Number("source.setPosition", tok, args, 1)

	if err != nil {
		return err, true
	}

	// Wrap the angle so a game can accumulate rotation without normalising it.
	angle = math.Mod(angle, 360)

	if angle < 0 {
		angle = angle + 360
	}

	source.spatial = &spatial{
		positional: true,
		angle:      int16(angle),
		distance:   uint8(clamp(distance, 0, 1) * 255),
	}

	source.applySpatial()

	return value.NULL, true
}

// clearEffects returns the source to plain centred stereo.
func (source *Source) clearEffects() (object.Object, bool) {
	source.spatial = nil

	if source.channel >= 0 {
		mix.SetPanning(source.channel, 255, 255)
		mix.SetPosition(source.channel, 0, 0)
	}

	return value.NULL, true
}

// applySpatial pushes the source's position onto whichever channel it is
// playing on. Doing nothing when the source is idle is correct: the settings are
// stored, and play() applies them once a channel exists.
func (source *Source) applySpatial() {
	if source.spatial == nil || source.channel < 0 {
		return
	}

	if source.spatial.positional {
		mix.SetPosition(source.channel, source.spatial.angle, source.spatial.distance)

		return
	}

	mix.SetPanning(source.channel, source.spatial.left, source.spatial.right)
}

// clone returns an independent handle to the same sound, letting one effect
// play several overlapping copies at different volumes.
func (source *Source) clone(tok token.Token) (object.Object, bool) {
	if source.Streaming {
		return streamingOnly("source.clone", tok), true
	}

	clone := &Source{
		Path:    source.Path,
		Looping: source.Looping,
		Volume:  source.Volume,
		Chunk:   source.Chunk,
		channel: -1,
	}

	if source.spatial != nil {
		settings := *source.spatial
		clone.spatial = &settings
	}

	return clone, true
}

// =============================================================================
// Helper methods

// isPlaying reports whether the source is currently producing sound.
func (source *Source) isPlaying() bool {
	if source.Streaming {
		return currentStream == source && mix.PlayingMusic() && !mix.PausedMusic()
	}

	return source.channel >= 0 && mix.Playing(source.channel) == 1 && mix.Paused(source.channel) == 0
}

// isPaused reports whether the source is paused mid-playback.
func (source *Source) isPaused() bool {
	if source.Streaming {
		return currentStream == source && mix.PausedMusic()
	}

	return source.channel >= 0 && mix.Paused(source.channel) == 1
}

// mixVolume converts the source's 0-1 volume into SDL_mixer's 0-128 range.
// Static sources deliberately leave the master volume out of it: SDL_mixer
// multiplies a chunk's volume by its channel's, and the master volume is applied
// to the channels. The music channel has no such second stage, so streaming
// sources fold the master volume in here.
func (source *Source) mixVolume() int {
	volume := source.Volume

	if source.Streaming {
		volume = volume * Lumen.MasterVolume
	}

	return int(clamp(volume, 0, 1) * audioMaxVolume)
}

// Release frees the decoded audio. Clones share their original's chunk and so
// are left alone.
func (source *Source) Release() {
	if source.Music != nil {
		source.Music.Free()
		source.Music = nil
	}

	if source.Chunk != nil {
		source.Chunk.Free()
		source.Chunk = nil
	}
}

// SourceKindNames are the ways a game says how a sound should be decoded, in
// the order a message listing them should read.
var SourceKindNames = []string{"static", "stream"}

// SourceKindFromName maps the source kind a game names onto the streaming flag.
func SourceKindFromName(name string) (bool, bool) {
	switch strings.ToLower(name) {
	case "static":
		return false, true
	case "stream", "streaming":
		return true, true
	}

	return false, false
}

// ApplyMasterVolume re-applies the master volume to whatever is playing now.
func (engine *Engine) ApplyMasterVolume() {
	if currentStream != nil {
		mix.VolumeMusic(currentStream.mixVolume())
	}

	mix.Volume(-1, int(clamp(engine.MasterVolume, 0, 1)*audioMaxVolume))
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}

// streamingOnly reports a method that a streamed source cannot answer. Panning,
// positioning, and cloning all work on a decoded chunk, and a stream does not
// have one — so the fix is never to the arguments, it is to how the source was
// loaded, and the message says so.
func streamingOnly(name string, tok token.Token) *object.Error {
	return State(name, tok, "is only available for `static` sources").
		WithHelp("load the sound with `audio.newSource(path)` instead of `audio.newSource(path, 'stream')`")
}
