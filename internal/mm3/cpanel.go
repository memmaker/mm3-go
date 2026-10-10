// Port of CPANEL.C (overlay 1816): experience, rest, gold/gems/food, the yes/no box, the
// control panel (sound, load, save, quit, Mr Wizard), game info, the automap, portraits,
// maze graphics and maze loading.
//
// cpanel.c's struct maze has surface_m1 at 0x2ff and edge[4] at 0x308 (North, East, South,
// West): mazeB(m, off) is the byte at that offset of g_maps[m] (m may be -1 .. 4).
package mm3

// Views where cpanel.c types a global differently from ds_gen.go (far pointers held in longs).
var (
	gPartyFacesF    = arr[Far](adrPartyFaces)
	gDseFaceGfxF    = at[Far](adrDseFaceGfx)
	gTilesetGfxF    = at[Far](adrTilesetGfx)
	gSkyGfxF        = at[Far](adrSkyGfx)
	gOutdoorObjGfxF = arr[Far](adrOutdoorObjGfx) // far *g_outdoor_obj_gfx[7]
	gMonD1l1F       = arr[Far](adrMonD1l1)       // far *g_mon_d1l1[]: the wall graphics by g_wall_gfx_ids
)

func mazeB(m, off int) *uint8 { return &ds[adrMaps+m*0x340+off] }

func getNextLevelExp(c *Chr) int32 {
	var sh, k int
	if lvl := int(c.Level); lvl < 12 {
		k = 0
		sh = lvl - 1
	} else {
		k = lvl - 11
		sh = 10
	}
	return int32(k)*1024000 + int32(gClassBaseExp[c.Cls])<<sh
}

func getExpToNextLevel(c *Chr) int32 {
	d := getNextLevelExp(c) - getTotalExp(c)
	if d < 1 {
		return 0
	}
	return d
}

func getTotalExp(c *Chr) int32 {
	var sh, k int
	i := int(c.Level) - 1
	if i < 12 {
		if i == 0 {
			return int32(c.Exp.Get())
		}
		k = 0
		sh = i - 1
	} else {
		k = i - 11
		sh = 10
	}
	t := int32(k)*1024000 + int32(gClassBaseExp[c.Cls])<<sh
	return t + int32(c.Exp.Get())
}

func charCannotAct(c *Chr) int {
	switch getWorstCondition(c) {
	case 8, 11, 12, 13, 14, 15:
		sprintf(gTmpText, gostr(str(*gMsgCantAct)), c.Name[:])
		showMessage(gTmpText)
		return 1
	}
	return 0
}

func loadMainIcons() {
	freeAndNull(gCurIcons)
	*gCurIcons = loadGraphic("main.icn")
	*gCpanelFrameImg = *gGlobalIcons
	*gCpanelButtonsImg = *gCurIcons
}

func drawMainIcons() {
	var buf [10]byte
	sprintf(buf[:], "\x03b\x05%04X", adrMainSprites)
	printWindowText(buf[:])
}

func setupMainUi() {
	loadMainIcons()
	setMainButtons()
}

func restParty() {
	var cls, old, n, starve int

	if getCellFlags(int(*gPartyX), int(*gPartyY), 4) != 0 || gMaps[*gMapQuadrant].Rest == 0 {
		showMessage(cs("Too dangerous to rest here!"))
		return
	}
	for i := 0; i < int(*gPartyCount); i++ {
		for j := 0; j <= 6; j++ {
			if getCharStat(&gParty[i], j, 0) < 1 {
				if askYesNo(cs("Some Chars may die. Rest anyway?"), 0) != 0 {
					goto rest
				}
				return
			}
		}
	}
rest:
	for i := 0; i < int(*gPartyCount); i++ {
		gParty[i].Cond[8] = 1 // asleep
	}
	drawPartyPortraits()
	old = int(*gGameMode)
	*gGameMode = 5
	if old != 12 {
		for i := 0; i < 10; i++ {
			endPartyTurn()
			refreshView()
			if *gGameMode == 1 {
				*gGameMode = uint8(old)
				return
			}
		}
		if *gOutdoors != 0 {
			passTime(0x17c)
		} else {
			passTime(0x1d6)
		}
	} else {
		passTime(0x1e0)
	}
	clearTempEffects(0)
	for i := 0; i < int(*gPartyCount); i++ {
		p := &gParty[i]
		p.Cond[8] = 0
		if *gFood != 0 {
			if cls = int(getWorstCondition(p)); cls < 13 || cls > 15 {
				*gFood--
				n++
				p.Cond[12], p.Bless, p.Shield, p.Holy, p.Hero = 0, 0, 0, 0, 0
				p.HP.Set(calcMaxHp(p))
				p.SP.Set(calcMaxSp(p))
			}
		} else {
			starve = 1
		}
	}
	drawPartyPortraits()
	*gGameMode = uint8(old)
	*gViewDirty = 1
	refreshView()
	s := "Hit Pts and Spell Pts restored."
	if starve != 0 {
		s = "\f07The Party is Starving!\fd"
	}
	sprintf(gTextBuf2, "\v000\t0008 hours pass.  Rest complete.\n%s\n%d food consumed.", s, n)
	showMessage(gTextBuf2)
	updatePartyDisabled()
}

