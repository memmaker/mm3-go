// Port of GUILD.C (overlay 182a): Corak's notes and the quest list, the quickfight options, spell
// costs, the guild's spell list (buy) and the known-spell list (select), and character creation
// (roll, exchange stats, pick class and alignment, name).
//
// guild's struct chr maps onto Chr by offset: stats[k][0] is byte 0x14+2k, res[k][0] 0x107+2k.
// Its locals passed to the window text as "\x05%p" draw lists live on the ds stack (dsAlloca).
package mm3

func showCoraksNotes() {
	var i, last, flag, idx int
	var n uint16
	var buf [0x13]uint8

	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	memset(buf[:], 0x13, 0x13)
	updateFreeMemKb()
	if *gFreeMemKb < 0x1e && *gOutdoorObjGfx != 0 {
		freeMazeGfx()
		flag = 1
	} else {
		flag = 0
	}
	if *gMapNum == 0x69 {
		idx = 0x40
	} else {
		idx = int(*gMapNum) - 1
	}
	loadTextTable(cs("corak.bin"))
	strncpy(gTextBuf2, gTextTable.Add(int(gTextOffsets[idx])).B(), 0x1f4)
	sprintf(gTextBuf, "\x03cCorak's Notes on %s\x03l\n\n%s\x0b120\t000\x03cPress a key\x03l", str(gMapNamesFrom1[idx]), gTextBuf2)
	openTextWindow(0, 0, 0x140, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	freeAndNull(gTextTable)
	loadTextTable(cs("quest.bin"))
	getInput()
	for i = 0; i < 0x13; i++ {
		switch {
		case i <= 12:
			if testBit(gGameFlags, i+0x9d) != 0 {
				buf[i] = uint8(i)
			}
		case i == 13:
			if testBit(gGameFlags, 0xaa) == 0 {
				buf[i] = uint8(i)
			}
		case i <= 18:
			if testBit(gGameFlags, int(gQuestFlagPairs[0][i-14])) != 0 && testBit(gGameFlags, int(gQuestFlagPairs[1][i-14])) == 0 {
				buf[i] = uint8(i)
			}
		}
	}
	printWindowText(str(*gStrQuestsHeader))
	i, n, last = 0, 0, 0
	for ; i < 0x13; i++ {
		if buf[gQuestOrder[i]] != 0x13 {
			n++
		}
		strncpy(gTextBuf2, gTextTable.Add(int(gTextOffsets[buf[gQuestOrder[i]]])).B(), 0x12c)
		sprintf(gTextBuf, "\t000\x0b%03d%s", int((n-1)%5)*0x16+10, gTextBuf2)
		printWindowText(gTextBuf)
		if (n == 5 || n == 10 || n == 15) && int(n) != last {
			printWindowText(cs("\x0b120\t000\x03cPress a key\x03l"))
			getInput()
			printWindowText(str(*gStrQuestsHeader))
			last = int(n)
		}
	}
	printWindowText(cs("\x0b120\t000\x03cPress a key\x03l"))
	printWindowText(cs("\x01"))
	freeAndNull(gTextTable)
	if flag != 0 {
		loadMazeGfx(int(*gMapNum) - 1)
	}
	getInput()
	closeWindows(1)
	flushInput()
	popButtons()
	*gGameMode = save
}

func quickfightOptions(c *Chr) {
	drawn := 0
	var key int

	pushButtons()
	pushIcons("train.icn")
	at[Spr](adrTrainSprites).Img.Set(uint32(*gCurIcons))
	*gBtnXs = adrTrainBtnXs
	*gBtnYs = adrTrainBtnYs
	*gBtnWs = adrTrainBtnWs
	*gBtnHs = adrTrainBtnHs
redraw:
	sprintf(gTextBuf, "\r\x05%04X\x03c\x0b000\t000QuickFight Options\n\n%s\x03l\n\nCurrent\x03r\n\t077%s\x02\x03c\x0b122\t021\x0c27N\x0cdext\t060Exit\x01",
		adrTrainSprites, bytesOf(c), str(gQuickfightNames[c.Quickfight]))
	if drawn != 0 {
		printWindowText(gTextBuf)
	} else {
		openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		drawn = 1
	}
	for {
		key = getInput()
		switch key {
		case 0, 0xd, 0x1b:
			key = 0x1b
		case 1, 0x4e:
			c.Quickfight = (c.Quickfight + 1) % 4
			goto redraw
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			key -= 0xc9
			if int(*gCombatPartyN) > key {
				c = &gParty[key]
				highlightPortrait(key)
			}
			goto redraw
		}
		if key == 0x1b {
			break
		}
	}
	closeWindows(1)
	popIcons()
	popButtons()
}

func spellSpCost(s, lvl int) int {
	v := int(gSpellSpCosts[s])
	if v < 0 {
		return int(int16(v * lvl * -1))
	}
	return v
}

func spellGoldCost(s, sh int) int {
	v := int(gSpellSpCosts[s])
	if v < 0 {
		return int(int16(v * -500 << sh))
	}
	return int(int16(v * 100 << sh))
}

// spellRow is g_spell_rows + k * 0x32.
func spellRow(k int) []byte { return ds[int(*gSpellRows)+k*0x32:] }

func buildSpellList(c *Chr, mode int) int {
	var i, j, lim, off, lvl, flag, kind, n int

	lvl = getCharLevel(c)
	switch mode & 0x7f {
	case 0:
		if int(gGuildMaxSpellLvl[*gMapNum-1]) < lvl {
			lim = int(gGuildMaxSpellLvl[*gMapNum-1])
		} else {
			lim = lvl
		}
		switch c.Cls {
		case 1:
			flag = 1
			fallthrough
		case 3:
			kind = 0
		case 2:
			flag = 1
			fallthrough
		case 4:
			kind = 1
		case 9:
			flag = 1
			fallthrough
		case 8:
			kind = 2
			if lim > 15 {
				lim = 15
			}
		default:
			goto check
		}
		for i, off = 0, 0; i < lim; i++ {
			for j = 0; j < int(gSpellsPerLevel[kind][i]); j++ {
				if c.Spells[off+j] == 0 || mode&0x80 != 0 {
					sp := int(gClassSpells[kind][off+j])
					sprintf(spellRow(n), "\x03l%s\x03r\t148%u", str(gSpellNames[sp]), spellGoldCost(sp, flag))
					gSpellRowIdx[n] = byte(off + j)
					n++
				}
			}
			off += int(gSpellsPerLevel[kind][i])
		}
		if n == 0 {
			if lvl <= lim {
				sprintf(spellRow(0), "Come back when you're\n\t010more experienced.")
			} else {
				sprintf(spellRow(0), "You have learned all we can\n\t010teach you.")
			}
		}
	check:
		if calcMaxSp(c) == 0 {
			sprintf(spellRow(0), "Not a spell caster...")
		}
	case 1:
		lim = 0x11
		switch c.Cls {
		case 1, 3:
			kind = 0
		case 2, 4:
			kind = 1
		case 8, 9:
			kind = 2
			lim = 0xf
		}
		if calcMaxSp(c) != 0 {
			for i, off = 0, 0; i < lim; i++ {
				for j = 0; j < int(gSpellsPerLevel[kind][i]); j++ {
					if c.Spells[off+j] != 0 {
						sp := int(gClassSpells[kind][off+j])
						sprintf(spellRow(n), "\x03l%s\x03r\t148%u/%u", str(gSpellNames[sp]), spellSpCost(sp, lvl), gSpellGemCosts[sp])
						gSpellRowIdx[n] = byte(off + j)
						n++
					}
				}
				off += int(gSpellsPerLevel[kind][i])
			}
			if n == 0 {
				sprintf(spellRow(0), "You don't know any spells...")
			}
		} else {
			sprintf(spellRow(0), "Not a spell caster...")
		}
	}
	return n
}

func spellListDialog(c *Chr, mode int) *Chr {
	var key, cnt, idx, flag, kind, spell, drawn, wait, omode, cost, top, cur int
	var sel [0x28]uint8
	var gold uint32

	pushButtons()
	pushIcons("scroll.icn")
	top, drawn, wait = 0, 0, 0
	*gSpellRows = adrDrawList
	gScrollSpr.Img.Set(uint32(*gCurIcons))
	*gBtnXs = adrScrollBtnX
	*gBtnYs = adrScrollBtnY
	*gBtnWs = adrScrollBtnW
	*gBtnHs = adrScrollBtnH
	omode = mode
	mode &= 0x7f
	if mode != 0 {
		gScrollCastSprites.Img.Set(uint32(*gCurIcons))
	} else {
		gScrollCastSprites.Img.Set(0)
	}
	memset(sel[:], 0xc, 0x28)
	cur = -1
	goto again
recount:
	cnt = buildSpellList(c, omode)
	if top+10 > cnt {
		if top = cnt - 10; top < 1 {
			top = 0
		}
	}
again:
	memset(gSpellRowIdx, 0, 0x24)
	memset(gDrawListB, 0, 0xfa0)
	cnt = buildSpellList(c, omode)
	if cnt != 0 {
		strcpy(gTmpText, cs("\x02\x03l\x0b015\t0011\n2\n3\n4\n5\n6\n7\n8\n9\n0"))
	} else {
		strcpy(gTmpText, cs(""))
	}
	{
		unit := "Gold"
		gold = *gPartyGold
		if mode != 0 {
			unit = "Spell Pts"
			gold = uint32(int32(c.SP.Get()))
		}
		args := []any{gTmpText, adrScrollSpr, adrScrollCastSprites, bytesOf(c)}
		for r := 0; r < 10; r++ {
			args = append(args, sel[top+r], spellRow(top+r))
		}
		args = append(args, unit, gold)
		sprintf(gTextBuf, "\r%s\x05%04X\x05%04X\x02\x03c\t000\x0b002Available Spells for %s\x03l\x0b015\t011\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\n\t010\x0c%2u%s\x0cd\x03l\t004\x0b110%s - %lu\x01", args...)
	}
	if drawn != 0 {
		printWindowText(gTextBuf)
	} else {
		openTextWindow(0x1b, 6, 0xb4, 0x88, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		drawn = 1
		for getMouseState(gMouseX, gMouseY) != 0 {
		}
	}
	if wait != 0 {
		for {
			if wait = getMouseState(gMouseX, gMouseY); wait == 0 {
				break
			}
		}
	}
	for {
		key = getInput()
		switch key {
		case 0, 0xf2:
			if top--; top < 0 {
				top = 0
			}
			goto again
		case 1, 0xf3:
			if top++; top > cnt-10 {
				if top = cnt - 10; top < 1 {
					top = 0
				}
			}
			goto again
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
			if key >= '1' && key <= '9' {
				key -= 0x2f
			} else if key == '0' {
				key = 0xb
			}
			if cnt == 0 {
				break
			}
			if key--; key > cnt {
				break
			}
			flag = 0
			switch c.Cls {
			case 1:
				flag = 1
				fallthrough
			case 3:
				kind = 0
			case 2:
				flag = 1
				fallthrough
			case 4:
				kind = 1
			case 9:
				flag = 1
				fallthrough
			case 8:
				kind = 2
			}
			key--
			idx = int(int8(gSpellRowIdx[key+top]))
			spell = int(gClassSpells[kind][idx])
			if mode != 0 {
				if key+top == cur {
					sel[cur] = 0xc
					cur = -1
				} else {
					memset(sel[:], 0xc, 0x28)
					cur = key + top
					sel[cur] = 0x12
				}
				wait = 1
				goto again
			}
			cost = spellGoldCost(spell, flag)
			strncpy(gTmpText, gTextTable.Add(int(gTextOffsets[gClassSpells[kind][idx]])).B(), 0xc7)
			if omode&0x80 != 0 {
				sprintf(gTextBuf2, "\x03c\x0c12%s\x0cd\x03l\n\n%s\x03c\t000\x0b100Press a Key!", str(gSpellNames[gClassSpells[kind][idx]]), gTmpText)
			} else {
				sprintf(gTextBuf2, "%s  Do you wish to purchase \x0c12%s\x0cd for %u?", gTmpText, str(gSpellNames[gClassSpells[kind][idx]]), cost)
			}
			if askYesNo(gTextBuf2, omode+1) == 0 {
				break
			}
			// guild declares pay_cost(int, int, int, int): the long amount is cost with a zero high word
			if payCost(0, uint32(uint16(cost)), 0) == 0 {
				break
			}
			c.Spells[idx]++
			goto recount
		case 0xc, 0x1b:
			cur = -1
			key = 0x1b
		case 0xd, 0x53, 0xde, 0xe, 0xdf, 0xf:
			if key == 0xf || (key == 0xd || key == 0x53) && *gMouseButtons == 0 { // esc:
				if cur != -1 {
					key = 0x1b
				}
				break
			}
			if key == 0xe || key == 0xdf {
				top += 10
			} else if top -= 10; top < 0 {
				top = 0
			}
			for getMouseState(gMouseX, gMouseY) != 0 { // wt:
			}
			goto recount
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if mode == 0 {
				break
			}
			if *gGameMode == 2 {
				break
			}
			key -= 0xc9
			if int(*gPartyCount) > key {
				memset(sel[:], 0xc, 0x28)
				cur = -1
				c = &gParty[key]
				*gCastChar = int8(key)
				highlightPortrait(key)
			}
			goto recount
		}
		if key == 0x1b {
			break
		}
	}
	closeWindows(1)
	popIcons()
	popButtons()
	if mode != 0 && cur != -1 {
		c.ReadySpell = gClassSpells[kind][int8(gSpellRowIdx[cur])]
	}
	return c
}

func pickStatExchange(n int) int {
	var key, sel, done int
	defer dsFree(dsSP)

	done = 0
	gCreateHotXs[7] = -1
	gCreateBtnXs[0] = 0x85
	gCreateBtnXs[1] = -1
	gCreateBtnYs[0] = 0x49
	adrS := dsAlloca(0x14)
	s := at[Spr1](adrS)
	*gCurIcons = loadGraphic("esc.icn")
	s.Img.Set(uint32(*gCurIcons))
	s.Flag.Set(-1)
	s.End.Set(-1)
	s.X.Set(0x76)
	s.Y.Set(0x3a)
	sprintf(gTextBuf2, "Exchange %s with ...\x05%04X", str(arr[uint16](adrStrGood)[n]), adrS)
	openTextWindow(7, 7, 0x9e, 0x5e, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
	for done == 0 {
		chk := false
		key = getInput()
		switch key {
		case 0, 0x1b:
			sel = 0
			done = 1
		case 'M', 'I', 'P', 'E', 'S', 'A', 'L':
			sel = map[int]int{'M': 0, 'I': 1, 'P': 2, 'E': 3, 'S': 4, 'A': 5, 'L': 6}[key]
			chk = true
		case 'V':
			sel = createHitButton()
			t := sel
			sel--
			chk = t != 0 && sel >= 0 && sel <= 6
		}
		if chk && n-1 != sel {
			sel++
			done = 1
		}
	}
	closeWindows(1)
	gCreateHotXs[7] = 0xe3
	gCreateBtnXs[0], gCreateBtnXs[1] = 0x84, 0x84
	gCreateBtnYs[0] = 0x62
	freeAndNull(gCurIcons)
	return sel
}

func bounceDice() {
	for i := 0; i < 3; i++ {
		k := i * 4
		gDiceFrame[i] = byte(int8(gDiceFrame[i]+1) % 7)
		fr := int(int8(gDiceFrame[i]))
		gDicePos[k] += int16(int8(gDiceDx[i]))
		gDicePos[k+1] += int16(int8(gDiceDy[i]))
		if gDicePos[k] < 0xd {
			gDicePos[k] = 0xd
			gDiceDx[i] = byte(-int8(gDiceDx[i]))
		}
		if int(gDicePos[k]) > 0xa3-int(gDieWidths[fr]) {
			gDicePos[k] = int16(0xa3 - int(gDieWidths[fr]))
			gDiceDx[i] = byte(-int8(gDiceDx[i]))
		}
		if gDicePos[k+1] < 0xc {
			gDicePos[k+1] = 0xc
			gDiceDy[i] = byte(-int8(gDiceDy[i]))
		}
		if int(gDicePos[k+1])+int(gDieHeights[fr]) > 0x5d {
			gDicePos[k+1] = int16(0x5d - int(gDieHeights[fr]))
			gDiceDy[i] = byte(-int8(gDiceDy[i]))
		}
		gDicePos[k+3] = int16(fr)
	}
	sprintf(gTextBuf2, "\x03q\x05%04X", adrDiceSprites)
	printWindowText(gTextBuf2)
	idleFrames(3)
}

var (
	createNeedClassText = cs("\x0b006\x03cSelect a Class and Alignment before saving.\x03l")
	createNameText      = cs("\x03cEnter a Name for this Character\n\n")
)

func finishCreateChar(cls, race, align, sex uint8, st []byte, c *Chr) int {
	var i, k, m, n int
	var name [0x14]byte

	if cls == 0xff || align == 0xff {
		pushButtons()
		setAnykeyButtons()
		openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, farOf(createNeedClassText))
		getInput()
		closeWindows(1)
		flushInput()
		popButtons()
		return 0
	}
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, farOf(createNameText))
	n = inputText(name[:], 10, 200)
	closeWindows(1)
	if n == 0 {
		return 0
	}
	b := bytesOf(c)
	memset(b, 0, 0x12f)
	for i = 0; i < n; i++ {
		c.Name[i] = name[i]
	}
	c.Inn.Set(int(*gInnMap))
	c.Sex = sex
	c.Race = race
	c.Cls = cls
	c.Align = align
	for k = 0; k < 7; k++ {
		b[0x14+2*k] = st[k]
	}
	c.Level = 1
	c.MagicRes = gRaceMagicRes[race]
	c.FireRes = gRaceFireRes[race]
	c.ElecRes = gRaceElecRes[race]
	c.ColdRes = gRaceColdRes[race]
	c.EnergyRes = gRaceEnergyRes[race]
	c.PoisonRes = gRacePoisonRes[race]
	c.BirthYear.Set(int(*gYear - 0x12))
	c.BirthDay = *gDay
	switch cls {
	case 3, 4, 8:
		c.Caster = 1
		for n = 0; n < 4; n++ {
			c.Spells[n] = 1
		}
		c.ReadySpell = 0
	case 1, 2, 9:
		c.Caster, c.Spells[0] = 1, 1
		c.ReadySpell = 0
	default:
		c.ReadySpell = 0xff
	}
	if k = int(gClassStartSkill[cls]); k != 0xff {
		c.Skill[k] = 1
	}
	if k = int(gRaceStartSkill[race]); k != 0xff {
		c.Skill[k] = 1
	}
	if m = int(gClassStartSpell[cls]); m != 0xff {
		c.Spells[m] = 1
	}
	c.HP.Set(calcMaxHp(c))
	c.SP.Set(calcMaxSp(c))
	return 1
}

func calcAllowedClasses(st, ok []byte) {
	s := func(k int) int { return int(int8(st[k])) }
	for i := 0; i <= 9; i++ {
		switch i {
		case 0:
			ok[i] = byte(b2i(s(0) >= 15))
		case 1:
			ok[i] = byte(b2i(s(0) >= 13 && s(2) >= 13 && s(3) >= 13))
		case 2:
			ok[i] = byte(b2i(s(1) >= 13 && s(5) >= 13))
		case 3:
			ok[i] = byte(b2i(s(2) >= 13))
		case 4:
			ok[i] = byte(b2i(s(1) >= 13))
		case 5:
			ok[i] = byte(b2i(s(6) >= 13))
		case 6:
			ok[i] = byte(b2i(s(4) >= 13 && s(5) >= 13))
		case 7:
			ok[i] = byte(b2i(s(3) >= 15))
		case 8:
			ok[i] = byte(b2i(s(1) >= 15 && s(2) >= 15))
		case 9:
			ok[i] = byte(b2i(s(1) >= 12 && s(2) >= 12 && s(3) >= 12 && s(4) >= 12))
		}
	}
}

func rollStats(st, ok []byte) {
	for i := 0; i < 7; i++ {
		st[i] = 0
	}
	for j := 0; j < 3; j++ {
		for i := 0; i < 7; i++ {
			st[i] += byte(uint8(randomRange(10, 0x4f)) / 10)
		}
	}
	calcAllowedClasses(st, ok)
}

func formatCreateStats(st, ok, out []byte, race, sex uint8) int {
	var a [10]int
	best := uint8(0xff)
	for i := range a {
		a[i] = 2
	}
	for i := 0; i <= 9; i++ {
		if ok[i] != 0 {
			if int(best) > i {
				best = uint8(i)
			}
			a[i] = 0
		}
	}
	args := []any{adrCreateButtonSprites, adrCreateAlignSprites, adrCreateStatboxSprites, str(gRaceNames[race]), str(gSexNames[sex])}
	for k := 0; k < 7; k++ {
		args = append(args, int8(st[k]))
	}
	for _, v := range a {
		args = append(args, v)
	}
	sprintf(out, "\x05%04X\x05%04X\x05%04X\t022\x0b142Race\t055: %s\n\t022Sex\t055: %s\n\t022Class\t055:\n\t022Align\t055:\x03r\t215\x0b031%d\t215\x0b055%d\t215\x0b079%d\t215\x0b103%d\t215\x0b127%d\t215\x0b151%d\t215\x0b175%d\x03l\t242\x0b020\x0c%2dKnight\t242\x0b031\x0c%2dPaladin\t242\x0b042\x0c%2dArcher\t242\x0b053\x0c%2dCleric\t242\x0b064\x0c%2dSorcerer\t242\x0b075\x0c%2dRobber\t242\x0b086\x0c%2dNinja\t242\x0b097\x0c%2dBarbarian\t242\x0b108\x0c%2dDruid\t242\x0b119\x0c%2dRanger\x0c00\t242\x0b148Good\t242\x0b160Neutral\t242\x0b172Evil", args...)
	return int(best)
}

func createMarkSelection(n, flag int) {
	sprintf(gTmpText, "\x05%04X", adrCreateMarkerSprites)
	printWindowText(gTmpText)
	y := 0x15 + 0xb*n
	if flag != 0 {
		y = 0x95 + 0xc*n
	}
	drawIcon(*gCurIcons, 0x3f, 0xdc, y, 0)
}

func createCharacters() {
	var b, flip, i, key, btn, save, sel int
	var cls, race, align, sex, count, cur, inited uint8
	var st [7]byte
	var ok [10]byte
	var img Far
	var list [0x1e]uint8
	var buf [0x190]byte
	defer dsFree(dsSP)

	adrSpr1 := dsAlloca(0x14)
	adrSpr2 := dsAlloca(0x14)
	spr1, spr2 := at[Spr1](adrSpr1), at[Spr1](adrSpr2)
	btn = 0
	*gDiceVga = loadGraphic("dice.vga")
	*gDiceSpritesImg = *gDiceVga
	gDicePos[0] = 0x14
	gDicePos[1] = 0x11
	gDicePos[4] = 0x70
	gDicePos[5] = 0x23
	gDicePos[8] = 0x3d
	gDicePos[9] = 0x32
	inited, gDiceFrame[0] = 0, 0
	gDiceFrame[1] = 2
	gDiceFrame[2] = 4
	gDiceDx[0] = 10
	gDiceDy[0] = 0xf6 // -10
	gDiceDx[1] = 0xf6
	gDiceDy[1] = 0xf6
	gDiceDx[2] = 0xf6
	gDiceDy[2] = 10
	save = int(*gGameMode)
	*gGameMode = 4
	*gBtnXs = adrCreateBtnXs
	*gBtnYs = adrCreateBtnYs
	*gBtnWs = adrCreateBtnWs
	*gBtnHs = adrCreateBtnHs
	clearScreen()
	showRawPicture("create.raw")
	setMouseCursor(0, uint32(*gMouseCursor))
	img = loadGraphic("create.icn")
	*gCurIcons = img
	spr1.Flag.Set(-1)
	spr1.End.Set(-1)
	spr2.Flag.Set(-1)
	spr2.End.Set(-1)
	spr2.X.Set(0x3e)
	for _, s := range []*Spr{&spr1.Spr, &spr2.Spr, gCreateAlignSprites, gCreateMarkerSprites, gCreateStatboxSprites, &gCreatePortraitSprites[1], gCreateButtonSprites} {
		s.Img.Set(uint32(img))
	}

	// the class (s 0..9) and alignment (s 0..2) picks of keys Enter and V
	pickClass := func(s int) {
		spr1.X.Set(int(gCreateHotXs[s+7]))
		spr1.Y.Set(int(gCreateHotYs[s+7]))
		spr1.Frame.Set((s+7)*2 + 11)
		spr2.Y.Set(0xa2)
		spr2.Frame.Set(0x3b)
		if cls == 0xff {
			strcpy(gTextBuf2, cs(""))
		} else {
			sprintf(gTextBuf2, "\t242\x0b%3u%s", int(cls)*0xb+0x14, str(gClassNames[cls]))
		}
		cls = uint8(s)
		sprintf(gTextBuf, "%s\x05%04X\t062\x0b162%s\x05%04X\x05%04X\x0c12\t242\x0b%3u%s\x0c00", gTextBuf2, adrSpr2, str(gClassNames[s]), adrCreateButtonSprites, adrSpr1, s*0xb+0x14, str(gClassNames[s]))
	}
	pickAlign := func(s int) {
		spr1.X.Set(int(gCreateHotXs[s+17]))
		spr1.Y.Set(int(gCreateHotYs[s+17]))
		spr1.Frame.Set((s+17)*2 + 11)
		spr2.Y.Set(0xac)
		spr2.Frame.Set(0x3c)
		if align == 0xff {
			strcpy(gTextBuf2, cs(""))
		} else {
			sprintf(gTextBuf2, "\t242\x0b%3u%s", int(align)*0xc+0x94, str(gAlignNames[align]))
		}
		align = uint8(s)
		sprintf(gTextBuf, "%s\x05%04X\t062\x0b172%s\x05%04X\x05%04X\x0c12\t242\x0b%3u%s\x0c00", gTextBuf2, adrSpr2, str(gAlignNames[s]), adrCreateAlignSprites, adrSpr1, s*0xc+0x94, str(gAlignNames[s]))
	}

top:
	cls, align = 0xff, 0xff
	memset(list[:], 0xff, 0x1e)
	flip, count, cur = 0, 0, 0
	for i = 0; i < 0x14; i++ {
		if rosterAt(i).Name[0] == 0 {
			list[count] = uint8(i)
			count++
		}
	}
	if count == 0 {
		goto done
	}
	gCreatePortraitSprites[0].Img.Set(uint32(gRosterFaces[list[cur]]))
	race = list[cur] / 4
	sex = list[cur] & 1
	rollStats(st[:], ok[:])
	sel = formatCreateStats(st[:], ok[:], buf[:], race, sex)
	sprintf(gTextBuf, "\x05%04X\x03c\x02\t144\x0b119\x0c27R\x0c00oll\t144\x0b149\x0c27C\x0c00reate\t144\x0b179\x0c27ESC\x0c00\x03l\x01\t195\x0b021\x0c27M\x0c00gt\t195\x0b045\x0c27I\x0c00nt\t195\x0b069\x0c27P\x0c00er\t195\x0b093\x0c27E\x0c00nd\t195\x0b116\x0c27S\x0c00pd\t195\x0b140\x0c27A\x0c00cy\t195\x0b164\x0c27L\x0c00ck%s", adrCreatePortraitSprites, buf[:])
	printWindowText(gTextBuf)
	if sel != 0xff {
		createMarkSelection(sel, flip)
	}
	bounceDice()
	if inited == 0 {
		fadeScreen(1, 4)
		inited = 1
	}
	for {
		const (
			lNone = iota
			lTab
			lShow
			lSearch
			lPick
			lStat
		)
		lbl := lNone
		key = getInput()
		playSfx(0x14)
		switch key {
		case 9:
			lbl = lTab
		case 0xf0:
			if flip == 0 {
				for key = sel - 1; key >= 0; key-- {
					if ok[key] != 0 {
						sel = key
						lbl = lShow
						break
					}
				}
				break
			}
			if sel--; sel < 0 {
				sel = 0
			}
			lbl = lShow
		case 0xf1:
			if flip == 0 {
				lbl = lSearch
				break
			}
			if sel++; sel > 2 {
				sel = 2
			}
			lbl = lShow
		case 0xd:
			if flip == 0 && sel == 0xff {
				break
			}
			if flip == 0 {
				pickClass(sel)
			} else {
				pickAlign(sel)
			}
			printWindowText(gTextBuf)
			lbl = lTab
		case 0, 0x20, 'R':
			rollStats(st[:], ok[:])
			align, cls = 0xff, 0xff
			flip = 0
			sel = formatCreateStats(st[:], ok[:], buf[:], race, sex)
			printWindowText(buf[:])
			if sel != 0xff {
				createMarkSelection(sel, flip)
			}
		case 1, 'C':
			*gGameMode = 0xff
			i = finishCreateChar(cls, race, align, sex, st[:], rosterAt(int(list[cur])))
			*gGameMode = 4
			if i != 0 {
				goto top
			}
		case 3, 0xf2:
			if cur == 0 {
				break
			}
			cur--
			lbl = lPick
		case 4, 0xf3:
			if cur++; cur == count {
				cur--
				break
			}
			lbl = lPick
		case 'M', 'I', 'P', 'E', 'S', 'A', 'L':
			b = map[int]int{'M': 0, 'I': 1, 'P': 2, 'E': 3, 'S': 4, 'A': 5, 'L': 6}[key]
			lbl = lStat
		case 'V':
			b = createHitButton()
			if b == 0 {
				break
			}
			b--
			switch {
			case b <= 6:
				lbl = lStat
			case b <= 16:
				if ok[b-7] != 0 {
					b -= 7
					pickClass(b)
					printWindowText(gTextBuf)
				}
				waitMouseRelease()
				if align != 0xff {
					sel = int(align)
				} else {
					sel = 0
				}
				flip = 1
				createMarkSelection(sel, flip)
			case b <= 19:
				b -= 0x11
				pickAlign(b)
				printWindowText(gTextBuf)
				waitMouseRelease()
				flip = 0
				if cls == 0xff {
					lbl = lSearch
				} else {
					sel = int(cls)
					createMarkSelection(sel, flip)
				}
			}
		case 2, 0x1b:
			goto done
		default:
			playSfx(0x15)
		}
		for lbl != lNone {
			switch lbl {
			case lTab:
				lbl = lNone
				if flip == 0 && sel == 0xff {
					break
				}
				flip ^= 1
				if flip == 0 {
					if sel = int(cls); sel == 0xff {
						lbl = lSearch
						break
					}
				} else if align != 0xff {
					sel = int(align)
				} else {
					sel = 0
				}
				lbl = lShow
			case lShow:
				createMarkSelection(sel, flip)
				lbl = lNone
			case lSearch:
				lbl = lNone
				for key = int(int8(sel)) + 1; key <= 9; key++ {
					if ok[key] != 0 {
						sel = key
						lbl = lShow
						break
					}
				}
			case lPick:
				align, cls = 0xff, 0xff
				gCreatePortraitSprites[0].Img.Set(uint32(gRosterFaces[list[cur]]))
				race = list[cur] / 4
				sex = list[cur] & 1
				sel = formatCreateStats(st[:], ok[:], buf[:], race, sex)
				sprintf(gTextBuf, "%s\x05%04X", buf[:], adrCreatePortraitSprites)
				printWindowText(gTextBuf)
				flip = 0
				lbl = lNone
				if sel == 0xff {
					lbl = lSearch
					break
				}
				createMarkSelection(sel, flip)
			case lStat:
				lbl = lNone
				*gGameMode = 0x56
				drawIcon(img, b*2+11, int(gCreateHotXs[b]), int(gCreateHotYs[b]), 0)
				btn = b
				b = pickStatExchange(btn + 1)
				*gCurIcons = img
				if b--; b+1 != 0 {
					drawIcon(img, b*2+11, int(gCreateHotXs[b]), int(gCreateHotYs[b]), 0)
					st[btn], st[b] = st[b], st[btn]
					calcAllowedClasses(st[:], ok[:])
					cls, align = 0xff, 0xff
					flip = 0
					sel = formatCreateStats(st[:], ok[:], buf[:], race, sex)
				} else {
					drawIcon(img, btn*2+10, int(gCreateHotXs[btn]), int(gCreateHotYs[btn]), 0)
				}
				waitMouseRelease()
				if b+1 != 0 {
					printWindowText(buf[:])
					if sel != 0xff {
						createMarkSelection(sel, flip)
					}
				}
				*gGameMode = 4
			}
		}
		if key == 0x1b {
			break
		}
	}
done:
	*gGameMode = uint8(save)
	freeAndNull(&img)
	freeAndNull(gDiceVga)
}

func createHitButton() int {
	my, mx := int(*gMouseY), int(*gMouseX)
	i := 0
	for {
		if int(gCreateHotYs[i]) <= my && int(gCreateHotYs[i])+int(gCreateHotHs[i]) > my && int(gCreateHotXs[i]) <= mx && int(gCreateHotXs[i])+int(gCreateHotWs[i]) > mx {
			return i + 1
		}
		if i++; gCreateHotXs[i] == -1 {
			break
		}
	}
	return 0
}
