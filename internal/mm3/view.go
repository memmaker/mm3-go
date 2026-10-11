// Port of VIEW.C: the clock, conditions and stats, the wizard eye, what is visible in the 3D
// view and the draw-list entries for it, the party portraits.
//
// A C `int *o` into g_draw_list is an int index into gDrawList here (addViewSpritesD3 etc.
// take and return it). Borland's "%p" of a near pointer (g_fmt_sprites, g_fmt_portraits) is
// four upper-case hex digits; libc.go's sprintf has no %p, so fmtP rewrites it to %04X.
package mm3

import "strings"

// dlPut is the C's `*++o = v` on g_draw_list: the next int, as a 16-bit store.
func dlPut(o, v int) int { o++; gDrawList[o] = int16(v); return o }

// dlPutL is `*(long *)++o = p`: a far pointer over the next two ints (the C then does o++).
func dlPutL(o int, p Far) int { o++; setLe32(ds[:], adrDrawList+2*o, uint32(p)); return o }

// fmtP is a ds format string with Borland's near %p as %04X.
func fmtP(f []byte) string { return strings.Replace(gostr(f), "%p", "%04X", 1) }

// hov is `(g_mon_hovers[g_mon_type[k]] ? d[phase] : 0)`: a hovering monster's bob offset.
func hov(d []byte, phase uint8, k int) int {
	if gMonHovers[gMonType[k]] != 0 {
		return int(int8(d[phase]))
	}
	return 0
}

// objFrame is g_map_objects[k].frame[0x13b] (frame is a near u8 pointer); monFrame is
// g_mon_frames[k][0x124] (g_mon_frames is u8 *[]).
func objFrame(k int) int { return int(ds[gMapObjects[k].Frame.Get()+0x13b]) }
func monFrame(k int) int { return int(ds[int(uint16(gMonFrames[k]))+0x124]) }

func recoverMonsterStatus() {
	for i := 0; i < int(*gMonCount); i++ {
		if gMonStatus[i] != 0 && monsterSavingThrow(int(gMonType[i])) != 0 {
			gMonStatus[i] = 0
		}
	}
}

func updatePartyDisabled() {
	flag := *gGameMode == 2
	n := int(*gPartyCount)
	if flag {
		n = int(*gCombatPartyN)
	}
	for i := 0; i < n; i++ {
		c := i
		if flag {
			c = int(gCombatOrder[i])
		}
		switch getWorstCondition(&gParty[c]) {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 16:
			*gPartyDefeated = 0
			return
		}
	}
	*gPartyDefeated = 1
}

func testBit(bits []byte, i int) int {
	i = int(uint16(i))
	return int(int8(bits[i>>3])) & (0x80 >> (i & 7))
}

func advanceClock(ticks int) {
	prev := int(*gDay)
	*gMinutes += uint16(ticks)
	for *gMinutes >= 1440 {
		*gMinutes -= 1440
		if *gDay++; *gDay >= 100 {
			*gDay -= 100
			*gYear++
			for i := 0x6f; i < 0x76; i++ {
				setBit(gGameFlags, i, 0)
			}
		}
	}
	if (*gDay%10 == 1 || ticks > 1440) && int(*gDay) != prev {
		restockShops()
		addBankInterest()
	}
	if int(*gDay) != prev {
		*gNewDayPending = 1
	}
	if *gNewDayPending != 0 && *gMinutes >= 300 {
		if *gGameMode != 9 {
			chargeHirelings()
			clearTempEffects(1)
			if *gOpenWindows == 0 {
				drawPartyPortraits()
			}
		}
		*gNewDayPending = 0
	}
}

func partyHasSkill(cond int) int {
	n := 0
	for i := 0; i < int(*gPartyCount); i++ {
		if gPartySkills[i*303+cond] != 0 {
			n++
			switch cond {
			case 0, 1, 2, 3, 4, 6, 7, 8, 12, 13, 15, 16, 17:
				return 1
			case 9, 10, 11:
				if n == 2 {
					return 1
				}
			case 5, 14:
				if int(*gPartyCount) == n {
					return 1
				}
			}
		}
	}
	return 0
}

func setBit(p []byte, bit int, set int) {
	byt := int(uint16(bit)) >> 3
	b := bit & 7
	if set != 0 {
		p[byt] |= byte(0x80 >> b)
	} else {
		p[byt] &= 0xff ^ byte(0x80>>b)
	}
}

var wizEyeBuf [20]byte // draw_wizard_eye's buf: the window gets its far address

// wizEyeLvl is the quadrant (of g_maps) that cell (x, y) relative to the party is in.
func wizEyeLvl(x, y int) int {
	lvl := int(*gMapQuadrant)
	if int(*gPartyX)+x > 15 {
		lvl++
	} else if int(*gPartyX)+x < 0 {
		lvl--
	}
	if int(*gPartyY)+y > 15 {
		lvl += 2
	} else if int(*gPartyY)+y < 0 {
		lvl -= 2
	}
	return lvl
}

func wizEyeShow(buf []byte) {
	if *gWizeyeShown != 0 {
		printWindowText(buf)
	} else {
		openTextWindow(0xe7, 1, 0x56, 0x48, 0, 0x96, 4, 6, farOf(buf))
		*gWizeyeShown = 1
	}
}

