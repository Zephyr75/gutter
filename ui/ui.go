package ui

import (
	// "fmt"
	"image"
	"image/color"
	// "sync"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/Zephyr75/gutter/utils"
	"strconv"
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
	SkipAlignmentNone  SkipAlignment = 0
	SkipAlignmentHoriz SkipAlignment = 1
	SkipAlignmentVert  SkipAlignment = 2
)

/*
Size
*/
type Size struct {
	Scale  ScaleType
	Width  int
	Height int
}

func (s Size) ToString() string {
	return strconv.Itoa(s.Width) + strconv.Itoa(s.Height) + strconv.FormatBool(bool(s.Scale))
}

type UIElement interface {
	Draw(img *image.RGBA, window *glfw.Window) []Area
	SetProperties(size Size, center Point) UIElement
	GetProperties() Properties
	Initialize(skip SkipAlignment) UIElement
	SetParent(parent *Properties) UIElement
	ToString() string
}

type UIType byte

const (
	UIContainer UIType = 0
	UIButton    UIType = 1
	UIImage     UIType = 2
	UIRow       UIType = 3
	UIColumn    UIType = 4
	UIText      UIType = 5
)

func (u UIType) ToString() string {
	switch u {
	case UIContainer:
		return "UIContainer"
	case UIButton:
		return "UIButton"
	case UIImage:
		return "UIImage"
	case UIRow:
		return "UIRow"
	case UIColumn:
		return "UIColumn"
	case UIText:
		return "UIText"
	default:
		return "Unknown"
	}
}

type Properties struct {
	Center      Point
	Size        Size
	Alignment   Alignment
	Padding     Padding
	Parent      *Properties
	Initialized bool
	Skip        SkipAlignment
	Type        UIType
}

func (p Properties) ToString() string {
	return p.Center.ToString() + p.Size.ToString() + p.Type.ToString()
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
			Center:      Point{utils.RESOLUTION_X / 2, utils.RESOLUTION_Y / 2},
			Size:        Size{ScalePixel, utils.RESOLUTION_X, utils.RESOLUTION_Y},
			Alignment:   AlignmentCenter,
			Padding:     PaddingEqual(ScalePixel, 0),
			Parent:      nil,
			Initialized: true,
			Skip:        skip,
			Type:        UIContainer,
		}
	}

	return Properties{
		Center:      newCenter,
		Size:        newSize,
		Alignment:   props.Alignment,
		Padding:     props.Padding,
		Parent:      newParent,
		Initialized: true,
		Skip:        skip,
		Type:        uitype,
	}
}

func DefaultStyle(style Style) Style {
	newStyle := style
	if newStyle.Color == nil {
		newStyle.Color = color.RGBA{0, 0, 0, 255}
	}
	return newStyle
}

type Style struct {
	Color color.Color
}

func (s Style) ToString() string {
	r, g, b, a := s.Color.RGBA()
	return strconv.Itoa(int(r)) + strconv.Itoa(int(g)) + strconv.Itoa(int(b)) + strconv.Itoa(int(a))
}

func DefaultStyleText(style StyleText) StyleText {
	newStyle := style
	if newStyle.Font == "" {
		newStyle.Font = "Arial"
	}
	if newStyle.FontSize == 0 {
		newStyle.FontSize = 12
	}
	if newStyle.FontColor == nil {
		newStyle.FontColor = color.RGBA{0, 0, 0, 255}
	}
	return newStyle
}

type StyleText struct {
	Font      string
	FontSize  int
	FontColor color.Color
}

func (s StyleText) ToString() string {
	r, g, b, a := s.FontColor.RGBA()
	return s.Font + strconv.Itoa(s.FontSize) + strconv.Itoa(int(r)) + strconv.Itoa(int(g)) + strconv.Itoa(int(b)) + strconv.Itoa(int(a))
}

type Point struct {
	X int
	Y int
}

func (p Point) ToString() string {
	return strconv.Itoa(p.X) + strconv.Itoa(p.Y)
}

type Area struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
	Function   func()
}

func (a Area) ToString() string {
	return strconv.Itoa(int(a.Top)) + strconv.Itoa(int(a.Right)) + strconv.Itoa(int(a.Bottom)) + strconv.Itoa(int(a.Left))
}