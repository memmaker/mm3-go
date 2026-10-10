// Port of MAZELOAD.C (overlay 1821): maze/event/monster loading and saving, party damage,
// found loot, the random item generator, party swap and dismiss, text and number input, the
// saved-game list and the options menu (new game / load).
//
// g_maze_script is a signed char far *: its bytes are read sign-extended. load_maze's monster
// and object loops index other tables with them; negative indices go through ds directly.
package mm3

func saveMazeState() {
	var p Far
	var i int

	if *gMazeScript != 0 {
		sprintf(gTmpText, "maze%02u.evt", *gCurTown)
		curWriteFile(gostr(gTmpText), gMazeScript.B())
		saveLoadedMazes()
		sprintf(gTmpText, "maze%02u.bin", *gCurTown)
		p = loadCCFile(gostr(gTmpText))
		b := p.B()
		i = 0
		for j := 0; j < int(*gMonCount); j++ {
			b[i] = byte(gMonX[j])
			i++
			b[i] = byte(gMonY[j])
			i++
			i++
		}
		i += 6
		for j := 0; j < int(*gObjCount); j++ {
			b[i] = byte(gMapObjects[j].X.Get())
			i++
			b[i] = byte(gMapObjects[j].Y.Get())
			i++
			i++
		}
		curWriteFile(gostr(gTmpText), p.B())
		freeAndNull(&p)
	}
}

func loadMaze(n int) {
	var j, t, i int
	var k uint16

	saveMazeState()
	freeAndNull(gMazeScript)
	freeAndNull(gMapText)
	k, i = 0, 0
	*gMonCount, *gObjCount = 0, 0
	sprintf(gTmpText, "maze%02u.bin", n+1)
	*gMazeScript = loadCCFile(gostr(gTmpText))
	*gMazeEvtSize = uint16(*gFileSize)
	ms := gMazeScript.B()
	sc := func(k uint16) int { return int(int8(ms[k])) }
	for sc(k) != -1 {
		gMonX[i] = int16(sc(k))
		k++
		gMonY[i] = int16(sc(k))
		k++
		j = int(ms[k]) >> 2
		if j != 0 {
			j += 0x28
		}
		gMonSpriteBank[i] = int16(sc(k) & 3)
		k++
		if j != 0 {
			t = j
		} else {
			m := 0x40
			if n != 0x68 {
				m = n
			}
			t = int(gMazeMonIds1[m*3+int(gMonSpriteBank[i])])
		}
		gMonFrames[i] = int16(randomRange(0, int(gMonFrameCounts[t])))
		gMonVisible[i], gMonStatus[i] = 0, 0
		gMonHp[i] = gMonMaxHp.I16(2 * t)
		gMonType[i] = int16(t)
		i++
		*gMonCount++
	}
	if *gMonCount > 0xaa {
		fadeScreen(1, 4)
		sprintf(gTextBuf, "Max Monsters Exceeded. (%u - %u)", *gMonCount, 0xaa)
		showMessage(gTextBuf)
		quitToDos()
	}
	for i = 0; i < 5; i++ {
		k++
		gTextBuf2[i] = ms[k]
		if gTextBuf2[i] == 0x2a {
			if n == 0x29 && testBit(gGameFlags, 0x41) != 0 {
				gTextBuf2[i] = 0x3e
			} else if n == 0x2f && testBit(gGameFlags, 0x4e) != 0 {
				gTextBuf2[i] = 0x3e
			} else if n == 0x38 && testBit(gGameFlags, 0x5b) != 0 {
				gTextBuf2[i] = 0x3e
			}
		}
	}
	k++
	i = 0
	if (*gMazeEvtSize-k)/3 > 0x50 {
		fadeScreen(1, 4)
		sprintf(gTextBuf, "Max Objects Exceeded. (%u- %u)", *gObjCount, 0x50)
		showMessage(gTextBuf)
		quitToDos()
	}
	for k < *gMazeEvtSize {
		o := &gMapObjects[i]
		o.X.Set(sc(k))
		k++
		o.Y.Set(sc(k))
		k++
		o.Sprite.Set(sc(k))
		k++
		o.ID.Set(int(int8(ds[adrTextBuf2+o.Sprite.Get()])))
		o.Frame.Set(randomRange(0, int(ds[adrObjFrameCounts+o.ID.Get()])))
		i++
		gMapObjects[i].Attr.Set(0)
		*gObjCount++
	}
	freeAndNull(gMazeScript)
	if testBit(gGameFlags, 0x6e) != 0 && n == 0x2d {
		*gMapObj12X = 5
		*gMapObj12Y = 2
	}
	n++
	sprintf(gTmpText, "maze%02u.evt", n)
	*gMazeScript = loadCCFile(gostr(gTmpText))
	*gMazeEvtSize = uint16(*gFileSize)
	sprintf(gTmpText, "text%02u.maz", n)
	*gMapText = loadCCFile(gostr(gTmpText))
	*gMapTextSize = uint16(*gFileSize)
	indexEventScript()
	for j = 0; j < 5; j++ {
		if v := int(int8(gTextBuf2[j])); v != -1 {
			sprintf(gTmpText, "%s.pic", str(le16(ds[:], adrPicFileNames+2*v)))
			gObjSprites[j] = loadGraphic(gostr(gTmpText))
		}
	}
	if n == 0x69 {
		n = 0x41
	}
	if n < 0x42 {
		for j = 0; j < 3; j++ {
			if gMazeMonIds[n*3+j] != 0xff {
				sprintf(gTmpText, "%s.mon", str(gMonFileNames[gMazeMonIds[n*3+j]]))
				gMonSprites[j] = loadGraphic(gostr(gTmpText))
			}
		}
	}
}

