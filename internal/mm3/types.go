package mm3

import "unsafe"

// The game's records. Every field is a byte or a byte array (I16, U16, U32 for the C ints
// and longs), so each Go struct has exactly the C layout: no padding, any alignment. They
// live in ds (the data segment) or in far blocks, are read and written there as raw bytes,
// and can be viewed at any address with at[T].

// Chr is a character record (struct chr / pc, 0x12f bytes): the party (gParty) and the
// roster file (gRoster). See results/character_record.md in the decompilation.
type Chr struct {
	Name                                                                              [0x10]byte  // 0x00
	Sex, Race, Align, Cls                                                             uint8       // 0x10
	Might, MightTmp, Intel, IntelTmp, Pers, PersTmp, Endur, EndurTmp, Speed, SpeedTmp uint8       // 0x14: stat k at 0x14+2k, its temporary bonus at 0x15+2k
	Acc, AccTmp, Luck, LuckTmp                                                        uint8       // 0x1e
	AcTmp, Level, LevelTmp, BirthDay, AgeMod                                          uint8       // 0x22
	Skill                                                                             [0x12]uint8 // 0x27
	Award                                                                             [0x1a]uint8 // 0x39
	Spells                                                                            [0x24]uint8 // 0x53
	BeaconMap, BeaconX, BeaconY, Caster, ReadySpell, Quickfight                       uint8       // 0x77
	ItemEq, ItemFlags, ItemElem, ItemMat, ItemAttr, ItemID, ItemPower                 [0x13]uint8 // 0x7d: 7 parallel arrays by item slot
	Bless, Shield, Holy, Hero, Donated                                                uint8       // 0x102
	FireRes, FireResTmp, ColdRes, ColdResTmp, ElecRes, ElecResTmp                     uint8       // 0x107
	PoisonRes, PoisonResTmp, EnergyRes, EnergyResTmp, MagicRes, MagicResTmp           uint8       // 0x10d
	Cond                                                                              [0x10]uint8 // 0x113: cursed .. eradicated
	Inn                                                                               U16         // 0x123
	HP, SP                                                                            I16         // 0x125
	BirthYear                                                                         U16         // 0x129
	Exp                                                                               U32         // 0x12b
}

// Maze is a map record (struct maze / map, 0x340 bytes): g_maps[4], the 16x16 quadrants
// around the party.
type Maze struct {
	Cell                                                        [16][16]U16   // 0x000: wall bits
	CellFlags                                                   [16][16]uint8 // 0x200
	Surface                                                     [8]uint8      // 0x300
	North, East, South, West                                    uint8         // 0x308: the adjacent maps
	Save, Lit, Rest, F30f, TavernTip, F311, LockLevel, StartPos uint8         // 0x30c
	F314                                                        [12]uint8     // 0x314
	Visited                                                     [0x20]uint8   // 0x320: bit per cell
}

// Spr is a draw list entry (struct spr): flag, image, position, attr (scale, bit 15 mirrors), frame.
type Spr struct {
	Flag              I16
	Img               U32 // a Far
	X, Y, Attr, Frame I16
}

// Spr1 is a one-entry draw list with its terminator (struct spr1).
type Spr1 struct {
	Spr
	End, EndImgOff, EndImgSeg I16
}

// Dspr is a two-entry draw list sharing one image (struct dspr).
type Dspr struct {
	Flag                                                         I16
	Img                                                          U32
	BotX, BotY, BotAttr, BotFrame, TopX, TopY, TopAttr, TopFrame I16
	End, EndImgOff, EndImgSeg                                    I16
}

// Evt is an event record of the maze script index (struct evt).
type Evt struct {
	X, Y, Dir, Line I16
	IP              U16
}

// Anim is a 12-byte map object animation record (struct anim).
type Anim struct {
	Frame U16
	Type  I16
	Pad   [8]byte
}

// Mobj is a 12-byte map object (struct mobj); Frame is a near pointer.
type Mobj struct {
	Y, X, Sprite I16
	Frame        U16
	ID, Attr     I16
}

func init() {
	for _, s := range []struct{ got, want uintptr }{
		{unsafe.Sizeof(Chr{}), 0x12f}, {unsafe.Sizeof(Maze{}), 0x340}, {unsafe.Sizeof(Spr{}), 14},
		{unsafe.Sizeof(Spr1{}), 20}, {unsafe.Sizeof(Dspr{}), 28}, {unsafe.Sizeof(Evt{}), 10},
		{unsafe.Sizeof(Anim{}), 12}, {unsafe.Sizeof(Mobj{}), 12},
	} {
		if s.got != s.want {
			panic("record layout")
		}
	}
}
