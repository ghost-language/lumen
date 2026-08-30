# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
make build              # Build dist/lumen for this machine
make run EXAMPLE=60_rpg # Build and run an example
make test               # Unit tests, headless through SDL's dummy driver
make examples           # Start every example for a few seconds, reporting any that fail
make check              # fmt, vet, test, examples — what CI runs
```

Lumen links SDL2 through cgo, so a build always targets the machine that ran it.
`go.mod` points at `../ghost`, so the Ghost that Lumen builds against is whatever
is checked out there rather than the version in `require`.

## Architecture Overview

Lumen is a game engine that runs games written in Ghost. A game is a folder with
a `main.gs` in it defining callbacks; Lumen drives them.

**Ghost source → engine.Run → callbacks (load, update, draw, input) → SDL**

- **cmd/** — the CLI: running a game, `build` (a standalone binary with the game
  fused into it), and `package` (a `.lumen` archive). It finds the game, hands
  it to Ghost, and starts the loop.
- **engine/** — the loop, the window, drawing, audio, input, and the objects a
  game holds: `Image`, `Quad`, `Font`, `Color`, `Target`, `Source`, `Transform`,
  `Spritesheet`, `Animation`. These implement Ghost's `object.Object`, with
  types numbered above Ghost's own enum (`engine/types.go`).
- **modules/** — the module layer a game calls: `canvas`, `window`, `image`,
  `font`, `audio`, `color`, `keyboard`, `mouse`, `joystick`, `timer`,
  `filesystem`, `system`, `lumen`. Registered with Ghost in `modules.Register()`
  under Lumen's own `lumen:` import scheme (`ghost.RegisterModuleForScheme`),
  the same way Ghost's own standard library sits behind `ghost:` — a game
  writes `import "lumen:canvas"`, not a bare `canvas.scale(...)`. Host resources
  a script builds with `new` — `Image`, `Spritesheet`, `Animation`, `Source`,
  `Font`, `Target`, `Quad` — are native classes (`ghost.RegisterClassForScheme`,
  `object.NativeClass`/`Constructible`) rather than factory methods: `new
  Image(path)`, not `image.load(path)`. The constructor functions live beside
  the module they belong to (`modules/image.go`'s `imageConstructor`, and so
  on) and are plain `object.GoFunction`s like any other library method.
- **resources/** — what Lumen carries inside its own binary, currently the
  built-in font.

### Key design patterns

- **One engine, reachable globally.** Ghost's module layer is free functions with
  no receiver, so the running engine is `engine.Lumen`.
- **Batching is automatic.** Consecutive draws sharing a texture are collected
  into one call (`engine/batch.go`) and off-screen sprites are dropped before
  they reach SDL (`culled` in `engine/graphics.go`). Anything that draws outside
  that path has to `Flush()` first.
- **The draw state resets every frame.** A game that forgets to pop a transform
  does not compound it.
- **Logical size is separate from window size.** A game with a fixed coordinate
  space is scaled and letterboxed into the window (`engine/viewport.go`); input
  positions are mapped back through it.

## Error handling

Every failure is a `*fault.Fault` — Ghost's structured diagnostic: a kind, a
sentence, a position with a width, an optional suggested fix, and the calls that
were in flight. Nothing formats a position into a message by hand. See
`engine/errors.go` for how one is built and `engine/report.go` for what happens
to it.

Read Ghost's own `CLAUDE.md` chapter on this first: Lumen's rules are Ghost's
rules, applied to a program that draws in a window.

### Rules

- **Failures are reported twice.** A game runs in its own window, often on a
  machine nobody opened a terminal on, so a report that only reaches stderr
  reaches nobody. `engine.Raise` writes to the console and puts the same report
  on screen (`engine/errorscreen.go`). Reporting is the loop's job: nothing
  below it prints, and nothing below it decides whether the game carries on.
- **Errors never reach Go.** A game cannot produce a Go panic or a Go traceback.
  Module methods read their arguments through the helpers in `engine/errors.go`
  rather than asserting types inline, and `engine.call` recovers anything that
  still escapes as an `internal error` that asks to be reported, with the stack
  attached only when `LUMEN_DEBUG` is set.
- **One vocabulary.** `Arity`, `Number`, `Text`, `ImageArgument`, and the rest of
  `engine/errors.go` are the only place an argument problem is worded, and
  `modules/args.go` wraps them rather than restating them. The sentences are
  deliberately identical to Ghost's in `object/arguments.go` — they are not
  simply calls into Ghost's because Ghost has no name for a Lumen value and
  answers "unknown"; `engine.TypeName` is what fills that hole, and every
  message goes through it.
- **Kinds, not "runtime error".** Pick the `fault.Kind` that says what sort of
  mistake it is: `Syntax`, `Name`, `Type`, `Argument`, `Index`, `Value`,
  `Property`, `Import`, `System`, `Internal`. A failure that fits none of them is
  usually a message that has not been thought through yet.
- **Messages describe, they do not locate.** Write ``"`%s` expects a color"``,
  not `"%d:%d: runtime error: ..."`. The position comes from the token, which
  every module method and every object method is handed. Messages start
  lower-case, end without a full stop, and quote code in backticks.
- **Help is for what to do next.** Where Lumen is holding the answer, offer it:
  `Choice` lists the modes a method accepts and suggests the nearest, and
  `AssetFailure` suggests the file sitting next to the one that was asked for.
  A failure with nothing useful to suggest should not have a help line.
- **Start-up is the exception.** Failures in `initSDL` happen before there is a
  window to report them in, so they print and exit (`fatal` in `engine/sdl.go`).
  Nothing after that point exits: it has somewhere to put a report and someone
  looking at it.
- **Headless runs do not hold a window open.** A game run with SDL's dummy or
  offscreen driver has nobody to dismiss an error screen, so the loop stops on a
  failure instead of waiting. A run that ended on one exits non-zero.

### Objects and tokens

Lumen's objects implement `Method(method string, tok token.Token, args []Object)`.
The token is the method's own name in the source, and it is what lets a method
report a bad argument at the call that made it. Thread it through to every
helper; never build an error without one where one is available.
