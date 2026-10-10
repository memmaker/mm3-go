package mm3

// Port of vga (MM3.CC), the game's 256-colour video driver: a raw 16-bit blob that main
// loads at offset 0 of its own segment and calls through the sys.asm thunks (push seg;
// push entry; retf). Only its VGA (mode 13h) half is ported; the Tandy/EGA/CGA paths
// (video modes 1-5, translated through xlat tables) and the copy protection are left out.
// The driver's own tables (palette, window border tiles, font colour sets, scale masks)
// are sliced from ccLoad("vga") at their offsets; its font is the file "fo.t".
//
// Entry points (offset in the jump table at 0 -> routine), Go signatures:
//
//	0x00 -> 0x0f3c  drawSprites(list []byte)        sprites over a copy of the offscreen, shown
//	0x03 -> 0x1498  clearScreen()                   NOT a clear: hides and disables the mouse cursor
//	0x06 -> 0x361e  closeWindows(n int)             restores the background of the n top windows
//	0x09 -> 0x0fbf  -                               retf, unused
//	0x0c -> 0x125c  showRawPicture(name string)     a 64000-byte .raw into the offscreen, shown
//	0x0f -> 0x1138  fadeScreen(in, step int)        in != 0: black -> palette, else palette -> black
//	0x12 -> 0x29f6  -                               shutdown (free buffers, close MM3.$$$); never called
//	0x15 -> 0x0e92  drawIcon(img Far, frame, x, y, attr int)  one frame into the offscreen + its rect shown
//	0x18 -> 0x1456  getMouseState(x, y *int16) int  buttons (bit 0 left, 1 right); waits one retrace
//	0x1b -> 0x1c10  vgaIdle1b()                     spins the text cursor at the text position
//	0x1e -> 0x3370  openTextWindow(x, y, w, h, color, fill, a7, a8 int, text Far)
//	0x21 -> 0x1312  loadGraphic(name string) Far
//	0x24 -> 0x1304  freeGraphic(p Far)
//	0x27 -> 0x13da  setMouseCursor(frame int, cur uint32)  cur: the graphic's far pointer (C long)
//	0x2a -> 0x0fc0  drawSprites2a(list []byte)      starfield + sprites (intro/ending)
//	0x2d -> 0x15b6  printWindowText(s []byte)
//	0x30 -> 0x2820  vgaInit()
//	0x33 -> 0x12ee  -                               load_cc_file again; not reached through a thunk
//
// openTextWindow's text is a far pointer (0: none); fill, a7 and a8 are ignored by the VGA
// driver (the fill colour is always 0xf0).
//
// Screen model. The driver draws everything into a 64000-byte offscreen buffer and copies
// rectangles of it to VRAM (vram here) after waiting for the vertical retrace (host.Yield).
// The mouse cursor is drawn into the offscreen too (its background saved and restored around
// every operation), so vram holds it; host.Frame's cursor overlay is not used. There is no
// interrupt work: the cursor only moves when the game calls into the driver.
//
// Graphic files (.icn/.fac/...; loadGraphic keeps them as they are):
//
//	word nframes; nframes * {word cell1, word cell2}   frame n's pair at 4n+2; cell2 0 = none
//	cell: word x, w, y, rows; then rows * row          x, y: offsets added to the draw position
//	row:  word len (0 = empty row; len counts the bytes after it), word xoff, tokens:
//	      00-7f: n+1 literal pixels follow;  80-bf: skip (n&3f)+1 (transparent);
//	      c0-ff: (n&3f)+3 copies of the next byte
//
// Rows with y outside 0..199 are skipped; x is not clipped (off-edge pixels wrap into the
// neighbouring row, as on the real thing) unless attr bit 1 asks for the 3D-view clip.
// attr: 0x0001 mirror, 0x0002 clip to x 8..223, 0x0100/0x0200/0x0300 scale down by the
// pixel masks 0xdb6d (~3/4), 0xa529 (~1/2), 0x8421 (~1/4); 0x8000 enlarge 2x, 0xc000 3x.
// Frames are always read from offset 0 of the graphic's segment: the offset half of a Far
// is ignored, as the driver does.
//
// Draw lists (drawSprites, drawSprites2a, text code 5): little-endian int16 words.
//
//	{-1, Far img}            (6 bytes) select the graphic (only its segment is used)
//	{-1, x, 0}               end: an img entry with segment 0 (so struct spr's {-1, NULL})
//	{x, y, attr, frame}      (8 bytes) draw; x, y are relative to the text window's text
//	                         origin unless text flag d, b or k is set (\x03d etc.)
//
// So struct spr {flag=-1, img, x, y, attr, frame} is a select + a draw entry, and a list of
// them ends with a {-1, NULL}. drawSprites and drawSprites2a set the text origin to 0,0 first.
//
// Text windows: openTextWindow(x, y, w, h, color, ...) saves the screen under the rectangle
// (a stack of 7; the original swaps it to MM3.$$$), fills the text area (x+8..x+w-8,
// y+8..y+h-8) with 0xf0, draws the 8x8 border tiles, sets the font colour set `color` and
// prints text. closeWindows(n) restores the rectangle and pops the previous window state
// (text area, cursor, alignment, colours).
//
// Text: bytes are masked with 0x7f. 0x21-0x7f draw a glyph of the current font (0x80 adds
// the small font), 0x20 advances 4 (small: 3). Lines wrap at the text area's right edge,
// breaking after ' ', '-', '/' or '\n' (else mid-word, one character early); a line that
// does not fit at the bottom scrolls the text area up unless mode f is set (then printing
// stops). Control codes (digits are decimal, ' ' counts as '0'):
//
//	00       end                       01 big font            02 small font
//	03 c     c: 'l' left, 'c' centre, 'r' right align; 't' scroll at the bottom, 'f' stop at
//	         the bottom; 'd' absolute coords (x, y and draw lists not relative to the text
//	         area, no wrap, and newline returns to x 116) and show rect 2,4 228x144 at the end;
//	         'b'/'k'/'q' show only rect 232,76 80x63 / 10,148 300x44 / 11,10 154x85 at the
//	         end (b and k also make draw lists absolute); 'mN' dim the 3D view (1-4:
//	         dither patterns, 4 = black)
//	04 nnn   fill nnn pixels from the cursor, 9 rows (small font 8), fill colour; right
//	         aligned: leftwards. The cursor does not move
//	05 hhhh  draw list at offset hhhh (4 upper-case hex digits, C's near "%p") of the string's
//	         segment: always DGROUP, so here ds (dsSeg) - the C's "\x05%p" works unchanged
//	06       draw glyph 0x20 (a hard space of the font's own width)
//	07 nnn   fill colour nnn (invalid: the window's default)
//	08 c     back up over c and erase it with the fill colour (' ': back 2/1 pixels)
//	09 nnn   x = left + nnn (clamped to the right edge)
//	0a       newline (10 pixels, small font 9); leading spaces of the next line skipped
//	0b nnn   y = top + nnn (clamped to the bottom)
//	0c nn    font colour set nn (invalid, e.g. "\x0cd": the window's default)
//	0d       clear the text area and home the cursor
//	0e-1f    ignored
//
// Surprises: clear_screen only turns the mouse cursor off; draw_sprites never changes the
// offscreen (it draws into a temporary copy), while draw_sprites_2a clears it; fade_screen
// leaves the driver's working palette filled with 0xff, which is also what restarts the
// starfield (the stars live in that buffer); the timer.drv the game hands to init is only
// checksummed (copy protection: it holds no timer code the driver uses), see vgaInit.

