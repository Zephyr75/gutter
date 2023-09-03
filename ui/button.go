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
  Image       string
}

func (button Button) Initialize(skip SkipAlignment) UIElement {
  button.Properties = DefaultProperties(button.Properties, skip, UIButton)
  return button
}

func (button Button) Draw(img *image.RGBA, window *glfw.Window) {

  // get color
  // _, _, b, _ := button.Style.Color.RGBA()

  // if b > 200 {
  //   fmt.Println("--------------------")
  //   fmt.Println(button.Properties.Parent)
  //   fmt.Println(button)
  // }

  if !button.Properties.Initialized {
    button = button.Initialize(SkipAlignmentNone).(Button)
  }

  button = ApplyRelative(button).(Button)

  button = ApplyAlignment(button).(Button)

  button = ApplyPadding(button).(Button)

  if button.Child != nil {
    button.Child = button.Child.SetParent(&button.Properties)
    button.Child = button.Child.Initialize(SkipAlignmentNone)
  }

  // if b > 200 {
  //   fmt.Println(button)
  // }

	Draw(img, window, button.Properties, button.Style, button.Image)
	
	if button.Child != nil {
    props := button.Child.GetProperties()
		button.Child.SetProperties(props.Size, button.Properties.Center)
		button.Child.Draw(img, window)
	}
}


func (button Button) SetProperties(size Size, center Point) UIElement {
	button.Properties.Size = size
	button.Properties.Center = center
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
