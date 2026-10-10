// Port of INIT.C (overlay 17f2): boss key, intro, config checks, equip rules.
package mm3

// bossKey (0000): the boss key: save the screen, show a DOS prompt, wait for a key.
func bossKey() {
	if *gVideoMode != 0 {
		*gBossKeyBusy = 1
		showMessage(cs("\x0b007Sorry,  Boss Key only works in\n256 color mode!"))
		*gBossKeyBusy = 0
	} else {
		if *gMusicEnabled != 0 {
			sndPlayMusic(0)
		}
		fadeScreen(0, 0x40)
		close(open(str(*gBossFileName), 0x8302, 0x180))
		fileOpen(str(*gBossFileName))
		fileWrite(vram[:], 0xfa00)
		fileClose()
		clearScreen()
		flushInput()
		// ponytail: int 10h mode 3 and the "C:\>" prompt are not shown; the screen stays black
		cputs("C:\\>")
		if readKey() == 0x11 {
			system("del boss.$$$")
			quitToDos()
		}
		if *gMusicEnabled != 0 && *gMusicOn != 0 {
			sndStartMusic(*gMusicBuf)
		}
		fadeScreen(0, 0x40)
		fileOpen(str(*gBossFileName))
		fileRead(vram[:], 0xfa00)
		fileClose()
		fadeScreen(1, 4)
		setMouseCursor(0, uint32(*gMouseCursor))
		system("del boss.$$$")
	}
}

// doRangedAttacks (0132): power-up animation, then the ranged attackers' turns.
func doRangedAttacks() {
	var i, k int

	for i = 0; i < 0x24; i++ {
		if gRangedAttackers[i] != 0xff {
			if k = int(gMissileIcons[gMonDmgType.B()[gRangedAttackers[i]]]); k != 12 {
				break
			}
		}
	}
	sprintf(gTmpText, "pow%d.icn", k)
	if *gPowIconFar != 0 {
		freeAndNull(gPowIconFar)
	}
	*gPowIconFar = loadGraphic(gostr(gTmpText))
	playSfx(int(gMissileSounds[gMonDmgType.B()[gRangedAttackers[i]]]))
	for *gPowIconFar != 0 {
		refreshView()
	}
	if *gGameMode != 2 {
		for i = 0; i < int(*gPartyCount); i++ {
			gCombatOrder[i] = uint8(i)
		}
		*gCombatPartyN = *gPartyCount
	}
	for i = 0; i < 0x24; i++ {
		if gRangedAttackers[i] != 0xff {
			monsterTurn(int(gRangedAttackers[i]))
		}
	}
	*gMonScanPending, *gMonMissileActive = 0, 0
	updatePartyDisabled()
	if *gGameMode == 5 {
		for i = 0; i < int(*gPartyCount); i++ {
			switch getWorstCondition(&gParty[i]) {
			case 8, 11, 12, 13, 14, 15:
			default:
				*gGameMode = 1
				return
			}
		}
	}
}

// breakAllItems (0299)
func breakAllItems(c *Chr) {
	for i := 0; i < 0x12; i++ {
		if c.ItemID[i] < 0x48 {
			c.ItemFlags[i] |= 0x80
		}
	}
}

// breakEquippedWeapons (02bd)
func breakEquippedWeapons(c *Chr) {
	for i := 0; i < 0x12; i++ {
		if c.ItemID[i] < 0x1e && c.ItemEq[i] != 0 {
			c.ItemFlags[i] |= 0x80
		}
	}
}

// loadCombatIcons (02e7)
func loadCombatIcons() {
	freeAndNull(gCurIcons)
	*gCurIcons = loadGraphic("combat.icn")
	*gCpanelFrameImg = *gGlobalIcons
	*gCombatMenuImg = *gCurIcons
	*gCpanelButtonsImg = *gCombatMenuImg
	setCombatButtons()
}

// showGuildMessage (0331)
func showGuildMessage() {
	pushButtons()
	setAnykeyButtons()
	strncpy(gTextBuf, gTextTable.Add(int(gGuildMsgTextOffs[*gMapNum])).B(), 0x12c)
	openTextWindow(0x5a, 0x2d, 0x82, 0x70, 0, 0x96, 4, 6, dsFar(adrTextBuf))
	flushInput()
	getInput()
	closeWindows(1)
	popButtons()
}

