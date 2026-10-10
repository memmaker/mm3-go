// Port of MISC1.C (overlay 1800): event helpers: rewinding the clock, indexing the event
// script, picking a party member, learning a scroll's spell, the take/give/set/compare event
// operations, the yes/no confirmation, password answers and opening locked objects.
//
// misc1's struct pc maps onto Chr by offset: award7/8/9/18 are Award[7/8/9/18], item_id_last is
// ItemID[0x11], raw[0x312] of a map is LockLevel and raw[0x31e] F314[10]. misc1 declares some
// int globals unsigned (g_minutes, g_light_count, the resists): compares use those types.
package mm3

// Literal strings the C passes as far pointers (package-level, see PORTING.md).
var howMuchText = cs("\x03cHow Much\n\n")

func rewindClock(n int) {
	m := at[int16](adrMinutes)
	*m -= int16(n)
	for *m < 0 {
		*m += 1440
		if *gDay--; int(*gDay) < 0 { // (int)(u8) is never negative: the C never wraps the year here
			*gDay += 100
			*gYear--
		}
	}
}

func checkEventCount() {
	save := *gBtnXs
	*gBtnXs = 0x41d
	max := uint16(505)
	if *gEventCount > *gEventCountMax {
		*gEventCountMax = *gEventCount
	}
	if *gEventCount > max {
		fadeScreen(1, 4)
		sprintf(gTmpText, "\x03cError ... Run to Mark...\nEvents New Max = %u - %u", *gEventCount, max)
		openTextWindow(0x32, 0x4c, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTmpText))
		getInput()
		closeWindows(1)
	}
	*gBtnXs = save
}

func indexEventScript() {
	var k, max, i, last int
	*gEventCount, *gEventAuxCount = 0, 0
	memset(ds[adrEvents:], 0x80, 0x13ba)
	for {
		e := &gEvents[*gEventCount]
		e.X.Set(ms(i + 1))
		e.Y.Set(ms(i + 2))
		e.Dir.Set(ms(i + 3))
		e.Line.Set(ms(i + 4))
		e.IP.Set(i)
		*gEventCount++
		i = int(int16(i + ms(i)))
		i++
		if uint16(i) >= *gMazeEvtSize {
			break
		}
	}
	memset(ds[adrMazeTextOffs:], 0, 0x8c)
	if *gMapText != 0 {
		k = 1
		i, max, last = 0, 0, 0
		t := gMapText.B()
		for ; uint16(i) < *gMapTextSize; i++ {
			if t[i] == 0 {
				gMazeTextOffs[k] = int16(i + 1)
				k++
				last = i - last
				if last > max {
					max = last
				}
				last = i
			}
		}
		*gMazeTextCount = uint8(k - 1)
	}
	checkEventCount()
}

