package mm3

// Sound driver interface: snddrv.asm (load_sound_driver and the five thunks) and the
// common behaviour of the drivers behind them (roland, blaster, adlib, covox, tandy, ibm,
// demo .drv). Each driver starts with a jump table, entered by the thunks with the C
// caller's stack: +0 init, +3 shutdown, +6 music, +9 timer/sfx, +0c sample.
//
// The timing is adlib.drv's (offsets below); the other drivers keep the same timer and
// argument conventions. With munt and the MT-32 ROMs (RomDir), music and sfx are
// roland.drv's (roland.go) and samples are mixed into the same output; without them
// everything is silent and only the state and timing are kept, so the game's waits
// ("sample finished", "fade done") behave.

import (
	"fmt"
	"os"

	"github.com/memmaker/mm3-go/internal/host"
	"github.com/memmaker/mm3-go/internal/snd"
)

// RomDir is where the MT-32 ROMs are ("" or missing: silent).
var RomDir string

// rol is the roland.drv player, nil when silent.
var rol *roland

const (
	pitHz        = 1193182
	sndTickDiv   = 0x4006 // adlib.drv 022e: PIT divisor of the music/timer interrupt, ~72.8 Hz
	sndSampleDiv = 0x95   // adlib.drv 0181: PIT divisor while a sample plays, ~8008 Hz
	sndSampleTk  = 0x6d   // adlib.drv 0347: one timer tick per 109 samples
)

var (
	sndLoaded     bool   // snddrv 008d: driver segment != 0
	sndTicks      uint16 // adlib.drv 000f
	sndMusicOn    bool   // adlib.drv 0013
	sndFadeRate   byte   // adlib.drv 0172
	sndFadeAcc    byte   // adlib.drv 0173
	sndFadeLevel  int8   // adlib.drv 0174
	sndSampleOn   bool   // adlib.drv 017f
	sndSampleLen  int    // adlib.drv 0179
	sndSamplePos  int    // adlib.drv 017b
	sndSampleTkN  int    // adlib.drv 0180
	sndPitAcc     int    // PIT clocks owed to the interrupt, in 1/70 units
	sndTimerAdded bool
)

// loadSoundDriver is snddrv 000e: n is MM3.CFG's sound byte, indexing roland, blaster,
// adlib, covox, tandy, ibm, demo. The driver code is not needed, only its presence.
func loadSoundDriver(n int) {
	if n < 0 || n > 6 {
		fatal("bad sound device")
	}
	sndLoaded = true
}

// sndInit is adlib.drv 0205: hook the timer interrupt at 72.8 Hz, reset the chip.
// port is the card's base port from MM3.CFG (unused here).
func sndInit(port int) {
	if !sndLoaded {
		return
	}
	sndMusicOn = false
	if rol == nil && RomDir != "" {
		if err := snd.Open(RomDir); err != nil {
			fmt.Fprintln(os.Stderr, "MT-32:", err, "- no music (speech still plays)")
		} else {
			rol = newRoland(ccLoad("roland.drv"), snd.Send)
		}
	}
	snd.StartOutput() // speech plays without the MT-32 too
	if rol != nil {
		rol.init()
	}
	if !sndTimerAdded {
		sndTimerAdded = true
		host.AddTimer(sndHostTick)
	}
}

// sndShutdown is adlib.drv 025f: unhook the timer, silence all channels.
func sndShutdown() {
	sndMusicOn = false
	if rol != nil {
		rol.shutdown()
		snd.PlaySample(nil, 0)
	}
}

// sndHostTick runs the driver's interrupts that fall within one 70 Hz host tick.
func sndHostTick() {
	sndPitAcc += pitHz
	for {
		if sndSampleOn {
			if sndPitAcc < sndSampleDiv*host.TPS {
				return
			}
			sndPitAcc -= sndSampleDiv * host.TPS
			sndSampleIRQ()
		} else {
			if sndPitAcc < sndTickDiv*host.TPS {
				return
			}
			sndPitAcc -= sndTickDiv * host.TPS
			sndTickIRQ()
		}
	}
}

