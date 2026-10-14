package mm3

// Port of roland.drv (MM3.CC), the MT-32 / LAPC-1 driver behind an MPU-401. Its entry
// points are init, shutdown, music, timer/sfx and sample; the timer interrupt (tick) plays
// one step of the song, else of the sound effect, else the fade and the pitch slides. The
// sound effects' scripts and the note table are the driver's data, in rolanddata.go
// (go generate with MM3_DATA set to the game directory); the song is the caller's buffer.
//
// The driver's channels 0..8 are MIDI channels 2..10. A script byte's high nibble is the
// command, its low nibble the channel:
//
//	command   song                                sound effect
//	0         call the offset w (16 deep)         -
//	1         wait b ticks                        wait b ticks
//	2         program map[ch] = the 14th of 14 b  program map[ch] = b
//	3         volume b (not while fading)         volume b (not while fading)
//	4         stop the pitch slide                stop the pitch slide
//	5         pitch slide: step w, start w (two 7-bit bytes each), stepped every idle tick
//	6         pan b                               pan b
//	8         note off                            note off (b ignored)
//	9         note on b: channel 8 plays key b, the others the note table's key of b's
//	          low five bits plus an octave for each of its top three; velocity 127
//	10, 11    skip b (the AdLib's)                10 nothing; 11 system exclusive to 247
//	12        program map[b]                      program map[b]
//	15        255 returns, or loops to 0;         255 loops to the start; others end
//	          others end
//	7, 13, 14 do nothing.

//go:generate sh -c "go run ../../cmd/rolanddata \"$MM3_DATA\" > rolanddata.tmp && mv rolanddata.tmp rolanddata.go"

import "github.com/memmaker/mm3-go/internal/snd"

// MIDI.
const (
	midiNoteOff    = 128
	midiNoteOn     = 144
	midiControl    = 176
	midiProgram    = 192
	midiBend       = 224
	midiSysex      = 240
	midiSysexEnd   = 247
	ctlVolume      = 7
	ctlPan         = 10
	ctlResetAll    = 121
	ctlAllNotesOff = 123
	bendCentre     = 64 // each 7-bit half of the centred pitch bend
)

const (
	rolandChannels  = 9
	rolandSongChs   = 6 // the song's slides are on channels 0..5, the effects' on 6..8
	rolandVelocity  = 127
	rolandVolume    = 95 // a song's start sets every channel to it, and pans to the centre
	rolandPan       = 63
	rolandMaxCalls  = 16
	rolandFadeFloor = 40 // the fade lowers the volumes down to this
	rolandMapSize   = 16
)

// The script commands.
const (
	rcCall = iota
	rcWait
	rcMap
	rcVolume
	rcSlideOff
	rcSlide
	rcPan
	_
	rcNoteOff
	rcNoteOn
	rcAdlib
	rcSysex
	rcProgram
	_
	_
	rcEnd
	rcLoop = 255
)

type roland struct {
	out func(msg []byte)

	song      []byte
	musicOn   bool
	pos       int
	wait      byte
	calls     []int // the return positions
	musicMap  [rolandMapSize]byte
	fadeRate  byte
	fadeAcc   byte
	fadeLevel int8

	sfx     []byte
	sfxOn   bool
	sfxPos  int
	sfxWait byte
	sfxMap  [rolandMapSize]byte

	sounding [rolandChannels]int // the key on, -1 none
	volume   [rolandChannels]byte
	slide    [rolandChannels]uint16
	bend     [rolandChannels]uint16
}

func newRoland(out func([]byte)) *roland {
	r := &roland{out: out}
	for c := range r.sounding {
		r.sounding[c] = -1
	}
	return r
}

func (r *roland) msg(b ...byte) { r.out(b) }

// control sends controller ctl = v to the channels from hi down to lo, as the driver loops.
func (r *roland) control(lo, hi int, ctl, v byte) {
	for c := hi; c >= lo; c-- {
		r.msg(midiControl+1+byte(c), ctl, v)
	}
}

