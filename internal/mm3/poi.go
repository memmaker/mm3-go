package mm3

import (
	"encoding/json"
	"fmt"
	"os"
)

// Map notes: with a Cartographer in the party, a click on a mapped automap cell asks for a
// label and a right click removes it; hovering shows it. They live in <work archive>.POI,
// copied along with it to and from the save slots.

func poiPath(archive []byte, write bool) string { return dosPath(gostr(archive)+".poi", write) }

func loadPOIs() map[string]string {
	m := map[string]string{}
	if d, err := os.ReadFile(poiPath(str(*gStrWorkArchive), false)); err == nil {
		json.Unmarshal(d, &m)
	}
	return m
}

func savePOIs(m map[string]string) {
	d, _ := json.Marshal(m)
	os.WriteFile(poiPath(str(*gStrWorkArchive), true), d, 0o644)
}

// copyPOIs follows copyFile: the notes go with the archive (none when it has none).
func copyPOIs(from, to []byte) {
	if d, err := os.ReadFile(poiPath(from, false)); err == nil {
		os.WriteFile(poiPath(to, true), d, 0o644)
	} else {
		os.Remove(poiPath(to, true))
	}
}

// automapPOIs runs the open automap: draws the note markers, shows a hovered note's label in
// place of the coordinate line, and waits for a key or click. True when the map should be
// reopened (a note was edited).
func automapPOIs(ox, oy int, seen *[16][16]byte, line []byte) bool {
	notes := loadPOIs()
	carto := partyHasSkill(4) != 0
	px, py := int(*gPartyX), int(*gPartyY)
	key := func(i, j int) string { // "" off the map or on an unmapped cell
		if i < 0 || j < 0 || i > 15 || j > 15 || seen[i][j] == 0 {
			return ""
		}
		x, y, idx, ok := cellQuadrant(px+ox+1+i, py+oy-1-j)
		if !ok {
			return ""
		}
		return fmt.Sprintf("%d:%d:%d", gQuadrantMaps[idx], x&15, y&15)
	}
	cursorHide() // the cursor is drawn into the offscreen: lift it off before drawing under it
	for j := 0; j < 16; j++ {
		for i := 0; i < 16; i++ {
			if notes[key(i, j)] == "" {
				continue
			}
			for d := 0; d < 9; d++ {
				put(rowOff(uint16(0x1c+10+j*8+2+d/3)<<1)+uint16(0x50+i*10+3+d%3), glyphCol[1])
			}
		}
	}
	flushText(true)
	flushInput()
	shown := ""
	for {
		b := getMouseState(gMouseX, gMouseY)
		if bioskey(1) != 0 {
			bioskey(0)
			return false
		}
		mx, my := int(*gMouseX)-0x50, int(*gMouseY)-0x1c-10
		k := ""
		if mx >= 0 && my >= 0 {
			k = key(mx/10, my/8)
		}
		if label := notes[k]; label != shown {
			shown = label
			var buf [64]byte
			if label != "" {
				sprintf(buf[:], "\x03l\v139\t000\x04160\x03c%s", cs(label))
				printWindowText(buf[:])
			} else {
				sprintf(buf[:], "\x03l\v139\t000\x04160%s", line)
				printWindowText(buf[:])
			}
		}
		if b == 0 {
			continue
		}
		waitMouseRelease()
		switch {
		case b&1 != 0 && carto && k != "":
			var buf [21]byte
			openTextWindow(0x58, 0x50, 0xa0, 0x26, 0, 0x96, 4, 6, 0)
			printWindowText(cs("\x03cNote:\n"))
			if inputText(buf[:], 20, 0x90) != 0 {
				notes[k] = gostr(buf[:])
				savePOIs(notes)
			}
			closeWindows(1)
			return true
		case b&2 != 0 && notes[k] != "":
			delete(notes, k)
			savePOIs(notes)
			return true
		}
		return false
	}
}