func drawWizardEyeOutdoor() {
	var k, lvl, v, x, y int
	buf := wizEyeBuf[:]
	px, py := int(*gPartyX), int(*gPartyY)
	dl := gDrawList

	k = 3
	y = 4
	for i := 0; i < 7; i++ {
		y--
		x = -4
		for j := 0; j < 7; j++ {
			x++
			if !minimapKnown(px+x, py+y) {
				continue
			}
			v = getCellWalls(px+x, py+y, 0x70) >> 4
			lvl = wizEyeLvl(x, y)
			dl[k] = gWizeyeColX[j]
			k++
			dl[k] = gWizeyeRowY[i]
			k++
			k++
			if v != 0 {
				dl[k] = int16(ds[adrMapSurfacesM1+lvl*0x340+v])
			} else {
				dl[k] = 0
			}
			k++
			v = getCellWalls(px+x, py+y, 7)
			if v != 0 {
				lvl = wizEyeLvl(x, y)
				dl[k] = gWizeyeColX[j]
				k++
				dl[k] = gWizeyeRowY[i]
				k++
				k++
				dl[k] = int16(ds[adrMapSurfacesM1+lvl*0x340+v])
				k++
			}
			v = getCellWalls(px+x, py+y, 0x700) >> 8
			if v != 0 {
				dl[k] = gWizeyeColX[j]
				k++
				dl[k] = gWizeyeRowY[i]
				k++
				k++
				dl[k] = int16(v + 0x16)
				k++
			}
		}
	}
	if partyHasSkill(6) != 0 { // port addition: the arrow needs Direction Sense
		dl[k] = 0x1e
		k++
		dl[k] = 0x18
		k++
		k++
		dl[k] = int16(*gPartyDir) + 0x13
		k++
	}
	dl[k] = -1
	sprintf(buf, fmtP(gFmtSprites), adrDrawList)
	wizEyeShow(buf)
	// the window keeps its last picture: black out the unexplored cells skipped above
	cursorHide()
	for i := 0; i < 7; i++ {
		for j := 0; j < 7; j++ {
			if minimapKnown(px+j-3, py+3-i) {
				continue
			}
			x, y := win.left+uint16(gWizeyeColX[j]), win.top+uint16(gWizeyeRowY[i])
			for d := uint16(0); d < 80; d++ {
				put(rowOff((y+d/10)<<1)+x+d%10, 0)
			}
		}
	}
	flushText(true)
	wizEyeMonsters()
}

// wizEyeMonsters (port addition): marks the party with a red X when there is no arrow; with the Wizard Eye spell up, a red dot marks each cell of
// the minimap's 7x7 that holds live monsters. The minimap itself only needs a Cartographer.
func wizEyeMonsters() {
	red := nearestColor(63, 0, 0)
	cursorHide() // the cursor is drawn into the offscreen: lift it off before drawing under it
	if partyHasSkill(6) == 0 { // no Direction Sense, no arrow: a red X marks the party
		x, y := win.left+uint16(gWizeyeColX[3])+1, win.top+uint16(gWizeyeRowY[3])
		for d := uint16(0); d < 8; d++ {
			put(rowOff((y+d)<<1)+x+d, red)
			put(rowOff((y+d)<<1)+x+7-d, red)
		}
	}
	if *gWizardEye == 0 {
		flushText(true)
		return
	}
	px, py := int(*gPartyX), int(*gPartyY)
	for m := 0; m < int(*gMonCount); m++ {
		dx, dy := int(gMonX[m])-px, int(gMonY[m])-py
		if uint16(gMonY[m]) >= 0x20 || gMonVisible[m] == 0 || dx < -3 || dx > 3 || dy < -3 || dy > 3 {
			continue
		}
		x, y := win.left+uint16(gWizeyeColX[dx+3])+3, win.top+uint16(gWizeyeRowY[3-dy])+2
		for d := uint16(0); d < 16; d++ {
			put(rowOff((y+d/4)<<1)+x+d%4, red)
		}
	}
	flushText(true)
}

// nearestColor is the game palette index closest to the 6-bit (r, g, b).
func nearestColor(r, g, b int) byte {
	best, bd := 0, 1<<30
	for i := 0; i < 256; i++ {
		p := drv[offPalette+i*3:]
		dr, dg, db := int(p[0])-r, int(p[1])-g, int(p[2])-b
		if d := dr*dr + dg*dg + db*db; d < bd {
			best, bd = i, d
		}
	}
	return byte(best)
}

func drawWizardEye(force int) {
	var k, v, x, y int
	buf := wizEyeBuf[:]
	dl := gDrawList

	if partyHasSkill(4) != 0 {
		markCellVisited(int(*gPartyX), int(*gPartyY))
	}
	if *gLightChanged != 0 {
		*gViewDirty = 1
		updateDarkness()
		if *gDimLevel == 0 {
			*gViewDirty = 0
			*gDimLevel = 4
			for *gDimLevel > 0 {
				paceFrame()
				refreshView()
				*gDimLevel--
			}
		}
		*gLightChanged = 0
		refreshView()
	}
	// port addition: a Cartographer in the party keeps the minimap up without Wizard Eye
	eye := *gWizardEye != 0 || partyHasSkill(4) != 0
	if !eye && *gWizeyeShown != 0 {
		closeWindows(1)
		*gWizeyeShown = 0
	}
	if !eye {
		return
	}
	if *gViewDirty == 0 && force == 0 {
		return
	}
	memset(ds[adrDrawList:], 0, 4000)
	dl[0] = -1
	*gDrawListImg = Far(*gTilesetGfx)
	if *gOutdoors != 0 {
		drawWizardEyeOutdoor()
		return
	}
	px, py := int(*gPartyX), int(*gPartyY)
	add := func(i, j, f int) {
		dl[k] = gWizeyeColX[j]
		k++
		dl[k] = gWizeyeRowY[i]
		k++
		k++
		dl[k] = int16(f)
		k++
	}
	k = 3
	y = 4
	for i := 0; i < 7; i++ {
		y--
		x = -4
		for j := 0; j < 7; j++ {
			x++
			if minimapKnown(px+x, py+y) && getCellWalls(px+x, py+y, 0x7777) != 0x1111 {
				add(i, j, 0)
			}
		}
	}
	y = 4
	for i := 0; i < 7; i++ {
		y--
		x = -4
		for j := 0; j < 7; j++ {
			x++
			if !minimapKnown(px+x, py+y) || getCellWalls(px+x, py+y, 0x7777) == 0x1111 {
				continue
			}
			v = -1
			switch getCellWalls(px+x, py+y, 0x7000) >> 12 {
			case 1, 3:
				v = 2
			case 4:
				v = 12
			case 5:
				v = 10
			case 2:
				v = 6
			case 6:
				v = 8
			case 7:
				v = 4
			}
			if v != -1 {
				if j != 6 {
					add(i, j, v)
				} else {
					add(i, j, (v>>1)+0x11)
				}
			}
			v = -1
			switch getCellWalls(px+x, py+y, 7) {
			case 1, 3:
				v = 3
			case 4:
				v = 13
			case 5:
				v = 11
			case 2:
				v = 7
			case 6:
				v = 9
			case 7:
				v = 5
			}
			if v != -1 {
				if i != 6 {
					add(i, j, v)
				} else {
					add(i, j, (v>>1)+0x17)
				}
			}
		}
	}
	y = 4
	for i := 0; i < 7; i++ {
		y--
		x = -4
		for j := 0; j < 7; j++ {
			x++
			if !minimapKnown(px+x, py+y) || getCellWalls(px+x, py+y, 0x7777) == 0x1111 {
				add(i, j, 1)
			}
		}
	}
	if partyHasSkill(6) != 0 { // port addition: the arrow needs Direction Sense
		dl[k] = 0x20
		k++
		dl[k] = 0x1a
		k++
		k++
		dl[k] = int16(*gPartyDir) + 14
		k++
	}
	dl[k] = -1
	sprintf(buf, fmtP(gFmtSprites), adrDrawList)
	wizEyeShow(buf)
	wizEyeMonsters()
}

