package mm3

// Port of adlib.drv (MM3.CC), the AdLib driver, onto an OPL2 (github.com/memmaker/opl2,
// behind internal/snd). It plays the same songs as roland.drv, reading the commands meant
// for the AdLib and skipping the MT-32's, and its own sound effects, whose scripts and
// note table are the driver's data, in adlibdata.go (go generate with MM3_DATA set to the
// game directory).
//
// Channels 0..6 are the song's, 7 and 8 the effects'. A script byte's high nibble is the
// command, its low nibble the channel:
//
//	command  song                               sound effect
//	0        call the offset w (16 deep)        call the offset w (16 deep)
//	1        wait b ticks                       wait b ticks
//	2        define instrument ch: 14 b         define instrument ch: 11 b
//	3        skip b (the MT-32's volume)        volume b (not while fading)
//	6        skip b (the MT-32's pan)           set note b, keyed off
//	7        key off                            key off
//	8        key off, keeping the key bit       set note b, keyed off
//	9        note on b                          note on b
//	10       volume b (not while fading)        -
//	11       skip b                             -
//	12       instrument b                       instrument b
//	13       -                                  stop the pitch slide
//	14       -                                  pitch slide: speed b, step w (high first)
//	15       255 returns, or loops to 0;        255 returns, or loops to the start;
//	         others end                         others end
//
// A note number b is the note table's frequency for its low five bits, raised an octave
// (a block) for each of its top three. A volume is the carrier's level register, and an
// instrument is its two operators' registers (flags, level, attack and decay, sustain
// and release, waveform: the modulator's then the carrier's), then the feedback.

//go:generate sh -c "go run ../../cmd/adlibdata \"$MM3_DATA\" > adlibdata.tmp && mv adlibdata.tmp adlibdata.go"

import "github.com/memmaker/opl2"

const (
	adlibChannels   = 9
	adlibSongChs    = 7 // channels 0..6
	adlibSilent     = 63
	adlibKeyOn      = 1 << 13 // in a channel's frequency word: the high register's bit 5
	adlibFnumMask   = 1<<10 - 1
	adlibBlockStep  = 1 << 10
	adlibBlockMask  = 7 << 10
	adlibMaxCalls   = 16
	adlibPrograms   = 16
	adlibSongInstr  = 14 // an instrument definition's bytes in a song (the MT-32's after it)
	adlibSfxInstr   = 11
	adlibSlideTop   = 734 // a rising slide past this frequency number goes up a block
	adlibSlideFloor = 388 // a falling one below this goes down
)

// The script commands (in a song or an effect, as the table above has them).
const (
	acCall = iota
	acWait
	acDefine
	acSfxVolume
	_
	_
	acSfxNote
	acKeyOff
	acRelease // a song's; an effect's sets a note, as acSfxNote
	acNoteOn
	acVolume
	acSkip
	acProgram
	acSlideOff
	acSlide
	acEnd
	acLoop = 255
)

// The effects the driver does not play.
var adlibIgnored = map[int]bool{33: true, 35: true}

type adlib struct {
	write func(reg int, val byte)

	song      []byte
	musicOn   bool
	pos       int
	wait      byte
	calls     []int
	songInstr [adlibPrograms]int // where each instrument is defined in the song
	fadeRate  byte
	fadeAcc   byte
	fadeLevel int8

	sfxOn    bool
	sfxStart int
	sfxPos   int
	sfxWait  byte
	sfxCalls []int
	sfxInstr [adlibPrograms]int

	freq       [adlibChannels]uint16
	volume     [adlibChannels]byte
	sliding    [adlibChannels]bool
	slideSpeed [adlibChannels]byte
	slideAcc   [adlibChannels]byte
	slideStep  [adlibChannels]int16

	sampleOn bool
}

func newAdlib(write func(reg int, val byte)) *adlib { return &adlib{write: write} }

func modulator(ch int) int { return ch%3 + ch/3*8 }
func carrier(ch int) int   { return modulator(ch) + 3 }

func (a *adlib) setFreq(ch int, f uint16) {
	a.write(opl2.RegFreqLow+ch, byte(f))
	a.write(opl2.RegFreqHigh+ch, byte(f>>8))
}

func (a *adlib) setVolume(ch int) { a.write(opl2.RegOpLevel+carrier(ch), a.volume[ch]) }

// silence keys channels hi down to lo off and turns their volume down.
func (a *adlib) silence(lo, hi int) {
	for ch := hi; ch >= lo; ch-- {
		a.freq[ch] = 0
		a.setFreq(ch, 0)
		a.volume[ch] = adlibSilent
		a.setVolume(ch)
	}
}

func (a *adlib) songNotesOff() { a.silence(0, adlibSongChs-1) }
func (a *adlib) sfxNotesOff()  { a.silence(adlibSongChs, adlibChannels-1) }

