// Port of EVENTS.C: character damage, the portrait highlight, condition stat modifiers, moving
// between outdoor quadrants, the maze event script interpreter, sound effects, draw-list size
// checks and monster movement.
package mm3

func charTakeDamage(p *Chr, dmg int) {
	p.HP.Set(p.HP.Get() - dmg)
	flag := p.HP.Get() <= -10
	if p.HP.Get() < 1 {
		d := int(int16(p.HP.Get() + calcMaxHp(p)))
		if d < 1 {
			p.Cond[0xd] = 1
			flag = true
			if p.HP.Get() > 0 {
				p.HP.Set(0)
			}
		} else {
			p.Cond[0xc] = 1
		}
		if flag {
			for d = 0; d < 0x12; d++ {
				if p.ItemID[d] > 0x21 && p.ItemID[d] < 0x2a && p.ItemEq[d] != 0 {
					p.ItemFlags[d] |= 0x80
					break
				}
			}
		}
	}
}

func highlightPortrait(row int) {
	hc := int(*gHighlightedChar)
	if hc != row && hc != -2 {
		if hc != -1 {
			drawIcon(*gGlobalIcons, hc+3, hc*0x26+10, 0x94, 0)
		}
		drawIcon(*gGlobalIcons, 2, row*0x26+10, 0x94, 0)
		*gHighlightedChar = int8(row)
	}
}

func getElementIndex(v int) int {
	for i := 0; ; i++ {
		if int(gElementRanges[i]) >= v {
			return i
		}
	}
}

func getBonusStatIndex(v int) int {
	for i := 0; ; i++ {
		if int(gBonusStatRanges[i]) >= v {
			return i
		}
	}
}

func hitWholeParty() {
	*gAttackScope = 2
	*gAttackType = 0
	spellAttackMonsters(0xb)
}

func getConditionStatMod(p *Chr, stat int) int {
	var might, intel, pers, endur, speed, acc, luck int
	c := &p.Cond
	if c[0xd] != 0 || c[0xe] != 0 || c[0xf] != 0 {
		return 0
	}
	if c[0] != 0 {
		luck -= int(c[0])
	}
	if v := int(c[5]); v != 0 {
		pers -= v
		intel -= v
		acc -= v
		might += v
		speed += v
	}
	if v := int(c[3]); v != 0 {
		might -= v
		speed -= v
		acc -= v
	}
	if v := int(c[4]); v != 0 {
		endur -= v
		pers -= v
		intel -= v
	}
	if v := int(c[1]); v != 0 {
		might -= v
		intel -= v
		pers -= v
		endur -= v
		speed -= v
		acc -= v
		luck -= v
	}
	if v := int(c[6]); v != 0 {
		might += v
		intel += v
		pers += v
		endur += v
		speed += v
		acc += v
		luck += v
	}
	if v := int(c[2]); v != 0 {
		might -= v
		intel -= v
		pers -= v
		endur -= v
		speed -= v
		acc -= v
		luck -= v
	}
	if v := int(c[7]); v != 0 {
		might -= v
		intel -= v
		pers += v
		endur -= v
		speed -= v
		acc -= v
		luck += v
	}
	switch stat {
	case 0:
		return might
	case 1:
		return intel
	case 2:
		return pers
	case 3:
		return endur
	case 4:
		return speed
	case 5:
		return acc
	}
	return luck
}

func updateOutdoorQuadrant() {
	if *gMapNum <= 0x21 {
		return
	}
	old := int(*gMapNum)
	enter := func() { *gMapNum = gQuadrantMaps[*gMapQuadrant]; *gCurTown = *gMapNum }
	if *gPartyX < 8 && *gMapQuadrant&1 == 0 {
		if *gMaps0West != 0 {
			loadMazeData(int(*gMaps0West) - 1)
			*gMapQuadrant |= 1
			enter()
		}
	} else if int8(*gPartyX) > 7 && *gMapQuadrant&1 != 0 {
		loadMazeData(int(*gMaps1BaseMap) - 1)
		*gMapQuadrant ^= 1
		enter()
	}
	if *gPartyY < 8 && *gMapQuadrant < 2 {
		if *gMaps0South != 0 {
			loadMazeData(int(*gMaps0South) - 1)
			*gMapQuadrant += 2
			enter()
		}
	} else if int8(*gPartyY) > 7 && *gMapQuadrant > 1 {
		loadMazeData(int(*gMaps2BaseMap) - 1)
		*gMapQuadrant -= 2
		enter()
	}
	m := &gMaps[*gMapQuadrant]
	switch {
	case int8(*gPartyX) < 0:
		if m.West != 0 {
			*gMapQuadrant--
			*gPartyX &= 0xf
		} else {
			*gPartyX = 0
		}
		*gMapNum = gQuadrantMaps[*gMapQuadrant]
	case int8(*gPartyX) > 0xf:
		if m.East != 0 {
			*gMapQuadrant++
			*gPartyX &= 0xf
		} else {
			*gPartyX = 0xf
		}
		*gMapNum = gQuadrantMaps[*gMapQuadrant]
	case int8(*gPartyY) < 0:
		if m.South != 0 {
			*gMapQuadrant -= 2
			*gPartyY &= 0xf
		} else {
			*gPartyY = 0
		}
		*gMapNum = gQuadrantMaps[*gMapQuadrant]
	case int8(*gPartyY) > 0xf:
		if m.North != 0 {
			*gMapQuadrant += 2
			*gPartyY &= 0xf
		} else {
			*gPartyY = 0xf
		}
		*gMapNum = gQuadrantMaps[*gMapQuadrant]
	}
	if int(*gMapNum) != old {
		showPleaseWait()
		freeMazeSprites()
	scan:
		for i := 0; i < 4; i++ {
			if int(gQuadrantMaps[i]) == old {
				for j := 0; j < 7; j++ {
					if gMaps[*gMapQuadrant].Surface[j] != gMaps[i].Surface[j] {
						save := *gCurTown
						loadMazeGfx(int(*gMapNum) - 1)
						*gCurTown = save
						break scan
					}
				}
			}
		}
		loadMaze(int(*gMapNum) - 1)
		*gCurTown = *gMapNum
		closeWindows(1)
	}
}