func clearTempEffects(keep int) {
	for i := 0; i < int(*gPartyCount); i++ {
		p := &gParty[i]
		p.MightTmp, p.IntelTmp, p.PersTmp, p.EndurTmp, p.SpeedTmp, p.AccTmp, p.LuckTmp, p.LevelTmp, p.AcTmp = 0, 0, 0, 0, 0, 0, 0, 0, 0
		p.Bless, p.Shield, p.Holy, p.Hero = 0, 0, 0, 0
		p.FireResTmp, p.ColdResTmp, p.ElecResTmp, p.PoisonResTmp, p.EnergyResTmp, p.MagicResTmp = 0, 0, 0, 0, 0, 0
	}
	if *gWizeyeShown != 0 && keep == 0 {
		closeWindows(1)
		*gWizeyeShown = 0
	}
	*gWalkOnWater, *gLevitate, *gWizardEye = 0, 0, 0
	*gLightCount, *gFireResist, *gElecResist, *gColdResist, *gPoisonResist = 0, 0, 0, 0, 0
}

func showNotEnough(kind, who, none int) {
	pushButtons()
	setAnykeyButtons()
	if none == 0 {
		sprintf(gTextBuf2, "\x03c\v012Not enough %s in the %s!", str(gResourceNames[kind]), str(gPurseNames[who]))
	} else {
		sprintf(gTextBuf2, "\x03c\v012No %s in the %s!", str(gResourceNames[kind]), str(gPurseNames[who]))
	}
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
	waitMouseRelease()
	getInput()
	closeWindows(1)
	waitMouseRelease()
	popButtons()
}

func payCost(kind int, amt uint32, who int) int {
	switch kind {
	case 0:
		if who != 0 {
			if *gBankGold < amt {
				showNotEnough(0, who, 0)
				return 0
			}
			*gBankGold -= amt
		} else {
			if *gPartyGold < amt {
				showNotEnough(0, who, 0)
				return 0
			}
			*gPartyGold -= amt
		}
	case 1:
		if who != 0 {
			if *gBankGems < amt {
				showNotEnough(1, who, 0)
				return 0
			}
			*gBankGems -= amt
		} else {
			if *gPartyGems < amt {
				showNotEnough(1, who, 0)
				return 0
			}
			*gPartyGems -= amt
		}
	case 2:
		if uint32(*gFood) < amt {
			showNotEnough(2, 0, 0)
			return 0
		}
		*gFood -= uint16(amt)
	}
	return 1
}

var askYesNoBuf [300]byte // ask_yes_no's buf: the window gets its far address

