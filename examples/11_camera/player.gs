import "lumen:keyboard"
import { Image } from "lumen:image"

class Player {
  constructor() {
    this.x = 0
    this.y = 0
    this.width = 16
    this.height = 16
    this.speed = 2
    this.sprite = new Image('resources/player.png')
  }

  update() {
    if (keyboard.isDown('s')) {
      this.y = this.y + this.speed
    }

    if (keyboard.isDown('w')) {
      this.y = this.y - this.speed
    }

    if (keyboard.isDown('a')) {
      this.x = this.x - this.speed
    }

    if (keyboard.isDown('d')) {
      this.x = this.x + this.speed
    }
  }

  draw() {
    this.sprite.draw(this.x, this.y)
  }
}