// Port of TOWN.C (overlay 1806): hireling charges, bank interest, the map teleporter, the
// jester, the arena, freeing the maze resources and the main walk loop (game_main_loop).
//
// town.c's struct chr has arena_wins at 0x4e (Chr.Award[0x15]); g_maps is u8[][0x340] there
// (mazeB). game_main_loop's gotos into other cases (turn, fwd, moved, turned) become an
// action that the code after the key switch carries out, in the C's fall-through order.
package mm3

func chargeHirelings() {
	var paid, saved, n, flag int
	var p Far
	var cost uint32
	var buf [7][0x32]byte

	pushButtons()
	setAnykeyButtons()
	for i := 0; i < int(*gPartyCount); i++ {
		if gPartyIds[i] < 0x14 {
			continue
		}
		if *gGameMode == 9 {
			continue
		}
		n++
		c := &gParty[i]
		switch getWorstCondition(c) {
		case 13, 14, 15:
			flag = 1
		default:
			flag = 0
		}
		cost = uint32(int32(int16(int(c.Level)*int(c.Level))) * 20)
		if flag != 0 {
			paid = flag
		} else if *gPartyGold >= cost {
			paid = 1
		} else {
			paid = 0
		}
		var shown uint32
		if flag == 0 {
			shown = cost
		}
		status := "Left"
		if flag != 0 {
			status = "UnAble"
		} else if paid != 0 {
			status = "Paid"
		}
		sprintf(buf[n-1][:], "%s\x03r\t080%u\x03l\t101%lu\x03c\t180%s\x03l\n", c.Name[:], c.Level, shown, status)
		if paid != 0 {
			if flag == 0 {
				*gPartyGold -= cost
			}
		} else {
			p = memAlloc(0x12f)
			copy(p.B(), bytesOf(&gParty[i]))
			saved = int(gPartyIds[i])
			gPartyIds[i] = 0xff
			freeAndNull(&gPartyFacesF[i])
			compactPartyFaces()
			compactParty()
			*gPartyCount--
			copy(bytesOf(&gParty[*gPartyCount]), p.B()[:0x12f])
			freeAndNull(&p)
			gPartyIds[*gPartyCount] = uint8(saved)
			i--
		}
	}
	if n != 0 {
		drawPartyPortraits()
		sprintf(gTextBuf, "\x03cHireling Charge Breakdown\x03l\n\nName\t065Lvl\t100Cost\t158Status\n\x02%s%s%s%s%s%s%s\x01\v106\x03cPress a Key!",
			buf[0][:], buf[1][:], buf[2][:], buf[3][:], buf[4][:], buf[5][:], buf[6][:])
		openTextWindow(8, 8, 0xd8, 0x84, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		flushInput()
		getInput()
		closeWindows(1)
	}
	popButtons()
}

func addBankInterest() {
	*gBankGold += *gBankGold / 100
	*gBankGems += *gBankGems / 100
}

func transportBoxTeleport() {
	var m int
again:
	pushButtons()
	setAnykeyButtons()
	sprintf(gTmpText, "Which Map (1-64) - Current = %u\n\n\t050", *gCurTown)
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, dsFar(adrTmpText))
	for {
		m = int(int16(inputNumber(2, 0xc8)))
		if m <= 0x40 {
			break
		}
	}
	closeWindows(1)
	popButtons()
	switch m {
	case 0x18, 0x19, 0x1a, 0x1d, 0x1e, 0x1f, 0x28:
		sprintf(gTmpText, "Maze %d is not within my reach!", m)
		showMessage(gTmpText)
		goto again
	}
	if m >= 1 {
		enterMaze(m)
		setPartyMapStart()
		updateOutdoorQuadrant()
		drawWizardEye(1)
	}
}

func drinkMightPotion() { ds[*gCurChar+0x15] += 5 }