// note is note number n's frequency word.
func note(n byte) uint16 { return adlibKeys[n%32] + uint16(n/32)*adlibBlockStep }

// program sets channel ch's instrument from the definition in script at pos.
func (a *adlib) program(ch int, script []byte, pos int) {
	at := func() byte {
		var b byte
		if pos < len(script) {
			b = script[pos]
		}
		pos++
		return b
	}
	for _, op := range []int{modulator(ch), carrier(ch)} {
		for _, reg := range []int{opl2.RegOpFlags, opl2.RegOpLevel, opl2.RegOpAttackDecay, opl2.RegOpSustainRelease, opl2.RegWaveform} {
			v := at()
			if op == carrier(ch) && reg == opl2.RegOpLevel && a.fadeRate != 0 {
				v = v&(3<<6) | byte(a.fadeLevel) // a fading song's new instruments start at its level
			}
			a.write(reg+op, v)
		}
	}
	a.write(opl2.RegFeedback+ch, at())
}

func (a *adlib) init() {
	a.write(opl2.RegTest, 1<<5)
	a.write(opl2.RegKeyboardSplit, 0)
	a.write(opl2.RegRhythm, 0)
	a.musicOn, a.sfxOn = false, false
}

func (a *adlib) shutdown() { a.songNotesOff(); a.sfxNotesOff() }

// music is the music entry with the argument's high word hi: 0 stops the song, 1..255
// fades it out at half that rate, -1 returns the fade rate (0 once it is over).
func (a *adlib) music(hi int) int {
	switch {
	case hi == 0:
		a.musicOn = false
		a.songNotesOff()
	case hi > 0 && hi <= 255:
		a.fadeRate, a.fadeLevel = byte(hi/2), adlibSilent
	case hi == -1:
		return int(a.fadeRate)
	}
	return 0
}

func (a *adlib) startSong(song []byte) {
	a.song = song
	a.calls = a.calls[:0]
	a.pos, a.wait = 0, 0
	a.musicOn = true
	a.songNotesOff()
}

// startSfx starts sound effect v; the driver ignores two, and past its list it does
// nothing here.
func (a *adlib) startSfx(v int) {
	if adlibIgnored[v] || v < 0 || v >= len(adlibSfx) {
		return
	}
	a.sfxStart = adlibSfx[v]
	a.sfxPos, a.sfxWait = a.sfxStart, 0
	a.sfxNotesOff()
	a.sfxOn = true
}

// sampleStart sets channel 7 up for the driver's sample playback, which steps its
// carrier's level with the samples; the mixer plays the samples instead, so the level
// stays down.
func (a *adlib) sampleStart() {
	op := carrier(7)
	a.write(opl2.RegOpFlags+op, 1<<5|1)
	a.write(opl2.RegOpAttackDecay+op, 15<<4)
	a.write(opl2.RegOpSustainRelease+op, 15<<4)
	a.write(opl2.RegFeedback+7, 1)
	a.write(opl2.RegWaveform+op, 0)
	a.write(opl2.RegOpLevel+op, adlibSilent)
	a.setFreq(7, 0)
	a.setFreq(7, adlibKeyOn|3*adlibBlockStep|655)
	a.setFreq(7, adlibKeyOn)
	a.sampleOn = true
}

// sampleStop keys channel 7 off again.
func (a *adlib) sampleStop() {
	a.sampleOn = false
	a.write(opl2.RegFreqHigh+7, 0)
	a.write(opl2.RegFreqLow+7, 0)
}

// tick is the timer interrupt: one step of the song, or else of the sound effect; after a
// step that waits, or when neither steps, the fade and the effects' pitch slides.
func (a *adlib) tick() {
	if a.musicOn {
		if a.wait == 0 {
			a.play(false)
			return
		}
		if a.wait--; a.wait == 0 {
			a.play(false)
			return
		}
	}
	if a.sfxOn {
		if a.sfxWait == 0 {
			a.play(true)
			return
		}
		if a.sfxWait--; a.sfxWait == 0 {
			a.play(true)
			return
		}
	}
	a.idle()
}

