package ui

import (
	"fmt"
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"

  "gutter/utils"
)

type Row struct {
	Properties Properties
	Style	   Style
	Children   []UIElement
}

func (row Row) Initialize() UIElement {
  row.Properties = DefaultProperties()

  // row.Properties.Size = Size{
  //   Scale:  ScalePixel,
  //   Width:  utils.RESOLUTION_X,
  //   Height: utils.RESOLUTION_Y,
  // }
    


  for i, child := range row.Children {
    row.Children[i] = child.Initialize()
  }
  return row
}

func (row Row) Draw(img *image.RGBA, window *glfw.Window) {


  row = row.Initialize().(Row)

	
	Draw(img, window, row.Properties, row.Style)

  availableWidth := row.Properties.Size.Width
  fmt.Println("availableWidth: ", availableWidth)

  for _, child := range row.Children {

    screenSize := Size{ScalePixel, utils.RESOLUTION_X, utils.RESOLUTION_Y}
    screenCenter := Point{utils.RESOLUTION_X / 2, utils.RESOLUTION_Y / 2}


    child.SetProperties(screenSize, screenCenter)
        

  }


  // Compute the available width
	for _, child := range row.Children {
    childProps := child.GetProperties()
    fmt.Println(childProps.Size.Height)
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

  fmt.Println("availableWidth: ", availableWidth)

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

  for _, child := range row.Children {
    fmt.Println(child)
  }

  // Compute the center of each child
  // currentX := row.Properties.Center.X - row.Properties.Size.Width / 2
  currentX := 0
  if row.Properties.Size.Scale == ScaleRelative {
    currentX = row.Properties.Center.X - availableWidth / 2
  }
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
    fmt.Println("Drawing child at ", currentX + pixelWidth / 2)
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
