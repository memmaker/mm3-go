// Port of ITEMS.C (overlay 1861): the item actions (equip, remove, use, discard; sell, repair,
// identify, buy; the recharge, duplicate and enchant spells), the item info text, the shop stock
// tables, the compass button, door bashing, the power-up animation over the monsters and the
// cast-spell dialog.
//
// items' struct chr maps onto Chr by offset: sk_thievery is Skill[0], sk_merchant Skill[8].
package mm3

var (
	cantRemoveCursedText = cs("\x03cYou cannot remove a cursed item!")
	pleaseWaitText       = cs("\x03c\x0b002Please Wait...")
)

func doItemAction(c *Chr, cat, slot, base, act int) int {
	var sv Far
	var a, n, f48, f64, f63 int
	var price int32
	defer dsFree(dsSP)

	if c.ItemID[0] == 0 {
		return 0
	}
	*gPortraitClickOn = 0
	if (slot < 0 || slot > 0x11) && act != 1 {
		gItemscrHotX[0] = 0xeb
		*gItemscrHotX1 = -1
		gItemscrHotY[0] = 0x6f
		if *gCurIcons != 0 {
			sv = *gCurIcons
		}
		adrS := dsAlloca(0x14)
		s := arr[int16](adrS)
		s[0], s[7] = -1, -1
		s[3] = 0xb1
		*gCurIcons = loadGraphic("esc.icn")
		setLe32(ds[:], adrS+2, uint32(*gCurIcons))
		sprintf(gTextBuf, "\t010\x0b005%s which item?\x05%04X", str(gItemActionLabels[cat][act]), adrS)
		openTextWindow(0x32, 0x67, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		for {
			slot = getInput()
			if slot == 0 || slot == 0x1b {
				slot = -1
				break
			}
			if slot >= 0x31 && slot <= 0x39 {
				slot -= 0x2a
			}
			if slot >= 7 && slot <= 0xf {
				slot = slot + base - 7
				if c.ItemID[slot] == 0 {
					continue
				}
				break
			}
		}
		closeWindows(1)
		freeAndNull(gCurIcons)
		*gCurIcons = sv
		gItemscrHotX[0] = 0x5b
		*gItemscrHotX1 = 0x81
		gItemscrHotY[0] = 0x6d
		if slot == -1 {
			goto out
		}
	}
	switch act {
	case 0, 8:
		if c.ItemID[slot] != 0 {
			switch cat {
			case 0:
				if c.ItemID[slot] > 0x47 {
					sprintf(gTextBuf2, "%s is not equippable!", str(gItemNames[c.ItemID[slot]]))
					showMessage(gTextBuf2)
				} else {
					equipItem(c, slot)
				}
			case 1:
				if c.ItemFlags[slot]&0x40 != 0 {
					showMessage(cantRemoveCursedText)
				} else {
					c.ItemEq[slot] = 0
				}
			case 2:
				if *gMapNum == 0 {
					showMessage(cs("\x0b007What's your hurry?\nWait till you get out of here!"))
					break
				}
				switch getWorstCondition(c) {
				case 8, 11, 12, 13, 14, 15:
					sprintf(gTmpText, gostr(str(*gMsgCantAct)), bytesOf(c))
					showMessage(gTmpText)
					*gPortraitClickOn = 1
				default:
					if *gItemsInCombat != 0 {
						showMessage(cs("\x0b007To use an item in Combat, invoke the Use command on your turn!"))
						*gPortraitClickOn = 1
						break
					}
					id := c.ItemID[slot]
					f48 = b2i(id == 0x48)
					f64 = b2i(id == 0x64)
					f63 = b2i(id == 0x63)
					if (c.ItemEq[slot] != 0 || id == 0x63 || (id >= 0x46 && id <= 0x48) || f64 != 0) && c.ItemFlags[slot]&0xc0 == 0 {
						if c.ItemPower[slot] != 0 || f48 != 0 || f64 != 0 || f63 != 0 {
							n = int(c.ItemFlags[slot] & 0x3f)
							if id == 0x63 {
								n = 2
							}
							n--
							c.ItemFlags[slot] = uint8(n)
							*gCurChar = nearOf(c)
							if *gGameMode == 2 {
								closeWindows(2)
							} else {
								closeWindows(3)
							}
							switch {
							case f48 != 0:
								useItemPower(1)
							case f64 != 0:
								useItemPower(0x4c)
							case f63 != 0:
								useItemPower(0x4d)
							default:
								useItemPower(int(c.ItemPower[slot]))
							}
							if n < 1 {
								clearItemSlot(c, slot)
								compactBackpack(c)
							}
							*gCombatActor = -1
							*gPortraitClickOn = 1
							return 1
						}
						sprintf(gTextBuf, "%shas no special abilities!", formatItemName(c, slot, gTmpText, 0))
					} else {
						sprintf(gTextBuf, "%s needs to be equipped to be used!", formatItemName(c, slot, gTmpText, 0))
					}
					showMessage(gTextBuf)
					*gPortraitClickOn = 1
				}
			case 3:
				if isQuestItem(int(c.ItemID[slot])) == 0 && c.ItemFlags[slot]&0x40 == 0 {
					sprintf(gTextBuf, "Permanently discard %s?", formatItemName(c, slot, gTmpText, 0))
					if askYesNo(gTextBuf, 0) == 0 {
						break
					}
					clearItemSlot(c, slot)
					compactBackpack(c)
					if act == 8 {
						return 2
					}
				} else if c.ItemFlags[slot]&0x40 != 0 {
					showMessage(cs("\x03cYou cannot discard a cursed item!"))
				} else {
					sprintf(gTextBuf, "The %s is necessary and cannot be discarded!", formatItemName(c, slot, gTmpText, 0))
					showMessage(gTextBuf)
				}
			}
		}
	case 2:
		if c.ItemID[slot] != 0 {
			switch cat {
			case 0:
				if (c.ItemID[slot] < 0x46 && c.ItemFlags[slot]&0x40 == 0) || c.ItemID[slot] == 0x4b || c.ItemID[slot] == 0x52 {
					price = getItemPrice(c, slot, act, b2i(c.Skill[8] == 0)|0x80)
					sprintf(gTextBuf, "Sell %s for %lu gold?", formatItemName(c, slot, gTmpText, 0), price)
					if askYesNo(gTextBuf, 0) == 0 {
						break
					}
					clearItemSlot(c, slot)
					*gPartyGold += uint32(price)
					compactBackpack(c)
				} else {
					sprintf(gTextBuf, "We have no need of this %s!", formatItemName(c, slot, gTmpText, 0))
					showMessage(gTextBuf)
				}
			case 1:
				if c.ItemFlags[slot]&0x80 == 0 {
					showMessage(cs("That item is not broken!"))
					break
				}
				price = fixIdentifyPrompt(c, slot, act, 0, cat)
				if askYesNo(gTextBuf, 0) == 0 {
					break
				}
				if payCost(0, uint32(price), 0) == 0 {
					break
				}
				c.ItemFlags[slot] &= 0x7f
			}
		}
	case 1:
		if c.ItemID[slot] != 0 {
			cc := curChr()
			if cc.ItemID[0x11] != 0 {
				sprintf(gTmpText, "\x03c%s's backpack is full.", bytesOf(cc))
				showMessage(gTmpText)
				break
			}
			price = getItemPrice(c, slot, act, 0)
			sprintf(gTextBuf, "Buy %s for %lu gold?", formatItemName(c, slot, gTmpText, 0), price)
			if askYesNo(gTextBuf, 0) == 0 {
				break
			}
			if payCost(0, uint32(price), 0) == 0 {
				break
			}
			cc.ItemFlags[0x11] = c.ItemFlags[slot]
			cc.ItemElem[0x11] = c.ItemElem[slot]
			cc.ItemMat[0x11] = c.ItemMat[slot]
			cc.ItemAttr[0x11] = c.ItemAttr[slot]
			cc.ItemID[0x11] = c.ItemID[slot]
			cc.ItemPower[0x11] = c.ItemPower[slot]
			clearItemSlot(c, slot)
			compactBackpack(c)
			compactBackpack(cc)
		}
	case 4:
		if c.ItemID[slot] != 0 {
			if c.ItemID[slot] <= 0x47 && c.ItemPower[slot] != 0 {
				a = randomRange(1, 6)
				playSfx(0x14)
				if int(c.ItemFlags[slot]&0xbf)+a > 0x3f {
					c.ItemFlags[slot] |= 0x3f
				} else {
					c.ItemFlags[slot] += uint8(a)
				}
			} else { // norech:
				playSfx(0x15)
				sprintf(gTmpText, "\x03cNot Recharagable.  %s", str(*gMsgSpellFailed))
				showMessage(gTmpText)
			}
		}
	case 5:
		if c.ItemID[slot] != 0 {
			if c.ItemID[0x11] != 0 {
				sprintf(gTmpText, "\x03cBackpack Full.  %s", str(*gMsgSpellFailed))
				showMessage(gTmpText)
				break
			}
			tooPow := false
			switch c.ItemAttr[slot] {
			case 9, 10, 0x11, 0x12, 0x19, 0x1a, 0x21, 0x22, 0x27, 0x28,
				0x2d, 0x2e, 0x32, 0x33, 0x38, 0x39, 0x3d, 0x3e, 0x47, 0x48:
				tooPow = true
			default:
				switch c.ItemElem[slot] {
				case 7, 8, 14, 15, 19, 20, 24, 25, 32, 33, 36:
					tooPow = true
				}
				if !tooPow && (c.ItemMat[slot] > 0x11 || c.ItemPower[slot] > 0x37 || c.ItemID[slot] > 0x48) {
					tooPow = true
				}
				if !tooPow {
					c.ItemFlags[0x11] = c.ItemFlags[slot] & 0xc0
					c.ItemElem[0x11] = c.ItemElem[slot]
					c.ItemMat[0x11] = c.ItemMat[slot]
					c.ItemAttr[0x11] = c.ItemAttr[slot]
					c.ItemID[0x11] = c.ItemID[slot]
					c.ItemPower[0x11] = c.ItemPower[slot]
					compactBackpack(c)
					playSfx(0x14)
				}
			}
			if tooPow { // toopow:
				sprintf(gTmpText, "\x03cItem too Powerful.  %s", str(*gMsgSpellFailed))
				playSfx(0x15)
				showMessage(gTmpText)
			}
		}
	case 6:
		if c.ItemID[slot] != 0 {
			if c.ItemElem[slot] != 0 || c.ItemMat[slot] != 0 || c.ItemAttr[slot] != 0 || c.ItemPower[slot] != 0 || c.ItemID[slot] > 0x47 {
				sprintf(gTmpText, "\x03cNot Enchantable.  %s", str(*gMsgSpellFailed))
				playSfx(0x15)
				showMessage(gTmpText)
				break
			}
			a = int(c.ItemID[slot])
			f63 = randomRange(1, getCharLevel(curChr())/10+1)
			if f63 > 5 {
				f63 = 5
			}
			generateRandomItem(f63, c, slot)
			c.ItemID[slot] = uint8(a)
			playSfx(0x14)
		}
	}
out:
	*gPortraitClickOn = 1
	return 0
}

// showItemInfo is the Smithy's old Identify window, now free on a right-click in any item
// list (port addition; the paid Identify service is gone).
func showItemInfo(c *Chr, slot int) {
	formatItemInfo(slot, c)
	sprintf(gTextBuf, "\x03cItem Info\x03l\n\n\x0b012\x0c22%s\n\n%s", formatItemName(c, slot, gTmpText, 0), gDrawListB)
	openTextWindow(0, 7, 0x140, 0x83, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	waitMouseRelease() // the right-click that opened it
	pushButtons()
	setAnykeyButtons()
	getInput()
	popButtons()
	closeWindows(1)
}

func formatItemInfo(slot int, c *Chr) {
	var i, k int8
	var s0, s1, s2, s3 [40]byte
	var s4, s5, s6 [10]byte
	out := gDrawListB

	id := int(c.ItemID[slot])
	mat := c.ItemMat[slot]
	memset(s0[:], 0, 0x28)
	for i, k = 0, 0; i <= 9; i++ {
		if classCanUseItem(int(i), id, 1) != 0 {
			cn := str(gClassNames[i])
			s0[k] = cn[0]
			k++
			s0[k] = cn[1]
			k++
			s0[k] = ' '
			k++
		}
	}
	if isQuestItem(id) != 0 {
		strcpy(s0[:], cs("None"))
	} else if k == 0x1e {
		strcpy(s0[:], cs("All"))
	}
	v30 := int(gWeaponDiceCount[id]) + int(int8(gMaterialDmgBonus[mat]))
	v26 := uint16(int(gWeaponDiceCount[id])*int(gWeaponDieSides[id]) + int(int8(gMaterialDmgBonus[mat])))
	v28 := int(gElemDmgBonus[c.ItemElem[slot]])
	v2a := int(gElemResistBonus[c.ItemElem[slot]])
	v2c := int(int8(gAttribBonusValues[c.ItemAttr[slot]]))
	v2e := int(gItemAcValues[id])
	if id >= 0x22 {
		v2e += int(int8(gMaterialAcBonus[mat]))
	}
	if id > 0x49 {
		if id == 0x64 {
			strcpy(s1[:], cs("+5 Might"))
		} else {
			strcpy(s1[:], cs("None"))
		}
		strcpy(s2[:], cs("None"))
		strcpy(s3[:], cs("None"))
		strcpy(s4[:], cs("None"))
		strcpy(s5[:], cs("None"))
		strcpy(s6[:], cs("None"))
	} else {
		sprintf(s1[:], "%+d %s", v2c, str(gAttribNames[getBonusStatIndex(int(c.ItemAttr[slot]))]))
		i = int8(getElementIndex(int(c.ItemElem[slot])))
		sprintf(s2[:], "%+d %s Damage", v28, str(gDamageTypeNames[i]))
		sprintf(s3[:], "%+d %s", v2a, str(gDamageTypeNames[i]))
		lo := v30
		if v30 < 1 {
			lo = 0
		}
		sprintf(s4[:], "%d to %d", lo, v26)
		sprintf(s5[:], "%+d", int8(gMaterialToHit[mat]))
		sprintf(s6[:], "%+d", v2e)
	}
	none := cs("None")
	or := func(cond bool, s []byte) []byte {
		if cond {
			return s
		}
		return none
	}
	sprintf(out, "Proficient Classes\t132:\t140%s\nto Hit Modifier\t132:\t140%s\nPhysical Damage\t132:\t140%s\n"+
		"Elemental Damage\t132:\t140%s\nElemental Resistance\t132:\t140%s\nArmor Class Bonus\t132:\t140%s\n"+
		"Attribute Bonus\t132:\t140%s\nSpecial Power\t132:\t140%s",
		s0[:],
		or(gMaterialToHit[mat] != 0 && id < 0x22, s5[:]),
		or(id != 0 && id < 0x22 && v26 > 0, s4[:]),
		or(c.ItemElem[slot] != 0 && id < 0x22, s2[:]),
		or(v2a != 0, s3[:]),
		or(v2e != 0 && id > 0x21, s6[:]),
		or(v2c != 0, s1[:]),
		str(gSpellNames[gPowerSpellIds[c.ItemPower[slot]]]))
}

func isQuestItem(id int) int {
	for i := 0; i < 0x10; i++ {
		if int(gQuestItemIds[i]) == id {
			return 1
		}
	}
	return 0
}

func showPleaseWait() {
	openTextWindow(0x67, 0x9c, 0x72, 0x1e, 0, 0x96, 4, 6, farOf(pleaseWaitText))
}

func getThievery(c *Chr) int {
	v := getCharLevel(c) << 1
	if c.Cls == 6 {
		v += 0xf
	} else if c.Cls == 5 {
		v += 0x1e
	}
	if c.Race == 1 || c.Race == 3 {
		v += 10
	} else if c.Race == 2 {
		v += 5
	} else if c.Race == 4 {
		v -= 10
	}
	v = int(int16(v + getItemStatBonus(c, 10)))
	if c.Skill[0] == 0 {
		v = 0
	}
	if v < 1 {
		return 0
	}
	return v
}

func saveShopStock(n int) {
	t := *gCurTown
	for i := 0; i < 9; i++ {
		gShopStockFlags[t][n][i] = gShopInvFlags[i]
		gShopStockElems[t][n][i] = gShopInvElems[i]
		gShopStockMaterials[t][n][i] = gShopInvMaterials[i]
		gShopStockAttribs[t][n][i] = gShopInvAttribs[i]
		gShopStockIds[t][n][i] = gShopInvIds[i]
		gShopStockPowers[t][n][i] = gShopInvPowers[i]
	}
}

func loadShopStock(n int) {
	t := *gCurTown
	for i := 0; i < 9; i++ {
		gShopInvFlags[i] = gShopStockFlags[t][n][i]
		gShopInvElems[i] = gShopStockElems[t][n][i]
		gShopInvMaterials[i] = gShopStockMaterials[t][n][i]
		gShopInvAttribs[i] = gShopStockAttribs[t][n][i]
		gShopInvIds[i] = gShopStockIds[t][n][i]
		gShopInvPowers[i] = gShopStockPowers[t][n][i]
	}
}

func getItemEquipCat(c *Chr, id int) int {
	if id > 0x48 {
		return 2
	}
	if classCanUseItem(int(c.Cls), id, 1) == 0 {
		return 0
	}
	switch {
	case id <= 0x11:
		return 1
	case id <= 0x1d:
		return 0xd
	case id <= 0x21:
		return 4
	case id <= 0x29:
		return 3
	case id <= 0x2a:
		return 2
	case id <= 0x2d:
		return 5
	case id <= 0x2e:
		return 6
	case id <= 0x2f:
		return 8
	case id <= 0x30:
		return 9
	case id <= 0x33:
		return 0xa
	case id <= 0x34:
		return 0xc
	case id <= 0x39:
		return 7
	case id <= 0x3c:
		return 0xb
	}
	return 2
}

func drawCompassButton(k int) {
	defer dsFree(dsSP)
	adr := dsAlloca(40)
	p := arr[int16](adr)
	p[0] = -1
	setLe32(ds[:], adr+2, uint32(*gGlobalIcons))
	p[3], p[4], p[5], p[6] = 0x64, 0x7c, 0, int16(k+0xc)
	p[7], p[8], p[9] = 0x6e, 0x83, 0
	if partyHasSkill(6) != 0 {
		if k != 0 {
			p[10] = int16(0x14 + int(*gPartyDir))
		} else {
			p[10] = int16(0xf + int(*gPartyDir))
		}
	} else if k != 0 {
		p[10] = 0x13
	} else {
		p[10] = 0xe
	}
	p[11] = -1
	setLe32(ds[:], adr+24, 0)
	sprintf(gTextBuf2, "\x03d\x05%04X", adr)
	printWindowText(gTextBuf2)
}

func bashDoor(x, y, dir int) {
	var ny, a, b, thr, i, v int

	if *gOutdoors != 0 {
		return
	}
	playSfx(0x19)
found:
	for i, a, b = 0, 0, 0; i < int(*gPartyCount); i++ {
		switch getWorstCondition(&gParty[i]) {
		case 8, 11, 12, 13, 14, 15:
		default:
			if a == 0 {
				a = i + 1
			} else {
				b = i + 1
				break found
			}
		}
	}
	charTakeDamage(&gParty[a-1], 2)
	drawIcon(*gCharpowSprites, 0, (a-1)*0x26+0xb, 0x95, 0)
	if b != 0 {
		drawIcon(*gCharpowSprites, 0, (b-1)*0x26+0xb, 0x95, 0)
		charTakeDamage(&gParty[b-1], 2)
	}
	if getCellWallsRegion(int(int8(gViewDxFrom7[dir][0]))+x, int(int8(gViewDyFrom7[dir][0]))+y, 0x7777) != 0x1111 {
		v = getCharStat(&gParty[a-1], 0, 0)
		if b != 0 {
			v += getCharStat(&gParty[b-1], 0, 0)
		}
		v += randomRange(1, 0x1e)
		d := *gPartyDir
		switch getCellWallsRegion(int(*gPartyX), int(*gPartyY), int(gViewWallMasksFrom2[d][0])) >> gViewWallShiftsFrom2[d][0] {
		case 4:
			thr = int(gMapDoorStrengths[*gMapQuadrant][1])
		case 2:
			thr = int(gMapDoorStrengths[*gMapQuadrant][0])
		default:
			thr = int(gMapDoorStrengths[*gMapQuadrant][2])
		}
		if v >= thr {
			setCellWall(x, y, dir, 5)
			i = x
			ny = y
			switch dir {
			case 0:
				ny = y + 1
				v = 1
			case 1:
				ny = y - 1
				v = 0
			case 2:
				i = x + 1
				v = 3
			case 3:
				i = x - 1
				v = 2
			}
			setCellWall(i, ny, v, 5)
		}
	}
	updatePartyDisabled()
	idleFramesOrKey(2)
	drawPartyPortraits()
}

func spellAttackMonsters(n int) {
	var s4b7c, found, sf02b, f, i int
	var buf [3]uint8
	var k uint8

	*gPartyCellNoTerrain = uint8(b2i(getCellWallsRegion(int(*gPartyX), int(*gPartyY), 0x70) == 0))
	if n == 11 {
		*gSpellHitsParty = 1
		if *gAttackType == 0 {
			*gAttackDamage = 0
			for i, found = 0, 0; i < int(*gPartyCount); i++ {
				if charCanShoot(&gParty[i]) != 0 {
					gMissilePos[i] = 1
					found = 1
				}
			}
		} else {
			gMissilePos[0] = 1
			found = 1
		}
		if found != 0 {
			playSfx(0xc)
		} else {
			playSfx(0x15)
			return
		}
	} else {
		gMissilePos[0] = 1
		*gSpellHitsParty = 0
	}
	*gSpellAnimActive = 1
	sprintf(gTmpText, "pow%d.icn", n)
	if *gPowIconFar != 0 {
		freeAndNull(gPowIconFar)
	}
	updateFreeMemKb()
	if *gFreeMemKb > 7 {
		*gPowIconFar = loadGraphic(gostr(gTmpText))
	} else if *gFreeMemKb != 0 {
		*gPowIconFar = loadGraphic("pow11.icn")
	}
	sf02b = int(*gTargetSlot)
	s4b7c = int(*gTargetMon)
	f, k = 0, 0
	fill := func(m1, m2, m3 *uint8) {
		k = 0
		memset(buf[:], 0, 3)
		for _, m := range []*uint8{m1, m2, m3} {
			if *m != 0 {
				buf[k] = *m
				k++
			}
		}
	}
	// hit attacks the monsters in buf; true: the attack ends here (clear_party_attack_flags, restore)
	hit := func(resistFirst bool) bool {
		for k = 0; k < 3; k++ {
			*gTargetSlot++
			if buf[k] != 0 {
				if resistFirst {
					memset(gPartyResisted, 0, 8)
					*gTargetMon = buf[k] - 1
				} else {
					*gTargetMon = buf[k] - 1
					memset(gPartyResisted, 0, 8)
				}
				partyAttackMonster(curChr(), 1)
				f = 1
				if *gAttackScope == 0 {
					return true
				}
			} else {
				refreshView()
			}
		}
		return f != 0 && *gAttackScope == 1
	}
	fill(gMonD0c1, gMonD0c2, gMonD0c3)
	*gTargetSlot = 0xff
	if *gTargetMon != 0xff {
		*gTargetSlot = uint8(sf02b - 1)
		buf[0] = uint8(s4b7c + 1)
	}
	refreshView()
	if hit(true) {
		clearPartyAttackFlags()
		goto restore
	}
	for _, m := range [][3]*uint8{{gMonD1c1, gMonD1c2, gMonD1c3}, {gMonD2c1, gMonD2c2, gMonD2c3}, {gMonD3c1, gMonD3c2, gMonD3c3}} {
		if *gPowIconFar == 0 {
			clearPartyAttackFlags()
			goto restore
		}
		fill(m[0], m[1], m[2])
		if hit(false) {
			clearPartyAttackFlags()
			goto restore
		}
	}
restore:
	*gTargetMon = uint8(s4b7c)
	*gTargetSlot = uint8(sf02b)
	awardFoundLoot()
}

func castSpellDialog(arg int) int {
	var key, mode, res, s17f, sp, r int
	var p *Chr

	mode = int(*gGameMode)
	*gGameMode = 3
	res = 0
	s17f = int(*gMonsterScanDelay)
	*gMonsterScanDelay = 0
	// engaged is a travel spell's refusal in combat; true: goto out
	engaged := func() bool {
		spellRefundCost(curChr(), int(p.ReadySpell))
		res = 0
		if arg != 0 {
			return true
		}
		showCantCastEngaged(int(p.ReadySpell))
		return false
	}
	// cast is the C's cast: label; true: goto out
	cast := func() bool {
		if r = spellPayCost(p, int(p.ReadySpell)); r != 0 {
			if arg != 0 {
				return true
			}
			showNotEnoughToCast(int(p.ReadySpell), r)
			return false
		}
		res = 1
		*gCurChar = nearOf(p)
		*gGameMode = uint8(mode)
		switch p.ReadySpell {
		case 0x0:
			spellLight()
		case 0x1:
			spellAwaken()
		case 0x2:
			spellFirstAid()
		case 0x3:
			spellFlyingFist()
		case 0x4:
			spellDetectMagic()
		case 0x5:
			spellElementalArrow()
		case 0x6:
			spellRevitalize()
		case 0x7:
			spellCureWounds()
		case 0x8:
			spellSparks()
		case 0x9:
			spellEnergyBlast()
		case 0xa:
			spellSleep()
		case 0xb:
			spellPain()
		case 0xc:
			spellCreateRope()
		case 0xd:
			spellToxicCloud()
		case 0xe:
			spellSuppressPoison()
		case 0xf:
			spellProtFromElements()
		case 0x10:
			spellTurnUndead()
		case 0x11:
			if mode != 2 {
				spellJump()
			} else if engaged() {
				return true
			}
		case 0x12:
			spellAcidStream()
		case 0x13:
			spellSuppressDisease()
		case 0x14:
			spellSilence()
		case 0x15:
			spellBlessed()
		case 0x16:
			spellLevitate()
		case 0x17:
			if mode != 2 {
				spellWizardEye()
			} else if engaged() {
				return true
			}
		case 0x18:
			spellIdentifyMonster()
		case 0x19:
			spellHolyBonus()
		case 0x1a:
			spellPowerCure()
		case 0x1b:
			spellNaturesCure()
		case 0x1c:
			spellLightningBolt()
		case 0x1d:
			spellImmobilize()
		case 0x1e:
			spellHeroism()
		case 0x1f:
			spellWalkOnWater()
		case 0x20:
			spellFrostBite()
		case 0x21:
			if mode != 2 {
				spellLloydsBeacon()
			} else if engaged() {
				return true
			}
		case 0x22:
			spellPowerShield()
		case 0x23:
			spellCurePoison()
		case 0x24:
			spellFireball()
		case 0x25:
			spellDetectMonster()
		case 0x26:
			spellAcidSpray()
		case 0x27:
			spellColdRay()
		case 0x28:
			spellCureDisease()
		case 0x29:
			if mode != 2 {
				spellNaturesGate()
			} else if engaged() {
				return true
			}
		case 0x2a:
			spellTimeDistortion()
		case 0x2b:
			spellFeebleMind()
		case 0x2c:
			spellDeadlySwarm()
		case 0x2d:
			if mode != 2 {
				spellTeleport()
			} else if engaged() {
				return true
			}
		case 0x2e:
			spellFingerOfDeath()
		case 0x2f:
			spellCureParalysis()
		case 0x30:
			spellParalyze()
		case 0x31:
			spellDragonBreath()
		case 0x32:
			if mode != 2 {
				spellSuperShelter()
			} else if engaged() {
				return true
			}
		case 0x33:
			spellFieryFlail()
		case 0x34:
			spellCreateFood()
		case 0x35:
			if mode != 2 {
				spellTownPortal()
			} else if engaged() {
				return true
			}
		case 0x36:
			spellStoneToFlesh()
		case 0x37:
			spellRechargeItem()
		case 0x38:
			spellFantasticFreeze()
		case 0x39:
			spellDuplication()
		case 0x3a:
			spellDisintegrate()
		case 0x3b:
			spellRaiseDead()
		case 0x3c:
			spellHalfForMe()
		case 0x3d:
			if mode != 2 {
				spellEtherealize()
			} else if engaged() {
				return true
			}
		case 0x3e:
			spellDancingSword()
		case 0x3f:
			spellPrismaticLight()
		case 0x40:
			spellMoonRay()
		case 0x41:
			spellMassDistortion()
		case 0x42:
			spellEnchantItem()
		case 0x43:
			spellIncinerate()
		case 0x44:
			spellElementalStorm()
		case 0x45:
			spellHolyWord()
		case 0x46:
			spellResurrect()
		case 0x47:
			spellMegaVolts()
		case 0x48:
			spellInferno()
		case 0x49:
			spellSunRay()
		case 0x4a:
			spellImplosion()
		case 0x4b:
			spellStarBurst()
		case 0x4c:
			spellDivineIntervention()
		}
		return false
	}
	text := func() {
		sp = int(p.ReadySpell)
		if p.ReadySpell == 0xff {
			sp = 0x4d
		}
		sprintf(gTextBuf, "\r\x05%04X\x02\x03c\x0b122\t013\x0c27C\x0cdast\t040\x0c27N\x0cdew\t067ESC\x01\t000\x0b000\x03cCast Spell\n\n%s\x03l\n\nSpell Ready:\x03c\n\n\x0c12%s\x0cd\x02\x03l\n\x0b082Cost\x03r\t077%u/%u\x03l\nCur SP\x03r\t077%u\x01",
			adrBankSprites, bytesOf(p), str(gSpellNames[sp]), spellSpCost(sp, getCharLevel(p)), gSpellGemCosts[sp], p.SP.Get())
	}

	if arg != 0 {
		p = &gParty[gCombatOrder[*gCombatActor]]
		cast()
		goto out
	}
	pushButtons()
	pushIcons("cast.icn")
	*gBankSpritesImg = *gCurIcons
	*gBtnXs = adrBankBtnXs
	*gBtnYs = adrBankBtnYs
	*gBtnWs = adrBankBtnWs
	*gBtnHs = adrBankBtnHs
	if mode == 2 {
		key = int(*gCombatActor)
	} else if *gCastChar < 0 || uint8(*gCastChar) >= *gPartyCount {
		for key = int(*gPartyCount) - 1; key >= 0; key-- {
			if gParty[key].Caster != 0 {
				*gCastChar = int8(key)
				break
			}
		}
	} else {
		key = int(*gCastChar)
	}
	if mode == 2 {
		p = &gParty[gCombatOrder[key]]
	} else {
		p = &gParty[key]
	}
	highlightPortrait(key)
	text()
	openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	for {
		key = getInput()
		switch key {
		case 0, 'C':
			if p.ReadySpell < 0x4d && charCannotAct(p) == 0 {
				if cast() {
					goto out
				}
				*gGameMode = 3
			}
		case 1, 'N':
			*gGameMode = uint8(mode)
			p = spellListDialog(p, 1)
			text() // show:
			printWindowText(gTextBuf)
			*gGameMode = 3
		case 2:
			key = 0x1b
			*gGameMode = uint8(mode)
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if mode == 2 {
				break
			}
			*gGameMode = uint8(mode)
			if key -= 0xc9; key >= int(*gPartyCount) {
				break
			}
			p = &gParty[key]
			highlightPortrait(key)
			*gCastChar = int8(key)
			text() // show:
			printWindowText(gTextBuf)
			*gGameMode = 3
		}
		if key == 0x1b || res != 0 {
			break
		}
	}
out:
	if arg == 0 {
		closeWindows(1)
		popIcons()
		*gCpanelButtonsImg = *gCurIcons
		popButtons()
		clearCharHighlight()
		if res != 0 && *gSpellMovedParty != 0 {
			drawWizardEye(1)
			*gSpellMovedParty = 0
		}
	}
	*gGameMode = uint8(mode)
	*gMonsterScanDelay = uint8(s17f)
	return res
}
