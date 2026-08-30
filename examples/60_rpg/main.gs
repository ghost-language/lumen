// A small top-down RPG built with nothing but Lumen and Ghost.
//
//   move        arrow keys / WASD / gamepad stick or d-pad
//   confirm     space or E (gamepad A)     - talk, choose a menu entry
//   cancel      escape or Q (gamepad B)    - go back one step
//   menu        tab                        - pack, equipment, status, save
//   debug       F1        fullscreen  F11        mute  M
//   quit        escape, from the world with no menu open
//
// Ghost scopes assignment to the function it happens in, so a bare `total = 1`
// inside update() makes a new local rather than changing a global. Everything
// that has to persist between frames therefore lives on `game`, a map, or on a
// class instance.

import "ghost:json"
import "ghost:math"
import "ghost:random"
import "lumen:canvas"
import "lumen:color"
import "lumen:filesystem"
import "lumen:font"
import "lumen:joystick"
import "lumen:lumen"
import "lumen:timer"
import "lumen:window"
import { Spritesheet } from "lumen:image"

import * from 'data'
import * from 'ui'
import Battle from 'battle'
import BattleView from 'battleview'
import Camera from 'camera'
import Dialogue from 'dialogue'
import FieldMenu from 'fieldmenu'
import Hud from 'hud'
import Npc from 'npc'
import Party from 'party'
import Player from 'player'
import Sounds from 'sounds'
import Tilemap from 'tilemap'

// The whole game is laid out for a canvas this size. Nothing is measured
// against the window, so the interface cannot drift as the window changes.
canvasWidth = 800
canvasHeight = 600

// How far the world is magnified while walking around. Whole numbers only:
// a fractional zoom lands 16-pixel tiles on fractions of a pixel and opens a
// seam between every one of them.
fieldZoom = 3

game = {
  mode: 'field',
  showDebug: false,
  message: '',
  messageTime: 0,
  fade: 1,
  encounterAt: 240,
  transition: null,
  pendingMonsters: null
}

function load() {
  window.setTitle('Lumen RPG')

  // Lumen scales this canvas to fill the window, keeping its shape and
  // centring what is left over. The game therefore draws in the same 800x600
  // whatever the window is doing: fullscreen makes everything bigger rather
  // than showing more map with the same tiny text in the corner of it.
  window.setLogicalSize(canvasWidth, canvasHeight)
  window.setMode(startingWidth(), startingHeight())

  filesystem.setIdentity('lumen-rpg')

  game.fonts = { body: font.system(24), small: font.system(19) }

  game.sounds = new Sounds(['blip', 'step', 'chime', 'hit', 'spell', 'fanfare'])
  game.sounds.setVolume(0.7)

  // Water, woodland, fences and buildings block movement. The road layer takes
  // collision away again, which is what makes the bridge over the river a
  // bridge rather than a picture of one.
  game.map = new Tilemap(
    'resources/map.json',
    ['Water', 'Trees', 'Props', 'Structures'],
    ['Paths']
  )

  game.characters = new Spritesheet('resources/characters.png', 16, 16)

  game.party = new Party(['aldric', 'sera', 'nell'])
  game.party.add('herb', 3)
  game.party.add('tonic', 1)
  game.party.add('leather', 1)
  game.party.add('buckler', 1)

  game.player = new Player(game.characters, 25, 25, game.map.tileSize, findHero('aldric').character)
  game.spawn = { x: game.player.x, y: game.player.y }

  game.camera = new Camera(fieldZoom)
  game.camera.setBounds(game.map.pixelWidth, game.map.pixelHeight)

  game.dialogue = new Dialogue(game.fonts.body)
  game.hud = new Hud(game.fonts.body, game.fonts.small)
  game.fieldMenu = new FieldMenu(game.party, game.fonts, game.sounds)
  game.battleView = new BattleView(game.characters, game.fonts)
  game.battle = null

  game.npcs = buildNpcs(game.characters, game.map.tileSize)

  canvas.setBackgroundColor(color.rgb(12, 14, 22))

  console.log('Lumen ' + lumen.version + ' - saves in ' + filesystem.getSaveDirectory())
}