func (r *roland) bendCentre() {
	for c := rolandChannels - 1; c >= 0; c-- {
		r.msg(midiBend+1+byte(c), bendCentre, bendCentre)
	}
}

// songNotesOff silences channels 0..6, sfxNotesOff 7 and 8.
func (r *roland) songNotesOff() { r.control(0, rolandSongChs, ctlAllNotesOff, 0) }
func (r *roland) sfxNotesOff()  { r.control(rolandSongChs+1, rolandChannels-1, ctlAllNotesOff, 0) }

// noteOff ends channel ch's note, if one sounds.
func (r *roland) noteOff(ch int) {
	if k := r.sounding[ch]; k >= 0 {
		r.sounding[ch] = -1
		r.msg(midiNoteOff+1+byte(ch), byte(k), rolandVelocity)
	}
}

// key is note number n's key: the table's for its low five bits, plus its octave.
func key(n byte) byte {
	return rolandKeys[n%32] + n/32*12
}

func (r *roland) init() { r.musicOn, r.sfxOn = false, false }

func (r *roland) shutdown() { r.songNotesOff(); r.sfxNotesOff() }

// music is the music entry with the argument's high word hi: 0 stops the song, 1..255
// fades it out at that rate, -1 returns the fade rate (0 once it is over).
func (r *roland) music(hi int) int {
	switch {
	case hi == 0:
		r.musicOn = false
		r.songNotesOff()
	case hi > 0 && hi <= 255:
		r.fadeRate, r.fadeLevel = byte(hi), 127
	case hi == -1:
		return int(r.fadeRate)
	}
	return 0
}

// startSong plays song from its start.
func (r *roland) startSong(song []byte) {
	r.song = song
	r.calls = r.calls[:0]
	r.pos, r.wait = 0, 0
	for c := range rolandSongChs {
		r.slide[c] = 0
	}
	r.bendCentre()
	r.control(0, rolandChannels-1, ctlVolume, rolandVolume)
	r.control(0, rolandChannels-1, ctlPan, rolandPan)
	r.musicOn = true
	r.songNotesOff()
}

// startSfx starts sound effect v; past the list it does nothing, and entry 150, which
// points outside the driver, only silences the effects' channels.
func (r *roland) startSfx(v int) {
	if v < 0 || v >= len(rolandSfx) {
		return
	}
	for c := rolandSongChs; c < rolandChannels; c++ {
		r.slide[c] = 0
	}
	r.sfxNotesOff()
	r.control(0, rolandChannels-1, ctlResetAll, 0)
	r.bendCentre()
	r.sfx, r.sfxPos, r.sfxWait = rolandSfx[v], 0, 0
	r.sfxOn = r.sfx != nil
}

// sampleStart stops the song: a sample's interrupt replaces the music's until it ends.
func (r *roland) sampleStart() {
	r.musicOn = false
	r.songNotesOff()
	r.sfxNotesOff()
}

// tick is the timer interrupt: one step of the song, or else of the sound effect, or else
// the fade and the pitch slides.
func (r *roland) tick() {
	if r.musicOn {
		if r.wait == 0 {
			r.play(false)
			return
		}
		if r.wait--; r.wait == 0 {
			r.play(false)
			return
		}
	}
	if r.sfxOn {
		if r.sfxWait == 0 {
			r.play(true)
			return
		}
		if r.sfxWait--; r.sfxWait == 0 {
			r.play(true)
			return
		}
	}
	r.fade()
	for c := rolandChannels - 1; c >= 0; c-- {
		if r.slide[c] != 0 {
			r.bend[c] += r.slide[c]
			r.msg(midiBend+1+byte(c), byte(r.bend[c])%128, byte(r.bend[c]>>8)%128)
		}
	}
}