import (
	"time"

	"github.com/memmaker/mm3-go/internal/host"
)

var (
	// vram is A000:0000 in mode 13h (the boss key reads and writes it directly).
	vram [64000]byte
	// palette is the VGA DAC, 6-bit values.
	palette [768]byte

	drv    []byte // the driver image: tables, and its 0x36 work palette / star buffer
	font   []byte // fo.t: 256 glyphs of 8 words (2 bits a pixel), then 256 widths
	screen []byte // offscreen buffer (0x9f8); 64K so every 16-bit offset is in range

	listImg  Far // 0x9fa: graphic selected by a draw list
	fontSel  byte
	glyphCol [4]byte // 0xb74 colours of glyph pixel values 1-3

	waitRetrace = host.Yield // 0x2a14; tests replace it
	frame       = host.Frame{Pix: vram[:], Palette: &palette}
)

const (
	offWorkPal   = 0x36   // working palette (what goes to the DAC) / star state
	offPalette   = 0x39c  // the game palette
	offColorSets = 0xb78  // font colour sets, 4 bytes each
	offTiles     = 0x37d2 // 20 window border tiles, 8x8
	offStarCols  = 0xe62
	offBreaks    = 0xe6b // 10 characters a line may break after
	offSpinner   = 0xe79 // text cursor glyphs
)

