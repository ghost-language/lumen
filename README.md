# Lumen
A lightweight 2D game engine for Ghost.

## Status
Currently in development.

## Requirements
Below are the requirements necessary if working with the Lumen source code directly:

- SDL2
- SDL2_image
- SDL2_ttf

Requirements can be installed through brew on Mac:

```bash
brew install sdl2 sdl2_image sdl2_ttf
```

## Notes
Lumen will come with a set of built-in modules to interact with and configure various aspects of the framework and game instance.

- [ ] Audio
- [ ] Lumen
- [x] Window
- [x] Canvas
- [x] Keyboard
- [x] Mouse
- [ ] Controller

## Game Loop
At the heart of every Lumen game, the game loop drives the lifecycle of the application. This looping mechanism is responsible for setting the initial state, collecting and processing input, updating state, and rendering graphics to the screen.

In the context of Lumen's game loop, the sequence is as follows:

1. `load()`
2. `update()`
3. `draw()`

### `load()`
Lumen begins with the `load()` function, which is invoked once at the very start. Primarily, this function is used to set the initial state of the game. For example, you might load assets, initialize variables, or set the screen size.

```typescript
function load() {
  state.player = {
    x: 0,
    y: 0,
    speed: 5,
    sprite: image.load('player.png')
  }
}
```

### `update()`
After initializing the game state, `update()` is consistently executed per game frame. This function should handle game logic controlling character behaviors, AI progress, physics, and other progression-related tasks.

```typescript
function update() {
  if (keyboard.isDown('right')) {
    state.player.x = state.player.x + state.player.speed
  }
}
```

### `draw()`
The `draw()` function is the final part of the cycle, executed after `update()`. The sole purpose of `draw()` is rendering; all calls to render the current game state onto the screen should be done here.

```typescript
function draw() {
  state.player.sprite.draw(state.player.x, state.player.y)
}
```

After `draw()` is executed, the game loop returns to `update()` and the cycle repeats until the game is closed. These three functions are the core of Lumen's game loop, and are the only functions that are required to be defined in a Lumen game.

## Modules
### Canvas
#### Methods
- `canvas.rectangle()`
- `canvas.filledRectangle()`
- `canvas.circle()`
- `canvas.filledCircle()`
- `canvas.line()`
- `canvas.point()`
- `canvas.clear()`
- `canvas.setColor()`
- `canvas.setFont()`
- `canvas.resetFont()`
- `canvas.print()`
- `canvas.scale()`
- `canvas.translate()`

### Color
#### Methods
- `color.rgb()`
- `color.hex()`

#### Properties
- `color.black`
- `color.white`

### Font
#### Methods
- `font.load()`

### Image
#### Methods
- `image.load()`

### Keyboard
#### Methods
- `keyboard.isDown()`
- `keyboard.isUp()`
- `keyboard.wasPressed()`
- `keyboard.wasReleased()`

### Mouse
#### Methods
- `mouse.showCursor()`
- `mouse.hideCursor()`
- `mouse.isButtonDown()`
- `mouse.isButtonUp()`
- `mouse.wasButtonPressed()`
- `mouse.wasButtonReleased()`

#### Properties
- `mouse.x`
- `mouse.y`

### Window
#### Methods
- `window.setTitle()`

#### Properties
- `window.fps`
- `window.width`
- `window.height`

## Objects
### Color
Represents a color.

### Font
Represents a font.

#### Methods
- `print()`

### Image
Represents an image.

#### Methods
- `draw()`
- `clip()`
- `getWidth()` (need to update ghost to include object properties)
- `getHeight()` (need to update ghost to include object properties)