func damageParty(dmg, typ, who int) {
	var i, a, b, f, pass, lim, save int
	var c *Chr
	var jump bool // the C's goto again into the loop body

	closeWindows(int(*gOpenWindows))
	*gOpenWindows = 0
	pass, f = 0, 0
	save = dmg
	lim = who + 1
	limit := func() int {
		if *gEventChar == 0 {
			return int(*gPartyCount)
		}
		return lim
	}
	if *gEventChar == 2 {
		for i, who, f = 0, 0, 0; i < int(*gPartyCount); i++ {
			switch getWorstCondition(&gParty[i]) {
			case 12, 13, 14, 15:
			default:
				if who == 0 {
					who = i + 1
				} else {
					f = i + 1
					who--
					jump = true
					goto again
				}
			}
		}
	}
	if *gEventChar == 0 {
		who = 0
	}
again:
	for jump || who < limit() {
		jump = false
		dmg = save
		c = &gParty[who]
		c.Cond[8] = 0
		a = 0
		switch typ {
		case 0:
			b = 4
		case 1:
			a = 6
			b = 0x12
		case 2:
			dmg = dmg - int(*gFireResist)
			a = 1
			b = 0xd
		case 3:
			dmg = dmg - int(*gElecResist)
			a = 2
			b = 0xe
		case 4:
			dmg = dmg - int(*gColdResist)
			a = 3
			b = 0xf
		case 5:
			dmg = dmg - int(*gPoisonResist)
			a = 4
			b = 0x11
		case 6:
			a = 5
			b = 0x10
		case 7:
			b = 0x35
		}
		dmg = int(int16(dmg))
		if typ != 0 {
			for {
				if charSavingThrow(c, typ) == 0 {
					break
				}
				dmg >>= 1
				if dmg == 0 {
					break
				}
			}
		}
		playSfx(b)
		drawIcon(*gCharpowSprites, a, who*0x26+0xb, 0x95, 0)
		if c.Shield != 0 {
			dmg -= int(c.Shield)
		}
		if dmg < 0 {
			dmg = 0
		}
		if typ == 7 {
			dmg = c.HP.Get()
			c.Cond[0xd] = 1
		}
		charTakeDamage(c, dmg)
		if f != 0 {
			break
		}
		who++
	}
	if f != 0 && pass == 0 {
		who = f - 1
		pass++
		jump = true
		goto again
	}
	idleFramesOrKey(2)
	drawPartyPortraits()
	updatePartyDisabled()
}

func partyHasHoloCards() int {
	n := 0
	for i := 0; i < int(*gPartyCount); i++ {
		for j := 0; j < 0x12; j++ {
			for k := 0; k < 6; k++ {
				if gParty[i].ItemID[j] == gHoloCardIds[k] {
					n++
				}
			}
		}
	}
	if n > 5 {
		return 1
	}
	return 0
}

