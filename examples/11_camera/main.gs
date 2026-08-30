import Camera from 'camera'
import Player from 'player'
import World from 'world'

camera = new Camera()
player = new Player()
world = new World()

function update() {
  world.update()
  
  player.update()
  camera.update()
  camera.follow(player)
}

function draw() {
  camera.draw()
  world.draw()

  player.draw()
}