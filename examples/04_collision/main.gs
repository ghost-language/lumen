import "lumen:canvas"
import Player from 'player'
import Enemy from 'enemy'

player = new Player()
enemy = new Enemy()

function update() {
  enemy.update()
  player.update()

  player.onCollision(enemy)
}

function draw() {
  canvas.scale(2)

  enemy.draw()
  player.draw()
}