func packsAllUndroppable() int {
	for i := 0; i < int(*gPartyCount); i++ {
		for j := 0; j < 0x12; j++ {
			if gParty[i].ItemFlags[j]&0x40 == 0 && isQuestItem(int(gParty[i].ItemID[j])) == 0 {
				return 0
			}
		}
	}
	return 1
}

func partyPacksFull() int {
	n := 0
	for i := 0; i < int(*gPartyCount); i++ {
		if gParty[i].ItemID[0x11] != 0 {
			n++
		}
	}
	if int(*gPartyCount) == n {
		return 1
	}
	return 0
}

func awardFoundLoot() {
	var k, save, i int
	var c *Chr
	m := func(a, i int) bool { return ds[a+i] != 0 } // g_mon_d0c_1[i] etc.

	if *gLootGold|*gLootGems|int32(*gLootCount) == 0 {
		return
	}
	if *gGameMode != 9 && (m(adrMonD0c1, 0) || m(adrMonD0c1, 3) || m(adrMonD0c1, 4) || m(adrMonD0c1, 5) || m(adrMonD2c1, 0) || m(adrMonD2l11, 0) || m(adrMonD2r11, 0) ||
		m(adrMonD2c1, 1) || m(adrMonD2l11, 1) || m(adrMonD2r11, 1) || m(adrMonD2l11, 2) || m(adrMonD2r11, 2) || m(adrMonD0c1, 1) || m(adrMonD0c1, 6) || m(adrMonD2c1, 2) ||
		m(adrMonD2l11, 3) || m(adrMonD2r11, 3) || m(adrMonD2c1, 3) || m(adrMonD2l11, 4) || m(adrMonD2r11, 4) || m(adrMonD0c1, 2) || m(adrMonD0c1, 7) || m(adrMonD2c1, 4) ||
		m(adrMonD2c1, 5) || m(adrMonD2l11, 5) || m(adrMonD2r11, 5)) {
		return
	}
	memset(gMissilePos, 0, 8)
	freeAndNull(gPowIconFar)
	*gSpellAnimActive = 0
	refreshView()
	if *gLootCount == 0 {
		playSfx(0x36)
	}
	pushButtons()
	setAnykeyButtons()
	flushInput()
	setMouseCursor(0, uint32(*gMouseCursor))
	sprintf(gTextBuf, "\x03cThe Party Found:\n\n\x03r\t072%lu Gold\n\t077%lu Gems", *gLootGold, *gLootGems)
	openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	if *gGameMode != 2 {
		*gGameMode = 7
	}
	if partyPacksFull() != 0 {
		if packsAllUndroppable() != 0 {
			for i = 0; i < int(*gPartyCount); i++ {
				memset(gParty[i].Cond[:], 1, 0x10)
			}
			goto done
		}
		showMessage(cs("\v007\f12Warning!  BackPacks Full!\fd\nPress a Key"))
	}
	for i = 0; i < int(*gLootCount); i++ {
		k = 0
		if partyPacksFull() != 0 {
			if isQuestItem(int(gLootIds[i])) == 0 {
				continue
			}
			showMessage(cs("\v007Your backpacks are full!\nPlease Drop an item!"))
			save = int(*gGameMode)
			*gGameMode = 1
			highlightPortrait(0)
			c = itemScreen(&gParty[0], 8)
			clearCharHighlight()
			*gGameMode = uint8(save)
		}
		for {
			if gParty[k].ItemID[0x11] == 0 {
				switch getWorstCondition(&gParty[k]) {
				case 8, 11, 12, 13, 14, 15:
				default:
					goto found
				}
			}
			if k++; k >= int(*gPartyCount) {
				break
			}
		}
		for k = 0; k < int(*gPartyCount); k++ {
			if gParty[k].ItemID[0x11] == 0 {
				goto found
			}
		}
		k = 0
	found:
		playSfx(0x24)
		c = &gParty[k]
		c.ItemFlags[0x11] = gLootFlags[i]
		c.ItemElem[0x11] = gLootElement[i]
		c.ItemMat[0x11] = gLootMaterial[i]
		c.ItemAttr[0x11] = gLootAttrib[i]
		c.ItemID[0x11] = gLootIds[i]
		c.ItemPower[0x11] = gLootCharges[i]
		sprintf(gTextBuf, "\x03l\v060\t000\x04077\n\x04077\n\x04077\n\x04077\n\x04077\n\x04077")
		printWindowText(gTextBuf)
		waitTicksOrKey(2)
		strcpy(gTextBuf2, str(gItemNames[c.ItemID[0x11]]))
		sprintf(gTextBuf, "\v060\t000\x03c%s found: %s", c.Name[:], gTextBuf2)
		printWindowText(gTextBuf)
		compactBackpack(c)
		waitTicksOrKey(0x14)
	}
	sprintf(gTextBuf, "\x03l\v120\t000\x04077\x03c\f27Hit a key\fd")
	printWindowText(gTextBuf)
	getInput()
done:
	if *gGameMode != 2 {
		*gGameMode = 1
	}
	closeWindows(1)
	*gPartyGold += uint32(*gLootGold)
	*gPartyGems += uint32(*gLootGems)
	*gLootCount = 0
	*gLootGold, *gLootGems = 0, 0
	if *gGameMode == 9 && *gObjD0c != 0 {
		o := &gMapObjects[*gObjD0c-1]
		switch o.ID.Get() {
		case 4, 0xc, 0x12, 0x14, 0x17, 0x21, 0x26, 0x28:
			o.X.Set(0x80)
			o.Y.Set(0x80)
			scanViewObjects()
			refreshView()
		}
	}
	popButtons()
}

