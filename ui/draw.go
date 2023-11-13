package ui

import (
	// "fmt"
	"image"
	"image/color"
	// "sync"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/nfnt/resize"
	"image/draw"

	"github.com/Zephyr75/gutter/utils"
)

func MouseInBounds(window *glfw.Window, area Area) bool {
	x, y := window.GetCursorPos()
	if x > area.Left && x < area.Right && y > area.Top && y < area.Bottom {
		return true
	}
	return false
}

func Draw(img *image.RGBA, window *glfw.Window, element UIElement) Area {

	props := element.GetProperties()
	style := Style{}
	file := ""
	hoverFile := ""

	width := props.Size.Width
	height := props.Size.Height
	centerX := props.Center.X
	centerY := props.Center.Y

	// Check if the mouse is in bounds
	darken := false
	area := Area{}
	if props.Type == UIButton {
		area = Area{
			Left: float64(centerX - width/2), 
			Right: float64(centerX + width/2), 
			Top: float64(centerY - height/2), 
			Bottom: float64(centerY + height/2),
			Function: element.(Button).Function,
		}
		if MouseInBounds(window, area) {
			darken = true
			// Call the function on click
			// if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
			// 	element.(Button).Function()
			// }
		}
	}

	// Get the style
	switch {
	case props.Type == UIButton:
		style = element.(Button).Style
		file = element.(Button).Image
		hoverFile = element.(Button).HoverImage
	case props.Type == UIRow:
		style = element.(Row).Style
	case props.Type == UIColumn:
		style = element.(Column).Style
	case props.Type == UIContainer:
		style = element.(Container).Style
	}

	var col color.Color
	col = color.RGBA{0, 0, 0, 0}
	if style.Color != nil {
		col = style.Color
	}

	var texture, hoverTexture, blackTexture image.Image
	texture = image.NewUniform(col)
	hoverTexture = image.NewUniform(col)
	blackTexture = image.NewUniform(color.RGBA{0, 0, 0, 55})
	if file != "" {
		texture, _ = utils.GetImageFromFilePath(file)
		texture = resize.Resize(uint(width), uint(height), texture, resize.Lanczos3)
	}
	if hoverFile != "" {
		hoverTexture, _ = utils.GetImageFromFilePath(hoverFile)
		hoverTexture = resize.Resize(uint(width), uint(height), hoverTexture, resize.Lanczos3)
	}

	rect := image.Rect(centerX-width/2, centerY-height/2, centerX+width/2, centerY+height/2)

	if !darken {
		draw.Draw(img, rect, texture, image.Point{}, draw.Over)
	} else {
		if hoverFile != "" {
			draw.Draw(img, rect, hoverTexture, image.Point{}, draw.Over)
		} else {
			draw.Draw(img, rect, texture, image.Point{}, draw.Over)
			draw.Draw(img, rect, blackTexture, image.Point{}, draw.Over)
		}
	}

	return area

}

func ApplyPadding(element UIElement) UIElement {
	props := element.GetProperties()

	oldWidth := props.Size.Width
	oldHeight := props.Size.Height

	horizPadding := props.Padding.Left + props.Padding.Right
	vertPadding := props.Padding.Top + props.Padding.Bottom
	horizOffset := props.Padding.Left - props.Padding.Right
	vertOffset := props.Padding.Top - props.Padding.Bottom
	if props.Padding.Scale == ScaleRelative {
		horizPadding = oldWidth * horizPadding / 100
		vertPadding = oldHeight * vertPadding / 100
		horizOffset = oldWidth * horizOffset / 100
		vertOffset = oldHeight * vertOffset / 100
	}
	newSize := Size{ScalePixel, oldWidth - horizPadding, oldHeight - vertPadding}
	newCenter := Point{props.Center.X + horizOffset/2, props.Center.Y + vertOffset/2}
	// newCenter := Point{props.Center.X, props.Center.Y}
	return element.SetProperties(newSize, newCenter)
}

func ApplyAlignment(element UIElement) UIElement {
	props := element.GetProperties()
	parent := props.Parent
	newX := props.Center.X
	newY := props.Center.Y

	switch props.Alignment {
	case AlignmentCenter:
		newX = parent.Center.X
		newY = parent.Center.Y
	case AlignmentBottom:
		newY = parent.Center.Y + parent.Size.Height/2 - props.Size.Height/2
	case AlignmentTop:
		newY = parent.Center.Y - parent.Size.Height/2 + props.Size.Height/2
	case AlignmentLeft:
		newX = parent.Center.X - parent.Size.Width/2 + props.Size.Width/2
	case AlignmentRight:
		newX = parent.Center.X + parent.Size.Width/2 - props.Size.Width/2
	case AlignmentTopLeft:
		newX = parent.Center.X - parent.Size.Width/2 + props.Size.Width/2
		newY = parent.Center.Y - parent.Size.Height/2 + props.Size.Height/2
	case AlignmentTopRight:
		newX = parent.Center.X + parent.Size.Width/2 - props.Size.Width/2
		newY = parent.Center.Y - parent.Size.Height/2 + props.Size.Height/2
	case AlignmentBottomLeft:
		newX = parent.Center.X - parent.Size.Width/2 + props.Size.Width/2
		newY = parent.Center.Y + parent.Size.Height/2 - props.Size.Height/2
	case AlignmentBottomRight:
		newX = parent.Center.X + parent.Size.Width/2 - props.Size.Width/2
		newY = parent.Center.Y + parent.Size.Height/2 - props.Size.Height/2
	}

	switch props.Skip {
	case SkipAlignmentHoriz:
		newX = props.Center.X
	case SkipAlignmentVert:
		newY = props.Center.Y
	}

	newCenter := Point{newX, newY}

	return element.SetProperties(props.Size, newCenter)
}

func ApplyRelative(element UIElement) UIElement {
	props := element.GetProperties()
	parent := props.Parent
	newWidth := props.Size.Width
	newHeight := props.Size.Height
	if props.Size.Scale == ScaleRelative {
		newWidth = parent.Size.Width * props.Size.Width / 100
		newHeight = parent.Size.Height * props.Size.Height / 100
	}
	newSize := Size{ScalePixel, newWidth, newHeight}
	return element.SetProperties(newSize, props.Center)
}

/*
Button
Text
Row
Column

Align
--------
Center
Left
Right
Top
Bottom
Top left
Top right
Bottom left
Bottom right



Padding
--------
Pixel : All around, Symmetric, Side by side
Relative : All around, Symmetric, Side by side



Style
--------
Background color
Border color
Border width
Border radius
Shadow
Text color
Text size
Text font



Parent





Color

Border radius
*/
