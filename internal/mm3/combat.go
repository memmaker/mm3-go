// Port of COMBAT.C (overlay 1839): the combat round, initiative, party and monster attacks,
// damage, saving throws and the combat main loop (run_combat).
//
// combat.c's struct chrat is the canonical Chr: res[k][0..1] at 0x107+2k (bytesOf), the
// conditions cursed..eradicated are Cond[0..15] (asleep 8, dead 13). Its view of the globals
// differs from ds_gen's: g_mon_hp and g_attack_type are unsigned, g_mon_hit_anim/hit_fx/
// elem_fx (u8) and g_mon_elem_sizes/hit_sizes (int) are arrays (local views below). The far
// monster tables are u8 except g_mon_exp (u32), g_mon_gold (long), g_mon_gems/max_hp (u16).
// monster_turn's labels (all, rnd, chk, hit) are a state loop; party_attack_monster's chk/two
// are closures over its tail.
package mm3

var (
	monHitAnim   = arr[uint8](adrMonHitAnim)
	monHitFx     = arr[uint8](adrMonHitFx)
	monElemFx    = arr[uint8](adrMonElemFx)
	monElemSizes = arr[int16](adrMonElemSizes)
	monHitSizes  = arr[int16](adrMonHitSizes)
)

func monType(m int) int { return int(gMonType[m]) }

// curMonType is g_mon_type[g_target_mon].
func curMonType() int { return monType(int(*gTargetMon)) }

func chrRes(c *Chr, k, j int) int { return int(bytesOf(c)[0x107+2*k+j]) }

func damageFxSize(n uint16) int {
	if n < 10 {
		return 0
	}
	if n < 50 {
		return 1
	}
	if n < 200 {
		return 2
	}
	return 3
}

func damageMonster(c *Chr, dmg uint16, kind int) {
	var t, killed, v int
	tm, ts := int(*gTargetMon), int(*gTargetSlot)

	if dmg != 0 {
		if gMonStatus[tm] == 7 {
			gMonStatus[tm] = 0
		}
		playSfx(1)
		monHitAnim[ts] = 3
		if kind != 2 {
			fx := 0
			if kind != 0 {
				at := uint16(*gAttackType)
				if at >= 7 {
					at = 6
				}
				fx = int(gElemHitFx[at])
			}
			monHitFx[ts] = uint8(fx + 1)
		} else {
			monHitFx[ts] = 2
		}
		if monType(tm) == 0x19 && kind == 0 {
			breakEquippedWeapons(c)
		}
		if kind == 0 || *gAttackType == 0 {
			if t = int(gMonResPhys.U8(monType(tm))); t != 0 {
				if t == 100 {
					dmg = 0
				} else {
					dmg = uint16(100-t) * dmg / 100
				}
			}
		}
		v = elemBonusDamage(kind)
		dmg += uint16(v)
		if v > 0 {
			monElemFx[ts] = uint8(getElementIndex(int(*gAtkElement)) + 1)
			monElemSizes[ts] = int16(damageFxSize(uint16(v)))
		} else if kind != 3 {
			monElemFx[ts] = 0
		}
		if kind != 0 && monsterSavingThrow(monType(tm)) != 0 {
			switch uint16(*gAttackType) {
			case 0xb, 0xd:
				dmg = 5
			case 0xf:
				dmg = 50
			default:
				dmg >>= 1
			}
		}
		monHitSizes[ts] = int16(damageFxSize(dmg))
		if uint16(gMonHp[tm]) <= dmg {
			killed = 1
			gMonHp[tm] = 0
		} else {
			gMonHp[tm] = int16(uint16(gMonHp[tm]) - dmg)
		}
	} else {
		playSfx(2)
	}
	refreshView()
	if killed != 0 {
		mt := monType(tm)
		playSfx(0)
		playSfx(3)
		awardExperience(gMonExp.U32(4 * mt))
		if *gMapNum != 0x6a {
			*gLootGold += int32(gMonGold.U32(4 * mt))
			*gLootGems += int32(gMonGems.U16(2 * mt))
			if e := gMonTreasure.U8(mt); e != 0 && randomRange(1, 100) <= int(gTreasureDropPct[e-1]) {
				if *gLootCount++; *gLootCount > 10 {
					*gLootCount = 10
				}
				p := &gParty[8]
				generateRandomItem(int(gMonTreasure.U8(mt)), p, 0)
				n := int(*gLootCount) - 1
				gLootFlags[n] = p.ItemFlags[0]
				gLootElement[n] = p.ItemElem[0]
				gLootMaterial[n] = p.ItemMat[0]
				gLootAttrib[n] = p.ItemAttr[0]
				gLootIds[n] = p.ItemID[0]
				gLootCharges[n] = p.ItemPower[0]
			}
		}
		gMonX[tm] = 0x80
		gMonY[tm] = 0x80
		monHitAnim[ts] = 0
		*gTargetMon = 0xff
		if ts < 12 {
			*[]*uint8{gMonD0c1, gMonD0c2, gMonD0c3, gMonD1c1, gMonD1c2, gMonD1c3,
				gMonD2c1, gMonD2c2, gMonD2c3, gMonD3c1, gMonD3c2, gMonD3c3}[ts] = 0
		}
		refreshView()
		if *gMonD0c1 != 0 {
			*gTargetMon = *gMonD0c1 - 1
			*gTargetSlot = 0
		}
	}
}