func loadTextTable(name []byte) {
	*gTextTable = loadCCFile(gostr(name))
	tt := gTextTable.B()
	n := 1
	for i := 0; i < int(uint16(*gFileSize)); i++ {
		if tt[i] == 0 {
			gTextOffsets[n] = int16(i + 1)
			n++
		}
	}
}

func generateRandomItem(lvl int, c *Chr, slot int) {
	var n, item, r1, r2, k int
	var t, u [3]int

	clearItemSlot(c, slot)
	item = 0
	lvl = int(uint8(lvl))
	if lvl == 0 {
		lvl = 1
	}
	r1 = randomRange(1, 100)
	r2 = randomRange(1, 100)
	if r1 <= 0x23 {
		if r2 <= 0x1e {
			item = randomRange(1, 6)
		} else if r2 <= 0x3c {
			item = randomRange(7, 0x11)
		} else if r2 <= 0x55 {
			item = randomRange(0x12, 0x1d)
		} else {
			item = randomRange(0x1e, 0x21)
		}
	} else if r1 <= 0x3c {
		if r2 <= 0x46 {
			item = randomRange(0x22, 0x29)
		} else {
			item = 0x2a
		}
	} else {
		switch {
		case r2 <= 10:
			item = randomRange(0x2b, 0x2d)
		case r2 <= 0x14:
			item = 0x2e
		case r2 <= 0x23:
			item = 0x2f
		case r2 <= 0x2d:
			item = 0x30
		case r2 <= 0x37:
			item = randomRange(0x31, 0x33)
		case r2 <= 0x41:
			item = 0x34
		case r2 <= 0x4b:
			item = randomRange(0x35, 0x39)
		case r2 <= 0x50:
			item = randomRange(0x3a, 0x3c)
		default:
			item = randomRange(0x3d, 0x45)
		}
	}
	c.ItemID[slot] = uint8(item)
	if lvl == 6 {
		n = 1
		t[0] = randomRange(1, 4)
		u[0] = lvl - 1
	} else {
		n = randomRange(1, 100)
		if n <= 0x5f {
			n = 1
		} else if n <= 0x63 {
			n = 2
		} else {
			n = 3
		}
		switch n {
		case 3:
			t[2] = randomRange(1, 4)
			u[2] = randomRange(0, lvl-1)
			fallthrough
		case 2:
			u[1] = randomRange(0, lvl-1)
			for {
				t[1] = randomRange(1, 4)
				if t[1] != t[2] {
					break
				}
			}
			u[0] = lvl - 1
			for {
				t[0] = randomRange(1, 4)
				if !(t[0] == t[1] || t[0] == t[2]) {
					break
				}
			}
		case 1:
			r1 = randomRange(1, 100)
			u[0] = lvl - 1
			if item >= 0x3d && r1 <= 0x5a {
				t[0] = 4
			} else if r1 <= 0x5d {
				t[0] = 3
			} else if r1 <= 0x60 {
				t[0] = 1
			} else if r1 <= 0x63 {
				t[0] = 2
			} else {
				t[0] = 4
			}
		}
	}
	for j := 0; j < n; j++ {
		switch t[j] {
		case 1:
			k = randomRange(1, 100)
			switch {
			case k <= 0x19:
				k = 0
			case k <= 0x2d:
				k = 1
			case k <= 0x3c:
				k = 2
			case k <= 0x4b:
				k = 3
			case k <= 0x5f:
				k = 4
			default:
				k = 5
			}
			c.ItemElem[slot] = uint8(randomRange(int(gItemElemRanges[k][u[j]][0]), int(gItemElemRanges[k][u[j]][1])))
			c.ItemElem[slot] += gItemElemBases[k]
		case 2:
			k = randomRange(1, 100)
			switch {
			case k <= 0xf:
				k = 0
			case k <= 0x19:
				k = 1
			case k <= 0x23:
				k = 2
			case k <= 0x32:
				k = 3
			case k <= 0x41:
				k = 4
			case k <= 0x50:
				k = 5
			case k <= 0x55:
				k = 6
			case k <= 0x5a:
				k = 7
			case k <= 0x5f:
				k = 8
			default:
				k = 9
			}
			c.ItemAttr[slot] = uint8(randomRange(int(gItemAttrRanges[k][u[j]][0]), int(gItemAttrRanges[k][u[j]][1])))
			c.ItemAttr[slot] += gItemAttrBases[k]
		case 3:
			k = randomRange(1, 100)
			if lvl == 6 {
				k = 1
			} else if k <= 0x46 {
				k = 0
			} else {
				k = 1
			}
			c.ItemMat[slot] = uint8(randomRange(int(gItemMatRanges[k][u[j]][0]), int(gItemMatRanges[k][u[j]][1])))
			c.ItemMat[slot] += gItemMatBases[k]
		case 4:
			c.ItemPower[slot] = uint8(randomRange(int(gItemPowerMin[lvl]), int(gItemPowerMax[lvl])))
			c.ItemFlags[slot] = uint8(randomRange(1, 10) + 5)
		}
	}
}

