import "ghost:math"
import "lumen:canvas"
import "lumen:color"
import "lumen:window"

import * from 'data'
import * from 'ui'
import Menu from 'ui'

// The menu the player opens in the field: the pack, equipment, the party's
// stats, and saving. Each screen is a Menu plus a detail panel beside it, and
// `screen` says which one is open. Cancel walks back one step at a time.
class FieldMenu {
  constructor(party, fonts, sounds) {
    this.party = party
    this.fonts = fonts
    this.sounds = sounds

    this.open = false
    this.screen = 'root'
    this.menu = null
    this.hero = null
    this.slot = null
    this.item = null
    this.notice = ''
    this.noticeTime = 0

    // Set when the player asks to save, and read and cleared by the game, which
    // is the only part that knows how a save is written.
    this.saveRequested = false
  }

  toggle() {
    if (this.open) {
      this.close()

      return false
    }

    this.open = true

    this.showRoot()

    return true
  }

  close() {
    this.open = false
    this.screen = 'root'
    this.menu = null
  }

  say(text) {
    this.notice = text
    this.noticeTime = 2.5
  }

  // ---- screens

  showRoot() {
    this.screen = 'root'
    this.hero = null
    this.slot = null

    this.menu = new Menu([
      row('Pack', 'items', this.party.pack.length().toString(), true),
      row('Equip', 'equip', null, true),
      row('Status', 'status', null, true),
      row('Rest', 'rest', null, true),
      row('Save', 'save', null, true),
      row('Close', 'close', null, true)
    ], { width: 250, title: 'Menu' })
  }

  showItems() {
    this.screen = 'items'

    rows = []

    for (index = 0; index < this.party.pack.length(); index++) {
      entry = this.party.pack[index]
      found = findItem(entry.id)

      if (found == null) {
        continue
      }

      detail = 'x' + entry.count.toString()

      rows.push(row(found.name, found.id, detail, true))
    }

    this.menu = new Menu(rows, { width: 330, title: 'Pack' })
  }

  showItemTarget() {
    this.screen = 'itemTarget'

    rows = []

    for (index = 0; index < this.party.members.length(); index++) {
      member = this.party.members[index]

      rows.push(row(
        member.name,
        member,
        member.health.toString() + '/' + member.maxHealth.toString(),
        true
      ))
    }

    this.menu = new Menu(rows, { width: 290, title: 'Use on' })
  }

  showHeroes(next) {
    this.screen = next

    rows = []

    for (index = 0; index < this.party.members.length(); index++) {
      member = this.party.members[index]

      rows.push(row(member.name, member, member.role, true))
    }

    this.menu = new Menu(rows, { width: 290, title: 'Who' })
  }

  showSlots() {
    this.screen = 'equipSlot'

    this.menu = new Menu([
      row('Weapon', 'weapon', this.hero.equippedName('weapon'), true),
      row('Armour', 'armour', this.hero.equippedName('armour'), true),
      row('Shield', 'shield', this.hero.equippedName('shield'), true)
    ], { width: 380, title: this.hero.name })
  }

  // showSlotItems lists what is in the pack for this slot, plus an option to
  // take off whatever is worn now.
  showSlotItems() {
    this.screen = 'equipItem'

    rows = []

    if (this.hero.equipment[this.slot] != null) {
      rows.push(row('Take off', 'none', this.hero.equippedName(this.slot), true))
    }

    stock = this.party.packOfKind(this.slot)

    for (index = 0; index < stock.length(); index++) {
      entry = stock[index]
      found = findItem(entry.id)

      change = this.party.previewEquip(this.hero, this.slot, found.id)

      rows.push(row(found.name, found.id, changeLabel(change), true))
    }

    this.menu = new Menu(rows, { width: 380, title: slotTitle(this.slot) })
  }

  showStatus() {
    this.screen = 'statusSheet'
    this.menu = null
  }

  // ---- input

