
package ui

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Row struct {
	Properties Properties
	Style	   Style
	Children   []UIElement
}

func (row Row) Draw(img *image.RGBA, window *glfw.Window) {
  row.Properties = DefaultProperties() 
	
	Draw(img, window, row.Properties, row.Style)

  availableWidth := row.Properties.Size.Width

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

  if childrenWidth == 0 {
   childrenWidth = 1
  }

  // Compute the width of each child
  for _, child := range row.Children {
    childProps := child.GetProperties()
    if childProps.Size.Scale == ScaleRelative { 
      child.SetProperties(
        Size{
          Scale:  childProps.Size.Scale,
          Width:  childProps.Size.Width * availableWidth / childrenWidth,
          Height: childProps.Size.Height,
        },
        Point{
          X: childProps.Center.X,
          Y: childProps.Center.Y,
        },
      )
    }
  }

  // Compute the center of each child
  currentX := row.Properties.Center.X - row.Properties.Size.Width / 2
  for _, child := range row.Children {
    childProps := child.GetProperties()
    pixelWidth := childProps.Size.Width
    if childProps.Size.Scale == ScaleRelative {
      pixelWidth = childProps.Size.Width * availableWidth / childrenWidth
    }
    child.SetProperties(
      Size{
        Scale:  childProps.Size.Scale,
        Width:  childProps.Size.Width,
        Height: childProps.Size.Height,
      },
      Point{
        X: currentX + pixelWidth / 2,
        Y: childProps.Center.Y,
      },
    )
    child.Draw(img, window)
    currentX += pixelWidth
  }




}


func (row Row) SetProperties(size Size, center Point) {
	row.Properties.Size = size
	row.Properties.Center = center
}

func (row Row) GetProperties() Properties {
  return row.Properties
}

func (row Row) Debug() {
	println(row.Properties.Center.Y)
}