func tellJesterJoke() {
	refreshView()
	loadTextTable(cs("jester.bin"))
	strncpy(gTextBuf2, gTextTable.Add(int(gTextOffsets[randomRange(0, 0x5e)])).B(), 0x12c)
	faceYesNoDialog(5, cs("Joke of the Day"), gTextBuf2, 1, 0)
	closeWindows(1)
	freeAndNull(gTextTable)
}

// arenaMonster is one monster record of the arena's waves, at slot g_mouse_x.
func arenaMonster(bank, t int) {
	mx := *gMouseX
	gMonX[mx] = int16(randomRange(1, 6) + 6)
	gMonY[mx] = int16(randomRange(1, 6) + 6)
	gMonSpriteBank[mx] = int16(bank)
	gMonFrames[mx] = int16(randomRange(0, int(gMonFrameCounts[t])))
	gMonVisible[mx] = 0
	gMonHp[mx] = gMonMaxHp.I16(2 * t)
	gMonType[mx] = int16(t)
	*gMouseX++
}

func runArena() {
	var i, k, j int
	var t [16]byte

	refreshView()
	playMusic(cs("combat.m"), 0xc0)
	if *gMonCount != 0 {
		for i = 0; i < int(*gMonCount); i++ {
			if (gMonY[i]|gMonX[i])&0xff != 0x80 {
				strncpy(gTextBuf2, mapText(4), 0x12c)
				faceYesNoDialog(8, cs("The Arena"), gTextBuf2, 1, 0)
				*gPartyX, *gPartyY = 7, 7
				*gPartyDir = 0
				closeWindows(1)
				goto done
			}
		}
		strncpy(gTextBuf, mapText(3), 0x12c)
		for i = 0; i < int(*gPartyCount); i++ {
			c := &gParty[i]
			c.Award[0x15]++ // arena_wins
			c.Exp.Set(c.Exp.Get() + uint32(int32(int16(int(c.Award[0x15])*1000))))
		}
		sprintf(gTextBuf2, gostr(gTextBuf), gParty[0].Award[0x15], uint32(gParty[0].Award[0x15])*1000)
		faceYesNoDialog(8, cs("The Arena"), gTextBuf2, 1, 0)
		*gMapNum = 1
		*gMonCount = 0
		i = int(*gGameMode)
		*gGameMode = 0
		exitBuilding(2, 4)
		*gGameMode = uint8(i)
		closeWindows(1)
		for i = 0; i < 5; i++ {
			sprintf(gTextBuf, "\x03d\x03m%u\x05%04X", i, adrBorderSprites)
			printWindowText(gTextBuf)
		}
	done:
		setPartyMapStart()
		drawWizardEye(1)
	} else {
		for i = 0; i < 16; i++ {
			t[i] = byte(i / 4)
		}
		strncpy(gTextBuf, mapText(0), 0x12c)
		freeAndNull(gCurIcons)
		i = buildingEntryDialog(gTextBuf, t[:], 0x10, 7)
		pushButtons()
		setAnykeyButtons()
		strncpy(gTextBuf, mapText(i+1), 0x12c)
		sprintf(gTextBuf2, "%s\x05%04X", gTextBuf, adrBldgFaceSprites)
		printWindowText(gTextBuf2)
		flushInput()
		drawMainIcons()
		buildingWaitInput()
		for *gMouseX = 0; *gMouseX < 2; *gMouseX++ {
			freeAndNull(&gBldgFaceImgs[*gMouseX])
		}
		popButtons()
		closeWindows(1)
		if i == 0 {
			*gMapNum = 1
			exitBuilding(2, 4)
			setPartyMapStart() // goto done
			drawWizardEye(1)
			refreshView()
			return
		}
		*gPartyX, *gPartyY = 7, 7
		*gPartyDir = 0
		*gMouseX = 0
		*gMonCount = 0
		i = int(gParty[0].Award[0x15]) + 5
		if i > 0x50 {
			i = 0x50
		}
		arenaMonster(0, i)
		sprintf(gTmpText, "%s.mon", str(gMonFileNames[i]))
		gMonSprites[0] = loadGraphic(gostr(gTmpText))
		*gMonCount++
		*gMouseY = int16(randomRange(1, int(*gPartyCount)))
		k = randomRange(1, i-1)
		if k > 0x50 {
			k = 0x50
		}
		sprintf(gTmpText, "%s.mon", str(gMonFileNames[k]))
		gMonSprites[1] = loadGraphic(gostr(gTmpText))
		for j = 0; j < int(*gMouseY); j++ {
			arenaMonster(1, k)
			*gMonCount++
		}
		*gMouseY = int16(randomRange(1, int(*gPartyCount)))
		k = randomRange(1, i-1) >> 1
		if k > 0x50 {
			k = 0x50
		}
		sprintf(gTmpText, "%s.mon", str(gMonFileNames[k]))
		gMonSprites[2] = loadGraphic(gostr(gTmpText))
		for j = 0; j < int(*gMouseY); j++ {
			arenaMonster(2, k)
			*gMonCount++
		}
	}
	refreshView()
}

