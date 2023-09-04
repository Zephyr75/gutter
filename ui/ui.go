package ui

import (
	// "fmt"
	"image"
	"image/color"
	// "sync"

  "math"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/Zephyr75/gutter/utils"

  "image/draw"
  "os"
  "github.com/nfnt/resize"
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

type SkipAlignment byte

const (
  SkipAlignmentNone SkipAlignment = 0
  SkipAlignmentHoriz SkipAlignment = 1
  SkipAlignmentVert SkipAlignment = 2
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
  Initialize(skip SkipAlignment) UIElement
  SetParent(parent *Properties) UIElement
}

type UIType byte

const (
  UIContainer UIType = 0
  UIButton UIType = 1
  UIImage UIType = 2
  UIRow UIType = 3
  UIColumn UIType = 4
  UIText UIType = 5
)

type Properties struct {
	Center    Point
	Size      Size
	Alignment Alignment
	Padding   Padding
	Function  func()
  Parent    *Properties
  Initialized bool
  Skip       SkipAlignment
  Type       UIType
}

func DefaultProperties(props Properties, skip SkipAlignment, uitype UIType) Properties {
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
      Skip: skip,
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
    Skip: skip,
    Type: uitype,
  }
}

func DefaultStyle (style Style) Style {
  newStyle := style
  if newStyle.Color == nil {
    newStyle.Color = color.RGBA{0, 0, 0, 255}
  }
  return newStyle
}