func selectPartyMember() int {
	var k int
	if *gPartyCount == 1 {
		return 0
	}
	old49c := *gPortraitClickOn
	*gPortraitClickOn = 1
	oldmode := *gGameMode
	if *gNpcWindowOpen != 0 {
		*gGameMode = 0xff
	} else if *gGameMode == 7 {
		*gGameMode = 7
	} else {
		*gGameMode = 6
	}
	pushButtons()
	setAnykeyButtons()
	head := "\x0b012"
	if oldmode == 1 {
		head = "\x0c07Unlock Door\x0cd\n\n"
	}
	sprintf(gTextBuf2, "\x03c%sSelect which char (F1 - F%d)?", head, *gPartyCount)
	openTextWindow(0xe7, 0x4a, 0x56, 0x48, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
loop:
	for {
		k = getInput()
		switch k {
		case 0x1b:
			k = -1
			break loop
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if k -= 0xc9; k < int(*gPartyCount) && charCannotAct(party(k)) == 0 {
				break loop
			}
		}
	}
	*gPortraitClickOn = old49c
	*gGameMode = oldmode
	popButtons()
	closeWindows(1)
	return k
}

// spellKind is the g_class_spells row of a class (cleric 0, sorcerer 1, druid 2), ok false for
// a class without spells.
func spellKind(cls uint8) (int, bool) {
	switch cls {
	case 1, 3:
		return 0, true
	case 2, 4:
		return 1, true
	case 8, 9:
		return 2, true
	}
	return 0, false
}

func learnScrollSpell(spell int) int {
	var can [8]byte
	var any byte
	var ret, who, j int
	pushButtons()
	setAnykeyButtons()
	for i := 0; i < int(*gPartyCount); i++ {
		j = 0
		kind, ok := spellKind(gParty[i].Cls)
		if !ok {
			continue
		}
		for {
			j++
			if int(gClassSpells[kind][j-1]) == spell {
				can[i], any = 1, 1
				break
			}
			if j >= 0x24 {
				break
			}
		}
	}
	if any == 0 {
		openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(int(*gMsgScrollCrumbles)))
		confirmYesNo(1)
		goto none
	}
	sprintf(gTextBuf, gostr(str(*gMsgFoundScroll)), str(gSpellNames[spell]))
	openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	if confirmYesNo(0) == 0 {
		goto none
	}
again:
	closeWindows(1)
	if who = selectPartyMember(); who == -1 {
		ret = 0
		goto out
	}
	if can[who] != 0 {
		ret = 1
		if kind, ok := spellKind(gParty[who].Cls); ok {
			j = kind
		}
		for i := 0; i < 0x24; i++ {
			if int(gClassSpells[j][i]) == spell {
				gParty[who].Spells[i] = 1
				break
			}
		}
		sprintf(gTextBuf, "\x03c\x0b012%s learns a \x0c12%s\x0cd spell.", gParty[who].Name[:], str(gSpellNames[spell]))
		openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		confirmYesNo(1)
	} else {
		sprintf(gTextBuf, gostr(str(*gMsgCannotReadScroll)), gParty[who].Name[:])
		openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		confirmYesNo(1)
		goto again
	}
	goto done
none:
	ret = 0
done:
	closeWindows(1)
out:
	popButtons()
	return ret
}

func askHowMuch() uint32 {
	old49c := *gPortraitClickOn
	*gPortraitClickOn = 1
	oldmode := *gGameMode
	*gGameMode = 6
	pushButtons()
	setAnykeyButtons()
	openTextWindow(0xe7, 0x4a, 0x56, 0x48, 0, 0x96, 4, 6, farOf(howMuchText))
	v := uint32(inputNumber(8, 0x46))
	*gPortraitClickOn = old49c
	*gGameMode = oldmode
	popButtons()
	closeWindows(1)
	return v
}

// classSpellSlot sets g_mouse_x (the C's scratch) to the class's spell row and returns the
// slot of spell v in it, or -1.
func classSpellSlot(p *Chr, v uint32) int {
	if kind, ok := spellKind(p.Cls); ok {
		*gMouseX = int16(kind)
	}
	for j := 0; j < 0x24; j++ {
		if uint32(gClassSpells[*gMouseX][j]) == v {
			return j
		}
	}
	return -1
}

// addCap is `if (x + v > 0xff) x = 0xff; else x += v;`.
func addCap(x *uint8, v uint32) {
	if uint32(*x)+v > 0xff {
		*x = 0xff
	} else {
		*x += uint8(v)
	}
}

