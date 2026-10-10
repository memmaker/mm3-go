// Port of SPELLS.C (overlay 1843): the spell effects, their element/target pickers, spell
// costs and refunds.
//
// spells.c's struct chr is the canonical Chr (endur 0x1a, age_mod 0x26, beacon 0x77, hp signed).
// Its g_spell_gem_costs is int (ds_gen: unsigned); g_food is int there.
package mm3

var onWhoText = cs("\x03c\x0b009On Who?")

// spellBuf is the buf[100] of spell_pick_element and the two messages: the window gets its far address.
var spellBuf [100]byte

// curChr is g_cur_char (a near pointer to a party record).
func curChr() *Chr { return at[Chr](int(*gCurChar)) }

func spellFailOutdoorsOnly() { showMessage(cs("Only works Outdoors.  Spell Failed.")) }

func spellFailed() { showMessage(str(*gMsgSpellFailed)) }

func clearPartyAttackFlags() {
	for i := 0; i < 8; i++ {
		gMissilePos[i] = 0
	}
}

func charCanShoot(p *Chr) int {
	for i := 0; i < 0x12; i++ {
		if p.ItemID[i] >= 0x1e && p.ItemID[i] <= 0x21 && p.ItemEq[i] != 0 {
			switch getWorstCondition(p) {
			case 8, 11, 12, 13, 14, 15:
				return 0
			default:
				return 1
			}
		}
	}
	return 0
}

func spellHealChar(p *Chr, n int) {
	switch getWorstCondition(p) {
	case 13, 14, 15:
		return
	}
	img := loadGraphic("spfx9.icn")
	k := int(int16(calcMaxHp(p)))
	if p.HP.Get() <= k {
		p.HP.Set(p.HP.Get() + n)
		if p.HP.Get() > k {
			p.HP.Set(k)
		}
		for k = 0; k < int(*gPartyCount); k++ {
			if &gParty[k] == p {
				break
			}
		}
		k = k*0x26 + 0xb
		save := *gMonsterScanDelay
		*gMonsterScanDelay = 0
		for i := 0; i < 4; i++ {
			drawIcon(img, i, k, 0x95, 0)
			refreshView()
		}
		*gMonsterScanDelay = save
	}
	if p.HP.Get() > 0 {
		p.Cond[12] = 0
	}
	freeAndNull(&img)
	drawPartyPortraits()
}

func spellPickElement(a int) int {
	var k int
	buf := &spellBuf

	save := *gGameMode
	*gGameMode = 0xff
	pushButtons()
	pushIcons("element.icn")
	*gBtnXs = adrElementBtnXs
	*gBtnYs = adrElementBtnYs
	*gBtnWs = adrElementBtnWs
	*gBtnHs = adrElementBtnHs
	*gElementDlgIcn = *gCurIcons
	sprintf(buf[:], "\x05%04X\x03cWhich Element?\x02\x0b034\t014\x0c27F\x0cdire\t044\x0c27E\x0cdlec\t074\x0c27C\x0cdold\t104\x0c27A\x0cdcid\x01", adrElementDlgSprites)
	openTextWindow(0x32, 0x47, 0x84, 0x3a, 0, 0x96, 4, 6, farOf(buf[:]))
loop:
	for {
		k = getInput()
		switch k {
		case 0x1b:
			k = -1
			spellRefundCost(curChr(), a)
		case 0, 'F':
			k = 2
		case 1, 'E':
			k = 3
		case 2, 'C':
			k = 4
		case 3, 'A':
			k = 5
		default:
			continue
		}
		break loop
	}
	closeWindows(1)
	popIcons()
	popButtons()
	*gGameMode = save
	return k
}