func askYesNo(msg []byte, flags int) int {
	var old, r int
	buf := askYesNoBuf[:]

	sav := *gCurIcons
	*gCurIcons = loadGraphic("confirm.icn")
	*gConfirmSpritesImg = *gCurIcons
	sprintf(buf, "%s\x05%04X", msg, adrConfirmSprites)
	pushButtons()
	*gBtnXs = adrConfirmBtnX
	*gBtnYs = adrConfirmBtnY
	*gBtnWs = adrConfirmBtnW
	*gBtnHs = adrConfirmBtnH
	if flags != 0 {
		if flags&0x80 != 0 {
			strcpy(buf, msg)
			setAnykeyButtons()
		}
		*gConfirmYesX = 0x25
		*gConfirmNoX = 0x5d
		*gConfirmNoY = 0x59
		*gConfirmYesY = 0x59
		gConfirmBtnX[0] = 0x78
		gConfirmBtnX[1] = 0xb0
		gConfirmBtnY[0], gConfirmBtnY[1] = 0x85, 0x85
		openTextWindow(0x4b, 0x24, 0xaa, 0x7e, 0, 0x96, 4, 6, farOf(buf))
	} else {
		*gConfirmYesX = 0x16
		*gConfirmNoX = 0x4e
		*gConfirmNoY = 0x2d
		*gConfirmYesY = 0x2d
		gConfirmBtnX[0] = 0x81
		gConfirmBtnX[1] = 0xb9
		gConfirmBtnY[0], gConfirmBtnY[1] = 0x70, 0x70
		openTextWindow(0x63, 0x3b, 0x8a, 0x52, 0, 0x96, 4, 6, farOf(buf))
	}
	old = int(*gPortraitClickOn)
	*gPortraitClickOn = 0
	waitMouseRelease()
	for {
		r = getInput()
		switch r {
		case 0, 'Y':
			r = 1
			goto done
		case 1, 0x1b, 'N':
			r = 0
			goto done
		}
		if flags&0x80 != 0 {
			r = 0
			break
		}
	}
done:
	closeWindows(1)
	waitMouseRelease()
	popButtons()
	*gPortraitClickOn = uint8(old)
	freeAndNull(gCurIcons)
	*gCurIcons = sav
	return r
}

