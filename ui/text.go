package ui

import (
	"image"
	"image/color"
	"os"
	"strings"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
)

type Text struct {
	Properties Properties
	StyleText  StyleText
	// Content is the string to draw. Embedded newlines start a new line.
	Content string
}

func (text Text) Initialize(in Input, skip SkipAlignment) UIElement {
	text.Properties = DefaultProperties(text.Properties, in, skip, UIText)
	text.StyleText = DefaultStyleText(text.StyleText)
	return text
}

func (text Text) Draw(dl *DrawList, in Input) []ClickArea {
	if !text.Properties.Initialized {
		text = text.Initialize(in, SkipAlignmentNone).(Text)
	}

	text.Properties = ApplyLayout(text.Properties)

	if text.Content == "" {
		return nil
	}

	props := text.Properties
	tex := textTexture(text.Content, text.StyleText, props.Size.Width, props.Size.Height)
	if tex == nil {
		return nil
	}

	// The bitmap is the text block itself, so it is placed rather than filled:
	// against the widget's left edge, centred on its vertical axis.
	dl.Add(Rect{
		X: props.Center.X - props.Size.Width/2,
		Y: props.Center.Y - tex.H/2,
		W: tex.W,
		H: tex.H,
	}, toNRGBA(text.StyleText.FontColor), tex)

	return nil
}

func (text Text) SetProperties(size Size, center Point) UIElement {
	text.Properties.Size = size
	text.Properties.Center = center
	return text
}

func (text Text) SetParent(parent *Properties) UIElement {
	text.Properties.Parent = parent
	return text
}

func (text Text) GetProperties() Properties {
	return text.Properties
}

func (text Text) Hash(h *Hasher) {
	text.Properties.Hash(h)
	text.StyleText.Hash(h)
	h.String(text.Content)
}

func (text Text) ToString() string {
	return text.Properties.ToString() +
		text.StyleText.ToString() +
		text.Content
}

// Reading and parsing a font file, and then building a rasterising context for
// it, used to happen once per Text per frame. Both are cached: the parse by
// path, the context by path and size.
//
// The context cache is the one that matters. freetype.Context keeps an internal
// cache of rasterised glyphs which SetFont and SetFontSize discard -- and both
// return early when the value is unchanged -- so a context that survives across
// frames rasterises each glyph once rather than once per frame. SetDst, SetSrc
// and SetClip leave that cache intact, which is why they can still be set per
// call.
var (
	parsedFonts   = map[string]*truetype.Font{}
	parsedFontErr = map[string]error{}
	textContexts  = map[textContextKey]*freetype.Context{}
)

type textContextKey struct {
	font string
	size int
}

func loadFont(path string) (*truetype.Font, error) {
	if f, ok := parsedFonts[path]; ok {
		return f, parsedFontErr[path]
	}

	b, err := os.ReadFile(path)
	if err != nil {
		parsedFonts[path], parsedFontErr[path] = nil, err
		return nil, err
	}
	f, err := freetype.ParseFont(b)
	parsedFonts[path], parsedFontErr[path] = f, err
	return f, err
}

func textContext(fontPath string, fontSize int) (*freetype.Context, error) {
	key := textContextKey{fontPath, fontSize}
	if c, ok := textContexts[key]; ok {
		if c == nil {
			return nil, parsedFontErr[fontPath]
		}
		return c, nil
	}

	f, err := loadFont(fontPath)
	if err != nil {
		textContexts[key] = nil
		return nil, err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(float64(fontSize))
	// Hinting is deliberately left at the default. SetHinting(font.HintingFull)
	// makes the first DrawString on a fresh context emit no pixels at all, which
	// is invisible when a context is thrown away every frame and very visible
	// once contexts are cached.
	textContexts[key] = c
	return c, nil
}

// A rendered string is cached as a white bitmap whose alpha is the glyph
// coverage, and coloured by the Cmd's tint. Colour is therefore not part of the
// key: the same string at the same size is one texture however many colours it
// is drawn in.
type textKey struct {
	content string
	font    string
	size    int
	width   int
	height  int
}

var (
	textTextures = map[textKey]*Texture{}
	measureDst   = image.NewNRGBA(image.Rect(0, 0, 1, 1))
	whiteFill    = image.NewUniform(color.White)
)

// textBucket rounds a bitmap's width up so that a string which changes every
// frame keeps producing the same byte size. Hosts that reallocate a texture
// whenever its size changes -- and that leak a descriptor slot each time they do
// -- would otherwise exhaust their slots within seconds on something as ordinary
// as a frame counter.
const textBucket = 64

func textTexture(content string, style StyleText, maxWidth, maxHeight int) *Texture {
	// Obtain or create a freetype.Context for the requested font and size.
	c, err := textContext(style.Font, style.FontSize)
	if err != nil {
		return nil
	}

	// Split content into separate lines to handle multiline strings.
	lines := strings.Split(content, "\n")

	// Helper values derived from the font size.
	size := float64(style.FontSize)
	lineHeight := int(c.PointToFixed(size*1.5) >> 6) // approximate line spacing
	ascent := int(c.PointToFixed(size) >> 6)         // distance above baseline

	// ------------------------------------------------------------
	// Measure phase: compute the maximum width needed by the string.
	// We render onto a dummy surface with an empty clip so no pixels are drawn,
	// but the pen position still advances as if drawing were performed.
	// ------------------------------------------------------------
	c.SetDst(measureDst)
	c.SetSrc(whiteFill)
	c.SetClip(image.Rectangle{})
	natural := 0
	for _, line := range lines {
		p, err := c.DrawString(line, freetype.Pt(0, 0))
		if err != nil {
			return nil
		}
		// Convert the fixed-point X position to integer width.
		if w := int(p.X >> 6); w > natural {
			natural = w
		}
	}

	// Clamp and bucketize the measured width; this guarantees a stable texture size across frames.
	width := roundUpTo(clampAbove(min(natural, maxWidth), 1), textBucket)
	height := clampAbove(min(lineHeight*len(lines), maxHeight), 1)

	// Check if we already have a cached texture for these parameters.
	key := textKey{content, style.Font, style.FontSize, width, height}
	if tex, ok := textTextures[key]; ok {
		return tex
	}

	// ------------------------------------------------------------
	// Rendering phase: create a real image and draw the string onto it.
	// ------------------------------------------------------------
	pix := image.NewNRGBA(image.Rect(0, 0, width, height))
	c.SetDst(pix)
	c.SetSrc(whiteFill)
	c.SetClip(pix.Bounds()) // clip to avoid drawing outside the texture

	pt := freetype.Pt(0, ascent) // start at baseline
	for _, line := range lines {
		if _, err := c.DrawString(line, pt); err != nil {
			return nil
		}
		// Move down by one line height (scaled appropriately).
		pt.Y += c.PointToFixed(size * 1.5)
	}

	// Create a hash key that uniquely identifies this texture's content and size.
	h := NewHasher()
	h.String("text")
	h.String(content)
	h.String(style.Font)
	h.Int(style.FontSize)
	h.Int(width)
	h.Int(height)

	tex := &Texture{Key: h.Sum(), Pixels: pix, W: width, H: height}
	textTextures[key] = tex
	return tex
}

func roundUpTo(v, multiple int) int {
	return ((v + multiple - 1) / multiple) * multiple
}

func clampAbove(v, low int) int {
	if v < low {
		return low
	}
	return v
}