// setBuildingExitPos (03ac)
func setBuildingExitPos(n int) {
	if n != 6 {
		*gPartyX = gBuildingExitXy[n][*gMapNum][0]
		*gPartyY = gBuildingExitXy[n][*gMapNum][1]
	} else {
		*gPartyX = 9
		*gPartyY = 0xb
	}
}

// exitBuilding (03fb)
func exitBuilding(a, b int) {
	if b != 0 && b != 4 {
		clearCharHighlight()
		showPleaseWait()
	}
	for i := 0; i < 4; i++ {
		freeAndNull(&gBldgFaceImgs[i])
	}
	if a != 1 {
		enterMaze(int(*gMapNum))
	}
	setBuildingExitPos(b)
	*gPartyDir ^= 1
	if b != 4 {
		closeWindows(1)
	}
	if a == 0 {
		drawWizardEye(1)
	}
	flushInput()
	playSfx(0)
}

// getCharStatText (047e): stat text.
func getCharStatText(c *Chr, sel int, buf []byte) []byte {
	var tmp [30]byte
	var i, n int

	switch sel {
	case 0, 1, 2, 3, 4, 5, 6, 15, 16, 17, 19, 20, 23:
		strncpy(buf, gTextTable.Add(int(gTextOffsets[sel])).B(), 0x1c)
	case 10:
		for i, n = 0, 0; i < 3; i++ {
			if testBit(gGameFlags, int(gAlignQuestBits[i])) != 0 {
				n++
			}
		}
		if n == 2 {
			for i = 0; i < 3; i++ {
				if testBit(gGameFlags, int(gAlignQuestBits[i])) == 0 {
					break
				}
			}
			strncpy(tmp[:], gTextTable.Add(int(gTextOffsets[sel])).B(), 0x1c)
			sprintf(buf, gostr(tmp[:]), str(gAlignmentNames[i]))
		}
	case 7, 8, 9, 11, 12, 13, 14, 18, 21, 22:
		strncpy(tmp[:], gTextTable.Add(int(gTextOffsets[sel])).B(), 0x1c)
		sprintf(buf, gostr(tmp[:]), c.Award[sel])
	}
	return buf
}

// fatalMissingFile (05ba)
func fatalMissingFile(n int) {
	directvideo = 0
	sprintf(gTextBuf, "\n\r%s file is missing or corrupt... Run Install to reconfigure.\n\n", str(gConfigFileNames[n]))
	cputs(gostr(gTextBuf))
	exit(1)
}

// checkConfigAndDisk (05f2): config / disk checks.
func checkConfigAndDisk() {
	var nb [2]byte
	var free uint32
	var fd int

	directvideo = 0
	memset(ds[adrCfgVideo:], 0, 4)
	if fd = open(str(gConfigFileNames[0]), 0x8001, 0x100); fd == -1 {
		fatalMissingFile(0)
	}
	read(fd, ds[adrCfgVideo:], 4)
	close(fd)
	if fd = open(str(gConfigFileNames[1]), 0x8001, 0x100); fd == -1 {
		fatalMissingFile(1)
	}
	close(fd)
	free = getFreeDiskBytes()
	if fd = open(cs("mm3.cc"), 0x8001, 0x100); fd == -1 {
		fatalMissingFile(1)
	}
	read(fd, nb[:], 2)
	close(fd)
	n := int(int16(le16(nb[:], 0)*8 + 2))
	if fileOpen(cs("mm3.cc")) != 0 {
		fatalMissingFile(1)
	}
	if fileSeek(uint32(uint16(n))) != 0 || fileRead(gTextBuf2, 0xed) != 0 || fileRead(gTmpText, 0x83) != 0 || fileClose() != 0 {
		fatalMissingFile(1)
	}
	if free < 100000 {
		setcursortype(0)
		openTextmodeBox(0x13, 6, 0x3d, 0x12, 0x14, cs("Alert!"))
		sprintf(gTextBuf, gostr(gTextBuf2), free)
		cprintf(gTextBuf)
		bioskey(0)
		restoreTextScreen()
		setcursortype(2)
		exit(1)
	}
	updateFreeMemKb()
	if *gFreeMemKb < 0x17b {
		free = 0x89c00
		setcursortype(0)
		openTextmodeBox(0x13, 6, 0x3d, 0x12, 0x14, cs("Alert!"))
		sprintf(gTextBuf, gostr(gTmpText), free, uint32(0x17b-*gFreeMemKb)<<10)
		cprintf(gTextBuf)
		bioskey(0)
		restoreTextScreen()
		setcursortype(2)
		exit(1)
	}
}