func saveGame(quiet int) {
	var tmp Far

	reload := 0
	if quiet == 0 {
		if gMaps[*gMapQuadrant].Save == 0 {
			showMessage(cs("\x03c\v000\t000The Gods of Game Restoration deem this area off limits!  Sorry,  no saving in this maze."))
			return
		}
		sprintf(gTextBuf2, "\x03c\v010Saving your game\n\n'%s'.", gSaveName)
		openTextWindow(0x63, 0x3b, 0x8a, 0x52, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
	}
	updateFreeMemKb()
	if gOutdoorObjGfxF[0] != 0 && *gFreeMemKb < 20 {
		freeMazeGfx()
		reload = 1
	}
	if uint8(*gSaveSlot) == 0xff {
		return
	}
	*gRosterFar = memAlloc(0x2382)
	tmp = memAlloc(0x396)
	copy(tmp.B()[:0x396], ds[adrPartyCount:])
	loadPartyData()
	copy(ds[adrPartyCount:adrPartyCount+0x396], tmp.B())
	for i := 0; i < int(*gPartyCountSaved); i++ {
		if int(*gPartyCount) <= i {
			gParty[i].Inn.Set(int(*gInnMap))
			*rosterAt(int(gPartyIds[i])) = gParty[i]
			gPartyIds[i] = 0xff
		} else {
			*rosterAt(int(gPartyIds[i])) = gParty[i]
		}
	}
	*gPartyCountSaved = *gPartyCount
	saveMazeState()
	writeSaveGame()
	freeAndNull(gRosterFar)
	freeAndNull(&tmp)
	if quiet == 0 {
		sprintf(gTextBuf2, "\r\x03c\v010Your Game '%s' has been saved.\n\nPress any key.", gSaveName)
		printWindowText(gTextBuf2)
		if reload != 0 {
			showPleaseWait()
			loadMazeGfx(int(*gCurTown) - 1)
			closeWindows(1)
		}
		flushInput()
		getInput()
		closeWindows(1)
		waitMouseRelease()
	}
}

func formatCpanelOptions() []byte {
	if *gSfxOn != 0 {
		*gStrSfxState = gStrOnOff[0]
		*gCpanelSfxFrame = 1
	} else {
		*gStrSfxState = gStrOnOff[1]
		*gCpanelSfxFrame = 0
	}
	if *gMusicOn != 0 {
		*gStrMusicState = gStrOnOff[0]
		*gCpanelMusicFrame = 1
	} else {
		*gStrMusicState = gStrOnOff[1]
		*gCpanelMusicFrame = 0
	}
	sprintf(gTmpText, "\x05%04X\x03c\v022\t062%s\t141load\v041\t062%s\t141save\v060\t062%d\t141DOS\v079\t102Help",
		adrCpanelSprites, str(*gStrSfxState), str(*gStrMusicState), *gMsgDelay)
	return gTmpText
}

func controlPanel() int {
	var o1, o2, t, old, ret, c, g int
	var img Far
	var redraw bool
	defer dsFree(dsSP)

	setMouseCursor(0, uint32(*gMouseCursor))
	pushButtons()
	ret = 0
	*gBtnXs = adrCtrlpanelBtnX
	*gBtnYs = adrCpanelBtnY
	*gBtnWs = adrCpanelBtnW
	*gBtnHs = adrCpanelBtnH
	*gPortraitClickOn = 0
	adrLst := dsAlloca(0x14)
	lst := at[Spr1](adrLst)
	img = loadGraphic("cpanel.icn")
	*gCpanelSpritesImg = img
	lst.Img.Set(uint32(img))
	lst.Flag.Set(-1)
	lst.End.Set(-1)
	lst.X.Set(0x2f)
	lst.Y.Set(0x3a)
	lst.Frame.Set(1)
top:
	sprintf(gTextBuf, "\x03f\x03cControl Panel\x03r\v022\t045\f27E\fdfx:\t124\f27L\fdoad:\v041\t045\f27M\fdusic:"+
		"\t124\f27S\fdave:\v060\t045\f27D\fdelay:\t124\f27Q\fduit:\v080\t084Mr \f27W\fdizard:%s\t000\v"+
		"096\x02Current Game = %s", formatCpanelOptions(), gSaveName)
	openTextWindow(0x50, 0x1c, 0xb0, 0x78, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	printWindowText(cs("\x01"))
	for {
		c = getInput()
		redraw = false
		switch c {
		case 0, 'E':
			*gSfxOn ^= 1
			if *gSfxOn != 0 {
				sndTimer(0x14)
			} else {
				sndTimer(0)
			}
			redraw = true
		case 1, 'M':
			*gMusicOn ^= 1
			if *gMusicOn != 0 {
				*gMusicLoadPending = 0
				sndPlayMusic(int32(*gMusicBuf))
				if *gSndDev5 != 0 {
					sndTimer(0)
				}
			} else {
				sndPlayMusic(0)
			}
			redraw = true
		case 2, 'D':
			if *gMsgDelay++; *gMsgDelay > 9 {
				*gMsgDelay = 0
			}
			sprintf(gTextBuf2, "\x05%04X\x03c\v061\t063%d", adrLst, *gMsgDelay)
			printWindowText(gTextBuf2)
			flushInput()
			redraw = true
		case 3, 'L':
			if *gGameMode == 2 {
				showMessage(cs("No Loading Allowed in Combat!"))
				break
			}
			closeWindows(1)
			findSaveFiles(cs("*.mm3"))
			if *gSaveCount == 0 {
				break
			}
			if *gSaveCount == 1 {
				*gSaveSlot = 0
			} else {
				if t = selectSavedGame(); t == -1 {
					goto top
				}
				*gSaveSlot = int8(t)
			}
			ret = 1
			copySaveToWork()
			goto out
		case 4, 'S':
			if *gGameMode == 2 {
				showMessage(cs("No Saving Allowed in Combat!"))
			} else {
				saveGame(0)
			}
		case 5, 'Q':
			old = int(*gDebugOverlay)
			*gDebugOverlay = 0
			if askYesNo(cs("Are you sure you want to quit?"), 0) != 0 {
				quitToDos()
			}
			*gDebugOverlay = int8(old)
		case 6, 'W':
			old = int(*gDebugOverlay)
			*gDebugOverlay = 0
			if askYesNo(cs("Are you sure you want Mr. Wizard's Help?"), 0) != 0 {
				closeWindows(1)
				playSfx(0x33)
				freeAndNull(&img)
				enterMaze(1)
				img = loadGraphic("cpanel.icn")
				setPartyMapStart()
				updateOutdoorQuadrant()
				for c = 0; c < int(*gPartyCount); c++ {
					if gParty[c].Level > 1 {
						gParty[c].Level--
					}
				}
				drawWizardEye(1)
				goto out
			}
			*gDebugOverlay = int8(old)
		case 'G':
			g = 1
		case 'O':
			if g == 0 {
				c = 0x1b
				g, o1, o2 = 0, 0, 0
				break
			}
			o1 = 1
		case 'B':
			if g == 0 || o1 == 0 {
				c = 0x1b
				g, o1, o2 = 0, 0, 0
				break
			}
			o2 = 1
		case 'R':
			if g != 0 && o1 != 0 && o2 != 0 {
				*gDebugOverlay = 1
			}
		default:
			c = 0x1b
			g, o1, o2 = 0, 0, 0
		}
		if redraw {
			printWindowText(formatCpanelOptions())
			flushInput()
		}
		if c == 0x1b {
			break
		}
	}
	closeWindows(1)
out:
	flushInput()
	popButtons()
	*gDebugOverlay = 0
	*gPortraitClickOn = 1
	freeAndNull(&img)
	return ret
}

func formatActiveSpells(n *int, out []byte) []byte {
	var light, fire, elec, cold, pois, lev, walk [30]byte
	k := 0
	*n = 0
	sep := func() int { // !k++ ? 10 : 1
		k++
		if k == 1 {
			return 10
		}
		return 1
	}
	if *gLightCount != 0 {
		*n++
		k++
		sprintf(light[:], "\x03l\n\n\t024Light\x03r\t124%u", *gLightCount)
	}
	if *gFireResist != 0 {
		*n++
		sprintf(fire[:], "%c\x03l\n\t024Fire\x03r\t124%u", sep(), *gFireResist)
	}
	if *gElecResist != 0 {
		*n++
		sprintf(elec[:], "%c\x03l\n\t024Elec\x03r\t124%u", sep(), *gElecResist)
	}
	if *gColdResist != 0 {
		*n++
		sprintf(cold[:], "%c\x03l\n\t024Cold\x03r\t124%u", sep(), *gColdResist)
	}
	if *gPoisonResist != 0 {
		*n++
		sprintf(pois[:], "%c\x03l\n\t024Poison/Acid\x03r\t124%u", sep(), *gPoisonResist)
	}
	if *gLevitate != 0 {
		*n++
		sprintf(lev[:], "%c\x03l\n\t024Levitate\x03r\t124%u", sep(), *gLevitate)
	}
	if *gWalkOnWater != 0 {
		*n++
		sprintf(walk[:], "%c\x03l\n\t024Walk on Water", sep())
	}
	sprintf(out, "%s%s%s%s%s%s%s", light[:], fire[:], elec[:], cold[:], pois[:], lev[:], walk[:])
	return out
}

func showGameInfo() {
	var n int
	var buf [200]byte

	setMouseCursor(0, uint32(*gMouseCursor))
	pushButtons()
	setAnykeyButtons()
	h := int(*gMinutes / 60)
	hh := h
	if h > 12 {
		hh = h - 12
	} else if h == 0 {
		hh = 12
	}
	ap := 'a'
	if h > 11 {
		ap = 'p'
	}
	sprintf(gTextBuf, "\x03c\v001\f27Might and Magic III\fd\nGame Information\n\nToday is \f27%ssday\fd\n\n"+
		"\t032Time\t072Day\t112Year\n\t032\f27%d:%02d%c\t072%u\t112%u\fd%s",
		str(gDayNames[*gDay%10]), hh, *gMinutes%60, int(ap), *gDay, *gYear, formatActiveSpells(&n, buf[:]))
	hgt := 0
	if n > 0 {
		hgt = n*9 + 0x10
	}
	openTextWindow(0x58, 0x1c, 0xa0, hgt+0x56, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	getInput()
	closeWindows(1)
	flushInput()
	popButtons()
}

func getAutomapOrigin(dx, dy *int8) {
	q := int(*gMapQuadrant)
	px, py := int(*gPartyX), int(*gPartyY)
	*dx = -8
	*dy = 8
	if *mazeB(q, 0x308+3) == 0 && px < 8 {
		*dx = int8(int(gRevCoord16[px]) - 16)
	} else if *mazeB(q, 0x308+1) == 0 && px > 7 {
		*dx = int8(int(gRevCoord16[px&15]) - 16)
	}
	if *mazeB(q, 0x308+2) == 0 && py < 8 {
		*dy = int8(16 - py)
	} else if *mazeB(q, 0x308+0) == 0 && py > 7 {
		*dy = int8((16 - py) & 15)
	}
}

// automapQuad is the g_maps index of cell (px+x, py+y).
func automapQuad(x, y int) int {
	px, py := int(*gPartyX), int(*gPartyY)
	m := int(*gMapQuadrant)
	if px+x > 15 {
		m++
	} else if px+x < 0 {
		m--
	}
	if py+y > 15 {
		m += 2
	} else if py+y < 0 {
		m -= 2
	}
	return m
}

func showAutomap() {
	var v, m, x, y, t, n int
	var ox, oy int8
	var seen [16][16]byte
	dl := gDrawList

	getAutomapOrigin(&ox, &oy)
	setMouseCursor(0, uint32(*gMouseCursor))
	setAnykeyButtons()
	mn := int(*gMapNum) - 1
	if *gMapNum > 0x68 {
		mn = int(*gMapNum) - 0x29
	}
	var dir []byte
	if partyHasSkill(6) != 0 {
		dir = str(gDirectionNames[*gPartyDir])
	}
	sprintf(gTextBuf, "\x03c%s\x05%04X\x03l\v139\t000X = %d\x03r\t160Y = %d\x03c\t000%s",
		str(gMapNamesFrom1[mn]), adrDrawList, *gPartyX, *gPartyY, dir)
	memset(gDrawListB, 0, 4000)
	dl[0] = -1
	*gDrawListImg = Far(*gTilesetGfx)
	n = 3
	px, py := int(*gPartyX), int(*gPartyY)
	cell := func(i, j, frame int) {
		dl[n] = int16(i * 10)
		n++
		dl[n] = int16(j*8 + 10)
		n += 2
		dl[n] = int16(frame)
		n++
	}
	if *gOutdoors != 0 {
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if isCellVisited(px+x, py+y) == 0 {
					continue
				}
				seen[i][j] = 1
				v = getCellWalls(px+x, py+y, 0x70) >> 4
				m = automapQuad(x, y)
				if v != 0 {
					cell(i, j, int(*mazeB(m, 0x2ff+v)))
				} else {
					cell(i, j, 0)
				}
			}
		}
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if seen[i][j] == 0 {
					continue
				}
				v = getCellWalls(px+x, py+y, 7)
				if v == 0 {
					continue
				}
				m = automapQuad(x, y)
				cell(i, j, int(*mazeB(m, 0x2ff+v)))
			}
		}
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if seen[i][j] == 0 {
					continue
				}
				v = getCellWalls(px+x, py+y, 0x700) >> 8
				if v == 0 {
					continue
				}
				cell(i, j, v+0x16)
			}
		}
		if ox != -8 {
			dl[n] = int16(px * 10)
		} else {
			dl[n] = 0x46
		}
		n++
		if oy != 8 {
			dl[n] = int16(int(gRevCoord16[py])*8 + 10)
		} else {
			dl[n] = 0x42
		}
		n += 2
		if partyHasSkill(6) != 0 {
			dl[n] = int16(int(*gPartyDir) + 0x13)
		} else {
			dl[n] = 0x1e
		}
		n++
	} else {
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if isCellVisited(px+x, py+y) == 0 {
					continue
				}
				seen[i][j] = 1
				if getCellWalls(px+x, py+y, 0x7777) == 0x1111 {
					continue
				}
				cell(i, j, 0)
			}
		}
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if seen[i][j] == 0 {
					continue
				}
				if getCellWalls(px+x, py+y, 0x7777) == 0x1111 {
					continue
				}
				t = -1
				switch getCellWalls(px+x, py+y, 0x7000) >> 12 {
				case 1, 3:
					t = 2
				case 4:
					t = 12
				case 5:
					t = 10
				case 2:
					t = 6
				case 6:
					t = 8
				case 7:
					t = 4
				}
				if t != -1 {
					if i != 15 {
						cell(i, j, t)
					} else {
						cell(i, j, (t>>1)+0x11)
					}
				}
				t = -1
				switch getCellWalls(px+x, py+y, 7) {
				case 1, 3:
					t = 3
				case 4:
					t = 13
				case 5:
					t = 11
				case 2:
					t = 7
				case 6:
					t = 9
				case 7:
					t = 5
				}
				if t != -1 {
					if j != 15 {
						cell(i, j, t)
					} else {
						cell(i, j, (t>>1)+0x17)
					}
				}
			}
		}
		y = int(oy)
		for j := 0; j < 16; j++ {
			y--
			x = int(ox)
			for i := 0; i < 16; i++ {
				x++
				if seen[i][j] == 0 || getCellWalls(px+x, py+y, 0x7777) == 0x1111 {
					cell(i, j, 1)
				}
			}
		}
		if ox != -8 {
			dl[n] = int16(px*10 + 2)
		} else {
			dl[n] = 0x46
		}
		n++
		if oy != 8 {
			dl[n] = int16(int(gRevCoord16[py])*8 + 12)
		} else {
			dl[n] = 0x42
		}
		n += 2
		if partyHasSkill(6) != 0 {
			dl[n] = int16(int(*gPartyDir) + 14)
		} else {
			dl[n] = 0x1e
		}
		n++
	}
	dl[n] = -1
	checkAutoListSize(n)
	openTextWindow(0x48, 0x14, 0xb0, 0xa3, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	flushInput()
	getInput()
	setMainButtons()
	closeWindows(1)
	waitMouseRelease()
}

