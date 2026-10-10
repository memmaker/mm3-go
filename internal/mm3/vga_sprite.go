package mm3

// Sprite renderer of vga (MM3.CC): the RLE frame decoder at 0x1d70 and its eight variants
// (plain, mirrored, scaled down, scaled and mirrored, each with or without the 3D-view clip),
// plus the 2x/3x enlarger. Ported register for register: 16-bit offsets wrap as on the 8086,
// so a sprite partly off the left edge lands where the original put it.

// Scaler state (driver words 0xb60..0xb68) and the clip counters (0xb5a..0xb5e).
var (
	scRowMask, scReload  uint16 // 0xb60 row mask, 0xb62 the mask << 1 it reloads from
	scColMask, scColBase uint16 // 0xb64 column mask at the row start, 0xb66 screen x
	clipRem              uint16 // 0xb5e pixels left before the right clip edge
	clipLo, clipHi       uint16 // 0xb5a/0xb5c row offsets of x = 8 and x = 223
)

func rd8(b []byte, i uint16) byte {
	if int(i) < len(b) {
		return b[i]
	}
	return 0
}

func rd16(b []byte, i uint16) uint16 { return uint16(rd8(b, i)) | uint16(rd8(b, i+1))<<8 }

// rowOff is the row-offset table at 0x69c (y*320 for 0..199; beyond it the driver bytes
// that follow, as the original reads them). bx2 is 2*y.
func rowOff(bx2 uint16) uint16 { return rd16(drv, 0x69c+bx2) }

func put(di uint16, c byte) { screen[di] = c }

// shl16 is `shl reg,1`: the bit shifted out and whether the result is zero.
func shl16(r *uint16) (carry, zero bool) {
	carry = *r&0x8000 != 0
	*r <<= 1
	return carry, *r == 0
}

// scaleH (0x20d8) counts how many of n pixels the column mask dx keeps.
func scaleH(n uint16, dx *uint16) uint16 {
	if n == 0 {
		return 0
	}
	var ax uint16
	for ; n > 0; n-- {
		c, z := shl16(dx)
		if z {
			*dx = scReload
			ax++
		} else if c {
			ax = ax&0xff00 | uint16(byte(ax)+byte(ax>>8)+1) // adc al,ah
		} else {
			ax = ax&0xff00 | uint16(byte(ax)+byte(ax>>8))
		}
	}
	return ax
}

// scaleV (0x2136) is scaleH on the row mask.
func scaleV(n uint16) uint16 {
	if n == 0 {
		return 0
	}
	var ax uint16
	for ; n > 0; n-- {
		c, z := shl16(&scRowMask)
		if z {
			scRowMask = scReload
			ax++
		} else if c {
			ax = ax&0xff00 | uint16(byte(ax)+byte(ax>>8)+1)
		} else {
			ax = ax&0xff00 | uint16(byte(ax)+byte(ax>>8))
		}
	}
	return ax
}

// drawFrame (0x1d70) draws the frame whose cell pair is at g[si] (si = 4*frame+2) at (x, y)
// of the offscreen buffer. attr: bit 15 enlarge (bit 14: 3x, else 2x); else bit 0 mirror,
// bit 1 clip to the 3D view (x 8..223), bits 8-9 scale down (1: 3/4, 2: 1/2... the masks at
// 0xb6a), 0 = full size.
func drawFrame(g []byte, si, x, y, attr uint16) {
	if attr&0x8000 != 0 {
		f := uint16(2)
		if attr&0x4000 != 0 {
			f = 3
		}
		forCells(g, si, func(c uint16) { cellEnlarged(g, c, x, y, f) })
		return
	}
	a := attr & 0x303
	if a&0x300 != 0 {
		mask := rd16(drv, 0xb6a+(a>>8)*2)
		scReload = mask << 1
		cell := cellScaled
		switch {
		case a&2 != 0 && a&1 != 0:
			cell = cellScaledMirrorClip
		case a&2 != 0:
			cell = cellScaledClip
		case a&1 != 0:
			cell = cellScaledMirror
		}
		forCells(g, si, func(c uint16) { scRowMask = mask; cell(g, c, x, y, mask) })
		return
	}
	cell := cellPlain
	switch a {
	case 1:
		cell = cellMirror
	case 2:
		cell = cellClip
	case 3:
		cell = cellClipMirror
	}
	forCells(g, si, func(c uint16) { cell(g, c, x, y) })
}