// getFreeDiskBytes (083e): free disk bytes (getdfree). ponytail: the host disk is not
// asked; the game only checks for at least 100000.
func getFreeDiskBytes() uint32 { return 0x7fffffff }

// drawTextBox (0887): text-mode box.
func drawTextBox(x1, y1, x2, y2, px int, s []byte) {
	textcolor(15)
	textbackground(4)
	clrscr()
	gotoxy(1, 1)
	putchar(0xda)
	for i := 2; i < x2-x1+1; i++ {
		putchar(0xc4)
	}
	putchar(0xbf)
	for i := 2; i < y2-y1+1; i++ {
		gotoxy(1, i)
		putchar(0xb3)
		gotoxy(x2-x1+1, i)
		putchar(0xb3)
	}
	gotoxy(1, y2-y1+1)
	putchar(0xc0)
	for i := 2; i < x2-x1+1; i++ {
		putchar(0xc4)
	}
	putchar(0xd9)
	gotoxy(1, 3)
	putchar(0xc3)
	for i := 2; i < x2-x1+1; i++ {
		putchar(0xc4)
	}
	putchar(0xb4)
	gotoxy(px, 2)
	cprintf(s)
}

// restoreTextScreen (0ad9)
func restoreTextScreen() {
	window(1, 1, 80, 25)
	textmode(-1)
	textbackground(0)
	clrscr()
}

// openTextmodeBox (0b0b)
func openTextmodeBox(x1, y1, x2, y2, px int, s []byte) {
	clrscr()
	window(x1, y1, x2, y2)
	drawTextBox(x1, y1, x2, y2, px, s)
	window(x1+1, y1+3, x2-1, y2-1)
	clrscr()
}

// rangedPathDistance (0b65): line-of-movement check.
func rangedPathDistance(dx, dy int) int {
	px, py := int(*gPartyX), int(*gPartyY)
	if dx != 0 {
		if dx > 0 {
			for i := 1; i <= dx; i++ {
				if getCellWallsRegion(px+i, py, 8) != 0 {
					return 0
				}
			}
			if *gPartyDir == 2 {
				return dx + 1
			}
			return 1
		}
		for i := dx; i != 0; i++ {
			if getCellWallsRegion(px+i, py, 0x800) != 0 {
				return 0
			}
		}
		if *gPartyDir == 3 {
			return -dx + 1
		}
		return 1
	}
	if dy > 0 {
		for i := 1; i <= dy; i++ {
			if getCellWallsRegion(px, py+i, 0x80) != 0 {
				return 0
			}
		}
		if *gPartyDir == 0 {
			return dy + 1
		}
		return 1
	}
	for i := dy; i != 0; i++ {
		if getCellWallsRegion(px, py+i, 0x8000) != 0 {
			return 0
		}
	}
	if *gPartyDir == 1 {
		return -dy + 1
	}
	return 1
}

// queueRangedAttack (0c7a)
func queueRangedAttack(c, x, y int) {
	for i := 0; i < 0x24; i++ {
		if gRangedAttackers[i] == 0xff {
			dy := y - int(*gPartyY)
			dx := x - int(*gPartyX)
			r := rangedPathDistance(dx, dy)
			if r == 0 {
				return
			}
			*gMonMissileActive = 1
			gRangedAttackers[i] = uint8(c)
			if r-1 == 0 {
				return
			}
			for i = 0; i < 8; i++ {
				if gMissilePos[i] == 0 {
					gMissilePos[i] = gRangedDistDelays[r]
					gMissileAttackers[i] = uint8(c)
					return
				}
			}
			return
		}
	}
}