func loadFaces(n int) {
	var buf [20]byte

	hire := n > 8
	for i := 0; i < n; i++ {
		face := i
		if !hire {
			face = int(gPartyIds[i])
			freeAndNull(&gPartyFacesF[i])
		}
		if face < 20 {
			sprintf(buf[:], "%s%d.fac", str(gRaceFacePrefixes[face>>2]), face%4+1)
		} else {
			sprintf(buf[:], "hire%d.fac", face-20)
		}
		if hire {
			gRosterFaces[i] = loadGraphic(gostr(buf[:]))
		} else {
			gPartyFacesF[i] = loadGraphic(gostr(buf[:]))
		}
	}
	if !hire {
		*gDseFaceGfxF = loadGraphic("dse.fac")
	}
}

func restorePrevMusic() {
	var buf [16]byte

	strcpy(buf[:], gCurMusic)
	strcpy(gCurMusic, gPrevMusic)
	playMusic(gCurMusic, 0xa0)
	strcpy(gPrevMusic, buf[:])
}

func freeMazeGfx() {
	for i := 0; i < 7; i++ {
		freeAndNull(&gOutdoorObjGfxF[i])
	}
	freeAndNull(gSkyGfxF)
	freeAndNull(gWaterGfx)
	freeAndNull(gTilesetGfxF)
}

