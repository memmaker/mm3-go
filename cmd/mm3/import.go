package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// importGame (-import, for goplay) finds the folder holding MM3.CC anywhere under src and
// copies what the game reads into dst: MM3.CC, MM3.CFG and any saved games (*.MM3, MM3.CUR).
// Without an MM3.CFG it writes one for VGA (the four bytes a GOG install has).
func importGame(src, dst string) error {
	var dir string
	filepath.WalkDir(src, func(p string, d fs.DirEntry, _ error) error {
		if d != nil && !d.IsDir() && strings.EqualFold(d.Name(), "MM3.CC") {
			dir = filepath.Dir(p)
			return fs.SkipAll
		}
		return nil
	})
	if dir == "" {
		return errors.New("no MM3.CC under " + src)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	cfg := false
	for _, e := range ents {
		n := strings.ToUpper(e.Name())
		if e.IsDir() || !(n == "MM3.CC" || n == "MM3.CFG" || n == "MM3.CUR" || filepath.Ext(n) == ".MM3") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
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
