// Port of MAZEIO.C (CODE_8, seg 1511): clock strings, key/mouse input, roster and save-slot
// file I/O.
package mm3

import "os"

// Views where mazeio.c types a global differently from ds_gen.go.
var (
	gRosterFar  = at[Far](adrRoster)  // char far *g_roster (ds_gen has a near type)
	gPowIconFar = at[Far](adrPowIcon) // long g_pow_icon, a far pointer
	gPartyBtnYc = arr[int8](adrPartyBtnY)
	gMainBtnYc  = arr[int8](adrMainBtnY)
	gEquipBtnYc = arr[int8](adrEquipListBtnY)
	gEquipBtnHc = arr[int8](adrEquipListBtnH)
	gDrawListB  = arr[byte](adrDrawList) // u8 g_draw_list[]: the 4000-byte copy buffer
)

// The current button table: near pointers to int x[] (ended by -1), u8 y[], w[], h[].
func btnX(i int) int { return int(int16(le16(ds[:], int(*gBtnXs)+2*i))) }
func btnY(i int) int { return int(ds[int(*gBtnYs)+i]) }
func btnW(i int) int { return int(ds[int(*gBtnWs)+i]) }
func btnH(i int) int { return int(ds[int(*gBtnHs)+i]) }

// formatClock12h: 12-hour clock string "hh:mm a/p" in g_text_buf2+11.
func formatClock12h(t int32) []byte {
	var h int

	setMouseCursor(0, uint32(*gMouseCursor))
	sprintf(gTextBuf2, "%s", ctime(t))
	gTextBuf2[13] = 0
	h = int(int16(atol(gTextBuf2[11:])))
	if h < 12 {
		strcpy(gTextBuf2[19:], cs("a"))
		if h == 0 {
			strcpy(gTextBuf2[11:], cs("12"))
		}
	} else {
		if h -= 12; h == 0 {
			strcpy(gTextBuf2[11:], cs("12"))
		} else {
			sprintf(gTextBuf2[11:], "%02d", h)
		}
		strcpy(gTextBuf2[19:], cs("p"))
	}
	gTextBuf2[13] = ':'
	if gTextBuf2[11] == '0' {
		gTextBuf2[11] = ' '
	}
	return gTextBuf2[11:]
}

// formatHms: seconds -> "hh:mm:ss" (h, m, sec are 16-bit unsigned, h*3600 too).
func formatHms(s uint32) []byte {
	var m, sec, h uint16

	h = uint16(s / 3600)
	s -= uint32(h * 3600)
	m = uint16(s / 60)
	sec = uint16(s % 60)
	sprintf(gTmpText, "%02u:%02u:%02u", h, m, sec)
	return gTmpText
}

func resetSessionTimer() { *gSessionStart = timeNow() }

func getPlayTime() int32 {
	t := timeNow()
	return t - *gSessionStart + *gPlayTime
}

func formatClockStatus() []byte {
	var buf [100]byte

	t := timeNow()
	sprintf(buf[:], "\x02\x03l\x0b000\t000\x04160%s\x03r\t160%s\x01", formatClock12h(t), formatHms(uint32(getPlayTime())))
	strcpy(gTextBuf2, buf[:])
	return gTextBuf2
}

// drawButtonState highlights button idx of the current button table.
func drawButtonState(idx, sel int) {
	if *gBtnXs == adrPartyBtnX {
		idx -= 0x59
		switch idx {
		case 0, 1, 2, 3, 4, 5:
			if *gPartyBtnXEsc == 0 {
				drawIcon(Far(*gInnIcons), idx*2+(sel&1), int(gInnBtnSprs[idx][0])+0x10, 100, 0)
			}
		case 6:
			if *gPartyBtnXEsc != 0 {
				drawIcon(*gCurIcons, sel&1, int(gPartyBtnX[idx]), int(gPartyBtnYc[idx]), 0)
			}
		}
		return
	}
	if *gGameMode == 10 && idx < 20 {
		drawCharsheetCursor(0)
		*gStatCursor = int8(idx)
		drawCharsheetCursor(1)
	}
	if *gBtnXs == adrMazeBtnX && idx == 9 {
		drawCompassButton(sel - 0x12)
		return
	}
	drawIcon(*gCurIcons, sel, btnX(idx), btnY(idx), 0)
}

