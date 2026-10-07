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
func Draw(drawList *DrawList, input Input, element UIElement) (ClickArea, bool) {

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
		hovered = MouseInBounds(input, clickArea)
	}

	source := file
	if hovered && hoverFile != "" {
		source = hoverFile
	}

	rect := Rect{X: centerX - width/2, Y: centerY - height/2, W: width, H: height}

	// An image is uploaded once at its native size and stretched over the quad
	// by the sampler, so there is nothing to resample per widget
	var texture *Texture
	if source != "" {
		tex, err := imageTexture(source)
	 	if err == nil {
			texture = tex
		}
		// On a missing or undecodable file, fall through to the background
		// colour rather than emitting a command with no bitmap
	}

	tint := white
	if texture == nil {
		tint = toNRGBA(style.Color)
	}
	drawList.Add(rect, tint, texture)

	// A hovered button with no dedicated hover image is darkened instead
	if hovered && hoverFile == "" {
		drawList.Add(rect, hoverTint, nil)
	}

	return clickArea, clickable
}

// ApplyLayout resolves the widget's own rect: the margin is taken off the
// parent's box, then the widget is sized and aligned in what is left
func ApplyLayout(props Properties) Properties {
	if props.Parent == nil {
		props.Size.Scale = ScalePixel
		return props
	}
	area := inset(*props.Parent, props.Margin)
	p := applyRelative(props, area)
	return applyAlignment(p, area)
}

// ContentBox is the rect a widget's children are laid out in: the widget's
// resolved rect (from ApplyLayout) with its padding taken off.
func ContentBox(props Properties) Properties {
	return inset(props, props.Padding)
}

// applyRelative and applyAlignment take the box to lay out in explicitly: it is
// the parent's box minus the widget's margin, not the parent itself
func applyRelative(p Properties, parent Properties) Properties {
	newWidth := p.Size.Width
	newHeight := p.Size.Height
	if p.Size.Scale == ScaleRelative {
		newWidth = parent.Size.Width * p.Size.Width / 100
		newHeight = parent.Size.Height * p.Size.Height / 100
	}
	p.Size = Size{ScalePixel, newWidth, newHeight}
	return p
}

func applyAlignment(p Properties, parent Properties) Properties {
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

// inset shrinks p's rect by s on each side: a widget's own rect by its
// padding, or its parent's box by its margin
func inset(p Properties, spacing Spacing) Properties {
	top, right, bottom, left := pixels(spacing, p.Size.Width, p.Size.Height)
	p.Size = Size{ScalePixel, p.Size.Width - left - right, p.Size.Height - top - bottom}
	p.Center = Point{p.Center.X + (left-right)/2, p.Center.Y + (top-bottom)/2}
	return p
}

// pixels resolves s to pixels. A relative side is a percentage of the box it
// is taken from: width for left and right, height for top and bottom.
func pixels(spacing Spacing, width, height int) (top, right, bottom, left int) {
	if spacing.Scale == ScaleRelative {
		return height * spacing.Top / 100, width * spacing.Right / 100,
			height * spacing.Bottom / 100, width * spacing.Left / 100
	}
	return spacing.Top, spacing.Right, spacing.Bottom, spacing.Left
}