// vgaInit is init (0x2820) for video mode 0: mode 13h with a black DAC, the row table,
// the offscreen buffer and the font. Left out: the "afts" checksum of main's code (it only
// yields the mode number 13h), the checksum of timer.drv (it only yields the row pitch 320),
// and creating MM3.$$$ (windows are saved in memory).
func vgaInit() {
	drv = append([]byte(nil), ccLoad("vga")...)
	for i := offWorkPal; i < offWorkPal+0x300; i++ {
		drv[i] = 0
	}
	for y := 0; y < 200; y++ {
		setLe16(drv, 0x69c+2*y, y*320)
	}
	screen = make([]byte, 0x10000)
	font = ccLoad("fo.t")
	copy(glyphCol[:], drv[0xb74:])
	fontSel = 0
	win = textWin{w: 320, h: 200, color: 1, fill: 0x8888, right: 320, bottom: 200, defColor: 1, defFill: 0x8080}
	winDepth = 0
	mouseX, mouseY, curOldW, curOldH = 0x11a, 0x5e, 4, 4
	starSeed1, starSeed2 = 0xb289, 0x13ca
	vram = [64000]byte{}
	host.ShowFrame = func() *host.Frame { return &frame }
	loadDAC()
}

// loadDAC (0x11e4) waits for the retrace and loads the DAC from the working palette.
func loadDAC() {
	waitRetrace()
	copy(palette[:], drv[offWorkPal:offWorkPal+0x300])
}

// copyRect (0x2a2a) copies h rows of w bytes at (x, y) from the offscreen to VRAM.
func copyRect(x, y, h, w uint16) {
	bx := y << 1
	for {
		di := rowOff(bx) + x
		for i := uint16(0); i < w; i++ {
			if o := di + i; o < 64000 {
				vram[o] = screen[o]
			}
		}
		bx += 2
		if h--; h == 0 {
			return
		}
	}
}

func showAll() { copy(vram[:], screen[:64000]) }

// segBytes is the memory of segment seg (nil if there is none: reads give 0).
func segBytes(seg uint16) []byte {
	if int(seg) < len(segs) {
		return segs[seg]
	}
	return nil
}

// drawList (0x1bb8) draws a draw list into the offscreen (format in the header).
func drawList(l []byte) {
	for p := uint16(0); ; {
		x := rd16(l, p)
		if x == 0xffff {
			seg := rd16(l, p+4)
			if seg == 0 {
				return
			}
			listImg = Far(uint32(seg) << 16)
			p += 6
			continue
		}
		y := rd16(l, p+2)
		if flagD == 0 && flagB == 0 && flagK == 0 {
			x += win.left
			y += win.top
		}
		drawFrame(segBytes(listImg.Seg()), rd16(l, p+6)*4+2, x, y, rd16(l, p+4))
		p += 8
	}
}

// drawSprites (0x0f3c) draws list over a copy of the offscreen and shows it, leaving the
// offscreen as it was.
func drawSprites(list []byte) {
	win.left, win.top = 0, 0
	saved := screen
	screen = append([]byte(nil), saved...)
	drawList(list)
	waitRetrace()
	showAll()
	screen = saved
}

// drawSprites2a (0x0fc0) clears the offscreen, draws the starfield and list into it, shows it.
func drawSprites2a(list []byte) {
	starTwinkle ^= 1
	win.left, win.top = 0, 0
	for i := 0; i < 0xfa00; i++ {
		screen[i] = 0
	}
	stars()
	drawList(list)
	waitRetrace()
	showAll()
}

var (
	starTwinkle          byte   // 0xe8c: star colours step down every other frame
	starSeed1, starSeed2 uint16 // 0x9c2, 0x9c4
)