  update(dt) {
    if (this.noticeTime > 0) {
      this.noticeTime = math.max(0, this.noticeTime - dt)
    }

    if (!this.open) {
      return null
    }

    if (this.screen == 'statusSheet') {
      if (pressedConfirm() or pressedCancel()) {
        this.showHeroes('status')
      }

      return null
    }

    if (this.menu == null) {
      return null
    }

    action = this.menu.update()

    if (action == null) {
      return null
    }

    if (action == 'cancel') {
      this.sounds.play('blip')
      this.back()

      return null
    }

    this.sounds.play('blip')
    this.confirm()
  }

  back() {
    switch (this.screen) {
      case 'root' { this.close() }
      case 'items' { this.showRoot() }
      case 'itemTarget' { this.showItems() }
      case 'status' { this.showRoot() }
      case 'equip' { this.showRoot() }
      case 'equipSlot' { this.showHeroes('equip') }
      case 'equipItem' { this.showSlots() }
    }
  }

  confirm() {
    chosen = this.menu.value()

    switch (this.screen) {
      case 'root' { this.confirmRoot(chosen) }
      case 'items' {
        this.item = chosen

        found = findItem(chosen)

        if (found.kind == 'potion') {
          this.showItemTarget()

          return null
        }

        this.say(found.text)
      }
      case 'itemTarget' {
        message = this.party.useItem(this.item, chosen)

        if (message == null) {
          this.say('It would do nothing right now.')

          return null
        }

        this.sounds.play('chime')
        this.say(message)
        this.showItems()
      }
      case 'status' {
        this.hero = chosen

        this.showStatus()
      }
      case 'equip' {
        this.hero = chosen

        this.showSlots()
      }
      case 'equipSlot' {
        this.slot = chosen

        this.showSlotItems()
      }
      case 'equipItem' { this.confirmEquip(chosen) }
    }
  }

  confirmRoot(chosen) {
    switch (chosen) {
      case 'items' { this.showItems() }
      case 'equip' { this.showHeroes('equip') }
      case 'status' { this.showHeroes('status') }
      case 'rest' {
        this.party.restAll()
        this.sounds.play('chime')
        this.say('The party rests. Everyone is restored.')
      }
      case 'save' {
        this.saveRequested = true

        this.close()
      }
      case 'close' { this.close() }
    }
  }

  confirmEquip(chosen) {
    // Every route into this screen sets both, but guarding costs nothing and
    // keeps a half-finished navigation from raising rather than doing nothing.
    if (this.hero == null or this.slot == null) {
      this.showRoot()

      return null
    }

    if (chosen == 'none') {
      this.party.equip(this.hero, this.slot, null)
      this.say(this.hero.name + ' puts it away.')
      this.showSlots()

      return null
    }

    if (!this.party.equip(this.hero, this.slot, chosen)) {
      this.say('It is not in the pack.')

      return null
    }

    this.sounds.play('chime')
    this.say(this.hero.name + ' equips ' + findItem(chosen).name + '.')
    this.showSlots()
  }

  // ---- drawing

  draw() {
    if (this.noticeTime > 0) {
      this.drawNotice()
    }

    if (!this.open) {
      return null
    }

    canvas.setColor(color.rgb(0, 0, 0, 0.55))
    canvas.filledRectangle(0, 0, window.width, window.height)

    if (this.screen == 'statusSheet') {
      this.drawStatusSheet()

      return null
    }

    if (this.menu == null) {
      return null
    }

    x = 48
    y = 112

    this.menu.draw(x, y, this.fonts.body)
    this.drawDetail(x + this.menu.width + 20, y)
    this.drawPartyStrip()
  }

  // drawDetail is the panel beside the menu describing the current row.
  drawDetail(x, y) {
    row = this.menu.current()

    if (row == null) {
      return null
    }

    width = window.width - x - 48

    if (this.screen == 'equipItem') {
      this.drawEquipDetail(x, y, width, row)

      return null
    }

    text = ''

    switch (this.screen) {
      case 'items' {
        found = findItem(row.value)

        if (found != null) {
          text = found.text
        }
      }
      case 'root' { text = rootHint(row.value) }
    }

    if (text == '') {
      return null
    }

    drawPanel(x, y, width, 116)

    canvas.setFont(this.fonts.small)
    canvas.setColor(theme.dim)
    canvas.printf(text, x + 16, y + 20, width - 32, 'left')
    canvas.resetFont()
  }