// forCells draws a frame's first cell, then its second one if there is one.
func forCells(g []byte, si uint16, f func(uint16)) {
	f(rd16(g, si))
	if c := rd16(g, si+2); c != 0 {
		f(c)
	}
}

// cellHead reads a cell header {x, w, y, rows}.
func cellHead(g []byte, si uint16) (x, w, y, rows, next uint16) {
	return rd16(g, si), rd16(g, si+2), rd16(g, si+4), rd16(g, si+6), si + 8
}

// 0x1dae
func cellPlain(g []byte, si, cx, bx uint16) {
	x, _, y, rows, si := cellHead(g, si)
	dx := cx + x
	bx = (bx + y) << 1
	for {
		n := rd16(g, si)
		si += 2
		if n != 0 {
			if bx >= 0x190 {
				si += n
			} else {
				bp := si + n
				di := rowOff(bx) + dx + rd16(g, si)
				si += 2
				for si != bp {
					al := rd8(g, si)
					si++
					switch {
					case al&0x80 == 0:
						for n := uint16(al) + 1; n > 0; n-- {
							put(di, rd8(g, si))
							di++
							si++
						}
					case al&0x40 == 0:
						di += uint16(al&0x3f) + 1
					default:
						c := rd8(g, si)
						si++
						for n := uint16(al&0x3f) + 3; n > 0; n-- {
							put(di, c)
							di++
						}
					}
				}
			}
		}
		bx += 2
		if rows--; rows == 0 {
			return
		}
	}
}

// 0x1e36: drawn right to left from x+w-1.
func cellMirror(g []byte, si, cx, bx uint16) {
	x, w, y, rows, si := cellHead(g, si)
	dx := cx + x + w - 1
	bx = (bx + y) << 1
	for {
		n := rd16(g, si)
		si += 2
		if n != 0 {
			if bx >= 0x190 {
				si += n
			} else {
				bp := si + n
				di := rowOff(bx) + dx - rd16(g, si)
				si += 2
				for si != bp {
					al := rd8(g, si)
					si++
					switch {
					case al&0x80 == 0:
						for n := uint16(al) + 1; n > 0; n-- {
							put(di, rd8(g, si))
							di--
							si++
						}
					case al&0x40 == 0:
						di -= uint16(al&0x3f) + 1
					default:
						c := rd8(g, si)
						si++
						for n := uint16(al&0x3f) + 3; n > 0; n-- {
							put(di, c)
							di--
						}
					}
				}
			}
		}
		bx += 2
		if rows--; rows == 0 {
			return
		}
	}
}

func lt16(a, b uint16) bool { return int16(a) < int16(b) }

// 0x2180: clipped to x 8..223 of the row (the 3D view). Quirks kept: a run or skip that
// crosses the left edge restarts the 216-pixel budget at the crossing, not at x = 8.
func cellClip(g []byte, si, cx, bx uint16) {
	x, _, y, rows, si := cellHead(g, si)
	dx := cx + x
	bx = (bx + y) << 1
	for ; ; bx += 2 {
		n := rd16(g, si)
		si += 2
		if n != 0 {
			if bx >= 0x190 {
				si += n
			} else {
				clipRow(g, si, si+n, bx, dx)
				si += n
			}
		}
		if rows--; rows == 0 {
			return
		}
	}
}

