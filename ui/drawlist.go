package ui

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/Zephyr75/gutter/utils"
)

// A widget tree paints by appending to a DrawList rather than by writing pixels,
// the host turns each Cmd into one textured, alpha-blended quad
type DrawList struct {
	Cmds []Cmd
}

// Rect is in pixels, with the origin top-left and Y increasing downwards, 
// which is the coordinate system the layout already works in
type Rect struct {
	X, Y, W, H int
}

// Texture is a bitmap the host uploads and then caches by Key : 
// Key is stable for the same Pixels, so the host uploads once and can look up afterwards
//
// Pixels is *image.NRGBA (Non-premultiplied RGBA) which is what a SrcAlpha/OneMinusSrcAlpha blend expects. Stride is always W*4, so
// the host can hand Pixels.Pix to an RGBA8 upload directly.
type Texture struct {
	Key    uint64
	Pixels *image.NRGBA
	W, H   int
}

// Cmd is one quad : Color multiplies Tex; with Tex nil it is the fill colour.
type Cmd struct {
	Rect  Rect
	Color color.NRGBA
	Tex   *Texture
}

// Reset keeps the backing array, so a list reused across frames stops allocating
// once it has reached its high-water mark.
func (d *DrawList) Reset() { d.Cmds = d.Cmds[:0] }

func (d *DrawList) Add(rect Rect, col color.NRGBA, tex *Texture) {
	d.Cmds = append(d.Cmds, Cmd{Rect: rect, Color: col, Tex: tex})
}

var white = color.NRGBA{255, 255, 255, 255}

// toNRGBA converts to straight alpha without going through color.Model.Convert,
// which returns a color.Color and so boxes its result on every call -- once per
// widget per frame.
func toNRGBA(c color.Color) color.NRGBA {
	switch v := c.(type) {
	case nil:
		return color.NRGBA{}
	case color.NRGBA:
		return v
	case color.RGBA:
		// color.RGBA is alpha-premultiplied; a blend expects straight
		switch v.A {
		case 0xff:
			return color.NRGBA{v.R, v.G, v.B, 0xff}
		case 0:
			return color.NRGBA{}
		}
		a := uint32(v.A)
		return color.NRGBA{
			uint8(uint32(v.R) * 0xff / a),
			uint8(uint32(v.G) * 0xff / a),
			uint8(uint32(v.B) * 0xff / a),
			v.A,
		}
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{
		uint8((r * 0xffff / a) >> 8),
		uint8((g * 0xffff / a) >> 8),
		uint8((b * 0xffff / a) >> 8),
		uint8(a >> 8),
	}
}

// Images are uploaded at their native size and scaled by the sampler, so unlike
// the old rasteriser there is nothing to resample per widget and nothing to
// re-cache when a widget changes size.

type imageEntry struct {
	tex *Texture
	err error
}

var imageTextures = map[string]imageEntry{}

func imageTexture(path string) (*Texture, error) {
	if e, ok := imageTextures[path]; ok {
		return e.tex, e.err
	}

	src, err := utils.GetImageFromFilePath(path)
	if err != nil {
		imageTextures[path] = imageEntry{nil, err}
		return nil, err
	}

	b := src.Bounds()
	pix := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(pix, pix.Bounds(), src, b.Min, draw.Src)

	h := NewHasher()
	h.String("image")
	h.String(path)

	tex := &Texture{Key: h.Sum(), Pixels: pix, W: b.Dx(), H: b.Dy()}
	imageTextures[path] = imageEntry{tex, nil}
	return tex, nil
}

// ClearTextureCache drops every cached image and text bitmap. Only useful in
// tests and when reloading assets that changed on disk.
func ClearTextureCache() {
	imageTextures = map[string]imageEntry{}
	textTextures = map[textKey]*Texture{}
}
