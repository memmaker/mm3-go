// Port of MAIN.C, the NOPROTECT build: main() without the copy protection, the EXE
// self-check and the vga.drv marker block; the same start-up, then the game loop.
package mm3

// Main is main(). The command line (the C's 'C' and 'S' switches, and "any argument: print
// the version and exit") is the cmd's business; the overlay manager's EMS/XMS message is gone.
func Main() {
	var timerDrv Far

	checkConfigAndDisk()
	installBreakHandlers()
	resetSessionTimer()
	*gArchiveOpen = 0
	switch *gCfgVideo {
	case 0:
		*gVideoMode = 0
	case 1:
		*gVideoMode = 4
	case 2:
		*gVideoMode = 3
	case 3:
		*gVideoMode = 1
	case 4:
		*gVideoMode = 5
	default:
		fatalMissingFile(0)
	}
	switch *gCfgSound {
	case 0:
		*gSndDev0 = 1
	case 1:
		*gSndDev1 = 1
	case 2:
		*gSndDev2 = 1
	case 3:
		*gSndDev3 = 1
	case 4:
		*gSndDev4 = 1
	case 5:
		*gSndDev5 = 1
	case 6:
		*gSndDev6 = 1
	default:
		fatalMissingFile(0)
	}
	openCCArchive()
	timerDrv = loadCCFile("timer.drv")
	loadVideoDriver(int(*gVideoMode), timerDrv)
	*gMusicEnabled = 0
	if *gSndDev6 != 0 || *gSndDev0 != 0 || *gSndDev2 != 0 || *gSndDev1 != 0 || *gSndDev4 != 0 || *gSndDev3 != 0 {
		*gMusicEnabled = 1
	}
	if *gNoSound != 0 {
		*gMusicEnabled = 0
	} else {
		loadSoundDriver(int(*gCfgSound))
		sndInit(int(*gCfgSoundPortLo) + int(*gCfgSoundPortHi)<<8)
	}
	*gMusicBuf = memAlloc(0x1890)
	*gSfxOn = 1
	*gMusicOn = 1
	playIntro()
	if *gCfgVideo == 4 {
		showRawPicture("logy5.raw")
	} else {
		showRawPicture("control.raw")
	}
	fadeScreen(1, 4)
	*gPortraitClickOn = 0
	*gRosterFar = 0
	*gMapText = 0
	*gMazeScript = 0
	*gMouseCursor = int32(loadGraphic("mouse.icn"))
	loadMonsterTables()
	strcpy(gCurMusic, cs("mm3theme.m"))
	flushInput()
	freeGraphic(timerDrv)
	again := 0
	setMouseCursor(0, uint32(*gMouseCursor))
	for {
		optionsMenu(again)
		again = 1
		for {
			innRosterScreen()
			if gameMainLoop() != 2 {
				break
			}
		}
	}
}
