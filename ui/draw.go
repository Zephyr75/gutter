package ui

import (
	"image/color"
)

// hoverTint is the default translucent black hover used when no dedicated hover image is provided
var hoverTint = color.NRGBA{0, 0, 0, 55}

// Draw appends the commands for one widget. It writes no pixels: the host turns
// each command into a quad.
//
// The second result reports whether the widget produced a clickable area at all.
// Only a Button does, so callers append on ok rather than collecting a placeholder
// for every widget in the tree -- the list the host hit-tests is then exactly the
// list of things that can be hit.
func Draw(dl *DrawList, in Input, element UIElement) (ClickArea, bool) {

	props := element.GetProperties()

	width := props.Size.Width
	height := props.Size.Height
	centerX := props.Center.X
	centerY := props.Center.Y

	// A widget squeezed to nothing has no quad and no clickable area
	if width <= 0 || height <= 0 {
		return ClickArea{}, false
	}

	// Pull the style and the background images out of the concrete widget
	style := Style{}
	file := ""
	hoverFile := ""

	switch props.Type {
	case UIButton:
		b := element.(Button)
		style, file, hoverFile = b.Style, b.Image, b.HoverImage
	case UIRow:
		r := element.(Row)
		style, file = r.Style, r.Image
	case UIColumn:
		c := element.(Column)
		style, file = c.Style, c.Image
	case UIContainer:
		c := element.(Container)
		style, file = c.Style, c.Image
	}

	// Only a button reports a clickable clickArea, and only a button reacts to hover
	clickArea := ClickArea{}
	clickable := false
	hovered := false
	if props.Type == UIButton {
		clickArea = ClickArea{
			Left:     float64(centerX - width/2),
			Right:    float64(centerX + width/2),
			Top:      float64(centerY - height/2),
			Bottom:   float64(centerY + height/2),
			Function: element.(Button).Function,
		}
		clickable = true
		hovered = MouseInBounds(in, clickArea)
	}

	source := file
	if hovered && hoverFile != "" {
		source = hoverFile
	}

	rect := Rect{X: centerX - width/2, Y: centerY - height/2, W: width, H: height}

	// An image is uploaded once at its native size and stretched over the quad
	// by the sampler, so there is nothing to resample per widget
	var tex *Texture
	if source != "" {
		if t, err := imageTexture(source); err == nil {
			tex = t
		}
		// On a missing or undecodable file, fall through to the background
		// colour rather than emitting a command with no bitmap
	}

	tint := white
	if tex == nil {
		tint = toNRGBA(style.Color)
	}
	dl.Add(rect, tint, tex)

	// A hovered button with no dedicated hover image is darkened instead
	if hovered && hoverFile == "" {
		dl.Add(rect, hoverTint, nil)
	}

	return clickArea, clickable
}

// ApplyLayout resolves a widget's geometry in the order the framework has
// always used: relative sizes against the parent, then alignment within the
// parent, then padding inside the result.
//
// It works on Properties rather than on UIElement. The three UIElement-shaped
// functions below each re-boxed the widget into an interface and copied the
// struct, so every widget paid three copies and three interface allocations per
// frame to compute this; going through Properties costs none.
func ApplyLayout(props Properties) Properties {
	p := applyRelative(props)
	p = applyAlignment(p)
	return applyPadding(p)
}

func applyRelative(p Properties) Properties {
	newWidth := p.Size.Width
	newHeight := p.Size.Height
	if p.Size.Scale == ScaleRelative && p.Parent != nil {
		newWidth = p.Parent.Size.Width * p.Size.Width / 100
		newHeight = p.Parent.Size.Height * p.Size.Height / 100
	}
	p.Size = Size{ScalePixel, newWidth, newHeight}
	return p
}

func applyAlignment(p Properties) Properties {
	parent := p.Parent
	if parent == nil {
		return p
	}

	newX := p.Center.X
	newY := p.Center.Y

	switch p.Alignment {
	case AlignmentCenter:
		newX = parent.Center.X
		newY = parent.Center.Y
	case AlignmentBottom:
		newY = parent.Center.Y + parent.Size.Height/2 - p.Size.Height/2
	case AlignmentTop:
		newY = parent.Center.Y - parent.Size.Height/2 + p.Size.Height/2
	case AlignmentLeft:
		newX = parent.Center.X - parent.Size.Width/2 + p.Size.Width/2
	case AlignmentRight:
		newX = parent.Center.X + parent.Size.Width/2 - p.Size.Width/2
	case AlignmentTopLeft:
		newX = parent.Center.X - parent.Size.Width/2 + p.Size.Width/2
		newY = parent.Center.Y - parent.Size.Height/2 + p.Size.Height/2
	case AlignmentTopRight:
		newX = parent.Center.X + parent.Size.Width/2 - p.Size.Width/2
		newY = parent.Center.Y - parent.Size.Height/2 + p.Size.Height/2
	case AlignmentBottomLeft:
		newX = parent.Center.X - parent.Size.Width/2 + p.Size.Width/2
		newY = parent.Center.Y + parent.Size.Height/2 - p.Size.Height/2
	case AlignmentBottomRight:
		newX = parent.Center.X + parent.Size.Width/2 - p.Size.Width/2
		newY = parent.Center.Y + parent.Size.Height/2 - p.Size.Height/2
	}

	// A row positions its children horizontally and a column vertically, so the
	// axis the parent owns is left as the parent set it
	switch p.Skip {
	case SkipAlignmentHoriz:
		newX = p.Center.X
	case SkipAlignmentVert:
		newY = p.Center.Y
	}

	p.Center = Point{newX, newY}
	return p
}

func applyPadding(p Properties) Properties {
	oldWidth := p.Size.Width
	oldHeight := p.Size.Height

	horizPadding := p.Padding.Left + p.Padding.Right
	vertPadding := p.Padding.Top + p.Padding.Bottom
	horizOffset := p.Padding.Left - p.Padding.Right
	vertOffset := p.Padding.Top - p.Padding.Bottom
	if p.Padding.Scale == ScaleRelative {
		horizPadding = oldWidth * horizPadding / 100
		vertPadding = oldHeight * vertPadding / 100
		horizOffset = oldWidth * horizOffset / 100
		vertOffset = oldHeight * vertOffset / 100
	}

	p.Size = Size{ScalePixel, oldWidth - horizPadding, oldHeight - vertPadding}
	p.Center = Point{p.Center.X + horizOffset/2, p.Center.Y + vertOffset/2}
	return p
}

// Deprecated: use ApplyLayout. Kept so existing callers keep compiling.
func ApplyPadding(element UIElement) UIElement {
	p := applyPadding(element.GetProperties())
	return element.SetProperties(p.Size, p.Center)
}

// Deprecated: use ApplyLayout. Kept so existing callers keep compiling.
func ApplyAlignment(element UIElement) UIElement {
	p := applyAlignment(element.GetProperties())
	return element.SetProperties(p.Size, p.Center)
}

// Deprecated: use ApplyLayout. Kept so existing callers keep compiling.
func ApplyRelative(element UIElement) UIElement {
	p := applyRelative(element.GetProperties())
	return element.SetProperties(p.Size, p.Center)
}
