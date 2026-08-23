import "lumen:keyboard"
import { Spritesheet, Animation } from "lumen:image"

class Player {
  constructor(x = 0, y = 0) {
    this.x = x
    this.y = y
    this.isMoving = false
    this.previousX = x
    this.previousY = y
    this.speed = 1
    this.size = 16
    this.direction = 'down'

    // The sheet slices the image into 16x16 frames once, up front. Frames are
    // numbered left to right, top to bottom, starting at zero.
    this.sheet = new Spritesheet('resources/characters.png', this.size)

    // Each animation holds a frame for a number of *seconds*, so the walk cycle
    // runs at the same speed whatever frame rate the machine manages. A
    // duration of zero makes a still pose out of the first frame.
    this.animations = {
      up: new Animation(this.sheet, [6, 7], 0.2),
      right: new Animation(this.sheet, [8, 9], 0.2),
      down: new Animation(this.sheet, [4, 5], 0.2),
      left: new Animation(this.sheet, [10, 11], 0.2),

      up_idle: new Animation(this.sheet, [1], 0),
      right_idle: new Animation(this.sheet, [2], 0),
      down_idle: new Animation(this.sheet, [0], 0),
      left_idle: new Animation(this.sheet, [3], 0),
    }
  }

  // current returns the animation the player's direction and movement call for.
  current() {
    if (this.isMoving) {
      return this.animations[this.direction]
    }

    return this.animations[this.direction + '_idle']
  }

  update() {
    this.previousX = this.x
    this.previousY = this.y

    if (keyboard.isDown('a')) {
      this.x = this.x - this.speed
      this.direction = 'left'
    }

    if (keyboard.isDown('d')) {
      this.x = this.x + this.speed
      this.direction = 'right'
    }

    if (keyboard.isDown('s')) {
      this.y = this.y + this.speed
      this.direction = 'down'
    }

    if (keyboard.isDown('w')) {
      this.y = this.y - this.speed
      this.direction = 'up'
    }

    // Determine if the player is moving or not based on the current
    // and previous positions update `isMoving` accordingly.
    if (this.x != this.previousX or this.y != this.previousY) {
      this.isMoving = true
    } else {
      this.isMoving = false
    }

    // update() with no argument advances by the time the last frame took.
    this.current().update()
  }

  draw() {
    this.current().draw(this.x, this.y)
  }
}