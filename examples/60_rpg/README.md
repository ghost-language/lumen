# Top-down RPG

A complete small RPG, written in Ghost with nothing but Lumen's modules: a world
to walk, people to talk to, random encounters, turn-based battles, levelling,
equipment, an inventory, and saving.

```bash
lumen examples/60_rpg
```

| | |
| --- | --- |
| move | arrow keys / WASD / gamepad stick or d-pad |
| confirm | space or E (gamepad A) |
| cancel | escape or Q (gamepad B) |
| menu | tab (gamepad start) |
| save / load | F5 / F9 |
| debug | F1 |
| mute | M |
| fullscreen | F11 |
| quit | escape, with nothing else open |

## What each file shows

| File | |
| --- | --- |
| `main.ghost` | game state, callbacks, encounters, interaction, saving, depth sorting |
| `data.ghost` | every item, spell, monster, and hero in one table |
| `combatant.ghost` | shared stats for heroes and monsters, damage, levelling |
| `party.ghost` | the party, the purse, the pack, and equipping |
| `battle.ghost` | the turn-based battle state machine and combat arithmetic |
| `battleview.ghost` | drawing a battle, with no rules in it |
| `ui.ghost` | the panel, bar, and scrolling menu every screen is built from |
| `fieldmenu.ghost` | pack, equipment, status sheets, resting, saving |
| `tilemap.ghost` | loading a Tiled JSON map, per-layer collision, culling |
| `camera.ghost` | smoothed following, map bounds, zoom, screen shake |
| `player.ghost` | dt-scaled movement, axis-separated collision, walk cycles |
| `npc.ghost` | characters that talk and hand over items |
| `dialogue.ghost` | typewriter text, wrapping, paging |
| `hud.ghost` | party health and gold while walking |
| `spritesheet.ghost` | slicing a sheet into quads and naming animations |
| `sounds.ghost` | a small sound bank |

## The battle

Front-view and turn-based, in the Dragon Quest mould. Each round every hero is
given an order, the monsters decide theirs, and then everyone acts in order of
agility — so a fast monster can strike before a hero who chose first.

Orders are **Fight**, **Spell**, **Item**, **Guard**, and **Run**. Guarding
doubles defence for the round. Running is likelier the faster the party is
relative to the monsters, and a failed attempt costs the turn.

Damage is `attack - defence/2`, multiplied by a roll between 0.85 and 1.15, with
a one-in-twenty-five chance of a critical hit that ignores armour. Spells ignore
most armour, which is what makes a mage worth bringing against something in
chain mail. Winning splits experience across the survivors and levels them up
along their own aptitudes: the knight keeps gaining health, the mage magic.

Losing costs half the purse and sends the party back where they started. It is
never a lost save.

Encounters are rolled from a weighted table gated on party level, so a level-one
party meets cutpurses and never an ogre.

## The menus

Tab opens the pack, equipment, status sheets, rest, and save. Equipment has
three slots per hero, and the item list shows what each piece would do to that
hero's attack and defence **before** it is equipped, which is the question the
player is actually asking. Taking something off puts it back in the pack.

Everything is built from one `Menu` in `ui.ghost`: rows with a cursor, disabled
entries that grey out and are skipped, a detail panel alongside, and scrolling
once a list outgrows its panel.

## Things worth copying

**Everything moves in pixels per second.** `update(dt)` gets the length of the
previous frame, and speeds are multiplied by it, so the game plays the same on
any machine.

**The tilemap only draws what the camera can see.** The map is 50x50 tiles
across six layers — 15,000 tiles. `camera.visibleTiles()` narrows that to the few
hundred actually on screen.

**Collision is resolved one axis at a time.** Moving x and y separately is what
lets a player pressing diagonally into a wall slide along it rather than stick.

**Characters are depth-sorted by y.** Someone standing lower on the screen is
drawn last, so they overlap whoever is behind them.

**The camera is pushed and popped.** The world is drawn inside
`camera.attach()` / `camera.detach()`; the HUD, dialogue, and menus are drawn
outside, in screen coordinates, unaffected by zoom.

**The battle rules never touch the canvas.** `battle.ghost` decides what
happens; `battleview.ghost` draws it. Combat can be reasoned about, and changed,
without a window open.

**Everything the player reads goes through a message queue.** A battle sits on
each line until it is dismissed or times out, so a whole round never resolves
inside one frame with nothing to show for it.

**Saves go to the save directory.** `filesystem` writes to the player's data
directory, not next to the game, and `filesystem.read()` returns `null` when
there is no save yet. Only what cannot be derived is written: levels, current
health, the pack, and what each hero is wearing.

## Packaging it

```bash
lumen package examples/60_rpg -o rpg.lumen   # one file, run with `lumen rpg.lumen`
lumen fuse examples/60_rpg -o rpg            # a standalone executable
```

## Assets

`tilesheet.png`, `characters.png`, and `map.json` come from the `53_top_down`
example. The `.wav` effects are synthesised square waves, filtered noise, and
sine arpeggios — placeholders, and small enough to keep in the repository.
