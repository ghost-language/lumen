import "ghost:math"
import "ghost:random"
import "lumen:canvas"
import "lumen:color"
import "lumen:window"

import * from 'data'
import * from 'ui'

// Draws a Battle. The rules live in battle.gs and never touch the canvas;
// this reads the battle's state and paints it, which means the combat maths can
// be reasoned about without a window open.
class BattleView {
  constructor(sheet, fonts) {
    this.sheet = sheet
    this.fonts = fonts
  }

  draw(battle) {
    canvas.setColor(color.rgb(10, 10, 18))
    canvas.filledRectangle(0, 0, window.width, window.height)

    this.drawBackdrop(battle)

    // A hit shakes the field, but never the interface: text that jitters while
    // you are trying to read it is worse than no feedback at all.
    canvas.push()

    if (battle.shake > 0) {
      canvas.translate(
        (random.random() - 0.5) * 8 * battle.shake,
        (random.random() - 0.5) * 8 * battle.shake
      )
    }

    this.drawMonsters(battle)

    canvas.pop()

    this.drawParty(battle)
    this.drawBottom(battle)
  }

  // drawBackdrop is a plain horizon: a lighter band of sky over dark ground, so
  // the monsters have something to stand on.
  drawBackdrop(battle) {
    horizon = 260

    canvas.setColor(color.rgb(26, 28, 48))
    canvas.filledRectangle(0, 0, window.width, horizon)

    canvas.setColor(color.rgb(18, 20, 32))
    canvas.filledRectangle(0, horizon, window.width, window.height - horizon)

    canvas.setColor(color.rgb(40, 44, 70))
    canvas.setLineWidth(2)
    canvas.line(0, horizon, window.width, horizon)

    // A few slow stars, seeded off their own index so they stay put.
    for (index = 0; index < 30; index++) {
      x = math.noise(index * 3.7, 0) * window.width
      y = math.noise(0, index * 2.3) * (horizon - 40) + 10

      twinkle = 0.4 + math.sin(battle.time * 1.5 + index) * 0.3

      canvas.setColor(color.rgb(200, 210, 255, twinkle * 0.5))
      canvas.filledCircle(x, y, 1.5)
    }
  }

  drawMonsters(battle) {
    living = battle.monsters
    count = living.length()

    if (count == 0) {
      return null
    }

    spacing = window.width / (count + 1)
    scale = 5
    size = 16 * scale

    for (index = 0; index < count; index++) {
      monster = living[index]

      if (!monster.isAlive()) {
        continue
      }

      x = spacing * (index + 1) - size / 2
      bob = math.sin(battle.time * 2 + index * 1.3) * 4
      y = 96 + bob

      // A struck monster flashes white by being drawn twice: once normally,
      // once tinted, with the tint fading out over a quarter of a second.
      canvas.setColor(color.rgb(0, 0, 0, 0.35))
      canvas.filledEllipse(x + size / 2, 238, size * 0.4, 10)

      canvas.setColor(color.white)
      this.sheet.draw(monster.character * 12, x, y, 0, scale, scale)

      if (monster.flash > 0) {
        canvas.setColor(color.rgb(255, 90, 90, monster.flash * 3))
        this.sheet.draw(monster.character * 12, x, y, 0, scale, scale)
      }

      this.drawMonsterLabel(battle, monster, x + size / 2, 248)
    }
  }

  drawMonsterLabel(battle, monster, centreX, y) {
    canvas.setFont(this.fonts.small)

    width = this.fonts.small.getWidth(monster.name)

    canvas.setColor(theme.text)
    canvas.print(monster.name, centreX - width / 2, y)

    drawBar(centreX - 30, y + 24, 60, 5, monster.healthFraction(), barColor(monster.healthFraction()))

    // Mark whichever monster the target cursor is on.
    if (battle.phase == 'target' and battle.menu != null) {
      if (battle.menu.value() == monster) {
        canvas.setColor(theme.accent)
        canvas.filledPolygon(
          centreX - 6, y - 16,
          centreX + 6, y - 16,
          centreX, y - 6
        )
      }
    }

    canvas.resetFont()
  }

