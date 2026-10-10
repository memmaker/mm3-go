// Port of SHOP.C (overlay 1867): the character screen, the item lists (equip, buy, sell, fix,
// identify, the item spells), the quick reference, class/item checks, max SP/HP and armor class,
// item prices and names.
//
// shop's struct chr maps onto Chr by offset: stats is bytes 0x14.., bless_res[k] byte 0x102+k
// (Bless .. the resistances). shop declares get_item_stat_bonus unsigned: calcMaxSp/calcMaxHp
// add it zero-extended to their longs, as the original does. The item list's per-row draw lists
// (struct spr1 hot[18], printed as "\x05%p") live on the ds stack (dsAlloca).
package mm3

func useItemPower(n int) {
	switch n {
	case 1:
		spellLight()
	case 2:
		spellAwaken()
	case 3:
		spellDetectMagic()
	case 4:
		spellElementalArrow()
	case 5:
		spellFirstAid()
	case 6:
		spellFlyingFist()
	case 7:
		spellEnergyBlast()
	case 8:
		spellSleep()
	case 9:
		spellRevitalize()
	case 10:
		spellCureWounds()
	case 11:
		spellSparks()
	case 12:
		spellCreateRope()
	case 13:
		spellToxicCloud()
	case 14:
		spellProtFromElements()
	case 15:
		spellPain()
	case 16:
		if *gGameMode == 2 {
			showCantCastEngaged(0x11)
		} else {
			spellJump()
		}
	case 17:
		spellAcidStream()
	case 18:
		spellTurnUndead()
	case 19:
		spellLevitate()
	case 20:
		if *gGameMode == 2 {
			showCantCastEngaged(0x17)
		} else {
			spellWizardEye()
		}
	case 21:
		spellSilence()
	case 22:
		spellBlessed()
	case 23:
		spellIdentifyMonster()
	case 24:
		spellLightningBolt()
	case 25:
		spellHolyBonus()
	case 26:
		spellPowerCure()
	case 27:
		spellNaturesCure()
	case 28:
		if *gGameMode == 2 {
			showCantCastEngaged(0x21)
		} else {
			spellLloydsBeacon()
		}
	case 29:
		spellPowerShield()
	case 30:
		spellHeroism()
	case 31:
		spellImmobilize()
	case 32:
		spellWalkOnWater()
	case 33:
		spellFrostBite()
	case 34:
		spellDetectMonster()
	case 35:
		spellFireball()
	case 36:
		spellColdRay()
	case 37:
		spellCurePoison()
	case 38:
		spellAcidSpray()
	case 39:
		spellTimeDistortion()
	case 40:
		spellFeebleMind()
	case 41:
		spellCureDisease()
	case 42:
		if *gGameMode == 2 {
			showCantCastEngaged(0x2b)
		} else {
			spellNaturesGate()
		}
	case 43:
		if *gGameMode == 2 {
			showCantCastEngaged(0x2d)
		} else {
			spellTeleport()
		}
	case 44:
		spellFingerOfDeath()
	case 45:
		spellCureParalysis()
	case 46:
		spellParalyze()
	case 47:
		spellDeadlySwarm()
	case 48:
		if *gGameMode == 2 {
			showCantCastEngaged(0x32)
		} else {
			spellSuperShelter()
		}
	case 49:
		spellDragonBreath()
	case 50:
		spellCreateFood()
	case 51:
		spellFieryFlail()
	case 52:
		spellRechargeItem()
	case 53:
		spellFantasticFreeze()
	case 54:
		if *gGameMode == 2 {
			showCantCastEngaged(0x35)
		} else {
			spellTownPortal()
		}
	case 55:
		spellStoneToFlesh()
	case 56:
		spellDuplication()
	case 57:
		spellDisintegrate()
	case 58:
		spellHalfForMe()
	case 59:
		spellRaiseDead()
	case 60:
		if *gGameMode == 2 {
			showCantCastEngaged(0x3d)
		} else {
			spellEtherealize()
		}
	case 61:
		spellDancingSword()
	case 62:
		spellMoonRay()
	case 63:
		spellMassDistortion()
	case 64:
		spellPrismaticLight()
	case 65:
		spellEnchantItem()
	case 66:
		spellIncinerate()
	case 67:
		spellHolyWord()
	case 68:
		spellResurrect()
	case 69:
		spellElementalStorm()
	case 70:
		spellMegaVolts()
	case 71:
		spellInferno()
	case 72:
		spellSunRay()
	case 73:
		spellImplosion()
	case 74:
		spellStarBurst()
	case 75:
		spellDivineIntervention()
	case 76:
		drinkMightPotion()
	case 77:
		*gTeleportPending = 1
	}
}

