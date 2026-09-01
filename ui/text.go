package ui

import (
	"image"
	"os"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
)

type Text struct {
	Properties Properties
	StyleText  StyleText
	// Content is the string to draw. Embedded newlines start a new line.
	Content string
}

func (text Text) Initialize(skip SkipAlignment) UIElement {
	text.Properties = DefaultProperties(text.Properties, skip, UIText)
	text.StyleText = DefaultStyleText(text.StyleText)
	return text
}

func (text Text) Draw(img *image.RGBA, window *glfw.Window) []Area {

	if !text.Properties.Initialized {
		text = text.Initialize(SkipAlignmentNone).(Text)
	}

	text.Properties = ApplyLayout(text.Properties)

	if text.Content == "" {
		return []Area{}
	}

	props := text.Properties
	rect := image.Rect(
		props.Center.X-props.Size.Width/2,
		props.Center.Y-props.Size.Height/2,
		props.Center.X+props.Size.Width/2,
		props.Center.Y+props.Size.Height/2,
	)

	drawText(img, strings.Split(text.Content, "\n"), text.StyleText, rect)

	return []Area{}

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

// textFill is reused for the glyph colour; the context reads it synchronously
// during DrawString and the UI is single-threaded.
var textFill image.Uniform

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

// drawText lays the lines out inside rect: left-aligned, and the block as a
// whole centred vertically. The clip is the widget's own box, so a string that
// is too long is cut off rather than painted over its neighbours.
func drawText(img *image.RGBA, lines []string, style StyleText, rect image.Rectangle) {
	c, err := textContext(style.Font, style.FontSize)
	if err != nil {
		return
	}

	textFill.C = style.FontColor
	c.SetDst(img)
	c.SetSrc(&textFill)
	c.SetClip(rect)

	size := float64(style.FontSize)
	lineHeight := int(c.PointToFixed(size*1.5) >> 6)
	ascent := int(c.PointToFixed(size) >> 6)

	top := rect.Min.Y + (rect.Dy()-lineHeight*len(lines))/2
	pt := freetype.Pt(rect.Min.X, top+ascent)
	for _, s := range lines {
		if _, err := c.DrawString(s, pt); err != nil {
			return
		}
		pt.Y += c.PointToFixed(size * 1.5)
	}
}
