package mm3

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The .CC archives (sys.asm open_cc_archive, cc_load_file_raw, cur_write_file). MM3.CC holds
// the game's data, each file LZHUF-compressed; MM3.CUR is the game in progress, its files
// stored plain. Files are found by a 16-bit hash of their name. Names starting with "MAZE"
// come from MM3.CUR, everything else from MM3.CC.

// GameDir is the game's directory (with MM3.CC); SaveDir holds MM3.CUR and the saved games,
// which DOS kept in the game directory.
var GameDir, SaveDir string

type ccEntry struct{ off, size int }

type ccArchive struct {
	path  string
	data  []byte // the whole file (MM3.CUR is written back in place)
	index map[uint16]ccEntry
}

var ccMain, ccCur *ccArchive

// ccHash is the archive's name hash: upper-cased 7-bit characters, rotate and add.
func ccHash(name string) uint16 {
	var x uint16
	for i := 0; i < len(name); i++ {
		c := uint16(name[i] & 0x7f)
		if c >= 0x60 {
			c -= 0x20
		}
		x = x<<8 | x>>8
		x = (x<<1 | x>>15) + c
	}
	return x
}

// gamePath finds name in dir without regard to case (DOS file names); a file that does not
// exist keeps the name given, upper-cased.
func gamePath(dir, name string) string {
	if ents, err := os.ReadDir(dir); err == nil {
		for _, e := range ents {
			if strings.EqualFold(e.Name(), name) {
				return filepath.Join(dir, e.Name())
			}
		}
	}
	return filepath.Join(dir, strings.ToUpper(name))
}

func openCC(path string) (*ccArchive, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(d) < 2 {
		return nil, fmt.Errorf("%s: not a CC archive", path)
	}
	n := int(le16(d, 0))
	if 2+8*n > len(d) {
		return nil, fmt.Errorf("%s: not a CC archive", path)
	}
	// the index is encrypted: rol 2, then add a key that starts at 0xac and grows by 0x67
	b := make([]byte, 8*n)
	k := byte(0xac)
	for i := range b {
		c := d[2+i]
		b[i] = (c<<2 | c>>6) + k
		k += 0x67
	}
	a := &ccArchive{path: path, data: d, index: map[uint16]ccEntry{}}
	for i := 0; i < len(b); i += 8 {
		a.index[le16(b, i)] = ccEntry{int(b[i+2]) | int(b[i+3])<<8 | int(b[i+4])<<16, int(le16(b, i+5))}
	}
	return a, nil
}

// openCCArchive is open_cc_archive: MM3.CC, read once.
func openCCArchive() {
	a, err := openCC(gamePath(GameDir, "mm3.cc"))
	if err != nil {
		fatal(err.Error())
	}
	ccMain = a
}

// openCurArchive is open_cur_archive (and reopen_cur_archive): MM3.CUR, re-read from disk.
func openCurArchive() {
	a, err := openCC(gamePath(SaveDir, "mm3.cur"))
	if err != nil {
		fatalDiskError()
	}
	ccCur = a
}

func closeCurArchive() {}

func (a *ccArchive) entry(name string) ccEntry {
	e, ok := a.index[ccHash(name)]
	if !ok || e.off+e.size > len(a.data) {
		fatalDiskError()
	}
	return e
}

// ccLoad is cc_load_file_raw: the file's contents (decompressed), its size in *gFileSize.
func ccLoad(name string) []byte {
	if len(name) >= 4 && strings.EqualFold(name[:4], "maze") {
		if ccCur == nil {
			openCurArchive()
		}
		e := ccCur.entry(name)
		*gFileSize = int16(e.size)
		return append([]byte(nil), ccCur.data[e.off:e.off+e.size]...)
	}
	e := ccMain.entry(name)
	out := lzhuf(ccMain.data[e.off : e.off+e.size])
	*gFileSize = int16(len(out))
	return out
}

// curWriteFile is cur_write_file: data overwrites the file's bytes in MM3.CUR (its size
// is the entry's, as in the original).
func curWriteFile(name string, data []byte) {
	if ccCur == nil {
		openCurArchive()
	}
	e := ccCur.entry(name)
	copy(ccCur.data[e.off:e.off+e.size], data)
	if err := os.WriteFile(ccCur.path, ccCur.data, 0o644); err != nil {
		fatalDiskError()
	}
}

