// Port of ENDING.C (overlay 17fd): the end-game cinematic (the CyberSpace computer and its
// code lock, the screens, the fight, the lift-off) and the final score.
//
// The C's macros are closures: S is gAnimSprites, W its int16 view (W[7k+3] = S[k].x,
// +4 y, +6 frame), DRAW/DRAW2 draw it with drawSprites/drawSprites2a, SONG plays a voice
// sample lip-synced from the cue tables, WALK and TICK are the walk and score-ticker steps.
package mm3

func endingAnimUntil(t, f int) {
	s := gAnimSprites
	for {
		*gAnimTicks = uint32(int32(sndTimer(-1)))
		if *gAnimTicks >= uint32(int32(t)) {
			break
		}
		s[1].Frame.Set(randomRange(0, 15))
		if f++; f > 11 {
			f = 0
		}
		s[2].Frame.Set(f)
		drawSprites(animSprites())
	}
}

func playEnding(score uint32) int {
	var c2, c4, key, i, j int
	var keys [6]int
	var digits [10]int
	s := gAnimSprites
	w := arr[int16](adrAnimSprites)
	draw := func() { drawSprites(animSprites()) }
	draw2 := func() { drawSprites2a(animSprites()) }
	rnd := func() int { return randomRange(0, 15) }
	clear := func(n int) {
		for i = 0; i < n; i++ {
			w[i] = 0
		}
		for i = 0; i < n; i += 7 {
			w[i] = -1
		}
	}
	img := func(k int, p Far) { s[k].Img.Set(uint32(p)) }
	frames := func(n int, f func(i int)) {
		for i = 0; i < n; i++ {
			f(i)
			draw()
		}
	}
	setFrame := func(i int) { s[0].Frame.Set(i) }
	setFrameRnd := func(i int) { s[0].Frame.Set(i); s[1].Frame.Set(rnd()) }
	rndOnly := func(int) { s[1].Frame.Set(rnd()) }
	song := func(p Far, n, row int) {
		sndTimer(-2)
		*gTalkFramePos = 0
		sndPlaySample(p, n)
		for {
			s[0].Frame.Set(int(gSongMouthFrames[row][*gTalkFramePos]))
			s[1].Frame.Set(rnd())
			if j++; j > 11 {
				j = 0
			}
			s[2].Frame.Set(j)
			draw()
			endingAnimUntil(int(int16(gSongCueTimes[row][*gTalkFramePos])), j)
			for {
				*gTalkFramePos++
				if uint32(gSongCueTimes[row][*gTalkFramePos]) >= *gAnimTicks {
					break
				}
			}
			if sndPlaySample(Far(1), 0) == 0 {
				break
			}
		}
		s[0].Frame.Set(0)
	}
	walk := func() {
		if c4++; c4 > 1 {
			c4 = 0
			if j++; j > 9 {
				j = 0
			}
			s[0].Frame.Set(int(gWalkFrames[j]))
			if gWalkImgSwap[j] != 0 {
				*gCineImg9 = *gCineImg0
				*gCineImg0 = *gCineImg5
				*gCineImg5 = *gCineImg9
				img(0, *gCineImg0)
			}
		}
	}
	tick := func() {
		if c4++; c4 > 12 {
			c4 = 0
		}
		s[2].Frame.Set(c4)
		draw2()
	}
	waitKey := func(f func()) {
		for bioskey(1) == 0 && *gUserBreak == 0 {
			f()
		}
	}
	rndDraw := func() { s[1].Frame.Set(rnd()); draw() }
	// the four-image door sequence of the start
	door := func(raw string, names [4]string) {
		if raw != "" {
			showRawPicture(raw)
		}
		imgs := []*Far{gCineImg0, gCineImg1, gCineImg2, gCineImg3}
		for k, n := range names {
			*imgs[k] = loadGraphic(n)
		}
		if raw == "" {
			clear(0x26)
			s[0].X.Set(0x50)
			s[0].Y.Set(0x24)
		}
		playSfx(0x18)
		for k, n := range []int{6, 8, 9, 9} {
			img(0, *imgs[k])
			frames(n, setFrame)
		}
		freeAndNull(gCineImg3)
		freeAndNull(gCineImg2)
		freeAndNull(gCineImg1)
		freeAndNull(gCineImg0)
	}
	hud := func(n int) {
		img(0, *gCineImg0)
		s[0].X.Set(0x16)
		s[0].Y.Set(0x11)
		showRawPicture("hud.raw")
		playSfx(0x3a)
		for j = 0; j < n; j++ {
			frames(14, setFrame)
		}
		freeAndNull(gCineImg0)
	}

	tries := 3
	memset(animSprites(), 0, 0xe6)
	if *gMusicEnabled != 0 {
		playMusic(cs("cyber.m"), 0)
	}
	fadeScreen(0, 8)
	clearScreen()
	showRawPicture("outh.raw")
	fadeScreen(1, 4)
	clear(0x26)
	door("", [4]string{"outa.vga", "outb.vga", "outc.vga", "outd.vga"})
	door("inh.raw", [4]string{"ina.vga", "inb.vga", "inc.vga", "ind.vga"})
	*gCineImg0 = loadGraphic("hud.vga")
	hud(4)
	playSfx(0)

	// the computer
	*gCineImg9 = loadGraphic("mess.vga")
	*gCineImg0 = loadGraphic("comf5.vga")
	*gCineImg1 = loadGraphic("comf4.vga")
	*gCineImg5 = loadGraphic("led.vga")
	*gCineImg6 = loadGraphic("pan.vga")
	*gCineImg7 = loadGraphic("lbut.vga")
	showRawPicture("computer.raw")
	*gCineImg2 = loadGraphic("comf3.vga")
	*gCineImg3 = loadGraphic("comf2.vga")
	*gCineImg4 = loadGraphic("comf1.vga")
	clear(0x26)
	s[0].X.Set(0x56)
	s[0].Y.Set(4)
	s[1].X.Set(0x4f)
	s[1].Y.Set(0x89)
	s[2].X.Set(0x25)
	s[2].Y.Set(0x16)
	playSfx(0x39)
	img(1, *gCineImg6)
	img(0, *gCineImg4)
	frames(0x1a, setFrameRnd)
	img(0, *gCineImg3)
	frames(7, setFrameRnd)
	img(0, *gCineImg2)
	frames(4, setFrameRnd)
	freeAndNull(gCineImg4)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	img(0, *gCineImg1)
	frames(7, setFrameRnd)
	img(0, *gCineImg0)
	s[0].Frame.Set(0)
	rndDraw()
	for i = 0; i < 5; i++ {
		rndDraw()
		sprintf(gCineNameBuf, "c%d", i+1)
		gCineVoices[i] = loadCCFile(gostr(gCineNameBuf))
		gCineVoiceSizes[i] = *gFileSize
	}
	playSfx(0)
	voice := func(k, row int) { song(gCineVoices[k], int(gCineVoiceSizes[k]), row) }
	closeDoor := func() {
		img(0, *gCineImg1)
		for i = 6; i > 0; i-- {
			s[0].Frame.Set(i - 1)
			rndDraw()
		}
	}
	for {
		s[2].X.Set(0x25)
		s[2].Y.Set(0x16)
		s[2].Frame.Set(0)
		img(3, *gCineImg9)
		s[3].X.Set(0x62)
		s[3].Y.Set(0xaa)
		s[3].Frame.Set(0)
		img(2, *gCineImg5)
		voice(0, 0)
		s[2].Y.Set(0xdc)
		rndDraw()
		s[2].X.Set(0)
		s[0].Frame.Set(0)
		flushInput()
		for c4 = 0; c4 < 6; c4++ {
			for {
				waitKey(rndDraw)
				key = readKey()
				if key >= '1' && key <= '6' {
					break
				}
			}
			playSfx(0x14)
			keys[c4] = key
			s[2].Frame.Set(key - '1')
			img(2, *gCineImg7)
			for i = 0; i < 4; i++ {
				s[2].Y.Set(0x88)
				rndDraw()
				s[2].Y.Set(0xdc)
				rndDraw()
			}
		}
		s[3].Y.Set(200)
		closeDoor()
		frames(10, rndOnly)
		frames(7, setFrameRnd)
		s[2].X.Set(0x25)
		s[2].Y.Set(0x16)
		s[2].Frame.Set(0)
		img(2, *gCineImg5)
		s[2].Y.Set(0x16)
		img(0, *gCineImg0)
		voice(1, 1)
		if keys[0] == '6' && keys[1] == '4' && keys[2] == '5' && keys[3] == '2' && keys[4] == '3' && keys[5] == '1' {
			break
		}
		s[3].Y.Set(0xaa)
		s[3].Frame.Set(1)
		img(2, *gCineImg5)
		voice(2, 2)
		if tries--; tries != 0 {
			s[2].Y.Set(0xdc)
			frames(0x14, rndOnly)
		} else {
			for i = 0; i < 5; i++ {
				freeAndNull(&gCineVoices[i])
			}
			freeAndNull(gCineImg0)
			freeAndNull(gCineImg1)
			freeAndNull(gCineImg5)
			freeAndNull(gCineImg6)
			freeAndNull(gCineImg7)
			fadeScreen(0, 4)
			flushInput()
			return 1
		}
	}
	s[3].Y.Set(0xaa)
	s[3].Frame.Set(2)
	for i = 3; i < 5; i++ {
		s[2].Y.Set(0x16)
		voice(i, i)
		s[2].Y.Set(0xdc)
		s[0].Frame.Set(0)
		for c4 = 0; c4 < 10; c4++ {
			rndDraw()
		}
	}
	s[3].Y.Set(200)
	closeDoor()
	for i = 0; i < 5; i++ {
		freeAndNull(&gCineVoices[i])
	}
	for i = 0; i < 2; i++ {
		rndDraw()
		sprintf(gCineNameBuf, "c%d", i+6)
		gCineVoices[i] = loadCCFile(gostr(gCineNameBuf))
		gCineVoiceSizes[i] = *gFileSize
	}
	frames(7, setFrameRnd)
	s[3].Y.Set(0xaa)
	s[3].Frame.Set(3)
	s[2].Y.Set(0x16)
	img(0, *gCineImg0)
	voice(0, 5)
	s[2].Y.Set(0xdc)
	s[0].Frame.Set(0)
	rndDraw()

	// the screens
	*gCineImg2 = loadGraphic("scn1.vga")
	img(0, *gCineImg2)
	frames(0x17, setFrameRnd)
	freeAndNull(gCineImg2)
	*gCineImg2 = loadGraphic("scn2.vga")
	img(0, *gCineImg2)
	s[2].X.Set(0x56)
	s[2].Y.Set(4)
	s[3].Img.Set(0)
	for i = 0; i < 5; i++ {
		s[0].Frame.Set(0)
		s[2].Img.Set(0)
		sprintf(gCineNameBuf, "txt%d.vga", i+1)
		*gCineImg3 = loadGraphic(gostr(gCineNameBuf))
		draw()
		img(2, *gCineImg3)
		for j = 0; j < 0x12; j++ {
			s[0].Frame.Set(j + 1)
			s[1].Frame.Set(rnd())
			s[2].Frame.Set(j)
			draw()
		}
		img(0, *gCineImg3)
		s[0].Frame.Set(0x11)
		draw()
		flushInput()
		waitKey(rndDraw)
		flushInput()
		img(0, *gCineImg2)
		for j = 0x12; j > 1; j-- {
			s[0].Frame.Set(j)
			s[1].Frame.Set(rnd())
			s[2].Frame.Set(j - 1)
			draw()
		}
		s[0].Frame.Set(0)
		s[2].Img.Set(0)
		draw()
		freeAndNull(gCineImg3)
	}
	freeAndNull(gCineImg2)
	*gCineImg2 = loadGraphic("scn1.vga")
	img(0, *gCineImg2)
	for i = 0x17; i > 1; i-- {
		s[0].Frame.Set(i - 1)
		rndDraw()
	}
	freeAndNull(gCineImg2)
	img(0, *gCineImg0)
	s[0].Frame.Set(0)
	rndDraw()
	img(3, *gCineImg9)
	img(2, *gCineImg9)
	s[2].X.Set(0x25)
	s[2].Y.Set(0x16)
	s[2].Frame.Set(0)
	s[3].X.Set(0x62)
	s[3].Y.Set(0xaa)
	s[3].Frame.Set(4)
	img(2, *gCineImg5)
	voice(1, 6)
	s[2].Y.Set(0xdc)
	rndDraw()
	for i = 0; i < 2; i++ {
		freeAndNull(&gCineVoices[i])
	}
	freeAndNull(gCineImg0)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg5)
	freeAndNull(gCineImg6)
	freeAndNull(gCineImg7)
	freeAndNull(gCineImg9)
	sndStartMusic(*gMusicBuf)
	*gCineImg0 = loadGraphic("hud.vga")
	clear(0x26)
	hud(2)

	// the fight
	fight := []*Far{gCineImg0, gCineImg1, gCineImg2, gCineImg3, gCineImg4, gCineImg5, gCineImg6, gCineImg7, gCineImg8, gCineImg9}
	for k, n := range []string{"fca1.vga", "fca2.vga", "fcb1.vga", "fcb2.vga", "fba1.vga", "fba2.vga", "fbb1.vga", "fbb2.vga", "fda1.vga", "fdb1.vga"} {
		*fight[k] = loadGraphic(n)
	}
	for k := 0; k < 6; k++ {
		s[k].Flag.Set(-1)
	}
	for k := 0; k < 5; k++ {
		s[k].Attr.Set(0)
		s[k].Frame.Set(0)
	}
	s[0].X.Set(0)
	s[1].X.Set(0)
	s[2].X.Set(0)
	s[5].Img.Set(0)
	s[0].Y.Set(4)
	s[2].Y.Set(4)
	s[3].Y.Set(4)
	s[1].Y.Set(0x2b)
	s[4].X.Set(0x7a)
	s[4].Y.Set(0xa6)
	s[3].X.Set(0xd1)
	all5 := func(i int) {
		for k := 0; k < 5; k++ {
			s[k].Frame.Set(i)
		}
	}
	for k, f := range []int{0, 1, 4, 5, 8} {
		img(k, *fight[f])
	}
	showRawPicture("front.raw")
	playSfx(0x32)
	frames(0x10, all5)
	for k, f := range []int{2, 3, 6, 7, 9} {
		img(k, *fight[f])
	}
	frames(0x10, all5)
	for k := 9; k >= 0; k-- {
		freeAndNull(fight[k])
	}
	clear(0x26)

	// the taking
	*gCineImg4 = loadGraphic("takb2.vga")
	*gCineImg3 = loadGraphic("takb1.vga")
	*gCineImg0 = loadGraphic("taks1.vga")
	img(1, *gCineImg3)
	img(2, *gCineImg4)
	s[0].Y.Set(0)
	s[1].Y.Set(0)
	s[2].Y.Set(0)
	j, c4 = 0, 0
	s[0].X.Set(0x5c)
	s[1].X.Set(0x24)
	s[2].X.Set(0xee)
	showRawPicture("take.raw")
	img(0, *gCineImg0)
	take := func(i int) {
		s[0].Frame.Set(i)
		s[1].Frame.Set(j)
		s[2].Frame.Set(j)
		j++
	}
	frames(0x1d, take)
	freeAndNull(gCineImg0)
	*gCineImg1 = loadGraphic("taks2.vga")
	img(0, *gCineImg1)
	frames(0x18, take)
	freeAndNull(gCineImg1)
	*gCineImg2 = loadGraphic("taks3.vga")
	img(0, *gCineImg2)
	frames(0x17, func(i int) { s[0].Frame.Set(i); s[2].Frame.Set(j); j++ })
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg4)

	// to the top
	*gCineImg0 = loadGraphic("topa1.vga")
	*gCineImg5 = loadGraphic("topa2.vga")
	*gCineImg1 = loadGraphic("toph.vga")
	*gCineImg2 = loadGraphic("topi.vga")
	*gCineImg3 = loadGraphic("topj.vga")
	*gCineImg4 = loadGraphic("topk.vga")
	img(0, *gCineImg0)
	s[2].Img.Set(0)
	s[1].Frame.Set(0)
	s[1].Attr.Set(0)
	s[1].Y.Set(0)
	s[1].Img.Set(0)
	s[0].Frame.Set(0)
	s[0].Attr.Set(0)
	s[0].X.Set(0)
	j, c4 = 0, 0
	s[0].Y.Set(0x82)
	s[1].X.Set(0x87)
	showRawPicture("topsea.raw")
	playSfx(0x3b)
	walkOnly := func(int) { walk() }
	walk1 := func(i int) { walk(); s[1].Frame.Set(i) }
	frames(0xf, walkOnly)
	img(1, *gCineImg1)
	frames(0x10, walk1)
	img(1, *gCineImg2)
	frames(0x10, walk1)
	playSfx(0x37)
	img(1, *gCineImg3)
	frames(5, walk1)
	img(1, *gCineImg4)
	frames(6, walk1)
	s[1].Img.Set(0)
	frames(0xf, walkOnly)
	freeAndNull(gCineImg4)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg5)
	freeAndNull(gCineImg0)

	// lift-off
	*gCineImg0 = loadGraphic("lytp1.vga")
	*gCineImg1 = loadGraphic("lytp2.vga")
	*gCineImg2 = loadGraphic("lytp3.vga")
	*gCineImg3 = loadGraphic("lyts1.vga")
	*gCineImg4 = loadGraphic("lyts2.vga")
	*gCineImg5 = loadGraphic("lyts3.vga")
	*gCineImg6 = loadGraphic("lyts4.vga")
	s[2].Img.Set(0)
	s[1].Attr.Set(0)
	s[1].Y.Set(0)
	s[0].Frame.Set(0)
	s[0].Attr.Set(0)
	s[0].Y.Set(0)
	s[0].X.Set(0)
	s[1].X.Set(0x4b)
	img(0, *gCineImg0)
	img(1, *gCineImg3)
	fadeScreen(0, 0x40)
	showRawPicture("stars.raw")
	draw()
	fadeScreen(1, 0x40)
	both := func(i int) { s[0].Frame.Set(i); s[1].Frame.Set(i) }
	frames(0xd, both)
	img(0, *gCineImg1)
	img(1, *gCineImg4)
	frames(0xd, both)
	img(0, *gCineImg2)
	img(1, *gCineImg5)
	frames(0xd, func(i int) { both(i); s[0].Y.Set(s[0].Y.Get() - 7) })
	s[1].Y.Set(0x15)
	s[0].Frame.Set(0xd)
	img(1, *gCineImg6)
	frames(0xb, func(i int) { s[1].Frame.Set(i); s[0].Y.Set(s[0].Y.Get() - 7) })
	freeAndNull(gCineImg6)
	freeAndNull(gCineImg5)
	freeAndNull(gCineImg4)
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg0)
	*gCineImg7 = loadGraphic("lyts5.vga")
	*gCineImg8 = loadGraphic("lyts6.vga")
	*gCineImg9 = loadGraphic("lyts7.vga")
	img(0, *gCineImg7)
	s[0].Y.Set(200)
	s[0].Frame.Set(0x2d)
	set1 := func(i int) { s[1].Frame.Set(i) }
	img(1, *gCineImg7)
	frames(5, set1)
	img(1, *gCineImg8)
	frames(5, set1)
	img(1, *gCineImg9)
	frames(3, set1)
	freeAndNull(gCineImg9)
	freeAndNull(gCineImg8)
	freeAndNull(gCineImg7)
	*gCineImgLyts9 = loadGraphic("lyts9.vga")
	*gCineImgLyts8 = loadGraphic("lyts8.vga")
	img(0, *gCineImgLyts9)
	playSfx(10)
	img(1, *gCineImgLyts8)
	frames(8, set1)
	img(1, *gCineImgLyts9)
	for i = 0; i < 0x2e; i++ {
		s[1].Frame.Set(i)
		draw()
		if i == 0x25 {
			playSfx(0x14)
		}
	}
	freeAndNull(gCineImg9)
	freeAndNull(gCineImgLyts8)
	freeAndNull(gCineImgLyts9)

	// the score
	for i = 0; i < 10; i++ {
		digits[i] = 0
		for gScoreDigitWeights[i] <= score {
			score -= gScoreDigitWeights[i]
			digits[i]++
		}
	}
	clear(0x68)
	c4 = 0
	s[1].Y.Set(0x78)
	s[1].Frame.Set(1)
	s[2].X.Set(0x7a)
	s[2].Y.Set(0xa6)
	s[3].Y.Set(0x34)
	*gCineImg0 = loadGraphic("frnt.vga")
	*gCineImg1 = loadGraphic("fda1.vga")
	*gCineImg2 = loadGraphic("scor.vga")
	*gCineImg3 = loadGraphic("scon.vga")
	*gCineImg4 = loadGraphic("send.vga")
	img(1, *gCineImg0)
	img(0, *gCineImg0)
	img(2, *gCineImg1)
	fadeScreen(0, 0x40)
	draw2()
	fadeScreen(1, 0x40)
	for i = 0; i < 0x28; i++ {
		tick()
	}
	img(3, *gCineImg2)
	for i = 0; i < 0xe; i++ {
		s[3].Frame.Set(i)
		tick()
		tick()
	}
	if c4++; c4 > 12 {
		c4 = 0
	}
	s[2].Frame.Set(c4)
	s[3].Frame.Set(0xe)
	for i = 0; i < 10; i++ {
		img(i+4, *gCineImg3)
		w[i*7+31] = int16(i*0x13 + 0x76)
		w[i*7+32] = 0x58
	}
	draw2()
	for j = 0; j < 0x14; j++ {
		for i = (0x13 - j) / 2; i < 10; i++ {
			w[i*7+34] = int16(randomRange(0, 9))
		}
		tick()
	}
	c2 = 11
	j = 10
	for {
		for i = 0; i < j; i++ {
			w[i*7+34] = int16(randomRange(0, 9))
		}
		if c2--; c2 == 0 {
			w[(j-1)*7+34] = int16(digits[j-1])
			c2 = 10
		}
		if digits[j-1] == int(w[(j-1)*7+34]) {
			j--
		}
		tick()
		if j == 0 {
			break
		}
	}
	img(14, *gCineImg4)
	s[15].Img.Set(0)
	s[14].Frame.Set(0)
	s[14].Attr.Set(0)
	s[14].X.Set(0)
	s[14].Y.Set(0x6e)
	s[15].Flag.Set(-1)
	flushInput()
	waitKey(tick)
	flushInput()
	freeAndNull(gCineImg3)
	freeAndNull(gCineImg2)
	freeAndNull(gCineImg1)
	freeAndNull(gCineImg0)
	fadeScreen(0, 1)
	flushInput()
	return 0
}
