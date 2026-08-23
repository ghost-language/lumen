import "ghost:math"

// Every item, spell, monster, and party member in the game, in one place.
//
// Ghost's maps iterate in an unspecified order, so anything the player sees as
// a list is stored as a list and looked up by id. Maps are used only where the
// order genuinely does not matter.

// ---------------------------------------------------------------------------
// Items
//
// kind is one of 'weapon', 'armour', 'shield', or 'potion'. Equipment adds its
// attack and defence to whoever wears it; a potion is consumed for its effect.

items = [
  { id: 'club',        name: 'Oak club',      kind: 'weapon', attack: 4,  defence: 0, price: 20,  text: 'Heavy, and honest about it.' },
  { id: 'copper',      name: 'Copper sword',  kind: 'weapon', attack: 9,  defence: 0, price: 90,  text: 'Keeps an edge for about a week.' },
  { id: 'steel',       name: 'Steel sword',   kind: 'weapon', attack: 16, defence: 0, price: 340, text: 'The blade most guards carry.' },
  { id: 'staff',       name: 'Birch staff',   kind: 'weapon', attack: 5,  defence: 1, price: 60,  text: 'Light enough to cast around.' },
  { id: 'rags',        name: 'Travel rags',   kind: 'armour', attack: 0,  defence: 2, price: 15,  text: 'Better than nothing. Just.' },
  { id: 'leather',     name: 'Leather vest',  kind: 'armour', attack: 0,  defence: 7, price: 110, text: 'Boiled hard and worn soft.' },
  { id: 'chain',       name: 'Chain mail',    kind: 'armour', attack: 0,  defence: 13, price: 380, text: 'Rings like rain when you run.' },
  { id: 'robe',        name: 'Quiet robe',    kind: 'armour', attack: 0,  defence: 5, price: 130, text: 'Wards a little of everything.' },
  { id: 'buckler',     name: 'Wooden buckler', kind: 'shield', attack: 0, defence: 3, price: 45,  text: 'Small, fast, splinters often.' },
  { id: 'kite',        name: 'Kite shield',   kind: 'shield', attack: 0,  defence: 8, price: 210, text: 'Covers you from chin to knee.' },
  { id: 'herb',        name: 'Medic herb',    kind: 'potion', attack: 0,  defence: 0, price: 12,  heals: 30, text: 'Restores about 30 health.' },
  { id: 'tonic',       name: 'Clear tonic',   kind: 'potion', attack: 0,  defence: 0, restores: 18, price: 25, text: 'Restores about 18 magic.' },
  { id: 'elixir',      name: 'Amber elixir',  kind: 'potion', attack: 0,  defence: 0, heals: 200, restores: 60, price: 200, text: 'Mends almost anything.' },
  { id: 'brasskey',    name: 'Brass key',     kind: 'quest',  attack: 0,  defence: 0, price: 0,   text: 'Nobody knows what it opens.' },
  { id: 'wetmap',      name: 'Waterlogged map', kind: 'quest', attack: 0, defence: 0, price: 0,   text: 'Half the ink ran off.' }
]

// ---------------------------------------------------------------------------
// Spells
//
// target is 'enemy', 'ally', or 'party'.

spells = [
  { id: 'spark',  name: 'Spark',  cost: 3,  target: 'enemy', power: 12, kind: 'damage', text: 'A short, bright shock.' },
  { id: 'blaze',  name: 'Blaze',  cost: 8,  target: 'enemy', power: 28, kind: 'damage', text: 'Fire, and plenty of it.' },
  { id: 'mend',   name: 'Mend',   cost: 4,  target: 'ally',  power: 34, kind: 'heal',   text: 'Closes what is open.' },
  { id: 'rally',  name: 'Rally',  cost: 6,  target: 'party', power: 24, kind: 'heal',   text: 'Mends the whole party a little.' },
  { id: 'guard',  name: 'Guard',  cost: 5,  target: 'ally',  power: 6,  kind: 'buff',   text: 'Hardens a friend for the fight.' }
]