// startingScale opens the window at a whole multiple of the canvas where the
// display has room for one, so the game starts out as big as it can be without
// running off the edge of the screen. The player can go fullscreen from there.
function startingScale() {
  desktop = window.getDesktopDimensions()

  if (desktop[0] >= canvasWidth * 2 + 100 and desktop[1] >= canvasHeight * 2 + 100) {
    return 2
  }

  if (desktop[0] >= canvasWidth * 1.5 + 80 and desktop[1] >= canvasHeight * 1.5 + 80) {
    return 1.5
  }

  return 1
}

function startingWidth() {
  return math.floor(canvasWidth * startingScale())
}

function startingHeight() {
  return math.floor(canvasHeight * startingScale())
}

function buildNpcs(sheet, tileSize) {
  return [
    new Npc(sheet, 22, 22, tileSize, {
      name: 'Wren',
      frame: 0,
      gives: null,
      lines: [
        'You must be the one from the coast road. We do not get many visitors this far in.',
        'Mind the long grass. Things live in it that would rather you did not walk through.',
        'Press tab if you want to check your pack. Everyone forgets.'
      ]
    }),
    new Npc(sheet, 29, 27, tileSize, {
      name: 'Osric',
      frame: 12,
      gives: null,
      lines: [
        'Forty years I have walked this map and never once found the edge of it.',
        'That is because there is no edge. Only a wall you cannot see.',
        'If a fight turns badly, guard and let the cleric work. Dying is expensive.'
      ]
    }),
    new Npc(sheet, 27, 21, tileSize, {
      name: 'Bram',
      frame: 24,
      gives: 'copper',
      lines: [
        'Here. My grandfather left me this and I have never once drawn it in anger.',
        'Put it on someone who will. Check the equip screen.'
      ]
    }),
    new Npc(sheet, 21, 30, tileSize, {
      name: 'Elsie',
      frame: 36,
      gives: 'elixir',
      lines: [
        'Pulled this out of the shallows last spring. Still sealed, somehow.',
        'Save it. You will know the moment when it comes.'
      ]
    })
  ]
}

// =============================================================================
// Frame

function update(dt) {
  // The fade is driven in every mode so the opening fade-in always finishes.
  game.fade = math.max(0, game.fade - dt * 1.5)

  if (game.messageTime > 0) {
    game.messageTime = math.max(0, game.messageTime - dt)
  }

  if (game.mode == 'transition') {
    updateTransition(dt)

    return null
  }

  if (game.mode == 'battle') {
    updateBattle(dt)

    return null
  }

  updateField(dt)
}

function updateField(dt) {
  game.fieldMenu.update(dt)

  if (game.fieldMenu.saveRequested) {
    game.fieldMenu.saveRequested = false

    saveGame()
  }

  game.party.update(dt)

  if (game.fieldMenu.open) {
    return null
  }

  if (game.dialogue.active) {
    game.dialogue.update(dt)

    return null
  }

  game.player.update(dt, game.map, game.sounds)

  for (index = 0; index < game.npcs.length(); index++) {
    game.npcs[index].update(dt)
  }

  game.camera.update(dt)
  game.camera.follow(game.player, dt)

  checkEncounter()
}

function updateBattle(dt) {
  game.battle.update(dt)

  // The battle stays on its closing message until it has been read; only then
  // does the world come back.
  if (game.battle.phase == 'over') {
    if (game.battle.currentMessage() == null) {
      endBattle()
    }
  }
}

// =============================================================================
// Encounters

// checkEncounter rolls once the player has walked far enough. Using distance
// rather than a timer means standing still is safe, which is what makes a town
// or a corner feel like a place to stop and think.
function checkEncounter() {
  if (game.player.walked < game.encounterAt) {
    return null
  }

  game.player.walked = 0
  game.encounterAt = random.random(160, 420)

  beginTransition('battle', rollEncounter())
}

// rollEncounter picks a group by weight from those the party is strong enough
// to meet. A weight table is far easier to tune than a chain of thresholds, and
// filtering by level first means a new party never draws a late-game group.
function rollEncounter() {
  level = averagePartyLevel()

  available = []
  total = 0

  for (index = 0; index < encounters.length(); index++) {
    group = encounters[index]

    if (group.level > level) {
      continue
    }

    available.push(group)
    total = total + group.weight
  }

  if (available.length() == 0) {
    return encounters[0].monsters
  }

  roll = random.random() * total

  for (index = 0; index < available.length(); index++) {
    roll = roll - available[index].weight

    if (roll <= 0) {
      return available[index].monsters
    }
  }

  return available[0].monsters
}

