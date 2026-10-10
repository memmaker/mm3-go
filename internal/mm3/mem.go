package mm3

import "fmt"

// Far is a DOS far pointer: segment << 16 | offset. The game stores far pointers in its
// draw lists and long variables, so they have to be numbers. A segment is a handle into
// segs: a block from memAlloc or a loaded file, or a Go byte slice registered with farOf.
// The offset is 16 bits and wraps like a real far pointer's; 0 is NULL.
type Far uint32

var (
	segs     = [][]byte{nil, ds[:]} // segment 0: NULL, 1: the data segment
	freeSegs []uint32
	sliceSeg = map[*byte]uint32{}
)

// newSeg registers b as a fresh segment and returns its far pointer.
func newSeg(b []byte) Far {
	var s uint32
	if n := len(freeSegs); n > 0 {
		s, freeSegs = freeSegs[n-1], freeSegs[:n-1]
		segs[s] = b
	} else {
		s = uint32(len(segs))
		segs = append(segs, b)
	}
	return Far(s << 16)
}

// memAlloc is sys.asm mem_alloc: n bytes, zeroed (DOS gave whatever was there; the game
// initialises what it reads).
func memAlloc(n int) Far { return newSeg(make([]byte, n)) }

// farOf gives a Go byte slice a far pointer (the same one each time), for the places the
// game passes a near array where a far pointer is expected.
func farOf(b []byte) Far {
	if len(b) == 0 {
		return 0
	}
	if s, ok := sliceSeg[&b[0]]; ok {
		return Far(s << 16)
	}
	p := newSeg(b)
	sliceSeg[&b[0]] = uint32(p >> 16)
	return p
}

// freeFar releases a block from memAlloc or a loaded file (free_graphic).
func freeFar(p Far) {
	s := uint32(p >> 16)
	if s == 0 || s >= uint32(len(segs)) || segs[s] == nil {
		return
	}
	if len(segs[s]) > 0 {
		delete(sliceSeg, &segs[s][0])
	}
	segs[s] = nil
	freeSegs = append(freeSegs, s)
}

// B is the memory from p to the end of its block.
func (p Far) B() []byte {
	s, o := uint32(p>>16), int(p&0xffff)
	if s >= uint32(len(segs)) || segs[s] == nil {
		panic(fmt.Sprintf("far pointer %08x: no such block", uint32(p)))
	}
	return segs[s][o:]
}

// Add is far pointer arithmetic: the offset wraps at 64K, the segment stays.
func (p Far) Add(n int) Far { return p&0xffff0000 | Far(uint16(int(p)+n)) }

func (p Far) Seg() uint16 { return uint16(p >> 16) }
func (p Far) Off() uint16 { return uint16(p) }

// Little-endian access, as the 8086 stores words.
func (p Far) U8(i int) byte          { return p.B()[i] }
func (p Far) SetU8(i int, v byte)    { p.B()[i] = v }
func (p Far) U16(i int) uint16       { return le16(p.B(), i) }
func (p Far) I16(i int) int16        { return int16(le16(p.B(), i)) }
func (p Far) SetU16(i int, v int)    { setLe16(p.B(), i, v) }
func (p Far) U32(i int) uint32       { return le32(p.B(), i) }
func (p Far) SetU32(i int, v uint32) { setLe32(p.B(), i, v) }

func le16(b []byte, i int) uint16       { return uint16(b[i]) | uint16(b[i+1])<<8 }
func setLe16(b []byte, i int, v int)    { b[i], b[i+1] = byte(v), byte(v>>8) }
func le32(b []byte, i int) uint32       { return uint32(le16(b, i)) | uint32(le16(b, i+2))<<16 }
func setLe32(b []byte, i int, v uint32) { setLe16(b, i, int(v)); setLe16(b, i+2, int(v>>16)) }

// I16 and U32 are C ints and longs inside byte-exact structs (the character record, the
// map records): byte arrays, so a Go struct of them has the C layout with no padding and can
// be read and written as raw bytes, as the game does.
type I16 [2]byte
type U16 [2]byte
type U32 [4]byte

func (x *I16) Get() int     { return int(int16(le16(x[:], 0))) }
func (x *I16) Set(v int)    { setLe16(x[:], 0, v) }
func (x *U16) Get() int     { return int(le16(x[:], 0)) }
func (x *U16) Set(v int)    { setLe16(x[:], 0, v) }
func (x *U32) Get() uint32  { return le32(x[:], 0) }
func (x *U32) Set(v uint32) { setLe32(x[:], 0, v) }