// objD0c is &g_map_objects[g_obj_d0c - 1] (index -1 when g_obj_d0c is 0, as the C reads it).
func objD0c() *Mobj { return at[Mobj](adrMapObjects + 12*(int(*gObjD0c)-1)) }

// ms and mu are g_maze_script[i] as the C's signed char and as (u8).
func ms(i int) int { return int(int8(gMazeScript.Add(i).U8(0))) }
func mu(i int) int { return int(gMazeScript.Add(i).U8(0)) }

// mapText is g_map_text + g_maze_text_offs[n].
func mapText(n int) []byte { return gMapText.Add(int(gMazeTextOffs[n])).B() }

// scriptValue reads an event operand after its type byte at ip: four bytes for 0x10/0x22, two
// for 0x19/0x23, one otherwise. It returns the value and the index after it.
func scriptValue(kind, ip int) (uint32, int) {
	switch kind {
	case 0x10, 0x22:
		return uint32(mu(ip+3))<<24 + uint32(mu(ip+2))<<16 + uint32(mu(ip+1))<<8 + uint32(mu(ip)), ip + 4
	case 0x19, 0x23:
		return uint32(mu(ip+1))<<8 + uint32(mu(ip)), ip + 2
	}
	return uint32(mu(ip)), ip + 1
}

