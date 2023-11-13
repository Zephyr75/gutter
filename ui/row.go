package ui

import (
	// "fmt"
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
	return row
}

func (row Row) Draw(img *image.RGBA, window *glfw.Window) []Area {
	// fmt.Println("--------------------")

	areas := []Area{}

	if !row.Properties.Initialized {
		row = row.Initialize(SkipAlignmentNone).(Row)
	}

	row = ApplyRelative(row).(Row)

	row = ApplyAlignment(row).(Row)

	row = ApplyPadding(row).(Row)

	for i, child := range row.Children {
		child = child.SetParent(&row.Properties)
		row.Children[i] = child.Initialize(SkipAlignmentHoriz)
	}

	areas = append(areas, Draw(img, window, row))

	availableWidth := row.Properties.Size.Width
	maxWidth := row.Properties.Size.Width
	if row.Properties.Size.Scale == ScaleRelative {
		availableWidth = row.Properties.Size.Width * maxWidth / 100
	}

	// Compute the available width
	for _, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScalePixel {
			availableWidth -= childProps.Size.Width
		}
	}

	// Compute the total percentage of width required by the children
	childrenWidth := 0
	for _, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScaleRelative {
			childrenWidth += childProps.Size.Width
		}
	}

	// Compute the width of each child
	for i, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScaleRelative {
			row.Children[i] = child.SetProperties(
				Size{
					Scale:  ScalePixel,
					Width:  childProps.Size.Width * availableWidth / childrenWidth,
					Height: row.Properties.Size.Height,
				},
				Point{
					X: childProps.Center.X,
					Y: childProps.Center.Y,
				},
			)
		}
	}

	// Compute the center of each child
	currentX := row.Properties.Center.X - maxWidth/2
	for i, child := range row.Children {
		childProps := child.GetProperties()
		pixelWidth := childProps.Size.Width
		if childProps.Size.Scale == ScaleRelative {
			pixelWidth = childProps.Size.Width * availableWidth / childrenWidth
		}
		row.Children[i] = child.SetProperties(
			Size{
				Scale:  childProps.Size.Scale,
				Width:  childProps.Size.Width,
				Height: childProps.Size.Height,
			},
			Point{
				X: currentX + pixelWidth/2,
				Y: row.Properties.Center.Y,
			},
		)
		currentX += pixelWidth
	}

	for _, child := range row.Children {
		// fmt.Println(child)
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

func (row Row) ToString() string {
	result := row.Properties.ToString() + row.Style.ToString()
	for _, child := range row.Children {
		result += child.ToString()
	}
	return result
}
