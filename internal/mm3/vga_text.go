package mm3

// Text windows and the text printer of vga (MM3.CC); formats in vga.go's header.

// textWin is the window state block at 0xa02..0xa21, pushed by openTextWindow.
type textWin struct {
	x, y, w, h  uint16 // 0xa04 the window
	color, fill uint16 // 0xa0c font colour set, 0xa0e fill word
	curX, curY  uint16 // 0xa10 text cursor
	align, mode byte   // 0xa14 0 left 1 right 2 centre; 0xa15 0x80 = stop at the bottom
	left, right uint16 // 0xa16 text area
	top, bottom uint16
	defColor    uint16 // 0xa1e
	defFill     uint16 // 0xa20
}

var (
	win      textWin
	winDepth int        // 0xa00
	winStack [7]textWin // 0xa22
	winSaves [][]byte   // MM3.$$$: the screen under each open window

	txt []byte // 0xb2c: the string being printed, tp the position in it
	tp  uint16

	flagD, flagB, flagK, flagQ byte // 0xe67, 0xe68, 0xe69, 0xe6a (text code 3 d/b/k/q)
	fromOpen                   bool // 0xe66
	textCursorOn               bool // 0xe78
	spinPhase                  int  // 0xb58
)

// nextChar is 0x1b08.
func nextChar() byte {
	c := rd8(txt, tp)
	tp++
	return c & 0x7f
}

func small() bool { return fontSel&0x80 != 0 }

// fillBytes is rep stosw/stosb of a fill word: low byte, high byte, low byte...
func fillBytes(di, n, ax uint16) {
	for i := uint16(0); i < n; i++ {
		if i&1 == 0 {
			put(di+i, byte(ax))
		} else {
			put(di+i, byte(ax>>8))
		}
	}
}

// setColors (0x1af0) selects font colour set win.color.
func setColors() {
	copy(glyphCol[:], drv[offColorSets+int(win.color)*4:])
}

// drawGlyph (0x1b42) draws character c at the text cursor and advances it; g, p, q, y sit one
// row lower.
func drawGlyph(c byte) {
	y := win.curY
	if c == 'g' || c == 'p' || c == 'q' || c == 'y' {
		y++
	}
	glyphAt(rowOff(y<<1)+win.curX, c)
	win.curX += uint16(font[0x1000+int(c|fontSel)])
}

// glyphAt (0x1c5c's loop) draws the 8x8 glyph c of the current font at offset di.
func glyphAt(di uint16, c byte) {
	si := int(c|fontSel) * 16
	for r := 0; r < 8; r++ {
		w := le16(font, si)
		si += 2
		for p := 0; p < 8; p++ {
			if b := w & 3; b != 0 {
				put(di, glyphCol[b])
			}
			w >>= 2
			di++
		}
		di += 0x138
	}
}

// eraseCell (0x1cb0) fills the 8x8 cell at the text cursor with the fill colour.
func eraseCell() {
	di := rowOff(win.curY<<1) + win.curX
	for r := 0; r < 8; r++ {
		fillBytes(di, 8, win.fill)
		di += 0x140
	}
}

// fillText (0x36ca) fills the text area with the fill colour. (Its "whole window" variant
// is dead code: the routine starts with clc.)
func fillText() {
	bx := win.top << 1
	w := win.right - win.left
	for bp := win.bottom - win.top; ; {
		fillBytes(rowOff(bx)+win.left, w, win.fill)
		bx += 2
		if bp--; bp == 0 {
			return
		}
	}
}

// flushText (0x371c) shows what was printed: the rectangle of a pending d/b/k/q flag
// (clearing that flag), else the whole window (all) or its text area.
func flushText(all bool) {
	cursorShow()
	var x, y, h, w uint16
	switch {
	case flagQ != 0:
		flagQ = 0
		x, y, h, w = 0xb, 0xa, 0x55, 0x9a
	case flagK != 0:
		flagK = 0
		x, y, h, w = 0xa, 0x94, 0x2c, 0x12c
	case flagB != 0:
		flagB = 0
		x, y, h, w = 0xe8, 0x4c, 0x3f, 0x50
	case flagD != 0:
		flagD = 0
		x, y, h, w = 2, 4, 0x90, 0xe4
	case all:
		x, y, h, w = win.x, win.y, win.h, win.w
	default:
		x, y, h, w = win.left, win.top, win.bottom-win.top, win.right-win.left
	}
	waitRetrace()
	copyRect(x, y, h, w)
	cursorFlush()
}