// stars (0x0ffd): 75 stars flying out from the centre (x 0x2800, y 0x1900 in 1/64 pixels),
// accelerating by 1/16 a frame, fading through a palette ramp. Their state is in the
// working palette buffer: x, vx, y, vy words at 0x36, +0x96, +0x12c, +0x1c2, colour and end
// colour bytes at +0x258. Seeded from the DOS clock, as the original.
func stars() {
	w := func(o uint16) uint16 { return le16(drv, offWorkPal+int(o)) }
	sw := func(o, v uint16) { setLe16(drv, offWorkPal+int(o), int(v)) }
	for cx := 0x4a; cx >= 0; cx-- {
		bp := uint16(cx * 2)
		vx := int16(w(bp + 0x96))
		vx += vx >> 4
		sw(bp+0x96, uint16(vx))
		x := w(bp) + uint16(vx)
		sw(bp, x)
		var y uint16
		reinit := x >= 0x5000
		if !reinit {
			vy := int16(w(bp + 0x1c2))
			vy += vy >> 4
			sw(bp+0x1c2, uint16(vy))
			y = w(bp+0x12c) + uint16(vy)
			sw(bp+0x12c, y)
			reinit = y >= 0x3200 || drv[offWorkPal+int(bp)+0x258] == drv[offWorkPal+int(bp)+0x259]
		}
		if reinit {
			c := drv[offStarCols+int((starSeed2-starSeed1)&3)]
			drv[offWorkPal+int(bp)+0x259] = c - 0x1f
			drv[offWorkPal+int(bp)+0x258] = c
			ch, cl, dh, dl := dosTime()
			dx := uint32(dh)<<8 | uint32(dl)
			dx += uint32(ch)<<8 | uint32(cl)
			carry := dx >> 16
			dx = dx&0xffff + uint32(starSeed1) + carry
			carry = dx >> 16
			starSeed1 = uint16(dx)
			r := uint32(starSeed1)<<1 | carry // rcl dx,1
			starSeed2 = uint16(uint32(starSeed2) - r&0xffff - (r >> 16))
			d := uint16(r)
			dl8 := byte(d) ^ byte(d>>8)
			sw(bp+0x96, uint16(int16(int8(dl8))))
			x = (cbwHi(dl8) | uint16(byte(starSeed2))) * 4
			x += 0x2800
			sw(bp, x)
			ch, cl, dh, dl = dosTime()
			d2 := (uint16(dh)<<8 | uint16(dl)) - (uint16(ch)<<8 | uint16(cl))
			d2 ^= starSeed2
			starSeed2 = d2
			starSeed1 += d2
			dl8 = byte(d2) + byte(d2>>8)
			sw(bp+0x1c2, uint16(int16(int8(dl8))))
			y = (cbwHi(dl8)|uint16(byte(starSeed1)))*4 + 0x1900
			sw(bp+0x12c, y)
		}
		ax := uint16(int16(y) >> 6)
		di := uint16(int16(x)>>6) + ax*320
		c := drv[offWorkPal+int(bp)+0x258]
		put(di, c)
		if c != drv[offWorkPal+int(bp)+0x259] {
			drv[offWorkPal+int(bp)+0x258] = c - starTwinkle
		}
	}
}

// cbwHi is the high byte cbw makes of al, as a word.
func cbwHi(al byte) uint16 {
	if al&0x80 != 0 {
		return 0xff00
	}
	return 0
}

// dosTime is int 21h function 2Ch: hour, minute, second, 1/100 s.
func dosTime() (ch, cl, dh, dl byte) {
	t := time.Now()
	return byte(t.Hour()), byte(t.Minute()), byte(t.Second()), byte(t.Nanosecond() / 1e7)
}

// fadeScreen (0x1138) fades the palette in (in != 0) or out in steps of step/128, one
// retrace a step, then fills the working palette with 0xff.
func fadeScreen(in, step int) {
	cx := 0x80
	for {
		lvl := cx
		if in != 0 {
			lvl = 0x80 - cx
		}
		for i := 0; i < 0x300; i++ {
			drv[offWorkPal+i] = byte(int(drv[offPalette+i]) * lvl * 2 >> 8)
		}
		cursorHide()
		cursorShow()
		loadDAC()
		cursorFlush()
		if cx == 0 {
			break
		}
		if cx -= step; cx < 0 {
			cx = 0
		}
	}
	for i := offWorkPal; i < offWorkPal+0x300; i++ {
		drv[i] = 0xff
	}
}

// showRawPicture (0x125c) loads a raw 320x200 picture into the offscreen and shows it.
func showRawPicture(name string) {
	copy(screen, ccLoad(name))
	waitRetrace()
	showAll()
}

// loadGraphic (0x1312) loads a graphic file; the VGA driver keeps it as it is.
func loadGraphic(name string) Far { return newSeg(ccLoad(name)) }

// freeGraphic (0x1304)
func freeGraphic(p Far) { freeFar(p) }

// drawIcon (0x0e92) draws a frame of img at (x, y) into the offscreen and shows its
// rectangle: the first cell's, or for two cells min(x)..max(w) by min(y)..max(rows) (sic).
func drawIcon(img Far, frame, x, y, attr int) {
	cursorHide()
	g := segBytes(img.Seg())
	si := uint16(frame)*4 + 2
	X, Y := uint16(x), uint16(y)
	drawFrame(g, si, X, Y, uint16(attr))
	c1, c2 := rd16(g, si), rd16(g, si+2)
	var rx, ry, rh, rw uint16
	if c2 == 0 {
		rx, rw, ry, rh = rd16(g, c1)+X, rd16(g, c1+2), rd16(g, c1+4)+Y, rd16(g, c1+6)
	} else {
		rx = min(rd16(g, c2)+X, rd16(g, c1)+X)
		rw = max(rd16(g, c2+2)+X, rd16(g, c1+2)+X) - rx
		ry = min(rd16(g, c2+4)+Y, rd16(g, c1+4)+Y)
		rh = max(rd16(g, c2+6)+Y, rd16(g, c1+6)+Y) - ry
	}
	waitRetrace()
	copyRect(rx, ry, rh, rw)
	cursorShow()
	cursorFlush()
}