func updateObjectFacing() {
	if *gObjD0c != 0 {
		for i := 0; i < 10; i++ {
			if arr[Mobj](adrMonHitFxD1c2)[*gObjD0c].ID.Get() == int(gFacingObjIds[i]) {
				*gLastMoveDir = *gPartyDir
				break
			}
		}
	}
}

func freeMazeResources(all int) {
	freeMazeGfx()
	freePartyFaces()
	*gPartyDefeated, *gOutdoors = 0, 0
	freeMazeSprites()
	freeAndNull(gGlobalIcons)
	freeAndNull(gCurIcons)
	freeAndNull(gGargoyleGfx)
	freeAndNull(gGrabberGfx)
	freeAndNull(gBatGfx)
	freeAndNull(gProtGemsGfx)
	freeAndNull(gHpbarsIcons)
	freeAndNull(gCharpowSprites)
	freeAndNull(gRestoreIcons)
	freeAndNull(gProtectIcons)
	freeAndNull(gCurIcons)
	freeAndNull(gMazeScript)
	freeAndNull(gMapText)
	if all != 0 {
		for _, p := range []*Far{gMonMaxHp, gMonAc, gMonSpeed, gMonDmgDice, gMonDmgSides, gMonNumAttacks, gMonExp,
			gMonDmgType, gMonAttackPref, gMonSpecial, gMonRanged, gMonHitBonus, gMonTreasure, gMonGold, gMonGems,
			gMonResMagic, gMonResFire, gMonResElec, gMonResCold, gMonResAcid, gMonResEnergy, gMonResPhys} {
			freeAndNull(p)
		}
	}
}

func initGameFlags() {
	memset(gGameFlags, 0, 0x20)
	for i := 0x10; i < 0x18; i++ {
		setBit(gGameFlags, i, 1)
	}
}

func respawnWaveMonsters(a, b int) {
	lo, hi := 1, 0x1d
	if a != 0 {
		lo, hi = 0x1d, 0x38
	}
	for i := lo; i < hi; i++ {
		if int(gMonSpriteBank[i]) == b {
			var t int
			if a != 0 {
				t = int(gForceWave2Mons[b])
			} else {
				t = int(gForceWave1Mons[b])
			}
			gMonType[i] = int16(t)
			gMonFrames[i] = int16(randomRange(0, int(gMonFrameCounts[t])))
			gMonVisible[i], gMonStatus[i] = 0, 0
			gMonHp[i] = gMonMaxHp.I16(2 * t)
		}
	}
}

// Where game_main_loop's key switch continues (the C's labels).
const (
	actNone   = iota
	actTurn   // turn: after a turn left/right
	actFwd    // fwd: step forward
	actMoved  // moved: after a step
	actTurned // turned: the view changed
)