// digits (0x1b16) reads n decimal digits (' ' = 0); ok false at a non-digit, which it
// consumes, with that character - '0' as the value.
func digits(n int) (uint16, bool) {
	var bx uint16
	for ; n > 0; n-- {
		al := nextChar()
		if al == ' ' {
			al = '0'
		}
		al -= '0'
		if al > 9 {
			return uint16(al), false
		}
		bx = bx*10 + uint16(al)
	}
	return bx, true
}

// measure (0x19c4) adds the width of the next character to bx; stop at a control code
// (left unread) other than 8 ' ' (backs 2) and 12 (its digits are skipped).
func measure(bx *uint16) (dl byte, stop bool) {
	al := nextChar()
	dl = al
	switch {
	case al > 0x20:
		*bx += uint16(font[0x1000+int(al|fontSel)])
	case al == 0x20:
		if small() {
			*bx += 3
		} else {
			*bx += 4
		}
	case al == 8:
		if nextChar() == 0x20 {
			*bx -= 2
			return
		}
		tp -= 2
		stop = true
	case al == 0xc:
		if nextChar() != 'd' {
			nextChar()
		}
	default:
		tp--
		stop = true
	}
	return
}

// newline (0x1a1c) skips spaces and moves to the next line, scrolling the text area if it
// does not fit. True if it scrolled or (mode f) ran out of room.
func newline() bool {
	for rd8(txt, tp)&0x7f == ' ' {
		tp++
	}
	win.curX = win.left
	if flagD != 0 {
		win.curX = 0x74
	}
	dx := uint16(10)
	if small() {
		dx = 9
	}
	win.curY += dx
	if flagD != 0 {
		return false
	}
	ax := win.bottom - win.curY
	dx--
	if int16(ax) >= int16(dx) {
		return false
	}
	if win.mode&0x80 != 0 {
		return true
	}
	amount := dx - ax
	flushText(false)
	cursorHide()
	w := win.right - win.left
	bx := win.top << 1
	bp := (win.top + amount) << 1
	for cx := win.bottom - (win.top + amount); ; {
		si, di := rowOff(bp)+win.left, rowOff(bx)+win.left
		for i := uint16(0); i < w; i++ {
			screen[di+i] = screen[si+i]
		}
		bx += 2
		bp += 2
		if cx--; cx == 0 {
			break
		}
	}
	win.curY = win.bottom - dx
	for cx := amount; cx > 0; cx-- {
		fillBytes(rowOff(bx)+win.left, w, win.fill)
		bx += 2
	}
	return true
}

func isBreak(c byte) bool {
	for _, b := range drv[offBreaks : offBreaks+10] {
		if c == b {
			return true
		}
	}
	return false
}

// printWindowText (0x15b6) prints s (NUL-terminated, control codes in vga.go's header) in the
// current text window and shows it.
func printWindowText(s []byte) {
	txt, tp = s, 0
	flagD = 0
	if !fromOpen {
		cursorHide()
		if textCursorOn {
			eraseCell()
		}
	}
	textCursorOn = false
	printLines(win.right)
	all := fromOpen // 0x19b6: shr byte [e66],1 hands the flag to the flush as CF
	fromOpen = false
	flushText(all)
}

// printLines is the line loop of printWindowText (0x15f6..0x19b3); limit is 0xb3a, the
// right edge at the start.
func printLines(limit uint16) {
	for { // 0x15f6
		overflow := false
		start := tp
		bx := win.curX
		if win.align != 0 {
			bx = win.left
		}
		stop := false
		for {
			if _, stop = measure(&bx); stop {
				break
			}
			if flagD != 0 || bx <= limit {
				continue
			}
			tp--
			overflow = true
			break
		}
		end := tp // 0xb3c
		tp = start
		if !stop && win.align != 1 && bx >= limit {
			// 0x165a: break after the last break character, or mid-word
			cx := end
			found := false
			if rd8(txt, cx) == ' ' {
				found = isBreak(' ')
			}
			for !found {
				if cx--; !(start < cx) {
					break
				}
				found = isBreak(rd8(txt, cx) & 0x7f)
			}
			if found {
				end = cx
			} else {
				end--
				if win.curX != win.left {
					if newline() && win.mode&0x80 != 0 {
						return
					}
					continue
				}
			}
		}
		if win.align >= 1 { // 0x16ab
			save := tp
			var w uint16
			for {
				dl, st := measure(&w)
				if st {
					break
				}
				if tp > end {
					if dl == ' ' {
						w -= 4
						if small() {
							w++
						}
					}
					break
				}
			}
			tp = save
			if win.align == 1 {
				win.curX -= w - 1
			} else {
				ax := win.curX * 2
				if win.curX == win.left {
					ax = win.left + limit + 1
				}
				win.curX = (ax - w) >> 1
			}
		}
		for { // 0x170e
			if tp > end {
				if win.align != 1 && overflow {
					if newline() && win.mode&0x80 != 0 {
						return
					}
				}
				break
			}
			if done := textCode(nextChar()); done {
				return
			}
		}
	}
}

