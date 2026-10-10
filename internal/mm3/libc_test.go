package mm3

import "testing"

func TestSprintf(t *testing.T) {
	for _, c := range []struct {
		f    string
		args []any
		want string
	}{
		{"%s has %u gold", []any{cs("Joe"), -1}, "Joe has 65535 gold"},
		{"%02u:%2d|%+d|%lu|%c|%3u|%03d", []any{5, 7, 3, uint32(100000), 'A', 12, -5}, "05: 7|+3|100000|A| 12|-05"},
		{"%-5s|%5s", []any{"ab", cs("cd")}, "ab   |   cd"},
	} {
		if got := string(csprintf(c.f, c.args...)); got != c.want {
			t.Errorf("%q: %q, want %q", c.f, got, c.want)
		}
	}
}