func gameMainLoop() int {
	var x, y, r, k, lvl, act int
	var p Far

	r = 1
	memset(gQuadrantMaps, 0xff, 4)
	clearPartyAttackFlags()
	gDeathSpr.BotY.Set(0xc6)
	gDeathSpr.TopY.Set(-0xb4)
	loadFaces(int(*gPartyCount))
	enterMaze(int(*gMapNum))
	updateOutdoorQuadrant()
	if *gGameMode == 0 {
		drawBackdrop()
	}
	drawPartyPortraits()
	*gHighlightedChar = -1
	drawMainIcons()
	runEventScript()
	drawWizardEye(1)
	*gGameMode = 1
	refreshView()
	for {
		k = getInput()
		act = actNone
		walls := func(x, y, mask int) int { return getCellWallsRegion(x, y, mask) }
		px, py, dir := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
		switch k {
		case 0xf0:
			switch *gPartyDir {
			case 0:
				*gPartyDir = 3
			case 1:
				*gPartyDir = 2
			case 2:
				*gPartyDir = 0
			case 3:
				*gPartyDir = 1
			}
			act = actTurn
		case 0xf1:
			switch *gPartyDir {
			case 0:
				*gPartyDir = 2
			case 1:
				*gPartyDir = 3
			case 2:
				*gPartyDir = 1
			case 3:
				*gPartyDir = 0
			}
			act = actTurn
		case 0xf2:
			if walls(px, py, int(gDirBlockBits[dir])) == 0 {
				act = actFwd
			} else if walls(px, py, int(gViewWallMasks[dir][2]))>>gViewWallShifts[dir][2] == 2 && *gOutdoors == 0 {
				if k = selectPartyMember(); k == -1 {
					break
				}
				if randomRange(1, 4) == 1 {
					damageParty(int(*mazeB(int(*gMapQuadrant), 0x31e)), randomRange(0, 6), k)
				}
				lock := int(*mazeB(int(*gMapQuadrant), 0x311))
				if getThievery(&gParty[k])+randomRange(1, 0x14) < lock {
					break
				}
				gParty[k].Exp.Set(gParty[k].Exp.Get() + uint32(uint16(getCharLevel(&gParty[k])*lock*10)))
				setCellWall(px, py, dir, 6)
				*gMouseX = int16(px)
				*gMouseY = int16(py)
				switch dir {
				case 0:
					*gMouseY++
					k = 1
				case 1:
					*gMouseY--
					k = 0
				case 2:
					*gMouseX++
					k = 3
				case 3:
					*gMouseX--
					k = 2
				}
				setCellWall(int(*gMouseX), int(*gMouseY), k, 6)
				playSfx(0x2a)
				*gViewDirty = 1
			} else if *gOutdoors != 0 {
				dx7, dy7 := int(int8(gViewDx[dir][7])), int(int8(gViewDy[dir][7]))
				if walls(px+dx7, py+dy7, 0x70) == 0 {
					if *gWalkOnWater != 0 {
						act = actFwd
					} else if walls(px+dx7, py+dy7, 0x7000) != 0 && partyHasSkill(0xe) != 0 {
						act = actFwd
					}
				} else {
					x = int(int8(gViewDx[dir][6])) + px
					y = int(int8(gViewDy[dir][6])) + py
					if k = walls(x, y, 7); k != 0 {
						lvl = int(*gMapQuadrant)
						if x > 0xf {
							lvl++
						} else if x < 0 {
							lvl--
						}
						if y > 0xf {
							lvl += 2
						} else if y < 0 {
							lvl -= 2
						}
						switch *mazeB(lvl, 0x2ff+k) {
						case 1, 6, 8, 11:
							if partyHasSkill(9) != 0 {
								act = actFwd
							}
						case 2, 3, 5, 7, 10:
							if partyHasSkill(0xb) != 0 {
								act = actFwd
							}
						}
					}
				}
				if act != actFwd {
					playSfx(9)
				}
			} else {
				playSfx(9)
			}
		case 0xf3:
			if walls(px, py, int(gDirBlockBits[gDirOpposite[dir]])) == 0 {
				*gLastMoveDir = *gPartyDir
				endPartyTurn()
				switch *gPartyDir {
				case 0:
					*gPartyY--
				case 1:
					*gPartyY++
				case 2:
					*gPartyX--
				case 3:
					*gPartyX++
				}
				act = actMoved
				break
			}
			playSfx(9)
		case 0, 0x53:
			hitWholeParty()
			endPartyTurn()
		case 0x3a:
			if *gKeyCtrl == 0 {
				break
			}
			fallthrough
		case 0xf4:
			if walls(px, py, int(gWallMaskStrafeL[dir])) == 0 {
				*gLastMoveDir = *gPartyDir
				endPartyTurn()
				switch *gPartyDir {
				case 0:
					*gPartyX--
				case 1:
					*gPartyX++
				case 2:
					*gPartyY++
				case 3:
					*gPartyY--
				}
				act = actTurned
				break
			}
			playSfx(9)
		case 1, 0x43:
			if castSpellDialog(0) != 0 {
				endPartyTurn()
			}
		case 2, 0x52:
			restParty()
		case 0x25:
			if *gKeyCtrl == 0 {
				break
			}
			fallthrough
		case 0xf5:
			if walls(px, py, int(gWallMaskStrafeR[dir])) == 0 {
				*gLastMoveDir = *gPartyDir
				endPartyTurn()
				switch *gPartyDir {
				case 0:
					*gPartyX++
				case 1:
					*gPartyX--
				case 2:
					*gPartyY--
				case 3:
					*gPartyY++
				}
				act = actTurned
				break
			}
			playSfx(9)
		case 3, 0x42:
			endPartyTurn()
			if walls(int(*gPartyX), int(*gPartyY), int(gDirBlockBits[*gPartyDir])) == 0 {
				act = actFwd
				break
			}
			bashDoor(int(*gPartyX), int(*gPartyY), int(*gPartyDir))
			*gViewDirty = 1
		case 4, 0x44:
			endPartyTurn()
			dismissPartyMember()
		case 5, 0x56:
			showCoraksNotes()
		case 6, 0x4d:
			showAutomap()
		case 7, 0x49:
			showGameInfo()
		case 0x20:
			endPartyTurn()
			*gViewDirty = 1
		case 8, 0x51:
			showQuickRef()
		case 9:
			if controlPanel() != 0 {
				r = 2
				goto out
			}
		case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
			if k -= 0xc9; k < int(*gPartyCount) {
				charScreen(k)
				if *gViewDirty != 0 {
					moveMonsters()
				}
				setupMainUi()
			}
		}
		switch act {
		case actTurn:
			updateObjectFacing()
			if *gMazeGfxIndex == 0 || *gOutdoors != 0 || *gMazeGfxIndex == 3 {
				*gViewBgFlip ^= 1
			}
			if *gMonsterScanDelay != 0 {
				moveMonsters()
			}
			act = actTurned
		case actFwd:
			act = mainLoopForward()
		}
		if act == actMoved {
			if *gMazeGfxIndex != 0 && *gOutdoors == 0 && *gMazeGfxIndex != 3 {
				*gViewBgFlip = *gViewFlip ^ 1
			}
			act = actTurned
		}
		if act == actTurned {
			*gViewDirty = 1
			memset(gViewCaption, 0, 0x96)
			*gViewFlip ^= 1
			if *gOutdoors != 0 && getCellWallsRegion(int(*gPartyX), int(*gPartyY), 0x70) == 0 {
				playSfx(0x1d)
			} else {
				gStepSoundToggle[0] ^= 1
				playSfx(int(gStepSoundToggle[0]) + 7)
			}
		}
		if *gViewDirty != 0 {
			switch r = runEventScript(); r {
			case 1, 2:
				goto out
			}
		}
		recoverMonsterStatus()
		if *gPartyDefeated != 0 {
			break
		}
	}
out:
	clearScreen()
	*gHighlightedChar = -2
	freeMazeResources(0)
	if r == 2 {
		setMouseCursor(0, uint32(*gMouseCursor))
		closeWindows(10)
	} else {
		if r != 1 {
			p = loadGraphic("death.vga")
			gDeathSpr.Img.Set(uint32(p))
			if *gMusicEnabled != 0 {
				sndPlayMusic(0)
			}
			flushInput()
			k = 0
			for {
				drawSprites(ds[adrDeathSpr:])
				gDeathSpr.TopY.Set(gDeathSpr.TopY.Get() + 8)
				gDeathSpr.BotY.Set(gDeathSpr.BotY.Get() - 4)
				if k++; k == 1 {
					playSfx(0x22)
				}
				if *gSndDev0 != 0 && k%9 == 0 {
					playSfx(0x23)
				}
				if !(gDeathSpr.TopY.Get() < 0 && bioskey(1) == 0 && *gUserBreak == 0) {
					break
				}
			}
			playSfx(0x21)
			gDeathSpr.TopY.Set(0)
			gDeathSpr.BotY.Set(0x6c)
			drawSprites(ds[adrDeathSpr:])
			setAnykeyButtons()
			idleFramesOrKey(0x1e)
			freeAndNull(&p)
			fadeScreen(0, 4)
		}
		closeWindows(10)
		flushInput()
		playMusic(cs("mm3theme.m"), 0xf0)
	}
	return r
}

