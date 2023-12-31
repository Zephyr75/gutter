package ui

import (
	// "fmt"
	"image"
	// "image/color"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Button struct {
	Properties Properties
	Style      Style
	Child      UIElement
	Function   func()
	Image      string
	HoverImage string
}

func (button Button) Initialize(skip SkipAlignment) UIElement {
	button.Properties = DefaultProperties(button.Properties, skip, UIButton)
	button.Style = DefaultStyle(button.Style)
	return button
}

func (button Button) Draw(img *image.RGBA, window *glfw.Window) []Area {

	areas := []Area{}

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

	areas = append(areas, Draw(img, window, button))

	if button.Child != nil {
		props := button.Child.GetProperties()
		button.Child.SetProperties(props.Size, button.Properties.Center)
		areas = append(areas, button.Child.Draw(img, window)...)
	}

	return areas
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

func (button Button) ToString() string {
	result := button.Properties.ToString() +
		button.Style.ToString() +
		button.Image +
		button.HoverImage
	if button.Child != nil {
		result += button.Child.ToString()
	}
	return result
}

// TODO: split between draw and setup
// TODO: create a method to check if mouse is over button
// TODO: add a parameter to get the result
// TODO: call setup before toString and draw after