type Style struct {
	Color color.Color
  BorderColor color.Color
  BorderWidth int
  CornerRadius int

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

func Draw(img *image.RGBA, window *glfw.Window, props Properties, style Style, file string) {

  // fmt.Println("Draw: ", props.Center.X, " ", props.Center.Y, " ", props.Size.Width, " ", props.Size.Height)

  width := props.Size.Width
  height := props.Size.Height
  centerX := props.Center.X
  centerY := props.Center.Y

  // fmt.Println("Center: ", centerX, " ", centerY, " ", width, " ", height)

	x, y := window.GetCursorPos()
  darken := false


	if x > float64(centerX - width/2) && x < float64(centerX + width/2) && y > float64(centerY - height/2) && y < float64(centerY + height/2) && props.Type == UIButton {
    darken = true
		
		if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
			if props.Function != nil {
				props.Function()
			}
		}
	}
  col := style.Color

  r2, g2, b2, _ := style.Color.RGBA()

  if style.BorderColor != nil {
    r2, g2, b2, _ = style.BorderColor.RGBA()
  }

  colBorder := color.RGBA{byte(r2), byte(g2), byte(b2), 255}

  borderWidth := style.BorderWidth

  cornerRadius := style.CornerRadius

  // Create a mask with rounded corners
  mask := image.NewRGBA(image.Rect(0, 0, width, height))
  draw.Draw(mask, mask.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

  offsetMask := image.NewRGBA(image.Rect(0, 0, width, height))
  draw.Draw(offsetMask, offsetMask.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)


  // Outside mask
	for y := 0; y <= cornerRadius; y++ {
		l := math.Round(float64(cornerRadius) - math.Sqrt(float64(2*y*cornerRadius-y*y)))
    for x := 0; x <= int(l); x++ {
      mask.Set(x-1, y-1, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      mask.Set(width-x, y-1, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      mask.Set(x-1, height-y, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      mask.Set(width-x, height-y, color.Transparent)
    }
	}

  smallCornerRadius := cornerRadius - borderWidth
  for y := 0; y <= smallCornerRadius; y++ {
		l := math.Round(float64(smallCornerRadius) - math.Sqrt(float64(2*y*smallCornerRadius-y*y)))
    for x := 0; x <= int(l); x++ {
      offsetMask.Set(borderWidth+x-1, borderWidth+y-1, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      offsetMask.Set(width-borderWidth-x, borderWidth+y-1, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      offsetMask.Set(borderWidth+x-1, height-borderWidth-y, color.Transparent)
    }
    for x := 0; x <= int(l); x++ {
      offsetMask.Set(width-borderWidth-x, height-borderWidth-y, color.Transparent)
    }
	}



  for x := 0; x <= width; x++ {
    for y := 0; y <= borderWidth; y++ {
      offsetMask.Set(x, y, color.Transparent)
      offsetMask.Set(x, height-y, color.Transparent)
    }
  }
  for y := 0; y <= height; y++ {
    for x := 0; x <= borderWidth; x++ {
      offsetMask.Set(x, y, color.Transparent)
      offsetMask.Set(width-x, y, color.Transparent)
    }
  }

  var texture, borderTexture, blackTexture image.Image
  texture = image.NewUniform(col)
  borderTexture = image.NewUniform(colBorder)
  blackTexture = image.NewUniform(color.RGBA{0, 0, 0, 55})
  if file != "" {
    texture, _ = getImageFromFilePath(file)
    texture = resize.Resize(uint(width - 2), uint(height - 2), texture, resize.Lanczos3)
  }


  rect := image.Rect(centerX - width/2, centerY - height/2, centerX + width/2, centerY + height/2)
  if style.BorderWidth > 0 {
    if style.CornerRadius > 0 {
      draw.DrawMask(img, rect, borderTexture, image.Point{}, mask, image.Point{}, draw.Over)
      draw.DrawMask(img, rect, texture, image.Point{}, offsetMask, image.Point{}, draw.Over)
      if darken {
        draw.DrawMask(img, rect, blackTexture, image.Point{}, mask, image.Point{}, draw.Over)
      }
    } else {
      draw.Draw(img, rect, borderTexture, image.Point{}, draw.Src)
      draw.Draw(img, rect, texture, image.Point{}, draw.Src)
      if darken {
        draw.Draw(img, rect, blackTexture, image.Point{}, draw.Over)
      }

    }
  } else {
    if style.CornerRadius > 0 {
      draw.DrawMask(img, rect, texture, image.Point{}, mask, image.Point{}, draw.Over)
      if darken {
        draw.DrawMask(img, rect, blackTexture, image.Point{}, mask, image.Point{}, draw.Over)
      }
    } else {
      draw.Draw(img, rect, texture, image.Point{}, draw.Src)
      if darken {
        draw.Draw(img, rect, blackTexture, image.Point{}, draw.Over)
      }
    }
  }

  
  
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
    newY = parent.Center.Y + parent.Size.Height / 2 - props.Size.Height / 2
  case AlignmentTop:
    newY = parent.Center.Y - parent.Size.Height / 2 + props.Size.Height / 2
  case AlignmentLeft:
    newX = parent.Center.X - parent.Size.Width / 2 + props.Size.Width / 2
  case AlignmentRight:
    newX = parent.Center.X + parent.Size.Width / 2 - props.Size.Width / 2
  case AlignmentTopLeft:
    newX = parent.Center.X - parent.Size.Width / 2 + props.Size.Width / 2
    newY = parent.Center.Y - parent.Size.Height / 2 + props.Size.Height / 2
  case AlignmentTopRight:
    newX = parent.Center.X + parent.Size.Width / 2 - props.Size.Width / 2
    newY = parent.Center.Y - parent.Size.Height / 2 + props.Size.Height / 2
  case AlignmentBottomLeft:
    newX = parent.Center.X - parent.Size.Width / 2 + props.Size.Width / 2
    newY = parent.Center.Y + parent.Size.Height / 2 - props.Size.Height / 2
  case AlignmentBottomRight:
    newX = parent.Center.X + parent.Size.Width / 2 - props.Size.Width / 2
    newY = parent.Center.Y + parent.Size.Height / 2 - props.Size.Height / 2
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

func getImageFromFilePath(filePath string) (image.Image, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    image, _, err := image.Decode(f)
    return image, err
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
