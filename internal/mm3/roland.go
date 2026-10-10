package mm3

// Port of roland.drv (MM3.CC, loaded by snddrv.asm at seg:0): the MT-32 / LAPC-1 driver
// behind an MPU-401. Jump table: 01a2 init, 01eb shutdown, 021f music, 029e timer/sfx,
// 02fd sample. Its variables and tables live in the driver image, so the port works on a
// 64K copy of it at the original offsets (m), reading the sfx scripts, the note table and
// the program maps from there; the song is the caller's buffer. Every "d0" MPU command
// (want to send data) followed by a MIDI message is one out(msg).
//
// Driver channels 0..8 are MIDI channels 2..10 (status + 1). A script byte's high nibble
// is the command, the low nibble the channel:
//
//	music (table 0162)                 sfx (table 0182, scripts in the driver, list 087c)
//	0 call w (rel. to 012f), max 16    -
//	1 delay b ticks                    1 delay b ticks
//	2 map[ch] = 14th of 14 bytes       2 sfxmap[ch] = b
//	3 volume b (cc 7), not while fading
//	4 pitch bend slide off
//	5 bend slide: delta w, start w (lsb, msb as 7-bit bytes), added every idle tick
//	6 pan b (cc 10)
//	8 note off                         8 note off (b ignored)
//	9 note on b: ch 8 key b, else table[b&1f] + 12*(b>>5); velocity 7f
//	a, b skip b (AdLib only)           b sysex: f0 ... f7
//	c program map[b]                   c program sfxmap[b]
//	ff return / loop to 0              ff loop to the sfx start
//	f0..fe end                         f0..fe end
//	7, d, e: nothing

const (
	rMusicOn   = 0x15
	rSfxOn     = 0x16
	rSfxMode   = 0x17
	rFadeDone  = 0x18
	rFadeRate  = 0x19
	rFadeAcc   = 0x1a
	rFadeLevel = 0x1b
	rSfxPos    = 0x1c
	rSfxStart  = 0x1e
	rSfxDelay  = 0x20
	rSfxMap    = 0x43
	rMusMap    = 0x54
	rCallStack = 0x65 // return offsets; 00a5 the targets
	rDepth     = 0xe5
	rNote      = 0xed // sounding key per channel, ff none
	rBendDelta = 0xf6
	rBend      = 0x108
	rDelay     = 0x11a
	rVelocity  = 0x11b
	rVolume    = 0x124
	rPos       = 0x12d
	rBase      = 0x12f
	rNoteTab   = 0x14a
	rSfxList   = 0x87c
)

type roland struct {
	m    []byte // the driver image, 64K
	song []byte // es of the music
	out  func(msg []byte)
}

func newRoland(drv []byte, out func([]byte)) *roland {
	m := make([]byte, 0x10000)
	copy(m, drv)
	return &roland{m: m, out: out}
}

func (r *roland) w(o int) int   { return int(r.m[o&0xffff]) | int(r.m[(o+1)&0xffff])<<8 }
func (r *roland) setW(o, v int) { r.m[o] = byte(v); r.m[o+1] = byte(v >> 8) }
func (r *roland) msg(b ...byte) { r.out(b) }
func (r *roland) allCh(lo, hi int, f func(ch int)) { // bx = hi down to lo, as the driver loops
	for ch := hi; ch >= lo; ch-- {
		f(ch)
	}
}

// es:si, music or sfx
func (r *roland) rd(si int) byte {
	si &= 0xffff
	if r.m[rSfxMode] == 1 {
		return r.m[si]
	}
	if si < len(r.song) {
		return r.song[si]
	}
	return 0
}
func (r *roland) rdW(si int) int { return int(r.rd(si)) | int(r.rd(si+1))<<8 }