  // drawEquipDetail shows what the highlighted piece would do to the hero's
  // numbers, which is the question the player is actually asking.
  drawEquipDetail(x, y, width, row) {
    drawPanel(x, y, width, 200)

    canvas.setFont(this.fonts.small)

    if (row.value == 'none') {
      canvas.setColor(theme.dim)
      canvas.printf('Remove what is equipped and put it back in the pack.', x + 16, y + 20, width - 32, 'left')
      canvas.resetFont()

      return null
    }

    found = findItem(row.value)

    if (found == null) {
      canvas.resetFont()

      return null
    }

    canvas.setColor(theme.accent)
    canvas.print(found.name, x + 16, y + 16)

    canvas.setColor(theme.dim)
    canvas.printf(found.text, x + 16, y + 46, width - 32, 'left')

    change = this.party.previewEquip(this.hero, this.slot, found.id)

    canvas.setColor(theme.dim)
    canvas.print('Attack', x + 16, y + 120)
    canvas.print('Defence', x + 16, y + 150)

    this.drawChange(x + 110, y + 120, this.hero.attackPower(), change.attack)
    this.drawChange(x + 110, y + 150, this.hero.defencePower(), change.defence)

    canvas.resetFont()
  }

  // drawChange prints "12 -> 19" with the arrow coloured by whether the swap is
  // an improvement.
  drawChange(x, y, current, delta) {
    canvas.setColor(theme.text)
    canvas.print(current.toString(), x, y)

    if (delta == 0) {
      return null
    }

    tint = theme.good

    if (delta < 0) {
      tint = theme.bad
    }

    canvas.setColor(tint)
    canvas.print('-> ' + (current + delta).toString(), x + 52, y)
  }

  // drawPartyStrip keeps everyone's health visible while menus are open, so a
  // player never has to leave a screen to check whether they need to heal.
  drawPartyStrip() {
    canvas.setFont(this.fonts.small)

    top = window.height - 100
    boxWidth = 220
    boxHeight = 80
    gap = 16
    total = this.party.members.length() * boxWidth + (this.party.members.length() - 1) * gap
    left = (window.width - total) / 2

    for (index = 0; index < this.party.members.length(); index++) {
      member = this.party.members[index]
      x = left + index * (boxWidth + gap)

      drawPanel(x, top, boxWidth, boxHeight)

      canvas.setColor(theme.text)
      canvas.print(member.name, x + 12, top + 8)

      canvas.setColor(theme.dim)
      label = 'Lv ' + member.level.toString()
      canvas.print(label, x + boxWidth - 12 - this.fonts.small.getWidth(label), top + 8)

      fraction = member.healthFraction()

      drawBar(x + 12, top + 36, boxWidth - 24, 7, fraction, barColor(fraction))

      canvas.setColor(theme.dim)
      canvas.print(member.health.toString() + '/' + member.maxHealth.toString(), x + 12, top + 50)

      if (member.maxMagic > 0) {
        magic = 'MP ' + member.magic.toString()
        canvas.setColor(color.rgb(130, 170, 240))
        canvas.print(magic, x + boxWidth - 12 - this.fonts.small.getWidth(magic), top + 50)
      }
    }

    canvas.setColor(theme.accent)
    gold = this.party.gold.toString() + ' gold'
    canvas.print(gold, window.width - 48 - this.fonts.small.getWidth(gold), 48)

    canvas.resetFont()
  }

