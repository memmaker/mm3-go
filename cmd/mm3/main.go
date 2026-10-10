// Command mm3 runs Might and Magic III on Ebiten from the original game files.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/memmaker/mm3-go/internal/host"
	"github.com/memmaker/mm3-go/internal/mm3"
)

func main() {
	defData, defRoms := "gamedata", os.Getenv("GOPLAY_MT32") // goplay (the browser): the dropped game; the page MT-32's ROMs
	if d := os.Getenv("GOPLAY_DATA"); d != "" {
		defData = d
	}
	data := flag.String("data", defData, "the game directory (with MM3.CC and MM3.CFG)")
	saves := flag.String("saves", "", "where MM3.CUR and the saved games go (default: the user config directory)")
	scale := flag.Int("scale", 3, "window scale")
	pace := flag.Int("pace", mm3.AnimRetraces, "70 Hz retraces per idle view animation frame (5: 386/486, 3: fast 486; higher is slower)")
	roms := flag.String("roms", defRoms, "the MT-32 ROM directory for music (default: roms/ in the working directory or next to the binary; none: silent)")
	importFrom := flag.String("import", "", "goplay: find the game under this folder, copy it to -data, and exit")
	flag.Parse()
	if *importFrom != "" {
		if err := importGame(*importFrom, *data); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	mm3.RomDir = romDir(*roms)
	mm3.AnimRetraces = *pace
	mm3.GameDir = *data
	mm3.SaveDir = *saves
	if mm3.SaveDir == "" && os.Getenv("GOPLAY_DATA") != "" {
		mm3.SaveDir = *data // goplay: saves beside the game
	}
	if mm3.SaveDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			dir = "."
		}
		mm3.SaveDir = filepath.Join(dir, "mm3-go")
	}
	if err := os.MkdirAll(mm3.SaveDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := host.Run("Might and Magic III", *scale, mm3.Main); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// romDir is -roms, else the first roms/ directory found in the working directory or next
// to the binary ("" if none).
func romDir(flagDir string) string {
	if flagDir != "" {
		return flagDir
	}
	dirs := []string{"roms"}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "roms"))
	}
	for _, d := range dirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			return d
		}
	}
	return ""
}
