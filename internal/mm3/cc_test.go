package mm3

import (
	"crypto/md5"
	"fmt"
	"os"
	"testing"
)

// The archive reader against tools/cc.py of the decompilation (md5 of a few files).
func TestCC(t *testing.T) {
	dir := os.Getenv("MM3_DATA")
	if dir == "" {
		t.Skip("MM3_DATA (the game directory) not set")
	}
	GameDir = dir
	openCCArchive()
	for name, sum := range ccSums {
		if got := fmt.Sprintf("%x", md5.Sum(ccLoad(name))); got != sum {
			t.Errorf("%s: md5 %s, want %s", name, got, sum)
		}
	}
}

var ccSums = map[string]string{
	"vga":       "b52ebac267e25ac8989de3d2e7295318",
	"timer.drv": "9a514c140cbc49991695063684e53c1e",
	"main.icn":  "acd17ddd2712732e8f66f2632d90bc3e",
	"back.raw":  "eee72e319c20d4088ed50f1e0e6ce798",
}