  drawStatusSheet() {
    x = 110
    y = 84
    width = window.width - 220

    drawPanel(x, y, width, 400)

    canvas.setFont(this.fonts.body)

    canvas.setColor(theme.accent)
    canvas.print(this.hero.name, x + 28, y + 22)

    canvas.setColor(theme.dim)
    canvas.print(this.hero.role + '  -  Level ' + this.hero.level.toString(), x + 28, y + 56)

    canvas.setColor(color.rgb(70, 88, 128))
    canvas.line(x + 24, y + 92, x + width - 24, y + 92)

    canvas.setFont(this.fonts.small)

    lines = [
      ['Health', this.hero.health.toString() + ' / ' + this.hero.maxHealth.toString()],
      ['Magic', this.hero.magic.toString() + ' / ' + this.hero.maxMagic.toString()],
      ['Attack', this.hero.attackPower().toString()],
      ['Defence', this.hero.defencePower().toString()],
      ['Agility', this.hero.agility.toString()],
      ['Experience', this.hero.experience.toString()],
      ['Next level', nextLevelText(this.hero)]
    ]

    for (index = 0; index < lines.length(); index++) {
      rowY = y + 112 + index * 30

      canvas.setColor(theme.dim)
      canvas.print(lines[index][0], x + 28, rowY)

      canvas.setColor(theme.text)
      canvas.print(lines[index][1], x + 180, rowY)
    }

    equipment = [
      ['Weapon', this.hero.equippedName('weapon')],
      ['Armour', this.hero.equippedName('armour')],
      ['Shield', this.hero.equippedName('shield')]
    ]

    for (index = 0; index < equipment.length(); index++) {
      rowY = y + 112 + index * 30

      canvas.setColor(theme.dim)
      canvas.print(equipment[index][0], x + 340, rowY)

      canvas.setColor(theme.text)
      canvas.print(equipment[index][1], x + 440, rowY)
    }

    if (this.hero.spells.length() > 0) {
      canvas.setColor(theme.dim)
      canvas.print('Spells', x + 340, y + 220)

      for (index = 0; index < this.hero.spells.length(); index++) {
        spell = findSpell(this.hero.spells[index])

        canvas.setColor(theme.text)
        canvas.print(spell.name + '  ' + spell.cost.toString() + ' MP', x + 440, y + 220 + index * 28)
      }
    }

    canvas.setColor(theme.dim)
    canvas.print('any key to go back', x + 28, y + 360)

    canvas.resetFont()
  }

  drawNotice() {
    canvas.setFont(this.fonts.body)

    width = this.fonts.body.getWidth(this.notice) + 40
    x = (window.width - width) / 2
    alpha = math.min(1, this.noticeTime / 0.5)

    canvas.setColor(color.rgb(18, 20, 34, alpha * 0.92))
    canvas.filledRectangle(x, 24, width, 46)

    canvas.setColor(theme.border)
    canvas.setLineWidth(2)
    canvas.rectangle(x, 24, width, 46)

    canvas.setColor(color.rgb(255, 255, 255, alpha))
    canvas.print(this.notice, x + 20, 34)

    canvas.resetFont()
  }
}

// changeLabel is the short "+5" shown against an equipment row.
// slotTitle capitalises a slot name for display, since the equipment map is
// keyed by lowercase names.
function slotTitle(slot) {
  switch (slot) {
    case 'weapon' { return 'Weapon' }
    case 'armour' { return 'Armour' }
    case 'shield' { return 'Shield' }
  }

  return slot
}

function changeLabel(change) {
  best = change.attack

  if (math.abs(change.defence) > math.abs(change.attack)) {
    best = change.defence
  }

  if (best == 0) {
    return '-'
  }

  if (best > 0) {
    return '+' + best.toString()
  }

  return best.toString()
}

function nextLevelText(hero) {
  needed = experienceForLevel(hero.level + 1) - hero.experience

  if (needed <= 0) {
    return 'ready'
  }

  return needed.toString() + ' more'
}

function rootHint(value) {
  switch (value) {
    case 'items' { return 'Look through what the party is carrying.' }
    case 'equip' { return 'Change what each member is wearing and holding.' }
    case 'status' { return 'Read a member\'s numbers in full.' }
    case 'rest' { return 'Restore health and magic. Free, for now.' }
    case 'save' { return 'Write the game to disk.' }
    case 'close' { return 'Back to the world.' }
  }

  return ''
}