// minimapKnown (port addition): without Wizard Eye the minimap shows only explored cells.
func minimapKnown(x, y int) bool { return *gWizardEye != 0 || isCellVisited(x, y) != 0 }

// viewAt: is (x, y) the cell at view position n for facing fac from (px, py)?
func viewAt(fac, n, px, py, x, y int) bool {
	return py+int(int8(gViewDyFrom2[fac][n])) == y && px+int(int8(gViewDxFrom2[fac][n])) == x
}

// The wall visibility tests shared by scan_view_objects and scan_view_monsters, by view
// position: (a == 0 || b == 0) chains on the wall flags.
func z(p *uint8) bool      { return *p == 0 }
func nz2(a, b *uint8) bool { return *a == 0 || *b == 0 }
func viewOK3() bool {
	return nz2(gWallD0c, gWallD0l1) && nz2(gWallD0c, gWallD0cLside) && *gWallD1cLside&*gWallD0l1 == 0 && nz2(gWallD1cLside, gWallD0cLside)
}
func viewOK7() bool {
	return nz2(gWallD0c, gWallD0r1) && nz2(gWallD0c, gWallD0cRside) && *gWallD1cRside&*gWallD0r1 == 0 && nz2(gWallD1cRside, gWallD0cRside)
}
func viewOK10() bool {
	return z(gWallD0c) && nz2(gWallD1c, gWallD1cLside) && nz2(gWallD1c, gWallD1l1) && nz2(gWallD1cLside, gWallD2cLside) && nz2(gWallD1l1, gWallD2cLside)
}
func viewOK14() bool {
	return z(gWallD0c) && nz2(gWallD1c, gWallD1cRside) && nz2(gWallD1c, gWallD1r1) && nz2(gWallD1cRside, gWallD2cRside) && nz2(gWallD1r1, gWallD2cRside)
}
func viewOK25() bool { return z(gWallD0c) && z(gWallD1c) && z(gWallD2c) }
func viewOK23() bool {
	return z(gWallD0c) && z(gWallD1c) && nz2(gWallD2c, gWallD2cLside) && nz2(gWallD2c, gWallD2l1) && nz2(gWallD2l1, gWallD3cLside) && nz2(gWallD2cLside, gWallD3cLside)
}
func viewOK21() bool {
	return z(gWallD0c) && nz2(gWallD1c, gWallD1l1) && nz2(gWallD1c, gWallD1cLside) && nz2(gWallD1l1, gWallD2cLside) && nz2(gWallD1cLside, gWallD2cLside) && z(gWallD2l1) && z(gWallD3l1Lside)
}
func viewOK27() bool {
	return z(gWallD0c) && z(gWallD1c) && nz2(gWallD2c, gWallD2cRside) && nz2(gWallD2c, gWallD2r1) && nz2(gWallD2r1, gWallD3cRside) && nz2(gWallD2cRside, gWallD3cRside)
}
func viewOK29() bool {
	return z(gWallD0c) && nz2(gWallD1c, gWallD1r1) && nz2(gWallD1c, gWallD1cRside) && nz2(gWallD1r1, gWallD2cRside) && nz2(gWallD1cRside, gWallD2cRside) && z(gWallD2r1) && z(gWallD3r1Rside)
}
func setFirst(i int, ps ...*uint8) {
	for _, p := range ps {
		if *p == 0 {
			*p = uint8(i + 1)
			return
		}
	}
}