function averagePartyLevel() {
  total = 0

  for (index = 0; index < game.party.members.length(); index++) {
    total = total + game.party.members[index].level
  }

  return total / game.party.members.length()
}

function startBattle(monsterIds) {
  game.mode = 'battle'
  game.dialogue.active = false

  game.battle = new Battle(game.party, monsterIds, game.fonts, game.sounds)
}

// =============================================================================
// Transitions
//
// A battle never simply appears. The screen flashes, the camera is knocked, and
// bars close over the world from alternating sides; the fight is built behind
// them and fades up out of the black. The same bars run backwards on the way
// out. None of it is decoration: the flash is what makes an encounter land as
// an interruption, and the bars are what cover the frame where one screen's
// worth of state is thrown away and another is built.
//
// The world keeps being drawn underneath all of it. That is what lets the same
// code run in both directions, and it is why the wipe is drawn after the field
// and before the fade rather than in place of anything.

// transitionTime is how long a wipe takes, in seconds. Long enough to read as
// deliberate, short enough that a player walking into fights all afternoon
// never waits on it.
transitionTime = 0.85

// barCount is how many bars the screen is cut into. More reads as a finer
// shutter; fewer as a heavier one.
barCount = 12

function beginTransition(into, monsterIds) {
  game.mode = 'transition'
  game.transition = { into: into, time: 0 }
  game.pendingMonsters = monsterIds

  if (into == 'battle') {
    game.dialogue.active = false

    game.sounds.play('spell')
    game.camera.knock(2, 0.35)
  }
}

function updateTransition(dt) {
  game.transition.time = game.transition.time + dt

  progress = math.min(1, game.transition.time / transitionTime)

  // The shake started by the flash still has to be driven, since the rest of
  // the world is deliberately frozen while the wipe runs.
  game.camera.update(dt)

  if (progress < 1) {
    return null
  }

  if (game.transition.into == 'battle') {
    startBattle(game.pendingMonsters)

    // The fight arrives behind the closed bars and fades up out of them.
    game.fade = 1
  } else {
    game.mode = 'field'
  }

  game.transition = null
  game.pendingMonsters = null
}

// drawTransition paints over the world, which is still being drawn underneath
// it: the wipe covers a frozen field on the way in and a restored one on the
// way out.
function drawTransition() {
  elapsed = game.transition.time
  progress = math.min(1, elapsed / transitionTime)

  if (game.transition.into == 'field') {
    progress = 1 - progress
  }

  drawTransitionFlash(elapsed)
  drawTransitionBars(progress)
}

// drawTransitionFlash is three quick washes of white over the first half of the
// wipe, fading as they go.
function drawTransitionFlash(elapsed) {
  if (game.transition.into != 'battle') {
    return null
  }

  if (elapsed > 0.42) {
    return null
  }

  strength = math.abs(math.sin(elapsed * 22)) * (1 - elapsed / 0.42)

  canvas.setColor(color.rgb(255, 255, 255, strength * 0.75))
  canvas.filledRectangle(0, 0, window.width, window.height)
}

// drawTransitionBars closes black bars in from alternating sides. Alternating
// them is the whole trick: bars that all travel the same way read as a curtain,
// and bars that interlock read as a shutter.
function drawTransitionBars(progress) {
  // The bars only start moving once the flashes are under way, so the two
  // halves of the wipe overlap rather than queueing behind one another.
  covered = math.clamp((progress - 0.3) / 0.7, 0, 1)

  if (covered <= 0) {
    return null
  }

  height = window.height / barCount
  width = covered * window.width

  canvas.setColor(color.rgb(0, 0, 0))

  for (index = 0; index < barCount; index++) {
    x = 0

    if (index % 2 == 1) {
      x = window.width - width
    }

    // The extra pixel of height closes the seam a fractional bar height would
    // otherwise leave between one bar and the next.
    canvas.filledRectangle(x, index * height, width, height + 1)
  }
}