func charScreen(n int) {
	var old, party, key, drawn int
	var c *Chr
	buf := gDrawListB
	who := func() int {
		if old == 2 {
			return int(gCombatOrder[n])
		}
		return n
	}

	party, drawn = 0, 0
	*gStatCursor = 0
	old = int(*gGameMode)
	*gGameMode = 10
	pushButtons()
	setMouseCursor(0, uint32(*gMouseCursor))
	updateFreeMemKb()
	if *gFreeMemKb < 0x1e && *gOutdoorObjGfx != 0 {
		freeMazeGfx()
		party = 1
	}
	loadCharscrIcons()
	c = &gParty[who()]
	highlightPortrait(n)
redraw:
	formatCharSheet(c, buf, who())
	sprintf(gTextBuf, "\t029\v018Mgt\t080Acy\t131H.P.\t196Experience\t029\v041Int\t080Lck\t131S.P.\t029\v064Per\t080Age\t131Resis\t196Party Gems\t029\v087End\t080Lvl\t131Skills\t196Party Food\t029\v110Spd\t080AC\t131Awrds\t196Condition\003c\t290\v025\01427I\014dtem\t290\v057\01427Q\014duick\t290\v089\01427E\014dxch\t290\v121Exit\003l\005%04X%s", adrCharscrSprites, buf)
	if drawn == 0 {
		openTextWindow(0, 0, 0x140, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		drawn = 1
	} else {
		printWindowText(gTextBuf)
	}
	drawCharsheetCursor(1)
	for {
		key = getInput()
		switch {
		case key >= 0 && key <= 19:
			*gGameMode = uint8(old)
			showCharsheetPopup(key, c)
			*gGameMode = 10
		case key == 0x14 || key == 'I':
			*gGameMode = uint8(old)
			*gItemsInCombat = int8(b2i(*gGameMode == 2))
			if c = itemScreen(c, 0); c == nil {
				*gViewDirty = 1
				goto quit
			}
			*gGameMode = 10
			goto redraw
		case key == 0x15 || key == 'Q':
			*gGameMode = uint8(old)
			showQuickRef()
			*gGameMode = 10
		case key == 0x16 || key == 'E':
			*gGameMode = uint8(old)
			c = exchangePartyMember(&n)
			*gGameMode = 10
		case key >= 0xc9 && key <= 0xd0:
			cnt := int(*gPartyCount)
			if old == 2 {
				cnt = int(*gCombatPartyN)
			}
			if key -= 0xc9; cnt > key {
				n = key
				c = &gParty[who()]
				highlightPortrait(n)
				formatCharSheet(c, buf, who())
				printWindowText(buf)
			}
		}
		if key == 0x17 || key == 0x1b {
			break
		}
	}
	closeWindows(1)
quit:
	clearCharHighlight()
	freeAndNull(gCurIcons)
	*gItemsInCombat = 0
	popButtons()
	*gGameMode = uint8(old)
	if party != 0 {
		showPleaseWait()
		loadMazeGfx(int(*gMapNum) - 1)
		closeWindows(1)
	}
	if *gTeleportPending != 0 {
		*gTeleportPending = 0
		transportBoxTeleport()
	}
}

// itemScreen is item_screen. Its gotos into other case bodies (c0, c1, c2, common, act, reset,
// pick) become a `next` step run after the key switch.
func itemScreen(c *Chr, mode int) *Chr {
	const (
		nBreak = iota
		nCommon
		nAct
		nReset
		nList
		nAgain
		nOut
	)
	var drawn, sel, tmp, cat, is7, oldcat, party, top, k, next int
	var orig *Chr
	var icon Far
	var cost uint32
	var name [20]byte
	var mark [40]uint8
	defer dsFree(dsSP)

	row := func(i int) []byte { return ds[adrDrawList+100*i:] }
	party1 := func(i int) *Chr {
		if *gGameMode == 2 {
			return &gParty[gCombatOrder[i]]
		}
		return &gParty[i]
	}
	orig = c
	if mode == 2 { // the removed Identify button (hotspot 2) takes no clicks
		w := ds[adrItemscrHotW+2]
		ds[adrItemscrHotW+2] = 0
		defer func() { ds[adrItemscrHotW+2] = w }()
	}
	if is7 = b2i(mode == 7); is7 != 0 {
		mode = 0
	}
	pushButtons()
	party, cat, drawn, top = 0, 0, 0, 0
	sel = -1
	updateFreeMemKb()
	if *gFreeMemKb < 10 && *gOutdoorObjGfx != 0 {
		freeMazeGfx()
		party = 1
	}
	setMouseCursor(0, uint32(*gMouseCursor))
	icon = loadGraphic("equip.icn")
	adrHot := dsAlloca(18 * 20)
	hot := arr[Spr1](adrHot)[:18]
	for k = 0; k < 18; k++ {
		hot[k].Flag.Set(-1)
		hot[k].End.Set(-1)
	}
	sprintf(name[:], "%s.icn", str(gItemscrIconNames[mode]))
	pushIcons(gostr(name[:]))
	gItemscrButtons[0].Img.Set(uint32(*gCurIcons))
	*gBtnXs = adrItemscrHotX
	*gBtnYs = adrItemscrHotY
	*gBtnWs = adrItemscrHotW
	*gBtnHs = adrItemscrHotH
	if mode == 1 {
		*gCurChar = nearOf(c)
		c = gScratchChar
		oldcat = 0
		loadShopStock(oldcat)
	}
	if c.ItemID[9] == 0 {
		*gItemscrDownFrame = 13
	} else {
		*gItemscrDownFrame = 12
	}
again:
	memset(mark[:], 0, 40)
list:
	for *gMouseX = int16(top); *gMouseX < 18; *gMouseX++ {
		mx := int(*gMouseX)
		if c.ItemID[mx] != 0 {
			if mode == 1 {
				c.ItemEq[mx] = uint8(getItemEquipCat(curChr(), int(c.ItemID[mx])))
			}
			h := &hot[mx]
			h.Img.Set(uint32(icon))
			h.Y.Set((mx-top)*9 + 10)
			h.Attr.Set(b2i(c.ItemEq[mx] == 2))
			switch eq := c.ItemEq[mx]; {
			case eq > 0 && eq < 3:
				h.Frame.Set(1)
			case eq == 13:
				h.Frame.Set(2)
			default:
				h.Frame.Set(int(eq))
			}
			fl := 0
			if mode != 1 {
				fl = b2i(c.Skill[8] == 0)
			}
			if mode == 2 {
				fl |= 0x80
			}
			cost = uint32(getItemPrice(c, mx, mode, fl))
			if mode != 0 && mode != 8 {
				ultoa(cost, name[:], 10)
			} else {
				strcpy(name[:], cs(""))
			}
			sprintf(row(mx), "\003r\014%2u\t023%2u)\003l\t028%s\003r\t304%s\014d\005%04X\n", mark[mx], mx-top+1,
				formatItemName(c, mx, gTmpText, int(mark[mx])), name[:], adrHot+20*mx)
		} else if mx == 0 {
			strcpy(row(mx), cs("\003c\n\t000No items available."))
		} else {
			strcpy(row(mx), cs(""))
		}
	}
	{
		rows := []any{}
		for i := 0; i < 9; i++ {
			rows = append(rows, row(top+i))
		}
		switch mode {
		case 0, 8:
			sprintf(gTextBuf, "\001\003f\r\003c\v000\t000Items for %s the %s\v011\002%s%s%s%s%s%s%s%s%s", append([]any{bytesOf(c), str(gClassNames[c.Cls])}, rows...)...)
		case 1:
			s := ""
			if cat == 0 {
				s = "s"
			}
			sprintf(gTextBuf, "\001\003f\r\003l\v000\t000Available %s%s\t150Gold - %lu\003r\t304Cost\003l\v011\002%s%s%s%s%s%s%s%s%s", append([]any{str(gItemActionLabels[cat][1]), s, *gPartyGold}, rows...)...)
		case 2:
			sprintf(gTextBuf, "\001\003f\r\003l\v000\t000Items for %s\t150Gold - %lu\003r\t304Cost\003l\v011\002%s%s%s%s%s%s%s%s%s", append([]any{bytesOf(c), *gPartyGold}, rows...)...)
		case 3, 4, 5, 6:
			sprintf(gTextBuf, "\001\003f\r\003l\v000\t000Items for %s\t150Gold - %lu\003r\t304Charges\003l\v011\002%s%s%s%s%s%s%s%s%s", append([]any{bytesOf(c), *gPartyGold}, rows...)...)
		}
	}
	if drawn == 0 {
		lbl2 := str(gItemActionLabels[2][mode])
		// the Smithy's Identify (mode 2, button 2) is removed: no label, and its draw-list
		// entry (x 159, frame 4) repeats button 1's, so nothing shows in its place
		btn2 := ds[adrItemscrButtons+6+16 : adrItemscrButtons+6+24]
		var saved [8]byte
		copy(saved[:], btn2)
		if mode == 2 {
			lbl2 = nil
			copy(btn2, ds[adrItemscrButtons+6+8:])
		}
		sprintf(gTmpText, "\002\005%04X\003c\v021\t019Up\t057Down\t095%s\t133%s\t171%s\t209%s\t286Exit", adrItemscrButtons,
			str(gItemActionLabels[0][mode]), str(gItemActionLabels[1][mode]), lbl2, str(gItemActionLabels[3][mode]))
		openTextWindow(0, 0x65, 0x140, 0x2d, 0, 0x96, 4, 6, dsFar(adrTmpText))
		copy(btn2, saved[:])
		openTextWindow(0, 0, 0x140, 0x6c, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		drawn = 1
	} else {
		printWindowText(gTextBuf)
	}
	printWindowText(cs("\001\014d"))
	drawItemPageArrows(c, top)
	waitMouseRelease()
	for {
		k = getInput()
		next = nBreak
		switch k {
		case 'W':
			if mode == 1 {
				k, next = 0, nCommon
			}
		case 'S':
			if mode == 2 {
				k, next = 0, nCommon
			}
		case 'E':
			if mode == 0 {
				k, next = 0, nCommon
			} else if mode == 6 {
				k, next = 1, nCommon
			}
		case 0:
			k, next = 0, nCommon
		case 'F':
			if mode == 2 {
				k, next = 1, nCommon
			}
		case 'A':
			if mode == 1 {
				k, next = 1, nCommon
			}
		case 'R', 1:
			if k == 'R' && mode != 0 && mode != 4 {
				break
			}
			k, next = 1, nCommon
		case 'M':
			if mode == 1 {
				k, next = 2, nCommon
			}
		case 'U', 2:
			if (k == 'U' && mode != 0) || mode == 2 { // mode 2's Identify is removed
				break
			}
			k, next = 2, nCommon
		case 'D', 3:
			if k == 'D' {
				if mode == 5 {
					k, next = 1, nCommon
					break
				}
				if mode != 0 {
					break
				}
			}
			k = 3
			if mode == 2 {
				break
			}
			next = nCommon
		case 5, 0xf2:
			if top == 0 {
				break
			}
			top -= 9
			next = nList
		case 6, 0xf3:
			if c.ItemID[9] == 0 {
				break
			}
			if top >= 9 {
				break
			}
			top += 9
			next = nList
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', 7, 8, 9, 10, 11, 12, 13, 14, 15:
			if k >= '1' {
				k -= 0x2a
			}
			if mode == 3 {
				break
			}
			if k < '1' && *gMouseButtons&2 != 0 { // right-click: the item's stats (port addition)
				if tmp = k + top - 7; c.ItemID[tmp] != 0 {
					showItemInfo(c, tmp)
				}
				next = nList
				break
			}
			if tmp = k + top - 7; tmp == sel {
				next = nReset
				break
			}
			if c.ItemID[tmp] == 0 {
				next = nList
				break
			}
			sel = tmp
			memset(mark[:], 0, 40)
			mark[sel] = 0x12
			if mode == 1 {
				k = cat
				next = nAct
				break
			}
			if mode != 8 {
				next = nList
				break
			}
			k = 3
			next = nAct
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if is7 != 0 {
				break
			}
			if mode == 3 {
				break
			}
			cnt := int(*gPartyCount)
			if *gGameMode == 2 {
				cnt = int(*gCombatPartyN)
			}
			if k -= 0xc9; cnt <= k {
				break
			}
			if mode == 1 {
				orig = party1(k)
				*gCurChar = nearOf(orig)
				highlightPortrait(k) // pick:
				next = nAgain
				break
			}
			if sel == -1 || mode == 1 {
				top = 0
				c = party1(k)
				orig = c
				highlightPortrait(k) // pick:
				next = nAgain
				break
			}
			tmp = k
			if c.ItemFlags[sel]&0x40 != 0 {
				showMessage(cs("\003cYou cannot remove a cursed item!"))
				next = nAgain
				break
			}
			p := &gParty[tmp]
			if p.ItemID[17] != 0 {
				sprintf(gTmpText, "\003c%s's backpack is full.", bytesOf(p))
				showMessage(gTmpText)
				next = nReset
				break
			}
			p.ItemFlags[17] = c.ItemFlags[sel]
			p.ItemElem[17] = c.ItemElem[sel]
			p.ItemMat[17] = c.ItemMat[sel]
			p.ItemAttr[17] = c.ItemAttr[sel]
			p.ItemID[17] = c.ItemID[sel]
			p.ItemPower[17] = c.ItemPower[sel]
			c.ItemID[sel] = 0
			p.ItemEq[17] = 0
			compactBackpack(c)
			compactBackpack(p)
			next = nReset
		case 0x1b:
			if mode != 8 {
				next = nOut
			}
		}
		if next == nCommon { // common:
			if mode == 1 {
				if k < 3 {
					cat = k
				}
				saveShopStock(oldcat)
				loadShopStock(cat)
				oldcat = cat
				next = nReset
			} else if mode == 3 || (mode == 4 || mode == 5 || mode == 6) && k != 1 {
				next = nBreak
			} else {
				next = nAct
			}
		}
		if next == nAct { // act:
			switch doItemAction(c, k, sel, top, mode) {
			case 1:
				c = nil
				goto done
			case 2:
				goto out
			}
			if mode == 4 || mode == 5 || mode == 6 {
				goto out
			}
			if c.ItemID[9] == 0 {
				top = 0
			}
			next = nReset
		}
		switch next {
		case nReset:
			memset(mark[:], 0, 40)
			sel = -1
			goto list
		case nList:
			goto list
		case nAgain:
			goto again
		case nOut:
			goto out
		}
		if k == 4 {
			break
		}
	}
out:
	closeWindows(2)
	c = orig
done:
	popIcons()
	popButtons()
	freeAndNull(&icon)
	drawPartyPortraits()
	if mode == 1 {
		saveShopStock(oldcat)
	}
	if party != 0 {
		showPleaseWait()
		loadMazeGfx(int(*gMapNum) - 1)
		closeWindows(1)
	}
	return c
}

func setCharscrHotspots() {
	*gBtnXs = adrCharscrHotX
	*gBtnYs = adrCharscrHotY
	*gBtnWs = adrCharscrHotW
	*gBtnHs = adrCharscrHotH
}

func loadCharscrIcons() {
	freeAndNull(gCurIcons)
	*gCurIcons = loadGraphic("view.icn")
	gCharscrSprites[0].Img.Set(uint32(*gCurIcons))
	gCharscrSprites2[0].Img.Set(uint32(*gCurIcons))
	setCharscrHotspots()
}

func countSkills(c *Chr) int {
	n := 0
	for i := 0; i < 0x12; i++ {
		if c.Skill[i] != 0 {
			n++
		}
	}
	return n
}

func countAwards(c *Chr) int {
	n := 0
	for i := 0; i < 0x1a; i++ {
		if c.Award[i] != 0 {
			n++
		}
	}
	return n
}

func calcMaxSp(c *Chr) int {
	var v, first int32
	var k, sk uint8
	twice := uint8(0)

	if c.Caster != 0 {
		if c.Cls == 4 || c.Cls == 2 {
			k = 1
			sk = 13
		} else {
			k = 2
			sk = 12
		}
		if c.Cls == 8 || c.Cls == 9 {
			sk = 2
		}
	again:
		v = int32(int16(statBonus(getCharStat(c, int(k), 0)) + 3))
		v += int32(int8(gRaceSpBonus[c.Race][k-1]))
		if c.Skill[sk] != 0 {
			v += 2
		}
		if v < 1 {
			v = 1
		}
		v *= int32(uint16(getCharLevel(c)))
		if c.Cls != 4 && c.Cls != 3 && c.Cls != 8 {
			v >>= 1
		}
		if twice == 0 && (c.Cls == 8 || c.Cls == 9) {
			k = 1
			twice = 1
			first = v
			goto again
		}
		if twice != 0 {
			v += first
			v >>= 1
		}
		v += int32(uint16(getItemStatBonus(c, 8)))
		if v < 0 {
			v = 0
		}
	}
	return int(int16(v))
}

func calcMaxHp(c *Chr) int {
	v := int32(gClassHpPerLevel[c.Cls])
	v += int32(int16(statBonus(getCharStat(c, 3, 0))))
	v += int32(int8(gRaceHpBonus[c.Race]))
	if c.Skill[3] != 0 {
		v++
	}
	if v < 1 {
		v = 1
	}
	v *= int32(uint16(getCharLevel(c)))
	v += int32(uint16(getItemStatBonus(c, 7)))
	if v < 0 {
		v = 0
	}
	return int(int16(v))
}

func calcArmorClass(c *Chr, base int) int {
	n := statBonus(getCharStat(c, 4, 0))
	n = int(int16(n + getItemStatBonus(c, 9)))
	if base == 0 {
		n += int(c.Bless)
		n += int(c.AcTmp)
	}
	if n < 1 {
		return 0
	}
	return n
}

func formatCharSheet(c *Chr, buf []byte, n int) {
	cond := int(getWorstCondition(c))
	days := *gFood / uint16(*gPartyCount) / 3
	b := bytesOf(c)
	br := func(k int) int { return int(b[0x102+k]) }
	col := func(kind int) int { return statColor(getCharStat(c, kind, 0), getCharStat(c, kind, 1)) }
	gold := *gPartyGold
	if gPartyIds[n] >= 0x14 {
		gold = uint32(int32(int16(int(c.Level)*int(c.Level))) * 20)
	}
	resis := br(5) + getItemStatBonus(c, 0xb) + br(6) + br(7) + getItemStatBonus(c, 0xd) + br(8) + br(9) +
		getItemStatBonus(c, 0xc) + br(10) + br(11) + getItemStatBonus(c, 0xe) + br(12) + br(13) +
		getItemStatBonus(c, 0xf) + br(14) + br(15) + getItemStatBonus(c, 0x10) + br(16)
	plus := func(k int) string {
		if cond == 16 && br(k) != 0 {
			return "\01417+"
		}
		return ""
	}
	dayS := 's'
	if days == 1 {
		dayS = ' '
	}
	sprintf(buf, "\v041\t196\004080%s\t000\v002\004275%s : %s %s %s %s\003r\t047\v028\004022\014%02u%u\014d\t099\004022\014%02u%u\014d\003l\t131\004037\014%02u%d\014d\t196\004076\01418%lu\014d\003r\t047\v051\004022\014%02u%u\014d\t099\004022\014%02u%u\014d\003l\t131\004037\014%02u%u\014d\t196\004076\01418%lu\014d\003r\t047\v074\004022\014%02u%u\014d\t099\004022\014%02u%u\014d\003l\t131\004037\01418%u\014d\t196\004076\01418%lu\014d\003r\t047\v097\004022\014%02u%u\014d\t099\004022\014%02u%u\014d\003l\t131\004037\01418%u\014d\t196\004076\01418%u day%c\014d\003r\t047\v120\004022\014%02u%u\014d\t099\004022\014%02u%u\014d\003l\t131\004037\01418%u\014d\t196\004080\014%02u%s\014d\t230%s%s%s%s\014d",
		str(gStrCostPartyGold[b2i(gPartyIds[n] < 0x14)]), c.Name[:], str(gSexNames[c.Sex]), str(gAlignNames[c.Align]), str(gRaceNames[c.Race]), str(gClassNames[c.Cls]),
		col(0), getCharStat(c, 0, 0),
		col(5), getCharStat(c, 5, 0),
		statColor(c.HP.Get(), calcMaxHp(c)), c.HP.Get(), getTotalExp(c),
		col(1), getCharStat(c, 1, 0),
		col(6), getCharStat(c, 6, 0),
		statColor(c.SP.Get(), calcMaxSp(c)), c.SP.Get(),
		gold,
		col(2), getCharStat(c, 2, 0),
		statColor(getCharAge(c, 0), getCharAge(c, 1)), getCharAge(c, 0),
		resis,
		*gPartyGems,
		col(3), getCharStat(c, 3, 0),
		statColor(getCharLevel(c), int(c.Level)), getCharLevel(c),
		countSkills(c), days, dayS,
		col(4), getCharStat(c, 4, 0),
		statColor(calcArmorClass(c, 0), calcArmorClass(c, 1)), calcArmorClass(c, 0),
		countAwards(c), gConditionColors[cond], str(gConditionNames[cond]),
		plus(0), plus(1), plus(2), plus(3))
}

func statColor(cur, max int) int {
	if cur < 1 {
		return 0x15
	}
	if cur > max {
		return 3
	}
	if cur == max {
		return 0x12
	}
	if max>>2 > cur {
		return 7
	}
	return 0xc
}

func showQuickRef() {
	var c *Chr
	var lines [8][100]byte

	pushButtons()
	setMouseCursor(0, uint32(*gMouseCursor))
	setAnykeyButtons()
	for *gMouseX = 0; int(*gMouseX) < int(*gPartyCount); *gMouseX++ {
		mx := int(*gMouseX)
		c = &gParty[mx]
		cond := int(getWorstCondition(c))
		cn, nm := str(gClassNames[c.Cls]), str(gConditionNames[cond])
		sprintf(lines[mx][:], "\v%3d\t007%u)\t027%s\t110%c%c%c\003r\t160\014%02u%u\014d\003l\t176\014%02u%d\014d\t212\014%02u%u\014d\t241\014%02u%u\014d\t270\014%02u%c%c%c%c\014d",
			mx*10+0x18, mx+1, bytesOf(c), cn[0], cn[1], cn[2],
			statColor(getCharLevel(c), int(c.Level)), getCharLevel(c),
			statColor(c.HP.Get(), calcMaxHp(c)), c.HP.Get(),
			statColor(c.SP.Get(), calcMaxSp(c)), c.SP.Get(),
			statColor(calcArmorClass(c, 0), calcArmorClass(c, 1)), calcArmorClass(c, 0),
			gConditionColors[cond], nm[0], nm[1], nm[2], nm[3])
	}
	days := int(*gFood / uint16(*gPartyCount) / 3)
	s := "s"
	if days == 1 {
		s = ""
	}
	sprintf(gTextBuf, "\003cQuick Reference Chart\v012\003l\t007#\t027Name\t110Cls\t145Lvl\t176H.P.\t212S.P.\t241A.C.\t270Cond%s%s%s%s%s%s%s%s\v110\t064\003cGold\t144Gems\t224Food\v119\t064\01418%lu\t144%lu\t224%u day%s\014d",
		lines[0][:], lines[1][:], lines[2][:], lines[3][:], lines[4][:], lines[5][:], lines[6][:], lines[7][:], *gPartyGold, *gPartyGems, days, s)
	openTextWindow(0, 0, 0x140, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	flushInput()
	getInput()
	closeWindows(1)
	waitMouseRelease()
	popButtons()
}

var classCanUseBuf [100]byte // class_can_use_item's buf (passed to the window as a far pointer)

func classCanUseItem(cls, id, quiet int) int {
	switch cls {
	case 0, 1:
		return 1
	case 2, 3, 4, 5, 6, 7, 8, 9:
		if int(gItemClassMasks[id])&int(gClassBits[cls-2]) == 0 {
			return 1
		}
	}
	if quiet == 0 {
		pushButtons()
		setAnykeyButtons()
		s := ""
		if id < 0x22 || id > 0x29 {
			s = "'s"
		}
		sprintf(classCanUseBuf[:], "\003c\v007%ss are not proficient with %s\b %s!", str(gClassNames[cls]), str(gItemNames[id]), s)
		openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, farOf(classCanUseBuf[:]))
		waitMouseRelease()
		getInput()
		closeWindows(1)
		popButtons()
		waitMouseRelease()
	}
	return 0
}

func fixIdentifyPrompt(c *Chr, slot, mode, ident, flags int) int32 {
	cost := getItemPrice(c, slot, mode, flags)
	if c.ItemID[slot] > 0x49 {
		cost = 1
	}
	s := "Fix"
	if ident != 0 {
		s = "Identify"
	}
	sprintf(gTextBuf, "%s %s for %lu gold?", s, formatItemName(c, slot, gTmpText, 0), cost)
	return cost
}

func drawItemPageArrows(c *Chr, top int) {
	var up, down int
	if top == 0 {
		up = 11
	} else {
		up = 10
	}
	if c.ItemID[9] == 0 || top != 0 {
		down = 13
	} else {
		down = 12
	}
	drawIcon(*gCurIcons, up, 0xf, 0x6d, 0)
	drawIcon(*gCurIcons, down, 0x35, 0x6d, 0)
}

func getItemPrice(c *Chr, slot, mode, flags int) int32 {
	var base, m2, m3, m4 uint32
	var res int32
	var ff int

	id := c.ItemID[slot]
	if id > 0x49 && id != 0x4b && id != 0x52 {
		return 0
	}
	if flags > 0x80 && mode == 2 {
		ff = 1
	} else {
		ff = flags & 0x80
	}
	flags &= 0x7f
	if id == 0x4b {
		base = 2000
	} else if id == 0x52 {
		base = 1000
	} else {
		base = uint32(gItemBasePrices[id])
	}
	switch c.ItemMat[slot] {
	case 1:
		base /= 10
	case 2:
		base >>= 2
	case 3:
		base >>= 1
	case 4:
		base -= base >> 2
	default:
		base *= uint32(int32(int8(gMaterialPriceMult[c.ItemMat[slot]])))
	}
	m2 = uint32(int32(int16(int(gElemDmgBonus[c.ItemElem[slot]]) * 100)))
	m3 = uint32(int32(int16(int(int8(gAttribBonusValues[c.ItemAttr[slot]])) * 100)))
	m4 = uint32(gPowerPrices[c.ItemPower[slot]])
	switch mode {
	case 1, 2:
		if c.ItemFlags[slot]&0xc0 != 0 && ff != 0 {
			base, m2, m3, m4 = 0, 0, 0, 0
		}
		res = int32((base + m2 + m3 + m4) / uint32(gPriceDivisors[flags]))
	case 3, 4, 5, 6:
		res = int32(c.ItemFlags[slot] & 0x3f)
	}
	if mode != 0 {
		return res
	}
	return 0
}

// itemStatSuffix is a port addition: Angband-style stats after a weapon's or armor's name,
// " (XdY) (+hit,+dmg)" and " [AC,+bonus]", with formatItemInfo's numbers (the material adds
// to hit, to damage and to armor class). The bonus part is left out when it is +0.
func itemStatSuffix(c *Chr, slot int) []byte {
	id, mat := int(c.ItemID[slot]), c.ItemMat[slot]
	var s [40]byte
	switch {
	case id > 0 && id < 0x22:
		sprintf(s[:], " (%dd%d)", gWeaponDiceCount[id], gWeaponDieSides[id])
		if h, d := int(int8(gMaterialToHit[mat])), int(int8(gMaterialDmgBonus[mat])); h|d != 0 {
			sprintf(s[strlen(s[:]):], " (%+d,%+d)", h, d)
		}
	case id < 0x4a && gItemAcValues[id] != 0:
		if b := int(int8(gMaterialAcBonus[mat])); b != 0 {
			sprintf(s[:], " [%d,%+d]", gItemAcValues[id], b)
		} else {
			sprintf(s[:], " [%d]", gItemAcValues[id])
		}
	}
	return append([]byte(nil), s[:strlen(s[:])+1]...)
}

func formatItemName(c *Chr, slot int, buf []byte, n int) []byte {
	broken := c.ItemFlags[slot]&0xc0 != 0
	pick := func(cond bool, s []byte) []byte {
		if cond {
			return nil
		}
		return s
	}
	var br, cu, of, bs string
	if c.ItemFlags[slot]&0x80 != 0 {
		br = "\01407broken "
	}
	if c.ItemFlags[slot]&0x40 != 0 {
		cu = "\01412cursed "
	}
	if !broken && c.ItemPower[slot] != 0 {
		of = "of "
	} else {
		bs = "\b "
	}
	sprintf(buf, "%s%s%s%s%s\014%02u%s%s%s%s",
		pick(broken, str(gElemPrefixNames[c.ItemElem[slot]])), pick(broken, str(gMaterialNames[c.ItemMat[slot]])), pick(broken, str(gAttribPrefixNames[c.ItemAttr[slot]])),
		br, cu, n, str(gItemNames[c.ItemID[slot]]), of, pick(broken, str(gPowerNames[c.ItemPower[slot]])), bs)
	if buf[0] == '\014' {
		buf[3] = byte(toupper(int(buf[3])))
	} else {
		buf[0] = byte(toupper(int(buf[0])))
	}
	if !broken {
		strcat(buf, itemStatSuffix(c, slot))
	}
	return buf
}
