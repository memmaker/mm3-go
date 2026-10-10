package mm3

import (
	"bytes"
	"strconv"
	"strings"
)

// What the game takes from Borland C++ 2.0's runtime library. C strings are NUL-terminated
// []byte (a ds view, a far block's bytes, or cs("literal")); ints are 16 bits.

func strlen(s []byte) int {
	if i := bytes.IndexByte(s, 0); i >= 0 {
		return i
	}
	return len(s)
}

// strcpy copies the C string src (with its NUL) to dst.
func strcpy(dst, src []byte) { n := strlen(src); copy(dst[:n], src); dst[n] = 0 }

// strncpy is _fstrncpy/strncpy: n bytes, NUL-padded, no NUL added when src is longer.
func strncpy(dst, src []byte, n int) {
	l := strlen(src)
	for i := 0; i < n; i++ {
		if i < l {
			dst[i] = src[i]
		} else {
			dst[i] = 0
		}
	}
}

func strcat(dst, src []byte) { strcpy(dst[strlen(dst):], src) }

func strcmp(a, b []byte) int { return bytes.Compare(a[:strlen(a)], b[:strlen(b)]) }

func stricmp(a, b []byte) int {
	return strings.Compare(strings.ToLower(gostr(a)), strings.ToLower(gostr(b)))
}

func memset(dst []byte, c byte, n int) {
	for i := range dst[:n] {
		dst[i] = c
	}
}

func memcpy(dst, src []byte, n int) { copy(dst[:n], src[:n]) }

func toupper(c int) int {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

func atol(s []byte) int32 {
	t := strings.TrimLeft(gostr(s), " \t")
	n := 0
	for n < len(t) && (t[n] >= '0' && t[n] <= '9' || n == 0 && (t[n] == '-' || t[n] == '+')) {
		n++
	}
	v, _ := strconv.ParseInt(t[:n], 10, 64)
	return int32(v)
}

// ultoa writes v in the given radix to dst as a C string.
func ultoa(v uint32, dst []byte, radix int) { strcpy(dst, cs(strconv.FormatUint(uint64(v), radix))) }

// sprintf is Borland's: %d %u %x %c %s with flags and width, l for longs. Without l an
// argument is a 16-bit int (%d signed, %u unsigned), as the C pushed it. %s takes a C string
// ([]byte) or a Go string. It writes the result and its NUL to dst and returns the length.
func sprintf(dst []byte, format string, args ...any) int {
	out := csprintf(format, args...)
	copy(dst, out)
	dst[len(out)] = 0
	return len(out)
}

func csprintf(format string, args ...any) []byte {
	var out []byte
	ai := 0
	next := func() any {
		if ai >= len(args) {
			return 0
		}
		ai++
		return args[ai-1]
	}
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' {
			out = append(out, c)
			continue
		}
		j := i + 1
		for j < len(format) && strings.IndexByte("-+ 0#", format[j]) >= 0 {
			j++
		}
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			j++
		}
		long := false
		if j < len(format) && (format[j] == 'l' || format[j] == 'h') {
			long = format[j] == 'l'
			j++
		}
		if j >= len(format) {
			break
		}
		spec, verb := format[i+1:j], format[j]
		spec = strings.TrimSuffix(strings.TrimSuffix(spec, "l"), "h")
		i = j
		var s string
		switch verb {
		case '%':
			out = append(out, '%')
			continue
		case 'd', 'i':
			v := toInt64(next())
			if long {
				v = int64(int32(v))
			} else {
				v = int64(int16(v))
			}
			s = fmtNum(spec, strconv.FormatInt(v, 10))
		case 'u':
			v := toInt64(next())
			if long {
				v = int64(uint32(v))
			} else {
				v = int64(uint16(v))
			}
			s = fmtNum(spec, strconv.FormatInt(v, 10))
		case 'x', 'X':
			v := toInt64(next())
			if long {
				v = int64(uint32(v))
			} else {
				v = int64(uint16(v))
			}
			s = strconv.FormatInt(v, 16)
			if verb == 'X' {
				s = strings.ToUpper(s)
			}
			s = fmtNum(spec, s)
		case 'c':
			s = fmtStr(spec, string([]byte{byte(toInt64(next()))}))
		case 's':
			switch v := next().(type) {
			case string:
				s = fmtStr(spec, v)
			case []byte:
				s = fmtStr(spec, gostr(v))
			}
		default:
			s = "%" + spec + string(verb)
		}
		out = append(out, s...)
	}
	return out
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint:
		return int64(x)
	case Far:
		return int64(x)
	}
	panic("sprintf: not a number")
}

// fmtNum applies flags and width (-, +, space, 0) to a formatted number.
func fmtNum(spec, digits string) string {
	left, plus, space, zero := false, false, false, false
	k := 0
	for ; k < len(spec) && strings.IndexByte("-+ 0#", spec[k]) >= 0; k++ {
		switch spec[k] {
		case '-':
			left = true
		case '+':
			plus = true
		case ' ':
			space = true
		case '0':
			zero = true
		}
	}
	w, _ := strconv.Atoi(spec[k:])
	neg := strings.HasPrefix(digits, "-")
	if neg {
		digits = digits[1:]
	}
	sign := ""
	switch {
	case neg:
		sign = "-"
	case plus:
		sign = "+"
	case space:
		sign = " "
	}
	pad := w - len(sign) - len(digits)
	if pad <= 0 {
		return sign + digits
	}
	if left {
		return sign + digits + strings.Repeat(" ", pad)
	}
	if zero {
		return sign + strings.Repeat("0", pad) + digits
	}
	return strings.Repeat(" ", pad) + sign + digits
}

func fmtStr(spec, s string) string {
	left := strings.HasPrefix(spec, "-")
	w, _ := strconv.Atoi(strings.TrimLeft(spec, "-+ 0#"))
	if pad := w - len(s); pad > 0 {
		if left {
			return s + strings.Repeat(" ", pad)
		}
		return strings.Repeat(" ", pad) + s
	}
	return s
}
