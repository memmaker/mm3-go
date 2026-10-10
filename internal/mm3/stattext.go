// Port of STATTEXT.C (CODE_6): the outdoor view: the per-frame animation tick and the draw list
// of terrain, walls, sprites and missiles, printed with the view caption and border.
package mm3

// gMonD1l1L is g_mon_d1l1 as stattext declares it: long[] (the outdoor terrain/wall images by
// cell value).
var gMonD1l1L = arr[int32](adrMonD1l1)

func refreshOutdoorView() {
	clearViewWalls()
	updateDarkness()
	scanViewObjects()
	if *gMonsterScanDelay != 0 && (*gGameMode == 1 || *gGameMode == 2) && *gMonMissileActive == 0 && *gLightChanged == 0 {
		if *gMonsterScanDelay--; *gMonsterScanDelay == 0 {
			moveMonsters()
		}
	}
	scanViewMonsters()
	drawOutdoorView()
}

func drawOutdoorView() {
	for i := 0; i < 8; i++ {
		gMissileSteps[i] = gMissilePos[i] - 1
	}
	x, y, d := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
	*gHoverPhase1 = (*gHoverPhase1 + 1) % 18
	*gHoverPhase2 = (*gHoverPhase2 + 1) % 18
	*gHoverPhase3 = (*gHoverPhase3 + 1) % 18
	*gOutdoorFlipTick = (*gOutdoorFlipTick + 1) % 4
	if *gOutdoorFlipTick == 0 {
		*gOutdoorFlip ^= 1
	}
	for j := 0; j < int(*gMonCount); j++ {
		if gMonStatus[j] == 0 {
			gMonFrames[j] = int16((int(uint16(gMonFrames[j])) + 1) % int(gMonFrameCounts[gMonType[j]]))
		}
	}
	for j := 0; j < int(*gObjCount); j++ {
		if *gGameMode != 7 || int(*gObjD0c)-1 != j {
			a := &gMapObjectsAnim[j]
			a.Frame.Set((a.Frame.Get() + 1) % int(gObjFrameCounts[a.Type.Get()]))
		}
	}
	gDrawList[0], gDrawList[7] = -1, -1
	gDrawList[3], gDrawList[4], gDrawList[10] = 8, 8, 8
	gDrawList[1], gDrawList[2], gDrawList[6], gDrawList[8], gDrawList[9], gDrawList[13] = 0, 0, 0, 0, 0, 0
	gDrawList[11] = 0x43
	p := 13
	gDrawList[5] = int16(*gViewBgFlip)
	gDrawList[12] = *gOutdoorFlip
	setLe32(ds[:], adrDrawList+2*1, uint32(*gSkyGfx))
	setLe32(ds[:], adrDrawList+2*8, uint32(*gWaterGfx))

	flip := int(*gViewFlip)
	nflip := b2i(flip == 0)
	cell := func(k, mask int) int {
		return getCellWallsRegion(int(int8(gViewDx[d][k]))+x, int(int8(gViewDy[d][k]))+y, mask)
	}
	// ent appends the image for cell value v (if any) at (sx, sy).
	ent := func(v, sx, sy, attr, frame int) {
		if v != 0 {
			p = dlPut(p, -1)
			p = dlPutL(p, Far(gMonD1l1L[v]))
			p++
			p = dlPut(p, sx)
			p = dlPut(p, sy)
			p = dlPut(p, attr)
			p = dlPut(p, frame)
		}
	}
	t := func(k int) int { return cell(k, 0x70) >> 4 } // terrain
	w := func(k int) int { return cell(k, 7) }         // wall/object
	ent(t(44), 8, 67, flip, int(gTerrainFramesD4x4[flip]))
	ent(t(36), 32, 67, flip, int(gTerrainFramesD4x3[flip]))
	ent(t(37), 56, 67, flip, int(gTerrainFramesD4x2[flip]))
	ent(t(38), 84, 67, flip, int(gTerrainFramesD4x1[flip]))
	ent(t(45), 176, 67, flip, int(gTerrainFramesD4x4[nflip]))
	ent(t(43), 152, 67, flip, int(gTerrainFramesD4x3[nflip]))
	ent(t(42), 115, 67, flip, int(gTerrainFramesD4x2[nflip]))
	ent(t(41), 115, 67, flip, int(gTerrainFramesD4x1[nflip]))
	ent(t(39), 105, 67, flip, 16)
	ent(t(22), 8, 73, flip, int(gTerrainFramesD3x2[flip]))
	ent(t(24), 36, 73, flip, int(gTerrainFramesD3x1[flip]))
	ent(t(31), 148, 73, flip, int(gTerrainFramesD3x2[nflip]))
	ent(t(29), 127, 73, flip, int(gTerrainFramesD3x1[nflip]))
	ent(t(26), 89, 73, flip, 11)
	ent(t(10), 8, 81, flip, int(gTerrainFramesD2x2[flip]))
	ent(t(11), 8, 81, flip, int(gTerrainFramesD2x1[flip]))
	ent(t(18), 196, 81, flip, int(gTerrainFramesD2x2[nflip]))
	ent(t(16), 143, 81, flip, int(gTerrainFramesD2x1[nflip]))
	ent(t(13), 65, 81, flip, 6)
	ent(t(5), 8, 93, flip, int(gTerrainFramesD1x1[flip]))
	ent(t(9), 167, 93, flip, int(gTerrainFramesD1x1[nflip]))
	ent(t(6), 33, 93, flip, 3)
	ent(t(0), 8, 109, flip, int(gTerrainFramesD0x1[flip]))
	ent(t(4), 199, 109, flip, int(gTerrainFramesD0x1[nflip]))
	ent(t(1), 8, 109, flip, 0)
	ent(w(44), 0, 56, 0x202, 2)
	ent(w(36), 24, 56, 0x200, 2)
	ent(w(37), 48, 56, 0x200, 2)
	ent(w(38), 72, 56, 0x200, 2)
	ent(w(45), 200, 56, 0x203, 2)
	ent(w(43), 176, 56, 0x201, 2)
	ent(w(42), 152, 56, 0x201, 2)
	ent(w(41), 128, 56, 0x201, 2)
	ent(w(39), 99, 58, 0x300, 1)
	p = addViewSpritesD3(p)

	// missiles of one layer; (ax, ay) is the cell a missile reaches at its third step, (bx, by)
	// the cell after it. It ends at a wall or blocked terrain there.
	missiles := func(layer, ax, ay, bx, by int, mx, my func(i int) int) {
		for i := 0; i < 8; i++ {
			if gMissilePos[i] != 0 && int(gMissileLayers[gMissileSteps[i]/3]) == layer {
				m := int(gMissileSteps[i] % 3)
				if m == 2 && *gMonMissileActive == 0 {
					draw := false
					switch getCellWallsRegion(ax, ay, 7) {
					case 1, 3, 5, 6, 7, 8, 11:
					default:
						if getCellWallsRegion(bx, by, 0x70) != 0 {
							draw = *gPartyCellNoTerrain == 0
						} else {
							draw = getCellWallsRegion(ax, ay, int(gDirBlockBits[d])) == 0 || getCellWallsRegion(bx, by, 0x70) == 0
						}
					}
					if !draw {
						gMissilePos[i] = 0
						playSfx(9)
						continue
					}
				}
				p = dlPut(p, -1)
				p = dlPutL(p, *gPowIconFar)
				p++
				p = dlPut(p, mx(i))
				p = dlPut(p, my(i))
				p = dlPut(p, layer|i%2)
				p = dlPut(p, m)
				if *gMonMissileActive != 0 {
					gMissilePos[i]--
				} else {
					gMissilePos[i] = (gMissilePos[i] + 1) % 12
				}
			}
		}
	}
	vx := func(k int) int { return x + int(int8(gViewDx[d][k])) }
	vy := func(k int) int { return y + int(int8(gViewDy[d][k])) }
	// g_step_sound_toggle[i + n]: the missile positions of distances 3 and 2 run on from it
	sst := func(n int) func(int) int { return func(i int) int { return int(gStepSoundToggle[i+n]) } }
	u8s := func(a []uint8) func(int) int { return func(i int) int { return int(a[i]) } }

	missiles(0x300, vx(29), vy(29), vx(40), vy(40), sst(2), sst(10))
	ent(w(22), 8, 49, 0x100, 2)
	ent(w(24), 44, 49, 0x100, 2)
	ent(w(31), 170, 49, 0x101, 2)
	ent(w(29), 136, 49, 0x101, 2)
	ent(w(26), 82, 50, 0x200, 1)
	p = addViewSpritesD2(p)
	missiles(0x200, vx(14), vy(14), vx(27), vy(27), sst(18), sst(26))
	ent(w(11), 8, 40, 0, 2)
	ent(w(16), 146, 40, 1, 2)
	ent(w(13), 60, 38, 0x100, 1)
	p = addViewSpritesD1(p)
	missiles(0x100, vx(7), vy(7), vx(14), vy(14), u8s(gMissileD1X), u8s(gMissileD1Y))
	ent(w(5), 8, 24, 0, 0)
	ent(w(9), 169, 24, 1, 0)
	ent(w(6), 32, 24, 0, 1)
	ent(w(0), -23, 40, 2, 0)
	ent(w(4), 200, 40, 3, 0)
	if v := getCellWallsRegion(x, y, 7); v != 0 {
		p = dlPut(p, -1)
		p = dlPutL(p, Far(gMonD1l1L[v]))
		p++
		for _, n := range []int{8, 0x2f, 0, 0, 0xa8, 0x2f, 1, 0, 0x20, 0x26, 0, 1} {
			p = dlPut(p, n)
		}
	}
	p = addViewSpritesD0(p)
	missiles(0, x, y, vx(7), vy(7), u8s(gMissileD0X), u8s(gMissileD0Y))
	p = dlPut(p, -1)
	p = dlPut(p, 0)
	p = dlPut(p, 0)
	if *gMonD0c1 != 0 {
		memset(gViewCaption, 0, 0x96)
	}
	addUiGadgetSprites()
	sprintf(gTextBuf, "\x03d\x05%04X%s\x03m%u\x03d\x03l\x05%04X", adrDrawList, gViewCaption, *gDimLevel, adrBorderSprites)
	checkWallListSize(p)
	printWindowText(gTextBuf)
	alive := false
	for i := 0; i < 8; i++ {
		if gMissilePos[i] != 0 {
			alive = true
		}
	}
	if *gPowIconFar != 0 && !alive {
		freeAndNull(gPowIconFar)
		*gSpellAnimActive = 0
	}
	if *gGameMode == 0 {
		fadeScreen(1, 4)
		*gGameMode = 1
	}
}
