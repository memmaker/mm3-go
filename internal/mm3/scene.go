// Port of SCENE.C: refreshing the 3D view, the border gadget sprites, the map cells around the
// party (walls, flags, visited bits) and the wall scan and near draw-list entries of the maze view.
package mm3

func refreshView() {
	if *gOutdoors != 0 {
		refreshOutdoorView()
	} else {
		scanViewWalls()
		drawMazeView()
	}
	*gViewDrawn = 1
	if (*gMonD0c1 != 0 || *gMonD0c2 != 0 || *gMonD0c3 != 0) && (*gGameMode == 1 || *gGameMode == 5) && *gMonMissileActive == 0 && *gLightChanged == 0 {
		runCombat()
		if *gEventsDeferred != 0 {
			runEventScript()
		}
	}
}

// addUiGadgetSprites fills the border's draw list (g_border_sprites_m1 + 1 on).
func addUiGadgetSprites() {
	o := 0
	put := func(v int) { o++; gBorderSpritesM1[o] = int16(v) }
	putL := func(p uint32) { o++; setLe32(ds[:], adrBorderSpritesM1+2*o, p); o++ }
	put(-1)
	putL(uint32(*gGlobalIcons))
	put(8)
	put(8)
	put(0)
	put(1)
	put(100)
	put(0x7c)
	put(0)
	put(0xb)
	put(0x6e)
	put(0x83)
	put(0)
	if partyHasSkill(6) != 0 {
		put(int(*gPartyDir) + 0xf)
	} else {
		put(0xe)
	}
	put(-1)
	putL(uint32(*gGargoyleGfx))
	put(8)
	put(0x47)
	put(0)
	if *gLevitate != 0 {
		put(int(*gLevitateGemFrame))
	} else {
		put(0)
	}
	*gLevitateGemFrame = (*gLevitateGemFrame + 1) % 12
	put(-1)
	putL(uint32(*gGrabberGfx))
	put(0xc5)
	put(0x5d)
	put(0)
	if *gSecretDoorAhead != 0 && partyHasSkill(0x10) != 0 {
		put(int(*gSpotdoorGemFrame))
	} else {
		put(0)
	}
	*gSpotdoorGemFrame = (*gSpotdoorGemFrame + 1) % 12
	put(-1)
	putL(uint32(*gBatGfx))
	put(0x69)
	put(8)
	put(0)
	if *gMonstersMoved != 0 && partyHasSkill(0x11) != 0 {
		put(int(*gDangerGemFrame))
	} else {
		put(0)
	}
	*gDangerGemFrame = (*gDangerGemFrame + 1) % 12
	put(-1)
	putL(uint32(*gProtGemsGfx))
	put(3)
	put(4)
	put(0)
	put(b2i(*gFireResist == 0))
	put(0xde)
	put(4)
	put(0)
	put(2 + b2i(*gElecResist == 0))
	put(2)
	put(0x88)
	put(0)
	put(4 + b2i(*gColdResist == 0))
	put(0xdf)
	put(0x88)
	put(0)
	put(6 + b2i(*gPoisonResist == 0))
	put(-1)
	put(0)
	put(0)
}

// b2i is C's value of a comparison.
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func setPartyMapStart() {
	var xo, yo int
	switch *gStartMapIdx {
	case 0:
		xo, yo = 0, 0
	case 1:
		xo, yo = 0x10, 0
	case 2:
		xo, yo = 0, 0x10
	default:
		xo, yo = 0x10, 0x10
	}
	*gPartyX = uint8(int(gMaps[*gStartMapIdx].StartPos&0xf) + xo)
	*gPartyY = uint8(int(gMaps[*gStartMapIdx].StartPos>>4) + yo)
}

func findLoadedMap(v int) int {
	for i := 0; i < 4; i++ {
		if int(gQuadrantMaps[i]) == int(uint16(v)) {
			return i
		}
	}
	return 0x1111
}