func loadSky(n int) {
	freeAndNull(gSkyGfxF)
	if mapHasDaylight() != 0 {
		if *gMinutes >= 300 && *gMinutes < 1260 {
			*gIsNight = 0
		} else {
			*gIsNight = 1
		}
		*gSkyGfxF = loadGraphic(gostr(str(gSkyFiles[*gIsNight])))
	} else {
		sprintf(gTmpText, "%s.sky", str(gMapGfxPrefixes[gMapGfxSetByNum[n]]))
		*gSkyGfxF = loadGraphic(gostr(gTmpText))
	}
}

func loadMazeGfx(n int) {
	var buf [20]byte

	updateFreeMemKb()
	*gMapNum = uint8(n + 1)
	*gCurTown = *gMapNum
	if n == -1 {
		n = int(*gMazeGfxIndex)
	} else {
		*gMazeGfxIndex = int8(n)
	}
	if n == 0x69 || n == 0x68 {
		n = 15
	}
	updateFreeMemKb()
	freeMazeGfx()
	updateFreeMemKb()
	if n < 40 {
		*gTilesetGfxF = loadGraphic(gostr(str(gTilesetFiles[gMapGfxSets[n]])))
	} else {
		*gTilesetGfxF = loadGraphic("out.til")
	}
	updateFreeMemKb()
	if n < 40 {
		for i := 0; i < 4; i++ {
			sprintf(buf[:], "%swl%u.vga", str(gMapGfxPrefixes[gMapGfxSets[n]]), gWallGfxIds[i])
			updateFreeMemKb()
			gMonD1l1F[gWallGfxIds[i]] = loadGraphic(gostr(buf[:]))
			updateFreeMemKb()
		}
	} else {
		q := int(*gMapQuadrant)
		for i := 0; i < 7; i++ {
			if s := *mazeB(q, 0x2ff+i+1); s != 0 {
				sprintf(buf[:], "%s.vga", str(gOutdoorObjFiles[s]))
				updateFreeMemKb()
				gOutdoorObjGfxF[i] = loadGraphic(gostr(buf[:]))
				updateFreeMemKb()
			}
		}
		*gWaterGfx = loadGraphic("water.vga")
		updateFreeMemKb()
	}
	loadSky(n + 1)
	updateFreeMemKb()
	strcpy(buf[:], gCurMusic)
	switch {
	case n < 5:
		strcpy(gCurMusic, cs("medieval.m"))
	case n < 15:
		strcpy(gCurMusic, cs("caves.m"))
	case n < 23:
		strcpy(gCurMusic, cs("eerie.m"))
	case n < 28:
		strcpy(gCurMusic, cs("city.m"))
	case n < 33:
		strcpy(gCurMusic, cs("eerie.m"))
	case n < 40:
		strcpy(gCurMusic, cs("cyber.m"))
	default:
		strcpy(gCurMusic, cs("venture.m"))
	}
	if stricmp(gCurMusic, buf[:]) != 0 {
		playMusic(gCurMusic, 0xc0)
		strcpy(gPrevMusic, buf[:])
	}
	if *gSndDev5 != 0 {
		playSfx(0)
	}
}

