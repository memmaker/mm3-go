// Port of BUILDING.C (overlay 1831): the town buildings (bank, guild, inn, tavern, temple,
// training grounds, smithy), their entry dialog and input loop, and the face yes/no dialog.
//
// building.c's struct chr has guild[7] at 0x38 (guild[town] = bytesOf(c)[0x38+town]), award6
// at 0x3f (Award[6]) and f106[2] at 0x106 (Donated). The animation sequence
// (g_bldg_anim_seq) can be a caller's local (the smithy's, face_yes_no_dialog's, the town's),
// so it is kept as a Go slice, bldgAnimSeq; the near pointer global is not used.
// goto redraw is a labelled outer loop; tavern's tip: (entered from the rumors case) a closure.
package mm3

var bldgAnimSeq []byte

func bldgSprites() []Spr { return arr[Spr](adrBldgFaceSprites) }

func formatAmountK(v uint32, buf []byte) []byte {
	if v > 9999 {
		sprintf(buf, "%uk", v/1000)
	} else {
		sprintf(buf, "%u", v)
	}
	return buf
}

func animateBuildingFace() {
	if !frameDue() { // in the input loop: skip, don't wait, so clicks stay immediate
		return
	}
	s := bldgSprites()
	pos := int(*gBldgAnimPos)
	seq := int(int8(bldgAnimSeq[pos]))
	if *gBldgType == 5 {
		if pos == 0 {
			playSfx(0x18)
		} else if pos == 15 {
			playSfx(0x18)
		}
	}
	if *gBldgType == 0 {
		s[0].Img.Set(uint32(gBldgFaceImgs[seq%4]))
		s[0].Frame.Set(seq >> 2)
	} else if *gBldgType == 5 {
		s[0].Img.Set(uint32(gBldgFaceImgs[pos&1]))
		s[0].Frame.Set(seq >> 1)
	} else if *gBldgType < 7 {
		s[0].Frame.Set(seq)
	} else {
		s[1].Frame.Set(seq)
	}
	*gBldgAnimPos = int16((pos + 1) % int(*gBldgAnimLen))
	addUiGadgetSprites()
	printWindowText(gBldgDrawStr)
	if *gBldgType == 5 {
		idleFrames(3)
	}
}

func buildingWaitInput() int {
	var k, i, f int
	for {
		if *gMusicLoadPending != 0 && sndPlayMusic(-1) == 0 {
			loadMusic()
		}
		animateBuildingFace()
		if getMouseState(gMouseX, gMouseY) != 0 {
			*gMouseButtons = 1
			k = hitTestButton()
			if *gBtnXs == adrBldgBtnXs && k > 1 {
				return k - 1
			}
			if *gBtnXs == adrAnyclickBtnXs && *gGameMode != 3 {
				return -1
			}
			if k != 0 {
				k--
				i = k
				drawButtonState(i, i*2+1)
				f = 1
				for {
					k = hitTestButton()
					if k != 0 && k-1 == i {
						k--
						if f == 0 {
							f = 1
							drawButtonState(i, i*2+1)
						}
					} else {
						if k != 0 {
							k--
						}
						if f != 0 {
							f = 0
							drawButtonState(i, i*2)
						}
					}
					animateBuildingFace()
					if getMouseState(gMouseX, gMouseY) == 0 {
						break
					}
				}
				drawButtonState(i, i*2)
				if f != 0 {
					if *gBtnXs == adrPartyBtnX {
						return i - 0x59
					}
					return i
				}
			} else {
				k = hitTestPortrait()
				if k != 0 {
					k--
					return k + 0xc9
				}
			}
		}
		if bioskey(1) != 0 || *gUserBreak != 0 {
			*gMouseButtons = 0
			if k = readKey(); k == 0xd2 {
				bossKey()
			} else {
				return k
			}
		}
	}
}