func awardExperience(xp uint32) {
	var n, cnt int
	if *gGameMode == 2 {
		n = int(*gCombatPartyN)
	} else {
		n = int(*gPartyCount)
	}
	for pass := 0; pass < 2; pass++ {
		for i := 0; i < n; i++ {
			k := i
			if *gGameMode == 2 {
				k = int(gCombatOrder[i])
			}
			p := &gParty[k]
			getWorstCondition(p) // an empty switch on it
			if pass == 0 {
				cnt++
			} else {
				p.Exp.Set(p.Exp.Get() + xp/uint32(int32(int16(cnt))))
			}
		}
	}
}

func elemBonusDamage(kind int) int {
	var r, s int
	mt := curMonType()
	if kind == 0 || kind == 3 {
		e := int(*gAtkElement)
		r = int(gElemDmgBonus[e])
		switch {
		case e == 0:
			s = 0
		case e < 9:
			s = int(gMonResFire.U8(mt))
		case e < 0x10:
			s = int(gMonResElec.U8(mt))
		case e < 0x15:
			s = int(gMonResCold.U8(mt))
		case e < 0x1a:
			s = int(gMonResAcid.U8(mt))
		case e < 0x22:
			s = int(gMonResEnergy.U8(mt))
		default:
			s = int(gMonResMagic.U8(mt))
		}
	} else {
		r = 0
		switch uint16(*gAttackType) {
		case 0:
			s = int(gMonResPhys.U8(mt))
		case 1:
			s = int(gMonResMagic.U8(mt))
		case 2:
			s = int(gMonResFire.U8(mt))
		case 3:
			s = int(gMonResElec.U8(mt))
		case 4:
			s = int(gMonResCold.U8(mt))
		case 5:
			s = int(gMonResAcid.U8(mt))
		case 6:
			s = int(gMonResEnergy.U8(mt))
		}
	}
	if s != 0 {
		if s == 100 {
			r = 0
		} else {
			r = int(int16((100-s)*r)) / 100
		}
	}
	return r
}

func inTargets(tbl []uint8, n, t int) bool {
	for j := 0; j < n; j++ {
		if t == int(tbl[j]) {
			return true
		}
	}
	return false
}

