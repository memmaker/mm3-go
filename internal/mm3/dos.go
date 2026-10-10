package mm3

// DOS and Borland runtime shims: the BIOS keyboard (bioskey), the C library's file calls
// (open/read/write/close on DOS handles, stdio FILEs, findfirst/findnext), sys.asm's
// one-handle file_* routines, the text-mode console (only used for fatal messages) and exit.
//
// Files: DOS kept everything in the game directory. Here a name is looked up (without regard
// to case) in SaveDir first, then GameDir; a file that is created or written goes to SaveDir.
// So MM3.CC and MM3.CFG are read from GameDir, and MM3.CUR, the saves and BOSS.$$$ live in
// SaveDir; nothing has to be copied on first run.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/memmaker/mm3-go/internal/host"
)

// bioskey is int 16h: 0 waits for a key and returns its word (scan<<8 | ascii), 1 returns the
// next key word without removing it (0: none), 2 the shift state.
func bioskey(cmd int) int {
	switch cmd {
	case 0:
		for len(host.KeyQueue) == 0 {
			host.Yield()
		}
		k := host.KeyQueue[0]
		host.KeyQueue = host.KeyQueue[1:]
		return int(k)
	case 1:
		if len(host.KeyQueue) == 0 {
			host.Busy()
			return 0
		}
		return int(host.KeyQueue[0])
	case 2:
		return host.Shift
	}
	return 0
}

// dosPath resolves a DOS file name: SaveDir if it is there (or for writing), else GameDir.
func dosPath(name string, write bool) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	p := gamePath(SaveDir, name)
	if write {
		return p
	}
	if _, err := os.Stat(p); err == nil {
		return p
	}
	if q := gamePath(GameDir, name); fileExists(q) {
		return q
	}
	return p
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// Borland fcntl.h flags.
const (
	oRdonly = 1
	oWronly = 2
	oRdwr   = 4
	oCreat  = 0x100
	oTrunc  = 0x200
)

var (
	dosFiles = map[int]*os.File{}
	dosNext  = 5 // 0..4 are the standard handles
)

func dosOpen(name string, flags int) *os.File {
	write := flags&(oWronly|oRdwr|oCreat|oTrunc) != 0
	p := dosPath(name, write)
	f := 0
	switch {
	case flags&oRdwr != 0:
		f = os.O_RDWR
	case flags&oWronly != 0:
		f = os.O_WRONLY
	}
	if flags&oCreat != 0 {
		f |= os.O_CREATE
	}
	if flags&oTrunc != 0 {
		f |= os.O_TRUNC
	}
	h, err := os.OpenFile(p, f, 0o644)
	if err != nil && f&(os.O_RDWR|os.O_WRONLY) == os.O_RDWR && flags&oCreat == 0 {
		h, err = os.Open(p) // a read-only game directory
	}
	if err != nil {
		return nil
	}
	return h
}

// open is io.h open: a DOS handle, or -1. mode (permissions) is ignored.
func open(name []byte, flags int, mode ...int) int {
	h := dosOpen(gostr(name), flags)
	if h == nil {
		return -1
	}
	fd := dosNext
	dosNext++
	dosFiles[fd] = h
	return fd
}

// read is io.h read: up to n bytes into buf, the count read, -1 on a bad handle.
func read(fd int, buf []byte, n int) int {
	h := dosFiles[fd]
	if h == nil {
		return -1
	}
	r, _ := io.ReadFull(h, buf[:n])
	return r
}

func close(fd int) int {
	h := dosFiles[fd]
	if h == nil {
		return -1
	}
	delete(dosFiles, fd)
	h.Close()
	return 0
}

// sys.asm's file_open .. file_write: one open file (cs:0140), opened read/write; each
// returns 0 on success, nonzero (the DOS error) on failure.
var sysFile *os.File

func fileOpen(name []byte) int {
	sysFile = dosOpen(gostr(name), oRdwr)
	if sysFile == nil {
		return 2
	}
	return 0
}

func fileClose() int {
	if sysFile == nil {
		return 6
	}
	err := sysFile.Close()
	sysFile = nil
	if err != nil {
		return 6
	}
	return 0
}

func fileSeek(off uint32) int {
	if sysFile == nil {
		return 6
	}
	if _, err := sysFile.Seek(int64(off), io.SeekStart); err != nil {
		return 1
	}
	return 0
}

