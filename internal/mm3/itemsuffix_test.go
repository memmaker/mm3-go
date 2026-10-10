package mm3

import "testing"

func TestItemStatSuffix(t *testing.T) {
	c := &Chr{}
	for id := 1; id < 0x4a; id++ {
		c.ItemID[0], c.ItemMat[0] = uint8(id), 0
		s := string(itemStatSuffix(c, 0))
		s = s[:len(s)-1]
		switch {
		case id < 0x22 && (len(s) < 6 || s[:2] != " ("):
			t.Errorf("weapon %#x: %q", id, s)
		case id >= 0x22 && gItemAcValues[id] != 0 && (len(s) < 4 || s[:2] != " ["):
			t.Errorf("armor %#x: %q", id, s)
		}
	}
	c.ItemID[0] = 0x50
	if s := itemStatSuffix(c, 0); s[0] != 0 {
		t.Errorf("misc item got %q", s)
	}
}