// restockShops (0cfd)
func restockShops() {
	memset(ds[adrShopStockFlagsT1:], 0, 0x32a)
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 9; k++ {
				if gFixedStockIds[i][j][k] != 0 {
					gShopStockIdsT1[i][j][k] = gFixedStockIds[i][j][k]
					gShopStockMatsT1[i][j][k] = gFixedStockMaterials[i][j][k]
					if j == 2 {
						gShopStockFlagsT1[i][j][k] = gFixedMiscCharges[i][k]
						gShopStockPowersT1[i][j][k] = gFixedMiscPowers[i][k]
					}
				} else {
					generateRandomItem(int(gShopItemLevels[i]), gScratchChar, 0)
					gShopStockIdsT1[i][j][k] = uint8(randomRange(int(gShopCatMinIds[j]), int(gShopCatMaxIds[j])))
					gShopStockMatsT1[i][j][k] = gShopInvMaterials[0]
					gShopStockPowersT1[i][j][k] = gShopInvPowers[0]
					gShopStockElemsT1[i][j][k] = gShopInvElems[0]
					gShopStockAttribsT1[i][j][k] = gShopInvAttribs[0]
					gShopStockFlagsT1[i][j][k] = gShopInvFlags[0]
				}
			}
		}
	}
}

// loadMusic (0f06)
func loadMusic() {
	p := loadCCFile(gostr(gCurMusic))
	copy(gMusicBuf.B()[:*gFileSize], p.B())
	freeAndNull(&p)
	if *gMusicOn != 0 {
		sndStartMusic(*gMusicBuf)
	}
	*gMusicLoadPending = 0
}

// playMusic (0f61)
func playMusic(name []byte, flag uint8) {
	strcpy(gPrevMusic, gCurMusic)
	strcpy(gCurMusic, name)
	if *gMusicEnabled != 0 {
		sndPlayMusic(int32(flag) * 65536)
		if flag != 0 {
			*gMusicLoadPending = 1
		} else {
			loadMusic()
		}
	}
}

// clearCharHighlight (0fac)
func clearCharHighlight() {
	if *gHighlightedChar > -1 {
		drawIcon(*gGlobalIcons, int(*gHighlightedChar)+3, int(*gHighlightedChar)*0x26+0xa, 0x94, 0)
		*gHighlightedChar = -1
	}
}

func animSprites() []byte { return ds[adrAnimSprites:] }

