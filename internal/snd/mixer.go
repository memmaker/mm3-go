// Package snd is the audio output: munt's MT-32 emulator (libmt32emu, loaded at run time
// with purego, mt32_dl.go; ROMs supplied by the player), or else an OPL2
// (github.com/memmaker/opl2), plus 8-bit PCM speech, mixed into one Ebiten audio stream.
// Adapted from UltimaUnderworld's internal/snd.
//
// The driver (mm3's roland.drv or adlib.drv port) runs on the game's timer interrupt; it
// queues its MIDI messages with Send, or its register writes with WriteOPL, and closes
// each ~72.8 Hz tick with Tick. The audio thread
// replays the queued ticks at the same rate on its own clock, so the music keeps its
// rhythm however the game's ticks are batched.
package snd

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/memmaker/opl2"
)

const (
	outRate = 44100
	tickPIT = 0x4006 // roland.drv's timer divisor
	pitHz   = 1193182
	backlog = 8 // queued ticks beyond this are played at once (the game ran ahead)
)

var (
	mu      sync.Mutex
	opened  bool
	ticks   [][]event       // closed ticks waiting to be played
	cur     []event         // the tick being filled
	chip    *opl2.Resampler // the OPL2, nil with the MT-32
	musicV  float64         // the desk's faders, read once per Read
	sfxV    float64
	tickAcc int     // frames*pitHz owed to the next tick
	speech  []byte  // unsigned 8-bit mono
	speechF float64 // position in speech
	speechD float64 // speech samples per output frame
	buf16   []int16
	started bool
)

// Open starts munt on the ROMs in romDir. It fails, leaving nothing set up, without the
// library or the ROMs.
func Open(romDir string) error {
	if err := mtOpen(romDir); err != nil {
		return err
	}
	opened = true
	return nil
}

// event is a MIDI message, or with none an OPL2 register write.
type event struct {
	msg []byte
	reg int
	val byte
}

// OpenOPL starts the OPL2, in place of the MT-32. The driver's channels 0..6 play the
// music, 7 and 8 the effects.
func OpenOPL() {
	chip = opl2.NewResampler(opl2.New(), outRate)
	chip.Mix = func(ch *[opl2.Channels]int) int {
		m, s := 0, 0
		for i, v := range ch {
			if i < 7 {
				m += v
			} else {
				s += v
			}
		}
		return int(float64(m)*musicV + float64(s)*sfxV)
	}
	opened = true
}

// StartOutput starts the Ebiten audio stream, once; without the MT-32 it carries speech only.
func StartOutput() {
	if started {
		return
	}
	started = true
	p, err := audio.NewContext(outRate).NewPlayer(Stream{})
	if err != nil {
		return
	}
	p.SetBufferSize(60 * time.Millisecond)
	p.Play()
}

// Send queues a MIDI message (a short one or a sysex) in the current tick.
func Send(msg []byte) {
	mu.Lock()
	cur = append(cur, event{msg: append([]byte(nil), msg...)})
	mu.Unlock()
}

// WriteOPL queues an OPL2 register write in the current tick.
func WriteOPL(reg int, val byte) {
	mu.Lock()
	cur = append(cur, event{reg: reg, val: val})
	mu.Unlock()
}

// Tick closes the current tick.
func Tick() {
	mu.Lock()
	ticks = append(ticks, cur)
	cur = nil
	mu.Unlock()
}

// PlaySample plays unsigned 8-bit mono PCM at hz, replacing any playing; nil stops it.
func PlaySample(pcm []byte, hz int) {
	mu.Lock()
	speech, speechF, speechD = append([]byte(nil), pcm...), 0, float64(hz)/outRate
	mu.Unlock()
}

func play(t []event) {
	for _, e := range t {
		m := e.msg
		if m == nil {
			if chip != nil {
				chip.Chip.Write(e.reg, e.val)
			}
			continue
		}
		if m[0] == 0xF0 {
			mtSysex(m)
			continue
		}
		v := uint32(0)
		for i, b := range m {
			v |= uint32(b) << (8 * i)
		}
		mtMsg(v)
	}
}

// Stream is the mixed output: 16-bit little-endian stereo at 44.1 kHz.
type Stream struct{}

func (Stream) Read(b []byte) (int, error) {
	frames := len(b) / 4
	mu.Lock()
	defer mu.Unlock()
	if !opened {
		ticks = nil
	}
	for len(ticks) > backlog {
		play(ticks[0])
		ticks = ticks[1:]
	}
	if cap(buf16) < frames*2 {
		buf16 = make([]int16, frames*2)
	}
	out := buf16[:frames*2]
	musicV, sfxV = Volume("music"), Volume("sfx")
	for f := 0; f < frames; {
		n := (tickPIT*outRate - tickAcc + pitHz - 1) / pitHz // frames to the next tick
		if n > frames-f {
			n = frames - f
		}
		if n > 0 {
			if chip != nil {
				for i := f; i < f+n; i++ {
					v := clamp16(chip.Next())
					out[i*2], out[i*2+1] = v, v
				}
			} else if opened {
				mtRender(out[f*2 : (f+n)*2])
			} else {
				clear(out[f*2 : (f+n)*2])
			}
		}
		f += n
		if tickAcc += n * pitHz; tickAcc >= tickPIT*outRate {
			tickAcc -= tickPIT * outRate
			if len(ticks) > 0 {
				play(ticks[0])
				ticks = ticks[1:]
			}
		}
	}
	sg := sfxV
	for f := 0; f < frames; f++ {
		l, r := int(out[f*2]), int(out[f*2+1])
		if speech != nil {
			if i := int(speechF); i < len(speech) {
				v := int(float64((int(speech[i])-128)<<7) * sg)
				l, r = l+v, r+v
				speechF += speechD
			} else {
				speech = nil
			}
		}
		binary.LittleEndian.PutUint16(b[f*4:], uint16(clamp16(l)))
		binary.LittleEndian.PutUint16(b[f*4+2:], uint16(clamp16(r)))
	}
	return frames * 4, nil
}

func clamp16(v int) int16 { return int16(max(-32768, min(32767, v))) }