func partyAttackMonster(c *Chr, kind int) {
	var d int
	var dmg uint16

	if kind == 0 {
		switch c.Cls {
		case 7:
			d = 4
		case 0, 6:
			d = 5
		case 1, 2, 5, 9:
			d = 6
		case 3, 8:
			d = 7
		case 4:
			d = 8
		}
		n := uint16(getCharLevel(c))/uint16(d) + 1
		for i := uint16(0); i < n; i++ {
			if rollToHit(c, 0) != 0 {
				dmg += uint16(rollDamage(c))
			}
		}
	} else if *gSpellHitsParty == 1 {
		memset(monElemFx, 0, 12)
		for j := 0; j < int(*gPartyCount); j++ {
			if gMissilePos[j] != 0 && gPartyResisted[j] == 0 {
				if rollToHit(&gParty[j], kind) != 0 {
					if *gAttackDamage != 0 {
						dmg = uint16(*gAttackDamage)
					} else {
						dmg = uint16(*gAtkDamage)
					}
					gMissilePos[j] = 0
					damageMonster(c, dmg, 3)
					if *gTargetMon == 0xff {
						return
					}
				} else {
					gPartyResisted[j]++
				}
			}
		}
		return
	} else {
		mt := curMonType()
		if r := int(gMonResMagic.U8(mt)); r != 0 && randomRange(1, getCharLevel(at[Chr](int(*gCurChar)))+100) < r {
			return
		}
		if *gTargetMon == 0xff {
			return
		}
		chk := func() {
			if monsterSavingThrow(mt) != 0 {
				return
			}
			gMonStatus[*gTargetMon] = *gAttackType
		}
		if *gAttackDamage != 0 {
			dmg = uint16(*gAttackDamage)
		} else {
			switch uint16(*gAttackType) {
			case 16:
				if inTargets(gFx16Targets, 5, mt) {
					chk()
				}
				return
			case 7:
				if inTargets(gFx07Targets, 5, mt) {
					chk()
				}
				return
			case 8:
				if inTargets(gFx08Targets, 13, mt) {
					chk()
				}
				return
			case 9:
				if inTargets(gFx09Targets, 13, mt) {
					chk()
				}
				return
			case 10:
				if inTargets(gFx10Targets, 0x18, mt) {
					chk()
				}
				return
			case 11:
				if !inTargets(gFx11Targets, 0x16, mt) {
					return
				}
				dmg = uint16(gMonHp[*gTargetMon])
				kind = 2
			case 14:
				dmg = 0
				if inTargets(gFx12Targets, 9, mt) {
					dmg = 0x19
				}
				kind = 2
			case 12:
				dmg = 0
				if inTargets(gFx12Targets, 9, mt) {
					dmg = gMonMaxHp.U16(2 * mt)
				}
				kind = 2
			case 15:
				dmg = gMonMaxHp.U16(2 * mt)
				if dmg > 0x96 {
					dmg = 0x32
					*gAttackType = 6
				}
			case 13:
				dmg = uint16(gMonHp[*gTargetMon]) >> 1
				if dmg == 0 {
					dmg++
				}
				kind = 2
			}
		}
	}
	damageMonster(c, dmg, kind)
	combatSortInitiative()
}

// monsAhead is the number of monsters in the three front slots.
func monsAhead() int {
	return b2i(*gMonD0c1 > 0) + b2i(*gMonD0c2 > 0) + b2i(*gMonD0c3 > 0)
}

func combatRoundOver() int {
	n := monsAhead()
	for i := 0; i < int(*gCombatPartyN)+n; i++ {
		if gCombatActed[i] == 0 {
			if i < int(*gCombatPartyN) {
				switch getWorstCondition(&gParty[gCombatOrder[i]]) {
				case 11, 12, 13, 14, 15:
				default:
					return 0
				}
			} else {
				return 0
			}
		}
	}
	return 1
}

func compactCombatOrder() {
	for i := 0; i < 8; i++ {
		if gCombatOrder[i] == 0xff {
			for j := i + 1; j < 8; j++ {
				if gCombatOrder[j] < 0xff {
					gCombatOrder[i] = gCombatOrder[j]
					gCombatActed[i] = gCombatActed[j]
					gCombatActed[j] = 0
					gCombatOrder[j] = 0xff
					break
				}
			}
		}
	}
	drawPartyPortraits()
}

func charSavingThrow(c *Chr, kind int) int {
	var r, roll int
	if kind == 0 {
		r = getCharLevel(c) + statBonus(getCharStat(c, 6, 0)*2)
		roll = randomRange(1, r+20)
	} else {
		k, b := -1, 0
		switch kind {
		case 1:
			k, b = 5, 0x10
		case 2:
			k, b = 0, 0xb
		case 3:
			k, b = 2, 0xc
		case 4:
			k, b = 1, 0xd
		case 5:
			k, b = 3, 0xe
		case 6:
			k, b = 4, 0xf
		}
		if k >= 0 {
			r = int(int16(chrRes(c, k, 0) + getItemStatBonus(c, b) + chrRes(c, k, 1)))
		}
		roll = randomRange(1, r+40)
	}
	return b2i(roll <= r)
}

// incCond is `if (!++x) x = 0xff`.
func incCond(p *uint8) {
	if *p++; *p == 0 {
		*p = 0xff
	}
}