func loadMazeData(n int) {
	var p Far
	var buf [20]byte

	saveLoadedMazes()
	*gStartMapIdx = -1
	memset(gQuadrantMaps, 0, 4)
	*gForceZone2Entered, *gForceZone1Entered = 0, 0
	for j := 0; j < 4; j++ {
		if j != 0 {
			n = int(*mazeB(int(gAdjMazeSrc[j]), 0x308+j%2))
		} else {
			n++
		}
		if n != 0 && gQuadrantMaps[gAdjMazeSrc[j]] != 0xff {
			sprintf(buf[:], "maze%02u.dat", n)
			p = loadCCFile(gostr(buf[:]))
			copy(bytesOf(&gMaps[j]), p.B()[:0x340])
			freeAndNull(&p)
			if gMaps[j].StartPos != 0 && uint8(*gStartMapIdx) == 0xff {
				*gStartMapIdx = int8(j)
			}
		} else {
			n = 0
		}
		if n != 0 {
			gQuadrantMaps[j] = uint8(n)
		} else {
			gQuadrantMaps[j] = 0xff
		}
	}
	if testBit(gGameFlags, 0x6e) != 0 {
		for j := 0; j < 4; j++ {
			if gQuadrantMaps[j] == 0x2e {
				for i := 0; i < 12; i++ {
					gMaps[j].Cell[gMazePatchY46[i]][gMazePatchX[i]].Set(0x50)
				}
			} else if gQuadrantMaps[j] == 0x2f {
				for i := 0; i < 12; i++ {
					gMaps[j].Cell[gMazePatchY47[i]][gMazePatchX[i]].Set(0x50)
				}
			}
		}
	}
}