// beginWhomPrompt's struct spr1 b is on its own stack frame in the C, gone when the window is
// still open; the window has drawn it by then. Here it stays reserved until endWhomPrompt.
func beginWhomPrompt(s1, s2 *int, verb int) {
	pushButtons()
	pushIcons("esc.icn")
	adrB := dsAlloca(0x14)
	b := at[Spr1](adrB)
	b.Flag.Set(-1)
	b.End.Set(-1)
	b.X.Set(0xb1)
	b.Img.Set(uint32(*gCurIcons))
	*gPartyBtnXEsc = 0xeb
	*gPartyBtnYEsc = 0x78
	*gPartyBtnWEsc = 0x18
	*gPartyBtnHEsc = 0x14
	*s1 = int(*gRosterSlotBtnX)
	*gRosterSlotBtnX = -1
	*gBtnXs = adrPartyBtnX
	*gBtnYs = adrPartyBtnY
	*gBtnWs = adrPartyBtnW
	*gBtnHs = adrPartyBtnH
	*s2 = int(*gPortraitClickOn)
	*gPortraitClickOn = 1
	sprintf(gTextBuf, "\t010\v005%s whom?\x05%04X", str(gWhomVerbs[verb]), adrB)
	openTextWindow(0x32, 0x70, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTextBuf))
}

func endWhomPrompt(s1, s2 int) {
	closeWindows(1)
	popIcons()
	popButtons()
	*gPartyBtnXEsc, *gPartyBtnYEsc, *gPartyBtnWEsc, *gPartyBtnHEsc = 0, 0, 0, 0
	*gRosterSlotBtnX = int16(s1)
	*gPortraitClickOn = uint8(s2)
	drawPartyPortraits()
}