func clipRow(g []byte, si, bp, bx, dx uint16) {
	di := rowOff(bx)
	cx := di + 8
	di += dx + rd16(g, si)
	si += 2
	if di < cx {
	skip: // 0x21d0
		for {
			if si == bp {
				return
			}
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				ax := uint16(al) + 1
				si += ax
				di += ax
				if lt16(di, cx) {
					continue
				}
				if di != cx {
					ax = di - cx
					si -= ax
					di = cx
					for ; ax > 0; ax-- {
						put(di, rd8(g, si))
						di++
						si++
					}
				}
				break skip
			case al&0x40 == 0:
				di += uint16(al&0x3f) + 1
				if lt16(di, cx) {
					continue
				}
				break skip
			default:
				ax := uint16(al&0x3f) + 3
				si++
				di += ax
				if lt16(di, cx) {
					continue
				}
				if di != cx {
					ax = di - cx
					di = cx
					c := rd8(g, si-1)
					for ; ax > 0; ax-- {
						put(di, c)
						di++
					}
				}
				break skip
			}
		}
		clipRem = 0xd8
	} else {
		clipRem = cx + 0xd8 - di
		if int16(clipRem) < 0 {
			return
		}
	}
	for si != bp { // 0x2236
		al := rd8(g, si)
		si++
		switch {
		case al&0x80 == 0:
			n := uint16(al) + 1
			clipRem -= n
			over := int16(clipRem) < 0
			if over {
				n += clipRem
			}
			for ; n > 0; n-- {
				put(di, rd8(g, si))
				di++
				si++
			}
			if over {
				return
			}
		case al&0x40 == 0:
			n := uint16(al&0x3f) + 1
			if clipRem -= n; int16(clipRem) < 0 {
				return
			}
			di += n
		default:
			n := uint16(al&0x3f) + 3
			c := rd8(g, si)
			si++
			if clipRem -= n; int16(clipRem) < 0 {
				n += clipRem
				for ; n > 0; n-- {
					put(di, c)
					di++
				}
				return
			}
			for ; n > 0; n-- {
				put(di, c)
				di++
			}
		}
	}
}

// 0x22ce: mirrored and clipped. Quirk kept: the left edge test is di - (row+8) & 0x7fff, so a
// row that starts left of the view is not clipped at all, and the leftmost pixel drawn is x = 9.
func cellClipMirror(g []byte, si, cx, bx uint16) {
	x, w, y, rows, si := cellHead(g, si)
	dx := cx + x + w - 1
	bx = (bx + y) << 1
	for ; ; bx += 2 {
		n := rd16(g, si)
		si += 2
		if n != 0 {
			if bx >= 0x190 {
				si += n
			} else {
				clipRowMirror(g, si, si+n, bx, dx)
				si += n
			}
		}
		if rows--; rows == 0 {
			return
		}
	}
}

func clipRowMirror(g []byte, si, bp, bx, dx uint16) {
	di := rowOff(bx)
	cx := di + 0xdf
	di += dx - rd16(g, si)
	si += 2
	if di > cx {
	skip: // 0x232c
		for {
			if si == bp {
				return
			}
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				ax := uint16(al) + 1
				si += ax
				di -= ax
				if int16(di) > int16(cx) {
					continue
				}
				if di != cx {
					ax = cx - di
					si -= ax
					di = cx
					for ; ax > 0; ax-- {
						put(di, rd8(g, si))
						di--
						si++
					}
				}
				break skip
			case al&0x40 == 0:
				di -= uint16(al&0x3f) + 1
				if int16(di) > int16(cx) {
					continue
				}
				break skip
			default:
				ax := uint16(al&0x3f) + 3
				si++
				di -= ax
				if int16(di) > int16(cx) {
					continue
				}
				if di != cx {
					ax = cx - di
					di = cx
					c := rd8(g, si-1)
					for ; ax > 0; ax-- {
						put(di, c)
						di--
					}
				}
				break skip
			}
		}
		clipRem = 0xd8
	} else {
		clipRem = (di - (cx - 0xd7)) & 0x7fff
	}
	for si != bp { // 0x23a4
		al := rd8(g, si)
		si++
		switch {
		case al&0x80 == 0:
			n := uint16(al) + 1
			clipRem -= n
			over := int16(clipRem) < 0
			if over {
				n += clipRem
			}
			for ; n > 0; n-- {
				put(di, rd8(g, si))
				di--
				si++
			}
			if over {
				return
			}
		case al&0x40 == 0:
			n := uint16(al&0x3f) + 1
			if clipRem -= n; int16(clipRem) < 0 {
				return
			}
			di -= n
		default:
			n := uint16(al&0x3f) + 3
			c := rd8(g, si)
			si++
			clipRem -= n
			over := int16(clipRem) < 0
			if over {
				n += clipRem
			}
			for ; n > 0; n-- {
				put(di, c)
				di--
			}
			if over {
				return
			}
		}
	}
}

