import "lumen:canvas"
import "lumen:window"
import World from 'world'

world = new World()

function load() {
  window.setTitle("Conway's Game of Life")
  
  world.load()
}

function update() {
  world.update()
}

function draw() {
  canvas.scale(2)
  canvas.print('FPS: ' + window.fps.toString(), 10, 10)
  
  world.draw()
}