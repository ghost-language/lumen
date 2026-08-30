import "ghost:file"
import "ghost:json"
import Spritesheet from 'spritesheet'

class Map {
  constructor(filepath, root = 'resources') {
    this.root = root
    this.map = this.loadMap(filepath)
    this.tilesets = this.loadTilesets()

    console.log('map loaded')
  }

  loadMap(filepath) {
    data = file.read(filepath)

    return json.decode(data)
  }

  loadTilesets() {
    tilesets = {}

    for (i = 0; i < this.map.tilesets.length(); i++) {
      tileset = this.map.tilesets[i]
      tilesets[tileset.name] = {
        spritesheet: new Spritesheet(this.root + '/' + tileset.image, tileset.tilewidth),
        firstgid: tileset.firstgid,
        lastgid: tileset.firstgid + tileset.tilecount,
        size: tileset.tilewidth
      }
    }

    console.log('loaded tilesets')
    console.log(tilesets)

    return tilesets
  }

  update() {
    // 
  }

  draw() {
    for (i = 0; i < this.map.layers.length(); i++) {
      layer = this.map.layers[i]

      if (layer.type == 'tilelayer') {
        this.drawTileLayer(layer)
      }
    }
  }

  drawTileLayer(layer) {
    for (y = 0; y < layer.height; y++) {
      for (x = 0; x < layer.width; x++) {
        index = x + y * layer.width
        gid = layer.data[index]

        if (gid == 0) {
          continue
        }

        tileset = this.getTileset(gid)
        
        frame = gid - tileset.firstgid

        tileset.spritesheet.draw(frame, x * tileset.size, y * tileset.size)
      }
    }
  }

  getTileset(gid) {
    foundTileset = null
    for (tileset in this.tilesets) {
      if (gid >= tileset.firstgid and gid <= tileset.lastgid) {
        foundTileset = tileset
      }
    }

    return foundTileset
  }

  // function drawTileLayer(layer) {
  //   for (y = 0; y < layer.height; y++) {
  //     for (x = 0; x < layer.width; x++) {
  //       index = x + y * layer.width
  //       gid = layer.data[index]

  //       if (gid == 0) {
  //         continue
  //       }

  //       tileset = this.getTileset(gid)

  //       if (tileset == null) {
  //         continue
  //       }

  //       tile = tileset.spritesheet.getTile(gid - tileset.firstgid)

  //       if (tile == null) {
  //         continue
  //       }

  //       tile.draw(x * tileset.spritesheet.tilewidth, y * tileset.spritesheet.tileheight)
  //     }
  //   }
  // }
}