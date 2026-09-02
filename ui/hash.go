package ui

import "image/color"

// Change detection compares the whole widget tree against the previous frame.
// Doing that by building a string per node allocates in proportion to the tree
// on every frame, and the string encoding lost field boundaries: Size{1, 23}
// and Size{12, 3} both serialised to "123", so genuine changes went unnoticed
// and distinct widgets shared a hover-state key.
//
// Hashing with a fixed-width, self-delimiting encoding fixes both: no
// allocation, and no two distinct field sequences collide by construction.

const (
	fnvOffset64 uint64 = 14695981039346656037
	fnvPrime64  uint64 = 1099511628211
)

// Hasher is an FNV-1a accumulator. Use NewHasher; the zero value is not seeded.
type Hasher struct{ sum uint64 }

func NewHasher() Hasher { return Hasher{sum: fnvOffset64} }

// Sum returns the hash of everything mixed in so far.
func (h *Hasher) Sum() uint64 { return h.sum }

func (h *Hasher) mix(b byte) {
	h.sum ^= uint64(b)
	h.sum *= fnvPrime64
}

// Uint64 mixes v in as eight bytes. Every other method funnels through this, so
// every field occupies a fixed width and field boundaries are unambiguous.
func (h *Hasher) Uint64(v uint64) {
	for i := 0; i < 8; i++ {
		h.mix(byte(v >> (8 * i)))
	}
}

func (h *Hasher) Int(v int) { h.Uint64(uint64(v)) }

func (h *Hasher) Bool(v bool) {
	if v {
		h.Uint64(1)
		return
	}
	h.Uint64(0)
}

// String mixes the bytes, then the length, so "ab"+"c" and "a"+"bc" differ.
func (h *Hasher) String(s string) {
	for i := 0; i < len(s); i++ {
		h.mix(s[i])
	}
	h.Uint64(uint64(len(s)))
}

// Color hashes the resolved RGBA rather than the dynamic type, so two different
// implementations of the same colour compare equal. A nil colour gets its own
// sentinel, since it means "unset" and is not the same as transparent black.
func (h *Hasher) Color(c color.Color) {
	if c == nil {
		h.Uint64(^uint64(0))
		return
	}
	r, g, b, a := c.RGBA()
	h.Uint64(uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a))
}

// TreeHash is the per-frame redraw key: it walks the tree once, allocating
// nothing, and returns a value that changes whenever any widget's geometry,
// style or content changed.
func TreeHash(element UIElement) uint64 {
	h := NewHasher()
	if element != nil {
		element.Hash(&h)
	}
	return h.Sum()
}

func (p Padding) Hash(h *Hasher) {
	h.Bool(bool(p.Scale))
	h.Int(p.Top)
	h.Int(p.Right)
	h.Int(p.Bottom)
	h.Int(p.Left)
}

func (s Size) Hash(h *Hasher) {
	h.Bool(bool(s.Scale))
	h.Int(s.Width)
	h.Int(s.Height)
}

func (p Point) Hash(h *Hasher) {
	h.Int(p.X)
	h.Int(p.Y)
}

func (p Properties) Hash(h *Hasher) {
	p.Center.Hash(h)
	p.Size.Hash(h)
	h.Uint64(uint64(p.Alignment))
	p.Padding.Hash(h)
	h.Uint64(uint64(p.Type))
	h.Uint64(uint64(p.Skip))
}

func (s Style) Hash(h *Hasher) {
	h.Color(s.Color)
}

func (s StyleText) Hash(h *Hasher) {
	h.String(s.Font)
	h.Int(s.FontSize)
	h.Color(s.FontColor)
}

func (a ClickArea) Hash(h *Hasher) {
	h.Uint64(uint64(int64(a.Top)))
	h.Uint64(uint64(int64(a.Right)))
	h.Uint64(uint64(int64(a.Bottom)))
	h.Uint64(uint64(int64(a.Left)))
}