// ---------------------------------------------------------------------------
// Monsters
//
// `character` indexes characters.png, where each character occupies twelve
// frames and the first of them faces the player.

monsters = [
  { id: 'cutpurse', name: 'Cutpurse',  character: 8,  maxHealth: 22,  attack: 8,  defence: 3,  agility: 4,  experience: 6,  gold: 7,  spells: [] },
  { id: 'goblin', name: 'Goblin',      character: 10, maxHealth: 34,  attack: 13, defence: 6,  agility: 9,  experience: 12, gold: 14, spells: [] },
  { id: 'raider', name: 'Hill raider', character: 13, maxHealth: 52,  attack: 19, defence: 10, agility: 11, experience: 24, gold: 30, spells: [] },
  { id: 'shade',  name: 'Pale shade',  character: 12, maxHealth: 44,  attack: 16, defence: 8,  agility: 15, experience: 28, gold: 26, spells: ['spark'] },
  { id: 'sentry', name: 'Iron sentry', character: 11, maxHealth: 78,  attack: 24, defence: 20, agility: 6,  experience: 48, gold: 60, spells: [] },
  { id: 'ogre',   name: 'Fen ogre',    character: 6,  maxHealth: 120, attack: 32, defence: 14, agility: 7,  experience: 90, gold: 140, spells: ['blaze'] }
]

// Which monsters appear together, how likely each group is, and the party level
// it starts appearing at. Gating on level is what keeps the first few fights
// winnable: a level-one party that walks into an ogre has lost before it has
// learned which button attacks.
encounters = [
  { monsters: ['cutpurse'],                 weight: 6, level: 1 },
  { monsters: ['cutpurse', 'cutpurse'],     weight: 4, level: 1 },
  { monsters: ['goblin'],                   weight: 4, level: 2 },
  { monsters: ['goblin', 'cutpurse'],       weight: 3, level: 3 },
  { monsters: ['goblin', 'goblin'],         weight: 3, level: 4 },
  { monsters: ['shade'],                    weight: 3, level: 5 },
  { monsters: ['raider'],                   weight: 3, level: 6 },
  { monsters: ['raider', 'goblin'],         weight: 2, level: 8 },
  { monsters: ['shade', 'shade'],           weight: 2, level: 9 },
  { monsters: ['sentry'],                   weight: 2, level: 10 },
  { monsters: ['ogre'],                     weight: 1, level: 12 }
]

// ---------------------------------------------------------------------------
// The party

heroes = [
  {
    id: 'aldric', name: 'Aldric', character: 15, role: 'Knight',
    maxHealth: 46, maxMagic: 0,  strength: 12, defence: 8, agility: 9,
    spells: [], weapon: 'club', armour: 'rags', shield: null
  },
  {
    id: 'sera', name: 'Sera', character: 14, role: 'Mage',
    maxHealth: 28, maxMagic: 26, strength: 6, defence: 4, agility: 12,
    spells: ['spark', 'blaze'], weapon: 'staff', armour: null, shield: null
  },
  {
    id: 'nell', name: 'Nell', character: 16, role: 'Cleric',
    maxHealth: 34, maxMagic: 20, strength: 8, defence: 6, agility: 8,
    spells: ['mend', 'rally', 'guard'], weapon: null, armour: null, shield: null
  }
]

// ---------------------------------------------------------------------------
// Lookups

function findItem(id) {
  return findById(items, id)
}

function findSpell(id) {
  return findById(spells, id)
}

function findMonster(id) {
  return findById(monsters, id)
}

function findHero(id) {
  return findById(heroes, id)
}

// findById walks a list looking for a matching id. The tables are small enough
// that a scan is cheaper than the map it would take to avoid one.
function findById(table, id) {
  for (index = 0; index < table.length(); index++) {
    if (table[index].id == id) {
      return table[index]
    }
  }

  return null
}

// experienceForLevel is the running total needed to reach a level. The curve is
// steep enough that early levels come quickly and later ones take a few fights.
function experienceForLevel(level) {
  if (level <= 1) {
    return 0
  }

  return math.floor(12 * math.pow(level - 1, 1.9))
}
