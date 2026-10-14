package mm3

import "testing"

// One second of host ticks gives ~72.8 driver ticks; a sample of n bytes ends after
// n samples at ~8 kHz; a fade with flag 0xc0 ends after 64 carries of 0x60.
func TestSndTimer(t *testing.T) {
	sndLoaded = true
	sndTimer(-2)
	for i := 0; i < 70; i++ {
		sndHostTick()
	}
	if n := sndTicks; n < 72 || n > 73 {
		t.Fatalf("ticks/s = %d", n)
	}
	sndSampleLen, sndSamplePos, sndSampleOn = 8008, 0, true
	for i := 0; i < 69; i++ {
		sndHostTick()
	}
	if !sndSampleOn {
		t.Fatal("sample ended early")
	}
	sndHostTick()
	sndHostTick()
	if sndSampleOn {
		t.Fatal("sample still playing")
	}
	a := newAdlib(func(int, byte) {})
	sound = a
	defer func() { sound = nil }()
	a.startSong([]byte{acWait*16 + 1, 200, acLoop})
	sndPlayMusic(0xc0 << 16)
	n := 0
	for sndPlayMusic(-1) != 0 {
		sndTickIRQ()
		n++
	}
	if a.musicOn || n < 168 || n > 172 {
		t.Fatalf("fade took %d ticks", n)
	}
}