// playIntro (0fe9): the intro sequence. The C's abort labels are the abortN closures.
func playIntro() {
	var i, f, k int
	s := gAnimSprites
	draw := func() { drawSprites2a(animSprites()) }

	memset(animSprites(), 0, 0xd8)
	s[0].Flag.Set(-1)
	s[1].Flag.Set(-1)
	s[2].Flag.Set(-1)
	s[0].Y.Set(0x6f)
	if *gMusicEnabled != 0 {
		playMusic(cs("mm3theme.m"), 0)
	}
	*gCineImg0 = loadGraphic("itit0.vga")
	*gCineImg1 = loadGraphic("itit1.vga")
	*gCineImg2 = loadGraphic("itit2.vga")
	*gCineImg3 = loadGraphic("itit3.vga")
	s[1].Img.Set(uint32(*gCineImg0))
	s[0].Img.Set(uint32(*gCineImg1))
	draw()
	fadeScreen(1, 4)
	idleFrames(0x78)
	abort1 := func() {
		freeAndNull(gCineImg3)
		freeAndNull(gCineImg2)
		freeAndNull(gCineImg1)
		freeAndNull(gCineImg0)
		playSfx(0)
		fadeScreen(0, 0x10)
	}
	for i = 1; i < 0x10; i++ {
		if introCheckKey() != 0 {
			abort1()
			return
		}
		s[0].Frame.Set(i)
		if i == 10 {
			playSfx(0x14)
		}
		sndTimer(-2)
		draw()
		introWaitTicks(5)
		if i == 0xf {
			playSfx(10)
		}
	}
	s[0].Img.Set(uint32(*gCineImg2))
	for i = 0; i < 8; i++ {
		if introCheckKey() != 0 {
			abort1()
			return
		}
		s[0].Frame.Set(i)
		sndTimer(-2)
		draw()
		introWaitTicks(5)
	}
	s[0].Img.Set(uint32(*gCineImg3))
	for i = 0; i < 0xe; i++ {
		if introCheckKey() != 0 {
			abort1()
			return
		}
		s[0].Frame.Set(i)
		sndTimer(-2)
		draw()
		introWaitTicks(2)
	}
	playSfx(0)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg0)
	*gCineImg4 = loadGraphic("itit4.vga")
	clear0 := func() {
		s[1].Img.Set(0)
		s[0].Frame.Set(0)
		s[0].Attr.Set(0)
		s[0].Y.Set(0)
		s[0].X.Set(0)
	}
	clear0()
	s[0].Img.Set(uint32(*gCineImg4))
	fadeScreen(0, 4)
	playSfx(0)
	draw()
	fadeScreen(1, 4)
	freeAndNull(gCineImg4)
	clear0()
	*gCineImg1 = loadGraphic("itit5.vga")
	*gCineImg2 = loadGraphic("itit6.vga")
	*gCineImg3 = loadGraphic("itit7.vga")
	*gCineImg4 = loadGraphic("itit8.vga")
	*gCineImg5 = loadGraphic("comet.vga")
	*gCineImg0 = loadGraphic("itxt.vga")
	s[0].Img.Set(0)
	abort2 := func() {
		freeAndNull(gCineImg5)
		freeAndNull(gCineImg4)
		freeAndNull(gCineImg3)
		freeAndNull(gCineImg2)
		freeAndNull(gCineImg1)
		freeAndNull(gCineImg0)
		playSfx(0)
		fadeScreen(0, 0x10)
	}
	// wait: draw until the timer reaches n; false if a key aborted
	wait := func(n int) bool {
		for sndTimer(-1) < n {
			if introCheckKey() != 0 {
				return false
			}
			draw()
		}
		return true
	}
	sndTimer(-2)
	if !wait(0x78) {
		abort2()
		return
	}
	s[0].Img.Set(uint32(*gCineImg0))
	sndTimer(-2)
	playSfx(0x25)
	if !wait(0x16c) {
		abort2()
		return
	}
	s[0].Img.Set(0)
	sndTimer(-2)
	playSfx(0)
	if !wait(0x78) {
		abort2()
		return
	}
	s[0].Img.Set(uint32(*gCineImg0))
	s[0].Frame.Set(1)
	sndTimer(-2)
	if !wait(0x16c) {
		abort2()
		return
	}
	s[0].Img.Set(0)
	sndTimer(-2)
	if !wait(0x78) {
		abort2()
		return
	}
	playSfx(0x25)
	s[0].Img.Set(uint32(*gCineImg0))
	s[0].Frame.Set(2)
	sndTimer(-2)
	if !wait(0x16c) {
		abort2()
		return
	}
	s[0].Img.Set(0)
	sndTimer(-2)
	playSfx(0)
	if !wait(0xda) || !wait(0xda) {
		abort2()
		return
	}
	playSfx(0x37)
	s[0].X.Set(0x44)
	s[0].Y.Set(0x16)
	s[0].Img.Set(uint32(*gCineImg1))
	playSfx(0x26)
	for _, st := range []struct {
		img *Far
		n   int
	}{{gCineImg1, 5}, {gCineImg2, 5}, {gCineImg3, 4}, {gCineImg4, 6}} {
		if st.img == gCineImg4 {
			playSfx(10)
		}
		s[0].Img.Set(uint32(*st.img))
		for i = 0; i < st.n; i++ {
			if introCheckKey() != 0 {
				abort2()
				return
			}
			s[0].Frame.Set(i)
			sndTimer(-2)
			draw()
			introWaitTicks(5)
		}
	}
	sndTimer(-2)
	for sndTimer(-1) < 0x78 {
		draw()
		if introCheckKey() != 0 {
			abort2()
			return
		}
	}
	s[0].X.Set(0x6a)
	s[0].Y.Set(8)
	s[0].Img.Set(uint32(*gCineImg5))
	for i = 0; i < 0x10; i++ {
		if introCheckKey() != 0 {
			abort2()
			return
		}
		s[0].Frame.Set(i)
		sndTimer(-2)
		draw()
		introWaitTicks(5)
	}
	freeAndNull(gCineImg5)
	freeAndNull(gCineImg4)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg0)
	*gCineImg6 = loadGraphic("talk.vga")
	s[0].X.Set(0x68)
	s[0].Y.Set(0x36)
	s[0].Frame.Set(0)
	s[0].Img.Set(uint32(*gCineImg6))
	s[0].Frame.Set(0)
	draw()
	for i = 0; i < 7; i++ {
		draw()
		sprintf(gCineNameBuf, "s%d.s", i+1)
		gCineVoices[i] = loadCCFile(gostr(gCineNameBuf))
		gCineVoiceSizes[i] = *gFileSize
	}
	abort3 := func() {
		freeAndNull(gCineImg6)
		for i := 0; i < 7; i++ {
			freeAndNull(&gCineVoices[i])
		}
		playSfx(0)
		fadeScreen(0, 0x10)
	}
	sndTimer(-2)
	if !wait(0x50) {
		abort3()
		return
	}
	for i = 0; i < 7; i++ {
		sndTimer(-2)
		*gTalkFramePos = 0
		sndPlaySample(gCineVoices[i], int(gCineVoiceSizes[i]))
		for {
			s[0].Frame.Set(int(gIntroTalkFrames[i][*gTalkFramePos]))
			draw()
			introWaitTicks(int(gIntroTalkTimes[i][*gTalkFramePos]))
			for {
				*gTalkFramePos++
				if uint32(gIntroTalkTimes[i][*gTalkFramePos]) >= *gAnimTicks {
					break
				}
			}
			if sndPlaySample(1, 0) == 0 {
				break
			}
		}
		if bioskey(1) != 0 || *gUserBreak != 0 {
			k = readKey()
			sndPlaySample(0, 0)
			if k != 0x11 {
				abort3()
				return
			}
			quitToDos()
		}
		*gIntroTalkEndTick = int16(*gAnimTicks)
		s[0].Frame.Set(0)
		draw()
	}
	s[0].Frame.Set(0)
	sndTimer(-2)
	if !wait(0x50) {
		abort3()
		return
	}
	freeAndNull(gCineImg6)
	for i = 0; i < 7; i++ {
		freeAndNull(&gCineVoices[i])
	}
	if *gSndDev0 != 0 {
		playMusic(cs("mm3theme.m"), 0)
	}
	for i = 0; i < 0x26; i++ {
		setLe16(ds[:], adrAnimSprites+2*i, 0)
	}
	for i = 0; i < 0x26; i += 7 {
		setLe16(ds[:], adrAnimSprites+2*i, -1)
	}
	s[0].X.Set(0x38)
	s[0].Y.Set(0x12)
	s[1].X.Set(0xd8)
	s[1].Y.Set(0x83)
	s[1].Frame.Set(1)
	for i = 0; i < 5; i++ {
		sprintf(gCineNameBuf, "cr%d.vga", i+1)
		gCineVoices[i] = loadGraphic(gostr(gCineNameBuf))
		draw()
	}
	abort4 := func() {
		for i := 0; i < 5; i++ {
			freeAndNull(&gCineVoices[i])
		}
		playSfx(0)
		fadeScreen(0, 0x10)
	}
	sndTimer(-2)
	if !wait(0x78) {
		abort4()
		return
	}
	f = 1
	for i = 0; i < 5; i++ {
		s[0].Img.Set(uint32(gCineVoices[i]))
		sndTimer(-2)
		for sndTimer(-1) < 0x168 {
			if introCheckKey() != 0 {
				abort4()
				return
			}
			if f != 0 && sndTimer(-1) > 0xb4 && i == 1 {
				f = 0
				s[1].Img.Set(uint32(gCineVoices[i]))
			}
			draw()
		}
		s[0].Img.Set(0)
		s[1].Img.Set(0)
		sndTimer(-2)
		if !wait(100) {
			abort4()
			return
		}
	}
	for i = 0; i < 5; i++ {
		freeAndNull(&gCineVoices[i])
	}
	playSfx(0)
	fadeScreen(0, 1)
}

