# Top-down RPG

A complete small RPG, written in Ghost with nothing but Lumen's modules.

```bash
lumen examples/60_rpg
```

| | |
| --- | --- |
| move | arrow keys / WASD / gamepad stick or d-pad |
| interact | space or E (gamepad A) |
| pack | tab (mouse wheel scrolls) |
| save / load | F5 / F9 |
| mute | M |
| debug overlay | F1 |
| fullscreen | F11 |
| quit | escape |

## What each file shows

| File | |
| --- | --- |
| `main.ghost` | game state, callbacks, interaction, saving, depth sorting, screen fade |
| `tilemap.ghost` | loading a Tiled JSON map, per-layer collision, culling to the camera |
| `camera.ghost` | smoothed following, map bounds, zoom, screen shake |
| `player.ghost` | dt-scaled movement, axis-separated collision, directional animation |
| `npc.ghost` | characters that talk and hand over items |
| `dialogue.ghost` | typewriter text, wrapping, paging |
| `hud.ghost` | hearts, coins, and a scissor-clipped inventory panel |
| `spritesheet.ghost` | slicing a sheet into quads and naming animations |
| `sounds.ghost` | a small sound bank |

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
`camera.attach()` / `camera.detach()`; the HUD and dialogue are drawn outside,
in screen coordinates, unaffected by zoom.

**Saves go to the save directory.** `filesystem` writes to the player's config
directory, not next to the game, and `filesystem.read()` returns `null` when
there is no save yet.

## Assets

`tilesheet.png`, `characters.png`, and `map.json` come from the `53_top_down`
example. The three `.wav` effects are synthesised square waves and filtered
noise — placeholders, and small enough to keep in the repository.
