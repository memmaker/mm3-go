package mm3

import (
	"fmt"
	"os"
)

// Port of sys.asm's own routines (its video-driver thunks are vga.go).

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "mm3:", msg)
	os.Exit(1)
}

// fatalDiskError: the original prints its disk error message and exits to DOS.
func fatalDiskError() { fatal("disk error (MM3.CC or MM3.CUR missing or damaged)") }

// randomState is random_range's seed, four words kept in sys.asm's code segment (L2a22..L2a28).
var randomState = [4]uint16{0x4e7b, 0xa109, 0xfc62, 0x83d5}

// adc16/sbb16 are adc/sbb: a+b+cf, a-b-cf and the carry out.
func adc16(a, b uint16, cf bool) (uint16, bool) {
	s := uint32(a) + uint32(b) + uint32(b2i(cf))
	return uint16(s), s > 0xffff
}

func sbb16(a, b uint16, cf bool) (uint16, bool) {
	c := uint32(b) + uint32(b2i(cf))
	return uint16(uint32(a) - c), uint32(a) < c
}

// randomRange returns a value in a..b (a when a == b). It mixes the DOS time (int 21h
// ah=2Ch; dx = seconds:hundredths) into the seed through one carry chain. The carry into the
// chain is the one cmp a,b left (int 21h keeps the caller's flags); the carry into the last
// sbb is the one `sub cx,a` left (div leaves the flags alone, as DOSBox does).
func randomRange(a, b int) int {
	ua, ub := uint16(a), uint16(b)
	if ua == ub {
		return int(int16(ua))
	}
	_, _, dh, dl := dosTime()
	dx := uint16(dh)<<8 | uint16(dl)
	s := &randomState
	cf := ua < ub
	dx, cf = sbb16(dx, s[0], cf)
	ax := s[1]
	s[0], cf = adc16(s[0], ax, cf)
	ax = s[2]
	s[1], cf = sbb16(s[1], ax, cf)
	ax = s[3]
	s[2], cf = adc16(s[2], ax, cf)
	s[3], _ = sbb16(s[3], dx, cf)
	cx := ub - ua
	cf = ub < ua
	cx++
	q, r := ax/cx, ax%cx
	s[2], _ = sbb16(s[2], q, cf)
	return int(int16(r + ua))
}