func runEventScript() int {
	var line, i, r, ret, who, tmp, tmp2, op2, op3, op1, redraw, ip2, x, y, sp, flag int
	var xs, ys, ls [5]int
	var v1, v2, v3 uint32
	var tm Far
	var ip, k int

	updateOutdoorQuadrant()
	redraw = 0
	scanViewMonsters()
	mode := *gGameMode
	if *gLootGold|*gLootGems|int32(*gLootCount) != 0 {
		*gSavedLootGold = *gLootGold
		*gSavedLootGems = *gLootGems
		*gSavedLootCount = *gLootCount
		for k = 0; k < 10; k++ {
			gSavedLootFlags[k] = gLootFlags[k]
			gSavedLootElement[k] = gLootElement[k]
			gSavedLootMaterial[k] = gLootMaterial[k]
			gSavedLootAttrib[k] = gLootAttrib[k]
			gSavedLootIds[k] = gLootIds[k]
			gSavedLootCharges[k] = gLootCharges[k]
		}
		*gLootGold, *gLootGems, *gLootCount = 0, 0, 0
	} else {
		*gSavedLootGold, *gSavedLootGems, *gSavedLootCount = 0, 0, 0
	}

	// clear is the C's `clear:` (OP_Remove): the events of the party's cell are removed and an
	// opened object goes away.
	clear := func() {
		for i := 0; i < int(*gEventCount); i++ {
			if gEvents[i].X.Get() == int(*gPartyX) && gEvents[i].Y.Get() == int(*gPartyY) {
				gMazeScript.Add(gEvents[i].IP.Get()+5).SetU8(0, 0)
			}
		}
		if *gObjD0c != 0 || flag != 0 {
			objD0c().Y.Set(0x80)
			objD0c().X.Set(0x80)
		}
		switch objD0c().ID.Get() {
		case 0x2d, 0x2e:
			w := 0
			if getCellWallsRegion(int(*gPartyX), int(*gPartyY), 0x800) != 0 {
				w = 8
			}
			setCellWall(int(*gPartyX), int(*gPartyY), 2, w)
			closeWindows(int(*gOpenWindows))
			*gOpenWindows = 0
			drawWizardEye(1)
			refreshView()
		}
		*gEventObjOpened = 0
	}
	// Port addition for the guild membership NPCs (SetChar, then TakeOrGive giving award 0-4):
	// guildAward is that award while such a sale runs, guildLine its SetChar line.
	// guildEligible: can act, a spell caster (the classes buildSpellList sells to) and not yet a
	// member; guildNext is the award the next line gives.
	guildAward, guildLine := -1, 0
	guildEligible := func(c *Chr) bool { return charCannotAct(c) == 0 && c.Caster != 0 && c.Award[guildAward] == 0 }
	anyGuildEligible := func() bool {
		for k := 0; k < int(*gPartyCount); k++ {
			if guildEligible(&gParty[k]) {
				return true
			}
		}
		return false
	}
	guildNext := func() int {
		for k := 0; k < int(*gEventCount); k++ {
			e := &gEvents[k]
			if e.X.Get() != x || e.Y.Get() != y || e.Line.Get() != line+1 ||
				(e.Dir.Get() != int(*gPartyDir) && e.Dir.Get() != 4) {
				continue
			}
			ip := int(e.IP.Get())
			if ms(ip+5) != 12 {
				return -1
			}
			_, ip2 := scriptValue(ms(ip+6), ip+7)
			if v, _ := scriptValue(ms(ip2), ip2+1); ms(ip2) == 0x0f && v <= 4 {
				return int(v)
			}
			return -1
		}
		return -1
	}
	// pay is the C's `pay:` (the ending): the score is the party's average experience in
	// 10000s times 100000 plus the play time as hours*100 + minutes.
	pay := func() {
		saveGame(1)
		freeMazeResources(1)
		sndPlayMusic(0)
		*gMusicOn, *gSfxOn = 1, 1
		v1 = 0
		for i = 0; i < int(*gPartyCount); i++ {
			v1 += uint32(getTotalExp(&gParty[i]))
		}
		v1 = v1 / uint32(*gPartyCount)
		v1 = v1 / 10000
		v1 *= 100000
		v3 = uint32(getPlayTime())
		i = int(int16(v3 % 3600))
		r = int(int16(v3 / 3600))
		v2 = uint32(int32(int16(r*100 + i/60)))
		v1 += v2
		playEnding(v1)
		ret = 3
	}

restart:
	flag, *gNpcWindowOpen, *gEventObjOpened, ret, line, *gOpenWindows, *gViewDrawn = 0, 0, 0, 0, 0, 0, 0
	tm = 0
	x = int(*gPartyX)
	y = int(*gPartyY)
	who = 1
	*gEventChar = 1
	*gEventDmgType = 0
	sp = -1
	drawWizardEye(0)
next:
	if *gMonD0c1 != 0 {
		*gEventsDeferred = 1
	} else {
		*gEventsDeferred = 0
		for k = 0; k < int(*gEventCount); k++ {
			e := &gEvents[k]
			if e.X.Get() != x || e.Y.Get() != y {
				continue
			}
			if x|y|int(*gPartyDir) == 0 {
				continue
			}
			if e.Dir.Get() != int(*gPartyDir) && e.Dir.Get() != 4 {
				continue
			}
			if e.Line.Get() != line {
				continue
			}
			if *gObjD0c != 0 && *gLastMoveDir != *gPartyDir && e.Dir.Get() == int(*gPartyDir) {
				*gLastMoveDir = *gPartyDir
			}
			if line == 0 && sp == -1 && !interacting && eventWaitsForSpace(e) {
				continue // port addition: an interaction waits for Space
			}
			if *gObjD0c == 0 || *gLastMoveDir == *gPartyDir || *gGameMode == 9 {
				*gGameMode = 9
				ip = int(e.IP.Get())
				strncpy(gTextBuf2, mapText(ms(ip+6)), 0x12c)
				switch ms(ip+5) - 1 {
				case 31: // OP_WhoWill
					if *gViewDrawn == 0 {
						refreshView()
					}
					sprintf(gTextBuf, gostr(gTextBuf2), ds[adrParty+0x12f*(who-1):]) // g_party[who - 1]: who may be 0
					if *gOpenWindows != 0 {
						printWindowText(gTextBuf)
					} else {
						openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf))
						*gOpenWindows++
					}
				case 0: // OP_Display0x01
					if *gViewDrawn == 0 {
						refreshView()
					}
					if *gNpcWindowOpen != 0 {
						closeWindows(int(*gOpenWindows))
						*gNpcWindowOpen, *gOpenWindows = 0, 0
					}
					if *gOpenWindows != 0 {
						sprintf(gTextBuf, gostr(gFmtNewlineStr), gTextBuf2)
						printWindowText(gTextBuf)
					} else {
						openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTextBuf2))
						*gOpenWindows++
					}
				case 1: // OP_DoorTextSml
					*gDoorIconImg = *gWallGfx4
					sprintf(gViewCaption, fmtP(gFmtDoorTextSml), adrDoorIconSprite, gTextBuf2)
				case 2: // OP_DoorTextLrg
					sprintf(gViewCaption, gostr(gFmtDoorTextLrg), gTextBuf2)
				case 3: // OP_SignText
					sprintf(gViewCaption, gostr(gFmtSignText), gTextBuf2)
				case 4: // OP_NPC
					if *gViewDrawn == 0 {
						refreshView()
					}
					if *gOpenWindows != 0 && *gNpcWindowOpen == 0 {
						closeWindows(int(*gOpenWindows))
						*gOpenWindows = 0
					}
					strncpy(gTmpText, mapText(ms(ip+7)), 0xfa)
					if faceYesNoDialog(ms(ip+8), gTextBuf2, gTmpText, ms(ip+9), int(*gOpenWindows)) != 0 {
						line = ms(ip+0xa) - 1
					}
					if *gOpenWindows != 0 {
						break
					}
					*gOpenWindows++
					*gNpcWindowOpen = uint8(*gOpenWindows)
				case 5: // OP_PlayFX
					playSfx(ms(ip + 6))
				case 6, 30: // OP_TeleportAndExit, OP_TeleportAndContinue
					if *gViewDrawn == 0 {
						refreshView()
					}
					if ms(ip+6) != 0 {
						tmp = ms(ip + 6)
						i = ms(ip + 7)
						r = ms(ip + 8)
					} else {
						i = int(gTeleportXs[tmp])
						r = int(gTeleportYs[tmp])
						tmp = int(gTeleportMaps[tmp])
						if i == 0 && r == 0 {
							i = -1
						}
						playSfx(0x33)
					}
					ip = ms(ip + 5)
					if int(*gCurTown) != tmp {
						into := tmp < 0x29 || tmp == 0x6a
						verb, name := gStrLeaving, []byte(nil)
						if into {
							verb = gStrEntering
						}
						if tmp == 0x6a {
							name = gStrTheArena
						} else if into {
							name = str(gMapNamesFrom1[tmp-1])
						} else {
							name = str(gMapNamesFrom1[int(*gCurTown)-1])
						}
						sprintf(gTmpText, gostr(gFmtEnteringLeaving), verb, name)
						openTextWindow(0, 0x8f, 0x140, 0x38, 0, 0x96, 4, 6, dsFar(adrTmpText))
						*gOpenWindows++
						switch objD0c().ID.Get() {
						case 0x2f:
							playSfx(0x2d)
						case 0x30:
							playSfx(0x2c)
						}
						enterMaze(tmp)
					} else {
						playSfx(0x33)
					}
					if i == -1 {
						setPartyMapStart()
					} else {
						*gPartyX = uint8(i)
						*gPartyY = uint8(r)
					}
					updateOutdoorQuadrant()
					if ip == 0x1f {
						*gLightChanged = 0
					}
					closeWindows(int(*gOpenWindows))
					*gOpenWindows = 0
					drawWizardEye(1)
					flushInput()
					if ip == 0x1f {
						goto restart
					}
					goto done
				case 7, 8, 9: // OP_If1, OP_If2, OP_If3
					v1, ip2 = scriptValue(ms(ip+6), ip+7)
					tmp2 = ms(ip2)
					if who == 0 && ms(ip+6) != 0x2c {
						i, r = 0, 0
						for ; i < int(*gPartyCount); i++ {
							r |= eventCompareValue(ms(ip+6), v1, ms(ip+5)-8, i)
						}
						if r == 0 {
							break
						}
						line = tmp2 - 1
						break
					}
					if eventCompareValue(ms(ip+6), v1, ms(ip+5)-8, who-1) == 0 {
						break
					}
					line = tmp2 - 1
				case 10: // OP_MoveObj
					o := &gMapObjects[ms(ip+6)]
					if o.Y.Get() == ms(ip+8) && o.X.Get() == ms(ip+7) {
						o.Attr.Set(o.Attr.Get() ^ 1)
					} else {
						o.Y.Set(ms(ip + 8))
						o.X.Set(ms(ip + 7))
					}
				case 11, 27, 28, 32: // OP_TakeOrGive, _2, _3, OP_RndDamage
					redraw = 1
					if *gViewDrawn == 0 {
						refreshView()
					}
					v1, ip2 = scriptValue(ms(ip+6), ip+7)
					op2 = ms(ip2)
					v2, ip2 = scriptValue(op2, ip2+1)
					op3 = ms(ip2)
					v3, _ = scriptValue(op3, ip2+1)
					if *gOpenWindows != 0 && op2 == 0x43 {
						closeWindows(int(*gOpenWindows))
						*gOpenWindows = 0
					}
					if ms(ip+5) == 0x1d {
						if who == 0 {
							for i = 0; i < int(*gPartyCount); i++ {
								if eventCompareValue(ms(ip+6), v1, 1, i) != 0 && eventCompareValue(op2, v2, 1, i) != 0 {
									eventTakeGive(0, 0, op3, v3, i)
									if op3 == 0x52 {
										break
									}
								}
							}
						} else {
							if eventCompareValue(ms(ip+6), v1, 1, who-1) == 0 {
								break
							}
							if eventCompareValue(op2, v2, 1, who-1) == 0 {
								break
							}
							eventTakeGive(0, 0, op3, v3, who-1)
						}
						break
					}
					if ms(ip+5) == 0x1c || ms(ip+5) == 0x21 {
						cmp := 1
						if ms(ip+5) == 0x21 {
							cmp = 2
						}
						if who == 0 {
							for i = 0; i < int(*gPartyCount); i++ {
								if eventCompareValue(ms(ip+6), v1, cmp, i) != 0 {
									eventTakeGive(0, 0, op2, v2, i)
									if op2 == 0x52 {
										break
									}
								}
							}
						} else {
							if eventCompareValue(ms(ip+6), v1, cmp, who-1) == 0 {
								break
							}
							eventTakeGive(0, 0, op2, v2, who-1)
						}
						break
					}
					if who == 0 {
						op1 = ms(ip + 6)
						for i = 0; i < int(*gPartyCount); i++ {
							r = eventTakeGive(op1, v1, op2, v2, i)
							switch op1 {
							case 8:
								op1 = 0
								fallthrough
							case 0x15, 0x42:
								if r == 0 {
									goto nextline
								}
							case 0x22, 0x23, 0x41:
								if r == 0 {
									goto nextline
								}
								goto done
							}
							switch op2 {
							case 0x52:
								op1 = 0
								fallthrough
							case 0x15, 0x22, 0x23, 0x41, 0x42:
								if r == 0 {
									goto nextline
								}
							}
						}
						break
					}
					if eventTakeGive(ms(ip+6), v1, op2, v2, who-1) != 0 {
						goto done
					}
					if guildAward >= 0 && op2 == 0x0f && int(v2) == guildAward {
						playSfx(0x36) // awardFoundLoot's gold "bling" (quest rewards)
						img := loadGraphic("spfx9.icn")
						sparklePortrait(img, who-1)
						freeAndNull(&img)
						drawPartyPortraits()
						if anyGuildEligible() {
							line = guildLine
							goto next
						}
						guildAward = -1
					}
					if op2 != 0x4f {
						break
					}
					closeWindows(int(*gOpenWindows))
					*gOpenWindows = 0
					drawWizardEye(1)
				case 12: // OP_NoAction
				case 13: // OP_Remove
					clear()
				case 14: // OP_SetChar
					if ms(ip+6) > 8 {
						if *gOpenWindows != 0 && *gNpcWindowOpen == 0 {
							closeWindows(int(*gOpenWindows))
							*gOpenWindows = 0
						}
						if guildAward = guildNext(); guildAward >= 0 {
							guildLine = line
							if !anyGuildEligible() {
								goto done
							}
							who = selectPartyMemberIf(guildEligible) + 1
						} else {
							who = selectPartyMember() + 1
						}
						if who == 0 {
							goto done
						}
						*gEventChar = 1
					} else {
						if ms(ip+6) != 0 {
							*gEventChar = uint8(ms(ip + 6))
						} else {
							who = ms(ip + 6)
							*gEventChar = uint8(who)
						}
					}
				case 15: // OP_Spawn
					m := ms(ip + 6)
					gMonY[m] = int16(ms(ip + 8))
					gMonX[m] = int16(ms(ip + 7))
					gMonFrames[m] = int16(randomRange(0, int(gMonFrameCounts[gMonType[m]])))
					gMonStatus[m] = 0
					gMonVisible[m] = 0
					gMonHp[m] = gMonMaxHp.I16(2 * int(gMonType[m]))
				case 16: // OP_DoTownEvent
					switch ms(ip + 6) {
					case 0:
						visitBank()
					case 1:
						visitSmithy()
					case 2:
						visitGuild()
					case 3:
						if visitInn() != 0 {
							ret = 2
							goto done
						}
					case 4:
						visitTavern()
					case 5:
						visitTemple()
					case 6:
						visitTraining()
					case 7:
						runArena()
					case 8:
						tellJesterJoke()
						goto done
					}
					*gViewDirty = 1
					ret = 1
					goto done
				case 17: // OP_Exit
					goto done
				case 18: // OP_AfterMap
					if ms(ip+8) == 4 {
						for i = 0; i <= 3; i++ {
							setCellWall(ms(ip+6), ms(ip+7), i, ms(ip+9))
						}
					} else {
						setCellWall(ms(ip+6), ms(ip+7), ms(ip+8), ms(ip+9))
					}
				case 19: // OP_GiveMulti
					redraw = 1
					v1, ip2 = scriptValue(ms(ip+6), ip+7)
					op2 = ms(ip2)
					v2, ip2 = scriptValue(op2, ip2+1)
					op3 = ms(ip2)
					v3, _ = scriptValue(op3, ip2+1)
					if *gViewDrawn == 0 {
						refreshView()
					}
					if *gOpenWindows != 0 {
						closeWindows(int(*gOpenWindows))
						*gOpenWindows = 0
					}
					w := who
					if who != 0 {
						w = who - 1
					}
					if eventGiveMulti(ms(ip+6), v1, op2, v2, op3, v3, w) != 0 {
						if *gEventObjOpened == -1 {
							*gEventObjOpened = 0
							goto done
						}
						if ms(ip+6) == 0x43 || op2 == 0x43 || op3 == 0x43 {
							flag = 1
						} else {
							goto done
						}
					} else {
						if ms(ip+6) == 0x43 || op2 == 0x43 || op3 == 0x43 {
							goto done
						}
					}
					*gEventObjOpened = 1
				case 20: // OP_ConfirmWord
					var buf [100]byte
					if *gViewDrawn == 0 {
						refreshView()
					}
					if ms(ip+8) != 0 {
						strncpy(gTmpText, mapText(ms(ip+8)), 0xfa)
					} else {
						strcpy(gTmpText, gTextBuf2)
					}
					if ms(ip+9) != 0 {
						strncpy(buf[:], mapText(ms(ip+9)), len(buf))
					} else {
						strcpy(buf[:], gStrWhatsThePassword)
					}
					if tmp = askAnswer(ms(ip+6), gTmpText, buf[:]); tmp == 0 {
						break
					}
					if tmp == 7 {
						pay()
						goto done
					}
					line = ms(ip+7) - 1
				case 21: // OP_Damage
					if *gViewDrawn == 0 {
						refreshView()
					}
					if *gOpenWindows != 0 {
						closeWindows(int(*gOpenWindows))
						*gOpenWindows = 0
					}
					damageParty(mu(ip+7)<<8+mu(ip+6), ms(ip+8), who-1)
				case 22: // OP_JumpRnd
					if randomRange(1, ms(ip+6)) != ms(ip+7) {
						break
					}
					line = ms(ip+8) - 1
				case 23: // OP_AfterEvent
					for i = 0; i < int(*gEventCount); i++ {
						ev := &gEvents[i]
						if ev.X.Get() == int(*gPartyX) && ev.Y.Get() == int(*gPartyY) && (ev.Dir.Get() == int(*gPartyDir) || ev.Dir.Get() == 4) &&
							ev.Line.Get() == ms(ip+6) {
							gMazeScript.Add(ev.IP.Get()+5).SetU8(0, uint8(ms(ip+7)))
						}
					}
				case 24: // OP_CallEvent
					sp++
					xs[sp] = x
					ys[sp] = y
					ls[sp] = line
					x = ms(ip + 6)
					y = ms(ip + 7)
					line = ms(ip+8) - 1
				case 25: // OP_Return
					x = xs[sp]
					y = ys[sp]
					line = ls[sp]
					sp--
				case 26: // OP_SetVar
					redraw = 1
					v1, _ = scriptValue(ms(ip+6), ip+7)
					if who == 0 {
						for i = 0; i < int(*gPartyCount); i++ {
							eventSetValue(ms(ip+6), v1, i)
						}
					} else {
						eventSetValue(ms(ip+6), v1, who-1)
					}
				case 29: // OP_CutsceneEndClouds
					if partyHasHoloCards() != 0 {
						pay()
						goto done
					}
					showMessage(gMsgNeedHoloCards)
					goto done
				}
			nextline:
				if *gPartyDefeated != 0 {
					break
				}
				line++
				goto next
			}
		}
	}
