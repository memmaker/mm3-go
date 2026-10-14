package mm3

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/opl2"
)

// adlibRender plays ticks driver ticks on an OPL2 at 44.1 kHz.
func adlibRender(a *adlib, out *opl2.Resampler, ticks int, each func(i int)) []int16 {
	var pcm []int16
	frames := 0
	for i := range ticks {
		if each != nil {
			each(i)
		}
		a.tick()
		for ; frames < (i+1)*44100*sndTickDiv/pitHz; frames++ {
			pcm = append(pcm, int16(max(-32768, min(32767, out.Next()))))
		}
	}
	return pcm
}

func rms(pcm []int16) float64 {
	s := 0.0
	for _, v := range pcm {
		s += float64(v) * float64(v)
	}
	return math.Sqrt(s / float64(max(1, len(pcm))))
}

func newTestAdlib() (*adlib, *opl2.Resampler) {
	chip := opl2.New()
	a := newAdlib(chip.Write)
	a.init()
	return a, opl2.NewResampler(chip, 44100)
}

// Every song plays audibly for a minute, with effects over it, and fades out; the title
// song goes to $TMPDIR/mm3theme-adlib.wav.
func TestAdlibEverySong(t *testing.T) {
	rolandData(t)
	for _, s := range []string{"bank.m", "caves.m", "city.m", "combat.m", "cyber.m", "dungeon.m", "eerie.m", "grounds.m", "guild.m", "honky.m", "medieval.m", "mm3theme.m", "temples.m", "towninn.m", "venture.m"} {
		a, out := newTestAdlib()
		a.startSong(ccLoad(s))
		pcm := adlibRender(a, out, 73*60, func(i int) {
			if i%977 == 500 {
				a.startSfx(i % len(adlibSfx))
			}
		})
		if r := rms(pcm); r < 100 {
			t.Errorf("%s: RMS %.0f, silent", s, r)
		}
		if s == "mm3theme.m" {
			writeWav(t, "mm3theme-adlib.wav", pcm)
		}
		a.music(255)
		adlibRender(a, out, 73*10, nil)
		if a.musicOn || a.music(-1) != 0 {
			t.Errorf("%s: still playing ten seconds into the fastest fade", s)
		}
	}
}

// Every effect sounds, but the two the driver ignores and those that play sfx 0's script,
// which keys both channels with an instrument whose carrier is turned all the way down.
func TestAdlibEverySfx(t *testing.T) {
	var all []int16
	for v := range adlibSfx {
		a, out := newTestAdlib()
		a.startSfx(v)
		pcm := adlibRender(a, out, 73*2, nil)
		silent := adlibIgnored[v] || adlibSfx[v] == adlibSfx[0]
		if r := rms(pcm); r < 100 != silent {
			t.Errorf("sfx %d: RMS %.0f", v, r)
		}
		all = append(all, pcm...)
	}
	writeWav(t, "mm3sfx-adlib.wav", all)
}

func writeWav(t *testing.T, name string, pcm []int16) {
	h := make([]byte, 44, 44+2*len(pcm))
	copy(h, "RIFF")
	binary.LittleEndian.PutUint32(h[4:], uint32(36+2*len(pcm)))
	copy(h[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(h[16:], 16)
	binary.LittleEndian.PutUint16(h[20:], 1)
	binary.LittleEndian.PutUint16(h[22:], 1)
	binary.LittleEndian.PutUint32(h[24:], 44100)
	binary.LittleEndian.PutUint32(h[28:], 2*44100)
	binary.LittleEndian.PutUint16(h[32:], 2)
	binary.LittleEndian.PutUint16(h[34:], 16)
	copy(h[36:], "data")
	binary.LittleEndian.PutUint32(h[40:], uint32(2*len(pcm)))
	for _, v := range pcm {
		h = binary.LittleEndian.AppendUint16(h, uint16(v))
	}
	p := filepath.Join(os.TempDir(), name)
	os.WriteFile(p, h, 0o644)
	t.Logf("%s", p)
}