// readKey reads a key, folding the shift state and extended codes.
func readKey() int {
	var k, hi int
	var sh int

	k = int(int16(bioskey(0)))
	sh = bioskey(2)
	*gKeyRshift = int16(sh & 1)
	*gKeyLshift = int16(sh & 2)
	*gKeyCtrl = int16(sh & 4)
	*gKeyAlt = int16(sh & 8)
	*gKeyExtended = 0
	if *gUserBreak != 0 {
		*gUserBreak = 0
		return 0xff
	}
	if k == int(int16(-0xc00)) { // 0xf400
		k = 0x4d36
	}
	if k&0xff != 0 {
		hi = k & 0xff00
		k &= 0xff
		if hi == 0x3500 || hi == 0x3700 || hi == 0x4a00 || hi == 0x4e00 {
			*gKeyExtended++
		}
	} else {
		k >>= 8 // a signed shift, as the C's int
		if k >= 0x3b && k <= 0x44 {
			k += 0x8e
		} else {
			*gKeyExtended++
			k = int(int8(ds[adrKeypadKeymap+k-0x47])) // keypad scan codes 0x47..
		}
	}
	k &= 0xff
	if (*gKeyCtrl != 0 && k != 0x3a && k != 0x25 && k != 0x11) || *gKeyAlt != 0 {
		k = 0xff
	}
	return toupper(k & 0xff)
}

func inMazeRect(x0, x1, y0, y1 int) bool {
	return int(*gMouseX) >= x0 && int(*gMouseX) <= x1 && int(*gMouseY) >= y0 && int(*gMouseY) <= y1
}