func exchangePartyMember(who *int) *Chr {
	var s1, s2, cur, key int
	defer dsFree(dsSP)

	c := &gParty[*who]
	beginWhomPrompt(&s1, &s2, 0)
	cur = *who
	for {
		key = getInput()
		switch key {
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			key -= 0xc9
			if key < int(*gPartyCount) {
				gParty[key], gParty[*who] = gParty[*who], gParty[key]
				gPartyIds[key], gPartyIds[*who] = gPartyIds[*who], gPartyIds[key]
				gPartyFacesF[key], gPartyFacesF[*who] = gPartyFacesF[*who], gPartyFacesF[key]
				cur = key
				*who = key
				c = &gParty[key]
				key = 0x1b
			}
		}
		if key == 0x1b || key == 6 {
			break
		}
	}
	endWhomPrompt(s1, s2)
	highlightPortrait(cur)
	return c
}

func dismissPartyMember() {
	var p Far
	var t, s1, s2, key int
	defer dsFree(dsSP)

	if gMapsCanDismiss[*gMapQuadrant][0] == 0 {
		showMessage(cs("Too dangerous to dismiss here!"))
		return
	}
	beginWhomPrompt(&s1, &s2, 1)
	for {
		key = getInput()
		switch key {
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			key -= 0xc9
			if key < int(*gPartyCount) {
				if *gPartyCount == 1 || countPartyNonHirelings() == 1 && *gPartyCount != 1 && gPartyIds[key] < 0x14 {
					showMessage(cs("You cannot dismiss your last character!"))
					goto out
				}
				p = memAlloc(0x12f)
				copy(p.B(), bytesOf(&gParty[key]))
				t = int(gPartyIds[key])
				gPartyIds[key] = 0xff
				freeAndNull(&gPartyFacesF[key])
				compactPartyFaces()
				compactParty()
				*gPartyCount--
				copy(bytesOf(&gParty[*gPartyCount]), p.B()[:0x12f])
				gPartyIds[*gPartyCount] = uint8(t)
				freeAndNull(&p)
				goto out
			}
		}
		if key == 0x1b || key == 6 {
			break
		}
	}
out:
	endWhomPrompt(s1, s2)
}

func inputText(buf []byte, max, width int) int {
	var ch uint8
	var tmp [50]byte

	n := 0
	sprintf(tmp[:], "\x03l\t000\x04%03d\x03c", width)
	printWindowText(tmp[:])
	for {
		ch = uint8(waitKey())
		show := false
		if (ch == 8 || ch == 0xf0) && n != 0 {
			n--
			show = true
		} else if ch >= 0x20 && ch <= 0x7f && n < max && (n != 0 || ch != 0x20) {
			buf[n] = ch
			n++
			show = true
		}
		if show {
			buf[n] = 0
			sprintf(tmp[:], "\x03l\t000\x04%03d\x03c%s", width, buf)
			printWindowText(tmp[:])
		}
		if ch == 0x1b || ch == 0xd {
			break
		}
	}
	if ch == 0x1b {
		n = 0
	}
	buf[n] = 0
	return n
}

func inputNumber(max, width int) int32 {
	var ch uint8
	var tmp, buf [20]byte

	n := 0
	sprintf(tmp[:], "\x03l\t000\x04%03d\x03c", width)
	printWindowText(tmp[:])
	for {
		ch = uint8(waitKey())
		show := false
		if (ch == 8 || ch == 0xf0) && n != 0 {
			n--
			show = true
		} else if ch >= '0' && ch <= '9' && n < max && (n != 0 || ch != ' ') {
			buf[n] = ch
			n++
			show = true
		}
		if show {
			buf[n] = 0
			sprintf(tmp[:], "\x03l\t000\x04%03d\x03c%s", width, buf[:])
			printWindowText(tmp[:])
		}
		if ch == 0x1b || ch == 0xd {
			break
		}
	}
	if ch == 0x1b {
		return 0
	}
	buf[n] = 0
	return atol(buf[:]) // sscanf "%lu": at most a few digits
}

func compareStrings(a, b []byte) int { return strcmp(a, b) }