done:
	if redraw != 0 {
		closeWindows(int(*gOpenWindows))
		*gOpenWindows = 0
		drawWizardEye(1)
		drawPartyPortraits()
	}
	updatePartyDisabled()
	if *gObjD0c != 0 && *gEventObjOpened != 0 {
		switch objD0c().ID.Get() {
		case 4, 5, 0xc, 0x12, 0x14, 0x17, 0x21, 0x26, 0x28:
			// goto clear: OP_Remove, then on at nextline.
			clear()
			if *gPartyDefeated != 0 {
				goto done
			}
			line++
			goto next
		}
	}
	*gRopeCreated = 0
	if *gLootGold|*gLootGems|int32(*gLootCount) != 0 {
		awardFoundLoot()
	}
	*gGameMode = mode
	freeAndNull(&tm)
	closeWindows(int(*gOpenWindows))
	if ret == 1 {
		goto restart
	}
	if *gSavedLootGold|*gSavedLootGems|int32(*gSavedLootCount) != 0 {
		*gLootGold = *gSavedLootGold
		*gLootGems = *gSavedLootGems
		*gLootCount = *gSavedLootCount
		for k = 0; k < 10; k++ {
			gLootFlags[k] = gSavedLootFlags[k]
			gLootElement[k] = gSavedLootElement[k]
			gLootMaterial[k] = gSavedLootMaterial[k]
			gLootAttrib[k] = gSavedLootAttrib[k]
			gLootIds[k] = gSavedLootIds[k]
			gLootCharges[k] = gSavedLootCharges[k]
		}
	}
	if ret == 3 {
		ret = 1
	}
	return ret
}

