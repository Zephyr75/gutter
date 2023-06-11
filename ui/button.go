
package ui

import (
	"image"
	"image/color"

	"github.com/go-gl/glfw/v3.3/glfw"
)


type Button struct {
	Properties *Properties
	Style	   Style
	Child      UIElement
}

func NewButton(props *Properties) Button {
	return Button{
		Properties: props,
		Style: Style{
			Color: color.RGBA{0, 0, 0, 0},
		},
	}
}

func (button Button) Draw(img *image.RGBA, window *glfw.Window) {
	Draw(img, window, button.Properties, button.Style)
	
	if button.Child != nil {
		button.Child.SetProperties(button.Properties.Size, button.Properties.Center)
		button.Child.Draw(img, window)
	}
}


func (button Button) SetProperties(size Size, center Point) {
	button.Properties.MaxSize = size
	button.Properties.Center = center
	//println("Button: ", center.X, " ", center.Y, " ", size.Width, " ", size.Height)
}

func (button Button) Debug() {
	println(button.Properties.Center.Y)
}