// scaledRows runs the row loop of the four scaled variants (0x1f3e etc.): the row mask drops
// rows (without advancing y); row draws one kept row.
func scaledRows(g []byte, si, bx, rows uint16, row func(si, bp, bx uint16)) {
	for {
		c, z := shl16(&scRowMask)
		n := rd16(g, si)
		si += 2
		if !c {
			si += n
		} else {
			if z {
				scRowMask = scReload
			}
			if n != 0 {
				if bx >= 0x190 {
					si += n
				} else {
					row(si, si+n, bx)
					si += n
				}
			}
			bx += 2
		}
		if rows--; rows == 0 {
			return
		}
	}
}

// scaledHead positions a scaled cell: x scaled by the column mask (which then carries on
// into every row), y by the row mask.
func scaledHead(g []byte, si, cx, bx, dx uint16, mirror bool) (uint16, uint16, uint16) {
	x, w, y, rows, si := cellHead(g, si)
	if mirror {
		x += w
	}
	scColBase = cx + scaleH(x, &dx)
	scColMask = dx
	bx = (bx + scaleV(y)) << 1
	return si, bx, rows
}

// 0x1f1c
func cellScaled(g []byte, si, cx, bx, dx uint16) {
	si, bx, rows := scaledHead(g, si, cx, bx, dx, false)
	scaledRows(g, si, bx, rows, func(si, bp, bx uint16) {
		di := rowOff(bx) + scColBase
		dx := scColMask
		di += scaleH(rd16(g, si), &dx)
		si += 2
		for si != bp {
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				for n := uint16(al) + 1; n > 0; n-- {
					b := rd8(g, si)
					si++
					if c, z := shl16(&dx); z {
						dx = scReload
						put(di, b)
						di++
					} else if c {
						put(di, b)
						di++
					}
				}
			case al&0x40 == 0:
				di += scaleH(uint16(al&0x3f)+1, &dx)
			default:
				n := scaleH(uint16(al&0x3f)+3, &dx)
				b := rd8(g, si)
				si++
				for ; n > 0; n-- {
					put(di, b)
					di++
				}
			}
		}
	})
}

// 0x2012
func cellScaledMirror(g []byte, si, cx, bx, dx uint16) {
	si, bx, rows := scaledHead(g, si, cx, bx, dx, true)
	scaledRows(g, si, bx, rows, func(si, bp, bx uint16) {
		di := rowOff(bx) + scColBase
		dx := scColMask
		di -= scaleH(rd16(g, si), &dx)
		si += 2
		for si != bp {
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				for n := uint16(al) + 1; n > 0; n-- {
					b := rd8(g, si)
					si++
					if c, z := shl16(&dx); z {
						dx = scReload
						put(di, b)
						di--
					} else if c {
						put(di, b)
						di--
					}
				}
			case al&0x40 == 0:
				di -= scaleH(uint16(al&0x3f)+1, &dx)
			default:
				n := scaleH(uint16(al&0x3f)+3, &dx)
				b := rd8(g, si)
				si++
				for ; n > 0; n-- {
					put(di, b)
					di--
				}
			}
		}
	})
}

// clipPut stores a pixel of a scaled clipped sprite only inside x 8..223.
func clipPut(di uint16, b byte) {
	if di >= clipLo && di <= clipHi {
		put(di, b)
	}
}