func playSfx(v int) {
	if *gSfxOn != 0 && v < 0x97 {
		sndTimer(v)
	} else if v > 0x96 && *gMusicOn != 0 {
		sndTimer(v)
	}
}

// checkWallListSize: o is the draw-list index the wall list reached.
func checkWallListSize(o int) {
	save := *gBtnXs
	*gBtnXs = 0x41d
	n := o
	max := 1999
	if uint16(n) > *gWallListMax {
		*gWallListMax = uint16(n)
	}
	if n > max {
		sprintf(gTmpText, gostr(gMsgWallListOverflow), n, max)
		openTextWindow(0x32, 0x4c, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTmpText))
		getInput()
		closeWindows(1)
	}
	*gBtnXs = save
}

func checkAutoListSize(n int) {
	save := *gBtnXs
	*gBtnXs = 0x41d
	max := 1999
	if uint16(n) > *gAutoListMax {
		*gAutoListMax = uint16(n)
	}
	if n > max {
		sprintf(gTmpText, gostr(gMsgAutoListOverflow), n, max)
		openTextWindow(0x32, 0x4c, 0xd8, 0x24, 0, 0x96, 4, 6, dsFar(adrTmpText))
		getInput()
		closeWindows(1)
	}
	*gBtnXs = save
}

// monsterCell is &g_monster_grid[y][x], unchecked as in the C.
func monsterCell(y, x int) *uint8 { return &ds[adrMonsterGrid+y*32+x] }