// play runs the song's or the effect's script up to its next wait or its end.
func (a *adlib) play(sfx bool) {
	script, pos, calls, instr := a.song, a.pos, &a.calls, &a.songInstr
	if sfx {
		script, pos, calls, instr = adlibScripts, a.sfxPos, &a.sfxCalls, &a.sfxInstr
	}
	at := func(i int) byte { // past the song's end the driver reads zeros
		if i < len(script) {
			return script[i]
		}
		return 0
	}
	for {
		op := at(pos)
		pos++
		ch := int(op % 16)
		cmd := int(op / 16)
		if sfx {
			switch cmd {
			case acRelease:
				cmd = acSfxNote
			case acVolume, acSkip:
				cmd = -1
			case acSfxVolume:
				cmd = acVolume
			}
		} else {
			switch cmd {
			case acSfxVolume, acSfxNote, acSkip:
				pos++
				cmd = -1
			case acSlideOff, acSlide:
				cmd = -1
			}
		}
		switch cmd {
		case acCall:
			if len(*calls) < adlibMaxCalls {
				*calls = append(*calls, pos+2)
				pos = int(at(pos)) | int(at(pos+1))<<8
			}
		case acWait:
			if sfx {
				a.sfxWait, a.sfxPos = at(pos), pos+1
			} else {
				a.wait, a.pos = at(pos), pos+1
			}
			a.idle()
			return
		case acDefine:
			instr[ch] = pos
			if sfx {
				pos += adlibSfxInstr
			} else {
				pos += adlibSongInstr
			}
		case acVolume:
			if a.fadeRate == 0 {
				a.volume[ch] = at(pos)
				a.setVolume(ch)
			}
			pos++
		case acSfxNote:
			a.freq[ch] = note(at(pos))
			pos++
			a.setFreq(ch, a.freq[ch])
		case acKeyOff:
			a.freq[ch] &^= adlibKeyOn
			a.write(opl2.RegFreqHigh+ch, byte(a.freq[ch]>>8))
		case acRelease:
			a.setFreq(ch, a.freq[ch]&(adlibKeyOn-1))
		case acNoteOn:
			f := note(at(pos))
			pos++
			a.setFreq(ch, f)
			a.freq[ch] = f | adlibKeyOn
			a.setFreq(ch, a.freq[ch])
		case acProgram:
			// a song's instrument past the 16 would be read from the effects' table
			if p := int(at(pos)); p < adlibPrograms {
				a.program(ch, script, instr[p])
			}
			pos++
		case acSlideOff:
			a.sliding[ch] = false
		case acSlide:
			a.slideSpeed[ch] = at(pos)
			a.slideStep[ch] = int16(uint16(at(pos+1))<<8 | uint16(at(pos+2)))
			a.slideAcc[ch], a.sliding[ch] = 255, true
			pos += 3
		case acEnd:
			switch {
			case op != acLoop:
				if sfx {
					a.sfxOn = false
				} else {
					a.musicOn = false
				}
				return
			case len(*calls) > 0:
				pos = (*calls)[len(*calls)-1]
				*calls = (*calls)[:len(*calls)-1]
			case sfx:
				pos = a.sfxStart
			default:
				pos = 0
			}
		}
	}
}

// idle runs the fade and the effects' pitch slides, unless a sample plays.
func (a *adlib) idle() {
	if a.sampleOn {
		return
	}
	a.fade()
	for ch := adlibChannels - 1; ch >= adlibSongChs; ch-- {
		a.slide(ch)
	}
}

// fade turns the volumes down: the rate is added to an 8-bit accumulator, and each carry
// lowers the level and every channel's volume a step; when the level runs out the song
// stops.
func (a *adlib) fade() {
	if a.fadeRate == 0 {
		return
	}
	s := int(a.fadeAcc) + int(a.fadeRate)
	a.fadeAcc = byte(s)
	if s < 256 {
		return
	}
	if a.fadeLevel--; a.fadeLevel < 0 {
		a.musicOn, a.fadeRate = false, 0
		a.songNotesOff()
		return
	}
	for ch := adlibChannels - 1; ch >= 0; ch-- {
		if a.volume[ch] != adlibSilent {
			a.volume[ch]++
			a.setVolume(ch)
		}
	}
}

// slide moves channel ch's frequency: the speed is added to an 8-bit accumulator, and
// each carry adds the step to the frequency number, which changes block, halved or
// doubled, when it leaves the range between the floor and the top.
func (a *adlib) slide(ch int) {
	if !a.sliding[ch] {
		return
	}
	s := int(a.slideAcc[ch]) + int(a.slideSpeed[ch])
	a.slideAcc[ch] = byte(s)
	if s < 256 {
		return
	}
	f := a.freq[ch]
	block, key := f&adlibBlockMask, f&adlibKeyOn
	n := int16(f&adlibFnumMask) + a.slideStep[ch]
	switch {
	case a.slideStep[ch] >= 0 && n >= adlibSlideTop:
		n = int16(uint16(n) >> 1)
		if n&adlibFnumMask == 0 {
			n++
		}
		block = (block + adlibBlockStep) & adlibBlockMask
	case a.slideStep[ch] < 0 && n <= adlibSlideFloor:
		n <<= 1
		if n&adlibFnumMask == 0 {
			n--
		}
		block = (block - adlibBlockStep) & adlibBlockMask
	}
	a.freq[ch] = uint16(n)&adlibFnumMask | block | key
	a.setFreq(ch, a.freq[ch])
}
