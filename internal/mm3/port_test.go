package mm3

import "testing"

func TestPortedHelpers(t *testing.T) {
	if got := gostr(formatHms(3725)); got != "01:02:05" {
		t.Errorf("formatHms = %q", got)
	}
	var b [4]byte
	setBit(b[:], 10, 1)
	if testBit(b[:], 10) == 0 || b[1] != 0x20 {
		t.Errorf("setBit/testBit: %v", b)
	}
	setBit(b[:], 10, 0)
	if b[1] != 0 {
		t.Errorf("clear: %v", b)
	}
	if got := fmtP(cs("\x05%p")); got != "\x05%04X" {
		t.Errorf("fmtP = %q", got)
	}
}