func tryMoveMonster(x, y, m int) {
	ny := int(gMonY[m]) + y
	nx := int(gMonX[m]) + x
	w := gMonsterCellWeights[gMonType[m]]
	if int(*monsterCell(ny, nx))+int(w) < 4 {
		if gMonStatus[m] == 0 && *gMonstersCanMove != 0 {
			*monsterCell(ny, nx) += w
			*monsterCell(int(gMonY[m]), int(gMonX[m])) -= w
			gMonY[m] = int16(ny)
			gMonX[m] = int16(nx)
			gMonsterMoved[m] = 1
		}
		*gMonstersMoved = 1
	}
}

func moveMonsters() {
	*gMonsterScanDelay = 0
	if *gSpellAnimActive != 0 {
		*gMonScanPending = 1
		return
	}
	memset(ds[adrMonsterGrid:], 0, 0x400)
	memset(ds[adrMonHitAnim:], 0, 0xc)
	memset(ds[adrMonElemFx:], 0, 0xc)
	memset(gMonsterMoved, 0, 0xaa)
	memset(gMonsterQueued, 0, 0xaa)
	memset(gRangedAttackers, 0xff, 0x24)
	*gMonstersMoved = 0
	px, py := int(*gPartyX), int(*gPartyY)
	fac := int(*gPartyDir)
	for k := 0; k < int(*gMonCount); k++ {
		if uint16(gMonY[k]) < 0x20 {
			*monsterCell(int(gMonY[k]), int(gMonX[k])) += gMonsterCellWeights[gMonType[k]]
		}
	}
	side := func(k, j int) {
		if k < 0x15 || k > 0x1b {
			tryMoveMonster(0, int(gMonSideStep[k]), j)
		} else {
			tryMoveMonster(int(gMonSideStep[k]), 0, j)
		}
	}
	for pass := 0; pass < 2; pass++ {
		k := -1
		for dy := 3; dy > -4; dy-- {
			for dx := -3; dx < 4; dx++ {
				k++
				xx := px + dx
				yy := py + dy
				for j := 0; j < int(*gMonCount); j++ {
					if int(gMonY[j]) == yy && int(gMonX[j]) == xx && (gMonVisible[j] != 0 || *gGameMode == 5) && gMonsterMoved[j] == 0 {
						if (px == xx || py == yy) && gMonRanged.U8(int(gMonType[j])) != 0 && gMonsterQueued[j] == 0 &&
							j+1 != int(*gMonD0c1) && j+1 != int(*gMonD0c2) && j+1 != int(*gMonD0c3) && gMonStatus[j] == 0 {
							queueRangedAttack(int(gMonType[j]), xx, yy)
							gMonsterQueued[j] = 1
						}
						step := func() bool { return getCellWallsRegion(xx, yy, int(gDirBlockBits[gMonStepWallDir[k]])) == 0 }
						switch fac >> 1 {
						case 0:
							if step() {
								tryMoveMonster(int(gMonStepDx[k]), int(gMonStepDy[k]), j)
							} else if getCellWallsRegion(xx, yy, int(gDirBlockBits[gMonSideWallDir[k]])) == 0 {
								side(k, j)
							}
						case 1:
							if getCellWallsRegion(xx, yy, int(gDirBlockBits[gMonSideWallDir[k]])) == 0 {
								side(k, j)
							} else if step() {
								tryMoveMonster(int(gMonStepDx[k]), int(gMonStepDy[k]), j)
							}
						}
					}
				}
			}
		}
	}
	if *gMonMissileActive != 0 {
		doRangedAttacks()
	}
}