// eventTakeGive takes (m1, v1) and gives (m2, v2) for party member idx; 1 = failed.
func eventTakeGive(m1 int, v1 uint32, m2 int, v2 uint32, idx int) int {
	if idx > 7 {
		idx &= 7
		m1 = 0
	}
	p := party(idx)
	if m1 != 0 && v1 == 0 {
		if v1 = askHowMuch(); v1 == 0 {
			return 1
		}
		if m2 != 0 && v2 == 0 {
			v2 = v1
		}
	} else if m1 == m2 && v1 == v2 {
		if v2 != 0 {
			v1 = uint32(int32(int16(randomRange(1, int(int16(v2))))))
		}
		m2, v2 = 0, 0
	}
	b := uint8(v1)
	switch m1 {
	case 8:
		damageParty(int(int16(v1)), int(*gEventDmgType), idx)
	case 9:
		if p.Caster != 0 {
			p.SP.Set(p.SP.Get() - int(int16(v1)))
			if p.SP.Get() < 1 {
				p.SP.Set(0)
			}
		}
	case 10:
		p.AcTmp -= b
	case 11:
		p.LevelTmp -= b
	case 12:
		p.AgeMod -= b
	case 13:
		p.Skill[v1] = 0
	case 15:
		p.Award[v1] = 0
	case 16:
		p.Exp.Set(p.Exp.Get() - v1)
	case 17:
		*gPoisonResist -= int16(v1)
	case 18:
		p.Cond[v1] = 0
	case 19:
		if j := classSpellSlot(p, v1); j >= 0 {
			p.Spells[j] = 0
		}
	case 20:
		setBit(gGameFlags, int(v1), 0)
	case 21:
		for j := 0; j < 0x12; j++ {
			if uint32(p.ItemID[j]) == v1 {
				clearItemSlot(p, j)
				compactBackpack(p)
				goto give
			}
		}
		return 1
	case 23:
		gQuestFlags[v1] = 0
	case 25:
		passTime(int(int16(v1)))
	case 34:
		if payCost(0, v1, 0) == 0 {
			return 1
		}
	case 35:
		if payCost(1, v1, 0) == 0 {
			return 1
		}
	case 37:
		p.MightTmp -= b
	case 38:
		p.IntelTmp -= b
	case 39:
		p.PersTmp -= b
	case 40:
		p.EndurTmp -= b
	case 41:
		p.SpeedTmp -= b
	case 42:
		p.AccTmp -= b
	case 43:
		p.LuckTmp -= b
	case 45:
		p.Might -= b
	case 46:
		p.Intel -= b
	case 47:
		p.Pers -= b
	case 48:
		p.Endur -= b
	case 49:
		p.Speed -= b
	case 50:
		p.Acc -= b
	case 51:
		p.Luck -= b
	case 52:
		p.FireRes -= b
	case 53:
		p.ElecRes -= b
	case 54:
		p.ColdRes -= b
	case 55:
		p.PoisonRes -= b
	case 56:
		p.EnergyRes -= b
	case 57:
		p.MagicRes -= b
	case 58:
		p.FireResTmp -= b
	case 59:
		p.ElecResTmp -= b
	case 60:
		p.ColdResTmp -= b
	case 61:
		p.PoisonResTmp -= b
	case 62:
		p.EnergyResTmp -= b
	case 63:
		p.MagicResTmp -= b
	case 64:
		p.Level -= b
	case 65:
		if payCost(2, v1, 0) == 0 {
			return 1
		}
	case 69:
		*gLevitate -= b
	case 70:
		*gLightCount -= int16(v1)
	case 71:
		*gFireResist -= int16(v1)
	case 72:
		*gElecResist -= int16(v1)
	case 73:
		*gColdResist -= int16(v1)
	case 74:
		*gLevitate -= b
		*gLightCount -= int16(v1)
		*gFireResist -= int16(v1)
		*gElecResist -= int16(v1)
		*gColdResist -= int16(v1)
		*gPoisonResist -= int16(v1)
		*gWalkOnWater = 0
	case 76:
		rewindClock(int(int16(v1 * 1440)))
	case 77:
		p.AcTmp -= b
	case 79:
		*gWizardEye = 0
	case 85:
		*gYear -= uint16(v1)
	case 94:
		*gWalkOnWater = 0
	}
give:
	b = uint8(v2)
	switch m2 {
	case 3:
		p.Sex = b
	case 4:
		p.Race = b
	case 5:
		p.Cls = b
	case 6:
		p.Align = b
	case 8:
		p.HP.Set(p.HP.Get() + int(int16(v2)))
	case 9:
		p.SP.Set(p.SP.Get() + int(int16(v2)))
	case 10:
		p.AcTmp += b
	case 11:
		p.LevelTmp += b
	case 12:
		p.AgeMod += b
	case 13:
		p.Skill[v2]++
	case 15:
		p.Award[v2]++
	case 16:
		p.Exp.Set(p.Exp.Get() + v2)
	case 17:
		*gPoisonResist += int16(v2)
	case 18:
		setCond(p, v2)
	case 19:
		if j := classSpellSlot(p, v2); j >= 0 {
			p.Spells[j] = 1
		}
	case 20:
		setBit(gGameFlags, int(v2), 1)
	case 21:
		for j := idx; j < int(*gPartyCount); j++ {
			if p.ItemID[0x11] == 0 {
				p.ItemID[0x11] = b
				compactBackpack(p)
				return 0
			}
		}
		return 1
	case 23:
		gQuestFlags[v2] = 1
	case 25:
		rewindClock(int(int16(v2)))
	case 34:
		*gPartyGold += v2
	case 35:
		*gPartyGems += v2
	case 37:
		addCap(&p.MightTmp, v2)
	case 38:
		addCap(&p.IntelTmp, v2)
	case 39:
		addCap(&p.PersTmp, v2)
	case 40:
		addCap(&p.EndurTmp, v2)
	case 41:
		addCap(&p.SpeedTmp, v2)
	case 42:
		addCap(&p.AccTmp, v2)
	case 43:
		addCap(&p.LuckTmp, v2)
	case 45:
		addCap(&p.Might, v2)
	case 46:
		addCap(&p.Intel, v2)
	case 47:
		addCap(&p.Pers, v2)
	case 48:
		addCap(&p.Endur, v2)
	case 49:
		addCap(&p.Speed, v2)
	case 50:
		addCap(&p.Acc, v2)
	case 51:
		addCap(&p.Luck, v2)
	case 52:
		p.FireRes += b
	case 53:
		p.ElecRes += b
	case 54:
		p.ColdRes += b
	case 55:
		p.PoisonRes += b
	case 56:
		p.EnergyRes += b
	case 57:
		p.MagicRes += b
	case 58:
		p.FireResTmp += b
	case 59:
		p.ElecResTmp += b
	case 60:
		p.ColdResTmp += b
	case 61:
		p.PoisonResTmp += b
	case 62:
		p.EnergyResTmp += b
	case 63:
		p.MagicResTmp += b
	case 64:
		p.Level += b
	case 65:
		*gFood += uint16(v2)
	case 66:
		for j := idx; j < int(*gPartyCount); j++ {
			if p.ItemID[0x11] == 0 {
				generateRandomItem(int(b), p, 0x11)
				compactBackpack(p)
				return 0
			}
		}
		return 1
	case 67:
		if learnScrollSpell(int(int16(v2))) != 0 {
			return 1
		}
	case 69:
		*gLevitate += b
	case 70:
		*gLightCount += int16(v2)
	case 71:
		*gFireResist += int16(v2)
	case 72:
		*gElecResist += int16(v2)
	case 73:
		*gColdResist += int16(v2)
	case 74:
		*gLevitate += b
		*gLightCount += int16(v2)
		*gFireResist += int16(v2)
		*gElecResist += int16(v2)
		*gColdResist += int16(v2)
		*gPoisonResist += int16(v2)
		*gWalkOnWater = 1
	case 76:
		advanceClock(int(int16(v2 * 1440)))
	case 77:
		p.AcTmp += b
	case 79:
		*gWizardEye = 1
	case 82:
		damageParty(int(int16(v2)), int(*gEventDmgType), idx)
	case 85:
		*gYear += uint16(v2)
	case 94:
		*gWalkOnWater = 1
	}
	return 0
}