// mainLoopForward is game_main_loop's fwd: the force zones of map 0x16, then the step.
// It returns where the C continues: moved (also when a Force repels the party).
func mainLoopForward() int {
	if *gMaps0BaseMap == 0x16 {
		if *gPartyY < 8 && *gForceZone1Entered == 0 {
			for k := 0x1d; k < 0x38; k++ {
				if (gMonY[k]|gMonX[k])&0xff != 0x80 {
					*gPartyX = 0x10
					*gPartyY = 0x10
					showMessage(cs("A Force repells you!"))
					return actMoved
				}
			}
			*gForceZone1Entered = 1
			*gForceZone2Entered = 1
			showPleaseWait()
			for k := 0; k < 3; k++ {
				freeAndNull(&gMonSprites[k])
				sprintf(gTmpText, "%s.mon", str(gMonFileNames[gForceWave1Mons[k]]))
				gMonSprites[k] = loadGraphic(gostr(gTmpText))
				respawnWaveMonsters(0, k)
			}
			closeWindows(1)
		} else if *gPartyY < 0x12 && *gForceZone2Entered == 0 {
			for k := 0x38; k < 0x92; k++ {
				if (gMonY[k]|gMonX[k])&0xff != 0x80 {
					*gPartyX = 0xf
					*gPartyY = 0x19
					showMessage(cs("A Force repells you!"))
					return actMoved
				}
			}
			showPleaseWait()
			for k := 0; k < 3; k++ {
				freeAndNull(&gMonSprites[k])
				sprintf(gTmpText, "%s.mon", str(gMonFileNames[gForceWave2Mons[k]]))
				gMonSprites[k] = loadGraphic(gostr(gTmpText))
				respawnWaveMonsters(1, k)
			}
			closeWindows(1)
			*gForceZone2Entered = 1
		}
	}
	*gLastMoveDir = *gPartyDir
	switch *gPartyDir {
	case 0:
		*gPartyY++
	case 1:
		*gPartyY--
	case 2:
		*gPartyX++
	case 3:
		*gPartyX--
	}
	*gViewFlip ^= 1
	endPartyTurn()
	*gViewFlip ^= 1
	return actMoved
}