// interacting is set while Space runs the party's cell events (port addition): only then do
// interactions start (see eventWaitsForSpace); traps, ambushes and the rest still fire on
// walking in.
var interacting bool

// hasCellInteraction: an event chain on the party's cell, facing its way, waits for Space.
func hasCellInteraction() bool {
	x, y, d := int(*gPartyX), int(*gPartyY), int(*gPartyDir)
	for i := 0; i < int(*gEventCount); i++ {
		e := &gEvents[i]
		if e.X.Get() == x && e.Y.Get() == y && e.Line.Get() == 0 && (e.Dir.Get() == d || e.Dir.Get() == 4) &&
			ms(int(e.IP.Get())+5) != 0 && eventWaitsForSpace(e) {
			return true
		}
	}
	return false
}

// eventWaitsForSpace (port addition): the event chain e starts (on the party's cell, facing
// its way) only on Space when it is an interaction and the player can see there is
// something: an object on the cell (chest, fountain, mirror...), or for an event tied to a
// facing indoors, a wall there that is not plain (door, torch, locked or bashed door, the
// special type 7). Everything else fires on walking in, as in the original.
func eventWaitsForSpace(e *Evt) bool {
	x, y, d := e.X.Get(), e.Y.Get(), int(*gPartyDir)
	if !eventNeedsInteract(x, y, d) {
		return false
	}
	for i := 0; i < int(*gObjCount); i++ {
		if gMapObjects[i].X.Get() == x && gMapObjects[i].Y.Get() == y {
			return true
		}
	}
	if e.Dir.Get() == 4 || *gOutdoors != 0 {
		return false
	}
	w := getCellWallsRegion(x, y, int(uint16(gViewWallMasks[d][2]))) >> uint(gViewWallShifts[d][2]) & 7
	return w >= 2
}