func scanViewObjects() {
	px, py, fac := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
	for _, p := range []*uint8{gObjD0c, gObjD1c, gObjD1l1, gObjD1r1, gObjD2c, gObjD2l1, gObjD2r1, gObjD3c, gObjD3l1, gObjD3r1, gObjD3l2, gObjD3r2} {
		*p = 0
	}
	for i := 0; i < int(*gObjCount); i++ {
		oy, ox := int(gMapObjects[i].Y.Get()), int(gMapObjects[i].X.Get())
		at := func(n int) bool { return viewAt(fac, n, px, py, ox, oy) }
		if at(0) && z(gObjD0c) {
			*gObjD0c = uint8(i + 1)
			if *gViewDirty != 0 {
				if id := gMapObjects[i].ID.Get(); id == 0x15 || id == 0x38 {
					playSfx(0x2e)
					*gObjSoundOn = 1
				} else if *gObjSoundOn != 0 {
					playSfx(0)
					*gObjSoundOn = 0
				}
			}
		}
		if at(5) && z(gWallD0c) && z(gObjD1c) {
			*gObjD1c = uint8(i + 1)
		}
		if at(3) && viewOK3() && z(gObjD1l1) {
			*gObjD1l1 = uint8(i + 1)
		}
		if at(7) && viewOK7() && z(gObjD1r1) {
			*gObjD1r1 = uint8(i + 1)
		}
		if at(12) && z(gWallD1c) && z(gWallD0c) && z(gObjD2c) {
			*gObjD2c = uint8(i + 1)
		}
		if at(10) && viewOK10() && z(gObjD2l1) {
			*gObjD2l1 = uint8(i + 1)
		}
		if at(14) && viewOK14() && z(gObjD2r1) {
			*gObjD2r1 = uint8(i + 1)
		}
		if at(25) && viewOK25() && z(gObjD3c) {
			*gObjD3c = uint8(i + 1)
		}
		if at(23) && viewOK23() && z(gObjD3l1) {
			*gObjD3l1 = uint8(i + 1)
		}
		if at(21) && viewOK21() && z(gObjD3l2) {
			*gObjD3l2 = uint8(i + 1)
		}
		if at(27) && viewOK27() && z(gObjD3r1) {
			*gObjD3r1 = uint8(i + 1)
		}
		if at(29) && viewOK29() && z(gObjD3r2) {
			*gObjD3r2 = uint8(i + 1)
		}
	}
}

func scanViewMonsters() {
	px, py, fac := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
	for _, p := range []*uint8{gMonD0c1, gMonD1c1, gMonD1l1, gMonD1r1, gMonD2c1, gMonD2l11, gMonD2r11, gMonD3c1, gMonD3l11, gMonD3r11, gMonD3l2, gMonD3r2,
		gMonD0c2, gMonD1c2, gMonD2c2, gMonD2l12, gMonD2r12, gMonD3c2, gMonD3l12, gMonD3r12, gMonD0c3, gMonD1c3, gMonD2c3, gMonD3c3, gMonD3l13, gMonD3r13} {
		*p = 0
	}
	for i := 0; i < int(*gMonCount); i++ {
		oy, ox := int(gMonY[i]), int(gMonX[i])
		fl := &gMonVisible[i]
		at := func(n int) bool { return viewAt(fac, n, px, py, ox, oy) }
		if at(0) {
			*fl = 1
			setFirst(i, gMonD0c1, gMonD0c2, gMonD0c3)
		}
		if at(5) {
			*fl = 1
			if z(gWallD0c) {
				setFirst(i, gMonD1c1, gMonD1c2, gMonD1c3)
			}
		}
		if at(3) && viewOK3() {
			*fl = 1
			setFirst(i, gMonD1l1)
		}
		if at(7) && viewOK7() {
			*fl = 1
			setFirst(i, gMonD1r1)
		}
		if at(12) {
			*fl = 1
			if z(gWallD1c) && z(gWallD0c) {
				setFirst(i, gMonD2c1, gMonD2c2, gMonD2c3)
			}
		}
		if at(10) && viewOK10() {
			*fl = 1
			setFirst(i, gMonD2l11, gMonD2l12)
		}
		if at(14) && viewOK14() {
			*fl = 1
			setFirst(i, gMonD2r11, gMonD2r12)
		}
		if at(25) && viewOK25() {
			*fl = 1
			setFirst(i, gMonD3c1, gMonD3c2, gMonD3c3)
		}
		if at(23) && viewOK23() {
			*fl = 1
			setFirst(i, gMonD3l11, gMonD3l12, gMonD3l13)
		}
		if at(21) && viewOK21() {
			*fl = 1
			setFirst(i, gMonD3l2)
		}
		if at(27) && viewOK27() {
			*fl = 1
			setFirst(i, gMonD3r11, gMonD3r12, gMonD3r13)
		}
		if at(29) && viewOK29() {
			*fl = 1
			setFirst(i, gMonD3r2)
		}
	}
}

func mapHasDaylight() int {
	m := *gMapNum
	if m < 6 || (m > 0x17 && m < 0x1d) || (m > 0x28 && m != 0x6a && m != 0x69) {
		return 1
	}
	return 0
}

func passTime(ticks int) {
	if (int(*gMinutes)+ticks)/480 != int(*gMinutes)/480 {
		for i := 0; i < int(*gPartyCount); i++ {
			c := &gParty[i]
			cd := &c.Cond
			if cd[13]|cd[14]|cd[15] == 0 {
				for j := 0; j < 7; j++ {
					if getCharStat(c, j, 0) < 1 {
						cd[13] = 1
					}
				}
			}
			if cd[1] != 0 {
				cd[1]++
				if cd[1] > 10 {
					cd[1] = 0
					cd[9] = 1
				}
			}
			cd[2], cd[11], cd[8], cd[10] = 0, 0, 0, 0
			for _, n := range []int{3, 4} {
				if cd[n] != 0 {
					if randomRange(1, 10) == 1 {
						if charSavingThrow(c, n) != 0 {
							cd[n] = 0
						} else {
							cd[n] <<= 1
						}
					} else {
						cd[n] <<= 1
					}
				}
			}
			if cd[5] != 0 {
				cd[5]++
			}
			for _, n := range []int{13, 14, 15} {
				if cd[n] != 0 {
					cd[n]++
					if cd[n] == 0 {
						cd[n] = 0xff
					}
				}
			}
			if cd[6] != 0 {
				if cd[6]++; cd[6] > 10 {
					cd[6] = 0
					cd[1] = 1
				}
			}
			cd[2] = cd[7]
			cd[7] = 0
			if cd[9] != 0 {
				cd[9]++
				cd[9] %= 4
			}
		}
	}
	advanceClock(ticks)
	for i := 0; i < int(*gPartyCount); i++ {
		c := &gParty[i]
		if c.Cond[10] != 0 {
			if randomRange(1, 3) == 1 {
				if charSavingThrow(c, 0) != 0 {
					c.Cond[10] = 0
				} else {
					c.Cond[10]--
				}
			}
		}
		if c.Cond[11] != 0 {
			if randomRange(1, 5) == 1 {
				c.Cond[11]--
			}
		}
	}
	if mapHasDaylight() != 0 {
		night := uint8(1)
		if *gMinutes >= 300 && *gMinutes < 0x4ec {
			night = 0
		}
		if night != *gIsNight {
			loadSky(int(*gMapNum))
		}
	}
}

