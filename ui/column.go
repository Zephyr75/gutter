
package ui

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
)


type Column struct {
	Properties *Properties
	Style	   Style
	Children   []UIElement
}

func (column Column) Draw(img *image.RGBA, window *glfw.Window) {
	
	Draw(img, window, column.Properties, column.Style)

	for child := range column.Children {
		
		column.Children[child].SetProperties(
			Size{
				Scale:  column.Properties.Size.Scale,
				Width:  column.Properties.Size.Width,
				Height: column.Properties.Size.Height / len(column.Children),
			},
			Point{
				X: column.Properties.Center.X,
				Y: column.Properties.Center.Y - column.Properties.MaxSize.Height/2 + (2*child+1)*column.Properties.MaxSize.Height/(len(column.Children)*2),
			},
		)
		column.Children[child].Draw(img, window)
	}
}


func (column Column) SetProperties(size Size, center Point) {
	column.Properties.MaxSize = size
	column.Properties.Center = center
}

func (column Column) Debug() {
	println(column.Properties.Center.Y)
}