// cellQuadrant is the common head of the cell accessors: (x, y) relative to quadrant 0 and the
// g_maps index of the map it is in; ok false where the C returns early (off the 32x32 area or
// in a map that is not loaded).
func cellQuadrant(x, y int) (int, int, int, bool) {
	idx := 0
	x += int(gQuadXOffsets[*gMapQuadrant])
	y += int(gQuadYOffsets[*gMapQuadrant])
	if x < 0 || y < 0 || x > 0x1f || y > 0x1f {
		return x, y, idx, false
	}
	if y > 0xf {
		if idx = findLoadedMap(int(gMaps[idx].North)); idx == 0x1111 {
			return x, y, idx, false
		}
	}
	if x > 0xf {
		if idx = findLoadedMap(int(gMaps[idx].East)); idx == 0x1111 {
			return x, y, idx, false
		}
	}
	return x, y, idx, true
}

// offMap is `g_outdoors ? 0 : 0x1111`.
func offMap() int {
	if *gOutdoors != 0 {
		return 0
	}
	return 0x1111
}

func markCellVisited(x, y int) {
	x, y, idx, ok := cellQuadrant(x, y)
	if !ok {
		return
	}
	setBit(gMaps[idx].Visited[:], (y&0xf)*16+(x&0xf), 1)
}

func isCellVisited(x, y int) int {
	x, y, idx, ok := cellQuadrant(x, y) // the C's (unsigned) > 0x1f test is the same range test
	if !ok {
		return 0
	}
	return testBit(gMaps[idx].Visited[:], (y&0xf)*16+(x&0xf))
}

func getCellWalls(x, y, mask int) int {
	x, y, idx, ok := cellQuadrant(x, y)
	if !ok {
		return offMap()
	}
	return int(int16(uint16(gMaps[idx].Cell[y&0xf][x&0xf].Get()) & uint16(mask)))
}

func getCellWallsRegion(x, y, mask int) int {
	x, y, idx, ok := cellQuadrant(x, y)
	if !ok {
		return offMap()
	}
	if *gOutdoors != 0 {
		switch gQuadrantMaps[*gMapQuadrant] {
		case 0x2d, 0x2e, 0x2f, 0x30:
			if gQuadrantMaps[idx] >= 0x31 && gQuadrantMaps[idx] <= 0x34 {
				return offMap()
			}
		case 0x31, 0x32, 0x33, 0x34:
			if gQuadrantMaps[idx] <= 0x30 {
				return offMap()
			}
		}
	}
	return int(int16(uint16(gMaps[idx].Cell[y&0xf][x&0xf].Get()) & uint16(mask)))
}

func getCellFlags(x, y, mask int) int {
	x, y, idx, ok := cellQuadrant(x, y)
	if !ok {
		return offMap()
	}
	return int(uint16(gMaps[idx].CellFlags[y&0xf][x&0xf]) & uint16(mask))
}

func setCellWall(x, y, k, v int) {
	idx := int(*gMapQuadrant)
	if y > 0xf {
		idx = findLoadedMap(int(gMaps[idx].North))
	}
	if x > 0xf {
		idx = findLoadedMap(int(gMaps[idx].East))
	}
	c := &gMaps[idx].Cell[y&0xf][x&0xf]
	c.Set(c.Get() & int(uint16(gViewWallMasks[8][k])))
	c.Set(c.Get() | v<<uint(gViewWallShifts[k][2]))
}

