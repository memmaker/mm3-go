package mm3

import (
	"bytes"
	"os"
	"testing"
)

// TestVGAProtectionWindow redraws the copy-protection question (control.raw, fade in, the
// window from copy.bin, input_text's prompt) and compares vram with a DOSBox dump of the same
// screen, MM3_VGA_FIXTURE (64000 bytes, kept outside the repo). The spinner cell is skipped.
func TestVGAProtectionWindow(t *testing.T) {
	if os.Getenv("MM3_DATA") == "" {
		t.Skip("MM3_DATA unset")
	}
	GameDir = os.Getenv("MM3_DATA")
	openCCArchive()
	waitRetrace = func() {}
	vgaInit()
	showRawPicture("control.raw")
	fadeScreen(1, 4)
	q := ccLoad("copy.bin")[:0x10e]
	text := append(csprintf(string(q[:bytes.IndexByte(append(q, 0), 0)]), 4, 4, 6), 0)
	openTextWindow(0x32, 0x23, 0xdc, 0x82, 0, 0x96, 4, 6, farOf(text))
	printWindowText([]byte("\x03l\t000\x04200\x03c\x00"))

	spinX, spinY := int(win.curX), int(win.curY) // the spinner cell, drawn by vgaIdle1b
	path := os.Getenv("MM3_VGA_FIXTURE")
	if path == "" {
		t.Log("MM3_VGA_FIXTURE unset; drew without comparing")
		return
	}
	want, err := os.ReadFile(path)
	if err != nil || len(want) != len(vram) {
		t.Fatalf("fixture %s: %v (%d bytes)", path, err, len(want))
	}
	bad := 0
	for i := range vram {
		x, y := i%320, i/320
		if x >= spinX && x < spinX+8 && y >= spinY && y < spinY+8 {
			continue
		}
		if vram[i] != want[i] {
			if bad < 10 {
				t.Errorf("(%d,%d) = %#x, want %#x", x, y, vram[i], want[i])
			}
			bad++
		}
	}
	if bad > 0 {
		t.Errorf("%d pixels differ", bad)
	}
}