// getInput is the main input loop: returns a key code or a button/region code.
func getInput() int {
	var flag, idx, i, k int

	for {
		if *gMusicLoadPending != 0 {
			if sndPlayMusic(-1) == 0 {
				loadMusic()
			}
		}
		if *gDebugOverlay != 0 {
			printWindowText(formatClockStatus())
		}
		if *gGameMode == 7 && *gObjD0c != 0 {
			o := &gObjects1based[*gObjD0c]
			switch o.Type.Get() {
			case 4, 0xc, 0x12, 0x14, 0x17, 0x26:
				*gObjAnimTick++
				o.Frame.Set(int(*gObjAnimTick)%4 + 1)
			case 0x21, 0x28:
				o.Frame.Set(1)
			}
		}
		if *gBtnXs == adrMazeBtnX || *gGameMode == 3 || *gGameMode == 6 || *gGameMode == 8 || *gGameMode == 7 || (*gGameMode == 2 && *gBtnXs == adrTrainBtnXs) {
			refreshView()
			if *gSpellAnimActive == 0 && *gMonScanPending != 0 {
				*gMonScanPending = 0
				moveMonsters()
			}
			*gViewDirty = 0
		} else if *gGameMode == 4 {
			bounceDice()
		} else if *gGameMode == 10 {
			drawCharsheetCursor(1)
		}
		if *gPartyDefeated != 0 {
			return -1
		}
		if *gGameMode == 2 && *gCombatActor != -1 {
			switch getWorstCondition(&gParty[gCombatOrder[*gCombatActor]]) {
			case 8, 11, 12, 13, 14, 15:
				return 0x42
			}
		}
		*gMouseButtons = int8(getMouseState(gMouseX, gMouseY))
		if *gMouseButtons != 0 {
			*gKeyRshift, *gKeyLshift, *gKeyCtrl, *gKeyAlt, *gKeyExtended = 0, 0, 0, 0, 0
			if *gBtnXs == adrMazeBtnX && *gGameMode != 2 {
				if inMazeRect(8, 0x3d, 8, 0x47) {
					return 0xf0
				}
				if inMazeRect(0xaa, 0xda, 8, 0x47) {
					return 0xf1
				}
				if inMazeRect(0x3e, 0xa9, 8, 0x47) {
					return 0xf2
				}
				if inMazeRect(0x3e, 0xa9, 0x48, 0x7c) {
					return 0xf3
				}
				if inMazeRect(8, 0x3d, 0x48, 0x7c) {
					return 0xf4
				}
				if inMazeRect(0xaa, 0xda, 0x48, 0x7c) {
					return 0xf5
				}
			}
			i = hitTestPortrait()
			if i != 0 {
				i--
				idx = i
				return idx + 0xc9
			}
			if *gBtnXs == adrAnyclickBtnXs && *gGameMode != 3 {
				return -1
			}
			i = hitTestButton()
			if i > 100 {
				return i - 0x60
			}
			if *gBtnXs == adrScrollBtnX && i > 2 {
				return i - 1
			}
			if *gGameMode == 2 && i > 10 {
				return i - 1
			}
			if i != 0 {
				i--
				idx = i
				if *gBtnXs == adrCtrlpanelBtnX {
					return idx
				}
				drawButtonState(idx, idx*2+1)
				if *gGameMode == 2 && i == 0 {
					*gCombatBtnHeld = 1
					return 0
				}
				playSfx(0x14)
				flag = 1
				for {
					i = hitTestButton()
					hit := false
					if i != 0 {
						i--
						hit = i == idx
					}
					if hit {
						if flag == 0 {
							flag = 1
							drawButtonState(idx, idx*2+1)
						}
					} else if flag != 0 {
						flag = 0
						drawButtonState(idx, idx*2)
					}
					if (*gBtnXs == adrMazeBtnX && i != 9) || *gGameMode == 3 || *gGameMode == 6 || *gGameMode == 8 || *gGameMode == 7 {
						refreshView()
					} else if *gGameMode == 4 {
						bounceDice()
					}
					if *gBtnXs == adrScrollBtnX {
						return i
					}
					if getMouseState(gMouseX, gMouseY) == 0 {
						break
					}
				}
				drawButtonState(idx, idx*2)
				if flag != 0 {
					if *gBtnXs == adrPartyBtnX {
						return idx - 0x59
					}
					return idx
				}
			} else if *gBtnXs == adrCtrlpanelBtnX {
				return 0x1b
			}
		}
		if *gCombatBtnHeld != 0 {
			*gCombatBtnHeld = 0
			if *gGameMode == 2 && *gBtnXs == adrMazeBtnX {
				drawIcon(*gCurIcons, 0, int(gMazeBtnX[0]), int(gMainBtnYc[0]), 0)
			}
		}
		k = 0
		if *gBtnXs == adrMazeBtnX && *gGameMode != 2 {
			for _, r := range [...]struct{ x0, x1, y0, y1, cur, ret int }{
				{8, 0x3d, 8, 0x47, 6, 0xf0}, {0xaa, 0xda, 8, 0x47, 5, 0xf1}, {0x3e, 0xa9, 8, 0x47, 1, 0xf2},
				{0x3e, 0xa9, 0x48, 0x7c, 2, 0xf3}, {8, 0x3d, 0x48, 0x7c, 4, 0xf4}, {0xaa, 0xda, 0x48, 0x7c, 3, 0xf5},
			} {
				if inMazeRect(r.x0, r.x1, r.y0, r.y1) {
					setMouseCursor(r.cur, uint32(*gMouseCursor))
					if getMouseState(gMouseX, gMouseY) != 0 {
						return r.ret
					}
					k++
				}
			}
		}
		if k == 0 {
			setMouseCursor(0, uint32(*gMouseCursor))
		}
		if *gMouseButtons != 0 {
			if *gGameMode == 0x56 || *gGameMode == 4 {
				return 0x56
			}
			playSfx(0x15)
			for getMouseState(gMouseX, gMouseY) != 0 {
			}
		}
		if bioskey(1) == 0 && *gUserBreak == 0 {
			continue
		}
		*gMouseButtons = 0
		k = readKey()
		if k == 0xd2 && *gBossKeyBusy == 0 {
			bossKey()
			continue
		}
		if *gGameMode == 10 {
			switch k {
			case 0xf1:
				if int(*gStatCursor)+5 >= 20 {
					continue
				}
				drawCharsheetCursor(0)
				*gStatCursor += 5
				continue
			case 0xf0:
				if int(*gStatCursor)-5 < 0 {
					continue
				}
				drawCharsheetCursor(0)
				*gStatCursor -= 5
				continue
			case 0xf2:
				if int(*gStatCursor)-1 < 0 {
					continue
				}
				drawCharsheetCursor(0)
				*gStatCursor--
				continue
			case 0xf3:
				if int(*gStatCursor)+1 >= 20 {
					continue
				}
				drawCharsheetCursor(0)
				*gStatCursor++
				continue
			case 0xd:
				return int(*gStatCursor)
			}
		}
		return k
	}
}