func endPartyTurn() {
	if *gPartyDefeated == 0 {
		t := 1
		if *gOutdoors != 0 {
			t = 10
		}
		passTime(t)
		if *gMonsterScanDelay != 0 {
			moveMonsters()
		}
		*gMonsterScanDelay = 3
	}
}

func statBonus(v int) int {
	i := 0
	for ; int(gStatRatingLimits[i]) <= int(uint16(v)); i++ {
	}
	return int(int8(gStatBonusValues[i]))
}

func getCharAge(m *Chr, flag int) int {
	d := int(int16(*gYear - uint16(m.BirthYear.Get())))
	if d > 0xfe {
		d = 0xfe
	}
	if flag != 0 {
		return d
	}
	return d + int(m.AgeMod)
}

func getWorstCondition(m *Chr) uint8 {
	for i := 15; i >= 0; i-- {
		if m.Cond[i] != 0 {
			return uint8(i)
		}
	}
	return 16
}

func getItemStatBonus(m *Chr, kind int) int {
	sum := 0
	for i := 0; i < 18; i++ {
		if m.ItemEq[i] != 0 && m.ItemFlags[i]&0xc0 == 0 {
			if kind < 11 && kind != 3 && m.ItemAttr[i] != 0 {
				v := getBonusStatIndex(int(m.ItemAttr[i]))
				if v > 2 {
					v++
				}
				if v == kind {
					sum += int(int8(gAttribBonusValues[m.ItemAttr[i]]))
				}
			}
			if kind > 10 && m.ItemElem[i] != 0 {
				v := getElementIndex(int(m.ItemElem[i])) + 11
				if v == kind {
					sum += int(gElemResistBonus[m.ItemElem[i]])
				}
			}
			if kind == 9 {
				sum += int(gItemAcValues[m.ItemID[i]])
				if m.ItemMat[i] != 0 {
					if m.ItemID[i] != 0x2a {
						switch m.ItemEq[i] {
						case 1, 4, 13:
							continue
						}
					}
					sum += int(int8(gMaterialAcBonus[m.ItemMat[i]]))
				}
			}
		}
	}
	return int(int16(sum))
}

func getCharLevel(m *Chr) int {
	v := int(m.Level) + int(m.LevelTmp)
	if v < 0 {
		return 0
	}
	return v
}

// getCharStat is get_char_stat (unsigned in the C; never negative).
func getCharStat(m *Chr, kind, flag int) int {
	var v, w, t int
	switch kind {
	case 0:
		v, w = int(m.Might), int(m.MightTmp)
	case 1:
		v, w, t = int(m.Intel), int(m.IntelTmp), 1
	case 2:
		v, w, t = int(m.Pers), int(m.PersTmp), 1
	case 3:
		v, w = int(m.Endur), int(m.EndurTmp)
	case 4:
		v, w = int(m.Speed), int(m.SpeedTmp)
	case 5:
		v, w = int(m.Acc), int(m.AccTmp)
	case 6:
		v, w, t = int(m.Luck), int(m.LuckTmp), 2
	}
	if t < 2 {
		a := getCharAge(m, 0)
		j := 0
		for ; gAgeThresholds[j] <= uint16(a); j++ {
		}
		v += int(gAgeStatMods[t][j])
	}
	v += getItemStatBonus(m, kind)
	if flag == 0 {
		v += getConditionStatMod(m, kind)
		v += w
	}
	v = int(int16(v))
	if v < 1 {
		return 0
	}
	return v
}

func drawPartyPortraits() {
	var buf [10]byte
	var pos uint32
	flag := *gGameMode == 2
	n := int(*gPartyCount)
	if flag {
		n = int(*gCombatPartyN)
	}
	sprintf(buf[:], fmtP(gFmtPortraits), adrDrawList)
	o := 0
	put := func(v int) { gDrawList[o] = int16(v); o++ }
	putL := func(v uint32) { setLe32(ds[:], adrDrawList+2*o, v); o += 2 }
	for i := 0; i < n; i++ {
		idx := i
		if flag {
			idx = int(gCombatOrder[i])
		}
		p := &gParty[idx]
		t := int(gCondPortraitFaces[getWorstCondition(&gParty[idx])])
		if t > 4 && gRosterFaces[0] == 0 {
			pos = uint32(*gDseFaceGfx)
		} else {
			pos = uint32(gPartyFaces[idx])
		}
		if t > 4 {
			t -= 5
		}
		x := i * 0x26
		put(-1)
		putL(pos)
		put(int(gPortraitX[i]))
		put(0x95)
		put(0)
		put(t)
		if p.Bless != 0 || p.Shield != 0 || p.Holy != 0 || p.Hero != 0 {
			put(-1)
			putL(uint32(*gProtectIcons))
			if p.Bless != 0 {
				put(x + 0x20)
				put(0x96)
				put(0)
				put(2)
			}
			if p.Shield != 0 {
				put(x + 0xc)
				put(0xa8)
				put(0)
				put(1)
			}
			if p.Holy != 0 {
				put(x + 0xc)
				put(0x96)
				put(0)
				put(3)
			}
			if p.Hero != 0 {
				put(x + 0x1e)
				put(0xa6)
				put(0)
				put(0)
			}
		}
	}
	if n < 8 {
		put(-1)
		putL(uint32(*gRestoreIcons))
		for i := n; i < 8; i++ {
			put(int(gPortraitX[i]) - 1)
			put(0x94)
			put(0)
			put(i)
		}
	}
	if *gHpbarsIcons != 0 {
		put(-1)
		putL(uint32(*gHpbarsIcons))
		for i := 0; i < n; i++ {
			put(i*0x26 + 0xe)
			put(0xb6)
			put(0)
			c := i
			if flag {
				c = int(gCombatOrder[i])
			}
			p := &gParty[c]
			t := calcMaxHp(p)
			hp := p.HP.Get()
			switch {
			case hp < 1:
				put(4)
			case hp > t:
				put(3)
			case hp == t:
				put(0)
			case t>>2 > hp:
				put(2)
			default:
				put(1)
			}
		}
	}
	put(-1)
	put(0)
	put(0)
	printWindowText(buf[:])
}