// sndSampleIRQ is adlib.drv 0314: one sample out; every 109th one runs the timer tick.
func sndSampleIRQ() {
	sndSamplePos++
	if sndSamplePos >= sndSampleLen {
		sndSampleOn = false // 0379
	}
	if sndSampleTkN++; sndSampleTkN >= sndSampleTk {
		sndSampleTkN = 0
		sndTickIRQ()
	}
}

// sndTickIRQ is adlib.drv 060f: count the tick; the music (skipped while a sample
// plays, 086f) fades: rate is added to an 8-bit accumulator, each carry lowers the
// level from 3f, and when it passes 0 the music stops (0889).
func sndTickIRQ() {
	sndTicks++
	if rol != nil {
		if !sndSampleOn {
			rol.tick() // roland.drv 0577; a sample's interrupt (039e) only counts
		}
		snd.Tick()
		return
	}
	if sndSampleOn || sndFadeRate == 0 {
		return
	}
	s := uint16(sndFadeAcc) + uint16(sndFadeRate)
	sndFadeAcc = byte(s)
	if s > 0xff {
		if sndFadeLevel--; sndFadeLevel < 0 {
			sndMusicOn = false
			sndFadeRate = 0
		}
	}
}

// sndTimer is adlib.drv 048c, the timer and sound-effect entry:
//
//	-1: return the tick count (16 bit, ~72.8 Hz, wraps)
//	other negative: reset the count to 0
//	>= 0: start sound effect v (0x21 and 0x23 are ignored by adlib)
//
// The return value is meaningful only for -1. A busy wait on it lets time pass.
func sndTimer(v int) int {
	if !sndLoaded {
		return 0
	}
	switch {
	case v == -1:
		host.Busy()
		return int(int16(sndTicks))
	case v < 0:
		sndTicks = 0
	case rol != nil:
		rol.sfx(v) // roland.drv 02c2: 0..96, no exceptions
	}
	return 0
}

// sndPlayMusic is adlib.drv 0433, decoding the long argument by its high word:
//
//	0: stop the music
//	1..0xff: fade out at speed v>>1 (play_music's flag)
//	0xffff: return the fade speed, 0 once the fade has ended
//	else: start the song at seg:0
//
// The Go port's segment handles are small numbers, so a buffer must be started with
// sndStartMusic; pass only the C's constant longs (0, flag<<16, -1) here.
func sndPlayMusic(v int32) int {
	if !sndLoaded {
		return 0
	}
	if rol != nil {
		if hi := int(int16(uint32(v) >> 16)); hi >= -1 && hi <= 0xff {
			return rol.music(hi)
		}
		fatal("sndPlayMusic: use sndStartMusic for a buffer")
	}
	switch hi := uint16(uint32(v) >> 16); {
	case hi == 0:
		sndMusicOn = false
	case hi <= 0xff:
		sndFadeRate = byte(hi >> 1)
		sndFadeLevel = 0x3f
	case hi == 0xffff:
		return int(sndFadeRate)
	default:
		fatal("sndPlayMusic: use sndStartMusic for a buffer")
	}
	return 0
}

// sndStartMusic is the song case of sndPlayMusic (adlib.drv 046e): play the .m data at p.
func sndStartMusic(p Far) {
	if !sndLoaded {
		return
	}
	sndMusicOn = true
	if rol != nil {
		rol.startSong(p.B())
	}
}

// sndPlaySample is adlib.drv 0293. p's offset != 0 (the C passes 1L): return 1 while a
// sample plays. p == 0: stop it. Else play the n unsigned 8-bit samples at p (~8 kHz).
func sndPlaySample(p Far, n int) int {
	if !sndLoaded {
		return 0
	}
	switch {
	case p.Off() != 0:
		if sndSampleOn {
			return 1
		}
		return 0
	case p == 0:
		sndSampleOn = false
		if rol != nil {
			snd.PlaySample(nil, 0)
		}
	default:
		n = int(uint16(n)) // the driver's count is an unsigned 16-bit register
		b := p.B()[:n]
		sndSampleLen, sndSamplePos, sndSampleOn = n, 0, n > 0
		if rol != nil {
			rol.sampleStart()
			snd.PlaySample(b, 8008)
		}
	}
	return 0
}