func monsterHitChar(c *Chr, slot, mt int) {
	var elem, snd, i, dmg int

	c.Cond[8] = 0
	for i = 0; i < int(gMonDmgDice.U8(mt)); i++ {
		dmg += randomRange(1, int(gMonDmgSides.U8(mt)))
	}
	snd = 4
	if i = int(gMonDmgType.U8(mt)); i != 0 {
		if charSavingThrow(c, 0) != 0 {
			dmg >>= 1
		}
		switch i {
		case 1:
			elem = 6
		case 2:
			dmg = dmg - int(*gFireResist)
			elem = 1
			snd = 0xd
		case 3:
			dmg = dmg - int(*gElecResist)
			elem = 2
			snd = 0xe
		case 4:
			dmg = dmg - int(*gColdResist)
			elem = 3
			snd = 0xf
		case 5:
			dmg = dmg - int(*gPoisonResist)
			elem = 4
			snd = 0x11
		case 6:
			elem = 5
		}
		for {
			if charSavingThrow(c, i) == 0 {
				break
			}
			dmg >>= 1
			if dmg == 0 {
				break
			}
		}
	}
	playSfx(snd)
	drawIcon(*gCharpowSprites, elem, slot*0x26+0xb, 0x95, 0)
	dmg -= int(c.Shield)
	if dmg < 1 {
		dmg = 0
	}
	zeroHP := func() {
		if c.HP.Get() > 0 {
			c.HP.Set(0)
		}
	}
	if sp := gMonSpecial.U8(mt); dmg != 0 && sp != 0 && charSavingThrow(c, 0) == 0 {
		switch sp {
		case 1:
			incCond(&c.Cond[3])
			playSfx(0x11)
		case 2:
			incCond(&c.Cond[8])
			playSfx(0x35)
		case 3:
			incCond(&c.Cond[5])
			playSfx(0x35)
		case 4:
			incCond(&c.Cond[4])
			playSfx(0x11)
		case 5:
			incCond(&c.Cond[10])
			playSfx(0x35)
		case 6:
			incCond(&c.Cond[0])
			playSfx(0x35)
		case 7:
			for i = 0; i < 0x12; i++ {
				if c.ItemID[i] >= 0x22 && c.ItemID[i] <= 0x29 && c.ItemEq[i] != 0 {
					c.ItemFlags[i] |= 0x80
					break
				}
			}
			playSfx(0x35)
		case 8:
			incCond(&c.Cond[2])
			playSfx(0x35)
		case 9:
			c.SP.Set(0)
			playSfx(0x35)
		case 10:
			incCond(&c.Cond[11])
			playSfx(0x35)
		case 11:
			c.AgeMod += 5
			playSfx(0x35)
		case 12:
			incCond(&c.Cond[12])
			playSfx(0x35)
			zeroHP()
		case 13:
			incCond(&c.Cond[15])
			breakAllItems(c)
			playSfx(0x35)
			zeroHP()
		case 14:
			incCond(&c.Cond[14])
			playSfx(0x35)
			zeroHP()
		case 15:
			incCond(&c.Cond[13])
			playSfx(0x35)
			zeroHP()
		case 16:
			for i = 0; i < 0x12; i++ {
				if c.ItemID[i] < 0x46 {
					c.ItemFlags[i] |= 0x40
				}
			}
			playSfx(0x35)
		}
	}
	charTakeDamage(c, dmg)
	idleFramesOrKey(2)
	drawPartyPortraits()
}

