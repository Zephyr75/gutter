package ui

import (
	"image"
	"image/color"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"

  "gutter/utils"
)

type ScaleType byte

const (
	ScalePixel    ScaleType = 0
	ScaleRelative ScaleType = 1
)

/*
Padding
*/
type Padding struct {
	Scale  ScaleType
	Top    int
	Right  int
	Bottom int
	Left   int
}

func PaddingEqual(scale ScaleType, padding int) Padding {
	return Padding{
		Scale:  scale,
		Top:    padding,
		Right:  padding,
		Bottom: padding,
		Left:   padding,
	}
}
func PaddingSymmetric(scale ScaleType, vertical, horizontal int) Padding {
	return Padding{
		Scale:  scale,
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}
func PaddingSideBySide(scale ScaleType, top, right, bottom, left int) Padding {
	return Padding{
		Scale:  scale,
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

/*
Alignment
*/
type Alignment byte

const (
	AlignmentCenter      Alignment = 0
	AlignmentTop         Alignment = 1
	AlignmentBottom      Alignment = 2
	AlignmentLeft        Alignment = 3
	AlignmentRight       Alignment = 4
	AlignmentTopLeft     Alignment = 5
	AlignmentTopRight    Alignment = 6
	AlignmentBottomLeft  Alignment = 7
	AlignmentBottomRight Alignment = 8
)

/*
Size
*/
type Size struct {
	Scale  ScaleType
	Width  int
	Height int
}

type UIElement interface {
	Draw(img *image.RGBA, window *glfw.Window)
	SetProperties(size Size, center Point)
	Debug()
}

type Properties struct {
	MaxSize   Size
	Center    Point
	Size      Size
	Alignment Alignment
	Padding   Padding
	Function   func()
}

type Style struct {
	Color color.Color
}

type StyleText struct {
	Font      string
	FontSize  int
	FontColor color.Color
}

type Point struct {
	X int
	Y int
}

func Draw(img *image.RGBA, window *glfw.Window, props *Properties, style Style) {
	maxWidth, maxHeight := GetMaxDimensions(props, window)
	width, height := GetDimensions(props, maxWidth, maxHeight)
	centerX, centerY := GetCenter(props, width, height, maxWidth, maxHeight)

	x, y := window.GetCursorPos()

	r, g, b, _ := style.Color.RGBA()

	if x > float64(centerX) && x < float64(centerX+width) && y > float64(centerY) && y < float64(centerY+height) {
		if r % 255 > 30 {
			r -= 30
		}
		if g % 255 > 30 {
			g -= 30
		}
		if b % 255 > 30 {
			b -= 30
		}
		if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
			if props.Function != nil {
				props.Function()
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(width)
	for i := 0; i < width; i++ {
		go func(i int) {
			for j := 0; j < height; j++ {
				trueJ := utils.RESOLUTION_Y - (centerY + j) - 1
				img.Set(centerX+i, (centerY + trueJ), color.RGBA{byte(r), byte(g), byte(b), 255})
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// func GetMaxDimensions(props *Properties, window *glfw.Window) (int, int) {
// 	var w, h = window.GetSize()

// 	if props.MaxSize.Width == 0 || props.MaxSize.Height == 0 {
// 		props.MaxSize.Width = w
// 		props.MaxSize.Height = h
// 		props.MaxSize.Scale = ScalePixel
// 	}

// 	maxWidth := props.MaxSize.Width
// 	maxHeight := props.MaxSize.Height
// 	if props.MaxSize.Scale == ScaleRelative {
// 		maxWidth = w * props.MaxSize.Width / 100
// 		maxHeight = h * props.MaxSize.Height / 100
// 	}

// 	return maxWidth, maxHeight
// }


func GetDimensions(props *Properties, maxWidth, maxHeight int) (int, int) {
	if props.Size.Width == 0 || props.Size.Height == 0 {
		props.Size.Width = props.MaxSize.Width
		props.Size.Height = props.MaxSize.Height
		props.Size.Scale = ScalePixel
	}

	width := props.Size.Width
	height := props.Size.Height
	if props.Size.Scale == ScaleRelative {
		width = maxWidth * props.Size.Width / 100
		height = maxHeight * props.Size.Height / 100
	}

	if props.Padding.Scale == ScaleRelative {
		height -= (maxHeight * props.Padding.Top / 100) + (maxHeight * props.Padding.Bottom / 100)
		width -= (maxWidth * props.Padding.Left / 100) + (maxWidth * props.Padding.Right / 100)
	} else {
		height -= props.Padding.Top + props.Padding.Bottom
		width -= props.Padding.Left + props.Padding.Right
	}

	return width, height
}

func GetCenter(props *Properties, width, height, maxWidth, maxHeight int) (int, int) {
	centerX := props.Center.X
	centerY := props.Center.Y
	
	switch props.Alignment {
	case AlignmentBottom:
		centerY -= height/2 - maxHeight/2
	case AlignmentTop:
		centerY += height/2 - maxHeight/2
	case AlignmentLeft:
		centerX += width/2 - maxWidth/2
	case AlignmentRight:
		centerX -= width/2 - maxWidth/2
	case AlignmentTopLeft:
		centerX += width/2 - maxWidth/2
		centerY += height/2 - maxHeight/2
	case AlignmentTopRight:
		centerX -= width/2 - maxWidth/2
		centerY += height/2 - maxHeight/2
	case AlignmentBottomLeft:
		centerX += width/2 - maxWidth/2
		centerY -= height/2 - maxHeight/2
	case AlignmentBottomRight:
		centerX -= width/2 - maxWidth/2
		centerY -= height/2 - maxHeight/2
	}

	if props.Padding.Scale == ScaleRelative {
		centerX += (maxWidth * props.Padding.Left / 100) - (maxWidth * props.Padding.Right / 100)
		centerY += (maxHeight * props.Padding.Top / 100) - (maxHeight * props.Padding.Bottom / 100)
	} else {
		centerX += props.Padding.Left - props.Padding.Right
		centerY += props.Padding.Top - props.Padding.Bottom
	}

	centerX -= width / 2
	centerY -= height / 2

	return centerX, centerY
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