  // drawParty is the row of status boxes along the bottom: one per hero, with
  // the hero currently giving an order picked out.
  drawParty(battle) {
    members = battle.party.members
    count = members.length()

    boxWidth = 184
    boxHeight = 84
    gap = 12
    total = count * boxWidth + (count - 1) * gap
    left = (window.width - total) / 2
    top = 288

    living = battle.party.living()
    active = null

    if (battle.actorIndex < living.length()) {
      if (battle.phase != 'resolve' and battle.phase != 'over') {
        active = living[battle.actorIndex]
      }
    }

    canvas.setFont(this.fonts.small)

    for (index = 0; index < count; index++) {
      member = members[index]
      x = left + index * (boxWidth + gap)

      drawPanel(x, top, boxWidth, boxHeight)

      if (member == active) {
        canvas.setColor(theme.accent)
        canvas.setLineWidth(2)
        canvas.rectangle(x - 3, top - 3, boxWidth + 6, boxHeight + 6)
      }

      nameColor = theme.text

      if (!member.isAlive()) {
        nameColor = theme.bad
      }

      canvas.setColor(nameColor)
      canvas.print(member.name, x + 12, top + 8)

      canvas.setColor(theme.dim)
      canvas.print('Lv ' + member.level.toString(), x + boxWidth - 12 - this.fonts.small.getWidth('Lv ' + member.level.toString()), top + 8)

      fraction = member.healthFraction()

      canvas.setColor(theme.dim)
      canvas.print('HP', x + 12, top + 32)
      drawBar(x + 46, top + 38, 84, 7, fraction, barColor(fraction))

      canvas.setColor(theme.text)
      canvas.print(member.health.toString(), x + 138, top + 32)

      if (member.maxMagic > 0) {
        canvas.setColor(theme.dim)
        canvas.print('MP', x + 12, top + 56)
        drawBar(x + 46, top + 62, 84, 7, member.magic / member.maxMagic, color.rgb(110, 160, 240))

        canvas.setColor(theme.text)
        canvas.print(member.magic.toString(), x + 138, top + 56)
      }

      if (member.flash > 0) {
        canvas.setColor(color.rgb(255, 80, 80, member.flash))
        canvas.filledRectangle(x, top, boxWidth, boxHeight)
      }
    }

    canvas.resetFont()
  }

  // drawBottom is either the message being read or the menu being chosen from.
  // Both sit against the bottom of the screen; the menu is as tall as it needs
  // to be, and the message box is a fixed shape so text never moves as one
  // message follows another.
  drawBottom(battle) {
    left = 24
    width = window.width - 48
    height = 160
    top = window.height - 16 - height

    message = battle.currentMessage()

    if (message != null) {
      drawPanel(left, top, width, height)

      canvas.setFont(this.fonts.body)
      canvas.setColor(theme.text)
      canvas.printf(message, left + 20, top + 22, width - 40, 'left')

      // The prompt pulses so it reads as waiting rather than stuck.
      if (math.sin(battle.time * 6) > 0) {
        canvas.setColor(theme.dim)
        canvas.print('>', left + width - 34, top + height - 38)
      }

      canvas.resetFont()

      return null
    }

    if (battle.phase == 'over') {
      return null
    }

    if (battle.menu != null) {
      menuTop = window.height - 16 - battle.menu.height()

      battle.menu.draw(left, menuTop, this.fonts.body)

      // The hint lines up with the top of the menu it explains, whatever the
      // menu's height turned out to be.
      this.drawHint(battle, left + battle.menu.width + 16, menuTop)
    }
  }

  // drawHint explains whatever the cursor is on, which is where a spell's cost
  // and an item's effect get told to the player.
  drawHint(battle, x, y) {
    row = null

    if (battle.menu != null) {
      row = battle.menu.current()
    }

    if (row == null) {
      return null
    }

    text = hintFor(battle, row)

    if (text == '') {
      return null
    }

    width = window.width - x - 24

    drawPanel(x, y, width, 104)

    canvas.setFont(this.fonts.small)
    canvas.setColor(theme.dim)
    canvas.printf(text, x + 16, y + 18, width - 32, 'left')
    canvas.resetFont()
  }
}

// hintFor is the line of help shown beside the current menu row.
function hintFor(battle, row) {
  switch (battle.phase) {
    case 'command' {
      switch (row.value) {
        case 'fight' { return 'Strike one monster with what you are holding.' }
        case 'spell' { return 'Spend magic on damage, healing, or defence.' }
        case 'item' { return 'Use something from the pack.' }
        case 'guard' { return 'Take the turn defending. Halves the next blow.' }
        case 'run' { return 'Try to leave. The faster the party, the likelier.' }
      }
    }
    case 'spell' {
      spell = findSpell(row.value)

      if (spell != null) {
        return spell.text
      }
    }
    case 'item' {
      found = findItem(row.value)

      if (found != null) {
        return found.text
      }
    }
    case 'target' {
      if (row.value != null) {
        return row.value.name
      }
    }
  }

  return ''
}