func selectSavedGame() int {
	var sel, shown, busy, save, top, k int
	var mark, lvl [10]uint8
	var names [10][0x1f]byte
	var strs [10][0x23]byte

	save = int(*gPortraitClickOn)
	*gPortraitClickOn = 0
	pushButtons()
	pushIcons("scroll.icn")
	top, shown, busy = 0, 0, 0
	gScrollSpr.Img.Set(uint32(*gCurIcons))
	*gBtnXs = adrScrollBtnX
	*gBtnYs = adrScrollBtnY
	*gBtnWs = adrScrollBtnW
	*gBtnHs = adrScrollBtnH
	gScrollCastSprites.Img.Set(uint32(*gCurIcons))
	sel = -1
	for k = 0; k < int(*gSaveCount); k++ {
		readSaveSummary(k, names[k][:], &lvl[k])
		sprintf(strs[k][:], "\x03r\t149\x04030\x03r\t149L %u\x03l", lvl[k])
	}
	page := func() {
		for getMouseState(gMouseX, gMouseY) != 0 {
		}
		if top+10 > int(*gSaveCount) {
			if top = int(*gSaveCount) - 10; top < 1 {
				top = 0
			}
		}
	}
redraw:
	{
		args := []any{adrScrollSpr, adrScrollCastSprites}
		for r := 0; r < 10; r++ {
			var s []byte
			if names[top+r][0] != 0 {
				s = strs[top+r][:]
			}
			args = append(args, mark[top+r], names[top+r][:], s)
		}
		var f []byte
		if sel >= 0 {
			f = gSaveFiles[sel][:]
		}
		args = append(args, f)
		sprintf(gTextBuf, "\r\x05%04X\x05%04X\x02\x03c\t000\v002Available Games to load\x03l\v015\t001"+
			"1\t010\f%2u%s%s\fd\n2\t010\f%2u%s%s\fd\n3\t010\f%2u%s%s\fd\n4\t010\f%2u%s%s\fd\n"+
			"5\t010\f%2u%s%s\fd\n6\t010\f%2u%s%s\fd\n7\t010\f%2u%s%s\fd\n8\t010\f%2u%s%s\fd\n"+
			"9\t010\f%2u%s%s\fd\n0\t010\f%2u%s%s\fd\x03l\t004\v110File: %s\x01", args...)
	}
	if shown != 0 {
		printWindowText(gTextBuf)
	} else {
		openTextWindow(0x1b, 6, 0xb4, 0x88, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		shown = 1
		for getMouseState(gMouseX, gMouseY) != 0 {
		}
	}
	if busy != 0 {
		for {
			if busy = getMouseState(gMouseX, gMouseY); busy == 0 {
				break
			}
		}
	}
	for {
		k = getInput()
		switch k {
		case 0, 0xf2:
			if top--; top < 0 {
				top = 0
			}
			goto redraw
		case 1, 0xf3:
			if top++; top > int(*gSaveCount)-10 {
				if top = int(*gSaveCount) - 10; top < 1 {
					top = 0
				}
			}
			goto redraw
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
			if k >= '1' {
				k -= 0x2f
			} else if k == '0' {
				k = 11
			}
			if k--; k > int(*gSaveCount) {
				break
			}
			if k+top-1 == sel {
				mark[sel] = 0
				sel = -1
			} else {
				mark = [10]uint8{}
				sel = k + top - 1
				mark[sel] = 0x12
			}
			busy = 1
			goto redraw
		case 0xc, 0x1b:
			sel = -1
			k = 0x1b
		case 0xd, 0xde:
			if k == 0xd && *gMouseButtons == 0 { // goto ok
				if sel != -1 {
					k = 0x1b
				}
				break
			}
			if top -= 10; top < 0 {
				top = 0
			}
			page()
			goto redraw
		case 0xe, 0xdf:
			top += 10
			page()
			goto redraw
		case 0xf, 'S':
			if sel != -1 {
				k = 0x1b
			}
		}
		if k == 0x1b {
			break
		}
	}
	closeWindows(1)
	popIcons()
	popButtons()
	*gPortraitClickOn = uint8(save)
	flushInput()
	return sel
}

var promptGameNameText = cs("\x03cEnter a Name for this Game\x03l\n\n")

func promptGameName(name []byte) int {
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, farOf(promptGameNameText))
	n := inputText(name, 0x18, 0xc8)
	closeWindows(1)
	if n == 0 {
		return 0
	}
	return 1
}