func buildingEntryDialog(title, seq []byte, n, typ int) int {
	var r int
	s := bldgSprites()
	drawWizardEye(1)
	showPleaseWait()
	memset(ds[adrBldgFaceImgs:], 0, 0x10)
	bldgAnimSeq = seq
	*gBldgAnimLen = int16(n)
	*gBldgType = int16(typ)
	*gBldgAnimPos = 0
	if typ != 7 {
		unloadMaze()
	}
	sprintf(gBldgDrawStr, "%s.out", str(gBldgOutNames[typ]))
	gBldgFaceImgs[0] = loadGraphic(gostr(gBldgDrawStr))
	if typ == 0 {
		gBldgFaceImgs[1] = loadGraphic("train2.out")
		gBldgFaceImgs[2] = loadGraphic("train3.out")
		gBldgFaceImgs[3] = loadGraphic("train4.out")
	} else if typ == 5 {
		gBldgFaceImgs[1] = loadGraphic("guild2.out")
	} else if typ > 6 {
		gBldgFaceImgs[1] = loadGraphic("frame.out")
	}
	if typ < 7 {
		sprintf(gBldgDrawStr, "\x03d\x05%04X\x05%04X", adrBldgFaceSprites, adrBorderSprites)
		s[1].Y.Set(8)
		s[0].Y.Set(8)
		s[0].X.Set(8)
	} else {
		sprintf(gBldgDrawStr, "\x05%04X", adrBldgFaceSprites)
		s[0].Y.Set(1)
		s[0].X.Set(1)
		s[1].Y.Set(7)
	}
	addUiGadgetSprites()
	*gConfirm2SpritesImg = *gGlobalIcons
	if typ < 7 {
		s[1].Img.Set(uint32(*gGlobalIcons))
		s[0].Img.Set(uint32(gBldgFaceImgs[0]))
	} else {
		s[1].Img.Set(uint32(gBldgFaceImgs[0]))
		s[0].Img.Set(uint32(gBldgFaceImgs[1]))
	}
	s[0].Frame.Set(0)
	s[1].Frame.Set(1)
	closeWindows(1)
	if typ < 7 {
		printWindowText(gBldgDrawStr)
	} else {
		sprintf(gTextBuf2, "\x03c%s\x05%04X", title, adrBldgFaceSprites)
		openTextWindow(8, 8, 0xd8, 0x84, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
	}
	if *gBldgType == 0 {
		playSfx(0x16)
	}
	setMouseCursor(0, uint32(*gMouseCursor))
	*gCurIcons = loadGraphic("confirm2.icn")
	*gConfirmSpritesImg2 = *gCurIcons
	sprintf(gTextBuf2, "\x03b\x05%04X", adrConfirm2Sprites)
	printWindowText(gTextBuf2)
	if typ < 7 {
		openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, farPtr(&title[0]))
	}
	pushButtons()
	*gBtnXs = adrConfirmBtnXs
	*gBtnYs = adrConfirmBtnYs
	*gBtnWs = adrConfirmBtnWs
	*gBtnHs = adrConfirmBtnHs
	setMouseCursor(0, uint32(*gMouseCursor))
	save := *gPortraitClickOn
	*gPortraitClickOn = 0
	waitMouseRelease()
loop:
	for {
		switch buildingWaitInput() {
		case 0, 'Y':
			r = 1
			break loop
		case 1, 0x1b, 'N':
			r = 0
			break loop
		}
	}
	if typ < 7 {
		closeWindows(1)
	}
	if r != 0 && typ < 7 {
		*gHighlightedChar = -1
		highlightPortrait(0)
	}
	popButtons()
	freeAndNull(gCurIcons)
	*gPortraitClickOn = save
	loadMainIcons()
	if typ < 7 {
		drawMainIcons()
	}
	return r
}

