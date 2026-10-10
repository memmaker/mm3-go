// Port of CHARINFO.C (overlay 180c): the character sheet's info popups, message box, icon and
// button-table stacks, party/backpack compaction, the inn roster screen (add, remove, create,
// delete characters).
package mm3

// Views where charinfo.c types a global differently from ds_gen.go.
var (
	gRosterSlotBtnXs = arr[int16](adrRosterSlotBtnX) // int g_roster_slot_btn_x[4]
	gInnIconsF       = at[Far](adrInnIcons)
)

// rosterAt is &g_roster[i] (g_roster: struct chr far *).
func rosterAt(i int) *Chr { return farAt[Chr](gRosterFar.Add(i * 0x12f)) }

func showCharsheetPopup(sel int, c *Chr) {
	var k, i, h, n, x, y, j, w int
	var a, b uint32
	var lines [20][30]byte

	h = 0x8f
	w = 0x34
	a = 0
	s := "s"
	x = int(gCharsheetPopupX[sel])
	y = int(gCharsheetPopupY[sel])
	label := str(gCharsheetLabels[sel])
	st := func(k int) []byte { // ST(k)
		if c.Award[k] != 0 {
			return getCharStatText(c, k, gDrawListB[k*0x22:])
		}
		return nil
	}
	switch sel {
	case 0, 1, 2, 3, 4, 5, 6:
		a = uint32(getCharStat(c, sel, 0))
		b = uint32(getCharStat(c, sel, 1))
		for i = 0; uint32(gStatRatingLimits[i]) <= a; i++ {
		}
		sprintf(gTextBuf, "\x02\x03c%s\nCurrent / Maximum\n\x03r\t054%lu\x03l\t058/ %lu\n\x03cRating: %s\x01",
			label, a, b, str(gRatingNames[i]))
	case 7:
		n = getCharAge(c, 0)
		j = getCharAge(c, 1)
		sprintf(gTextBuf, "\x02\x03c%s\nCurrent / Natural\n\x03r\t057%u\x03l\t061/ %u\n\x03cBorn: %u / %u\x01",
			label, n, j, c.BirthDay, c.BirthYear.Get())
		w = 0x35
	case 8:
		if j = getCharLevel(c)/int(gClassAttackLevels[c.Cls]) + 1; j == 1 {
			s = ""
		}
		sprintf(gTextBuf, "\x02\x03c%s\nCurrent / Maximum\n\x03r\t054%u\x03l\t058/ %u\n\x03c%u Attack%s/Round\x01",
			label, getCharLevel(c), c.Level, j, s)
		w = 0x35
	case 9, 11:
		if sel == 9 {
			a = uint32(uint16(calcArmorClass(c, 0)))
			b = uint32(uint16(calcArmorClass(c, 1)))
		} else {
			a = uint32(int32(c.SP.Get()))
			b = uint32(uint16(calcMaxSp(c)))
		}
		sprintf(gTextBuf, "\x02\x03c%s\nCurrent / Maximum\n\x03r\t054%lu\x03l\t058/ %lu", label, a, b)
		w = 0x2a
	case 10:
		sprintf(gTextBuf, "\x02\x03c%s\nCurrent / Maximum\n\x03r\t054%d\x03l\t058/ %lu", label, c.HP.Get(),
			uint32(uint16(calcMaxHp(c))))
		w = 0x2a
	case 12:
		res := bytesOf(c)[0x107:]
		sprintf(gTextBuf, "\x02\x03c%s\x03l\n\t020Fire\t100%u\n\t020Cold\t100%u\n\t020Electricity\t100%u\n\t020Poison\t100%u\n\t020Energy\t100%u\n\t020Magic\t100%u",
			label,
			int(res[0])+getItemStatBonus(c, 0xb)+int(res[1]),
			int(res[2])+getItemStatBonus(c, 0xd)+int(res[3]),
			int(res[4])+getItemStatBonus(c, 0xc)+int(res[5]),
			int(res[6])+getItemStatBonus(c, 0xe)+int(res[7]),
			int(res[8])+getItemStatBonus(c, 0xf)+int(res[9]),
			int(res[10])+getItemStatBonus(c, 0x10)+int(res[11]))
		w = 0x50
	case 13:
		n = countSkills(c)
		if n != 0 {
			for j = 0; j < 0x12; j++ {
				if c.Skill[j] != 0 {
					if j == 0 {
						sprintf(lines[j][:], "\n\t020%s%u", str(gSkillNames[j]), getThievery(c))
					} else {
						sprintf(lines[j][:], "\n\t020%s", str(gSkillNames[j]))
					}
				} else {
					strcpy(lines[j][:], cs(""))
				}
			}
		} else {
			strcpy(lines[0][:], cs("\n\t020None"))
			for j = 1; j < 0x12; j++ {
				strcpy(lines[j][:], cs(""))
			}
			n = 1
		}
		l := func(k int) []byte { return lines[k][:] }
		sprintf(gTextBuf, "\x02\x03c%s\x03l%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", label,
			l(0), l(1), l(2), l(3), l(4), l(5), l(17), l(6), l(7), l(8),
			l(9), l(10), l(11), l(12), l(13), l(16), l(14), l(15))
		w = n*9 + 0x1a
		y -= (n >> 1) << 3
		if y+w > 0xc7 {
			y = 0xc7 - w
		}
	case 14:
		loadTextTable(cs("award.bin"))
		memset(gDrawListB, 0, 0xfa0)
		sprintf(gTextBuf, "\x01\x03c%s for %s the %s\x03l\x02\n\v012%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\v012\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\n\t152%s\x01",
			label, c.Name[:], str(gClassNames[c.Cls]),
			st(0), st(1), st(2), st(3), st(4), st(5), st(21), st(6), st(18), st(14), st(22), st(20),
			st(19), st(15), st(16), st(17), st(11), st(12), st(13), st(7), st(9), st(8), st(10), st(23))
		openTextWindow(0, 0, 0x140, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		freeAndNull(gTextTable)
		goto done
	case 15:
		if b = uint32(getExpToNextLevel(c)); b == 0 {
			strcpy(gTmpText, cs("\f12Eligible\fd"))
		} else {
			ultoa(b, gTmpText, 10)
		}
		sprintf(gTextBuf, "\x02\x03c%s\x03l\n\t010Current:\t070%lu\n\t010Next Level:\t070%s\x01", label,
			getTotalExp(c), gTmpText)
		w = 0x2b
	case 16, 17:
		if sel == 16 {
			a = *gPartyGold
			b = *gBankGold
		} else {
			a = *gPartyGems
			b = *gBankGems
		}
		sprintf(gTextBuf, "\x02\x03cParty %s\n%lu on hand\n%lu in bank\x01\x03l", label, a, b)
		w = 0x2b
	case 18:
		if j = int(*gFood) / int(*gPartyCount) / 3; j == 1 {
			s = ""
		}
		sprintf(gTextBuf, "\x02\x03cParty %s\n%u on hand\nEnough for %u day%s\x01\x03l", label, *gFood, j, s)
		w = 0x2b
	case 19:
		j = 0
		for k = 0; k < 0x10; k++ {
			if c.Cond[k] != 0 {
				if k < 0xc {
					sprintf(lines[k][:], "\n\t020%s\t095-%d", str(gConditionNames[k]), c.Cond[k])
				} else {
					sprintf(lines[k][:], "\n\t020%s", str(gConditionNames[k]))
				}
				j++
			} else {
				strcpy(lines[k][:], cs(""))
			}
		}
		if getWorstCondition(c) == 0x10 {
			sprintf(lines[0][:], "\n\t020%s", str(*gStrGood))
			j++
		}
		for k = 16; k < 20; k++ {
			strcpy(lines[k][:], cs(""))
		}
		if c.Bless != 0 {
			sprintf(lines[16][:], "\n\t020Blessed\t095%+d", c.Bless)
			j++
		}
		if c.Shield != 0 {
			sprintf(lines[17][:], "\n\t020Power Shield\t095%+d", c.Shield)
			j++
		}
		if c.Holy != 0 {
			sprintf(lines[18][:], "\n\t020Holy Bonus\t095%+d", c.Holy)
			j++
		}
		if c.Hero != 0 {
			sprintf(lines[19][:], "\n\t020Heroism\t095%+d", c.Hero)
			j++
		}
		var args []any
		args = append(args, label)
		for k = 0; k < 20; k++ {
			args = append(args, lines[k][:])
		}
		sprintf(gTextBuf, "\x02\x03c%s\x03l%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s\x01", args...)
		w = j*9 + 0x1a
		y -= ((j - 1) >> 1) << 3
		if y+w > 0xc7 {
			y = 0xc7 - w
		}
	}
	openTextWindow(x, y, h, w, 0, 0x96, 4, 6, dsFar(adrTextBuf))
done:
	setAnykeyButtons()
	getInput()
	setCharscrHotspots()
	printWindowText(cs("\x01"))
	closeWindows(1)
	flushInput()
}

func freePartyFaces() {
	for i := 0; i < 8; i++ {
		freeAndNull(&gPartyFacesF[i])
	}
	freeAndNull(gDseFaceGfxF)
}

func drawCharsheetCursor(n int) {
	if *gStatCursor < 0x14 {
		drawIcon(*gCurIcons, n+0x30, int(gCharsheetFieldX[*gStatCursor/5]), int(gCharsheetFieldY[*gStatCursor%5]), 0)
	}
}

func partyHasOnlyHirelings() int {
	for i := 0; i < int(*gPartyCount); i++ {
		if gPartyIds[i] < 0x14 {
			return 0
		}
	}
	return 1
}

func countPartyNonHirelings() int {
	n := 0
	for i := 0; i < int(*gPartyCount); i++ {
		if gPartyIds[i] < 0x14 {
			n++
		}
	}
	return n
}

func drawBackdrop() {
	fadeScreen(0, 4)
	clearScreen()
	updateFreeMemKb()
	showRawPicture("back.raw")
	sprintf(gTextBuf, "\x03d\x05%04X", adrBorderSprites)
	printWindowText(gTextBuf)
	setMouseCursor(0, uint32(*gMouseCursor))
}

var showMessageHdr = cs("\x03c\v010") // show_message's window text (a far pointer)

func showMessage(msg []byte) {
	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	openTextWindow(0x2a, 0x4c, 0xe0, 0x30, 0, 0x96, 4, 6, farOf(showMessageHdr))
	printWindowText(msg)
	waitMouseRelease()
	getInput()
	closeWindows(1)
	waitMouseRelease()
	popButtons()
	*gGameMode = save
}

func pushIcons(name string) {
	*gIconStackTop++
	gIconStack[*gIconStackTop] = *gCurIcons
	*gCurIcons = loadGraphic(name)
}

func popIcons() {
	freeAndNull(gCurIcons)
	*gCurIcons = gIconStack[*gIconStackTop]
	*gIconStackTop--
}

func pushButtons() {
	*gBtnStackTop++
	gBtnStackX[*gBtnStackTop] = *gBtnXs
	gBtnStackY[*gBtnStackTop] = *gBtnYs
	gBtnStackW[*gBtnStackTop] = *gBtnWs
	gBtnStackH[*gBtnStackTop] = *gBtnHs
}

func popButtons() {
	*gBtnXs = gBtnStackX[*gBtnStackTop]
	*gBtnYs = gBtnStackY[*gBtnStackTop]
	*gBtnWs = gBtnStackW[*gBtnStackTop]
	*gBtnHs = gBtnStackH[*gBtnStackTop]
	*gBtnStackTop--
}

func formatBigNumber(v uint32, buf []byte) []byte {
	if v > 999999 {
		sprintf(buf, "%lu mil", v/1000000)
	} else {
		sprintf(buf, "%lu", v)
	}
	return buf
}

func setCombatButtons() {
	*gBtnXs = adrMazeBtnX
	*gBtnYs = adrMainBtnY
	*gBtnWs = adrMainBtnW
	*gBtnHs = adrMainBtnH
	gMazeBtnX[10] = 0xef
}

func setMainButtons() {
	*gBtnXs = adrMazeBtnX
	*gBtnYs = adrMainBtnY
	*gBtnWs = adrMainBtnW
	*gBtnHs = adrMainBtnH
	gMazeBtnX[10] = -1
}

func setAnykeyButtons() {
	*gBtnXs = adrAnyclickBtnXs
	*gBtnYs = adrAnykeyBtnY
	*gBtnWs = adrAnykeyBtnW
	*gBtnHs = adrAnykeyBtnH
}

func compactParty() {
	for i := 0; i < int(*gPartyCount); i++ {
		if gPartyIds[i] == 0xff {
			for j := i + 1; j < int(*gPartyCount); j++ {
				if gPartyIds[j] < 0xff {
					gPartyIds[i] = gPartyIds[j]
					gParty[i] = gParty[j]
					gPartyIds[j] = 0xff
					gParty[j] = Chr{}
					break
				}
			}
		}
	}
}

func compactPartyFaces() {
	for i := 0; i < 8; i++ {
		if gPartyFacesF[i] == 0 {
			for j := i + 1; j < 8; j++ {
				if gPartyFacesF[j] != 0 {
					gPartyFacesF[i] = gPartyFacesF[j]
					gPartyFacesF[j] = 0
					break
				}
			}
		}
	}
}

func clearItemSlot(c *Chr, slot int) {
	c.ItemElem[slot], c.ItemMat[slot], c.ItemAttr[slot], c.ItemID[slot], c.ItemPower[slot], c.ItemEq[slot], c.ItemFlags[slot] = 0, 0, 0, 0, 0, 0, 0
}

func compactBackpack(c *Chr) {
	for i := 0; i < 0x12; i++ {
		if c.ItemID[i] == 0 {
			clearItemSlot(c, i)
			for j := i + 1; j <= 0x12; j++ {
				if c.ItemID[j] != 0 {
					c.ItemElem[i] = c.ItemElem[j]
					c.ItemMat[i] = c.ItemMat[j]
					c.ItemAttr[i] = c.ItemAttr[j]
					c.ItemID[i] = c.ItemID[j]
					c.ItemPower[i] = c.ItemPower[j]
					c.ItemEq[i] = c.ItemEq[j]
					c.ItemFlags[i] = c.ItemFlags[j]
					clearItemSlot(c, j)
					break
				}
			}
		}
	}
}

func buildRosterPage(start int, list []byte) {
	var p [4]*Chr
	var names [4][16]byte
	var race, cls, st [4][15]byte

	hire := cs("\f12For Hire\fd")
	for i := 0; i < 4; i++ {
		in := 0
		v := int(list[start+i])
		for j := 0; j < int(*gPartyCount); j++ {
			if int(gPartyIds[j]) == v {
				in = 1
			}
		}
		if v != 0xff {
			if i&1 != 0 {
				gRosterSlotBtnXs[i] = 0x75
			} else {
				gRosterSlotBtnXs[i] = 0x10
			}
			p[i] = rosterAt(int(list[start+i]))
			if in != 0 {
				copy(names[i][:15], "\f12In Party\fd\x00\x00")
			} else {
				copy(names[i][:], p[i].Name[:]) // movedata: 16 bytes
			}
			copy(race[i][:5], str(gRaceNames[p[i].Race]))
			if v > 0x13 {
				strcpy(cls[i][:], str(gClassNames[p[i].Cls]))
				strcpy(st[i][:], hire)
			} else {
				strcpy(cls[i][:], str(gSexNames[p[i].Sex]))
				strcpy(st[i][:], str(gClassNames[p[i].Cls]))
			}
		} else {
			gRosterSlotBtnXs[i] = -1
			break
		}
	}
	drawPartyPortraits()
	for i := 0; i < 4; i++ {
		if j := int(list[start+i]); j != 0xff {
			gRosterFaceSprites[i].Img.Set(uint32(gRosterFaces[j]))
		} else {
			gRosterFaceSprites[i].Img.Set(0)
		}
	}
	sprintf(gTextBuf2, "\r\x03l\x02\v001\t035%s\v009\t035%s\v017\t035%s\v025\t035%s\v001\t136%s\v009\t136%s\v017\t136%s\v025\t136%s\v044\t035%s\v052\t035%s\v060\t035%s\v068\t035%s\v044\t136%s\v052\t136%s\v060\t136%s\v068\t136%s",
		names[0][:], race[0][:], cls[0][:], st[0][:], names[1][:], race[1][:], cls[1][:], st[1][:],
		names[2][:], race[2][:], cls[2][:], st[2][:], names[3][:], race[3][:], cls[3][:], st[3][:])
}

func pickRosterChar(mode int, list []byte, start int) int {
	var save, v, k int
	var icn Far
	defer dsFree(dsSP)

	icn = 0
	if *gCurIcons != 0 {
		icn = *gCurIcons
	}
	adrB := dsAlloca(0x14)
	b := at[Spr1](adrB)
	b.Flag.Set(-1)
	b.End.Set(-1)
	b.X.Set(0xb1)
	*gCurIcons = loadGraphic("esc.icn")
	b.Img.Set(uint32(*gCurIcons))
	sprintf(gTextBuf, "\t010\v005%s which character?\x05%04X", str(gActionVerbs[mode]), adrB)
	openTextWindow(0x32, mode*0x24+0x4c, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	*gPartyBtnXEsc = 0xeb
	*gPartyBtnYEsc = int8(mode*0x24 + 0x54)
	*gPartyBtnWEsc = 0x18
	*gPartyBtnHEsc = 0x14
	if mode == 0 {
		save = int(gRosterSlotBtnXs[0])
		gRosterSlotBtnXs[0] = -1
	}
	*gPortraitClickOn ^= uint8(mode)
	for {
		k = getInput()
		switch k {
		case 0:
			k = 1
		case 6, 0x1b:
			k = 0
		case 7, 8, 9, 10, 0x31, 0x32, 0x33, 0x34:
			if k <= 10 {
				k = k + 0x2a
			}
			if mode == 0 {
				break
			}
			v = int(list[k+start-0x31])
			if v == 0xff {
				break
			}
			goto out
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if mode == 0 && k-0xc9 < int(*gPartyCount) {
				goto out
			}
		}
		if k == 0 {
			break
		}
	}
out:
	closeWindows(1)
	freeAndNull(gCurIcons)
	*gCurIcons = icn
	*gPartyBtnXEsc, *gPartyBtnYEsc, *gPartyBtnWEsc, *gPartyBtnHEsc = 0, 0, 0, 0
	if mode == 0 {
		gRosterSlotBtnXs[0] = int16(save)
	}
	*gPortraitClickOn ^= uint8(mode)
	return k
}

func storePartyInRoster() {
	for i := 0; i < int(*gPartyCount); i++ {
		*rosterAt(int(gPartyIds[i])) = gParty[i]
	}
}

// buildInnList fills list with the characters at this inn: available hirelings first.
func buildInnList(list []byte, n *uint8) {
	memset(list, 0xff, 0x1e)
	*n = 0
	for i := 0x14; i < 0x1e; i++ {
		if r := rosterAt(i); r.Name[0] != 0 && r.Inn.Get() == int(*gInnMap) && gHirelingAvail[i] != 0 {
			list[*n] = uint8(i)
			*n++
		}
	}
	for i := 0; i < 0x14; i++ {
		if r := rosterAt(i); r.Name[0] != 0 && r.Inn.Get() == int(*gInnMap) {
			list[*n] = uint8(i)
			*n++
		}
	}
}

func innRosterScreen() {
	var i, v, pg, j int
	var n, flag uint8
	var cost uint32
	var pc *Chr
	// u8 list[30]; build_roster_page reads up to 3 past the end (the C's stack): 0xff here.
	var list [33]byte
	memset(list[:], 0xff, len(list))
	const fmtRoster = "%s\x05%04X\x05%04X\x02\x03c\v106\t013Up\t048Down\t083\f27D\fdel\t118\f27R\fdem\t153\f27C\fdreate\t188E\f27x\fdit\x01"

	updateFreeMemKb()
	*gWizeyeShown = 0
	freeAndNull(gGlobalIcons)
	freeAndNull(gGargoyleGfx)
	freeAndNull(gGrabberGfx)
	freeAndNull(gBatGfx)
	freeAndNull(gProtGemsGfx)
	freeAndNull(gProtectIcons)
	freeAndNull(gHpbarsIcons)
	freeAndNull(gRestoreIcons)
	freeAndNull(gCharpowSprites)
	updateFreeMemKb()
	*gGlobalIcons = loadGraphic("global.icn")
	*gGargoyleGfx = loadGraphic("gargoyle.brd")
	*gGrabberGfx = loadGraphic("grabber.brd")
	*gBatGfx = loadGraphic("bat.brd")
	*gProtGemsGfx = loadGraphic("fecp.brd")
	*gProtectIcons = loadGraphic("protect.icn")
	*gHpbarsIcons = loadGraphic("hpbars.icn")
	*gRestoreIcons = loadGraphic("restore.icn")
	*gCharpowSprites = loadGraphic("charpow.icn")
	updateFreeMemKb()
	addUiGadgetSprites()
	*gDimLevel = 0
	flag = 0
	sprintf(gTextBuf, "\x03d\x05%04X", adrBorderSprites)
	*gRosterFar = memAlloc(0x2382)
	memset(ds[adrRosterFaces:], 0, 0x78)
	*gInnIconsF = loadGraphic("inn.icn")
	gRosterFrameSprite.Img.Set(uint32(*gInnIconsF))
top:
	if uint8(*gSaveSlot) == 0xff {
		quitToDos()
	}
	loadPartyData()
	for i = 0; i < int(*gPartyCount); i++ {
		gParty[i] = *rosterAt(int(gPartyIds[i]))
	}
	if *gMinutes < 0x12c {
		*gNewDayPending = 1
	}
	if *gMapNum != 0 {
		*gGameMode = 0
		*gPortraitClickOn = 1
		goto end
	}
	if flag == 0 {
		playMusic(cs("towninn.m"), 0x50)
		flag = 1
	}
	if gRosterFaces[0] == 0 {
		loadFaces(0x1e)
	}
	for i = 0; i < int(*gPartyCount); i++ {
		gPartyFacesF[i] = gRosterFaces[gPartyIds[i]]
	}
	*gGameMode = 1
	buildInnList(list[:], &n)
	gTextBuf2[0] = 0
	pg = 0
	*gPortraitClickOn = 1
	*gBtnXs = adrPartyBtnX
	*gBtnYs = adrPartyBtnY
	*gBtnWs = adrPartyBtnW
	*gBtnHs = adrPartyBtnH
	drawBackdrop()
	buildRosterPage(pg, list[:])
	sprintf(gTextBuf, fmtRoster, gTextBuf2, adrRosterFaceSprites, adrRosterFrameSprite)
	openTextWindow(8, 8, 0xd8, 0x84, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	fadeScreen(1, 4)
	goto loop
redraw:
	for i = 0; i < int(*gPartyCount); i++ {
		gPartyFacesF[i] = gRosterFaces[gPartyIds[i]]
	}
	gTextBuf2[0] = 0
	buildRosterPage(pg, list[:])
	sprintf(gTextBuf, fmtRoster, gTextBuf2, adrRosterFaceSprites, adrRosterFrameSprite)
	printWindowText(gTextBuf)
loop:
	for {
		i = getInput()
		switch i {
		case 0, 0xf2:
			if pg != 0 {
				pg -= 4
				goto redraw
			}
		case 1, 0xf3:
			if int(n)-5 >= pg {
				pg += 4
				goto redraw
			}
		case 2, 'D':
			*gPortraitClickOn = 0
			if n == 0 {
				goto delOut
			}
			i = pickRosterChar(1, list[:], pg)
			if i == 0 {
				goto delOut
			}
			v = int(list[i+pg-0x31])
			pc = rosterAt(v)
			if v < 0x14 {
				strncpy(gTmpText, pc.Name[:], 0x10)
				for i = 0; i < 0x12; i++ {
					if isQuestItem(int(pc.ItemID[i])) != 0 {
						she := ""
						if pc.Sex != 0 {
							she = "s"
						}
						sprintf(gTextBuf2, "%s cannot be deleted because %she is holding special items!", gTmpText, she)
						showMessage(gTextBuf2)
						goto delOut
					}
				}
				sprintf(gTextBuf2, "Are you sure you want to delete %s the %s?", gTmpText, str(gClassNames[pc.Cls]))
				if askYesNo(gTextBuf2, 0) == 0 {
					goto delOut
				}
				*pc = Chr{}
				buildInnList(list[:], &n)
				for j = 0; j < int(*gPartyCount); j++ {
					if int(gPartyIds[j]) == v {
						i = j + 0xc9
						gPartyIds[i-0xc9] = 0xff
						compactParty()
						*gPartyCount--
					}
				}
				storePartyInRoster()
				writeSaveGame()
				goto redraw
			}
			showMessage(cs("You cannot delete a Hireling!"))
		delOut:
			*gPortraitClickOn = 1
		case 3, 'R':
			if *gPartyCount == 0 {
				break
			}
			i = pickRosterChar(0, list[:], pg)
			if i == 0 {
				break
			}
			storePartyInRoster()
			gPartyIds[i-0xc9] = 0xff
			compactParty()
			*gPartyCount--
			goto redraw
		case 4, 'C':
			if n < 0x14 {
				fadeScreen(0, 4)
				closeWindows(1)
				*gPortraitClickOn = 0
				createCharacters()
				storePartyInRoster()
				writeSaveGame()
				*gPortraitClickOn = 1
				goto top
			}
			showMessage(cs("Your Roster is full!"))
		case 5, 0x1b, ' ', 'E', 'X':
			if *gPartyCount == 0 {
				break
			}
			if countPartyNonHirelings() == 0 {
				showMessage(cs("\v007You must have at least one character accompany the hirelings!"))
				break
			}
			*gPartyCountSaved = *gPartyCount
			*gMapNum = *gInnMap
			storePartyInRoster()
			writeSaveGame()
			goto end
		case 0x31, 0x32, 0x33, 0x34, 7, 8, 9, 10:
			if i >= 0x31 {
				i -= 0x2a
			}
			if int(n) <= i+pg-7 {
				break
			}
			for j = 0; j < int(*gPartyCount); j++ {
				if list[i+pg-7] == gPartyIds[j] {
					continue loop
				}
			}
			if *gPartyCount != 8 {
				if list[i+pg-7] > 0x13 {
					if *gPartyCount == 7 && partyHasOnlyHirelings() != 0 {
						showMessage(cs("\v007You must have at least one character accompany the hirelings!"))
						break
					}
					r := rosterAt(int(list[i+pg-7]))
					cost = uint32(getCharLevel(r))
					strncpy(gTmpText, bytesOf(r), 0x14)
					sprintf(gTextBuf, "\x03c%s speaks: My cost is %lu per day.  Do you accept?", gTmpText, cost*cost*0x14)
					if askYesNo(gTextBuf, 0) == 0 {
						break
					}
					if payCost(0, uint32(r.Level)*100, 0) == 0 {
						break
					}
				} else if countPartyNonHirelings() > 5 {
					showMessage(cs("\v007You can only have six characters but you may add more hirelings!"))
					break
				}
				gPartyIds[*gPartyCount] = list[i+pg-7]
				// g_party[g_party_count++] = g_roster[g_party_ids[g_party_count]]: the id just stored
				gParty[*gPartyCount] = *rosterAt(int(gPartyIds[*gPartyCount]))
				*gPartyCount++
				goto redraw
			}
			playSfx(0x15)
		case 0x11:
			quitToDos()
			fallthrough
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if i -= 0xc9; i < int(*gPartyCount) {
				charScreen(i)
			}
		}
	}
end:
	for i = 0; i < 0x1e; i++ {
		freeAndNull(&gRosterFaces[i])
	}
	memset(ds[adrPartyFaces:], 0, 0x20)
	freeAndNull(gInnIconsF)
	freeAndNull(gRosterFar)
	if *gGameMode != 0 {
		for *gDimLevel = 0; *gDimLevel < 5; *gDimLevel++ {
			sprintf(gTextBuf, "\x03d\x03m%u\x05%04X", *gDimLevel, adrBorderSprites)
			printWindowText(gTextBuf)
		}
	}
	closeWindows(1)
}