// setCond is the condition give/set: 16 clears all, 6 sets, others count up; dead, stoned or
// eradicated end positive hit points.
func setCond(p *Chr, v uint32) {
	if v == 16 {
		memset(p.Cond[:], 0, 16)
	} else if v == 6 {
		p.Cond[v] = 1
	} else {
		p.Cond[v]++
	}
	if v >= 13 && v <= 15 && p.HP.Get() > 0 {
		p.HP.Set(0)
	}
}

func confirmYesNo(quiet int) int {
	var key, r int
	pushButtons()
	oldmode := *gGameMode
	if *gGameMode != 7 {
		*gGameMode = 8
	}
	if quiet == 0 {
		pushIcons("confirm2.icn")
		*gConfirmSpritesImg2 = *gCurIcons
		*gConfirm2SpritesImg = *gGlobalIcons
		sprintf(gTextBuf2, "\x03b\x05%04X", adrConfirm2Sprites)
		printWindowText(gTextBuf2)
		*gBtnXs = 0x455
		*gBtnYs = 0x45b
		*gBtnWs = 0x45d
		*gBtnHs = 0x45f
	} else {
		setAnykeyButtons()
	}
	setMouseCursor(0, uint32(*gMouseCursor))
	old49c := *gPortraitClickOn
	*gPortraitClickOn = 0
	waitMouseRelease()
	for {
		if oldmode == 11 {
			key = buildingWaitInput()
		} else {
			key = getInput()
		}
		if quiet != 0 {
			r = 1
			break
		}
		switch key {
		case 0, 'Y':
			r = 1
		case 1, 0x1b, 'N':
			r = 0
		default:
			continue
		}
		break
	}
	popButtons()
	*gPortraitClickOn = old49c
	*gGameMode = oldmode
	if quiet == 0 {
		popIcons()
		drawMainIcons()
	}
	return r
}

