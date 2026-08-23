# Lumen

A lightweight 2D game engine for [Ghost](https://github.com/ghost-language/ghost).

Lumen gives Ghost a game loop, a renderer, input, audio, and file access, so a
whole game can be written in Ghost and nothing else. It is built on SDL2 and
takes its shape from [LÖVE](https://love2d.org): the same game loop, the same
transform stack, the same drawing model.

```ghost
import "lumen:keyboard"
import { Image } from "lumen:image"

game = { x: 100, y: 100 }

function load() {
  game.sprite = new Image('resources/player.png')
}

function update(dt) {
  if (keyboard.isDown('right', 'd')) {
    game.x = game.x + 120 * dt
  }
}

function draw() {
  game.sprite.draw(game.x, game.y)
}
```

## Requirements

Lumen links against SDL2 through cgo, so the SDL development libraries have to
be present to build it:

- SDL2
- SDL2_image
- SDL2_ttf
- SDL2_mixer

```bash
# macOS
brew install sdl2 sdl2_image sdl2_ttf sdl2_mixer

# Debian / Ubuntu
apt install libsdl2-dev libsdl2-image-dev libsdl2-ttf-dev libsdl2-mixer-dev
```

Ghost is expected as a sibling checkout (`../ghost`), which is what the
`replace` directive in `go.mod` points at.

```bash
make build                  # builds dist/lumen
make run EXAMPLE=60_rpg     # builds and runs an example
make examples               # starts every example briefly and reports failures
```

## Running a game

```bash
lumen main.ghost      # run a specific file
lumen examples/60_rpg # run the main.ghost inside a directory
lumen game.lumen      # run a packaged game
lumen                 # run the main.ghost beside the binary, or in the
                      # working directory
```

Asset paths (`new Image(path)`, `new Font(path, size)`, `new Source(path)`,
`filesystem.readAsset`) resolve relative to the directory the entry file is in.

## Shipping a game

A game is a folder. To hand it to someone who does not have Lumen, build it.

```bash
lumen build mygame -o mygame            # a standalone executable; players just run it
lumen package mygame -o mygame.lumen    # one file; players run `lumen mygame.lumen`
```

A `.lumen` file is a zip of the game's sources and assets with `main.ghost` at
the root. `build` appends that archive to a copy of the Lumen binary, so the
result is one file with both the engine and the game in it. `lumen fuse` is an
older name for the same command and still works.

On macOS a built binary is re-signed with an ad-hoc signature as the last step,
because appending anything to an executable invalidates its existing signature
and the system kills a binary whose signature does not match its contents. That
needs `codesign`, which arrives with the Xcode command line tools; shipping to
other people's machines wants a real identity and notarisation on top.

A packaged game is unpacked into a cache directory the first time it runs, and
run from there. That is deliberate rather than incidental: Ghost resolves
`import` and its own `file` module against the real filesystem, so serving Lumen's
loaders out of the zip while Ghost's imports still needed real paths would give
a game two disagreeing views of its own files. Unpacking costs a moment on first
launch and keeps every path in the system pointing at the same thing. Archives
are keyed by a hash of their contents, so a second launch reuses what is there
and a changed game never reads a stale mixture.

## Game loop

Lumen calls `load()` once, then `update(dt)` and `draw()` every frame until the
game quits.

### `load()`

Runs once, before the first frame. Load assets and build initial state here. If
`load()` raises an error Lumen reports it and exits, rather than running a game
whose state was never finished.

### `update(dt)`

`dt` is how many seconds the previous frame took. **Scale anything that moves by
it.** A speed written as pixels-per-frame changes with the frame rate; a speed
written as pixels-per-second does not.

```ghost
function update(dt) {
  game.x = game.x + 90 * dt   // 90 pixels per second, on any machine
}
```

`dt` is capped at 0.25s, so a frame that stalls cannot teleport everything
through a wall. A game that ignores the argument still works — Ghost drops
arguments a function does not declare.

### `draw()`

Renders one frame. The draw state — transform, color, line width, blend mode,
scissor — is reset to its defaults before `draw()` runs, so each frame is
self-contained. The current font is the exception: it is a choice a game makes
once and keeps.

### Event callbacks

Every callback below is optional; define the ones a game needs.

| Callback | When |
| --- | --- |
| `keypressed(key, isRepeat)` | a key goes down |
| `keyreleased(key)` | a key comes up |
| `textinput(text)` | text is typed, between `keyboard.startTextInput()` and `stopTextInput()` |
| `mousepressed(x, y, button, clicks)` | a mouse button goes down |
| `mousereleased(x, y, button)` | a mouse button comes up |
| `mousemoved(x, y, dx, dy)` | the pointer moves |
| `wheelmoved(x, y)` | the wheel turns |
| `resize(width, height)` | the window is resized |
| `focus(hasFocus)` | the window gains or loses focus |
| `joystickadded(count)` / `joystickremoved(count)` | a controller is plugged in or unplugged |
| `quit()` | the window is closed; return `true` to cancel |

Use `keypressed` rather than `keyboard.isDown` for menus and dialogue: it fires
once per physical press, where `isDown` is true on every frame the key is held.

## Coordinates and transforms

The origin is the top-left of the window, x to the right and y down. Every
coordinate passed to the canvas goes through the transform on top of the
transform stack, which is how cameras and zoom work.

```ghost
function draw() {
  canvas.push()                 // save the current transform
  canvas.scale(3)               // 16px tiles drawn at 48px
  canvas.translate(-camera.x, -camera.y)

  map.draw()                    // world coordinates
  player.draw()

  canvas.pop()                  // back to screen coordinates

  canvas.print('Score', 20, 20) // unaffected by the camera
}
```

Transforms **compose**: `canvas.scale(2)` after `canvas.scale(3)` gives 6x, it
does not replace the 3x. `canvas.push('all')` also saves the color, line width,
point size, blend mode, and scissor, and `canvas.pop()` restores them.

`canvas.toWorld(x, y)` and `mouse.getWorldPosition()` map a screen position back
through the current transform, which is how a game finds what the player clicked
on while a camera is active.

`canvas.getVisible()` goes the other way and reports `[x, y, width, height]`:
the region of the current coordinate space that is on screen. A game with a
world larger than the window uses it to loop over only the part of the world the
player can see, without having to redo the camera's arithmetic itself.

```ghost
visible = canvas.getVisible()

left = math.max(0, math.floor(visible[0] / tileSize))
right = math.min(map.width - 1, math.ceil((visible[0] + visible[2]) / tileSize))
```

## Drawing performance

Sprites are not handed to SDL one at a time. Consecutive draws that share a
texture, blend mode, and scissor box are collected into a single call, and any
sprite whose transformed corners fall entirely off screen is dropped before it
gets that far. A tilemap drawing a few thousand tiles from one tileset costs one
call, not a few thousand.

Two things are worth knowing, because both are visible from Ghost:

**A primitive between two sprites ends the batch.** Batching only ever merges
draws with the ones immediately before them, because reordering them would put
sprites through each other. Drawing a rectangle between every two sprites is
therefore correct but slow. Drawing the sprites together and the primitives
together is the same picture for a fraction of the calls.

**Changing texture ends it too.** Sprites clipped from one sheet batch together;
alternating between two sheets does not. This is the usual argument for packing a
game's art into as few sheets as it can stand.

Neither is something a game has to manage, and neither changes what is drawn —
they are the two things that decide whether a frame is one call or a thousand.

Measured on `53_top_down`, a 50x50 map across six layers: 33 FPS before, 116
after, drawing a pixel-identical frame. `60_rpg`, which was already culling its
tilemap by hand, gains a fifth on top of that.

Loading the same file twice hands back the same image rather than uploading a
second copy of it, and `clip()` reuses the view it made for a region last time,
so a spritesheet costs one texture and a fixed set of views however many times a
frame reaches for them.

## Colors

**Red, green, and blue run 0-255. Alpha runs 0-1.** The two ranges are
deliberately different: a single range that accepts both and guesses from the
value cannot tell `1` (nearly transparent) from `1.0` (fully opaque), and gets
it silently wrong either way.

```ghost
color.rgb(255, 128, 0)          // opaque orange
color.rgb(255, 128, 0, 0.5)     // the same orange at half opacity
color.hex('#ff8800')            // #rgb, #rgba, #rrggbb, and #rrggbbaa
color.hsl(30, 1, 0.5)           // hue in degrees, saturation and lightness 0-1
color.white.withAlpha(0.25)
```

The current color tints everything drawn, images and text included. Set it back
to `color.white` before drawing sprites you do not want tinted.

## Modules and imports

Ghost's standard library is import-only — `console` and `type` are the only
names reachable without one — and Lumen's own modules follow the same rule,
registered under their own `lumen:` scheme rather than borrowed from Ghost's
`ghost:`. A game imports whatever it uses, from whichever scheme it lives
under:

```ghost
import "ghost:math"                    // Ghost's own standard library
import "lumen:canvas"                  // the whole module, bound to `canvas`
import "lumen:canvas" as gfx           // aliased
import { setColor, print } from "lumen:canvas"   // named imports
```

A class a module exports — `Image`, `Spritesheet`, `Animation`, `Source`,
`Font`, `Target`, `Quad` — is pulled in the same way and `new`-ed exactly like
a Ghost-defined class:

```ghost
import { Image } from "lumen:image"

sprite = new Image('resources/player.png')
```

The reference below names, for each module, which scheme it lives under —
`lumen:name` for everything in this section, `ghost:name` for Ghost's own
`math`, `random`, `json`, and the rest — and which classes it exports, if any.

### `canvas`

Drawing, the drawing state, and the transform stack.

**Shapes** — `rectangle(x, y, w, h)`, `filledRectangle(...)`,
`circle(x, y, r, [segments])`, `filledCircle(...)`,
`ellipse(x, y, rx, ry, [segments])`, `filledEllipse(...)`,
`arc(x, y, r, startAngle, endAngle, [segments])`, `filledArc(...)`,
`polygon(x1, y1, x2, y2, x3, y3, ...)` (or one list of coordinates),
`filledPolygon(...)`, `line(x1, y1, x2, y2, ...)` (any number of points),
`point(x, y, ...)`.

**State** — `clear([color])`, `setColor(color)` or `setColor(r, g, b, [a])`,
`getColor()`, `setBackgroundColor(color)`, `setLineWidth(n)`, `getLineWidth()`,
`setPointSize(n)`, `setBlendMode('alpha'|'add'|'multiply'|'none')`,
`setScissor(x, y, w, h)`, `clearScissor()`.

**Text** — `print(text, x, y, [rotation, sx, sy, ox, oy])`,
`printf(text, x, y, limit, ['left'|'center'|'right'], [rotation, sx, sy, ox, oy])`,
`setFont(font)`, `getFont()`, `resetFont()`.

**Transforms** — `push(['all'])`, `pop()`, `origin()`, `translate(x, y)`,
`rotate(radians)`, `scale(x, [y])`, `shear(x, y)`, `toScreen(x, y)`,
`toWorld(x, y)`, `getVisible()`.

**Targets** — `setTarget([target])`, `screenshot(filename)`. Setting a target
resets the transform, because a target is its own screen: (0, 0) is its own
corner, not the window's. Clearing it restores the transform the window draws
with. `new Target(w, h)` and `new Quad(x, y, w, h)` — see **Target** and
**Quad** below — are exported classes, imported with
`import { Target, Quad } from "lumen:canvas"`.

**Properties** — `canvas.width`, `canvas.height` (of the render target when one
is set, of the window otherwise).

### `color`

`rgb(r, g, b, [a])`, `rgba(...)`, `hex(string)`, `hsl(h, s, l, [a])`.

Named colors: `black`, `white`, `transparent`, `red`, `green`, `blue`, `yellow`,
`orange`, `purple`, `cyan`, `magenta`, `brown`, `gray`, `lightGray`, `darkGray`.

### `image`

No free functions of its own — it exists to hold the classes below.
`import { Image, Spritesheet, Animation } from "lumen:image"`.

`new Image(path)` hands back the image already in memory when a path has been
loaded before, so loading the same sheet from two places costs one texture.
`new Spritesheet(path, frameSize)` or `new Spritesheet(path, frameWidth,
frameHeight)`, and `new Animation(sheet, frames, secondsPerFrame,
['loop'|'once'|'pingpong'])` — see **Image**, **Spritesheet**, and
**Animation** below.

### `font`

`system(size)` for the built-in font at a size, `system()` for the current
default. `new Font(path, size)` or `new Font(size)` for the built-in font —
see **Font** below; `import { Font } from "lumen:font"`.

### `audio`

`play(source)`, `stop([source])`, `pause()`, `resume()`, `setVolume(0-1)`,
`getVolume()`. `new Source(path, ['static'|'stream'])` — see **Source**
below; `import { Source } from "lumen:audio"`.

`'static'` decodes the whole sound up front and can overlap with itself — use it
for effects. `'stream'` decodes while playing — use it for music. WAV, OGG, and
MP3 are supported.

### `keyboard`

`isDown(key, ...)`, `isUp(key, ...)`, `wasPressed(key, ...)`,
`wasReleased(key, ...)`, `startTextInput()`, `stopTextInput()`,
`isTextInputActive()`.

Each takes any number of key names and is true if any of them matches, so one
action can be bound to several keys: `keyboard.isDown('left', 'a')`. Names are
SDL key names and are matched case-insensitively (`'left'`, `'space'`,
`'escape'`, `'return'`, `'f1'`, `'a'`). An unrecognised name is an error rather
than a silent false.

### `mouse`

`showCursor()`, `hideCursor()`, `isVisible()`, `isButtonDown(button)`,
`isButtonUp(button)`, `wasButtonPressed(button)`, `wasButtonReleased(button)`,
`getPosition()`, `setPosition(x, y)`, `getWorldPosition()`,
`setRelativeMode(bool)`, `setGrabbed(bool)`.

Buttons are `'left'`, `'middle'`, `'right'`, `'x1'`, `'x2'`.

**Properties** — `mouse.x`, `mouse.y`, `mouse.wheel`, `mouse.wheelX`.

### `joystick`

`isDown(index, button)`, `isUp(...)`, `wasPressed(...)`, `wasReleased(...)`,
`getAxis(index, axis, [deadZone])`, `getName(index)`, `isConnected(index)`,
`vibrate(index, strength, [strength2], [seconds])`.

**Property** — `joystick.count`.

Controllers are numbered from 1. Buttons use SDL game-controller names: `'a'`,
`'b'`, `'x'`, `'y'`, `'start'`, `'back'`, `'guide'`, `'leftshoulder'`,
`'rightshoulder'`, `'leftstick'`, `'rightstick'`, `'dpup'`, `'dpdown'`,
`'dpleft'`, `'dpright'`. Axes are `'leftx'`, `'lefty'`, `'rightx'`, `'righty'`,
`'triggerleft'`, `'triggerright'`; sticks report -1 to 1 and triggers 0 to 1,
with a 0.15 dead zone by default. Checking a controller that is not plugged in
reads as "not pressed" rather than raising.

### `timer`

`getDelta()`, `getFps()`, `getTime()`, `sleep(seconds)`.

**Properties** — `timer.delta`, `timer.averageDelta`, `timer.fps`, `timer.time`
(seconds since the game started), `timer.frame`.

### `window`

`setTitle(title)`, `setMode(w, h, [fullscreen])`, `setSize(w, h)`,
`setLogicalSize(w, h)`, `clearLogicalSize()`, `getLogicalSize()`,
`setPixelPerfect(bool)`, `setFullscreen(bool)`, `toggleFullscreen()`,
`setResizable(bool)`, `setBorderless(bool)`, `setVsync(bool)`, `setIcon(image)`,
`setPosition(x, y)`, `center()`, `maximize()`, `minimize()`, `restore()`,
`getDesktopDimensions()`, `getDimensions()`.

**Properties** — `window.width`, `window.height`, `window.scale`, `window.title`,
`window.fps`, `window.fullscreen`, `window.focused`.

#### A fixed canvas

By default a game draws in window pixels, so a bigger window shows more of the
world at the same size and leaves the interface stranded in the corner it was
placed in. That is the right answer for an application and the wrong one for a
game, where the layout of the screen is part of the design.

```js
window.setLogicalSize(320, 180)   // the game now draws in a 320x180 space
window.setPixelPerfect(true)      // ...scaled by whole numbers only
```

With a logical size set, Lumen scales that space to fill as much of the window
as it can, keeps its proportions, centres it, and paints black bars around what
is left. Going fullscreen makes the game bigger rather than wider.
`window.width` and `window.height` report the logical size — the space the game
actually draws in — and mouse positions and `mousepressed` arrive in the same
coordinates, so nothing in a game has to know that scaling is happening.
`window.getDimensions()` still reports the real window, and `window.scale` is
how many screen pixels one game pixel currently covers.

`setPixelPerfect(true)` rounds that scale down to a whole number, which is what
keeps pixel art from developing uneven edges; it costs more of the screen to the
bars, and is worth it for a small canvas and not for a large one. Scaling *down*
stays fractional either way, since a window smaller than the game still has to
show all of it.

### `filesystem`

Saved games belong in the player's own data directory, not next to the program:
a game installed read-only cannot write to its own folder. Ghost's built-in
`file` module reads and writes next to the source, which is right for assets
and wrong for saves, so Lumen adds this.

Saves land in `~/.local/share/lumen/<identity>` on Linux (honouring
`XDG_DATA_HOME`), `~/Library/Application Support/lumen/<identity>` on macOS, and
`%AppData%\lumen\<identity>` on Windows.

`setIdentity(name)` (call once in `load()`), `getSaveDirectory()`,
`write(name, contents)`, `append(name, contents)`, `read(name)`, `exists(name)`,
`remove(name)`, `createDirectory(name)`, `getDirectoryItems([name])`,
`readAsset(path)`.

`read()` returns `null` when the file does not exist, so "no save yet" is an
ordinary case rather than an error. Save paths cannot escape the save directory.

`readAsset(path)` is the read-only counterpart, resolved against the game's own
directory — use it for maps, dialogue, and other shipped data.

### `system`

`getClipboardText()`, `setClipboardText(text)`, `openUrl(url)` (http and https
only), `getPowerInfo()`.

**Properties** — `system.os`, `system.processorCount`.

### `lumen`

`quit()`, `setTargetFps(n)` (0 for uncapped), `getTargetFps()`.

**Property** — `lumen.version`.

### `math`

`import "ghost:math"` — this is Ghost's own module, not one of Lumen's.
Lumen extends it rather than competing with it. Added:
`floor`, `ceil`, `round`, `sqrt`, `pow`, `exp`, `log`, `sign`, `asin`, `acos`,
`atan`, `atan2`, `degrees`, `radians`, `clamp(v, low, high)`,
`lerp(from, to, amount)`, `distance(x1, y1, x2, y2)`, `angle(x1, y1, x2, y2)`,
`random()`, `random(n)`, `random(low, high)`, `randomSeed(n)`,
`noise(x, [y])`.

## Objects

Ghost cannot expose properties on objects a host program defines, only methods,
so these are all method calls.

`Image`, `Spritesheet`, `Animation`, `Source`, `Font`, `Target`, and `Quad` are
native classes: built and driven by Lumen, `new`-able exactly like a
Ghost-defined class once imported from the module that exports them (noted
under each one below). Everything else below is returned by a call rather than
constructed directly — `color.rgb(...)` for a `Color`, and so on.

### Image

`draw(x, y, [rotation, sx, sy, ox, oy])`,
`drawQuad(quad, x, y, [rotation, sx, sy, ox, oy])`,
`clip(x, y, size)` or `clip(x, y, w, h)`, `getWidth()`, `getHeight()`,
`getDimensions()`, `getPixel(x, y)`, `setFilter('nearest'|'linear')`.

Rotation is in radians, applied about the origin offset `(ox, oy)`. Negative
scale flips: `sprite.draw(x, y, 0, -1, 1)` mirrors horizontally.

`clip()` returns a lightweight view onto part of the image, which is how one
spritesheet becomes many sprites. `getPixel()` reads from the source image,
which lets collision or spawn data be baked into a map image.

### Spritesheet

`draw(frame, x, y, [rotation, sx, sy, ox, oy])`,
`getQuad(frame)`, `getImage()`, `getCount()`, `getColumns()`, `getRows()`,
`getFrameWidth()`, `getFrameHeight()`, `getFrameDimensions()`.

One image cut into a grid of equally sized frames, numbered left to right and
top to bottom from zero. The quads are built once, when the sheet is made, so
drawing a frame never allocates. `import { Spritesheet } from "lumen:image"`.

```ghost
sheet = new Spritesheet('characters.png', 16)   // square frames
sheet = new Spritesheet('characters.png', 16, 24)
sheet.draw(4, x, y)
```

The first argument may also be an image that is already loaded, which is how two
sheets get cut from one file at different frame sizes.

### Animation

`update([dt])`, `draw(x, y, [rotation, sx, sy, ox, oy])`, `play()`, `pause()`,
`resume()`, `stop()`, `reset()`, `seek(seconds)`, `clone()`, `isPlaying()`,
`isPaused()`, `isFinished()`, `getFrame()`, `getFrameCount()`, `getElapsed()`,
`getLength()`, `getSheet()`, `getDuration()`, `setDuration(n)`, `getSpeed()`,
`setSpeed(n)`, `getMode()`, `setMode(name)`.

A sequence of a sheet's frames played against a clock, built with
`new Animation(sheet, frames, secondsPerFrame, ['loop'|'once'|'pingpong'])`;
`import { Animation } from "lumen:image"`.

**The clock is time, not frames drawn.** `secondsPerFrame` is what it says, so a
walk cycle runs at the same speed on a machine managing 30 FPS as on one managing
144. Counting draws instead is the obvious version of this and is wrong; it is
what two of the examples in this repository used to do.

```ghost
walk = new Animation(sheet, [4, 5, 6, 7], 0.16)

function update(dt) {
  walk.update(dt)     // or walk.update() to use the last frame's time
}

function draw() {
  walk.draw(x, y)
}
```

A duration of `0` holds the first frame, which is how a still pose is written.
`'once'` stops on the last frame and sets `isFinished()`; `'loop'` (the default)
and `'pingpong'` never finish. `setSpeed(-1)` runs it backwards.

Each animation carries its own playhead, so two characters showing the same
walk cycle need one each — `clone()` is the cheap way to get one. A game that
already keeps its own clock can skip `update()` and call `seek(elapsed)` instead.

### Quad

`getX()`, `getY()`, `getWidth()`, `getHeight()`, `setViewport(x, y, w, h)`.

A quad is a rectangle of a texture. Building one per animation frame up front and
reusing it beats allocating one per draw. `new Quad(x, y, w, h)`;
`import { Quad } from "lumen:canvas"`.

### Font

`print(text, x, y, [...])`, `printf(text, x, y, limit, [align], [...])`,
`getWidth(text)`, `getHeight()`, `getLineHeight()`, `setLineHeight(n)`,
`getAscent()`, `getDescent()`, `getBaseline()`, `getWrap(text, limit)`,
`getSize()`.

`new Font(path, size)` loads a TrueType font at a pixel size; `new Font(size)`
returns Lumen's built-in font at that size. `import { Font } from "lumen:font"`.

`getWidth()` is what centring text and sizing a dialogue box need. Rendered
strings are cached, so drawing the same text every frame costs almost nothing;
text that changes every frame is evicted automatically.

### Color

`getRed()`, `getGreen()`, `getBlue()`, `getAlpha()`, `toHex()`,
`withAlpha(0-1)`, `lerp(other, amount)`.

### Target

`draw(x, y, [...])`, `getWidth()`, `getHeight()`, `getDimensions()`.

An off-screen surface. Draw into it with `canvas.setTarget(target)`, return to
the window with `canvas.setTarget()`, then draw it like any image.
`new Target(w, h)`; `import { Target } from "lumen:canvas"`.

### Source

`play()`, `stop()`, `pause()`, `resume()`, `isPlaying()`, `isPaused()`,
`setLooping(bool)`, `isLooping()`, `setVolume(0-1)`, `getVolume()`,
`fadeIn(seconds)`, `fadeOut(seconds)`, `clone()`,
`setPanning(left, right)`, `setPosition(angle, distance)`, `clearEffects()`.

`new Source(path, ['static'|'stream'])`; `import { Source } from "lumen:audio"`.

`setPanning` takes a volume per speaker, each 0 to 1, so `setPanning(1, 0)` is
hard left. `setPosition` is the convenient form for a sound with a place in the
world: an angle in degrees, where 0 is ahead and 90 is to the right, and a
distance from 0 to 1. Both apply to `'static'` sources only — SDL_mixer places
channels, and music does not play on one.

A sound that cannot find a free channel is dropped rather than raising. In a
loud moment the least important effect should go missing, not the frame that
triggered it.

## Writing Ghost for Lumen

A few of Ghost's rules surprise people coming from other languages, and every
one of them bites hardest inside a game loop.

**Assignment is local to the function it happens in.** A bare `score = score + 1`
inside `update()` creates a new local each frame; the outer `score` never
changes. Keep mutable state on a map or a class instance, where assignment goes
through a property and mutates in place:

```ghost
game = { score: 0 }

function update(dt) {
  game.score = game.score + 1   // works
}
```

**`and` and `or` evaluate both sides.** They do not short-circuit, so a guard
cannot protect the test beside it:

```ghost
if (j >= 0 and list[j].y > 0) { }   // list[j] is read even when j is -1
```

Split it into nested `if`s instead. A negative or out-of-range list index reads
as `null` rather than raising, so the failure shows up later as "cannot read
property of null".

**An imported name shadows anything else you'd call it.** If a file imports
`"lumen:font"`, naming a variable or parameter `font` in that file hides the
module for the rest of it. The same goes for a named class import — a file that
imports `{ Image }` cannot also use `Image` as a variable name. Name them
`bodyFont`, `sprite`, and so on; it only matters in the file that did the
importing, since imports are not global.

**`default` is a keyword.** It cannot be used as a method name, which is why the
built-in font is `font.system(size)`.

## When something goes wrong

A game runs in its own window, and often on a machine nobody started it from a
terminal on. So every failure is reported twice: once to the console, and once
into the window, where whoever is looking at the game will actually see it.

Both reports say the same things in the same order — what sort of failure it is,
what happened, where, the line it happened on with the offending part marked,
what was in flight at the time, and what to do about it.

```
argument error: `canvas.print()` expects argument 1 to be a string, got number
 --> main.ghost:12:5
   |
12 |     canvas.print(score, 10, 10)
   |     ^^^^^^^^^^^^
   |
   = in drawScore(), called at main.ghost:30:3
   = in draw()
   = help: did you mean `text(score)`?
```

The window shows that report on an error screen, and the game holds still on it:

| Key | What it does |
| --- | --- |
| `esc` or `q` | close the game |
| `enter` or `space` | carry on, if carrying on is possible |
| `c` | copy the report to the clipboard |

Carrying on is offered for a failure inside `update()`, `draw()`, or an input
callback — one bad frame is worth watching past. It is not offered for a failure
in `load()` or in the game's source, because neither ever built the state every
later frame reads. A failure carried on past is not stopped for again: it is
counted in the console instead, so the same broken frame cannot bury everything
above it.

While the error screen is up, no game code runs at all — not `update()`, not
`draw()`, and not the input callbacks, so the keys above always belong to the
error screen even in a game that has bound them.

### What the kinds mean

The first two words say what sort of mistake it is before the sentence explains
it, and they are Ghost's own kinds, used the same way:

| Kind | What it means |
| --- | --- |
| `syntax error` | source that could not be read |
| `name error` | a reference to something that was never defined |
| `type error` | an operation applied to the wrong sort of value |
| `argument error` | a call whose arguments do not fit what it is calling |
| `index error` | a subscript outside what it indexes — a frame the sheet does not have, a pixel outside the image |
| `value error` | a value of the right type the operation cannot accept — a negative size, a mode that does not exist |
| `property error` | a member the value does not have |
| `system error` | the world outside the game refusing: a missing asset, a file that will not open |
| `internal error` | a bug in Lumen or Ghost, with a note asking for it to be reported |

Where Lumen is holding the answer, it offers it. A misspelled asset name is
answered with the file sitting next to it, and a misremembered mode, key, button,
or axis name with the nearest real one:

```
system error: `Image()` could not load `playr.png`: No such file or directory
 --> main.ghost:2:20
  |
2 |     player = new Image("playr.png")
  |                    ^^^^
  |
  = in load()
  = help: did you mean `player.png`?
```

### Running without a window

A game run headlessly — in CI, in a build script, over ssh — has a window nobody
can see and no way to dismiss what is on it, so Lumen writes the report to the
console and stops rather than holding a window open that nothing will ever
close. A run that ended on a failure exits non-zero.

Console reports are colored when the terminal can show it, and follow the usual
switches: `NO_COLOR`, `CLICOLOR=0`, `FORCE_COLOR`, `CLICOLOR_FORCE`.

Set `LUMEN_DEBUG=1` to attach the Go stack to an `internal error`. It is only
worth doing when filing a bug: it says nothing about the game, and everything
about where Lumen broke.

## Examples

`make run EXAMPLE=<name>`, or `lumen examples/<name>`.

| Example | Shows |
| --- | --- |
| `01_draw` | shapes and colors |
| `02_input` | reading the keyboard |
| `03_modular` | splitting a game across files |
| `04_collision` | rectangle overlap |
| `05_translate` | moving the world under a camera |
| `06_spritesheets` | slicing an image into tiles |
| `07_animations` | frame animation |
| `08_tilemap` | drawing a grid of tiles |
| `11_camera` | following a player |
| `12_mouse` | pointer position and buttons |
| `13_mouse_select` | click-and-drag selection |
| `50_conway` | Conway's Game of Life |
| `51_player_animations` | directional walk cycles |
| `52_tiled_maps` | loading a map exported from Tiled |
| `53_top_down` | a tiled world with a following camera |
| `60_rpg` | **a complete top-down RPG with turn-based battles** |

`60_rpg` is the one to read first if you are building something. It has a Tiled
map with per-layer collision and view culling, a smoothed camera with bounds and
screen shake, dt-driven walk animations, depth-sorted characters, NPCs and
typewriter dialogue, weighted random encounters, front-view turn-based battles
with spells and items, levelling, a party with equipment and a shared pack,
scrolling menus, saving and loading, sound, and gamepad support.

## What Lumen still needs

Lumen can now carry a real game from an empty folder to a file you hand someone.
`60_rpg` is the proof: a world, battles, menus, saving, sound, and a single-file
build, in Ghost and nothing else. What follows is an honest account of what is
still missing, roughly in the order it will bite.

**Cross-compiling.** This is the sharpest edge. Lumen links SDL2 through cgo, so
`lumen build` produces a binary for the machine that ran it, and only that. There
is no way to build a Windows executable from a Mac. Shipping to three platforms
today means building on three platforms. Until that is solved with CI that
builds and fuses per platform, "shippable" has an asterisk on it.

**A test suite.** What a failure says is tested; what a frame looks like is not.
The engine's drawing correctness still rests on running the examples and looking
at them. The pieces to fix that already exist — SDL's dummy video driver runs the
whole engine headless, and `canvas.screenshot()` can capture a frame — so
golden-image tests of the drawing paths are reachable, and worth having before
the module surface grows further.

**Asset hot-reloading.** Changing a sprite or a line of dialogue means restarting
the game. For a tool people iterate in, watching the game directory and
reloading changed images, fonts, and sounds is among the highest-value things
left.

**Text input polish.** The pieces are there — `keyboard.startTextInput()` and the
`textinput` callback — but every game that wants a name entry field has to build
caret movement, selection, and clipboard handling itself.

**Shaders and particles.** Neither exists. Particles can be written in Ghost and
will be fine for most games; shaders cannot be worked around, and rule out
lighting, palette swaps, and whole-screen effects.

**Physics.** There is no Box2D equivalent. Axis-aligned collision is a few lines
of Ghost, which covers a top-down RPG and most puzzle games, and covers nothing
that needs slopes, joints, or stacking.

**Ghost-side gaps that surface as engine gaps.** Ghost has no string slicing (the
dialogue box in `60_rpg` builds its typewriter effect a character at a time), no
list removal or sorting (the party rebuilds lists to remove an item, and sorts by
hand), and no way for a host program to expose properties on its own objects, so
everything Lumen hands back is a method call. None of these belong to Lumen, but
every one of them is felt while writing a game in it.

None of that stops a game shipping today on the platform it was built on. It is
the difference between an engine that works and one that gets out of your way.

## Differences from LÖVE

Lumen follows LÖVE's model but is not a port, and does not try to be.

- Drawing methods live on the objects being drawn (`sprite.draw(x, y)`) rather
  than on the graphics module (`love.graphics.draw(sprite, x, y)`), which is the
  shape the rest of Ghost's standard library already has.
- The graphics module is `canvas`; render targets are `Target` objects.
- Color channels are 0-255 and alpha is 0-1, where LÖVE uses 0-1 for both.
- The draw state resets every frame, where LÖVE carries color and line width
  across frames.
- Batching is automatic rather than an object a game builds. There is no
  `SpriteBatch` to fill: consecutive draws that share a texture are collected
  into one call as they are made, and off-screen sprites are dropped before they
  reach SDL. See **Drawing performance** below for what keeps that working.
- Spritesheets and animations are engine objects. LÖVE leaves both to libraries
  like `anim8`; Lumen has no package ecosystem to leave them to, so
  `Spritesheet` and `Animation` ship with it as native classes.
- Not implemented: shaders, particle systems, physics (Box2D), meshes, threads,
  video, and touch. Simple axis-aligned collision is a few lines of Ghost —
  `04_collision` and `60_rpg` both show it.