func monsterTurn(mt int) {
	var m, n, st, flag, i, target int
	var p *Chr
	var list [8]int
	const (
		sAll = iota
		sRnd
		sChk
		sHit
		sNext
	)

	if *gMonMissileActive == 0 {
		switch int(*gCombatActor) - int(*gCombatPartyN) {
		case 0:
			m = int(*gMonD0c1) - 1
		case 1:
			m = int(*gMonD0c2) - 1
		case 2:
			m = int(*gMonD0c3) - 1
		}
		if gMonStatus[m] != 0 && gMonStatus[m] != 1 {
			return
		}
		mt = monType(m)
	}
	pref := gMonAttackPref.U8(mt)
	np := int(*gCombatPartyN)
	for k := 0; k < int(gMonNumAttacks.U8(mt)); k++ {
		target = -1
		s := sRnd
		if pref == 0 {
			s = sAll
		} else if pref != 1 {
		scan:
			for i = 0; i < np; i++ {
				p = &gParty[gCombatOrder[i]]
				switch getWorstCondition(p) {
				case 11, 12, 13, 14, 15:
				default:
					var hit bool
					switch pref {
					case 2:
						hit = p.Cls == 3
					case 3:
						hit = p.Cls == 4
					case 4:
						hit = p.Cls == 8
					case 5:
						hit = p.Cls == 1
					case 6:
						hit = p.Race == 2
					}
					if hit {
						target = i
						s = sHit
						break scan
					}
				}
			}
		}
		for s != sNext {
			switch s {
			case sAll:
				flag = 1
				target++
				if np > target {
					s = sChk
				} else {
					s = sNext
				}
			case sRnd:
				switch np {
				case 1:
					target = 0
				case 2, 3, 4, 5:
					target = randomRange(0, np-1)
				case 6:
					if i = randomRange(1, 8); i == 8 {
						target = 5
					} else {
						target = randomRange(0, 4)
					}
				case 7:
					if i = randomRange(1, 10); i == 10 {
						target = 6
					} else {
						target = randomRange(0, 5)
					}
				case 8:
					if i = randomRange(1, 12); i == 12 {
						target = 7
					} else {
						target = randomRange(0, 6)
					}
				}
				s = sChk
			case sChk:
				s = sHit
				if flag != 0 {
					break
				}
				st = int(getWorstCondition(&gParty[gCombatOrder[target]]))
				switch st {
				case 11, 12, 13, 14, 15:
					for j := range list {
						list[j] = -1
					}
					n = 0
					for i = 0; i < np; i++ {
						st = int(getWorstCondition(&gParty[gCombatOrder[i]]))
						switch st {
						case 11, 12, 13, 14, 15: // case 11, 12 jump to hit when flag is set: it is 0 here
						default:
							list[n] = int(gCombatOrder[i])
							n++
						}
					}
					if n == 0 {
						*gPartyDefeated = 1
						*gGameMode = 1
						return
					}
					target = list[randomRange(0, n-1)]
					for i = 0; i < np; i++ {
						if int(gCombatOrder[i]) == target {
							target = i
							break
						}
					}
				}
			case sHit:
				p = &gParty[gCombatOrder[target]]
				if gMonDmgType.U8(mt) != 0 || p.Cond[8] != 0 {
					monsterHitChar(p, target, mt)
				} else {
					miss := false
					if i = randomRange(1, 20); i == 20 {
						monsterHitChar(p, target, mt)
					} else if i == 1 {
						miss = true
					}
					if !miss {
						hb := int(gMonHitBonus.U8(mt))
						i += hb>>2 + randomRange(1, hb)
						blk := 10
						if gCombatBlocking[target] != 0 {
							blk = getCharLevel(p)>>1 + 15
						}
						if calcArmorClass(p, 0)+blk <= i {
							monsterHitChar(p, target, mt)
						} else {
							miss = true
						}
					}
					if miss {
						playSfx(2)
					}
				}
				if flag != 0 {
					s = sAll
				} else {
					s = sNext
				}
			}
		}
	}
	drawPartyPortraits()
}

func partyCannotAct() int {
	for i := 0; i < int(*gCombatPartyN); i++ {
		switch getWorstCondition(&gParty[gCombatOrder[i]]) {
		case 8, 11, 12, 13, 14, 15:
		default:
			return 0
		}
	}
	return 1
}

func combatNextActor() {
	n := monsAhead()
	if combatRoundOver() != 0 {
		return
	}
	if *gMonD0c1 == 0 && *gMonD0c2 == 0 && *gMonD0c3 == 0 || *gCombatPartyN == 0 {
		*gGameMode = 1
		return
	}
top:
	for {
		updatePartyDisabled()
		if *gPartyDefeated != 0 {
			*gGameMode = 1
			return
		}
	actors:
		for i := 0; i < int(*gCombatPartyN)+n; i++ {
			if *gCombatActor != -1 {
				gCombatActed[*gCombatActor] = 1
			}
			*gCombatTurnPos = int8(int(int8(*gCombatTurnPos+1)) % (int(*gCombatPartyN) + n))
			*gCombatActor = int8(gInitiativeOrder[*gCombatTurnPos])
			if combatRoundOver() != 0 {
				if partyCannotAct() == 0 {
					return
				}
				combatSortInitiative()
				*gCombatActor = -1
				*gCombatTurnPos = -1
				memset(gCombatActed, 0, 12)
				continue top
			}
			if uint8(*gCombatActor) < *gCombatPartyN {
				switch getWorstCondition(&gParty[gCombatOrder[*gCombatActor]]) {
				case 8, 11, 12, 13, 14, 15:
					continue actors
				}
			}
			break
		}
		if uint8(*gCombatActor) >= *gCombatPartyN {
			monsterTurn(0)
			if *gPartyDefeated == 0 {
				updatePartyDisabled()
			}
			if *gPartyDefeated != 0 {
				return
			}
			continue top
		} else if combatRoundOver() == 0 {
			highlightPortrait(int(*gCombatActor))
		}
		return
	}
}