// lzhuf is the archive's compression: Okumura's LZHUF (ring 4096, lookahead 60). The header
// is the word the ring buffer is filled with, then the length, big-endian.
func lzhuf(d []byte) []byte {
	const (
		N  = 4096
		F  = 60
		TH = 2
		NC = 256 - TH + F
		T  = NC*2 - 1
		R  = T - 1
	)
	var freq [T + 1]uint16
	var son [T]int
	var prnt [T + NC]int
	for i := 0; i < NC; i++ {
		freq[i] = 1
		son[i] = i + T
		prnt[i+T] = i
	}
	for i, j := 0, NC; j <= R; i, j = i+2, j+1 {
		freq[j] = freq[i] + freq[i+1]
		son[j] = i
		prnt[i], prnt[i+1] = j, j
	}
	freq[T] = 0xffff
	prnt[R] = 0
	reconst := func() {
		j := 0
		for i := 0; i < T; i++ {
			if son[i] >= T {
				freq[j] = (freq[i] + 1) / 2
				son[j] = son[i]
				j++
			}
		}
		for i, j := 0, NC; j < T; i, j = i+2, j+1 {
			f := freq[i] + freq[i+1]
			freq[j] = f
			k := j - 1
			for f < freq[k] {
				k--
			}
			k++
			copy(freq[k+1:j+1], freq[k:j])
			freq[k] = f
			copy(son[k+1:j+1], son[k:j])
			son[k] = i
		}
		for i := 0; i < T; i++ {
			k := son[i]
			prnt[k] = i
			if k < T {
				prnt[k+1] = i
			}
		}
	}
	update := func(c int) {
		if freq[R] == 0x8000 {
			reconst()
		}
		c = prnt[c+T]
		for {
			freq[c]++
			k := freq[c]
			l := c + 1
			if k > freq[l] {
				for k > freq[l+1] {
					l++
				}
				freq[c] = freq[l]
				freq[l] = k
				i := son[c]
				prnt[i] = l
				if i < T {
					prnt[i+1] = l
				}
				j := son[l]
				son[l] = i
				prnt[j] = c
				if j < T {
					prnt[j+1] = c
				}
				son[c] = j
				c = l
			}
			c = prnt[c]
			if c == 0 {
				break
			}
		}
	}
	pos := 32 // bit position, after the 4-byte header
	bit := func() int {
		i, sh := pos>>3, 7-pos&7
		pos++
		if i >= len(d) {
			return 0
		}
		return int(d[i]>>sh) & 1
	}
	byt := func() int {
		v := 0
		for k := 0; k < 8; k++ {
			v = v<<1 | bit()
		}
		return v
	}
	var dl, dc [256]int
	{
		i := 0
		for _, g := range []struct{ n, l int }{{32, 3}, {48, 4}, {64, 5}, {48, 6}, {48, 7}, {16, 8}} {
			for k := 0; k < g.n; k++ {
				dl[i] = g.l
				i++
			}
		}
		i = 0
		// d_code: 0 x32, 1 x16, 2 x16, 3 x16, then 4..11 x8 each, 12..23 x4 each, 24..47 x2 each, 48..63 x1
		for k := 0; k < 32; k++ {
			dc[i] = 0
			i++
		}
		for c := 1; c <= 3; c++ {
			for k := 0; k < 16; k++ {
				dc[i] = c
				i++
			}
		}
		for _, g := range [][3]int{{4, 12, 8}, {12, 24, 4}, {24, 48, 2}, {48, 64, 1}} {
			for c := g[0]; c < g[1]; c++ {
				for k := 0; k < g[2]; k++ {
					dc[i] = c
					i++
				}
			}
		}
	}
	n := int(d[2])<<8 | int(d[3])
	var ring [N]byte
	for i := 0; i < N; i += 2 {
		ring[i], ring[i+1] = d[0], d[1]
	}
	r := N - F
	out := make([]byte, 0, n)
	for len(out) < n {
		c := son[R]
		for c < T {
			c = son[c+bit()]
		}
		c -= T
		update(c)
		if c < 256 {
			out = append(out, byte(c))
			ring[r] = byte(c)
			r = (r + 1) & (N - 1)
			continue
		}
		i := byt()
		hi := dc[i] << 6
		for k := dl[i] - 2; k > 0; k-- {
			i = i<<1 | bit()
		}
		i = (r - (hi | i&0x3f) - 1) & (N - 1)
		for k := 0; k < c-255+TH; k++ {
			c2 := ring[(i+k)&(N-1)]
			out = append(out, c2)
			ring[r] = c2
			r = (r + 1) & (N - 1)
		}
	}
	return out[:n]
}

// ReadCC returns the file name from the MM3.CC in dir, decompressed (for the tools).
func ReadCC(dir, name string) ([]byte, error) {
	a, err := openCC(gamePath(dir, "mm3.cc"))
	if err != nil {
		return nil, err
	}
	e, ok := a.index[ccHash(name)]
	if !ok || e.off+e.size > len(a.data) {
		return nil, fmt.Errorf("%s: no %s", a.path, name)
	}
	return lzhuf(a.data[e.off : e.off+e.size]), nil
}
