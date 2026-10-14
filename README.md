# mm3-go

Might and Magic III: Isles of Terra (1991), ported to Go and [Ebiten](https://ebitengine.org) from the byte-exact
decompilation [mm3-decomp](https://github.com/memmaker/mm3-decomp). It runs on the original game's data files
(`MM3.CC`); no game data is in this repository.

Work in progress.

```sh
go run ./cmd/mm3 -data /path/to/MM3
```

Music and sound effects play on munt's MT-32 emulation (roland.drv's player, preferring a CM-32L, whose
extra drum sounds the effects use): install munt's libmt32emu (build it from munt or use a package) and put
the MT-32 ROMs (`mt32_ctrl_*.rom`, `mt32_pcm.rom`, or the CM-32L's) in `roms/` (working directory or next to
the binary) or pass `-roms DIR`. Without them they play on an emulated AdLib (adlib.drv's player on
[github.com/memmaker/opl2](https://github.com/memmaker/opl2)). Speech plays either way.

Both drivers are ported to Go (internal/mm3/roland.go, adlib.go). Their effect scripts and note tables are
generated from the drivers in MM3.CC: `MM3_DATA=/path/to/MM3 go generate ./internal/mm3`
(cmd/rolanddata, cmd/adlibdata). Until opl2's `Resampler.Mix` is tagged, `../opl2` must sit beside this
repository (go.mod's `replace`).

See PORTING.md for how the C maps to Go.

## In the browser

The WebAssembly build (`GOOS=js GOARCH=wasm go build ./cmd/mm3`) runs in
[goplay](https://ruzzoli.de/goplay/) on the player's own files: `GOPLAY_DATA`
names the game's folder (saves go beside it). `-import SRC` finds MM3.CC in
whatever the player dropped (`SRC` is the folder holding the dropped items):
plain folders and loose files, zip files, CD-ROM images (ISO 9660, raw
BIN/CUE in mode 1 or 2) and GOG installers (Inno Setup 5.5 to 6.4
`setup_*.exe`, with any `setup_*-N.bin` slices beside it), nested in any
order. It copies MM3.CC, MM3.CFG (writing a VGA one if missing) and any saved
games from the folder MM3.CC lies in to `-data` under upper-case names, and
exits. The search lives in `internal/gameimport`, written to be copied into
other ports. When the page's MT-32 has its
ROMs, goplay sets `GOPLAY_MT32` and music and sound effects play on it (munt
compiled to WebAssembly, offered by the page as `goplayMT32`); otherwise on the
AdLib. Speech plays either way.