// introWaitTicks (1b77): wait n ticks, redrawing.
func introWaitTicks(n int) {
	for {
		*gAnimTicks = uint32(int32(sndTimer(-1)))
		if *gAnimTicks >= uint32(int32(n)) {
			break
		}
		drawSprites2a(animSprites())
	}
}

// introCheckKey (1bae): key check during the intro.
func introCheckKey() int {
	if bioskey(1) != 0 || *gUserBreak != 0 {
		k := readKey()
		if k == 0x11 {
			quitToDos()
		} else if k == 0x1b {
			return 1
		}
	}
	return 0
}

var equipConflictBuf [100]byte // show_equip_conflict's buf: the window gets its far address

// showEquipConflict (1be9): equip-conflict message.
func showEquipConflict(c *Chr, a, b int) {
	buf := equipConflictBuf[:]

	if a == b {
		return
	}
	pushButtons()
	setAnykeyButtons()
	if b < 0 {
		w := "medal"
		if b == -1 {
			w = "ring"
		}
		sprintf(buf, "\x03c\x0b007You have equipped all the %ss you can!", w)
	} else {
		sprintf(buf, "\x03c\x0b007You must remove %sto equip %s\x08 !", str(gItemNames[c.ItemID[b]]), str(gItemNames[c.ItemID[a]]))
	}
	openTextWindow(0x32, 0x4c, 0xd8, 0x30, 0, 0x96, 4, 6, farOf(buf))
	flushInput()
	getInput()
	closeWindows(1)
	popButtons()
	flushInput()
}

