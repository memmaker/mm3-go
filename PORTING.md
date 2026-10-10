# Porting rules (C decompilation -> internal/mm3)

Source: ~/Projects/decomp-might-magic/src (byte-exact BC++ 2.0 decompilation). One Go file per C file
(view.c -> view.go), functions in C order, C names in camelCase (`get_char_stat` -> `getCharStat`).
Port faithfully: same logic, same branches, same arithmetic. No redesign.

## Memory
- Globals are typed views into `ds` (the data segment) at their original addresses: `ds_gen.go` (generated,
  do not edit). `g_party_count` -> `gPartyCount` (*uint8: use `*gPartyCount`), arrays -> slices that run to the
  end of ds (C's run-on indexing works). `adrXxx` constants are the addresses.
- When a module uses a global with a different type than the generated view, make a local view:
  `at[int16](adrBankBtn0Y)`, `arr[Spr](adrBankSprites)`. Labels that point inside another array
  (g_party_skills = g_party[0].skill) keep working because it is all one memory.
- Records: `types.go` (Chr, Maze, Spr, Spr1, Dspr, Evt, Anim, Mobj) are byte-exact: ints are I16/U16/U32
  with .Get()/.Set(). Map a module's local struct variant to the canonical one by **byte offset**
  (results/character_record.md in the decompilation has the Chr offsets). Raw byte access:
  `bytesOf(c)[0x14+2*k]`.
- `char far *` -> `Far` (mem.go): `p.B()` bytes, `p.Add(n)`, `memAlloc`, `freeFar`. A far pointer to ds is
  `dsFar(adr)`, to any Go pointer into ds/blocks `farPtr(&x)`; a record in far memory `farAt[Chr](p)`.
  Near pointers stored in memory are uint16 ds offsets: `str(p)` gives the C string, `nearOf(&x)` the offset.
- Locals stay Go locals (C `u8 buf[0x396]` -> `var buf [0x396]byte`).

## Types and arithmetic
- C int is 16 bits, char is SIGNED, long 32. Keep truncation/sign where it matters: compute in Go int, store
  with the global's type (`*gMinutes = int16(...)`). Watch: unsigned 16-bit compares/division, `(u8)` casts,
  char comparisons (`int8(x) > 7`), shifts of 16-bit values, int overflow wrap.
- C strings are NUL-terminated []byte: `cs("literal")`, `gostr(b)`, `strcpy`, `sprintf(dst, fmt, args...)`
  (libc.go, Borland semantics: %d/%u are 16-bit unless %ld/%lu).
- `goto` is fine in Go if it doesn't jump over declarations; restructure minimally otherwise.

## Platform
- Drivers: vga.go (video, being written: loadGraphic, drawSprites(list []byte), printWindowText([]byte), ...),
  snd.go (sndTimer, sndPlaySample, sndPlayMusic/sndStartMusic), cc.go (ccLoad, curWriteFile, open/close
  archive), sys.go. bioskey(n) and file I/O shims go in dos.go (create functions as needed, documented).
- A busy-wait loop that waits for input/time must reach host.Busy()/host.Yield() (bioskey and sndTimer do).
- MK_FP(0xa000,0) is `vram[:]` (vga.go).
- Missing callee from a module not yet ported: add a stub to stubs.go with the final signature
  (`// stub: <c file>`), panicking `panic("todo: name")`. Whoever ports the real one deletes the stub.

## Done means
`go vet ./...` clean for your files, package compiles (with stubs), and a top comment
`// Port of X.C: ...` as in ~/Projects/UltimaUnderworld/internal/uw.
