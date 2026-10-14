package snd

import (
	"encoding/binary"
	"testing"

	"github.com/memmaker/opl2"
)

// Register writes queued in a tick sound once the stream reaches it.
func TestOPLStream(t *testing.T) {
	OpenOPL()
	for _, w := range [][2]int{
		{opl2.RegOpFlags + 3, 1}, {opl2.RegOpLevel + 3, 0}, {opl2.RegOpAttackDecay + 3, 240},
		{opl2.RegOpSustainRelease + 3, 15}, {opl2.RegOpLevel, 63},
		{opl2.RegFreqLow, 68}, {opl2.RegFreqHigh, 1<<5 | 4<<2 | 2},
	} {
		WriteOPL(w[0], byte(w[1]))
	}
	Tick()
	b := make([]byte, 4*4410)
	Stream{}.Read(b)
	peak := 0
	for i := 0; i < len(b); i += 2 {
		peak = max(peak, abs(int(int16(binary.LittleEndian.Uint16(b[i:])))))
	}
	if peak < 1000 {
		t.Fatalf("peak %d: silent", peak)
	}
}

func abs(v int) int { return max(v, -v) }