func (r *roland) bendCenter() { r.allCh(0, 8, func(c int) { r.msg(0xe1+byte(c), 0x40, 0x40) }) } // 042b
func (r *roland) notesOff17() { r.allCh(0, 6, func(c int) { r.msg(0xb1+byte(c), 0x7b, 0) }) }    // 0453
func (r *roland) notesOff89() { r.msg(0xb9, 0x7b, 0); r.msg(0xb8, 0x7b, 0) }                     // 0472
func (r *roland) resetCtl()   { r.allCh(0, 8, func(c int) { r.msg(0xb1+byte(c), 0x79, 0) }) }    // 049b
func (r *roland) volumes()    { r.allCh(0, 8, func(c int) { r.msg(0xb1+byte(c), 0x07, 0x5f) }) } // 04b8
func (r *roland) pans()       { r.allCh(0, 8, func(c int) { r.msg(0xb1+byte(c), 0x0a, 0x3f) }) } // 04d5

// noteOff is 0533.
func (r *roland) noteOff(ch int) {
	if r.m[rNote+ch]&0x80 == 0 {
		n := r.m[rNote+ch] & 0x7f
		r.m[rNote+ch] = 0xff
		r.msg(0x81+byte(ch), n, r.m[rVelocity+ch])
	}
}

// key is 055e: note-table entry plus 12 per octave (bits 5..7).
func (r *roland) key(b byte) byte {
	return r.m[rNoteTab+int(b&0x1f)] + (b>>5)*8 + (b>>5)*4
}

// init is 01a2 (the MPU reset and the timer hook are the host's).
func (r *roland) init() { r.m[rMusicOn], r.m[rSfxOn] = 0, 0 }

// shutdown is 01eb.
func (r *roland) shutdown() { r.notesOff17(); r.notesOff89() }

// music is 021f with the argument's high word hi (start a song: startSong).
func (r *roland) music(hi int) int {
	switch {
	case hi == 0:
		r.m[rMusicOn] = 0
		r.notesOff17()
	case hi > 0 && hi <= 0xff:
		r.m[rFadeRate], r.m[rFadeLevel] = byte(hi), 0x7f
	case hi == -1:
		return int(r.m[rFadeRate])
	}
	return 0
}

// startSong is 025c.
func (r *roland) startSong(song []byte) {
	r.song = song
	r.setW(rDepth, 0)
	r.setW(rPos, 0)
	r.m[rDelay] = 0
	for c := 0; c < 6; c++ {
		r.setW(rBendDelta+2*c, 0)
	}
	r.bendCenter()
	r.volumes()
	r.pans()
	r.m[rMusicOn] = 1
	r.notesOff17()
}

// sfx is 02c2: start effect v (0..96).
func (r *roland) sfx(v int) {
	if v > 0x96 {
		return
	}
	for c := 6; c < 9; c++ {
		r.setW(rBendDelta+2*c, 0)
	}
	r.notesOff89()
	r.resetCtl()
	r.bendCenter()
	p := r.w(rSfxList + 2*v)
	r.setW(rSfxPos, p)
	r.setW(rSfxStart, p)
	r.m[rSfxOn], r.m[rSfxDelay] = 1, 0
}

// sampleStart is 032e's effect on the music: it stops (the sample's interrupt replaces
// the music's until it ends).
func (r *roland) sampleStart() {
	r.m[rMusicOn] = 0
	r.notesOff17()
	r.notesOff89()
}

