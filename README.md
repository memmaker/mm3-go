# mm3-go

Might and Magic III: Isles of Terra (1991), ported to Go and [Ebiten](https://ebitengine.org) from the byte-exact
decompilation [mm3-decomp](https://github.com/memmaker/mm3-decomp). It runs on the original game's data files
(`MM3.CC`); no game data is in this repository.

Work in progress.

```sh
go run ./cmd/mm3 -data /path/to/MM3
```

See PORTING.md for how the C maps to Go.