func clearViewWalls() {
	for _, p := range []*uint8{
		gWallD3l3,
		gWallD3l2,
		gWallD3l1,
		gWallD3r3,
		gWallD3r2,
		gWallD3r1,
		gWallD3c,
		gWallD3cLside,
		gWallD3l1Lside,
		gWallD3cRside,
		gWallD3r1Rside,
		gWallD2l2,
		gWallD2l1,
		gWallD2r2,
		gWallD2r1,
		gWallD2c,
		gWallD2l1Lside,
		gWallD2cLside,
		gWallD2r1Rside,
		gWallD2cRside,
		gWallD1l1,
		gWallD1r1,
		gWallD1c,
		gWallD1cLside,
		gWallD1cRside,
		gWallD0l1,
		gWallD0r1,
		gWallD0c,
		gWallD0cLside,
		gWallD0cRside,
		gWallD3l2Lside,
		gWallD3r2Rside,
		gWallD3l3Lside,
		gWallD3r3Rside,
		gWallD3l4,
		gWallD3r4,
		gWallD4cLside,
		gWallD4l1Lside,
		gWallD4l2Lside,
		gWallD4l3Lside,
		gWallD4cRside,
		gWallD4r1Rside,
		gWallD4r2Rside,
		gWallD4r3Rside,
		gWallD3l3T7,
		gWallD3l2T7,
		gWallD3l1T7,
		gWallD3r3T7,
		gWallD3r2T7,
		gWallD3r1T7,
		gWallD3cT7,
		gWallD3cLsideT7,
		gWallD3l1LsideT7,
		gWallD3cRsideT7,
		gWallD3r1RsideT7,
		gWallD2l2T7,
		gWallD2l1T7,
		gWallD2r2T7,
		gWallD2r1T7,
		gWallD2cT7,
		gWallD2l1LsideT7,
		gWallD2cLsideT7,
		gWallD2r1RsideT7,
		gWallD2cRsideT7,
		gWallD1l1T7,
		gWallD1r1T7,
		gWallD1cT7,
		gWallD1cLsideT7,
		gWallD1cRsideT7,
		gWallD0l1T7,
		gWallD0r1T7,
		gWallD0cT7,
		gWallD0cLsideT7,
		gWallD0cRsideT7,
		gWallD3l2LsideT7,
		gWallD3r2RsideT7,
		gWallD3l3LsideT7,
		gWallD3r3RsideT7,
		gWallD3l4T7,
		gWallD3r4T7,
		gWallD4cLsideT7,
		gWallD4l1LsideT7,
		gWallD4l2LsideT7,
		gWallD4l3LsideT7,
		gWallD4cRsideT7,
		gWallD4r1RsideT7,
		gWallD4r2RsideT7,
		gWallD4r3RsideT7,
		gWallD3l4Locked,
		gWallD3l3Locked,
		gWallD3l2Locked,
		gWallD3l1Locked,
		gWallD3r3Locked,
		gWallD3r4Locked,
		gWallD3r2Locked,
		gWallD3r1Locked,
		gWallD3cLocked,
		gWallD2l2Locked,
		gWallD2l1Locked,
		gWallD2r2Locked,
		gWallD2r1Locked,
		gWallD2cLocked,
		gWallD1l1Locked,
		gWallD1r1Locked,
		gWallD1cLocked,
		gWallD0l1Locked,
		gWallD0r1Locked,
		gWallD0cLocked,
		gWallD3l4Unlocked,
		gWallD3l3Unlocked,
		gWallD3l2Unlocked,
		gWallD3l1Unlocked,
		gWallD3r3Unlocked,
		gWallD3r4Unlocked,
		gWallD3r2Unlocked,
		gWallD3r1Unlocked,
		gWallD3cUnlocked,
		gWallD2l2Unlocked,
		gWallD2l1Unlocked,
		gWallD2r2Unlocked,
		gWallD2r1Unlocked,
		gWallD2cUnlocked,
		gWallD1l1Unlocked,
		gWallD1r1Unlocked,
		gWallD1cUnlocked,
		gWallD0l1Unlocked,
		gWallD0r1Unlocked,
		gWallD0cUnlocked,
		gWallD3l4Door,
		gWallD3l3Door,
		gWallD3l2Door,
		gWallD3l1Door,
		gWallD3r3Door,
		gWallD3r4Door,
		gWallD3r2Door,
		gWallD3r1Door,
		gWallD3cDoor,
		gWallD2l2Door,
		gWallD2l1Door,
		gWallD2r2Door,
		gWallD2r1Door,
		gWallD2cDoor,
		gWallD1l1Door,
		gWallD1r1Door,
		gWallD1cDoor,
		gWallD0l1Door,
		gWallD0r1Door,
		gWallD0cDoor,
		gWallD3l4Torch,
		gWallD3l3Torch,
		gWallD3l2Torch,
		gWallD3l1Torch,
		gWallD3r3Torch,
		gWallD3r4Torch,
		gWallD3r2Torch,
		gWallD3r1Torch,
		gWallD3cTorch,
		gWallD2l2Torch,
		gWallD2l1Torch,
		gWallD2r2Torch,
		gWallD2r1Torch,
		gWallD2cTorch,
		gWallD1l1Torch,
		gWallD1r1Torch,
		gWallD1cTorch,
		gWallD0l1Torch,
		gWallD0r1Torch,
		gWallD0cTorch,
		gWallD3l4Bashed,
		gWallD3l3Bashed,
		gWallD3l2Bashed,
		gWallD3l1Bashed,
		gWallD3r3Bashed,
		gWallD3r4Bashed,
		gWallD3r2Bashed,
		gWallD3r1Bashed,
		gWallD3cBashed,
		gWallD2l2Bashed,
		gWallD2l1Bashed,
		gWallD2r2Bashed,
		gWallD2r1Bashed,
		gWallD2cBashed,
		gWallD1l1Bashed,
		gWallD1r1Bashed,
		gWallD1cBashed,
		gWallD0l1Bashed,
		gWallD0r1Bashed,
		gWallD0cBashed,
		gSecretDoorAhead,
		gPartyCellFlag20} {
		*p = 0
	}
	*gSecretDoorAhead, *gPartyCellFlag20 = 0, 0
}

