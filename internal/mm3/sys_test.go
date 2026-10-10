package mm3

import "testing"

func TestRandomRange(t *testing.T) {
	if got := randomRange(7, 7); got != 7 {
		t.Fatalf("randomRange(7,7) = %d", got)
	}
	for i := 0; i < 1000; i++ {
		if v := randomRange(1, 20); v < 1 || v > 20 {
			t.Fatalf("randomRange(1,20) = %d", v)
		}
	}
	// sbb/adc carry chain: 0 - 1 - 1 borrows, 0xffff + 0 + 1 carries.
	if v, c := sbb16(0, 1, true); v != 0xfffe || !c {
		t.Fatal("sbb16")
	}
	if v, c := adc16(0xffff, 0, true); v != 0 || !c {
		t.Fatal("adc16")
	}
}