// eventNeedsInteract: whether the event chain on (x, y) facing dir is an interaction with
// the environment rather than something that happens to the party. It follows the chain in
// line order (into CallEvent subroutines and back) to the first op that decides: a question
// to the player (Yes/No, an NPC, a password, picking a character) makes it an interaction; a
// trap or ambush (damage, monsters, a random jump), a move (teleport, stairs, a building
// door), a door/sign caption or a chain that only gives or takes (treasure piles) does not.
func eventNeedsInteract(x, y, dir int) bool {
	find := func(x, y, line int) (int, bool) {
		for i := 0; i < int(*gEventCount); i++ {
			e := &gEvents[i]
			if e.X.Get() == x && e.Y.Get() == y && (e.Dir.Get() == dir || e.Dir.Get() == 4) && e.Line.Get() == line {
				return int(e.IP.Get()), true
			}
		}
		return 0, false
	}
	var stack [][3]int
	cx, cy, line := x, y, 0
	for steps := 0; steps < 64; steps++ {
		ip, ok := find(cx, cy, line)
		if !ok {
			if len(stack) == 0 {
				break
			}
			t := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			cx, cy, line = t[0], t[1], t[2]+1
			continue
		}
		switch ms(ip+5) - 1 { // the op as runEventScript's switch numbers it
		case 4, 20, 31: // NPC, ConfirmWord, WhoWill
			return true
		case 14: // SetChar: 9+ asks who
			if ms(ip+6) > 8 {
				return true
			}
		case 7, 8, 9: // If: kind 0x2c is a Yes/No question
			if ms(ip+6) == 0x2c {
				return true
			}
		case 21, 32, 15, 22: // Damage, RndDamage, Spawn, JumpRnd
			return false
		case 6, 30, 16, 1, 2, 3: // teleports, town buildings, door and sign captions
			return false
		case 24: // CallEvent
			if len(stack) < 8 {
				stack = append(stack, [3]int{cx, cy, line})
				cx, cy, line = ms(ip+6), ms(ip+7), ms(ip+8)
				continue
			}
		case 25: // Return
			if len(stack) > 0 {
				t := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				cx, cy, line = t[0], t[1], t[2]+1
				continue
			}
		}
		line++
	}
	return false
}