func spellPickPartyMember(a int) int {
	var k int
	pushButtons()
	setAnykeyButtons()
	s49c := *gPortraitClickOn
	*gPortraitClickOn = 1
	s520 := *gGameMode
	*gGameMode = 3
	openTextWindow(0xe4, 0x6a, 0x5b, 0x28, 0, 0x96, 4, 6, farOf(onWhoText))
loop:
	for {
		k = getInput()
		switch k {
		case 0x1b:
			k = -1
			spellRefundCost(curChr(), a)
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			k -= 0xc9
			if k >= int(*gPartyCount) {
				continue
			}
		default:
			continue
		}
		break loop
	}
	closeWindows(1)
	popButtons()
	*gGameMode = s520
	*gPortraitClickOn = s49c
	return k
}

func spellMessageBox(buf []byte) {
	openTextWindow(0x2a, 0x4c, 0xe0, 0x30, 0, 0x96, 4, 6, farOf(buf))
	waitMouseRelease()
	getInput()
	closeWindows(1)
	waitMouseRelease()
}

func showCantCastEngaged(a int) {
	buf := &spellBuf
	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	playSfx(0x15)
	sprintf(buf[:], "\x03c\x0b007Can't cast %s while engaged!", str(gSpellNames[a]))
	spellMessageBox(buf[:])
	popButtons()
	*gGameMode = save
}

func showNotEnoughToCast(a, b int) {
	buf := &spellBuf
	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	playSfx(0x15)
	sprintf(buf[:], "\x03c\x0b010Not enough %s to Cast %s", str(gCostNames[b]), str(gSpellNames[a]))
	spellMessageBox(buf[:])
	popButtons()
	*gGameMode = save
}

// spellCost is the gems and spell points of spell s for p.
func spellCost(p *Chr, s int) (gems, cost int) {
	gems = int(int16(gSpellGemCosts[s]))
	cost = int(gSpellSpCosts[s])
	if cost < 1 {
		cost = int(int16(uint16(getCharLevel(p)) * uint16(cost*-1)))
	}
	return
}

func spellPayCost(p *Chr, s int) int {
	gems, cost := spellCost(p, s)
	if p.SP.Get() < cost {
		return 1
	}
	if uint32(int32(gems)) > *gPartyGems {
		return 2
	}
	p.SP.Set(p.SP.Get() - cost)
	*gPartyGems -= uint32(int32(gems))
	return 0
}

func spellRefundCost(p *Chr, s int) {
	gems, cost := spellCost(p, s)
	p.SP.Set(p.SP.Get() + cost)
	*gPartyGems += uint32(int32(gems))
}

func spellLight() {
	*gLightCount++
	if *gDimLevel != 0 {
		*gLightChanged = 1
		*gViewDirty = 1
	}
	*gSpellMovedParty = 1
	playSfx(0x2c)
}

func spellAwaken() {
	for i := 0; i < int(*gPartyCount); i++ {
		gParty[i].Cond[8] = 0
	}
	drawPartyPortraits()
	playSfx(0x34)
}

// spellHealPicked is the shape of first aid, cure wounds, power cure and nature's cure:
// pick, fail on dead/stone/eradicated, else heal by n().
func spellHealPicked(a int, n func() int) {
	k := spellPickPartyMember(a)
	if k == -1 {
		return
	}
	p := &gParty[k]
	if c := int(getWorstCondition(p)); c >= 13 && c <= 15 {
		spellFailed()
	} else {
		playSfx(0x34)
		spellHealChar(p, n())
	}
}

func spellFirstAid() { spellHealPicked(2, func() int { return 6 }) }

// spellAttack sets the attack globals, plays snd and attacks with pow animation pow.
func spellAttack(dmg, typ, scope, snd, pow int) {
	*gAttackDamage = int16(dmg)
	*gAttackType = int16(typ)
	*gAttackScope = int16(scope)
	playSfx(snd)
	spellAttackMonsters(pow)
}

func spellFlyingFist() { spellAttack(6, 0, 0, 0x12, 0xe) }

// spellItemScreen is detect magic, recharge, duplication and enchant: item_screen in a mode.
func spellItemScreen(a, mode, snd int) {
	if k := spellPickPartyMember(a); k != -1 {
		save := *gGameMode
		*gGameMode = 0xff
		if snd != 0 {
			playSfx(snd)
		}
		itemScreen(&gParty[k], mode)
		*gGameMode = save
	}
}