// play runs the song's or the effect's script up to its next wait or its end.
func (r *roland) play(sfx bool) {
	script, pos, maps := r.song, r.pos, &r.musicMap
	if sfx {
		script, pos, maps = r.sfx, r.sfxPos, &r.sfxMap
	}
	at := func(i int) byte { // past the song's end the driver reads zeros
		if i < len(script) {
			return script[i]
		}
		return 0
	}
	word := func(i int) uint16 { return uint16(at(i)) | uint16(at(i+1))<<8 }
	for {
		op := at(pos)
		pos++
		ch := int(op % 16)
		switch op / 16 {
		case rcCall:
			if !sfx && len(r.calls) < rolandMaxCalls {
				r.calls = append(r.calls, pos+2)
				pos = int(word(pos))
			}
		case rcWait:
			if sfx {
				r.sfxWait, r.sfxPos = at(pos), pos+1
			} else {
				r.wait, r.pos = at(pos), pos+1
			}
			return
		case rcMap:
			if sfx {
				maps[ch] = at(pos)
				pos++
			} else {
				maps[ch] = at(pos + 13)
				pos += 14
			}
		case rcVolume:
			if r.fadeRate == 0 {
				r.volume[ch] = at(pos)
				r.msg(midiControl+1+byte(ch), ctlVolume, r.volume[ch])
			}
			pos++
		case rcSlideOff:
			r.slide[ch] = 0
		case rcSlide:
			r.slide[ch], r.bend[ch] = word(pos), word(pos+2)
			pos += 4
		case rcPan:
			r.msg(midiControl+1+byte(ch), ctlPan, at(pos))
			pos++
		case rcNoteOff:
			if sfx {
				pos++
			}
			r.noteOff(ch)
		case rcNoteOn:
			k := at(pos)
			pos++
			if ch != rolandChannels-1 {
				k = key(k)
			}
			k %= 128
			r.noteOff(ch)
			r.sounding[ch] = int(k)
			// the desk's fader: music and sfx share the synth, so their notes are scaled
			kind := "music"
			if sfx {
				kind = "sfx"
			}
			r.msg(midiNoteOn+1+byte(ch), k, byte(rolandVelocity*snd.Volume(kind)+0.5))
		case rcAdlib:
			if !sfx {
				pos++
			}
		case rcSysex:
			if !sfx {
				pos++
				break
			}
			m := []byte{midiSysex}
			for {
				b := at(pos)
				pos++
				m = append(m, b)
				if b == midiSysexEnd {
					break
				}
			}
			r.out(m)
		case rcProgram:
			// three effects ask for programs past the map, where the driver keeps the
			// song's call targets; those are 0 unless the song nests calls
			var p byte
			if i := int(at(pos)); i < rolandMapSize {
				p = maps[i]
			}
			pos++
			r.msg(midiProgram+1+byte(ch), p)
		case rcEnd:
			if op != rcLoop {
				if sfx {
					r.sfxOn = false
				} else {
					r.musicOn = false
				}
				return
			}
			switch {
			case sfx:
				pos = 0
			case len(r.calls) > 0:
				pos = r.calls[len(r.calls)-1]
				r.calls = r.calls[:len(r.calls)-1]
			default:
				pos = 0
			}
		}
	}
}

// fade lowers the song's volumes: the rate is added to an 8-bit accumulator; on each
// carry the level drops and every channel above the floor loses a step. When the level
// runs out or every channel is at the floor, the song stops.
func (r *roland) fade() {
	if r.fadeRate == 0 {
		return
	}
	s := int(r.fadeAcc) + int(r.fadeRate)
	r.fadeAcc = byte(s)
	if s < 256 {
		return
	}
	stop := true
	if r.fadeLevel--; r.fadeLevel >= 0 {
		for c := rolandChannels - 1; c >= 0; c-- {
			if int8(r.volume[c]) >= rolandFadeFloor {
				r.volume[c]--
				stop = false
				r.msg(midiControl+1+byte(c), ctlVolume, r.volume[c])
			}
		}
	}
	if stop {
		r.musicOn, r.fadeRate = false, 0
		r.songNotesOff()
	}
}
