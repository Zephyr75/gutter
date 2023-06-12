package ui

import (
	// "fmt"
	"image"
	// "image/color"

	"github.com/go-gl/glfw/v3.3/glfw"
)


type Button struct {
	Properties Properties
	Style	   Style
	Child      UIElement
}

func (button Button) Initialize() UIElement {
  button.Properties = DefaultProperties(button.Properties)
  return button
}

func (button Button) Draw(img *image.RGBA, window *glfw.Window) {
  // fmt.Println("--------------------")
  // fmt.Println(button)

  if !button.Properties.Initialized {
    button = button.Initialize().(Button)
  }

  button = ApplyRelative(button).(Button)

  button = ApplyAlignment(button).(Button)

  button = ApplyPadding(button).(Button)





  // fmt.Println(button)

	Draw(img, window, button.Properties, button.Style)
	
	// if button.Child != nil {
	// 	button.Child.SetProperties(button.Properties.Size, button.Properties.Center)
	// 	button.Child.Draw(img, window)
	// }
}


func (button Button) SetProperties(size Size, center Point) UIElement {
	button.Properties.Size = size
	button.Properties.Center = center
	//println("Button: ", center.X, " ", center.Y, " ", size.Width, " ", size.Height)
  return button
}

func (button Button) SetParent(parent *Properties) UIElement {
  button.Properties.Parent = parent
  return button
}

func (button Button) GetProperties() Properties {
  return button.Properties
}

func (button Button) Debug() {
	println(button.Properties.Center.Y)
}