func eventSetValue(mode int, v uint32, idx int) {
	p := party(idx)
	b := uint8(v)
	w := int16(v)
	switch mode {
	case 3:
		p.Sex = b
	case 4:
		p.Race = b
	case 5:
		p.Cls = b
	case 6:
		p.Align = b
	case 8:
		p.HP.Set(int(w))
	case 9:
		p.SP.Set(int(w))
	case 10:
		p.AcTmp = b
	case 11:
		p.LevelTmp = b
	case 12:
		p.AgeMod = b
	case 16:
		p.Exp.Set(v)
	case 17:
		*gPoisonResist = w
	case 18:
		setCond(p, v)
	case 25:
		*gMinutes = uint16(v)
	case 34:
		*gPartyGold = v
	case 35:
		*gPartyGems = v
	case 37:
		p.MightTmp = b
	case 38:
		p.IntelTmp = b
	case 39:
		p.PersTmp = b
	case 40:
		p.EndurTmp = b
	case 41:
		p.SpeedTmp = b
	case 42:
		p.AccTmp = b
	case 43:
		p.LuckTmp = b
	case 45:
		p.Might = b
	case 46:
		p.Intel = b
	case 47:
		p.Pers = b
	case 48:
		p.Endur = b
	case 49:
		p.Speed = b
	case 50:
		p.Acc = b
	case 51:
		p.Luck = b
	case 52:
		p.FireRes = b
	case 53:
		p.ElecRes = b
	case 54:
		p.ColdRes = b
	case 55:
		p.PoisonRes = b
	case 56:
		p.EnergyRes = b
	case 57:
		p.MagicRes = b
	case 58:
		p.FireResTmp = b
	case 59:
		p.ElecResTmp = b
	case 60:
		p.ColdResTmp = b
	case 61:
		p.PoisonResTmp = b
	case 62:
		p.EnergyResTmp = b
	case 63:
		p.MagicResTmp = b
	case 64:
		p.Level = b
	case 65:
		*gFood = uint16(v)
	case 69:
		*gLevitate = b
	case 70:
		*gLightCount = w
	case 71:
		*gFireResist = w
	case 72:
		*gElecResist = w
	case 73:
		*gColdResist = w
	case 74:
		// a chain of assignments through misc1's u8 g_walk_on_water: everything gets (u8)v
		*gWalkOnWater = int8(b)
		*gPoisonResist = int16(b)
		*gWizardEye = b
		*gColdResist = int16(b)
		*gElecResist = int16(b)
		*gFireResist = int16(b)
		*gLightCount = int16(b)
		*gLevitate = b
	case 76:
		*gDay = b
	case 77:
		p.AcTmp = b
	case 79:
		*gWizardEye = 1
	case 83:
		*gEventDmgType = b
	case 84:
		*gPartyDir = b
		if *gOpenWindows == 0 {
			drawWizardEye(1)
		}
	case 85:
		*gYear = uint16(v)
	case 94:
		*gWalkOnWater = int8(b)
	}
}