// Mouse cursor state (0xb3e..0xb56, 0xe75).
var (
	mouseOn              bool   // 0xe75
	mouseButtons         uint16 // 0xb3e
	mouseX, mouseY       uint16 // 0xb40
	mouseOldX, mouseOldY uint16 // 0xb44
	curH, curW           uint16 // 0xb48, 0xb4a: the cursor cell's rows and w
	curOldH, curOldW     uint16 // 0xb4c: those of the last shown cursor
	curFrame             uint16 // 0xb50: 4*frame+2
	curImg               Far    // 0xb52
	curSaveX, curSaveRow uint16 // 0xb54, 0xb56
	curSave              []byte // 0xc22
)

// readMouse is int 33h function 3 within the ranges set by setMouseCursor (x 4..309, y 3..186).
func readMouse() {
	mouseButtons = uint16(host.MouseButtons)
	mouseX = uint16(min(max(host.MouseX, 4), 309))
	mouseY = uint16(min(max(host.MouseY, 3), 186))
}

// cursorShow (0x14b2) reads the mouse, saves the offscreen under the cursor and draws it.
func cursorShow() {
	if !mouseOn {
		return
	}
	mouseOldX, mouseOldY = mouseX, mouseY
	readMouse()
	curSaveX, curSaveRow = mouseX, mouseY<<1
	n := (curW + 1) >> 1 * 2
	curSave = curSave[:0]
	bx := curSaveRow
	for r := curH; ; {
		si := rowOff(bx) + curSaveX
		for i := uint16(0); i < n; i++ {
			curSave = append(curSave, screen[si+i])
		}
		bx += 2
		if r--; r == 0 {
			break
		}
	}
	drawFrame(segBytes(curImg.Seg()), curFrame, mouseX, mouseY, 0)
}

// cursorHide (0x1534) puts back the offscreen under the cursor.
func cursorHide() {
	if !mouseOn {
		return
	}
	n := (curW + 1) >> 1 * 2
	bx, k := curSaveRow, 0
	for r := curH; ; {
		di := rowOff(bx) + curSaveX
		for i := uint16(0); i < n; i++ {
			if k < len(curSave) {
				screen[di+i] = curSave[k]
			}
			k++
		}
		bx += 2
		if r--; r == 0 {
			break
		}
	}
}

// cursorFlush (0x1572) shows the cursor's old and new rectangles.
func cursorFlush() {
	if !mouseOn {
		return
	}
	copyRect(mouseOldX, mouseOldY, curOldH, curOldW)
	curOldH, curOldW = curH, curW
	copyRect(mouseX, mouseY, curH, curW)
}

// setMouseCursor (0x13da) turns the mouse cursor on with frame of img (its hot spot is the
// cell's x, y offset).
func setMouseCursor(frame int, cur uint32) {
	img := Far(cur)
	if !mouseOn {
		mouseOn = true // int 33h reset, hide, position (0x11a, 0x5e): the host owns the pointer
	} else {
		cursorHide()
	}
	curFrame = uint16(frame)*4 + 2
	curImg = img
	g := segBytes(img.Seg())
	c := rd16(g, curFrame)
	curW, curH = rd16(g, c+2), rd16(g, c+6)
	cursorShow()
	cursorFlush()
}

// clearScreen (0x1498) despite its name only turns the mouse cursor off (until the next
// setMouseCursor).
func clearScreen() {
	if mouseOn {
		cursorHide()
	}
	mouseOn = false
	curSaveX = 0
}

// getMouseState (0x1456) moves the cursor, waits for a retrace and returns the buttons and
// the position (0, 0, 0 while the cursor is off).
func getMouseState(x, y *int16) int {
	if !mouseOn {
		waitRetrace()
		*x, *y = 0, 0
		return 0
	}
	cursorHide()
	cursorShow()
	waitRetrace()
	cursorFlush()
	*x, *y = int16(mouseX), int16(mouseY)
	return int(mouseButtons)
}