// 0x24aa
func cellScaledClip(g []byte, si, cx, bx, dx uint16) {
	si, bx, rows := scaledHead(g, si, cx, bx, dx, false)
	scaledRows(g, si, bx, rows, func(si, bp, bx uint16) {
		di := rowOff(bx)
		clipLo = di + 8
		clipHi = clipLo + 0xd7
		di += scColBase
		dx := scColMask
		di += scaleH(rd16(g, si), &dx)
		si += 2
		for si != bp {
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				for n := uint16(al) + 1; n > 0; n-- {
					b := rd8(g, si)
					si++
					if c, z := shl16(&dx); z {
						dx = scReload
						clipPut(di, b)
						di++
					} else if c {
						clipPut(di, b)
						di++
					}
				}
			case al&0x40 == 0:
				di += scaleH(uint16(al&0x3f)+1, &dx)
			default:
				n := scaleH(uint16(al&0x3f)+3, &dx)
				b := rd8(g, si)
				si++
				for ; n > 0; n-- {
					clipPut(di, b)
					di++
				}
			}
		}
	})
}

// 0x25ec
func cellScaledMirrorClip(g []byte, si, cx, bx, dx uint16) {
	si, bx, rows := scaledHead(g, si, cx, bx, dx, true)
	scaledRows(g, si, bx, rows, func(si, bp, bx uint16) {
		di := rowOff(bx)
		clipLo = di + 8
		clipHi = clipLo + 0xd7
		di += scColBase
		dx := scColMask
		di -= scaleH(rd16(g, si), &dx)
		si += 2
		for si != bp {
			al := rd8(g, si)
			si++
			switch {
			case al&0x80 == 0:
				for n := uint16(al) + 1; n > 0; n-- {
					b := rd8(g, si)
					si++
					if c, z := shl16(&dx); z {
						dx = scReload
						clipPut(di, b)
						di--
					} else if c {
						clipPut(di, b)
						di--
					}
				}
			case al&0x40 == 0:
				di -= scaleH(uint16(al&0x3f)+1, &dx)
			default:
				n := scaleH(uint16(al&0x3f)+3, &dx)
				b := rd8(g, si)
				si++
				for ; n > 0; n-- {
					clipPut(di, b)
					di--
				}
			}
		}
	})
}

// 0x2722: every pixel f wide, every row f high, offsets times f. Quirk kept: a literal run's
// count is (high byte of the screen x) << 8 | count byte, so literals at x >= 256 run long.
func cellEnlarged(g []byte, si, cx, bx, f uint16) {
	x, _, y, rows, si := cellHead(g, si)
	dx := cx + x*f
	bx = (bx + y*f) * 2
	for {
		start := si // 0xe90
		for rep := f; ; {
			n := rd16(g, si)
			si += 2
			if n != 0 {
				if bx >= 0x190 {
					si += n
				} else {
					bp := si + n
					di := rowOff(bx) + dx + rd16(g, si)*f
					si += 2
					for si != bp {
						al := rd8(g, si)
						si++
						switch {
						case al&0x80 == 0:
							n := dx&0xff00 | uint16(al)
							n++
							px := func() {
								b := rd8(g, si)
								si++
								for k := f; k > 0; k-- {
									put(di, b)
									di++
								}
							}
							q := n >> 1
							if n&1 != 0 {
								px()
							}
							if n&1 == 0 || q != 0 {
								for {
									px()
									px()
									if q--; q == 0 {
										break
									}
								}
							}
						case al&0x40 == 0:
							di += (uint16(al&0x3f) + 1) * f
						default:
							k := (uint16(al&0x3f) + 3) * f
							b := rd8(g, si)
							si++
							for ; k > 0; k-- {
								put(di, b)
								di++
							}
						}
					}
				}
			}
			bx += 2
			if rep--; rep == 0 {
				break
			}
			si = start
		}
		if rows--; rows == 0 {
			return
		}
	}
}
