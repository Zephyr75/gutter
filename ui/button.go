
package ui

import (
	"image"
	// "image/color"

	"github.com/go-gl/glfw/v3.3/glfw"
)


type Button struct {
	Properties Properties
	Style	   Style
	Child      UIElement
}


func (button Button) Draw(img *image.RGBA, window *glfw.Window) {

  button.Properties = DefaultProperties() 

	Draw(img, window, button.Properties, button.Style)
	
	// if button.Child != nil {
	// 	button.Child.SetProperties(button.Properties.Size, button.Properties.Center)
	// 	button.Child.Draw(img, window)
	// }
}


func (button Button) SetProperties(size Size, center Point) {
	button.Properties.Size = size
	button.Properties.Center = center
	//println("Button: ", center.X, " ", center.Y, " ", size.Width, " ", size.Height)
}

func (button Button) GetProperties() Properties {
  return button.Properties
}

func (button Button) Debug() {
	println(button.Properties.Center.Y)
}
