import audio, { Source } from "lumen:audio"

// A tiny sound bank. Loading every effect once and playing by name keeps audio
// calls out of the gameplay code, and means a missing file is reported at
// startup rather than the first time something happens to trigger it.
class Sounds {
  constructor(names) {
    this.effects = {}
    this.enabled = true

    for (index = 0; index < names.length(); index++) {
      name = names[index]
      this.effects[name] = new Source('resources/' + name + '.wav')
    }
  }

  play(name) {
    if (!this.enabled) {
      return null
    }

    source = this.effects[name]

    if (source != null) {
      source.play()
    }
  }

  setVolume(volume) {
    audio.setVolume(volume)
  }

  toggle() {
    this.enabled = !this.enabled

    if (this.enabled) {
      audio.setVolume(0.7)
    } else {
      audio.setVolume(0)
    }

    return this.enabled
  }
}