// eventCompareValue compares (what) of party member idx with val; mode 0 >=, 1 ==, 2 <=. v is
// uninitialised in the C where a case leaves it unset; here it is 0.
func eventCompareValue(what int, val uint32, mode, idx int) int {
	var v uint32
	p := party(idx)
	u16 := func(x int) uint32 { return uint32(uint16(x)) }
	switch what {
	case 3:
		v = uint32(p.Sex)
	case 4:
		v = uint32(p.Race)
	case 5:
		v = uint32(p.Cls)
	case 6:
		v = uint32(p.Align)
	case 8:
		v = uint32(int32(p.HP.Get()))
	case 9:
		v = uint32(int32(p.SP.Get()))
	case 10:
		v = u16(calcArmorClass(p, 0))
	case 11:
		v = uint32(p.LevelTmp)
	case 12:
		v = u16(getCharAge(p, 0))
	case 13:
		if p.Skill[val] != 0 {
			v = val
		}
	case 15:
		if p.Award[val] != 0 {
			v = val
		}
	case 16:
		v = p.Exp.Get()
	case 17:
		v = u16(int(*gPoisonResist))
	case 18:
		if val == 16 || p.Cond[val] != 0 {
			v = val
		}
	case 19:
		if kind, ok := spellKind(p.Cls); ok {
			*gMouseX = int16(kind)
		}
		for j := 0; j < 0x24; j++ {
			if uint32(gClassSpells[*gMouseX][j]) == val {
				if p.Spells[j] != 0 {
					v = val
				}
				break
			}
		}
	case 20:
		if testBit(gGameFlags, int(val)) != 0 {
			v = val
		} else {
			v = 0xffffffff
		}
	case 21:
		v = 100
		if val == 0x49 && *gRopeCreated != 0 {
			v = val
		} else {
			for j := 0; j < 0x12; j++ {
				if uint32(p.ItemID[j]) == val {
					v = val
				}
			}
		}
	case 23:
		if gQuestFlags[val] != 0 {
			v = val
		}
	case 25:
		v = uint32(*gMinutes)
	case 34:
		v = *gPartyGold
	case 35:
		v = *gPartyGems
	case 37:
		v = uint32(p.MightTmp)
	case 38:
		v = uint32(p.IntelTmp)
	case 39:
		v = uint32(p.PersTmp)
	case 40:
		v = uint32(p.EndurTmp)
	case 41:
		v = uint32(p.SpeedTmp)
	case 42:
		v = uint32(p.AccTmp)
	case 43:
		v = uint32(p.LuckTmp)
	case 44:
		v = uint32(int32(int16(confirmYesNo(int(int16(val))))))
		if val == 0 && v != 0 || val != 0 {
			v = val
		} else {
			v = 2
		}
	case 45:
		v = uint32(p.Might)
	case 46:
		v = uint32(p.Intel)
	case 47:
		v = uint32(p.Pers)
	case 48:
		v = uint32(p.Endur)
	case 49:
		v = uint32(p.Speed)
	case 50:
		v = uint32(p.Acc)
	case 51:
		v = uint32(p.Luck)
	case 52:
		v = uint32(p.FireRes)
	case 53:
		v = uint32(p.ElecRes)
	case 54:
		v = uint32(p.ColdRes)
	case 55:
		v = uint32(p.PoisonRes)
	case 56:
		v = uint32(p.EnergyRes)
	case 57:
		v = uint32(p.MagicRes)
	case 58:
		v = uint32(p.FireResTmp)
	case 59:
		v = uint32(p.ElecResTmp)
	case 60:
		v = uint32(p.ColdResTmp)
	case 61:
		v = uint32(p.PoisonResTmp)
	case 62:
		v = uint32(p.EnergyResTmp)
	case 63:
		v = uint32(p.MagicResTmp)
	case 64:
		v = uint32(p.Level)
	case 65:
		v = uint32(*gFood)
	case 69:
		v = uint32(*gLevitate)
	case 70:
		v = u16(int(*gLightCount))
	case 71:
		v = u16(int(*gFireResist))
	case 72:
		v = u16(int(*gElecResist))
	case 73:
		v = u16(int(*gColdResist))
	case 76:
		v = uint32(*gDay)
	case 77:
		v = uint32(p.AcTmp)
	case 78:
		v = uint32(b2i(uint16(calcMaxHp(p)) < uint16(p.HP.Get())))
	case 79:
		v = uint32(*gWizardEye)
	case 81:
		v = uint32(b2i(uint16(calcMaxSp(p)) < uint16(p.SP.Get())))
	case 84:
		v = uint32(*gPartyDir)
	case 85:
		v = uint32(*gYear)
	case 86, 87, 88, 89, 90, 91, 92:
		v = u16(getCharStat(p, what-86, 0))
	case 93:
		v = uint32(*gDay % 10)
	case 94:
		v = uint32(uint8(*gWalkOnWater))
	case 95:
		v = uint32(p.Award[18])
	case 96:
		v = uint32(p.Award[7])
	case 97:
		v = uint32(p.Award[9])
	case 98:
		v = uint32(p.Award[8])
	case 99:
		if partyHasSkill(int(int16(val))) != 0 {
			v = val
		} else {
			v = 0xffffffff
		}
	}
	switch mode {
	case 1:
		if v == val {
			return 1
		}
	case 2:
		if v <= val {
			return 1
		}
	case 0:
		if v >= val {
			return 1
		}
	}
	return 0
}

