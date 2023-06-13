package ui

import (
	// "fmt"
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"

)

type Column struct {
	Properties Properties
	Style	   Style
	Children   []UIElement
}

func (column Column) Initialize(skip SkipAlignment) UIElement {
  column.Properties = DefaultProperties(column.Properties, skip)
  return column
}

func (column Column) Draw(img *image.RGBA, window *glfw.Window) {
  // fmt.Println("--------------------")

  if !column.Properties.Initialized {
    column = column.Initialize(SkipAlignmentNone).(Column)
  }

  column = ApplyRelative(column).(Column)

  column = ApplyAlignment(column).(Column)

  column = ApplyPadding(column).(Column)

  for i, child := range column.Children {
    child = child.SetParent(&column.Properties)
    column.Children[i] = child.Initialize(SkipAlignmentVert)
  }

  // fmt.Println("Column")
  // fmt.Println(column.Properties)
	
	Draw(img, window, column.Properties, column.Style)

  availableHeight := column.Properties.Size.Height
  maxHeight := column.Properties.Size.Height
  if column.Properties.Size.Scale == ScaleRelative {
    availableHeight = column.Properties.Size.Height * maxHeight / 100
  }

  // Compute the available width
	for _, child := range column.Children {
    childProps := child.GetProperties()
    if childProps.Size.Scale == ScalePixel { 
      availableHeight -= childProps.Size.Height
    }
	}

  // Compute the total percentage of width required by the children
  childrenHeight := 0
  for _, child := range column.Children {
    childProps := child.GetProperties()
    if childProps.Size.Scale == ScaleRelative { 
      childrenHeight += childProps.Size.Height
    }
  }

  // Compute the width of each child
  for i, child := range column.Children {
    childProps := child.GetProperties()
    if childProps.Size.Scale == ScaleRelative { 
      column.Children[i] = child.SetProperties(
        Size{
          Scale:  ScalePixel,
          Width: column.Properties.Size.Width,
          Height:  childProps.Size.Height * availableHeight / childrenHeight,
        },
        Point{
          X: childProps.Center.X,
          Y: childProps.Center.Y,
        },
      )
    }
  }


  // Compute the center of each child
  currentY := column.Properties.Center.Y - maxHeight / 2
  for i, child := range column.Children {
    childProps := child.GetProperties()
    pixelHeight := childProps.Size.Height
    if childProps.Size.Scale == ScaleRelative {
      pixelHeight = childProps.Size.Height * availableHeight / childrenHeight
    }
    column.Children[i] = child.SetProperties(
      Size{
        Scale:  childProps.Size.Scale,
        Width: childProps.Size.Width,
        Height:  childProps.Size.Height,
      },
      Point{
        X: column.Properties.Center.X,
        Y: currentY + pixelHeight / 2,
      },
    )
    currentY += pixelHeight
  }


  for _, child := range column.Children {
    // fmt.Println("child")
    // fmt.Println(child)
    child.Draw(img, window)
  }


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

func (column Column) Debug() {
	println(column.Properties.Center.Y)
}

