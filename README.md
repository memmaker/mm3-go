# mm3-go

Might and Magic III: Isles of Terra (1991), ported to Go and [Ebiten](https://ebitengine.org) from the byte-exact
decompilation [mm3-decomp](https://github.com/memmaker/mm3-decomp). It runs on the original game's data files
(`MM3.CC`); no game data is in this repository.

Work in progress.

```sh
go run ./cmd/mm3 -data /path/to/MM3
```

Music and sound effects play on munt's MT-32 emulation (roland.drv's player): install munt's libmt32emu
(build it from munt or use a package) and put the MT-32 ROMs (`mt32_ctrl_*.rom`, `mt32_pcm.rom`) in
`roms/` (working directory or next to the binary) or pass `-roms DIR`. Without them the game runs silently.

See PORTING.md for how the C maps to Go.
