package host

import "github.com/hajimehoshi/ebiten/v2"

// The keyboard is fed to the game as the BIOS keyboard buffer saw it (int 16h): one word per
// key press, the scan code in the high byte and the ASCII code in the low byte. The game reads
// it through bioskey (see the mm3 package).

// KeyQueue holds key words the game has not read yet.
var KeyQueue []uint16

// Shift is the BIOS shift state (bioskey(2)): bit 0/1 right/left shift, bit 2 ctrl, bit 3 alt.
var Shift int

type bios struct{ scan, ascii, shifted, ctrl byte }

// keys not delivered through AppendInputChars (or whose BIOS word carries a scan code the
// game tests): scan code and the ASCII byte plain / with shift / with ctrl
var specialKeys = map[ebiten.Key]bios{
	ebiten.KeyEscape: {0x01, 0x1b, 0x1b, 0x1b}, ebiten.KeyBackspace: {0x0e, 8, 8, 0x7f},
	ebiten.KeyTab: {0x0f, 9, 0, 0}, ebiten.KeyEnter: {0x1c, 13, 13, 10}, ebiten.KeyNumpadEnter: {0x1c, 13, 13, 10},
	ebiten.KeyF1: {0x3b, 0, 0, 0}, ebiten.KeyF2: {0x3c, 0, 0, 0}, ebiten.KeyF3: {0x3d, 0, 0, 0},
	ebiten.KeyF4: {0x3e, 0, 0, 0}, ebiten.KeyF5: {0x3f, 0, 0, 0}, ebiten.KeyF6: {0x40, 0, 0, 0},
	ebiten.KeyF7: {0x41, 0, 0, 0}, ebiten.KeyF8: {0x42, 0, 0, 0}, ebiten.KeyF9: {0x43, 0, 0, 0},
	ebiten.KeyF10:  {0x44, 0, 0, 0},
	ebiten.KeyHome: {0x47, 0, '7', 0}, ebiten.KeyArrowUp: {0x48, 0, '8', 0}, ebiten.KeyPageUp: {0x49, 0, '9', 0},
	ebiten.KeyArrowLeft: {0x4b, 0, '4', 0}, ebiten.KeyArrowRight: {0x4d, 0, '6', 0}, ebiten.KeyEnd: {0x4f, 0, '1', 0},
	ebiten.KeyArrowDown: {0x50, 0, '2', 0}, ebiten.KeyPageDown: {0x51, 0, '3', 0}, ebiten.KeyInsert: {0x52, 0, '0', 0},
	ebiten.KeyDelete: {0x53, 0, '.', 0},
}

// scan codes of the printable keys, by unshifted ASCII
var asciiScan = map[byte]byte{}

func init() {
	rows := []struct {
		first byte
		keys  string
	}{{0x02, "1234567890-="}, {0x10, "qwertyuiop[]"}, {0x1e, "asdfghjkl;'`"}, {0x2b, "\\zxcvbnm,./"}}
	for _, r := range rows {
		for i := 0; i < len(r.keys); i++ {
			asciiScan[r.keys[i]] = r.first + byte(i)
		}
	}
	asciiScan[' '] = 0x39
	for k, v := range map[byte]byte{'!': '1', '@': '2', '#': '3', '$': '4', '%': '5', '^': '6', '&': '7', '*': '8',
		'(': '9', ')': '0', '_': '-', '+': '=', '{': '[', '}': ']', ':': ';', '"': '\'', '~': '`', '|': '\\',
		'<': ',', '>': '.', '?': '/'} {
		asciiScan[k] = asciiScan[v]
	}
	for c := byte('A'); c <= 'Z'; c++ {
		asciiScan[c] = asciiScan[c+32]
	}
}

// PushChar queues the key word of a printable character.
func PushChar(c byte) { KeyQueue = append(KeyQueue, uint16(asciiScan[c])<<8|uint16(c)) }

var keysDown = map[ebiten.Key]bool{}

// Mouse state, in the game's 320x200 screen (top row 0).
var (
	MouseX, MouseY int
	MouseButtons   int // bit 0 left, bit 1 right
)

func pollInput(g *game) {
	if runScript() {
		return
	}
	Shift = 0
	if ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		Shift |= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		Shift |= 2
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl {
		Shift |= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		Shift |= 8
	}
	for k, b := range specialKeys {
		down := ebiten.IsKeyPressed(k)
		if down && !keysDown[k] {
			a := b.ascii
			if Shift&3 != 0 {
				a = b.shifted
			} else if ctrl {
				a = b.ctrl
			}
			KeyQueue = append(KeyQueue, uint16(b.scan)<<8|uint16(a))
		}
		keysDown[k] = down
	}
	if ctrl { // ctrl+letter: the control code (the game's ^S, ^Q, ...)
		for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
			down := ebiten.IsKeyPressed(k)
			if down && !keysDown[k] {
				c := byte('a' + k - ebiten.KeyA)
				KeyQueue = append(KeyQueue, uint16(asciiScan[c])<<8|uint16(c-'a'+1))
			}
			keysDown[k] = down
		}
	} else {
		for _, r := range ebiten.AppendInputChars(nil) {
			if r < 0x80 && r >= ' ' {
				PushChar(byte(r))
			}
		}
	}
	if len(KeyQueue) > 16 { // the BIOS buffer holds 15 keys
		KeyQueue = KeyQueue[len(KeyQueue)-16:]
	}

	cx, cy := ebiten.CursorPosition()
	if g.scaleX > 0 {
		MouseX = min(max(int((float64(cx)-g.offX)/g.scaleX), 0), ScreenW-1)
		MouseY = min(max(int((float64(cy)-g.offY)/g.scaleY), 0), ScreenH-1)
	}
	b := 0
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		b |= 1
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		b |= 2
	}
	MouseButtons = b
}

// Timers are callbacks run every host tick before the game runs, as the DOS timer
// interrupt ran them.
var timers []func()

// AddTimer registers a callback run once per host tick.
func AddTimer(f func()) { timers = append(timers, f) }

func tickTimers() {
	for _, f := range timers {
		f()
	}
}