func updateFreeMemKb() {
	var seg uint16
	*gFreeMemKb = uint16(allocmem(0xffff, &seg) >> 6)
}

func addViewSpritesD3(o int) int {
	var yo, k, xo int
	if int(*gObjD3c) != 0 {
		k = int(*gObjD3c) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x67)
		o = dlPut(o, 0x36)
		o = dlPut(o, 0x300|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD3l1) != 0 {
		k = int(*gObjD3l1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x2f)
		o = dlPut(o, 0x36)
		o = dlPut(o, 0x300|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD3r1) != 0 {
		k = int(*gObjD3r1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x9f)
		o = dlPut(o, 0x36)
		o = dlPut(o, 0x300|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD3l2) != 0 {
		k = int(*gObjD3l2) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0xfff7)
		o = dlPut(o, 0x36)
		o = dlPut(o, 0x300|gMapObjects[k].Attr.Get()|2)
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD3r2) != 0 {
		k = int(*gObjD3r2) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0xd7)
		o = dlPut(o, 0x36)
		o = dlPut(o, 0x300|gMapObjects[k].Attr.Get()|2)
		o = dlPut(o, objFrame(k))
	}
	if int(*gMonD3l11) != 0 {
		k = int(*gMonD3l11) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3l12) != 0 && int(*gMonD3l13) == 0 {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase2, k)+0x34)
		} else {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase2, k)+0x2b)
		}
		o = dlPut(o, hov(gHoverDy, *gHoverPhase2, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3l12) != 0 {
		k = int(*gMonD3l12) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3l13) == 0 {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x20)
		} else {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x19)
		}
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3l13) != 0 {
		k = int(*gMonD3l13) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)+0x3d)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3r11) != 0 {
		k = int(*gMonD3r11) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3r12) != 0 && int(*gMonD3r13) == 0 {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x91)
		} else {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x9b)
		}
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3r12) != 0 {
		k = int(*gMonD3r12) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3r13) == 0 {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase2, k)+0xa4)
		} else {
			o = dlPut(o, hov(gHoverDx, *gHoverPhase2, k)+0x89)
		}
		o = dlPut(o, hov(gHoverDy, *gHoverPhase2, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3r13) != 0 {
		k = int(*gMonD3r13) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)+0xad)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3l2) != 0 {
		k = int(*gMonD3l2) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)+2)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x3e)
		o = dlPut(o, 0x302)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3r2) != 0 {
		k = int(*gMonD3r2) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0xc1)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x3e)
		o = dlPut(o, 0x302)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD3c1) != 0 {
		k = int(*gMonD3c1) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase2]))
			yo = int(int8(gHoverDy[*gHoverPhase2]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3c2) != 0 && int(*gMonD3c3) == 0 {
			o = dlPut(o, xo+0x6b)
			xo = 0x6b
		} else {
			o = dlPut(o, xo+0x63)
			xo = 0x63
		}
		o = dlPut(o, yo+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD3c1) != 0 {
			*gMonHitAnimD3c1--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD3c1) != 0 {
				o = dlPut(o, xo+0xf)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(*gMonHitFxD3c1)-1)
			}
			if int(*gMonElemFxD3c1) != 0 {
				o = dlPut(o, xo+0xf)
				o = dlPut(o, 0x4a)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(*gMonElemFxD3c1))
			}
		}
	}
	if int(*gMonD3c2) != 0 {
		k = int(*gMonD3c2) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase1]))
			yo = int(int8(gHoverDy[*gHoverPhase1]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD3c3) == 0 {
			o = dlPut(o, xo+0x59)
			xo = 0x59
		} else {
			o = dlPut(o, xo+0x51)
			xo = 0x51
		}
		o = dlPut(o, yo+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD3c2) != 0 {
			*gMonHitAnimD3c2--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(ds[adrObjects1based]) != 0 {
				o = dlPut(o, xo+0xf)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(ds[adrObjects1based])-1)
			}
			if int(*gMonElemFxD3c2) != 0 {
				o = dlPut(o, xo+0xf)
				o = dlPut(o, 0x4a)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(*gMonElemFxD3c2))
			}
		}
	}
	if int(*gMonD3c3) != 0 {
		k = int(*gMonD3c3) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase3]))
			yo = int(int8(gHoverDy[*gHoverPhase3]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, xo+0x75)
		o = dlPut(o, yo+0x3e)
		o = dlPut(o, 0x300)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD3c3) != 0 {
			*gMonHitAnimD3c3--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD3c3) != 0 {
				o = dlPut(o, 0x84)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(*gMonHitFxD3c3)-1)
			}
			if int(*gMonElemFxD3c3) != 0 {
				o = dlPut(o, 0x84)
				o = dlPut(o, 0x4a)
				o = dlPut(o, 0x300)
				o = dlPut(o, int(*gMonElemFxD3c3))
			}
		}
	}
	return o
}

