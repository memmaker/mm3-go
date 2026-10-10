package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/memmaker/mm3-go/internal/gameimport"
)

// importGame (-import, for goplay) finds MM3.CC in whatever src holds (folders, loose files,
// zips, CD images, GOG installers; see gameimport) and copies what the game reads from the
// folder it lies in into dst, upper-case: MM3.CC, MM3.CFG and any saved games (*.MM3,
// MM3.CUR). Without an MM3.CFG it writes one for VGA (the four bytes a GOG install has).
func importGame(src, dst string) error {
	files, err := gameimport.Find(src, func(name string) bool {
		n := strings.ToUpper(path.Base(name))
		return n == "MM3.CC" || n == "MM3.CFG" || n == "MM3.CUR" || path.Ext(n) == ".MM3"
	})
	if err != nil {
		return err
	}
	dir := "\x00"
	for name := range files {
		if strings.EqualFold(path.Base(name), "MM3.CC") && (dir == "\x00" || name < dir) {
			dir = path.Dir(name) // the first in name order, when there are several
		}
	}
	if dir == "\x00" {
		return errors.New("no MM3.CC in " + src)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	cfg := false
	for name, b := range files {
		if path.Dir(name) != dir {
			continue
		}
		n := strings.ToUpper(path.Base(name))
		if err := os.WriteFile(filepath.Join(dst, n), b, 0o644); err != nil {
			return err
		}
		cfg = cfg || n == "MM3.CFG"
	}
	if !cfg {
		return os.WriteFile(filepath.Join(dst, "MM3.CFG"), []byte{0, 1, 0x20, 2}, 0o644)
	}
	return nil
}