// askAnswer asks for an answer; mode 1: compare with ans (or g_text_buf2), mode 0: the index in
// g_answer_words[] + 1.
func askAnswer(mode int, ans, msg []byte) int {
	var buf [20]byte
	shown, i, r := 0, 0, 0
	sprintf(gTextBuf, "\r\x03c%s\n\n", msg)
	if *gOpenWindows == 0 || *gNpcWindowOpen != 0 {
		openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		shown = 1
	} else {
		printWindowText(gTextBuf)
	}
	if inputText(buf[:], 0x12, 0x130) != 0 {
		if mode != 0 {
			if stricmp(buf[:], gTextBuf2) == 0 || stricmp(buf[:], ans) == 0 {
				r = 1
			}
		} else {
			for {
				if stricmp(buf[:], str(gAnswerWords[i])) == 0 {
					r = i + 1
					i = 13
				}
				i++
				if i-1 >= 13 {
					break
				}
			}
		}
	}
	if shown != 0 {
		closeWindows(1)
	}
	return r
}

// eventGiveMulti opens a trapped or locked object first (who picks the lock), then gives up to
// three things.
func eventGiveMulti(m1 int, v1 uint32, m2 int, v2 uint32, m3 int, v3 uint32, idx int) int {
	var who int
	snd, again, flag := 0, 0, 0
	if *gObjD0c != 0 && *gEventObjOpened == 0 {
		o := &gObjects1based[*gObjD0c]
		switch o.Type.Get() {
		case 4:
			snd = 0x14
		case 12:
			snd = 0x2a
		case 18:
			snd = 0x2a
		case 20:
			snd = 0x14
		case 23:
			snd = 0x14
		case 38:
			snd = 0x14
		case 33:
			snd = 0x2b
		case 40:
			snd = 0x2b
		default:
			goto give
		}
		*gEventObjOpened++
		o.Frame.Set(1)
		if o.Type.Get() == 0x14 {
			goto give
		}
	top:
		if who = selectPartyMember(); who == -1 {
			goto fail
		}
		m := &gMaps[*gMapQuadrant]
		if randomRange(1, 4) == 1 {
			damageParty(int(m.F314[10]), randomRange(0, 6), who)
		}
		if getThievery(party(who))+randomRange(1, 20) >= int(m.LockLevel) {
			c := party(who)
			c.Exp.Set(c.Exp.Get() + uint32(uint16(getCharLevel(c)*int(m.LockLevel)*10)))
			*gGameMode = 7
			sprintf(gTextBuf2, "\x03c\x0b010%s picks the lock!\n\nPress any key.", c.Name[:])
			playSfx(snd)
			openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
			confirmYesNo(1)
			closeWindows(1)
			if again != 0 {
				goto top
			}
			goto give
		}
		sprintf(gTextBuf2, "\x03c\x0b010%s was unable to pick the lock!\nPress any key.", gParty[who].Name[:])
		showMessage(gTextBuf2)
		goto fail
	}
give:
	{
		ms := [3]int{m1, m2, m3}
		vs := [3]uint32{v1, v2, v3}
		for i := 0; i < 3; i++ {
			switch ms[i] {
			case 0x22:
				*gLootGold += int32(vs[i])
			case 0x23:
				*gLootGems += int32(vs[i])
			case 0x15:
				if *gLootCount++; *gLootCount > 10 {
					*gLootCount = 10
				}
				n := int(*gLootCount) - 1
				gLootFlags[n], gLootElement[n], gLootMaterial[n], gLootAttrib[n], gLootCharges[n] = 0, 0, 0, 0, 0
				gLootIds[n] = uint8(vs[i])
			case 0x42:
				if *gLootCount++; *gLootCount > 10 {
					*gLootCount = 10
				}
				c := gScratchChar
				generateRandomItem(int(uint8(vs[i])), c, 0)
				n := int(*gLootCount) - 1
				gLootFlags[n] = c.ItemFlags[0]
				gLootElement[n] = c.ItemElem[0]
				gLootMaterial[n] = c.ItemMat[0]
				gLootAttrib[n] = c.ItemAttr[0]
				gLootIds[n] = c.ItemID[0]
				gLootCharges[n] = c.ItemPower[0]
			case 0x43:
				flag = 1
				fallthrough
			default:
				if eventTakeGive(0, 0, ms[i], vs[i], idx) != 0 {
					return 1
				}
				if flag != 0 {
					return 0
				}
			}
		}
	}
	return 0
fail:
	*gEventObjOpened = -1
	return 1
}