func addViewSpritesD2(o int) int {
	var yo, k, xo int
	if int(*gObjD2c) != 0 {
		k = int(*gObjD2c) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x58)
		o = dlPut(o, 0x29)
		o = dlPut(o, 0x200|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD2l1) != 0 {
		k = int(*gObjD2l1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 8)
		o = dlPut(o, 0x29)
		o = dlPut(o, 0x200|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD2r1) != 0 {
		k = int(*gObjD2r1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0xa8)
		o = dlPut(o, 0x29)
		o = dlPut(o, 0x200|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gMonD2l11) != 0 {
		k = int(*gMonD2l11) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)+0xd)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x3a)
		o = dlPut(o, 0x200)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD2l12) != 0 {
		k = int(*gMonD2l12) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)-0x15)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x3a)
		o = dlPut(o, 0x202)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD2r11) != 0 {
		k = int(*gMonD2r11) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x96)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x3a)
		o = dlPut(o, 0x200)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD2r12) != 0 {
		k = int(*gMonD2r12) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)+0xb8)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x3a)
		o = dlPut(o, 0x202)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD2c1) != 0 {
		k = int(*gMonD2c1) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase2]))
			yo = int(int8(gHoverDy[*gHoverPhase2]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD2c2) != 0 && int(*gMonD2c3) == 0 {
			o = dlPut(o, xo+0x62)
			xo = 0x62
		} else {
			o = dlPut(o, xo+0x51)
			xo = 0x51
		}
		o = dlPut(o, yo+0x3a)
		o = dlPut(o, 0x200)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD2c1) != 0 {
			*gMonHitAnimD2c1--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD2c1) != 0 {
				o = dlPut(o, xo+0x1c)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonHitFxD2c1)-1)
			}
			if int(*gMonElemFxD2c1) != 0 {
				o = dlPut(o, xo+0x1c)
				o = dlPut(o, 0x49)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonElemFxD2c1))
			}
		}
	}
	if int(*gMonD2c2) != 0 {
		k = int(*gMonD2c2) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase1]))
			yo = int(int8(gHoverDy[*gHoverPhase1]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD2c3) == 0 {
			o = dlPut(o, xo+0x40)
			xo = 0x40
		} else {
			o = dlPut(o, xo+0x2f)
			xo = 0x2f
		}
		o = dlPut(o, yo+0x3a)
		o = dlPut(o, 0x200)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD2c2) != 0 {
			*gMonHitAnimD2c2--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD2c2) != 0 {
				o = dlPut(o, xo+0x1c)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonHitFxD2c2)-1)
			}
			if int(*gMonElemFxD2c2) != 0 {
				o = dlPut(o, xo+0x1c)
				o = dlPut(o, 0x49)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonElemFxD2c2))
			}
		}
	}
	if int(*gMonD2c3) != 0 {
		k = int(*gMonD2c3) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase3]))
			yo = int(int8(gHoverDy[*gHoverPhase3]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, xo+0x73)
		o = dlPut(o, yo+0x3a)
		o = dlPut(o, 0x202)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD2c3) != 0 {
			*gMonHitAnimD2c3--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD2c3) != 0 {
				o = dlPut(o, 0x8f)
				o = dlPut(o, 0x44)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonHitFxD2c3)-1)
			}
			if int(*gMonElemFxD2c3) != 0 {
				o = dlPut(o, 0x8f)
				o = dlPut(o, 0x49)
				o = dlPut(o, 0x200)
				o = dlPut(o, int(*gMonElemFxD2c3))
			}
		}
	}
	return o
}

func addViewSpritesD1(o int) int {
	var yo, k, xo int
	if int(*gObjD1c) != 0 {
		k = int(*gObjD1c) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x46)
		o = dlPut(o, 0x17)
		o = dlPut(o, 0x100|gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD1l1) != 0 {
		k = int(*gObjD1l1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0xffea)
		o = dlPut(o, 0x17)
		o = dlPut(o, 0x100|gMapObjects[k].Attr.Get()|2)
		o = dlPut(o, objFrame(k))
	}
	if int(*gObjD1r1) != 0 {
		k = int(*gObjD1r1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0xa2)
		o = dlPut(o, 0x17)
		o = dlPut(o, 0x100|gMapObjects[k].Attr.Get()|2)
		o = dlPut(o, objFrame(k))
	}
	if int(*gMonD1l1) != 0 {
		k = int(*gMonD1l1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase3, k)-0x14)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase3, k)+0x33)
		o = dlPut(o, 0x102)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD1r1) != 0 {
		k = int(*gMonD1r1) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, hov(gHoverDx, *gHoverPhase1, k)+0x85)
		o = dlPut(o, hov(gHoverDy, *gHoverPhase1, k)+0x33)
		o = dlPut(o, 0x102)
		o = dlPut(o, monFrame(k))
	}
	if int(*gMonD1c1) != 0 {
		k = int(*gMonD1c1) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase2]))
			yo = int(int8(gHoverDy[*gHoverPhase2]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD1c2) != 0 && int(*gMonD1c3) == 0 {
			o = dlPut(o, xo+0x4a)
			xo = 0x4a
		} else {
			o = dlPut(o, xo+0x3a)
			xo = 0x3a
		}
		o = dlPut(o, yo+0x33)
		o = dlPut(o, 0x100)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD1c1) != 0 {
			*gMonHitAnimD1c1--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD1c1) != 0 {
				o = dlPut(o, xo+0x2e)
				o = dlPut(o, 0x46)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonHitFxD1c1)-1)
			}
			if int(*gMonElemFxD1c1) != 0 {
				o = dlPut(o, xo+0x2e)
				o = dlPut(o, 0x4b)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonElemFxD1c1))
			}
		}
	}
	if int(*gMonD1c2) != 0 {
		k = int(*gMonD1c2) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase1]))
			yo = int(int8(gHoverDy[*gHoverPhase1]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD1c3) == 0 {
			o = dlPut(o, xo+0x1d)
			xo = 0x1d
		} else {
			o = dlPut(o, xo+2)
			xo = 2
		}
		o = dlPut(o, yo+0x33)
		o = dlPut(o, 0x100)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD1c2) != 0 {
			*gMonHitAnimD1c2--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD1c2) != 0 {
				o = dlPut(o, xo+0x2e)
				o = dlPut(o, 0x46)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonHitFxD1c2)-1)
			}
			if int(*gMonElemFxD1c2) != 0 {
				o = dlPut(o, xo+0x2e)
				o = dlPut(o, 0x4b)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonElemFxD1c2))
			}
		}
	}
	if int(*gMonD1c3) != 0 {
		k = int(*gMonD1c3) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase3]))
			yo = int(int8(gHoverDy[*gHoverPhase3]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, xo+0x72)
		o = dlPut(o, yo+0x33)
		o = dlPut(o, 0x100)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD1c3) != 0 {
			*gMonHitAnimD1c3--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD1c3) != 0 {
				o = dlPut(o, 0xa0)
				o = dlPut(o, 0x46)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonHitFxD1c3)-1)
			}
			if int(*gMonElemFxD1c3) != 0 {
				o = dlPut(o, 0xaa)
				o = dlPut(o, 0x4b)
				o = dlPut(o, 0x100)
				o = dlPut(o, int(*gMonElemFxD1c3))
			}
		}
	}
	return o
}