// tick is the timer interrupt 0577 (the tick count is the caller's): one music step, or
// else one sfx step, or else the fade and the bend slides.
func (r *roland) tick() {
	m := r.m
	si := 0
	if m[rMusicOn] == 1 {
		m[rSfxMode] = 0
		si = r.w(rPos)
		if m[rDelay] == 0 {
			goto run
		}
		if m[rDelay]--; m[rDelay] == 0 {
			goto run
		}
	}
	if m[rSfxOn] == 0 {
		goto idle
	}
	si = r.w(rSfxPos)
	m[rSfxMode] = 1
	if m[rSfxDelay] != 0 {
		if m[rSfxDelay]--; m[rSfxDelay] != 0 {
			goto idle
		}
	}
run:
	for {
		al := r.rd(si)
		si++
		ch := int(al & 0xf)
		sfx := m[rSfxMode] == 1
		switch al >> 4 {
		case 0:
			if !sfx { // 070c
				if bx := r.w(rDepth); bx < 0x20 {
					r.setW(rCallStack+bx, si-r.w(rBase)+2)
					si = (r.rdW(si) + r.w(rBase)) & 0xffff
					r.setW(0xa5+bx, si)
					r.setW(rDepth, bx+2)
				}
			}
		case 1: // 073a, 0748
			if sfx {
				m[rSfxDelay] = r.rd(si)
				r.setW(rSfxPos, si+1)
			} else {
				m[rDelay] = r.rd(si)
				r.setW(rPos, si+1)
			}
			return
		case 2:
			if sfx { // 0756
				m[rSfxMap+ch] = r.rd(si)
				si++
			} else { // 0766
				m[rMusMap+ch] = r.rd(si + 13)
				si += 14
			}
		case 3: // 0645
			if m[rFadeRate] == 0 {
				m[rVolume+ch] = r.rd(si)
				r.msg(0xb1+byte(ch), 7, m[rVolume+ch])
			}
			si++
		case 4: // 06a9
			r.setW(rBendDelta+2*ch, 0)
		case 5: // 068d
			r.setW(rBendDelta+2*ch, r.rdW(si))
			r.setW(rBend+2*ch, r.rdW(si+2))
			si += 4
		case 6: // 06b9
			r.msg(0xb1+byte(ch), 0x0a, r.rd(si))
			si++
		case 8: // 05ec, 05f6
			if sfx {
				si++
			}
			r.noteOff(ch)
		case 9: // 0607
			k := r.rd(si)
			si++
			if ch != 8 {
				k = r.key(k)
			}
			r.noteOff(ch)
			m[rNote+ch] = k & 0x7f
			r.msg(0x91+byte(ch), k&0x7f, m[rVelocity+ch])
		case 0xa:
			if !sfx {
				si++
			}
		case 0xb:
			if sfx { // 0674
				msg := []byte{0xf0}
				for {
					b := r.rd(si)
					si++
					msg = append(msg, b)
					if b == 0xf7 {
						break
					}
				}
				r.out(msg)
			} else {
				si++
			}
		case 0xc: // 06d9
			p := int(r.rd(si))
			si++
			mp := rMusMap
			if sfx {
				mp = rSfxMap
			}
			r.msg(0xc1+byte(ch), m[mp+p])
		case 0xf:
			if sfx { // 07a3
				if al != 0xff {
					m[rSfxOn] = 0
					return
				}
				si = r.w(rSfxStart)
			} else { // 0779
				if al != 0xff {
					m[rMusicOn] = 0
					return
				}
				if bx := r.w(rDepth); bx != 0 {
					si = (r.w(0x63+bx) + r.w(rBase)) & 0xffff
					r.setW(rDepth, bx-2)
				} else {
					si = 0
				}
			}
		}
		si &= 0xffff
	}
idle:
	if m[rFadeRate] != 0 { // 07b6
		s := int(m[rFadeAcc]) + int(m[rFadeRate])
		m[rFadeAcc] = byte(s)
		if s > 0xff {
			stop := false
			if m[rFadeLevel]--; int8(m[rFadeLevel]) < 0 {
				stop = true
			} else {
				m[rFadeDone] = 1
				for c := 8; c >= 0; c-- {
					if int8(m[rVolume+c]) >= 0x28 {
						m[rVolume+c]--
						m[rFadeDone] = 0
						r.msg(0xb1+byte(c), 7, m[rVolume+c])
					}
				}
				stop = m[rFadeDone] == 1
			}
			if stop { // 07cf
				m[rMusicOn], m[rFadeRate] = 0, 0
				r.notesOff17()
			}
		}
	}
	for c := 8; c >= 0; c-- { // 0818
		if d := r.w(rBendDelta + 2*c); d != 0 {
			b := (r.w(rBend+2*c) + d) & 0xffff
			r.setW(rBend+2*c, b)
			r.msg(0xe1+byte(c), byte(b)&0x7f, byte(b>>8)&0x7f)
		}
	}
}
