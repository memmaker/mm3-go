package mm3

import (
	"bytes"
	"unsafe"
)

// ds is the data segment (DGROUP): every global of the C code at its original address. The
// decompiled C treats globals as addresses: one label read as different types by different
// modules, labels inside other arrays, arrays that run on into the next global, saves that
// dump address ranges. With the globals as typed views of one block of memory (ds_gen.go),
// all of that behaves as it did. Near pointers are offsets into ds.
var ds [0x10000]byte

// dsSeg is ds's far segment: a far pointer to ds offset a is dsFar(a).
const dsSeg = 1

// at is a view of ds at address a as a T.
func at[T any](a int) *T { return (*T)(unsafe.Pointer(&ds[a])) }

// arr is a view of ds from address a to its end as a []T (a C array: its size is the
// program's business).
func arr[T any](a int) []T {
	var z T
	return unsafe.Slice((*T)(unsafe.Pointer(&ds[a])), (len(ds)-a)/int(unsafe.Sizeof(z)))
}

// dsFar is the far pointer to ds offset a.
func dsFar(a int) Far { return Far(dsSeg<<16 | a&0xffff) }

// nearOf is the near pointer (ds offset) of p, which must point into ds.
func nearOf[T any](p *T) uint16 {
	o := uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(&ds[0]))
	if o >= uintptr(len(ds)) {
		panic("nearOf: not in the data segment")
	}
	return uint16(o)
}

// farPtr is the far pointer of p: into ds, or into a far block.
func farPtr[T any](p *T) Far {
	a := uintptr(unsafe.Pointer(p))
	if o := a - uintptr(unsafe.Pointer(&ds[0])); o < uintptr(len(ds)) {
		return dsFar(int(o))
	}
	for s, b := range segs[2:] {
		if len(b) > 0 {
			if o := a - uintptr(unsafe.Pointer(&b[0])); o < uintptr(len(b)) {
				return Far(uint32(s+2)<<16 | uint32(o))
			}
		}
	}
	panic("farPtr: not in a known block")
}

// farAt is a view of far memory at p as a T (e.g. a Chr in the roster).
func farAt[T any](p Far) *T { return (*T)(unsafe.Pointer(&p.B()[0])) }

// bytesOf is the memory of *p as bytes (a record's raw bytes).
func bytesOf[T any](p *T) []byte {
	var z T
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), unsafe.Sizeof(z))
}

// str is the C string at near pointer p: ds from p (up to the end; strings end at a NUL).
func str(p uint16) []byte { return ds[p:] }

// cs is a NUL-terminated copy of s, for passing Go string literals where the C passed one.
func cs(s string) []byte { return append([]byte(s), 0) }

// gostr is the C string in b, up to its NUL.
func gostr(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

type (
	i16s []int16
	u16s []uint16
	i32s []int32
	u32s []uint32
)

type dsBlock struct {
	a int
	v any
}

func init() {
	for _, b := range dsInit {
		switch v := b.v.(type) {
		case string:
			copy(ds[b.a:], v)
			ds[b.a+len(v)] = 0
		case []byte:
			copy(ds[b.a:], v)
		case i16s:
			for i, x := range v {
				setLe16(ds[:], b.a+2*i, int(x))
			}
		case u16s:
			for i, x := range v {
				setLe16(ds[:], b.a+2*i, int(x))
			}
		case i32s:
			for i, x := range v {
				setLe32(ds[:], b.a+4*i, uint32(x))
			}
		case u32s:
			for i, x := range v {
				setLe32(ds[:], b.a+4*i, x)
			}
		}
	}
}

// dsSP is the C stack, which lived in DGROUP above the last global (SS = DS). A local the
// game passes by near pointer (a draw list in "\x05%p", a struct spr1 button) must have a ds
// address: dsAlloca reserves n zeroed bytes there; the function does `defer dsFree(dsSP)`
// first (the argument is evaluated then), which pops them all on return.
var dsSP = len(ds)

func dsAlloca(n int) int {
	dsSP -= n
	if dsSP < 0xf800 {
		panic("dsAlloca: stack overflow")
	}
	memset(ds[dsSP:], 0, n)
	return dsSP
}

func dsFree(sp int) { dsSP = sp }
