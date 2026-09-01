package ui

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Column struct {
	Properties Properties
	Style      Style
	Children   []UIElement
	Image      string
}

func (column Column) Initialize(skip SkipAlignment) UIElement {
	column.Properties = DefaultProperties(column.Properties, skip, UIColumn)
	column.Style = DefaultStyle(column.Style)
	return column
}

func (column Column) Draw(img *image.RGBA, window *glfw.Window) []Area {

	areas := []Area{}

	if !column.Properties.Initialized {
		column = column.Initialize(SkipAlignmentNone).(Column)
	}

	column.Properties = ApplyLayout(column.Properties)

	for i, child := range column.Children {
		child = child.SetParent(&column.Properties)
		column.Children[i] = child.Initialize(SkipAlignmentVert)
	}

	areas = append(areas, Draw(img, window, column))

	// Fixed-size children claim their height first
	availableHeight := column.Properties.Size.Height
	maxHeight := column.Properties.Size.Height
	for _, child := range column.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScalePixel {
			availableHeight -= childProps.Size.Height
		}
	}

	// What is left is shared among the relative children, in proportion to the
	// percentages they declared
	childrenHeight := 0
	for _, child := range column.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScaleRelative {
			childrenHeight += childProps.Size.Height
		}
	}

	// Size and place every child in one pass
	currentY := column.Properties.Center.Y - maxHeight/2
	for i, child := range column.Children {
		childProps := child.GetProperties()

		pixelWidth := childProps.Size.Width
		pixelHeight := childProps.Size.Height
		if childProps.Size.Scale == ScaleRelative {
			// childrenHeight is zero when every relative child declared a height
			// of zero, which used to divide by zero and take the process with it
			pixelHeight = 0
			if childrenHeight > 0 {
				pixelHeight = childProps.Size.Height * availableHeight / childrenHeight
			}
			// A relative child fills the column's width
			pixelWidth = column.Properties.Size.Width
		}

		column.Children[i] = child.SetProperties(
			Size{
				Scale:  ScalePixel,
				Width:  pixelWidth,
				Height: pixelHeight,
			},
			Point{
				X: column.Properties.Center.X,
				Y: currentY + pixelHeight/2,
			},
		)
		currentY += pixelHeight
	}

	for _, child := range column.Children {
		areas = append(areas, child.Draw(img, window)...)
	}

	return areas
}

func (column Column) SetProperties(size Size, center Point) UIElement {
	column.Properties.Size = size
	column.Properties.Center = center
	return column
}

func (column Column) SetParent(parent *Properties) UIElement {
	column.Properties.Parent = parent
	return column
}

func (column Column) GetProperties() Properties {
	return column.Properties
}

func (column Column) Hash(h *Hasher) {
	column.Properties.Hash(h)
	column.Style.Hash(h)
	h.String(column.Image)
	h.Int(len(column.Children))
	for _, child := range column.Children {
		child.Hash(h)
	}
}

func (column Column) ToString() string {
	result := column.Properties.ToString() + column.Style.ToString()
	for _, child := range column.Children {
		result += child.ToString()
	}
	return result
}