// fileRead and fileWrite move n bytes (DOS reports a short count as success; so does this).
func fileRead(buf []byte, n int) int {
	if sysFile == nil {
		return 6
	}
	if _, err := sysFile.Read(buf[:n]); err != nil && err != io.EOF {
		return 5
	}
	return 0
}

func fileWrite(buf []byte, n int) int {
	if sysFile == nil {
		return 6
	}
	if _, err := sysFile.Write(buf[:n]); err != nil {
		return 5
	}
	return 0
}

// stdio, as copy_work_to_save uses it: binary files.
func fopen(name []byte, mode string) *os.File {
	if strings.HasPrefix(mode, "w") {
		return dosOpen(gostr(name), oWronly|oCreat|oTrunc)
	}
	return dosOpen(gostr(name), oRdonly)
}

func fread(buf []byte, size, n int, f *os.File) int {
	r, _ := io.ReadFull(f, buf[:size*n])
	return r / size
}

func fwrite(buf []byte, size, n int, f *os.File) int {
	r, _ := f.Write(buf[:size*n])
	return r / size
}

func fseek(f *os.File, off int32, whence int) int {
	if _, err := f.Seek(int64(off), whence); err != nil {
		return -1
	}
	return 0
}

func ftell(f *os.File) int32 { o, _ := f.Seek(0, io.SeekCurrent); return int32(o) }
func rewind(f *os.File)      { f.Seek(0, io.SeekStart) }
func fclose(f *os.File) int  { f.Close(); return 0 }

// ffblk is dir.h's find block: only the name is used.
type ffblk struct {
	ffName [13]byte
	rest   []string
}

// findfirst lists the files matching the DOS pattern in SaveDir and GameDir (a name once).
func findfirst(pat []byte, ff *ffblk, attr int) int {
	p := strings.ToUpper(gostr(pat))
	seen := map[string]bool{}
	ff.rest = nil
	for _, dir := range []string{SaveDir, GameDir} {
		ents, _ := os.ReadDir(dir)
		for _, e := range ents {
			n := strings.ToUpper(e.Name())
			if ok, _ := filepath.Match(p, n); ok && !e.IsDir() && len(n) <= 12 && !seen[n] {
				seen[n] = true
				ff.rest = append(ff.rest, n)
			}
		}
	}
	sort.Strings(ff.rest)
	return findnext(ff)
}

func findnext(ff *ffblk) int {
	if len(ff.rest) == 0 {
		return -1
	}
	strcpy(ff.ffName[:], cs(ff.rest[0]))
	ff.rest = ff.rest[1:]
	return 0
}

// system runs the one command the game uses: "del name".
func system(cmd string) int {
	if f := strings.Fields(cmd); len(f) == 2 && strings.EqualFold(f[0], "del") {
		os.Remove(dosPath(f[1], true))
	}
	return 0
}

// time and ctime: the host's local clock.
func timeNow() int32 { return int32(time.Now().Unix()) }

func ctime(t int32) string { return time.Unix(int64(t), 0).Format("Mon Jan _2 15:04:05 2006\n") }

// allocmem is dos.h allocmem asking for more than there is: it returns the largest block
// free, in paragraphs. The port has plenty: 640K's worth.
func allocmem(size uint16, seg *uint16) int { return 0xa000 }

// The text-mode console. The game uses it only for its fatal messages before exiting, which
// go to stderr; window/colour/cursor calls do nothing.
var directvideo int

func cputs(s string)                      { fmt.Fprint(os.Stderr, strings.ReplaceAll(s, "\r", "")) }
func cprintf(s []byte)                    { cputs(gostr(s)) }
func putchar(c int)                       {}
func textmode(m int)                      {}
func textcolor(c int)                     {}
func textbackground(c int)                {}
func clrscr()                             {}
func gotoxy(x, y int)                     {}
func window(x1, y1, x2, y2 int)           {}
func setcursortype(t int)                 {}
func installBreakHandlers()               {}
func restoreBreakHandlers()               {}
func loadVideoDriver(mode int, timer Far) { vgaInit() } // sys.asm: load vga.drv and call its init

// loadCCFile is sys.asm load_cc_file: the file from the archive in a new far block.
func loadCCFile(name string) Far { return newSeg(ccLoad(name)) }

// exit ends the program (exit(), int 21h 4ch).
func exit(code int) { host.Exit(code) }
