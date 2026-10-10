// Port of ITEMNAME.C (root segment 0e40, named for its place in the link map): drawMazeView,
// the per-frame animation tick and the draw list of the 3D view's far walls (distance 3 and 2),
// then the sprites of view.go/scene.go and the window text that draws it all.
package mm3

// w is a wall flag's C truth value.
func w(p *uint8) bool { return *p != 0 }

func drawMazeView() {
	var o, i, f int
	for i = 0; i < 8; i++ {
		gMissileSteps[i] = gMissilePos[i] - 1
	}
	*gHoverPhase1 = (*gHoverPhase1 + 1) % 18
	*gHoverPhase2 = (*gHoverPhase2 + 1) % 18
	*gHoverPhase3 = (*gHoverPhase3 + 1) % 18
	*gTorchFrame = (*gTorchFrame + 1) % 3
	for i = 0; i < int(*gMonCount); i++ {
		if gMonStatus[i] == 0 { // g_mon_frames is u16[] here
			gMonFrames[i] = int16((uint16(gMonFrames[i]) + 1) % uint16(gMonFrameCounts[gMonType[i]]))
		}
	}
	for i = 0; i < int(*gObjCount); i++ {
		if *gGameMode != 7 || int(*gObjD0c)-1 != i {
			a := &gMapObjectsAnim[i]
			a.Frame.Set((a.Frame.Get() + 1) % int(gObjFrameCounts[a.Type.Get()]))
		}
	}
	gDrawList[0], gDrawList[7] = -1, -1
	gDrawList[3], gDrawList[4], gDrawList[10] = 8, 8, 8
	gDrawList[1], gDrawList[2], gDrawList[6], gDrawList[8], gDrawList[9] = 0, 0, 0, 0, 0
	gDrawList[13] = 0x1d
	gDrawList[11] = 0x43
	o = 13
	gDrawList[5] = int16(*gViewBgFlip)
	gDrawList[12] = int16(*gViewFlip)
	setLe32(ds[:], adrDrawList+2, uint32(*gSkyGfx))
	setLe32(ds[:], adrDrawList+16, uint32(*gWallGfx4))
	if !w(gWallD0c) && !w(gWallD1l1) && !w(gWallD1cLside) && !w(gWallD2l1) && !w(gWallD3l1Lside) && !w(gWallD3l2Lside) {
		if w(gWallD4l3Lside) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 13)
		} else if w(gWallD4l3LsideT7) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 28)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2cLside) && !w(gWallD2l1) && !w(gWallD3l1Lside) {
		if w(gWallD4l2Lside) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 11)
		} else if w(gWallD4l2LsideT7) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 26)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) && !w(gWallD3l1) && !w(gWallD3cLside) {
		if w(gWallD4l1Lside) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 9)
		} else if w(gWallD4l1LsideT7) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 24)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) && !w(gWallD3c) {
		if w(gWallD4cLside) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(gClassHpPerLevel[int(*gViewFlip)+10]))
		} else if w(gWallD4cLsideT7) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(gSidewallD1T7Frames[int(*gViewFlip)+6]))
		}
	}
	if !w(gWallD0c) && !w(gWallD1r1) && !w(gWallD1cRside) && !w(gWallD2r1) && !w(gWallD3r1Rside) && !w(gWallD3r2Rside) {
		if w(gWallD4r3Rside) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 13)
		} else if w(gWallD4r3RsideT7) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 28)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2cRside) && !w(gWallD2r1) && !w(gWallD3r1Rside) {
		if w(gWallD4r2Rside) {
			o = dlPut(o, 144)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 11)
		} else if w(gWallD4r2RsideT7) {
			o = dlPut(o, 144)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 26)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) && !w(gWallD3r1) && !w(gWallD3cRside) {
		if w(gWallD4r1Rside) {
			o = dlPut(o, 131)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 9)
		} else if w(gWallD4r1RsideT7) {
			o = dlPut(o, 131)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, 24)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) && !w(gWallD3c) {
		if w(gWallD4cRside) {
			o = dlPut(o, 120)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, int(gClassHpPerLevel[b2i(*gViewFlip == 0)+10]))
		} else if w(gWallD4cRsideT7) {
			o = dlPut(o, 120)
			o = dlPut(o, 60)
			o = dlPut(o, 1)
			o = dlPut(o, int(gSidewallD1T7Frames[b2i(*gViewFlip == 0)+6]))
		}
	}
	if !w(gWallD0l1) && !w(gWallD0cLside) && !w(gWallD1l1) && !w(gWallD2l2) && !w(gWallD2l1Lside) && !w(gWallD3l2Lside) && !w(gWallD3l3Lside) {
		if w(gWallD3l4Locked) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3l4T7) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 18)
		} else if w(gWallD3l4Torch) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3l4Unlocked) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3l4Bashed) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3l4Door) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3l4) {
			o = dlPut(o, 8)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0r1) && !w(gWallD0cRside) && !w(gWallD1r1) && !w(gWallD2r2) && !w(gWallD2r1Rside) && !w(gWallD3r2Rside) && !w(gWallD3r3Rside) {
		if w(gWallD3r4Locked) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3r4T7) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3r4Torch) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3r4Unlocked) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3r4Bashed) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3r4Door) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3r4) {
			o = dlPut(o, 200)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0l1) && !w(gWallD0cLside) && !w(gWallD1l1) && !w(gWallD2l2) && !w(gWallD2l1Lside) && !w(gWallD3l2Lside) {
		if w(gWallD3l3Lside) {
			o = dlPut(o, 8)
			o = dlPut(o, 58)
			o = dlPut(o, 0)
			o = dlPut(o, 12)
		} else if w(gWallD3l3LsideT7) {
			o = dlPut(o, 8)
			o = dlPut(o, 58)
			o = dlPut(o, 0)
			o = dlPut(o, 27)
		}
	}
	if !w(gWallD0r1) && !w(gWallD0cRside) && !w(gWallD1r1) && !w(gWallD2r2) && !w(gWallD2r1Rside) && !w(gWallD3r2Rside) {
		if w(gWallD3r3Rside) {
			o = dlPut(o, 200)
			o = dlPut(o, 58)
			o = dlPut(o, 1)
			o = dlPut(o, 12)
		} else if w(gWallD3r3RsideT7) {
			o = dlPut(o, 200)
			o = dlPut(o, 58)
			o = dlPut(o, 1)
			o = dlPut(o, 27)
		}
	}
	if !(w(gWallD0cLside) && w(gWallD0c)) && !(w(gWallD0cLside) && w(gWallD2l1)) && !(w(gWallD0cLside) && w(gWallD1cLside)) && !(w(gWallD0cLside) && w(gWallD3l1Lside)) && !(w(gWallD0l1) && w(gWallD0c)) && !(w(gWallD0l1) && w(gWallD2l1)) && !(w(gWallD0l1) && w(gWallD1cLside)) && !(w(gWallD0l1) && w(gWallD3l1Lside)) && !(w(gWallD2l2) && w(gWallD0c)) && !(w(gWallD2l2) && w(gWallD2l1)) && !(w(gWallD2l2) && w(gWallD1cLside)) && !(w(gWallD2l2) && w(gWallD3l1Lside)) && !(w(gWallD2cLside) && w(gWallD0c)) && !(w(gWallD2cLside) && w(gWallD2l1)) && !(w(gWallD2cLside) && w(gWallD1cLside)) && !(w(gWallD2cLside) && w(gWallD3l1Lside)) && !w(gWallD1l1) {
		if w(gWallD3l2Lside) {
			o = dlPut(o, 8)
			o = dlPut(o, 55)
			o = dlPut(o, 0)
			o = dlPut(o, 10)
		} else if w(gWallD3l2LsideT7) {
			o = dlPut(o, 8)
			o = dlPut(o, 55)
			o = dlPut(o, 0)
			o = dlPut(o, 25)
		}
	}
	if !(w(gWallD0cRside) && w(gWallD0c)) && !(w(gWallD0cRside) && w(gWallD2r1)) && !(w(gWallD0cRside) && w(gWallD1cRside)) && !(w(gWallD0cRside) && w(gWallD3r1Rside)) && !(w(gWallD0r1) && w(gWallD0c)) && !(w(gWallD0r1) && w(gWallD2r1)) && !(w(gWallD0r1) && w(gWallD1cRside)) && !(w(gWallD0r1) && w(gWallD3r1Rside)) && !(w(gWallD2r2) && w(gWallD0c)) && !(w(gWallD2r2) && w(gWallD2r1)) && !(w(gWallD2r2) && w(gWallD1cRside)) && !(w(gWallD2r2) && w(gWallD3r1Rside)) && !(w(gWallD2cRside) && w(gWallD0c)) && !(w(gWallD2cRside) && w(gWallD2r1)) && !(w(gWallD2cRside) && w(gWallD1cRside)) && !(w(gWallD2cRside) && w(gWallD3r1Rside)) && !w(gWallD1r1) {
		if w(gWallD3r2Rside) {
			o = dlPut(o, 176)
			o = dlPut(o, 55)
			o = dlPut(o, 1)
			o = dlPut(o, 10)
		} else if w(gWallD3r2RsideT7) {
			o = dlPut(o, 176)
			o = dlPut(o, 55)
			o = dlPut(o, 1)
			o = dlPut(o, 25)
		}
	}
	if !w(gWallD0c) && !w(gWallD1l1) && !w(gWallD2l1) && !w(gWallD1cLside) && !w(gWallD3l1Lside) && !w(gWallD3l2Lside) {
		if w(gWallD3l3Locked) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3l3T7) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3l3Torch) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3l3Unlocked) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3l3Bashed) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3l3Door) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3l3) {
			o = dlPut(o, 32)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1l1)) && !(w(gWallD1c) && w(gWallD1cLside)) && !(w(gWallD1l1) && w(gWallD2cLside)) && !(w(gWallD1cLside) && w(gWallD2cLside)) && !w(gWallD2l1) && !w(gWallD3l1Lside) {
		if w(gWallD3l2Locked) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3l2T7) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3l2Torch) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3l2Unlocked) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3l2Bashed) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3l2Door) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3l2) {
			o = dlPut(o, 56)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !(w(gWallD2c) && w(gWallD2cLside)) && !(w(gWallD2c) && w(gWallD2l1)) && !(w(gWallD2l1) && w(gWallD3cLside)) && !(w(gWallD2cLside) && w(gWallD3cLside)) {
		if w(gWallD3l1Locked) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3l1T7) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3l1Torch) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3l1Unlocked) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3l1Bashed) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3l1Door) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3l1) {
			o = dlPut(o, 80)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !w(gWallD1r1) && !w(gWallD2r1) && !w(gWallD1cRside) && !w(gWallD3r1Rside) && !w(gWallD3r2Rside) {
		if w(gWallD3r3Locked) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3r3T7) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3r3Torch) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3r3Unlocked) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3r3Bashed) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3r3Door) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3r3) {
			o = dlPut(o, 176)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1r1)) && !(w(gWallD1c) && w(gWallD1cRside)) && !(w(gWallD1r1) && w(gWallD2cRside)) && !(w(gWallD1cRside) && w(gWallD2cRside)) && !w(gWallD2r1) && !w(gWallD3r1Rside) {
		if w(gWallD3r2Locked) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3r2T7) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3r2Torch) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3r2Unlocked) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3r2Bashed) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3r2Door) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3r2) {
			o = dlPut(o, 152)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !(w(gWallD2c) && w(gWallD2cRside)) && !(w(gWallD2c) && w(gWallD2r1)) && !(w(gWallD2r1) && w(gWallD3cRside)) && !(w(gWallD2cRside) && w(gWallD3cRside)) {
		if w(gWallD3r1Locked) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3r1T7) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3r1Torch) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3r1Unlocked) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3r1Bashed) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3r1Door) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3r1) {
			o = dlPut(o, 128)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) {
		if w(gWallD3cLocked) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		} else if w(gWallD3cT7) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 18)
		} else if w(gWallD3cTorch) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD3cUnlocked) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 15)
		} else if w(gWallD3cBashed) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 17)
		} else if w(gWallD3cDoor) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 16)
		} else if w(gWallD3c) {
			o = dlPut(o, 104)
			o = dlPut(o, 60)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) {
		if w(gWallD3cLside) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, int(gClassHpPerLevel[int(*gViewFlip)+12]))
		} else if w(gWallD3cLsideT7) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, int(gSidewallD1T7Frames[int(*gViewFlip)+4]))
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1cLside)) && !(w(gWallD1c) && w(gWallD1l1)) && !(w(gWallD2cLside) && w(gWallD1cLside)) && !(w(gWallD2cLside) && w(gWallD1l1)) && !w(gWallD2l1) {
		if w(gWallD3l1Lside) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 8)
		} else if w(gWallD3l1LsideT7) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 23)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) && !w(gWallD2c) {
		if w(gWallD3cRside) {
			o = dlPut(o, 128)
			o = dlPut(o, 52)
			o = dlPut(o, 1)
			o = dlPut(o, int(gClassHpPerLevel[b2i(*gViewFlip == 0)+12]))
		} else if w(gWallD3cRsideT7) {
			o = dlPut(o, 128)
			o = dlPut(o, 52)
			o = dlPut(o, 1)
			o = dlPut(o, int(gSidewallD1T7Frames[b2i(*gViewFlip == 0)+4]))
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1cRside)) && !(w(gWallD1c) && w(gWallD1r1)) && !(w(gWallD2cRside) && w(gWallD1cRside)) && !(w(gWallD2cRside) && w(gWallD1r1)) && !w(gWallD2r1) {
		if w(gWallD3r1Rside) {
			o = dlPut(o, 152)
			o = dlPut(o, 52)
			o = dlPut(o, 1)
			o = dlPut(o, 8)
		} else if w(gWallD3r1RsideT7) {
			o = dlPut(o, 152)
			o = dlPut(o, 52)
			o = dlPut(o, 1)
			o = dlPut(o, 23)
		}
	}
	o = addViewSpritesD3(o)
	o = addNearMissiles(o, 0x300, w(gWallD3c) || w(gWallD3cDoor), gStepSoundToggle[2:], gStepSoundToggle[10:])
	o = dlPut(o, -1)
	o = dlPutL(o, Far(*gWallGfx3))
	o++
	if !w(gWallD0l1) && !w(gWallD0cLside) && !w(gWallD1l1) && !w(gWallD2l1Lside) {
		if w(gWallD2l2Locked) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 7)
		} else if w(gWallD2l2T7) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 11)
		} else if w(gWallD2l2Door) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 9)
		} else if w(gWallD2l2Torch) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD2l2Unlocked) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 8)
		} else if w(gWallD2l2Bashed) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 10)
		} else if w(gWallD2l2) {
			o = dlPut(o, -24)
			o = dlPut(o, 52)
			o = dlPut(o, 2)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1cLside)) && !(w(gWallD1c) && w(gWallD1l1)) && !(w(gWallD1cLside) && w(gWallD2cLside)) && !(w(gWallD1l1) && w(gWallD2cLside)) {
		if w(gWallD2l1Locked) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 7)
		} else if w(gWallD2l1T7) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 11)
		} else if w(gWallD2l1Bashed) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 10)
		} else if w(gWallD2l1Unlocked) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 8)
		} else if w(gWallD2l1Torch) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD2l1Door) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 9)
		} else if w(gWallD2l1) {
			o = dlPut(o, 32)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0r1) && !w(gWallD0cRside) && !w(gWallD1r1) && !w(gWallD2r1Rside) {
		if w(gWallD2r2Locked) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, 7)
		} else if w(gWallD2r2T7) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 2|int(*gViewFlip))
			o = dlPut(o, 11)
		} else if w(gWallD2r2Door) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, 9)
		} else if w(gWallD2r2Torch) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD2r2Unlocked) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, 8)
		} else if w(gWallD2r2Bashed) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, 10)
		} else if w(gWallD2r2) {
			o = dlPut(o, 200)
			o = dlPut(o, 52)
			o = dlPut(o, 3)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0c) && !(w(gWallD1c) && w(gWallD1cRside)) && !(w(gWallD1c) && w(gWallD1r1)) && !(w(gWallD1cRside) && w(gWallD2cRside)) && !(w(gWallD1r1) && w(gWallD2cRside)) {
		if w(gWallD2r1Locked) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 7)
		} else if w(gWallD2r1T7) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 11)
		} else if w(gWallD2r1Door) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 9)
		} else if w(gWallD2r1Torch) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD2r1Unlocked) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 8)
		} else if w(gWallD2r1Bashed) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 10)
		} else if w(gWallD2r1) {
			o = dlPut(o, 144)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD1c) && !w(gWallD0c) {
		if w(gWallD2cLocked) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 7)
		} else if w(gWallD2cT7) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, int(*gViewFlip))
			o = dlPut(o, 11)
		} else if w(gWallD2cDoor) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 9)
		} else if w(gWallD2cTorch) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, int(*gTorchFrame)+1)
		} else if w(gWallD2cUnlocked) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 8)
		} else if w(gWallD2cBashed) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 10)
		} else if w(gWallD2c) {
			o = dlPut(o, 88)
			o = dlPut(o, 52)
			o = dlPut(o, 0)
			o = dlPut(o, 0)
		}
	}
	if !w(gWallD0l1) && !w(gWallD0cLside) && !w(gWallD1l1) {
		if w(gWallD2l1Lside) {
			o = dlPut(o, 8)
			o = dlPut(o, 48)
			o = dlPut(o, 0)
			o = dlPut(o, 6)
		} else if w(gWallD2l1LsideT7) {
			o = dlPut(o, 8)
			o = dlPut(o, 48)
			o = dlPut(o, 0)
			o = dlPut(o, 14)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) {
		if w(gWallD2cLside) {
			o = dlPut(o, 64)
			o = dlPut(o, 40)
			o = dlPut(o, 0)
			o = dlPut(o, int(gClassHpPerLevel[int(*gViewFlip)+14]))
		} else if w(gWallD2cLsideT7) {
			o = dlPut(o, 64)
			o = dlPut(o, 40)
			o = dlPut(o, 0)
			o = dlPut(o, int(gSidewallD1T7Frames[int(*gViewFlip)+2]))
		}
	}
	if !w(gWallD0r1) && !w(gWallD0cRside) && !w(gWallD1r1) {
		if w(gWallD2r1Rside) {
			o = dlPut(o, 200)
			o = dlPut(o, 48)
			o = dlPut(o, 1)
			o = dlPut(o, 6)
		} else if w(gWallD2r1RsideT7) {
			o = dlPut(o, 200)
			o = dlPut(o, 48)
			o = dlPut(o, 1)
			o = dlPut(o, 14)
		}
	}
	if !w(gWallD0c) && !w(gWallD1c) {
		if w(gWallD2cRside) {
			o = dlPut(o, 144)
			o = dlPut(o, 40)
			o = dlPut(o, 1)
			o = dlPut(o, int(gClassHpPerLevel[b2i(*gViewFlip == 0)+14]))
		} else if w(gWallD2cRsideT7) {
			o = dlPut(o, 144)
			o = dlPut(o, 40)
			o = dlPut(o, 1)
			o = dlPut(o, int(gSidewallD1T7Frames[b2i(*gViewFlip == 0)+2]))
		}
	}
	o = addViewSpritesD2(o)
	o = addNearMissiles(o, 0x200, w(gWallD2c) || w(gWallD2cDoor), gStepSoundToggle[18:], gStepSoundToggle[26:])
	if *gMonD0c1 != 0 {
		memset(gViewCaption, 0, 150)
	}
	addViewSpritesNear(o)
	addUiGadgetSprites()
	sprintf(gTextBuf, "\x03d\x05%04X%s\x03m%u\x03d\x03l\x05%04X", adrDrawList, gViewCaption, int(*gDimLevel), adrBorderSprites)
	checkWallListSize(o)
	printWindowText(gTextBuf)
	for i, f = 0, 0; i < 8; i++ {
		if gMissilePos[i] != 0 {
			f = 1
		}
	}
	if *gPowIconFar != 0 && f == 0 {
		freeAndNull(gPowIconFar)
		*gSpellAnimActive = 0
	}
	if *gGameMode == 0 {
		fadeScreen(1, 4)
		*gGameMode = 1
	}
}