func findFreeSaveSlot() {
	var name [16]byte

	sprintf(name[:], "save%02d.mm3", *gSaveCount)
	if probeFile(name[:]) != -1 {
		for i := int8(0); i < *gSaveCount; i++ {
			sprintf(name[:], "save%02d.mm3", i)
			if probeFile(name[:]) == -1 {
				*gSaveCount = i
				return
			}
		}
	}
}

var (
	optionsBuf     [100]byte // options_menu's buf: the window gets its far address
	maxSavesText   = cs("You have reached the maximum number of saved games.  Delete old games not in use or reinstall MM3 in a new directory.")
)

func optionsMenu(intro int) {
	var sel, key int
	var name [0x20]byte
	buf := optionsBuf[:]

	if *gMonMaxHp == 0 {
		loadMonsterTables()
	}
	if intro != 0 {
		clearScreen()
		if *gCfgVideo == 4 {
			showRawPicture("logy5.raw")
		} else {
			showRawPicture("control.raw")
		}
		fadeScreen(1, 4)
		setMouseCursor(0, uint32(*gMouseCursor))
	}
	*gCurIcons = loadGraphic("start.icn")
	gOptionsSpr[0].Img.Set(uint32(*gCurIcons))
	sprintf(buf, "\r\x05%04X\x03c\v003\t000Might and Magic III Options\x02\x03c\t000\v070Copyright 1991 NWC, Inc.\x01", adrOptionsSpr)
	setAnykeyButtons()
	*gSaveSlot = 0
	if intro != 0 {
		openTextWindow(0x4b, 0x3b, 0xac, 0x5f, 0, 0x96, 4, 6, farOf(buf))
	} else {
		openTextWindow(0x4b, 0x3b, 0xac, 0x5f, 0, 0x96, 4, 6, dsFar(int(*gMsgCopyright)))
		key = getInput()
		if key == 0x11 {
			quitToDos()
		}
		printWindowText(buf)
	}
	*gBtnXs = adrOptionsBtnX
	*gBtnYs = adrOptionsBtnY
	*gBtnWs = adrOptionsBtnW
	*gBtnHs = adrOptionsBtnH
	for {
		key = getInput()
		switch key {
		case 0, 'S':
			findSaveFiles(cs("save??.mm3"))
			if *gSaveCount >= 10 {
				openTextWindow(0x23, 0x46, 0xeb, 0x37, 0, 0x96, 4, 6, farOf(maxSavesText))
				for getMouseState(gMouseX, gMouseY) != 0 {
				}
				getInput()
				closeWindows(1)
				break
			}
			if promptGameName(name[:]) == 0 {
				break
			}
			findFreeSaveSlot()
			sprintf(gSaveFiles[0][:], "save%02d.mm3", *gSaveCount)
			*gRosterFar = memAlloc(0x2382)
			initWorkArchive()
			if *gArchiveOpen++; *gArchiveOpen == 1 {
				openCurArchive()
			}
			loadPartyData()
			strcpy(gSaveName, name[:])
			restockShops()
			*gDay = 1
			*gYear = 0x1f4
			writeSaveGame()
			freeAndNull(gRosterFar)
			goto saved
		case 1, 'L':
			findSaveFiles(cs("save??.mm3"))
			if *gSaveCount != 0 {
				if *gSaveCount == 1 {
					*gSaveSlot = 0
				} else {
					if sel = selectSavedGame(); sel == -1 {
						break
					}
					*gSaveSlot = int8(sel)
				}
				closeWindows(1)
				copySaveToWork()
				goto loaded
			}
			sprintf(gTmpText, "\x03c\v010No saved file to load...")
			openTextWindow(0x2a, 0x4c, 0xe0, 0x30, 0, 0x96, 4, 6, dsFar(adrTmpText))
			pushButtons()
			setAnykeyButtons()
			getInput()
			popButtons()
			closeWindows(1)
			for getMouseState(gMouseX, gMouseY) != 0 {
			}
		case 0x11:
			quitToDos()
		}
	}
saved:
	closeWindows(1)
loaded:
	freeAndNull(gCurIcons)
	if *gArchiveOpen++; *gArchiveOpen == 1 {
		openCurArchive()
	}
	setAnykeyButtons()
	if key == 0x1b {
		quitToDos()
	}
}