// equipItem (1cae): equip the item in slot.
func equipItem(c *Chr, slot int) {
	var i, n int
	cls, id := int(c.Cls), int(c.ItemID[slot])
	// conflict: the first equipped slot whose kind is one of ks
	conflict := func(ks ...uint8) bool {
		for i = 0; i < 0x12; i++ {
			for _, k := range ks {
				if c.ItemEq[i] == k {
					showEquipConflict(c, slot, i)
					return true
				}
			}
		}
		return false
	}
	count := func(k uint8) int {
		for i, n = 0, 0; i < 0x12; i++ {
			if c.ItemEq[i] == k {
				n++
			}
		}
		return n
	}

	for i = 0; i < 0x12; i++ {
		if c.ItemElem[slot] == c.ItemElem[i] && c.ItemMat[slot] == c.ItemMat[i] && c.ItemAttr[slot] == c.ItemAttr[i] &&
			c.ItemID[slot] == c.ItemID[i] && c.ItemPower[slot] == c.ItemPower[i] && i != slot && c.ItemEq[i] != 0 {
			showMessage(cs("You cannot equip two of the same items!"))
			return
		}
	}
	var eq uint8
	switch {
	case id <= 0x11:
		if classCanUseItem(cls, id, 0) == 0 || conflict(13, 1) {
			return
		}
		eq = 1
	case id <= 0x1d:
		if classCanUseItem(cls, id, 0) == 0 || conflict(1, 2, 13) {
			return
		}
		eq = 13
	case id <= 0x21:
		if classCanUseItem(cls, id, 0) == 0 || conflict(4) {
			return
		}
		eq = 4
	case id <= 0x29:
		if classCanUseItem(cls, id, 0) == 0 || conflict(3) {
			return
		}
		eq = 3
	case id <= 0x2a:
		if classCanUseItem(cls, id, 0) == 0 || conflict(2, 13) {
			return
		}
		eq = 2
	case id <= 0x2d:
		if conflict(5) {
			return
		}
		eq = 5
	case id <= 0x2e:
		if conflict(6) {
			return
		}
		eq = 6
	case id <= 0x2f:
		if count(8) > 9 {
			showEquipConflict(c, slot, -1)
			return
		}
		eq = 8
	case id <= 0x30:
		if conflict(9) {
			return
		}
		eq = 9
	case id <= 0x33:
		if conflict(10) {
			return
		}
		eq = 10
	case id <= 0x34:
		if conflict(12) {
			return
		}
		eq = 12
	case id <= 0x39:
		if count(7) > 3 {
			showEquipConflict(c, slot, -2)
			return
		}
		eq = 7
	case id <= 0x3c:
		if conflict(11) {
			return
		}
		eq = 11
	default:
		if conflict(2, 13) {
			return
		}
		eq = 2
	}
	c.ItemEq[slot] = eq
}

// quitToDos (2040)
func quitToDos() {
	restoreBreakHandlers()
	sndShutdown()
	cputs("Thank you for playing Might and Magic ]I[\n\n\r")
	exit(0)
}
