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

## Shared decisions (from batch 1)
- Draw lists: a C `int *o` into g_draw_list is an int index into gDrawList; `dlPut(o, v)` = `*++o = v`,
  `dlPutL(o, far)` = `*(long *)++o = p`. `fmtP(fmt)` rewrites Borland `%p` to `%04X` (pass the ds address).
- Mistyped views: gRosterFar, gPowIconFar (far pointers); gMouseCursor/gInnIcons/gTilesetGfx/gPartyFaces/
  gDseFaceGfx are int32 holding Fars; gDrawListB byte view; btnX/btnY/btnW/btnH read the near-pointer button tables;
  monFrame/objFrame for near byte pointers.
- Signatures: freeAndNull(*Far), loadCCFile(string) Far, getWorstCondition(*Chr) uint8,
  getCharStat(*Chr, kind, flag int) int, getCharLevel(*Chr) int, testBit([]byte,int) int, setBit([]byte,bit,set int),
  playMusic([]byte, uint8), readKey() int, getInput() int, equipItem(*Chr, slot int). Music buffers: sndStartMusic(Far).
- Whoever rewrites MM3.CUR sets `ccCur = nil` afterwards.
- Buffers passed to vga as Far must be package-level (farOf never frees). Index `ds[adr+k]` directly where the C
  indexes negatively.

## Shared decisions (scene/events/misc1/stattext)
- `b2i(bool) int` (scene.go) is a C comparison's value (`x ? 0 : 1` -> `b2i(x == 0)`).
- Cell accessors (scene.go): getCellWalls/getCellWallsRegion return the masked u16 as a C int
  (int16-extended); getCellFlags, setCellWall(x, y, k, v), isCellVisited, findLoadedMap, setPartyMapStart.
- Event script (events.go): `ms(i)`/`mu(i)` = g_maze_script[i] signed/(u8); `mapText(n)` = g_map_text +
  g_maze_text_offs[n]; `objD0c()` = &g_map_objects[g_obj_d0c - 1] via ds (index -1 when 0, as the C reads).
  `runEventScript() int`. A C label jumped to from outside its block (`clear:`, `pay:`) is a closure.
- misc1.go: event values are uint32 (`eventTakeGive(m1, v1 uint32, m2, v2 uint32, idx)`, eventSetValue,
  eventCompareValue(what, val uint32, mode, idx) int, eventGiveMulti); `spellKind(cls)` is the
  g_class_spells row; `confirmYesNo(quiet) int`, `selectPartyMember() int`, `askAnswer(mode, ans, msg []byte)`.
- Literal C strings passed to vga as Far: a package-level `var xText = cs("...")` + farOf.
- `checkWallListSize(o int)` takes the draw-list index; `gMonD1l1L` is g_mon_d1l1 as long[].
- Stub signatures: damageParty(dmg, typ, who int), inputNumber(max, width) int32, inputText([]byte, max,
  width) int, faceYesNoDialog(face, a, b []byte, mode, shown) int, visitInn() int, playEnding(uint32) int,
  calcArmorClass(*Chr, base) int, calcMaxSp(*Chr) int, getThievery(*Chr) int, buildingWaitInput() int.

## Shared decisions (cpanel/charinfo/mazeload/town)
- C locals passed by near pointer (draw lists `\x05%04X`, sprite records) live on an emulated C stack in ds:
  `defer dsFree(dsSP)` then `adr := dsAlloca(n)` (ds.go; region above 0xf800, zeroed).
- `waitKey() int` (it returns bioskey(0)). Exp helpers getNextLevelExp/getExpToNextLevel/getTotalExp return int32.
- Signatures: pushIcons(name string), askYesNo(msg []byte, flags) int, showMessage([]byte), loadTextTable([]byte),
  payCost(kind int, amt uint32, who int) int, controlPanel() int, generateRandomItem(lvl, *Chr, slot),
  exchangePartyMember(*int) *Chr, freeMazeResources(all int), gameMainLoop() int.
- Helpers: `rosterAt(i) *Chr` (roster slot via gRosterFar), `mazeB(m, off) *uint8` (g_maps[m][off]);
  Far views of the long-typed globals: gPartyFacesF, gDseFaceGfxF, gTilesetGfxF, gSkyGfxF, gInnIconsF (cpanel/charinfo).
- game_main_loop's cross-case gotos (turn/fwd/moved/turned) are an `act` state run after the key switch.
