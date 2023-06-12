package ui

import (
	// "fmt"
	"image"
	"image/color"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"

	"gutter/utils"
)

type ScaleType bool

const (
	ScalePixel    ScaleType = true
	ScaleRelative ScaleType = false
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
	SetProperties(size Size, center Point) UIElement
  GetProperties() Properties
	Debug()
  Initialize() UIElement
  SetParent(parent *Properties) UIElement
}

type Properties struct {
	Center    Point
	Size      Size
	Alignment Alignment
	Padding   Padding
	Function  func()
  Parent    *Properties
  Initialized bool
}

func DefaultProperties(props Properties) Properties {
  newSize := props.Size
  if props.Size.Width == 0 && props.Size.Height == 0 {
    newSize = Size{ScaleRelative, 100, 100}
    if props.Parent == nil {
      // fmt.Println("Parent is nil")
      newSize = Size{ScalePixel, utils.RESOLUTION_X, utils.RESOLUTION_Y}
    }
  }

  newCenter := props.Center
  if props.Center.X == 0 && props.Center.Y == 0 {
    newCenter = Point{utils.RESOLUTION_X / 2, utils.RESOLUTION_Y / 2}
  }

  newParent := props.Parent
  if props.Parent == nil {
    newParent = &Properties{
      Center: Point{utils.RESOLUTION_X / 2, utils.RESOLUTION_Y / 2},
      Size: Size{ScalePixel, utils.RESOLUTION_X, utils.RESOLUTION_Y},
      Alignment: AlignmentCenter,
      Padding: PaddingEqual(ScalePixel, 0),
      Function: nil,
      Parent: nil,
      Initialized: true,
    }
  }

    
  return Properties{
    Center: newCenter,
    Size: newSize,
    Alignment: props.Alignment,
    Padding: props.Padding,
    Function: props.Function,
    Parent: newParent,
    Initialized: true,
  }
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

func Draw(img *image.RGBA, window *glfw.Window, props Properties, style Style) {

  // fmt.Println("Draw: ", props.Center.X, " ", props.Center.Y, " ", props.Size.Width, " ", props.Size.Height)

	width, height := GetScreenSize(props)
	centerX, centerY := GetScreenCenter(props)

  // fmt.Println("Center: ", centerX, " ", centerY, " ", width, " ", height)

	x, y := window.GetCursorPos()

	r, g, b, _ := style.Color.RGBA()

	if x > float64(centerX - width/2) && x < float64(centerX + width/2) && y > float64(centerY - height/2) && y < float64(centerY + height/2) {
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
        trueI := centerX - width/2 + i
				trueJ := centerY - height/2 + j
        trueJ = utils.RESOLUTION_Y - trueJ
				img.Set(trueI, trueJ, color.RGBA{byte(r), byte(g), byte(b), 255})
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func GetScreenSize(props Properties) (int, int) {
	width := props.Size.Width
	height := props.Size.Height
	if props.Size.Scale == ScaleRelative {
    parentProps := props.Parent
    width = parentProps.Size.Width * props.Size.Width / 100
    height = parentProps.Size.Height * props.Size.Height / 100
	}
	return width, height
}

func GetScreenCenter(props Properties) (int, int) {
	centerX := props.Center.X
	centerY := props.Center.Y
	
	// switch props.Alignment {
	// case AlignmentBottom:
	// 	centerY -= height/2 - maxHeight/2
	// case AlignmentTop:
	// 	centerY += height/2 - maxHeight/2
	// case AlignmentLeft:
	// 	centerX += width/2 - maxWidth/2
	// case AlignmentRight:
	// 	centerX -= width/2 - maxWidth/2
	// case AlignmentTopLeft:
	// 	centerX += width/2 - maxWidth/2
	// 	centerY += height/2 - maxHeight/2
	// case AlignmentTopRight:
	// 	centerX -= width/2 - maxWidth/2
	// 	centerY += height/2 - maxHeight/2
	// case AlignmentBottomLeft:
	// 	centerX += width/2 - maxWidth/2
	// 	centerY -= height/2 - maxHeight/2
	// case AlignmentBottomRight:
	// 	centerX -= width/2 - maxWidth/2
	// 	centerY -= height/2 - maxHeight/2
	// }

	// if props.Padding.Scale == ScaleRelative {
	// 	centerX += (maxWidth * props.Padding.Left / 100) - (maxWidth * props.Padding.Right / 100)
	// 	centerY += (maxHeight * props.Padding.Top / 100) - (maxHeight * props.Padding.Bottom / 100)
	// } else {
	// 	centerX += props.Padding.Left - props.Padding.Right
	// 	centerY += props.Padding.Top - props.Padding.Bottom
	// }

	// centerX -= width / 2
	// centerY -= height / 2

	return centerX, centerY
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
  newCenter := Point{props.Center.X + horizOffset / 2, props.Center.Y + vertOffset / 2}
  // newCenter := Point{props.Center.X, props.Center.Y}
  return element.SetProperties(newSize, newCenter)
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