// hitTestButton: which button of the current table is under the mouse.
func hitTestButton() int {
	var i int
	mx, my := int(*gMouseX), int(*gMouseY)

	i = 0
	for {
		if btnY(i) <= my && my < btnY(i)+btnH(i) && btnX(i) <= mx && mx < btnX(i)+btnW(i) {
			if *gBtnXs == adrPartyBtnX {
				return i + 0x5a
			}
			return i + 1
		}
		i++
		if btnX(i) == -1 {
			break
		}
	}
	if *gBtnXs == adrItemscrHotX {
		i = 0
		for {
			if int(gEquipBtnYc[i]) <= my && my < int(gEquipBtnYc[i])+int(gEquipBtnHc[i]) && int(gEquipListBtnX[i]) <= mx && mx < int(gEquipListBtnX[i])+int(gEquipListBtnW[i]) {
				return i + 0x65
			}
			i++
			if gEquipListBtnX[i] == -1 {
				break
			}
		}
	}
	if *gGameMode == 2 {
		if *gMonD0c2 != 0 && *gMonD0c3 == 0 {
			i = 3
		} else {
			i = 0
		}
		for {
			if my >= 0x2a && my < 0x80 && int(gCombatHitX[i]) <= mx && mx < int(gCombatHitX[i])+int(gCombatHitW[i]) {
				return i + 0xe
			}
			i++
			if i >= 5 {
				break
			}
		}
	}
	return 0
}

// hitTestPortrait: which party member portrait is under the mouse.
func hitTestPortrait() int {
	if *gPortraitClickOn == 0 {
		return 0
	}
	for i := 0; i < int(*gPartyCount); i++ {
		if *gMouseY >= 0x95 && *gMouseY < 0xb5 && gPortraitX[i] <= *gMouseX && int(*gMouseX) < int(gPortraitX[i])+0x20 {
			return i + 1
		}
	}
	return 0
}

func waitMouseRelease() {
	for getMouseState(gMouseX, gMouseY) != 0 {
	}
}

func flushInput() {
	for bioskey(1) != 0 {
		bioskey(0)
	}
	waitMouseRelease()
}

func waitKey() {
	for {
		vgaIdle1b()
		getMouseState(gMouseX, gMouseY)
		if bioskey(1) != 0 || *gUserBreak != 0 {
			break
		}
	}
	printWindowText(cs(""))
	bioskey(0)
}

func idleFrames(n int) {
	for {
		getMouseState(gMouseX, gMouseY)
		if *gBtnXs == adrMazeBtnX || *gGameMode == 3 {
			refreshView()
			*gViewDirty = 0
		}
		n = int(int16(n - 1))
		if n == 0 {
			break
		}
	}
}

func idleFramesOrKey(n int) {
	var r int

	for {
		r = getMouseState(gMouseX, gMouseY)
		n--
		if *gBtnXs == adrMazeBtnX || *gGameMode == 3 {
			refreshView()
			*gViewDirty = 0
		}
		if !(n != 0 && bioskey(1) == 0 && r == 0) {
			break
		}
	}
}