// textCode (0x173e) prints one character or runs one control code; true at the end.
func textCode(al byte) bool {
	switch {
	case al > 0x20:
		drawGlyph(al)
	case al == 0x20:
		win.curX += 4
		if small() {
			win.curX--
		}
	case al > 0xd:
	case al == 0xd:
		fillText()
		win.curX, win.curY = win.left, win.top
	case al == 1:
		fontSel = 0
	case al == 2:
		fontSel = 0x80
	case al == 3:
		switch c := nextChar(); c {
		case 'q':
			flagQ = 1
		case 'k':
			flagK = 1
		case 'b':
			flagB = 1
		case 'd':
			flagD = 2
		case 't':
			win.mode = 0
		case 'f':
			win.mode = 0x80
		case 'm':
			dimView(uint16(nextChar() - '0'))
		case 'c':
			win.align = 2
		case 'r':
			win.align = 1
		default:
			win.align = 0
		}
	case al == 4:
		n, _ := digits(3)
		di := rowOff(win.curY<<1) + win.curX
		rows := 9
		if fontSel != 0 {
			rows = 8
		}
		if win.align == 1 {
			di -= n
		}
		for ; rows > 0; rows-- {
			fillBytes(di, n, win.fill)
			di += 0x140
		}
	case al == 5:
		var seg uint16
		for i := 0; i < 4; i++ {
			d := nextChar() - '0'
			if d > 9 {
				d -= 7
			}
			seg = seg<<4 | uint16(d)
		}
		drawList(Far(dsSeg<<16 | uint32(seg)).B()) // the string's segment: DGROUP
	case al == 6:
		drawGlyph(0x20)
	case al == 7:
		if n, ok := digits(3); ok {
			win.fill = n&0xff | n<<8
		} else {
			win.fill = win.defFill
		}
	case al == 8:
		c := nextChar()
		if c == ' ' {
			c = 0
			win.curX -= 2
			if small() {
				win.curX++
			}
		} else {
			if c == 6 {
				c = ' '
			}
			win.curX -= uint16(font[0x1000+int(c|fontSel)])
		}
		if win.left >= win.curX {
			win.curX = win.left
		}
		if c != 0 {
			x, cols := win.curX, glyphCol
			glyphCol[1], glyphCol[2], glyphCol[3] = byte(win.fill), byte(win.fill), byte(win.fill>>8)
			drawGlyph(c)
			win.curX, glyphCol = x, cols
		}
	case al == 9:
		n, _ := digits(3)
		if flagD == 0 {
			n = min(n+win.left, win.right)
		}
		win.curX = n
	case al == 0xa:
		if newline() && win.mode&0x80 != 0 {
			return true
		}
	case al == 0xb:
		n, _ := digits(3)
		if flagD == 0 {
			n = min(n+win.top, win.bottom)
		}
		win.curY = n
	case al == 0xc:
		n, ok := digits(2)
		if !ok {
			n = win.defColor
		}
		win.color = n
		setColors()
	default:
		return true
	}
	return false
}

// openTextWindow (0x3370) opens a text window (up to 7 deep) and prints text in it (0:
// none). fill, a7 and a8 are not used by the VGA driver: the fill colour is always 0xf0.
func openTextWindow(x, y, w, h, color, fill, a7, a8 int, text Far) {
	if winDepth >= 7 {
		return
	}
	winStack[winDepth] = win
	winDepth++
	win.x, win.y, win.w, win.h = uint16(x), uint16(y), uint16(w), uint16(h)
	win.color, win.fill = uint16(color), 0xf0f0
	win.defColor, win.defFill = win.color, win.fill
	cursorHide()
	save := make([]byte, 0, int(win.w)*int(win.h))
	bx := win.y << 1
	for bp := win.h; ; {
		o := rowOff(bx) + win.x
		for i := uint16(0); i < win.w; i++ {
			save = append(save, screen[o+i])
		}
		bx += 2
		if bp--; bp == 0 {
			break
		}
	}
	winSaves = append(winSaves, save)
	win.align, win.mode = 0, 0
	win.left = win.x + 8
	win.right = win.x + win.w - 8
	win.top = win.y + 8
	win.bottom = win.y + win.h - 8
	fillText()
	// border: 0 top-left, 1-4 top, 5 top-right, 6-9 left, 10-13 right, 14 bottom-left,
	// 15-18 bottom, 19 bottom-right
	win.curX, win.curY = win.x, win.y
	tile(0)
	for i, t := (win.w-9)>>3, 1; i > 0; i-- {
		tile(t)
		if t++; t == 5 {
			t = 1
		}
	}
	win.curX = win.x + win.w - 8
	tile(5)
	for i, t := (win.h-9)>>3, 6; i > 0; i-- {
		win.curY += 8
		win.curX = win.x
		tile(t)
		win.curX = win.x + win.w - 8
		tile(t + 4)
		if t++; t == 10 {
			t = 6
		}
	}
	win.curX, win.curY = win.x, win.y+win.h-8
	tile(14)
	for i, t := (win.w-9)>>3, 15; i > 0; i-- {
		tile(t)
		if t++; t == 19 {
			t = 15
		}
	}
	win.curX = win.x + win.w - 8
	tile(19)
	win.curX, win.curY = win.left, win.top
	setColors()
	if text != 0 {
		fromOpen = true
		printWindowText(text.B())
		return
	}
	flushText(true)
}