func updateDarkness() {
	if *gViewDirty == 0 {
		return
	}
	if getCellFlags(int(*gPartyX), int(*gPartyY), 8) != 0 && *gLightCount != 0 {
		*gLightCount--
	}
	if *gLightCount != 0 || gMaps[*gMapQuadrant].Lit != 0 {
		*gDimLevel = 0
	} else {
		*gDimLevel = 4
	}
}

func scanViewWalls() {
	k := 0
	px, py := int(*gPartyX), int(*gPartyY)
	fac := int(*gPartyDir)
	clearViewWalls()
	updateDarkness()
	cell := func(k, mask int) int {
		return getCellWallsRegion(int(int8(gViewDx[fac][k]))+px, int(int8(gViewDy[fac][k]))+py, mask)
	}
	switch cell(2, int(uint16(gViewWallMasks[fac][2]))) >> uint(gViewWallShifts[fac][2]) {
	case 1, 3:
		if cell(7, 0x7777) != 0x1111 {
			*gSecretDoorAhead = 1
		}
	}
	// wall is `get_cell_walls_region(view cell k, g_view_wall_masks[fac][k]) >> g_view_wall_shifts[fac][k++]`.
	wall := func() int {
		v := cell(k, int(uint16(gViewWallMasks[fac][k]))) >> uint(gViewWallShifts[fac][k])
		k++
		return v
	}
	switch wall() {
	case 3:
		*gWallD0l1++
		*gWallD0l1Torch = *gWallD0l1
	case 4:
		*gWallD0l1Door++
	case 5:
		*gWallD0l1Bashed++
	case 1:
		*gWallD0l1++
	case 2:
		*gWallD0l1++
		*gWallD0l1Locked = *gWallD0l1
	case 6:
		*gWallD0l1Unlocked++
	case 7:
		*gWallD0l1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD0cLside++
	case 7:
		*gWallD0cLsideT7++
	}
	switch wall() {
	case 3:
		*gWallD0c++
		*gWallD0cTorch = *gWallD0c
	case 4:
		*gWallD0cDoor++
	case 5:
		*gWallD0cBashed++
	case 1:
		*gWallD0c++
	case 6:
		*gWallD0cUnlocked++
	case 2:
		*gWallD0c++
		*gWallD0cLocked = *gWallD0c
	case 7:
		*gWallD0cT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD0cRside++
	case 7:
		*gWallD0cRsideT7++
	}
	switch wall() {
	case 3:
		*gWallD0r1++
		*gWallD0r1Torch = *gWallD0r1
	case 4:
		*gWallD0r1Door++
	case 5:
		*gWallD0r1Bashed++
	case 1:
		*gWallD0r1++
	case 6:
		*gWallD0r1Unlocked++
	case 2:
		*gWallD0r1++
		*gWallD0r1Locked = *gWallD0r1
	case 7:
		*gWallD0r1T7++
	}
	switch wall() {
	case 3:
		*gWallD1l1++
		*gWallD1l1Torch = *gWallD1l1
	case 4:
		*gWallD1l1Door++
	case 5:
		*gWallD1l1Bashed++
	case 1:
		*gWallD1l1++
	case 6:
		*gWallD1l1Unlocked++
	case 2:
		*gWallD1l1++
		*gWallD1l1Locked = *gWallD1l1
	case 7:
		*gWallD1l1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD1cLside++
	case 7:
		*gWallD1cLsideT7++
	}
	switch wall() {
	case 3:
		*gWallD1c++
		*gWallD1cTorch = *gWallD1c
	case 4:
		*gWallD1cDoor++
	case 5:
		*gWallD1cBashed++
	case 1:
		*gWallD1c++
	case 6:
		*gWallD1cUnlocked++
	case 2:
		*gWallD1c++
		*gWallD1cLocked = *gWallD1c
	case 7:
		*gWallD1cT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD1cRside++
	case 7:
		*gWallD1cRsideT7++
	}
	switch wall() {
	case 3:
		*gWallD1r1++
		*gWallD1r1Torch = *gWallD1r1
	case 4:
		*gWallD1r1Door++
	case 5:
		*gWallD1r1Bashed++
	case 1:
		*gWallD1r1++
	case 6:
		*gWallD1r1Unlocked++
	case 2:
		*gWallD1r1++
		*gWallD1r1Locked = *gWallD1r1
	case 7:
		*gWallD1r1T7++
	}
	switch wall() {
	case 3:
		*gWallD2l2++
		*gWallD2l2Torch = *gWallD2l2
	case 4:
		*gWallD2l2Door++
	case 5:
		*gWallD2l2Bashed++
	case 1:
		*gWallD2l2++
	case 6:
		*gWallD2l2Unlocked++
	case 2:
		*gWallD2l2++
		*gWallD2l2Locked = *gWallD2l2
	case 7:
		*gWallD2l2T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD2l1Lside++
	case 7:
		*gWallD2l1LsideT7++
	}
	switch wall() {
	case 3:
		*gWallD2l1++
		*gWallD2l1Torch = *gWallD2l1
	case 4:
		*gWallD2l1Door++
	case 5:
		*gWallD2l1Bashed++
	case 1:
		*gWallD2l1++
	case 6:
		*gWallD2l1Unlocked++
	case 2:
		*gWallD2l1++
		*gWallD2l1Locked = *gWallD2l1
	case 7:
		*gWallD2l1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD2cLside++
	case 7:
		*gWallD2cLsideT7++
	}
	switch wall() {
	case 3:
		*gWallD2c++
		*gWallD2cTorch = *gWallD2c
	case 4:
		*gWallD2cDoor++
	case 5:
		*gWallD2cBashed++
	case 1:
		*gWallD2c++
	case 6:
		*gWallD2cUnlocked++
	case 2:
		*gWallD2c++
		*gWallD2cLocked = *gWallD2c
	case 7:
		*gWallD2cT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD2cRside++
	case 7:
		*gWallD2cRsideT7++
	}
	switch wall() {
	case 3:
		*gWallD2r1++
		*gWallD2r1Torch = *gWallD2r1
	case 4:
		*gWallD2r1Door++
	case 5:
		*gWallD2r1Bashed++
	case 1:
		*gWallD2r1++
	case 6:
		*gWallD2r1Unlocked++
	case 2:
		*gWallD2r1++
		*gWallD2r1Locked = *gWallD2r1
	case 7:
		*gWallD2r1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD2r1Rside++
	case 7:
		*gWallD2r1RsideT7++
	}
	switch wall() {
	case 3:
		*gWallD2r2++
		*gWallD2r2Torch = *gWallD2r2
	case 4:
		*gWallD2r2Door++
	case 5:
		*gWallD2r2Bashed++
	case 1:
		*gWallD2r2++
	case 6:
		*gWallD2r2Unlocked++
	case 2:
		*gWallD2r2++
		*gWallD2r2Locked = *gWallD2r2
	case 7:
		*gWallD2r2T7++
	}
	switch wall() {
	case 3:
		*gWallD3l4++
		*gWallD3l4Torch = *gWallD3l4
	case 4:
		*gWallD3l4Door++
	case 5:
		*gWallD3l4Bashed++
	case 1:
		*gWallD3l4++
	case 6:
		*gWallD3l4Unlocked++
	case 2:
		*gWallD3l4++
		*gWallD3l4Locked = *gWallD3l4
	case 7:
		*gWallD3l4T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3l3Lside++
	case 7:
		*gWallD3l3LsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3l3++
		*gWallD3l3Torch = *gWallD3l3
	case 4:
		*gWallD3l3Door++
	case 5:
		*gWallD3l3Bashed++
	case 1:
		*gWallD3l3++
	case 6:
		*gWallD3l3Unlocked++
	case 2:
		*gWallD3l3++
		*gWallD3l3Locked = *gWallD3l3
	case 7:
		*gWallD3l3T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3l2Lside++
	case 7:
		*gWallD3l2LsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3l2++
		*gWallD3l2Torch = *gWallD3l2
	case 4:
		*gWallD3l2Door++
	case 5:
		*gWallD3l2Bashed++
	case 1:
		*gWallD3l2++
	case 6:
		*gWallD3l2Unlocked++
	case 2:
		*gWallD3l2++
		*gWallD3l2Locked = *gWallD3l2
	case 7:
		*gWallD3l2T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3l1Lside++
	case 7:
		*gWallD3l1LsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3l1++
		*gWallD3l1Torch = *gWallD3l1
	case 4:
		*gWallD3l1Door++
	case 5:
		*gWallD3l1Bashed++
	case 1:
		*gWallD3l1++
	case 6:
		*gWallD3l1Unlocked++
	case 2:
		*gWallD3l1++
		*gWallD3l1Locked = *gWallD3l1
	case 7:
		*gWallD3l1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3cLside++
	case 7:
		*gWallD3cLsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3c++
		*gWallD3cTorch = *gWallD3c
	case 4:
		*gWallD3cDoor++
	case 5:
		*gWallD3cBashed++
	case 1:
		*gWallD3c++
	case 6:
		*gWallD3cUnlocked++
	case 2:
		*gWallD3c++
		*gWallD3cLocked = *gWallD3c
	case 7:
		*gWallD3cT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3cRside++
	case 7:
		*gWallD3cRsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3r1++
		*gWallD3r1Torch = *gWallD3r1
	case 4:
		*gWallD3r1Door++
	case 5:
		*gWallD3r1Bashed++
	case 1:
		*gWallD3r1++
	case 6:
		*gWallD3r1Unlocked++
	case 2:
		*gWallD3r1++
		*gWallD3r1Locked = *gWallD3r1
	case 7:
		*gWallD3r1T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3r1Rside++
	case 7:
		*gWallD3r1RsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3r2++
		*gWallD3r2Torch = *gWallD3r2
	case 4:
		*gWallD3r2Door++
	case 5:
		*gWallD3r2Bashed++
	case 1:
		*gWallD3r2++
	case 6:
		*gWallD3r2Unlocked++
	case 2:
		*gWallD3r2++
		*gWallD3r2Locked = *gWallD3r2
	case 7:
		*gWallD3r2T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3r2Rside++
	case 7:
		*gWallD3r2RsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3r3++
		*gWallD3r3Torch = *gWallD3r3
	case 4:
		*gWallD3r3Door++
	case 5:
		*gWallD3r3Bashed++
	case 1:
		*gWallD3r3++
	case 6:
		*gWallD3r3Unlocked++
	case 2:
		*gWallD3r3++
		*gWallD3r3Locked = *gWallD3r3
	case 7:
		*gWallD3r3T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD3r3Rside++
	case 7:
		*gWallD3r3RsideT7++
	}
	switch wall() {
	case 3:
		*gWallD3r4++
		*gWallD3r4Torch = *gWallD3r4
	case 4:
		*gWallD3r4Door++
	case 5:
		*gWallD3r4Bashed++
	case 1:
		*gWallD3r4++
	case 6:
		*gWallD3r4Unlocked++
	case 2:
		*gWallD3r4++
		*gWallD3r4Locked = *gWallD3r4
	case 7:
		*gWallD3r4T7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4l3Lside++
	case 7:
		*gWallD4l3LsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4l2Lside++
	case 7:
		*gWallD4l2LsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4l1Lside++
	case 7:
		*gWallD4l1LsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4cLside++
	case 7:
		*gWallD4cLsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4cRside++
	case 7:
		*gWallD4cRsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4r1Rside++
	case 7:
		*gWallD4r1RsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4r2Rside++
	case 7:
		*gWallD4r2RsideT7++
	}
	switch wall() {
	case 1, 2, 3, 4, 5, 6:
		*gWallD4r3Rside++
	case 7:
		*gWallD4r3RsideT7++
	}
	*gPartyCellFlag20 = gMaps[0].CellFlags[py][px] & 0x20
	scanViewObjects()
	if *gMonsterScanDelay != 0 && (*gGameMode == 1 || *gGameMode == 2) && *gMonMissileActive == 0 && *gLightChanged == 0 {
		if *gMonsterScanDelay--; *gMonsterScanDelay == 0 {
			moveMonsters()
		}
	}
	scanViewMonsters()
}