func waitTicksOrKey(n int) {
	r := 0
	n = int(int16(n * int(*gMsgDelay)))
	sndTimer(-2)
	for sndTimer(-1) < n && bioskey(1) == 0 && r == 0 {
		r = getMouseState(gMouseX, gMouseY)
		refreshView()
		*gViewDirty = 0
	}
	for bioskey(1) != 0 {
		bioskey(0)
	}
}

// writeSaveGame saves the game.
func writeSaveGame() {
	*gPlayTime = getPlayTime()
	curWriteFile("maze.nam", gSaveName)
	curWriteFile("maze.chr", gRosterFar.B())
	curWriteFile("maze.pty", ds[adrPartyCount:])
	copyWorkToSave()
	resetSessionTimer()
}

// initWorkArchive writes the outdoor spell files into the work archive (MM3.CUR).
func initWorkArchive() {
	var p Far

	if *gArchiveOpen != 0 {
		closeCurArchive()
	}
	close(open(str(*gStrWorkArchive), 0x8302, 0x180))
	fail := func() { quitToDos() }
	if fileOpen(str(*gStrWorkArchive)) != 0 {
		fail()
	}
	for _, n := range []string{"out0.spl", "out1.spl", "out2.spl", "out3.spl"} {
		p = loadCCFile(n)
		if fileWrite(p.B(), int(*gFileSize)) != 0 {
			fail()
		}
		freeAndNull(&p)
	}
	if fileClose() != 0 {
		fail()
	}
	ccCur = nil // the cached MM3.CUR is stale: ccLoad re-reads it
	if *gArchiveOpen != 0 {
		openCurArchive()
	}
}

// copyFile is the body shared by copy_work_to_save and copy_save_to_work: 4000-byte blocks
// through g_draw_list.
func copyFile(from, to []byte) {
	var f, g *os.File

	if *gArchiveOpen != 0 {
		closeCurArchive()
	}
	if f = fopen(from, "rb"); f == nil {
		quitToDos()
	}
	fseek(f, 0, 2)
	l := uint32(ftell(f))
	rewind(f)
	if g = fopen(to, "wb"); g == nil {
		quitToDos()
	}
	nblk := l / 4000
	rem := l % 4000
	for i := 0; uint32(int16(i)) < nblk; i++ {
		fread(gDrawListB, 1, 4000, f)
		fwrite(gDrawListB, 1, 4000, g)
	}
	fread(gDrawListB, 1, int(rem), f)
	fwrite(gDrawListB, 1, int(rem), g)
	fclose(g)
	fclose(f)
	ccCur = nil // MM3.CUR may have changed: ccLoad re-reads it
	if *gArchiveOpen != 0 {
		openCurArchive()
	}
}

// copyWorkToSave copies the work archive to the current save slot.
func copyWorkToSave() { copyFile(str(*gStrWorkArchive), gSaveFiles[*gSaveSlot][:]) }

// copySaveToWork copies the current save slot to the work archive.
func copySaveToWork() { copyFile(gSaveFiles[*gSaveSlot][:], str(*gStrWorkArchive)) }

func writeFileOrDie(name []byte, buf []byte, n int) int {
	fd := open(name, 0x8302, 0x180)
	close(fd)
	if fileOpen(name) != 0 || fileWrite(buf, n) != 0 || fileClose() != 0 {
		textmode(-1)
		cputs(gostr(str(*gMsgAllocError)))
		exit(1)
	}
	return 0 // the C returns an uninitialised local
}

// saveLoadedMazes saves the four loaded maps.
func saveLoadedMazes() {
	for i := 0; i < 4; i++ {
		if gQuadrantMaps[i] != 0xff {
			sprintf(gTmpText, "maze%02u.dat", gQuadrantMaps[i])
			curWriteFile(gostr(gTmpText), bytesOf(&gMaps[i]))
		}
	}
}

func readFileInto(name []byte, buf []byte, n int) int {
	fd := open(name, 0x8001, 0x180)
	r := read(fd, buf, n)
	close(fd)
	return r
}