func spellDetectMagic() { spellItemScreen(4, 3, 0x14) }

func spellElementalArrow() {
	*gAttackDamage = 8
	if *gAttackType = int16(spellPickElement(5)); *gAttackType != -1 {
		*gAttackScope = 0
		spellAttackMonsters(0xb)
	}
}

// spellCureCond picks a member, heals by 0 and clears condition cond.
func spellCureCond(a, snd, cond int) {
	if k := spellPickPartyMember(a); k != -1 {
		playSfx(snd)
		spellHealChar(&gParty[k], 0)
		gParty[k].Cond[cond] = 0
		drawPartyPortraits()
	}
}

func spellRevitalize()  { spellCureCond(6, 0x34, 2) }
func spellCureWounds()  { spellHealPicked(7, func() int { return 0xf }) }
func spellSparks()      { spellAttack(getCharLevel(curChr())*2, 3, 1, 0xe, 5) }
func spellEnergyBlast() { spellAttack(randomRange(2, 6)*getCharLevel(curChr()), 6, 0, 0x10, 0xd) }
func spellSleep()       { spellAttack(0, 7, 1, 0x12, 0xe) }
func spellPain()        { spellAttack(8, 0, 1, 0x12, 0xe) }

func spellCreateRope() {
	*gRopeCreated = 1
	*gViewDirty = 1
	playSfx(0x14)
}

func spellToxicCloud() { spellAttack(0xa, 5, 1, 0x11, 0xa) }

func spellSuppress(a, cond int) {
	k := spellPickPartyMember(a)
	if k == -1 {
		return
	}
	p := &gParty[k]
	if p.Cond[cond] != 0 {
		if p.Cond[cond] < 4 {
			p.Cond[cond] = 1
		} else {
			p.Cond[cond] -= 3
		}
	}
	playSfx(0x14)
	spellHealChar(p, 0)
	drawPartyPortraits()
}

func spellSuppressPoison() { spellSuppress(0xe, 3) }

func spellProtFromElements() {
	v := getCharLevel(curChr())*2 + 5
	if v > 200 {
		v = 200
	}
	switch spellPickElement(0xf) {
	case -1:
		return
	case 2:
		*gFireResist = int16(v)
	case 3:
		*gElecResist = int16(v)
	case 4:
		*gColdResist = int16(v)
	case 5:
		*gPoisonResist = int16(v)
	}
	playSfx(0x14)
	drawPartyPortraits()
}

func spellTurnUndead() { spellAttack(0, 0xe, 1, 0x12, 0xe) }

func dxFrom7(d, i int) int { return int(int8(gViewDxFrom7[d][i])) }
func dyFrom7(d, i int) int { return int(int8(gViewDyFrom7[d][i])) }

func spellJump() {
	d := int(*gPartyDir)
	x := int(*gPartyX) + dxFrom7(d, 0)
	y := int(*gPartyY) + dyFrom7(d, 0)
	if getCellWallsRegion(int(*gPartyX), int(*gPartyY), int(gDirBlockBits[d])) == 0 && getCellWallsRegion(x, y, int(gDirBlockBits[d])) == 0 {
		*gPartyX += uint8(dxFrom7(d, 7))
		*gPartyY += uint8(dyFrom7(d, 7))
		playSfx(0x33)
		refreshView()
		*gSpellMovedParty = 1
		*gViewDirty = 1
	} else {
		spellFailed()
	}
}

func spellAcidStream()      { spellAttack(0x19, 5, 0, 0x11, 0xa) }
func spellSuppressDisease() { spellSuppress(0x13, 4) }
func spellSilence()         { spellAttack(0, 0x10, 1, 0x14, 0xe) }