func addViewSpritesD0(o int) int {
	var yo, k, xo int

	if int(*gObjD0c) != 0 && int(*gLastMoveDir) == int(*gPartyDir) {
		k = int(*gObjD0c) - 1
		o = dlPut(o, -1)
		o = dlPutL(o, gObjSprites[gMapObjects[k].Sprite.Get()])
		o++
		o = dlPut(o, 0x2f)
		o = dlPut(o, 8)
		o = dlPut(o, gMapObjects[k].Attr.Get())
		o = dlPut(o, objFrame(k))
	}
	if int(*gMonD0c2) != 0 {
		k = int(*gMonD0c2) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase1]))
			yo = int(int8(gHoverDy[*gHoverPhase1]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD0c3) == 0 {
			o = dlPut(o, xo+-0x11)
			xo = -0xd
		} else {
			o = dlPut(o, xo+-0x24)
			xo = -0x20
		}
		o = dlPut(o, yo+0x2b)
		o = dlPut(o, 2)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD0c2) != 0 {
			*gMonHitAnimD0c2--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD0c2) != 0 {
				o = dlPut(o, xo+int(int8(gSplashDx[*gMonHitSizeD0c2])))
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizeD0c2]))+0x2b)
				o = dlPut(o, int(gSplashFlags[*gMonHitSizeD0c2]))
				o = dlPut(o, int(*gMonHitFxD0c2)-1)
			}
			if int(*gMonElemFxD0c2) != 0 {
				o = dlPut(o, xo+int(int8(gSplashDx[*gMonHitSizeD0c2]))+10)
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizeD0c2]))+0x35)
				o = dlPut(o, int(gSplashFlags[*gMonElemSizeD0c2]))
				o = dlPut(o, int(*gMonElemFxD0c2))
			}
		}
	}
	if int(*gMonD0c3) != 0 {
		k = int(*gMonD0c3) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase3]))
			yo = int(int8(gHoverDy[*gHoverPhase3]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		o = dlPut(o, xo+0x5c)
		o = dlPut(o, yo+0x2b)
		o = dlPut(o, 2)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnimD0c3) != 0 {
			*gMonHitAnimD0c3--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFxD0c3) != 0 {
				o = dlPut(o, int(int8(gSplashDx[*gMonHitSizeD0c3]))+0x58)
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizeD0c3]))+0x2b)
				o = dlPut(o, int(gSplashFlags[*gMonHitSizeD0c3]))
				o = dlPut(o, int(*gMonHitFxD0c3)-1)
			}
			if int(*gMonElemFxD0c3) != 0 {
				o = dlPut(o, int(int8(gSplashDx[*gMonHitSizeD0c3]))+0x62)
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizeD0c3]))+0x35)
				o = dlPut(o, int(gSplashFlags[*gMonElemSizeD0c3]))
				o = dlPut(o, int(*gMonElemFxD0c3))
			}
		}
	}
	if int(*gMonD0c1) != 0 {
		k = int(*gMonD0c1) - 1
		if int(gMonHovers[gMonType[k]]) != 0 {
			xo = int(int8(gHoverDx[*gHoverPhase2]))
			yo = int(int8(gHoverDy[*gHoverPhase2]))
		} else {
			xo, yo = 0, 0
		}
		o = dlPut(o, -1)
		o = dlPutL(o, gMonSprites[gMonSpriteBank[k]])
		o++
		if int(*gMonD0c2) != 0 && int(*gMonD0c3) == 0 {
			o = dlPut(o, xo+0x48)
			xo = 0x48
		} else {
			o = dlPut(o, xo+0x1c)
			xo = 0x1c
		}
		o = dlPut(o, yo+0x2b)
		o = dlPut(o, 0)
		o = dlPut(o, monFrame(k))
		if int(*gMonHitAnim) != 0 {
			*gMonHitAnim--
			o = dlPut(o, -1)
			o = dlPutL(o, *gCharpowSprites)
			o++
			if int(*gMonHitFx) != 0 {
				o = dlPut(o, xo+int(int8(gSplashDx[*gMonHitSizes])))
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizes]))+0x2b)
				o = dlPut(o, int(gSplashFlags[*gMonHitSizes]))
				o = dlPut(o, int(*gMonHitFx)-1)
			}
			if int(*gMonElemFx) != 0 {
				o = dlPut(o, xo+int(int8(gSplashDx[*gMonHitSizes]))+10)
				o = dlPut(o, int(int8(gSplashDy[*gMonHitSizes]))+0x35)
				o = dlPut(o, int(gSplashFlags[*gMonElemSizes]))
				o = dlPut(o, int(*gMonElemFx))
			}
		}
	}
	return o
}