// loadPartyData loads party name, roster and party from the archive.
func loadPartyData() {
	var p Far

	p = loadCCFile("maze.nam")
	copy(gSaveName[:0x1f], p.B())
	freeAndNull(&p)
	p = loadCCFile("maze.chr")
	copy(gRosterFar.B()[:0x2382], p.B())
	freeAndNull(&p)
	p = loadCCFile("maze.pty")
	copy(ds[adrPartyCount:adrPartyCount+0x396], p.B())
	freeAndNull(&p)
}

// readSaveSummary reads the party name and highest level from a save slot.
func readSaveSummary(slot int, name []byte, maxlvl *uint8) {
	var offs [2]byte
	var buf [0x396]byte

	fd := open(gSaveFiles[slot][:], 0x8001, 0x180)
	read(fd, offs[:], 2)
	close(fd)
	o := uint16(le16(offs[:], 0)*8 + 2)
	*gRosterFar = memAlloc(0x2382)
	if fileOpen(gSaveFiles[slot][:]) != 0 || fileSeek(uint32(o)) != 0 || fileRead(name, 0x1f) != 0 ||
		fileSeek(uint32(o+0x1f)) != 0 || fileRead(gRosterFar.B(), 0x2382) != 0 || fileRead(buf[:], 0x396) != 0 || fileClose() != 0 {
		textmode(-1)
		cputs("Error reading Roster file")
		exit(1)
	}
	*maxlvl = 0
	for i := 0; i < int(buf[0]); i++ {
		p := gRosterFar.Add(int(buf[i+2]) * 0x12f)
		lvl := getCharLevel(farAt[Chr](p))
		if p.U8(0) != 0 && int(*maxlvl) < lvl {
			*maxlvl = uint8(lvl)
		}
	}
	freeAndNull(gRosterFar)
}

func probeFile(name []byte) int {
	fd := open(name, 0x400)
	close(fd)
	return fd
}

func freeAndNull(p *Far) {
	if *p != 0 {
		freeGraphic(*p)
	}
	*p = 0
}

// findSaveFiles collects up to 10 save-slot file names.
func findSaveFiles(pat []byte) {
	var ff ffblk

	*gSaveCount = 0
	memset(ds[adrSaveFiles:], 0, 0x82)
	r := findfirst(pat, &ff, 0)
	for r == 0 && *gSaveCount < 10 {
		strcpy(gSaveFiles[*gSaveCount][:], ff.ffName[:])
		*gSaveCount++
		r = findnext(&ff)
	}
}

// loadMonsterTables loads the monster stat tables.
func loadMonsterTables() {
	*gMonMaxHp = loadCCFile("MonHP.dat")
	*gMonAc = loadCCFile("MonAC.dat")
	*gMonSpeed = loadCCFile("MonSpd.dat")
	*gMonDmgDice = loadCCFile("MonDmgN.dat")
	*gMonDmgSides = loadCCFile("MonDmgS.dat")
	*gMonNumAttacks = loadCCFile("MonNumA.dat")
	*gMonExp = loadCCFile("MonExp.dat")
	*gMonDmgType = loadCCFile("MonDmgT.dat")
	*gMonAttackPref = loadCCFile("MonAttP.dat")
	*gMonSpecial = loadCCFile("MonSpec.dat")
	*gMonRanged = loadCCFile("MonRang.dat")
	*gMonHitBonus = loadCCFile("MonHitB.dat")
	*gMonTreasure = loadCCFile("MonTrea.dat")
	*gMonGold = loadCCFile("MonGold.dat")
	*gMonGems = loadCCFile("MonGems.dat")
	*gMonResMagic = loadCCFile("MonMagi.dat")
	*gMonResFire = loadCCFile("MonFire.dat")
	*gMonResElec = loadCCFile("MonElec.dat")
	*gMonResCold = loadCCFile("MonCold.dat")
	*gMonResAcid = loadCCFile("MonAcid.dat")
	*gMonResEnergy = loadCCFile("MonEner.dat")
	*gMonResPhys = loadCCFile("MonPhys.dat")
}