func doQuickfightAction() {
	p := &gParty[gCombatOrder[*gCombatActor]]
	switch p.Quickfight {
	case 0:
		partyAttackMonster(p, 0)
	case 1:
		if castSpellDialog(1) == 0 {
			partyAttackMonster(p, 0)
		}
	case 2:
		combatBlock()
	case 3:
		combatTryRun()
	}
}

func rollToHit(c *Chr, kind int) int {
	var v, d int
	loadAttackWeapon(c, kind)
	v = statBonus(getCharStat(c, 5, 0))
	v += int(*gAtkHitBonus)
	switch c.Cls {
	case 0, 7:
		d = 1
	case 1, 2, 5, 6, 9:
		d = 2
	case 3, 8:
		d = 3
	case 4:
		d = 4
	}
	v = v + int(uint16(getCharLevel(c))/uint16(d))
	for {
		d = randomRange(1, 20)
		v += d
		if d != 20 {
			break
		}
	}
	v -= int(c.Cond[0])
	if gMonStatus[*gTargetMon] != 0 {
		v += 20
	}
	ac := 10
	ac += int(gMonAc.U8(curMonType()))
	return b2i(v >= ac)
}

func loadAttackWeapon(c *Chr, kind int) {
	*gAtkElement, *gAtkWeaponFef, *gAtkHitBonus, *gAtkDiceN, *gAtkDieSides = 0, 0, 0, 0, 0
	*gAtkDamage = 0
	for i := 0; i < 0x12; i++ {
		var ok bool
		if kind != 0 {
			ok = c.ItemEq[i] == 4
		} else {
			ok = c.ItemEq[i] == 1 || c.ItemEq[i] == 13
		}
		if ok {
			*gAtkElement = int8(c.ItemElem[i])
			*gAtkWeaponFef = int8(c.ItemPower[i])
			*gAtkHitBonus = int8(int(int8(gMaterialToHit[c.ItemMat[i]])) + int(c.Hero))
			*gAtkDamage = int16(int(int8(gMaterialDmgBonus[c.ItemMat[i]])) + int(c.Holy))
			*gAtkDiceN = int8(gWeaponDiceCount[c.ItemID[i]])
			*gAtkDieSides = int8(gWeaponDieSides[c.ItemID[i]])
			for k := 0; k < int(*gAtkDiceN); k++ {
				*gAtkDamage += int16(randomRange(1, int(*gAtkDieSides)))
			}
		}
	}
	if *gAtkDamage < 1 {
		*gAtkDamage = 0
	}
}

func rollDamage(c *Chr) int {
	v := statBonus(getCharStat(c, 0, 0))
	v += int(*gAtkDamage)
	if v < 1 {
		return 0
	}
	return v
}

func monsterSavingThrow(lvl int) int { return b2i(randomRange(1, lvl+50) <= lvl) }

func combatTryRun() {
	if randomRange(1, 100) < int(gMapsRunChance[*gMapQuadrant][0]) {
		gCombatOrder[*gCombatActor] = 0xff
		*gCombatPartyN--
		compactCombatOrder()
		combatSortInitiative()
		*gCombatTurnPos--
		*gCombatActor = -1
		*gPartyFled = 1
	}
}

func combatBlock() { gCombatBlocking[*gCombatActor] = 1 }

func monsterHealthColor(m int) int {
	hp := gMonMaxHp.U16(2 * monType(m))
	if uint16(gMonHp[m]) == hp {
		return 0x12
	}
	if uint16(gMonHp[m]) < hp>>2 {
		return 7
	}
	return 0xc
}