// spellBoost picks a member, heals by 0 and sets one of bless/holy/hero/shield to the caster's level.
func spellBoost(a, snd int, f func(*Chr) *uint8) {
	if k := spellPickPartyMember(a); k != -1 {
		playSfx(snd)
		spellHealChar(&gParty[k], 0)
		*f(&gParty[k]) = uint8(getCharLevel(curChr()))
		drawPartyPortraits()
	}
}

func spellBlessed() { spellBoost(0x15, 0x34, func(c *Chr) *uint8 { return &c.Bless }) }

func spellLevitate() {
	*gLevitate = 1
	playSfx(0x14)
}

func spellWizardEye() {
	*gWizardEye = 1
	*gSpellMovedParty = 1
	playSfx(0x14)
}

func spellIdentifyMonster() {
	var s [3][50]byte
	var k int
	if *gMonD0c1 == 0 && *gMonD0c2 == 0 && *gMonD0c3 == 0 {
		spellFailed()
		return
	}
	pushButtons()
	for i, d := range []*uint8{gMonD0c1, gMonD0c2, gMonD0c3} {
		if *d == 0 {
			continue
		}
		k = int(*d) - 1
		m := monType(k)
		sprintf(s[i][:], "\x03l\n%s\x03c\t100%u\t140%u\t180%u\x03r\t250%s",
			str(gMonsterNames[m]), gMonHp[k], gMonAc.U8(m), gMonNumAttacks.U8(m), str(gMonsterSpecialNames[gMonSpecial.U8(m)]))
	}
	playSfx(0x14)
	sprintf(gTextBuf, "Name\x03c\t100HP\t140AC\t177#Atks\x03r\t250Special%s%s%s", s[0][:], s[1][:], s[2][:])
	openTextWindow(0x14, 0x8e, 0x10e, 0x39, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	getInput()
	closeWindows(1)
	popButtons()
}

func spellHolyBonus() { spellBoost(0x19, 0x34, func(c *Chr) *uint8 { return &c.Holy }) }
func spellPowerCure() {
	spellHealPicked(0x1a, func() int { return randomRange(2, 0xc) * getCharLevel(curChr()) })
}
func spellNaturesCure()   { spellHealPicked(0x1b, func() int { return 0x19 }) }
func spellLightningBolt() { spellAttack(randomRange(4, 6)*getCharLevel(curChr()), 3, 1, 0xe, 3) }
func spellImmobilize()    { spellAttack(0, 8, 1, 0x12, 0xe) }
func spellHeroism()       { spellBoost(0x1e, 0x34, func(c *Chr) *uint8 { return &c.Hero }) }

func spellWalkOnWater() {
	*gWalkOnWater = 1
	playSfx(0x14)
}

func spellFrostBite() { spellAttack(0x23, 4, 0, 0xf, 8) }

func spellLloydsBeacon() {
	if gMapSpellFlags[*gMapQuadrant][1] == 0 {
		spellFailed()
		return
	}
	c := curChr()
	pushButtons()
	pushIcons("lloyds.icn")
	*gTrainSpritesImg = *gCurIcons
	*gBtnXs = adrTrainBtnXs
	*gBtnYs = adrTrainBtnYs
	*gBtnWs = adrTrainBtnWs
	*gBtnHs = adrTrainBtnHs
	if c.BeaconMap == 0 {
		c.BeaconMap = 1
		c.BeaconX = 2
		c.BeaconY = 5
	}
	sprintf(gTextBuf, "\x05%04X\x03cLloyd's Beacon\n\nLast Location\n\n%s\x03l\nx = %d\x03r\t077y = %d\x03c\x02\x0b122\t021\x0c27S\x0cdet\t060\x0c27R\x0cdeturn\x01",
		adrTrainSprites, str(gMapNames[c.BeaconMap]), c.BeaconX, c.BeaconY)
	openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
loop:
	for {
		switch getInput() {
		case 0, 'R':
			playSfx(0x33)
			popIcons()
			popIcons()
			if c.BeaconMap != *gMapNum {
				enterMaze(int(c.BeaconMap))
			}
			pushIcons("cast.icn")
			pushIcons("lloyds.icn")
			*gPartyX = c.BeaconX
			*gPartyY = c.BeaconY
			updateOutdoorQuadrant()
		case 1, 'S':
			playSfx(0x14)
			c.BeaconMap = *gCurTown
			c.BeaconX = *gPartyX
			c.BeaconY = *gPartyY
		case 0x1b:
		default:
			continue
		}
		break loop
	}
	closeWindows(1)
	popIcons()
	popButtons()
	*gSpellMovedParty = 1
}

func spellPowerShield() { spellBoost(0x22, 0x14, func(c *Chr) *uint8 { return &c.Shield }) }
func spellCurePoison()  { spellCureCond(0x23, 0x34, 3) }
func spellFireball()    { spellAttack(randomRange(3, 7)*getCharLevel(curChr()), 2, 1, 0xd, 0) }

func spellDetectMonster() {
	var cnt [7][7]uint8
	pushButtons()
	setAnykeyButtons()
	img := loadGraphic("detctmon.icn")
	memset(ds[adrDrawList:], 0, 0xfa0)
	gDrawList[0] = -1
	*gDrawListImg = img
	*gDrawListX = 4
	*gDrawListY = 0x18
	n := 7
	dy := 4
	for r := 0; r < 7; r++ {
		dy--
		dx := -4
		for c := 0; c < 7; c++ {
			dx++
			for j := 0; j < int(*gMonCount); j++ {
				if int(gMonX[j]) == int(*gPartyX)+dx && int(gMonY[j]) == int(*gPartyY)+dy {
					gDrawList[n] = int16(c*9 + 5)
					n++
					gDrawList[n] = int16(r*7 + 0x19)
					n++
					n++
					if cnt[r][c]++; cnt[r][c] > 3 {
						cnt[r][c] = 3
					}
					gDrawList[n] = int16(cnt[r][c])
					n++
				}
			}
		}
	}
	gDrawList[n] = 0x1f
	n++
	gDrawList[n] = 0x2d
	n++
	n++
	gDrawList[n] = int16(int(*gPartyDir) + 4)
	n++
	gDrawList[n] = -1
	playSfx(0x14)
	sprintf(gTextBuf, "\x05%04X\x03cDetect Monsters", adrDrawList)
	openTextWindow(0xe7, 0x30, 0x56, 0x5d, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	getInput()
	closeWindows(1)
	freeAndNull(&img)
	popButtons()
}

func spellAcidSpray()   { spellAttack(0xf, 5, 2, 0x11, 0xa) }
func spellColdRay()     { spellAttack(randomRange(2, 4)*getCharLevel(curChr()), 4, 2, 0xf, 8) }
func spellCureDisease() { spellCureCond(0x28, 0x34, 4) }

func spellNaturesGate() {
	if gMapSpellFlags[*gMapQuadrant][5] == 0 {
		spellFailed()
		return
	}
	playSfx(0x33)
	popIcons()
	enterMaze(int(gNatgateMaps[*gDay%10]))
	pushIcons("cast.icn")
	*gPartyX = gNatgateXs[*gDay%10]
	*gPartyY = gNatgateYs[*gDay%10]
	updateOutdoorQuadrant()
	*gSpellMovedParty = 1
}

func spellTimeDistortion() {
	if gMapSpellFlags[*gMapQuadrant][2] == 0 {
		spellFailed()
		return
	}
	setPartyMapStart()
	updateOutdoorQuadrant()
	playSfx(0x33)
	refreshView()
}

func spellFeebleMind()  { spellAttack(0, 9, 1, 0x12, 0xe) }
func spellDeadlySwarm() { spellAttack(0x28, 0, 1, 0xd, 0xe) }

func spellTeleport() {
	if gMapSpellFlags[*gMapQuadrant][0] == 0 {
		spellFailed()
		return
	}
	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	sprintf(gTextBuf, "\x03cTeleport\nHow many squares %s (1-9)\n", str(gDirectionNames[*gPartyDir]))
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	n := int(int16(inputNumber(1, 200)))
	closeWindows(1)
	popButtons()
	*gGameMode = save
	if n == 0 {
		return
	}
	mask, want := 0x7777, 0x1111
	if *gOutdoors != 0 {
		mask, want = 0x70, 0
	}
	x, y := int(*gPartyX), int(*gPartyY)
	switch *gPartyDir {
	case 0:
		y += n
	case 1:
		y -= n
	case 3:
		x -= n
	case 2:
		x += n
	}
	if *gPartyDir < 4 {
		if getCellWallsRegion(x, y, mask) == want {
			spellFailed()
			return
		}
		*gPartyX, *gPartyY = uint8(x), uint8(y)
	}
	updateOutdoorQuadrant()
	*gSpellMovedParty = 1
	memset(gViewCaption, 0, 0x96)
	playSfx(0x33)
}

func spellFingerOfDeath() { spellAttack(0, 0xb, 1, 0x12, 0xe) }
func spellCureParalysis() { spellCureCond(0x2f, 0x34, 11) }
func spellParalyze()      { spellAttack(0, 0xa, 1, 0x12, 0xe) }

// spellElementBlast is dragon breath and elemental storm: dmg of a picked element.
func spellElementBlast(dmg, a int) {
	*gAttackDamage = int16(dmg)
	if *gAttackType = int16(spellPickElement(a)); *gAttackType != -1 {
		*gAttackScope = 2
		e := int(uint16(*gAttackType) - 2)
		playSfx(int(gElementSounds[e]))
		spellAttackMonsters(int(gElementPowAnims[e]))
	}
}

func spellDragonBreath() { spellElementBlast(getCharLevel(curChr())*5, 0x31) }

func spellSuperShelter() {
	if gMapSpellFlags[*gMapQuadrant][3] == 0 {
		spellFailed()
		return
	}
	save := *gGameMode
	*gGameMode = 0xc
	playSfx(0x34)
	restParty()
	*gGameMode = save
}

func spellFieryFlail() { spellAttack(0x64, 2, 0, 0xd, 2) }

func spellCreateFood() {
	*gFood += uint16(*gPartyCount)
	playSfx(0x14)
}

func spellTownPortal() {
	var t int
	if gMapSpellFlags[*gMapQuadrant][4] == 0 {
		spellFailed()
		return
	}
	pushButtons()
	setAnykeyButtons()
	save := *gGameMode
	*gGameMode = 0xff
	sprintf(gTextBuf, "\x03cTown Portal\x03l\n\n\t0101. Fountain Head\n\t0102. Bay Watch\n\t0103. Wildabar\n\t0104. Swamp Town\n\t0105. Blistering Heights\x03c\n\nTo which Town (1-5)\n\n")
	openTextWindow(0x48, 0x25, 0xb0, 0x7e, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	for {
		if t = int(int16(inputNumber(1, 0xa0))); t <= 5 {
			break
		}
	}
	closeWindows(1)
	popButtons()
	*gGameMode = save
	if t != 0 {
		playSfx(0x33)
		popIcons()
		enterMaze(t)
		pushIcons("cast.icn")
		setPartyMapStart()
		*gSpellMovedParty = 1
	}
}

func spellStoneToFlesh()    { spellCureCond(0x36, 0x34, 14) }
func spellRechargeItem()    { spellItemScreen(0x37, 4, 0) }
func spellFantasticFreeze() { spellAttack(0x28, 4, 1, 0xf, 9) }
func spellDuplication()     { spellItemScreen(0x39, 5, 0) }
func spellDisintegrate()    { spellAttack(0, 0xf, 1, 0xd, 0xe) }

func spellRaiseDead() {
	k := spellPickPartyMember(0x3b)
	if k == -1 {
		return
	}
	p := &gParty[k]
	if p.Cond[13] == 0 {
		spellFailed()
		return
	}
	p.Cond[12], p.Cond[13] = 0, 0
	p.HP.Set(0)
	playSfx(0x34)
	spellHealChar(p, 1)
	if p.Endur--; int(p.Endur) < 1 {
		p.Endur = 1
	}
	drawPartyPortraits()
}

func spellHalfForMe() {
	k := spellPickPartyMember(0x3c)
	if k == -1 {
		return
	}
	p := &gParty[k]
	c := int(getWorstCondition(p))
	if p == curChr() || c >= 13 && c <= 15 {
		spellFailed()
		return
	}
	m := int(int16(calcMaxHp(p)))
	charTakeDamage(curChr(), (m-p.HP.Get())>>1)
	playSfx(0x14)
	spellHealChar(p, m)
	drawPartyPortraits()
}

func spellEtherealize() {
	if gMapSpellFlags[*gMapQuadrant][6] == 0 {
		spellFailed()
		return
	}
	x, y, d := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
	if getCellWallsRegion(dxFrom7(d, 0)+x, dyFrom7(d, 0)+y, 0x7777) == 0x1111 {
		spellFailed()
		return
	}
	*gPartyX += uint8(dxFrom7(d, 0))
	*gPartyY += uint8(dyFrom7(d, 0))
	playSfx(0x33)
	save := *gGameMode
	*gGameMode = 0xff
	refreshView()
	*gGameMode = save
	*gSpellMovedParty = 1
}

func spellDancingSword()   { spellAttack(randomRange(6, 0xe)*getCharLevel(curChr()), 0, 1, 0x12, 0xe) }
func spellPrismaticLight() { spellAttack(0x50, randomRange(0, 6), 2, 0x12, 0xe) }

func spellMoonRay() {
	spellAttack(0x1e, 6, 2, 0x10, 0xd)
	for i := 0; i < int(*gPartyCount); i++ {
		playSfx(0x34)
		spellHealChar(&gParty[i], randomRange(1, 0x1e))
	}
	drawPartyPortraits()
}

func spellMassDistortion() { spellAttack(0, 0xd, 1, 0x12, 0xe) }
func spellEnchantItem()    { spellItemScreen(0x42, 6, 0) }
func spellIncinerate()     { spellAttack(0xfa, 2, 0, 0xd, 0) }
func spellElementalStorm() { spellElementBlast(0x96, 0x44) }
func spellHolyWord()       { spellAttack(0, 0xc, 2, 0x12, 0xe) }

func ageBy5(c *Chr) {
	if int(c.AgeMod)+5 > 0xfa {
		c.AgeMod = 0xfa
	} else {
		c.AgeMod += 5
	}
}

func spellResurrect() {
	k := spellPickPartyMember(0x46)
	if k == -1 {
		return
	}
	p := &gParty[k]
	if p.Cond[15] == 0 {
		spellFailed()
		return
	}
	playSfx(0x34)
	spellHealChar(p, 0)
	p.Cond[15] = 0
	p.Endur--
	if int(p.Endur) < 1 {
		p.Endur = 1
	}
	ageBy5(p)
	drawPartyPortraits()
}

func spellMegaVolts() { spellAttack(0x96, 3, 1, 0xe, 4) }
func spellInferno()   { spellAttack(0xfa, 2, 1, 0xd, 0) }
func spellSunRay()    { spellAttack(0xc8, 6, 2, 0x10, 0xd) }
func spellImplosion() { spellAttack(0x3e8, 6, 0, 0x12, 0xd) }
func spellStarBurst() { spellAttack(0x1f4, 0, 2, 0xd, 2) }

func spellDivineIntervention() {
	ageBy5(curChr())
	for i := 0; i < int(*gPartyCount); i++ {
		p := &gParty[i]
		v := p.Cond[15]
		p.Cond = [0x10]uint8{}
		p.Cond[15] = v
		if v == 0 {
			p.HP.Set(calcMaxHp(p))
		}
	}
	playSfx(0x14)
	drawPartyPortraits()
}
