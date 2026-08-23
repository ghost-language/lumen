import { Image } from "lumen:image"

class Spritesheet {
  constructor(path, size) {
    this.sprite = new Image(path)
    this.size = size
  }

  draw(x, y, frame) {
    columns = this.sprite.getWidth() / this.size
    
    sx = (frame % columns) * this.size
    sy = (frame / columns).floor() * this.size

    this.sprite.clip(sx, sy, this.size).draw(x, y)
  }
}