func formatMonsterList() {
	var s1, s2, s3 [40]byte
	one := func(s []byte, d *uint8, f string) {
		if *d != 0 {
			c := int8(monsterHealthColor(int(*d) - 1))
			sprintf(s, f, int(c), str(gMonsterNames[monType(int(*d)-1)]))
		}
	}
	one(s1[:], gMonD0c1, "\n\v020\f%2u%s\fd")
	one(s2[:], gMonD0c2, "\n\v030\f%2u%s\fd")
	one(s3[:], gMonD0c3, "\n\v040\f%2u%s\fd")
	if *gTargetSlot == 2 && *gMonD0c3 != 0 {
		*gTargetMon = *gMonD0c3 - 1
	} else if *gTargetSlot == 1 && *gMonD0c2 != 0 {
		*gTargetMon = *gMonD0c2 - 1
	} else {
		*gTargetMon = *gMonD0c1 - 1
		*gTargetSlot = 0
	}
	*gCombatCursorY = int16(int(*gTargetSlot)*10 + 0x12)
	sprintf(gTextBuf, "\r\v000\t000\x03c\x02Combat%s%s%s\x05%04X\x01", s1[:], s2[:], s3[:], adrCombatCursorSprites)
}

func combatSortInitiative() {
	var i, v, n, max int
	var sp [12]int

	had := *gCombatTurnPos != -1
	cur := int8(ds[adrInitiativeOrder+int(*gCombatTurnPos)])
	memset(gInitiativeOrder, 0xff, 12)
	for j := range sp {
		sp[j] = -1
	}
	for i = 0; i < int(*gCombatPartyN); i++ {
		v = int(int16(getCharStat(&gParty[gCombatOrder[i]], 4, 0)))
		sp[i] = v
		if v > max {
			max = v
		}
	}
	for _, d := range []*uint8{gMonD0c1, gMonD0c2, gMonD0c3} {
		if *d != 0 {
			v = int(gMonSpeed.U8(monType(int(*d) - 1)))
			sp[i] = v
			i++
			if v > max {
				max = v
			}
		}
	}
	for i = max; i >= 0; i-- {
		for v = 0; v < int(*gCombatPartyN)+3; v++ {
			if sp[v] == i {
				gInitiativeOrder[n] = byte(v)
				n++
			}
		}
	}
	if had && int8(ds[adrInitiativeOrder+int(*gCombatTurnPos)]) != cur {
		n = monsAhead()
		for i = 0; i < int(*gCombatPartyN)+n; i++ {
			if int8(gInitiativeOrder[i]) == cur {
				*gCombatTurnPos = int8(i)
				break
			}
		}
	}
}