func bankTransferDialog(dep int) {
	var k int
	var amt, a, b int32
	var s1, s2 [10]byte

	*gBankBtn2Y = 0x4a
	*gBankBtn1Y = 0x4a
	*gBankBtn0Y = 0x4a
	balances := func() {
		if dep != 0 {
			a = int32(*gBankGold)
			b = int32(*gBankGems)
		} else {
			a = int32(*gPartyGold)
			b = int32(*gPartyGems)
		}
	}
	balances()
	sprintf(gTextBuf, "\x03c%s\x03l\n\nGold\x03r\t093%s\x03l\nGems\x03r\t093%s\x02\x03c\x0b096\t013G\x0c27o\x0cdld\t040G\x0c27e\x0cdms\t067ESC\x01\x05%04X",
		str(gBankOpNames[dep]), formatBigNumber(uint32(a), s1[:]), formatBigNumber(uint32(b), s2[:]), adrBankSprites)
	openTextWindow(0xe2, 0x1a, 0x5d, 0x78, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	amount := func() {
		if dep != 0 && *gBankGems == 0 && k != 0 || dep != 0 && *gBankGold == 0 && k == 0 || dep == 0 && *gPartyGems == 0 && k != 0 || dep == 0 && *gPartyGold == 0 && k == 0 {
			showNotEnough(k, dep, 1)
			return
		}
		sprintf(gTextBuf, "\x03c\t000\x0b051Amount\x03l\n")
		printWindowText(gTextBuf)
		amt = inputNumber(10, 0x4d)
		if amt != 0 {
			if k != 0 {
				if payCost(1, uint32(amt), dep) != 0 {
					if dep != 0 {
						*gPartyGems += uint32(amt)
					} else {
						*gBankGems += uint32(amt)
					}
				}
			} else {
				if payCost(0, uint32(amt), dep) != 0 {
					if dep != 0 {
						*gPartyGold += uint32(amt)
					} else {
						*gBankGold += uint32(amt)
					}
				}
			}
		}
		balances()
		sprintf(gTextBuf, "\t000\x0b000\x03c%s\x03l\n\n\x04077Gold\x03r\t093%s\x03l\n\x04077Gems\x03r\t093%s\x03l\t000\x0b051\x04077\n\x04077",
			str(gBankOpNames[dep]), formatBigNumber(uint32(a), s1[:]), formatBigNumber(uint32(b), s2[:]))
		printWindowText(gTextBuf)
	}
	for {
		k = buildingWaitInput()
		switch k {
		case 'O':
			k = 0
			amount()
		case 'E':
			k = 1
			amount()
		case 0, 1:
			amount()
		case 0x1b:
			k = 2
		}
		if k == 2 {
			break
		}
	}
	closeWindows(1)
	*gBankBtn2Y = 100
	*gBankBtn1Y = 100
	*gBankBtn0Y = 100
}

func visitBank() {
	var k int
	var s1, s2, s3, s4 [10]byte

	if *gMinutes < 0x21c || *gMinutes > 0x3fc {
		showMessage(cs("\x0b007Sorry, the Bank's closed! Come back between 9am and 5pm."))
		*gViewFlip ^= 1
		setBuildingExitPos(6)
		return
	}
	playMusic(cs("bank.m"), 0xc0)
	playSfx(0x97)
	strncpy(gTextBuf, gMapText.Add(int(*gBankTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimBank, 10, 6) != 0 {
		pushButtons()
		pushIcons("bank.icn")
		p := loadGraphic("bank2.icn")
		save := *gCurIcons
		*gBankSpritesImg = *gCurIcons
		*gBtnXs = adrBankBtnXs
		*gBtnYs = adrBankBtnYs
		*gBtnWs = adrBankBtnWs
		*gBtnHs = adrBankBtnHs
		sprintf(gTextBuf, "\x05%04X\x02\x03c\x0b122\t013\x0c27D\x0cdep\t040\x0c27W\x0cdith\t067ESC\x01\t000\x0b000Bank of Cancomar\x0b015\nBank\x03l\nGold\x03r\t093%s\x03l\nGems\x03r\t093%s\x03c\n\nParty\x03l\nGold\x03r\t093%s\x03l\nGems\x03r\t093%s",
			adrBankSprites, formatBigNumber(*gBankGold, s1[:]), formatBigNumber(*gBankGems, s2[:]), formatBigNumber(*gPartyGold, s3[:]), formatBigNumber(*gPartyGems, s4[:]))
		openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		trans := func() {
			*gBankSpritesImg = p
			*gCurIcons = p
			bankTransferDialog(k)
			*gCurIcons = save
			*gBankSpritesImg = *gCurIcons
			sprintf(gTextBuf, "\x01\x03c\x0b015\nBank\x03l\n\x04077Gold\x03r\t093%s\x03l\n\x04077Gems\x03r\t093%s\x03c\n\nParty\x03l\n\x04077Gold\x03r\t093%s\x03l\n\x04077Gems\x03r\t093%s",
				formatBigNumber(*gBankGold, s1[:]), formatBigNumber(*gBankGems, s2[:]), formatBigNumber(*gPartyGold, s3[:]), formatBigNumber(*gPartyGems, s4[:]))
			printWindowText(gTextBuf)
		}
		for {
			k = buildingWaitInput()
			switch k {
			case 'D':
				k = 0
				trans()
			case 'W':
				k = 1
				trans()
			case 0, 1:
				trans()
			case 0x1b:
				k = 2
			}
			if k == 2 {
				break
			}
		}
		closeWindows(1)
		popIcons()
		popButtons()
		advanceClock(0x3c)
	}
	exitBuilding(0, 6)
}

func chrGuild(c *Chr) uint8 { return bytesOf(c)[0x38+int(*gCurTown)] }

func formatGuildOptions(c *Chr) {
	var s [20]byte
	if chrGuild(c) != 0 {
		sprintf(gTextBuf2, "\x03c\x0b027\t039%s\x03l\x0b046\t012\x0c27G\x0cduild Info\n\t012\x0c27B\x0cduy Spells\n\t012\x0c27S\x0cdpell Info", bytesOf(c))
	} else {
		sprintf(gTextBuf2, "\n\nYou have to be a member to shop here.")
	}
	sprintf(gTextBuf, "\r\x05%04X\x03c\x0b000\t000Guild Options%s\x03l\t000\x0b090Gold\x03r\t077%s\x02\x03c\x0b122\t040ESC\x01",
		adrEscSprites, gTextBuf2, formatBigNumber(*gPartyGold, s[:]))
}

func visitGuild() {
	var k int
	if *gIsNight == 0 {
		showMessage(cs("\x0b007Sorry, the Guild's closed! Come back during night hours."))
		setBuildingExitPos(5)
		*gViewFlip ^= 1
		return
	}
	playMusic(cs("guild.m"), 0xc0)
	playSfx(0x98)
	strncpy(gTextBuf, gMapText.Add(int(*gGuildTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimGuild, 0x19, 5) != 0 {
		loadTextTable(cs("spldesc.bin"))
		drawIcon(*gGlobalIcons, 0, 0xe8, 0x4d, 0)
		pushButtons()
		pushIcons("esc.icn")
		*gEscSpritesImg = *gCurIcons
		*gBtnXs = adrBldgBtnXs
		*gBtnYs = adrBldgBtnYs
		*gBtnWs = adrBldgBtnWs
		*gBtnHs = adrBldgBtnHs
		c := &gParty[0]
	redraw:
		for {
			formatGuildOptions(c)
			openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
			for {
				k = buildingWaitInput()
				switch k {
				case 0:
					k = 0x1b
				case 1, 'G':
					if chrGuild(c) != 0 {
						showGuildMessage()
					}
				case 2, 'B':
					if charCannotAct(c) == 0 && chrGuild(c) != 0 {
						closeWindows(1)
						c = spellListDialog(c, 0)
						continue redraw
					}
				case 3, 'S':
					c = spellListDialog(c, 0x80)
				case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
					k -= 0xc9
					if int(*gPartyCount) > k {
						c = &gParty[k]
						highlightPortrait(k)
						formatGuildOptions(c)
						printWindowText(gTextBuf)
					}
				}
				if k == 0x1b {
					break redraw
				}
			}
		}
		freeAndNull(gTextTable)
		popIcons()
		popButtons()
		closeWindows(1)
		drawMainIcons()
		advanceClock(0x3c)
	}
	exitBuilding(0, 5)
}

func visitInn() int {
	i := 0
	playMusic(cs("towninn.m"), 0xc0)
	playSfx(0x99)
	strncpy(gTextBuf, gMapText.Add(int(*gInnTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimInn, 0xd, 4) != 0 {
		*gPartyX = gInnPositions[*gMapNum][0]
		*gPartyY = gInnPositions[*gMapNum][1]
		*gPartyDir ^= 1
		*gInnMap = *gMapNum
		for i = 0; i < int(*gPartyCount); i++ {
			gParty[i].Inn.Set(int(*gMapNum))
		}
		*gMapNum = 0
		advanceClock(0x5a0)
		saveGame(1)
		i = 1
	}
	exitBuilding(i, 4)
	return i
}

func formatTavernOptions(c *Chr) {
	var s [20]byte
	sprintf(gTextBuf, fmtP(str(*gFmtTavernOptions)), adrEscSprites, bytesOf(c), "\x03l\t017\x0b046\x0c27D\x0cdrink\n\t017\x0c27F\x0cdood\n\t017\x0c27T\x0cdip\n\t017\x0c27R\x0cdumors",
		formatBigNumber(*gPartyGold, s[:]))
}

func visitTavern() {
	var k, drunk, shown, mode int
	var cost uint16

	if *gMinutes < 0x438 && *gMinutes > 0x12c {
		showMessage(cs("\x0b007Sorry, the Tavern's closed! Come back between 6pm and 5am."))
		setBuildingExitPos(3)
		*gViewFlip ^= 1
		return
	}
	playMusic(cs("honky.m"), 0xc0)
	playSfx(0x9a)
	strncpy(gTextBuf, gMapText.Add(int(*gTavernTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimTavern, 0xd, 3) != 0 {
		mode = int(*gGameMode)
		*gGameMode = 0xb
		pushButtons()
		pushIcons("esc.icn")
		loadTextTable(cs("tavern.bin"))
		*gEscSpritesImg = *gCurIcons
		*gBtnXs = adrBldgBtnXs
		*gBtnYs = adrBldgBtnYs
		*gBtnWs = adrBldgBtnWs
		*gBtnHs = adrBldgBtnHs
		drunk = 0
		shown = 0
		advanceClock(0x78)
		c := &gParty[0]
		tip := func() {
			sprintf(gTextBuf, "\x03c\x0b012%s", gTmpText)
			openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
			waitMouseRelease()
			buildingWaitInput()
			closeWindows(1)
		}
		hitKey := func(msg string, s []byte) {
			sprintf(gTextBuf, fmtP(str(*gFmtTavernOptions)), adrEscSprites, bytesOf(c), msg, formatBigNumber(*gPartyGold, s))
			printWindowText(gTextBuf)
		}
	redraw:
		for {
			formatTavernOptions(c)
			if shown != 0 {
				printWindowText(gTextBuf)
			} else {
				openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
				shown = 1
			}
			for {
				var s [20]byte
				k = buildingWaitInput()
				switch k {
				case 0:
					k = 0x1b
				case 1, 'D':
					if charCannotAct(c) != 0 {
						break
					}
					if payCost(0, 1, 0) == 0 {
						continue redraw
					}
					drunk = 1
					hitKey("\n\nGood Stuff\n\nHit a key!", s[:])
					if randomRange(1, 100) < 0x1a {
						c.Cond[7]++
						drawPartyPortraits()
					}
					waitMouseRelease()
					buildingWaitInput()
					continue redraw
				case 2, 'F':
					strncpy(gTextBuf2, gTextTable.Add(int(gTavernFoodTextOffs[*gMapNum])).B(), 300)
					closeWindows(1)
					shown = 0
					openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
					if confirmYesNo(0) != 0 {
						cost = uint16(int(gFoodDaysPerTown[*gMapNum-1]) * int(*gPartyCount) * 3)
						if *gFood < cost {
							if payCost(0, uint32(gFoodPricePerTown[*gMapNum-1]), 0) != 0 {
								*gFood = cost
							}
						} else {
							showMessage(cs("\x0b007Your food packs are already full!"))
						}
					}
					closeWindows(1)
					continue redraw
				case 3, 'T':
					if charCannotAct(c) != 0 {
						break
					}
					if drunk != 0 {
						drunk = 0
						if c.Cond[7] != 0 {
							hitKey("\n\nYou're Drunk\n\nHit a key!", s[:])
							waitMouseRelease()
							buildingWaitInput()
							continue redraw
						}
						if payCost(0, 1, 0) == 0 {
							continue redraw
						}
						m := &gMaps[*gMapQuadrant]
						base := gTavernTipBase[*gMapNum-1]
						strncpy(gTmpText, gTextTable.Add(int(gTextOffsets[int(m.TavernTip)+int(base)])).B(), 0xc7)
						m.TavernTip++
						if int(m.TavernTip)+int(base) > int(gTavernTipBase[*gMapNum])-1 {
							m.TavernTip--
						}
						tip()
						continue redraw
					}
					hitKey("\n\nHave a Drink\n\nHit a key!", s[:])
					waitMouseRelease()
					buildingWaitInput()
					continue redraw
				case 4, 'R':
					strncpy(gTmpText, gTextTable.Add(int(gTextOffsets[(int(*gMapNum)-1)*10+int(*gDay)%10])).B(), 0xc7)
					tip()
					continue redraw
				case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
					k -= 0xc9
					if int(*gPartyCount) > k {
						c = &gParty[k]
						drunk = 0
						highlightPortrait(k)
						formatTavernOptions(c)
						printWindowText(gTextBuf)
					}
				}
				if k == 0x1b {
					break redraw
				}
			}
		}
		*gGameMode = uint8(mode)
		freeAndNull(gTextTable)
		popIcons()
		popButtons()
		closeWindows(1)
		advanceClock(0x3c)
		saveLoadedMazes()
	}
	exitBuilding(0, 3)
}

func formatTempleOptions(c *Chr, heal, don int32, unc uint32) {
	var s1, s2 [20]byte
	sprintf(gTextBuf, "\r\x05%04X\x03c\x0b000\t000Temple Options\t039\x0b027%s\x03l\t000\x0b046\x0c27H\x0cdeal\x03r\t077%lu\x03l\n\x0c27D\x0cdonation\x03r\t077%lu\x03l\n\x0c27U\x0cdnCurse\x03r\t077%s\x03l\t000\x0b090Gold\x03r\t077%s\x02\x03c\x0b122\t040ESC\x01",
		adrEscSprites, bytesOf(c), uint32(heal), uint32(don), formatAmountK(unc, s2[:]), formatBigNumber(*gPartyGold, s1[:]))
}

func visitTemple() {
	var k, tmp, shown, cnt int
	var heal, x, y, unc, don, lvl, t int32

	playMusic(cs("temples.m"), 0xc0)
	playSfx(0x9b)
	t, shown, cnt = 0, 0, 0
	strncpy(gTextBuf, gMapText.Add(int(*gTempleTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimTemple, 0xb, 2) != 0 {
		pushButtons()
		pushIcons("esc.icn")
		*gEscSpritesImg = *gCurIcons
		*gBtnXs = adrBldgBtnXs
		*gBtnYs = adrBldgBtnYs
		*gBtnWs = adrBldgBtnWs
		*gBtnHs = adrBldgBtnHs
		c := &gParty[0]
		town := int(*gCurTown) - 1
	redraw:
		for {
			heal, x, y, unc, don = 0, 0, 0, 0, 0
			lvl = int32(getCharLevel(c))
			if uint16(calcMaxHp(c)) > uint16(c.HP.Get()) {
				heal = lvl*10 + int32(gTempleHealCost[town])
			}
			for k = 1; k <= 12; k++ {
				if c.Cond[k] != 0 {
					heal += lvl * 10
				}
			}
			x = 0
			if c.Cond[13] != 0 {
				x += lvl*100 + int32(int(c.Cond[13])*50) + int32(gTempleRaiseCost[town])
			}
			if c.Cond[14] != 0 {
				x += lvl*100 + int32(int(c.Cond[14])*50) + int32(gTempleRaiseCost[town])
			}
			if c.Cond[15] != 0 {
				y = lvl*1000 + int32(int16(int(c.Cond[15])*500)) + int32(gTempleEradCost[town])
			}
			for k = 0; k < 0x12; k++ {
				unc |= int32(c.ItemFlags[k] & 0x40)
			}
			if unc != 0 || c.Cond[0] != 0 {
				unc = lvl*20 + int32(gTempleUncurseCost[town])
			}
			don = int32(gTempleDonation[town])
			heal = heal + x + y
			formatTempleOptions(c, heal, don, uint32(unc))
			if shown == 0 {
				openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
				shown = 1
			} else {
				printWindowText(gTextBuf)
			}
			for {
				k = buildingWaitInput()
				switch k {
				case 0:
					k = 0x1b
				case 1, 'H':
					if heal == 0 || payCost(0, uint32(heal), 0) == 0 {
						break
					}
					c.MightTmp, c.IntelTmp, c.PersTmp, c.EndurTmp, c.SpeedTmp, c.AccTmp, c.LuckTmp, c.LevelTmp, c.AcTmp = 0, 0, 0, 0, 0, 0, 0, 0, 0
					c.FireResTmp, c.ColdResTmp, c.ElecResTmp, c.PoisonResTmp, c.EnergyResTmp, c.MagicResTmp = 0, 0, 0, 0, 0, 0
					c.HP.Set(calcMaxHp(c))
					tmp = int(c.Cond[0])
					c.Cond = [0x10]uint8{}
					c.Cond[0] = uint8(tmp)
					t += 0x5a0
					drawPartyPortraits()
					continue redraw
				case 2, 'D':
					if don == 0 || payCost(0, uint32(don), 0) == 0 {
						break
					}
					for k = 0; k < int(*gPartyCount); k++ {
						p := bytesOf(&gParty[k])
						setBit(p[0x106:], town, 1)
						if p[0x106]&0xf8 == 0xf8 {
							gParty[k].Award[6] = 1
						}
					}
					cnt = (cnt + 1) % 10
					if int(*gDay)%10 == cnt {
						v := uint8(cnt)
						if cnt == 0 {
							v = 10
						}
						c.Bless, c.Shield, c.Holy, c.Hero = v, v, v, v
						drawPartyPortraits()
					}
					waitMouseRelease()
					continue redraw
				case 3, 'U':
					if unc == 0 || payCost(0, uint32(unc), 0) == 0 {
						break
					}
					c.Cond[0] = 0
					for k = 0; k < 0x12; k++ {
						if c.ItemFlags[k]&0x40 != 0 {
							c.ItemFlags[k] ^= 0x40
						}
					}
					t += 0x5a0
					drawPartyPortraits()
					continue redraw
				case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
					k -= 0xc9
					if int(*gPartyCount) > k {
						c = &gParty[k]
						highlightPortrait(k)
						cnt = 0
						continue redraw
					}
				}
				if k == 0x1b {
					break redraw
				}
			}
		}
		popIcons()
		popButtons()
		closeWindows(1)
		t += 0x5a0
	}
	advanceClock(int(int16(t)))
	drawPartyPortraits()
	exitBuilding(0, 2)
}

func formatTrainingStatus(c *Chr) int32 {
	need := getExpToNextLevel(c)
	max := gTrainMaxLevel[*gMapNum-1]
	if need != 0 && c.Level < max {
		sprintf(gTextBuf2, "%s needs %lu experience for level %u.", bytesOf(c), uint32(need), int(c.Level)+1)
	} else if c.Level >= max {
		sprintf(gTextBuf2, "%s has learned all we can teach!", bytesOf(c))
		need = 1
	} else {
		sprintf(gTextBuf2, "%s is eligible for level %d.\x03l\n\x0b081Cost\x03r\t077%lu", bytesOf(c), int(c.Level)+1,
			uint32(c.Level)*uint32(c.Level)*10)
	}
	return need
}

func visitTraining() {
	var k, shown, idx int
	var done [8]byte
	var need int32

	if *gIsNight != 0 {
		showMessage(cs("\x0b007Sorry, the Training grounds are closed! Come back during daylight hours."))
		setBuildingExitPos(0)
		*gViewFlip ^= 1
		return
	}
	shown = 0
	freePartyFaces()
	freeAndNull(gCharpowSprites)
	freeAndNull(gProtectIcons)
	freeAndNull(gRestoreIcons)
	freeAndNull(gHpbarsIcons)
	playMusic(cs("grounds.m"), 0xc0)
	playSfx(0x9c)
	strncpy(gTextBuf, gMapText.Add(int(gMazeTextOffs[0])).B(), 300)
	if buildingEntryDialog(gTextBuf, gAnimTraining, 0x10, 0) != 0 {
		updateFreeMemKb()
		pushButtons()
		freeAndNull(gCurIcons)
		*gCurIcons = loadGraphic("train.icn")
		*gTrainSpritesImg = *gCurIcons
		*gBtnXs = adrTrainBtnXs
		*gBtnYs = adrTrainBtnYs
		*gBtnWs = adrTrainBtnWs
		*gBtnHs = adrTrainBtnHs
		idx = 0
		c := &gParty[idx]
	redraw:
		for {
			need = formatTrainingStatus(c)
			{
				var s [40]byte
				sprintf(gTextBuf, "\r\x05%04X\x03cTraining Options\n\n%s\x03l\x0b090\t000Gold\x03r\t077%s\x02\x03c\x0b122\t021\x0c27T\x0cdrain\t060ESC\x01",
					adrTrainSprites, gTextBuf2, formatBigNumber(*gPartyGold, s[:]))
			}
			if shown == 0 {
				openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
				shown = 1
			} else {
				printWindowText(gTextBuf)
			}
			for {
				k = buildingWaitInput()
				switch k {
				case 0:
					k = 0x1b
				case 1, 'T':
					if need != 0 || charCannotAct(c) != 0 {
						break
					}
					if payCost(0, uint32(c.Level)*uint32(c.Level)*10, 0) == 0 {
						break
					}
					c.Exp.Set(c.Exp.Get() - uint32(getNextLevelExp(c)-(getTotalExp(c)-int32(c.Exp.Get()))))
					c.Level++
					if done[idx] == 0 {
						advanceClock(0x5a0)
						done[idx] = 1
					}
					clearTempEffects(1)
					c.HP.Set(calcMaxHp(c))
					c.SP.Set(calcMaxSp(c))
					continue redraw
				case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
					k -= 0xc9
					if int(*gPartyCount) > k {
						idx = k
						c = &gParty[idx]
						highlightPortrait(idx)
						continue redraw
					}
				}
				if k == 0x1b {
					break redraw
				}
			}
		}
		closeWindows(1)
		freeAndNull(gCurIcons)
		popButtons()
	}
	for k = 0; k < 4; k++ {
		freeAndNull(&gBldgFaceImgs[k])
	}
	showPleaseWait()
	*gHpbarsIcons = loadGraphic("hpbars.icn")
	*gRestoreIcons = loadGraphic("restore.icn")
	*gProtectIcons = loadGraphic("protect.icn")
	loadFaces(int(*gPartyCount))
	*gCharpowSprites = loadGraphic("charpow.icn")
	exitBuilding(0, 0)
	clearCharHighlight()
	drawPartyPortraits()
}

func formatSmithyOptions(c *Chr) {
	var s [20]byte
	sprintf(gTextBuf, "\r\x05%04X\x03c\x0b000\t000Store Options\t039\x0b027%s\x03l\x0b046\n\t011\x0c27B\x0cduy Items\n\t011\x0c27I\x0cd\t015nventory\t000\x0b090Gold\x03r\t077%s\x02\x03c\x0b122\t040ESC\x01",
		adrEscSprites, bytesOf(c), formatBigNumber(*gPartyGold, s[:]))
}

func visitSmithy() {
	var k int
	if *gIsNight != 0 {
		showMessage(cs("\x0b007Sorry, the Smithy's closed! Come back during daylight hours."))
		setBuildingExitPos(1)
		*gViewFlip ^= 1
		return
	}
	var seq [12]byte
	playMusic(cs("dungeon.m"), 0xc0)
	playSfx(0x9d)
	for k = 0; k < 12; k++ {
		seq[k] = byte(k)
	}
	strncpy(gTextBuf, gMapText.Add(int(*gSmithyTextOffs)).B(), 300)
	if buildingEntryDialog(gTextBuf, seq[:], 12, 1) != 0 {
		pushButtons()
		pushIcons("esc.icn")
		*gEscSpritesImg = *gCurIcons
		*gBtnXs = adrBldgBtnXs
		*gBtnYs = adrBldgBtnYs
		*gBtnWs = adrBldgBtnWs
		*gBtnHs = adrBldgBtnHs
		c := &gParty[0]
		formatSmithyOptions(c)
		openTextWindow(0xe2, 0, 0x5d, 0x92, 0, 0x96, 4, 6, dsFar(adrTextBuf))
		for k != 0x1b {
			k = buildingWaitInput()
			switch k {
			case 0:
				k = 0x1b
			case 2, 'B':
				c = itemScreen(c, 1)
				formatSmithyOptions(c)
				printWindowText(gTextBuf)
			case 3, 'I':
				c = itemScreen(c, 2)
				formatSmithyOptions(c)
				printWindowText(gTextBuf)
			case 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0:
				k -= 0xc9
				if int(*gPartyCount) > k {
					c = &gParty[k]
					highlightPortrait(k)
					formatSmithyOptions(c)
					printWindowText(gTextBuf)
				}
			}
		}
		popIcons()
		popButtons()
		closeWindows(1)
		advanceClock(0x3c)
	}
	exitBuilding(0, 1)
}

func faceYesNoDialog(face int, a, b []byte, mode, shown int) int {
	var i int
	var seq [16]byte
	s := bldgSprites()

	setMouseCursor(0, uint32(*gMouseCursor))
	for i = 0; i < 16; i++ {
		seq[i] = byte(i / 4)
	}
	pushButtons()
	if mode != 0 {
		setAnykeyButtons()
	} else {
		*gBtnXs = adrConfirmBtnXs
		*gBtnYs = adrConfirmBtnYs
		*gBtnWs = adrConfirmBtnWs
		*gBtnHs = adrConfirmBtnHs
	}
	bldgAnimSeq = seq[:]
	*gBldgAnimLen = 16
	*gBldgType = 7
	*gBldgAnimPos = 0
	if face > 0x1e {
		sprintf(gBldgDrawStr, "hire%u.fac", face-0x1f)
		*gBldgAnimLen = 1
	} else {
		sprintf(gBldgDrawStr, "eface%02u.out", face)
	}
	gBldgFaceImgs[0] = loadGraphic(gostr(gBldgDrawStr))
	gBldgFaceImgs[1] = loadGraphic("frame.out")
	sprintf(gBldgDrawStr, "\x05%04X", adrBldgFaceSprites)
	s[0].Y.Set(1)
	s[0].X.Set(1)
	s[1].Y.Set(7)
	s[1].Img.Set(uint32(gBldgFaceImgs[0]))
	s[0].Img.Set(uint32(gBldgFaceImgs[1]))
	s[1].Frame.Set(0)
	s[0].Frame.Set(0)
	sprintf(gTextBuf, "\r\n\n\t125\x03c%s\t000\x0b054%s\x05%04X", a, b, adrBldgFaceSprites)
	if shown != 0 {
		printWindowText(gTextBuf)
	} else {
		openTextWindow(8, 8, 0xd8, 0x84, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	}
	if mode == 0 {
		*gConfirm2SpritesImg = *gGlobalIcons
		pushIcons("confirm2.icn")
		*gConfirmSpritesImg2 = *gCurIcons
		sprintf(gTextBuf, "\x03b\x05%04X", adrConfirm2Sprites)
		printWindowText(gTextBuf)
	}
	waitMouseRelease()
loop:
	for {
		i = buildingWaitInput()
		if mode != 0 {
			i = 0
			break
		}
		switch i {
		case 0, 'Y':
			i = 1
			break loop
		case 1, 0x1b, 'N':
			i = 0
			break loop
		}
	}
	for *gMouseX = 0; *gMouseX < 2; *gMouseX++ {
		freeAndNull(&gBldgFaceImgs[*gMouseX])
	}
	popButtons()
	if mode == 0 {
		popIcons()
		drawMainIcons()
	}
	return i
}