// endBattle returns to the world, handling the three ways a fight can finish.
function endBattle() {
  outcome = game.battle.result

  game.battle = null
  game.player.walked = 0

  // The world comes back behind the same bars it left behind, opening this time.
  beginTransition('field', null)

  if (outcome == 'victory') {
    game.sounds.play('fanfare')

    return null
  }

  if (outcome == 'defeat') {
    // Losing costs half the purse and sends the party back where they started,
    // which is the traditional bargain: a setback, never a lost save.
    game.party.gold = math.floor(game.party.gold / 2)
    game.party.restAll()

    game.player.x = game.spawn.x
    game.player.y = game.spawn.y

    // The camera goes with them. Easing across half the map behind the closing
    // bars would show the whole world flying past when they open again.
    game.camera.snapTo(game.player)

    notify('The party wakes where they began. Half the purse is gone.')
  }
}

// =============================================================================
// Input

// keypressed fires once per physical press. Menus and battles poll the shared
// helpers in ui.gs instead, so this only handles what is global or what
// belongs to walking around.
function keypressed(key) {
  switch (key) {
    case 'F1' { game.showDebug = !game.showDebug }
    case 'F5' { saveGame() }
    case 'F9' { loadGame() }
    case 'F11' { window.toggleFullscreen() }
    case 'M' { toggleSound() }
  }

  // A battle, and the wipe on either side of it, own the keyboard entirely.
  if (game.mode != 'field') {
    return null
  }

  switch (key) {
    case 'Tab' { toggleMenu() }
    case 'Escape' { handleEscape() }
    case 'Space' { interact() }
    case 'E' { interact() }
  }
}

function toggleMenu() {
  if (game.dialogue.active) {
    return null
  }

  game.sounds.play('blip')
  game.fieldMenu.toggle()
}

function handleEscape() {
  if (game.dialogue.active) {
    game.dialogue.active = false

    return null
  }

  // The field menu handles its own cancel key, so escape only quits from a
  // world with nothing open on top of it.
  if (game.fieldMenu.open) {
    return null
  }

  lumen.quit()
}

function toggleSound() {
  if (game.sounds.toggle()) {
    notify('Sound on')
  } else {
    notify('Sound off')
  }
}

function wheelmoved(x, y) {
  //
}

// Gamepad buttons are polled rather than delivered as events, so the field
// confirm is checked here where a press can be seen as an edge.
function gamepadCheck() {
  if (game.mode != 'field') {
    return null
  }

  if (game.fieldMenu.open) {
    return null
  }

  if (joystick.isConnected(1) and joystick.wasPressed(1, 'a')) {
    interact()
  }

  if (joystick.isConnected(1) and joystick.wasPressed(1, 'start')) {
    toggleMenu()
  }
}

function interact() {
  if (game.fieldMenu.open) {
    return null
  }

  if (game.dialogue.active) {
    game.dialogue.advance(game.sounds)

    return null
  }

  point = game.player.facingPoint()
  npc = nearestNpc(point.x, point.y, 14)

  if (npc == null) {
    return null
  }

  game.dialogue.start(npc.name, npc.talk())
  game.sounds.play('blip')

  // A giver hands its item over the first time it is spoken to.
  if (npc.gives != null and !npc.given) {
    npc.given = true

    game.party.add(npc.gives, 1)
    game.party.gold = game.party.gold + 15

    game.sounds.play('chime')

    notify('Received ' + findItem(npc.gives).name)
  }
}

function nearestNpc(x, y, radius) {
  found = null

  for (index = 0; index < game.npcs.length(); index++) {
    npc = game.npcs[index]

    if (npc.isNear(x, y, radius)) {
      found = npc
    }
  }

  return found
}

function notify(message) {
  game.message = message
  game.messageTime = 2.5
}

// =============================================================================
// Saving

function saveGame() {
  opened = []

  for (index = 0; index < game.npcs.length(); index++) {
    if (game.npcs[index].given) {
      opened.push(index)
    }
  }

  filesystem.write('slot1.json', json.encode({
    x: game.player.x,
    y: game.player.y,
    facing: game.player.facing,
    party: game.party.toSave(),
    opened: opened
  }))

  notify('Saved')
  game.sounds.play('chime')
}

