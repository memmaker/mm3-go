package mm3

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/mm3-go/internal/snd"
)

func rolandData(t *testing.T) (drv, song []byte) {
	dir := os.Getenv("MM3_DATA")
	if dir == "" {
		t.Skip("MM3_DATA (the game directory) not set")
	}
	GameDir = dir
	openCCArchive()
	return ccLoad("roland.drv"), ccLoad("mm3theme.m")
}

// The title song through the roland.drv port for a minute: notes go on and off in
// pairs on channels 2..10 with sane keys, and an sfx script runs too.
func TestRolandSequencer(t *testing.T) {
	drv, song := rolandData(t)
	on := map[[2]byte]bool{}
	ons, offs, other := 0, 0, 0
	r := newRoland(drv, func(m []byte) {
		st, ch := m[0]&0xf0, m[0]&0xf
		if m[0] < 0xf0 && (ch < 1 || ch > 9) {
			t.Fatalf("message % x on channel %d", m, ch+1)
		}
		switch st {
		case 0x90:
			if m[1] > 108 || m[2] == 0 {
				t.Fatalf("note on % x", m)
			}
			on[[2]byte{ch, m[1]}] = true
			ons++
		case 0x80:
			if !on[[2]byte{ch, m[1]}] {
				t.Fatalf("note off % x without its note on", m)
			}
			delete(on, [2]byte{ch, m[1]})
			offs++
		default:
			other++
		}
	})
	r.init()
	r.startSong(song)
	for i := 0; i < 73*60; i++ {
		if i == 300 {
			r.sfx(0x14)
		}
		r.tick()
	}
	t.Logf("%d note ons, %d note offs, %d other, music on %d", ons, offs, other, r.m[rMusicOn])
	if ons < 200 || offs < ons-9 {
		t.Fatalf("%d note ons, %d note offs", ons, offs)
	}
}

// With munt and the ROMs (MT32_ROMS), render 8 s of the title song to $TMPDIR/mm3theme.wav
// and check it is audible.
func TestRolandRender(t *testing.T) {
	drv, song := rolandData(t)
	roms := os.Getenv("MT32_ROMS")
	if roms == "" {
		t.Skip("MT32_ROMS not set")
	}
	if err := snd.Open(roms); err != nil {
		t.Skip(err)
	}
	r := newRoland(drv, snd.Send)
	r.init()
	r.startSong(song)
	var pcm []byte
	buf := make([]byte, 4*441)
	done := 0
	for i := 0; i < 800; i++ { // 8 s in 10 ms reads, ~0.73 ticks each
		for ; float64(done) <= float64(i)*0.728; done++ {
			r.tick()
			snd.Tick()
		}
		snd.Stream{}.Read(buf)
		pcm = append(pcm, buf...)
	}
	sum := 0.0
	for i := 0; i+1 < len(pcm); i += 2 {
		v := float64(int16(binary.LittleEndian.Uint16(pcm[i:])))
		sum += v * v
	}
	rms := math.Sqrt(sum / float64(len(pcm)/2))
	wav := filepath.Join(os.TempDir(), "mm3theme.wav")
	h := make([]byte, 44)
	copy(h, "RIFF")
	binary.LittleEndian.PutUint32(h[4:], uint32(36+len(pcm)))
	copy(h[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(h[16:], 16)
	binary.LittleEndian.PutUint16(h[20:], 1)
	binary.LittleEndian.PutUint16(h[22:], 2)
	binary.LittleEndian.PutUint32(h[24:], 44100)
	binary.LittleEndian.PutUint32(h[28:], 44100*4)
	binary.LittleEndian.PutUint16(h[32:], 4)
	binary.LittleEndian.PutUint16(h[34:], 16)
	copy(h[36:], "data")
	binary.LittleEndian.PutUint32(h[40:], uint32(len(pcm)))
	os.WriteFile(wav, append(h, pcm...), 0o644)
	t.Logf("%s: RMS %.0f (of 32767)", wav, rms)
	if rms < 100 {
		t.Fatalf("RMS %.0f: silent", rms)
	}
}