// addViewSpritesNear appends the near walls (distance 1 and 0), their sprites and missiles to
// the draw list at index o and terminates it.
func addViewSpritesNear(o int) {
	put4 := func(a, b, c, d int) {
		o = dlPut(o, a)
		o = dlPut(o, b)
		o = dlPut(o, c)
		o = dlPut(o, d)
	}
	flip := int(*gViewFlip)
	nflip := b2i(flip == 0)
	tf := int(*gTorchFrame) + 1
	o = dlPut(o, -1)
	o = dlPutL(o, Far(*gWallGfx2))
	o++
	if !(*gWallD0c != 0 && *gWallD0l1 != 0) && !(*gWallD0c != 0 && *gWallD0cLside != 0) && *gWallD1cLside&*gWallD0l1 == 0 && !(*gWallD1cLside != 0 && *gWallD0cLside != 0) {
		switch {
		case *gWallD1l1Locked != 0:
			put4(-40, 40, 2, 6)
		case *gWallD1l1T7 != 0:
			put4(-40, 40, 2|flip, 10)
		case *gWallD1l1Bashed != 0:
			put4(-40, 40, 2, 9)
		case *gWallD1l1Unlocked != 0:
			put4(-40, 40, 2, 7)
		case *gWallD1l1Door != 0:
			put4(-40, 40, 2, 8)
		case *gWallD1l1Torch != 0:
			put4(-40, 40, 2, tf)
		case *gWallD1l1 != 0:
			put4(-40, 40, 2, 0)
		}
	}
	if !(*gWallD0c != 0 && *gWallD0r1 != 0) && !(*gWallD0c != 0 && *gWallD0cRside != 0) && *gWallD1cRside&*gWallD0r1 == 0 && !(*gWallD1cRside != 0 && *gWallD0cRside != 0) {
		switch {
		case *gWallD1r1Locked != 0:
			put4(168, 40, 3, 6)
		case *gWallD1r1T7 != 0:
			put4(168, 40, 2|flip, 10)
		case *gWallD1r1Bashed != 0:
			put4(168, 40, 3, 9)
		case *gWallD1r1Unlocked != 0:
			put4(168, 40, 3, 7)
		case *gWallD1r1Door != 0:
			put4(168, 40, 3, 8)
		case *gWallD1r1Torch != 0:
			put4(168, 40, 3, tf)
		case *gWallD1r1 != 0:
			put4(168, 40, 3, 0)
		}
	}
	if *gWallD0c == 0 {
		switch {
		case *gWallD1cLocked != 0:
			put4(64, 40, 0, 6)
		case *gWallD1cT7 != 0:
			put4(64, 40, flip, 10)
		case *gWallD1cBashed != 0:
			put4(64, 40, 0, 9)
		case *gWallD1cUnlocked != 0:
			put4(64, 40, 0, 7)
		case *gWallD1cDoor != 0:
			put4(64, 40, 0, 8)
		case *gWallD1cTorch != 0:
			put4(64, 40, 0, tf)
		case *gWallD1c != 0:
			put4(64, 40, 0, 0)
		}
	}
	if *gWallD0c == 0 {
		if *gWallD1cLside != 0 {
			put4(32, 24, 0, int(gSidewallD1Frames[flip]))
		} else if *gWallD1cLsideT7 != 0 {
			put4(32, 24, 0, int(gSidewallD1T7Frames[flip]))
		}
	}
	if *gWallD0c == 0 {
		if *gWallD1cRside != 0 {
			put4(168, 24, 1, int(gSidewallD1Frames[nflip]))
		} else if *gWallD1cRsideT7 != 0 {
			put4(168, 24, 1, int(gSidewallD1T7Frames[nflip]))
		}
	}
	o = addViewSpritesD1(o)
	o = addNearMissiles(o, 0x100, *gWallD1c != 0 || *gWallD1cDoor != 0, gMissileD1X, gMissileD1Y)
	o = dlPut(o, -1)
	o = dlPutL(o, Far(*gOutdoorObjGfx))
	o++
	if *gWallD0cLside == 0 {
		if *gWallD0l1Locked != 0 || *gWallD0l1 != 0 || *gWallD0l1Torch != 0 || *gWallD0l1Unlocked != 0 || *gWallD0l1Bashed != 0 || *gWallD0l1Door != 0 {
			put4(-136, 24, 2, 0)
		} else if *gWallD0l1T7 != 0 {
			put4(-136, 24, 2|flip, 10)
		}
	}
	if *gWallD0cRside == 0 {
		if *gWallD0r1Locked != 0 || *gWallD0r1 != 0 || *gWallD0r1Torch != 0 || *gWallD0r1Unlocked != 0 || *gWallD0r1Bashed != 0 || *gWallD0r1Door != 0 {
			put4(200, 24, 3, 0)
		} else if *gWallD0r1T7 != 0 {
			put4(200, 24, 2|flip, 10)
		}
	}
	switch {
	case *gWallD0cLocked != 0:
		put4(32, 24, 0, 6)
	case *gWallD0cT7 != 0:
		put4(32, 24, flip, 10)
	case *gWallD0cTorch != 0:
		put4(32, 24, flip, tf)
	case *gWallD0cUnlocked != 0:
		put4(32, 24, flip, 7)
	case *gWallD0cDoor != 0:
		put4(32, 24, flip, 8)
	case *gWallD0cBashed != 0:
		put4(32, 24, 0, 9)
	case *gWallD0c != 0:
		put4(32, 24, flip, 0)
	}
	if *gWallD0cLside != 0 {
		put4(8, 12, 0, int(gSidewallD0Frames[flip]))
	} else if *gWallD0cLsideT7 != 0 {
		put4(8, 12, 0, int(gSidewallD0T7Frames[flip]))
	}
	if *gWallD0cRside != 0 {
		put4(200, 12, 1, int(gSidewallD0Frames[nflip]))
	} else if *gWallD0cRsideT7 != 0 {
		put4(200, 12, 1, int(gSidewallD0T7Frames[nflip]))
	}
	o = addViewSpritesD0(o)
	o = addNearMissiles(o, 0, *gWallD0c != 0 || *gWallD0cDoor != 0, gMissileD0X, gMissileD0Y)
	o = dlPut(o, -1)
	o = dlPut(o, 0)
	dlPut(o, 0)
}

// addNearMissiles is add_view_sprites_near's missile loop for one layer (0x100: distance 1, 0:
// distance 0): a missile hitting the wall in front at its third step ends there.
func addNearMissiles(o, layer int, wall bool, xs, ys []uint8) int {
	for i := 0; i < 8; i++ {
		if gMissilePos[i] != 0 && int(gMissileLayers[gMissileSteps[i]/3]) == layer {
			r := int(gMissileSteps[i] % 3)
			if wall && r == 2 && *gMonMissileActive == 0 {
				gMissilePos[i] = 0
				playSfx(9)
			} else {
				o = dlPut(o, -1)
				o = dlPutL(o, *gPowIconFar)
				o++
				o = dlPut(o, int(xs[i]))
				o = dlPut(o, int(ys[i]))
				o = dlPut(o, layer|i%2)
				o = dlPut(o, r)
				if *gMonMissileActive != 0 {
					gMissilePos[i]--
				} else {
					gMissilePos[i] = (gMissilePos[i] + 1) % 12
				}
			}
		}
	}
	return o
}
