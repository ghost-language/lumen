import Spritesheet from 'spritesheet'

class Map {
  constructor(spritesheet, size) {
    this.layers = []
    this.spritesheet = new Spritesheet(spritesheet, size)
  }

  addLayer(layer) {
    this.layers.push(layer)
  }

  draw() {
    for (i = 0; i < this.layers.length(); i++) {
      for (row = 0; row < this.layers[i].length(); row++) {
        for (col = 0; col < this.layers[i][row].length(); col++) {
          tile = this.layers[i][row][col]
          x = col * this.spritesheet.size
          y = row * this.spritesheet.size

          this.spritesheet.draw(x, y, tile)
        }
      }
    }
  }
}