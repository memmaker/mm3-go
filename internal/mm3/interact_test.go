package mm3

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEventNeedsInteract classifies known cells of the original data (MM3_DATA: a folder with
// MM3.CC and the shipped MM3.CUR; skipped without it). loadMaze writes to a copy of MM3.CUR.
func TestEventNeedsInteract(t *testing.T) {
	dir := os.Getenv("MM3_DATA")
	if dir == "" {
		t.Skip("MM3_DATA not set")
	}
	cur, err := os.ReadFile(gamePath(dir, "mm3.cur"))
	if err != nil { // GOG keeps the shipped one here
		cur, err = os.ReadFile(gamePath(filepath.Join(dir, "__support", "save"), "mm3.cur"))
	}
	if err != nil {
		t.Skip(err)
	}
	tmp := filepath.Join(t.TempDir(), "MM3.CUR")
	os.WriteFile(tmp, cur, 0o644)
	if ccMain, err = openCC(gamePath(dir, "mm3.cc")); err != nil {
		t.Skip(err)
	}
	ccCur, _ = openCC(tmp)
	*gMonMaxHp = loadCCFile("MonHP.dat")
	for _, c := range []struct {
		maze, x, y int
		want       bool
		what       string
	}{
		{0, 14, 14, true, "wooden chest (asks to open)"},
		{0, 1, 10, true, "mirror portal (asks)"},
		{56, 5, 12, false, "items buried in snow (asks, nothing to see)"},
		{0, 1, 5, false, "building door"},
		{0, 2, 5, false, "door caption"},
		{8, 15, 1, false, "trap"},
		{13, 28, 16, false, "treasure pile (only gives)"},
	} {
		*gMazeScript, *gMapQuadrant = 0, 0
		loadMazeData(c.maze)
		loadMaze(c.maze)
		var e *Evt
		for i := 0; i < int(*gEventCount); i++ {
			if g := &gEvents[i]; g.X.Get() == c.x && g.Y.Get() == c.y && g.Line.Get() == 0 {
				e = g // the party faces the event's way
			}
		}
		if e == nil {
			t.Fatalf("maze %d (%d,%d): no event", c.maze, c.x, c.y)
		}
		*gPartyDir = uint8(e.Dir.Get() & 3)
		if got := eventWaitsForSpace(e); got != c.want {
			t.Errorf("maze %d (%d,%d) %s: waits for Space %v, want %v", c.maze, c.x, c.y, c.what, got, c.want)
		}
	}
}