// tile (0x35d4) draws border tile t at the text cursor and moves it 8 right.
func tile(t int) {
	si := offTiles + t*64
	di := rowOff(win.curY<<1) + win.curX
	win.curX += 8
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if b := drv[si]; b != 0 {
				put(di, b)
			}
			si++
			di++
		}
		di += 0x138
	}
}

// closeWindows (0x361e) closes the n topmost windows, restoring what was under them.
func closeWindows(n int) {
	for ; n > 0; n-- {
		if winDepth--; winDepth < 0 {
			winDepth = 0
			return
		}
		cursorHide()
		save := winSaves[len(winSaves)-1]
		winSaves = winSaves[:len(winSaves)-1]
		bx, k := win.y<<1, 0
		for bp := win.h; ; {
			o := rowOff(bx) + win.x
			for i := uint16(0); i < win.w; i++ {
				if k < len(save) {
					screen[o+i] = save[k]
				}
				k++
			}
			bx += 2
			if bp--; bp == 0 {
				break
			}
		}
		flushText(true)
		win = winStack[winDepth]
	}
}

// vgaIdle1b (0x1c10) steps the spinning text cursor at the text position.
func vgaIdle1b() {
	cursorHide()
	eraseCell()
	textCursorOn = true
	if spinPhase--; spinPhase < 0 {
		spinPhase = 5
	}
	glyphAt(rowOff(win.curY<<1)+win.curX, drv[offSpinner+spinPhase])
	cursorShow()
	waitRetrace()
	copyRect(win.curX, win.curY, 8, 8)
	cursorFlush()
}

// dimView (0x1cda) darkens the 3D view (216x132 at 8,8) with colour 0: 1 every 4th pixel
// of every 2nd row... 2 a checkerboard, 3 three in four, 4 and up black.
func dimView(ax uint16) {
	ax--
	if int16(ax) < 0 {
		return
	}
	bp := int16(ax)
	di := uint16(0xa08)
	si := uint16(0x68)
	bx, dx := uint16(0x42), uint16(0x6c)
	stos := func(n uint16) {
		for ; n > 0; n-- {
			put(di, 0)
			di++
		}
	}
	if bp--; bp < 0 { // mode 1
		si += 0x140
		bx >>= 1
		dx >>= 1
		step := uint16(3)
		for ; bx > 0; bx-- {
			for cx := dx; cx > 0; cx-- {
				stos(1)
				di += step
			}
			di += si + step - 1
			for cx := dx; cx > 0; cx-- {
				stos(1)
				di += step
			}
			di += si - step + 1
		}
		return
	}
	if bp--; bp < 0 { // mode 2
		for ; bx > 0; bx-- {
			for cx := dx; cx > 0; cx-- {
				stos(1)
				di++
			}
			di += si
			for cx := dx; cx > 0; cx-- {
				di++
				stos(1)
			}
			di += si
		}
		return
	}
	if bp--; bp < 0 { // mode 3
		bx >>= 1
		dx >>= 1
		for ; bx > 0; bx-- {
			for cx := dx; cx > 0; cx-- {
				stos(2)
				di++
				stos(1)
			}
			di += si
			stos(dx * 4)
			di += si
			for cx := dx; cx > 0; cx-- {
				di++
				stos(3)
			}
			di += si
			stos(dx * 4)
			di += si
		}
		return
	}
	for bx <<= 1; bx > 0; bx-- { // mode 4+
		stos(dx * 2)
		di += si
	}
}