func runCombat() {
	var i, r int

	*gGameMode = 2
	loadCombatIcons()
	drawMainIcons()
	memset(gCombatOrder, 0xff, 9)
	memset(gCombatActed, 0, 12)
	memset(gCombatBlocking, 0, 12)
	monHitAnim[0], monHitAnim[1], monHitAnim[2] = 0, 0, 0
	*gTargetSlot = 0
	*gPartyFled = 0
	*gTargetMon = *gMonD0c1 - 1
	for i = 0; i < int(*gPartyCount); i++ {
		gCombatOrder[i] = uint8(i)
	}
	*gCombatPartyN = *gPartyCount
	combatSortInitiative()
	*gCombatTurnPos = -1
	*gCombatActor = -1
	*gHighlightedChar = -1
	combatNextActor()
	if *gPartyDefeated != 0 {
		return
	}
	if *gMonsterScanDelay != 0 {
		moveMonsters()
		refreshView()
	}
	highlightPortrait(int(*gCombatActor))
	formatMonsterList()
	openTextWindow(0xe7, 1, 0x59, 0x48, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	turn := func() {
		if *gMazeGfxIndex == 0 || *gOutdoors != 0 || *gMazeGfxIndex == 3 {
			*gViewBgFlip ^= 1
		}
		if *gMonsterScanDelay != 0 {
			moveMonsters()
		}
		*gViewDirty = 1
		*gViewFlip ^= 1
	}
	for {
		if *gMonD0c1 != 0 {
			playSfx(curMonType() + 0x3c)
		}
		combatSortInitiative()
		formatMonsterList()
		printWindowText(gTextBuf)
		cur := func() *Chr { return &gParty[gCombatOrder[*gCombatActor]] }
		switch i = getInput(); i {
		case 0xf0:
			if d := *gPartyDir; d < 4 {
				*gPartyDir = [4]uint8{3, 2, 0, 1}[d]
			}
			turn()
		case 0xf1:
			if d := *gPartyDir; d < 4 {
				*gPartyDir = [4]uint8{2, 3, 1, 0}[d]
			}
			turn()
		case 0, 0x46:
			doQuickfightAction()
			combatNextActor()
		case 1, 0x43:
			if castSpellDialog(0) != 0 {
				combatNextActor()
				break
			}
			highlightPortrait(int(gCombatOrder[*gCombatActor]))
		case 2, 0x41:
			partyAttackMonster(cur(), 0)
			combatNextActor()
		case 3, 0x55:
			i = int(*gCombatActor)
			itemScreen(cur(), 7)
			if int(*gCombatActor) != i {
				*gCombatActor = int8(i)
				combatNextActor()
				break
			}
			highlightPortrait(int(*gCombatActor))
			loadCombatIcons()
		case 4, 0x52:
			combatTryRun()
			combatNextActor()
			if *gGameMode != 1 {
				break
			}
			*gLootGems = 0
			*gLootCount = 0
			*gLootGold = 0
			setPartyMapStart()
			goto done
		case 5, 0x42:
			combatBlock()
			combatNextActor()
		case 6, 0x4f:
			quickfightOptions(cur())
			highlightPortrait(int(*gCombatActor))
		case 7, 0x49:
			showGameInfo()
			highlightPortrait(int(*gCombatActor))
		case 8, 0x51:
			showQuickRef()
			highlightPortrait(int(*gCombatActor))
		case 9:
			controlPanel()
			highlightPortrait(int(*gCombatActor))
		case 10, 13, 16, 0x31:
			if *gMonD0c1 != 0 {
				*gTargetMon = *gMonD0c1 - 1
				*gTargetSlot = 0
			}
		case 11, 14, 17, 0x32:
			if *gMonD0c2 != 0 {
				*gTargetMon = *gMonD0c2 - 1
				*gTargetSlot = 1
			}
		case 12, 15, 0x33:
			if *gMonD0c3 != 0 {
				*gTargetMon = *gMonD0c3 - 1
				*gTargetSlot = 2
			}
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if i -= 0xc9; i < int(*gCombatPartyN) {
				charScreen(i)
				highlightPortrait(int(*gCombatActor))
				loadCombatIcons()
			}
		}
		if combatRoundOver() != 0 {
			r++
			memset(gCombatActed, 0, 12)
			memset(gCombatBlocking, 0, 12)
			combatSortInitiative()
			*gCombatActor = -1
			*gCombatTurnPos = -1
			combatNextActor()
			for i = 0; i < int(*gMonCount); i++ {
				if gMonType[i] == 0x37 {
					gMonHp[i] = int16(gMonMaxHp.U16(2 * 0x37))
				}
			}
			recoverMonsterStatus()
			moveMonsters()
			scanViewMonsters()
			passTime(1)
		}
		if *gMonD0c1 == 0 && *gMonD0c2 == 0 && *gMonD0c3 == 0 {
			passTime(1)
			refreshView()
			if *gMonD0c1 == 0 && *gMonD0c2 == 0 && *gMonD0c3 == 0 {
				break
			}
		}
		updatePartyDisabled()
		if !(*gPartyDefeated == 0 && *gGameMode == 2) {
			break
		}
	}
done:
	*gGameMode = 1
	if *gPartyFled != 0 && (*gMonD0c1 != 0 || *gMonD0c2 != 0 || *gMonD0c3 != 0) {
		updatePartyDisabled()
		if *gPartyDefeated == 0 {
			setPartyMapStart()
			for i = 0; i < int(*gPartyCount); i++ {
				switch getWorstCondition(&gParty[i]) {
				case 8, 11, 12, 14, 15:
					gParty[i].Cond[13] = 1
				}
			}
		}
	}
	closeWindows(1)
	flushInput()
	*gGameMode = 2
	refreshView()
	awardFoundLoot()
	*gGameMode = 1
	*gViewDirty = 1
	drawWizardEye(1)
	clearCharHighlight()
	for i = 0; i < int(*gPartyCount); i++ {
		gCombatOrder[i] = uint8(i)
	}
	*gCombatPartyN = *gPartyCount
	drawPartyPortraits()
	setupMainUi()
	drawMainIcons()
	*gTargetMon = 0xff
	playSfx(0)
}
