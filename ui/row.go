package ui

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Row struct {
	Properties Properties
	Style      Style
	Children   []UIElement
	Image      string
}

func (row Row) Initialize(skip SkipAlignment) UIElement {
	row.Properties = DefaultProperties(row.Properties, skip, UIRow)
	row.Style = DefaultStyle(row.Style)
	return row
}

func (row Row) Draw(img *image.RGBA, window *glfw.Window) []Area {

	areas := []Area{}

	if !row.Properties.Initialized {
		row = row.Initialize(SkipAlignmentNone).(Row)
	}

	row.Properties = ApplyLayout(row.Properties)

	for i, child := range row.Children {
		child = child.SetParent(&row.Properties)
		row.Children[i] = child.Initialize(SkipAlignmentHoriz)
	}

	areas = append(areas, Draw(img, window, row))

	// Fixed-size children claim their width first
	availableWidth := row.Properties.Size.Width
	maxWidth := row.Properties.Size.Width
	for _, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScalePixel {
			availableWidth -= childProps.Size.Width
		}
	}

	// What is left is shared among the relative children, in proportion to the
	// percentages they declared
	childrenWidth := 0
	for _, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScaleRelative {
			childrenWidth += childProps.Size.Width
		}
	}

	// Size and place every child in one pass
	currentX := row.Properties.Center.X - maxWidth/2
	for i, child := range row.Children {
		childProps := child.GetProperties()

		pixelWidth := childProps.Size.Width
		pixelHeight := childProps.Size.Height
		if childProps.Size.Scale == ScaleRelative {
			// childrenWidth is zero when every relative child declared a width of
			// zero, which used to divide by zero and take the process with it
			pixelWidth = 0
			if childrenWidth > 0 {
				pixelWidth = childProps.Size.Width * availableWidth / childrenWidth
			}
			// A relative child fills the row's height
			pixelHeight = row.Properties.Size.Height
		}

		row.Children[i] = child.SetProperties(
			Size{
				Scale:  ScalePixel,
				Width:  pixelWidth,
				Height: pixelHeight,
			},
			Point{
				X: currentX + pixelWidth/2,
				Y: row.Properties.Center.Y,
			},
		)
		currentX += pixelWidth
	}

	for _, child := range row.Children {
		areas = append(areas, child.Draw(img, window)...)
	}

	return areas

}

func (row Row) SetProperties(size Size, center Point) UIElement {
	row.Properties.Size = size
	row.Properties.Center = center
	return row
}

func (row Row) SetParent(parent *Properties) UIElement {
	row.Properties.Parent = parent
	return row
}

func (row Row) GetProperties() Properties {
	return row.Properties
}

func (row Row) Hash(h *Hasher) {
	row.Properties.Hash(h)
	row.Style.Hash(h)
	h.String(row.Image)
	h.Int(len(row.Children))
	for _, child := range row.Children {
		child.Hash(h)
	}
}

func (row Row) ToString() string {
	result := row.Properties.ToString() + row.Style.ToString()
	for _, child := range row.Children {
		result += child.ToString()
	}
	return result
}