func freeMazeSprites() {
	for i := 0; i < 5; i++ {
		freeAndNull(&gObjSprites[i])
	}
	for i := 0; i < 3; i++ {
		freeAndNull(&gMonSprites[i])
	}
}

func unloadMaze() {
	saveMazeState()
	freeMazeSprites()
	freeMazeGfx()
	freeAndNull(gCurIcons)
	freeAndNull(gMazeScript)
	freeAndNull(gMapText)
}

func enterMaze(n int) {
	n--
	if n >= 40 && n != 0x69 && n != 0x68 {
		*gOutdoors = 1
	} else {
		*gOutdoors = 0
	}
	unloadMaze()
	setupMainUi()
	*gMapQuadrant = 0
	loadMazeData(n)
	loadMazeGfx(n)
	loadMaze(n)
	addUiGadgetSprites()
	if *gGameMode != 0 && *gGameMode != 9 {
		for *gDimLevel = 0; *gDimLevel < 5; *gDimLevel++ {
			sprintf(gTextBuf, "\x03d\x03m%u\x05%04X", *gDimLevel, adrBorderSprites)
			printWindowText(gTextBuf)
		}
	}
	*gLootGold, *gLootGems, *gLootCount = 0, 0, 0
	*gLightChanged = 1
	memset(gViewCaption, 0, 0x96)
}