function loadGame() {
  contents = filesystem.read('slot1.json')

  if (contents == null) {
    notify('No save yet')

    return null
  }

  save = json.decode(contents)

  game.player.x = save.x
  game.player.y = save.y
  game.player.facing = save.facing

  game.party.loadSave(save.party)

  for (index = 0; index < game.npcs.length(); index++) {
    game.npcs[index].given = false
  }

  // An empty list survives a trip through JSON as null, so every list read back
  // out of a save is checked before it is walked. A game that has been saved
  // before anyone handed anything over would otherwise fail to load at all.
  opened = save.opened

  if (opened != null) {
    for (index = 0; index < opened.length(); index++) {
      game.npcs[opened[index]].given = true
    }
  }

  notify('Loaded')
  game.sounds.play('blip')
}

// =============================================================================
// Drawing

function draw() {
  gamepadCheck()

  if (game.mode == 'battle') {
    game.battleView.draw(game.battle)

    drawNotification()
    drawFade()

    return null
  }

  // World, seen through the camera.
  game.camera.attach()

  game.map.draw(game.camera.visibleTiles(game.map.tileSize))

  drawInteractionHint()
  drawSortedCharacters()

  game.camera.detach()

  // Interface, in screen space. The field menu dims the world and draws its own
  // party strip, so the walking-around HUD steps out of its way.
  if (!game.fieldMenu.open) {
    game.hud.draw(game.party, game.showDebug)
  }

  game.dialogue.draw()
  game.fieldMenu.draw()

  if (game.transition != null) {
    drawTransition()
  }

  drawNotification()
  drawFade()
}

function drawSortedCharacters() {
  actors = [game.player]

  for (index = 0; index < game.npcs.length(); index++) {
    actors.push(game.npcs[index])
  }

  // A short insertion sort: with a handful of actors it costs nothing, and it
  // keeps the ordering stable frame to frame.
  //
  // Ghost's `and` evaluates both sides, so the bounds check cannot share a
  // condition with the comparison: at j = -1 the index would still be read, and
  // a negative index reads as null rather than raising.
  for (i = 1; i < actors.length(); i++) {
    current = actors[i]
    j = i - 1
    placed = false

    while (!placed) {
      if (j < 0) {
        placed = true
      } else if (actors[j].y <= current.y) {
        placed = true
      } else {
        actors[j + 1] = actors[j]
        j = j - 1
      }
    }

    actors[j + 1] = current
  }

  for (index = 0; index < actors.length(); index++) {
    actors[index].draw()
  }
}

// drawInteractionHint marks the NPC the player is facing, so it is obvious what
// pressing space will do.
function drawInteractionHint() {
  if (game.dialogue.active) {
    return null
  }

  if (game.fieldMenu.open) {
    return null
  }

  point = game.player.facingPoint()
  npc = nearestNpc(point.x, point.y, 14)

  if (npc == null) {
    return null
  }

  bob = math.sin(timer.time * 6) * 1.5

  canvas.setColor(color.rgb(255, 235, 140))
  canvas.filledPolygon(
    npc.centerX() - 3, npc.centerY() - 14 + bob,
    npc.centerX() + 3, npc.centerY() - 14 + bob,
    npc.centerX(), npc.centerY() - 9 + bob
  )
}

function drawNotification() {
  if (game.messageTime <= 0) {
    return null
  }

  canvas.setFont(game.fonts.body)

  width = game.fonts.body.getWidth(game.message) + 32
  x = (window.width - width) / 2

  // The banner fades out over its last half second rather than vanishing.
  alpha = math.min(1, game.messageTime / 0.5)

  canvas.setColor(color.rgb(20, 22, 36, alpha * 0.9))
  canvas.filledRectangle(x, 70, width, 40)

  canvas.setColor(color.rgb(255, 255, 255, alpha))
  canvas.print(game.message, x + 16, 77)

  canvas.resetFont()
}

function drawFade() {
  if (game.fade <= 0) {
    return null
  }

  canvas.setColor(color.rgb(0, 0, 0, game.fade))
  canvas.filledRectangle(0, 0, window.width, window.height)
}
