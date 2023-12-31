package ui

import (
	"image"
	"image/color"

	"github.com/go-gl/glfw/v3.3/glfw"

	"io/ioutil"
	"log"

	"github.com/goki/freetype"
)

type Text struct {
	Properties Properties
	StyleText  StyleText
}

func (text Text) Initialize(skip SkipAlignment) UIElement {
	text.Properties = DefaultProperties(text.Properties, skip, UIText)
	text.StyleText = DefaultStyleText(text.StyleText)
	return text
}

func (text Text) Draw(img *image.RGBA, window *glfw.Window) []Area {
	//Draw(img, window, text.Properties, Style{})

	if !text.Properties.Initialized {
		text = text.Initialize(SkipAlignmentNone).(Text)
	}

	text = ApplyRelative(text).(Text)

	text = ApplyAlignment(text).(Text)

	text = ApplyPadding(text).(Text)

	drawText(img, []string{"Hello, World!", "sdgsg"}, text.StyleText.Font, float64(text.StyleText.FontSize), text.StyleText.FontColor, text.Properties.Center.X, text.Properties.Center.Y)

	return []Area{}

}

func (text Text) SetProperties(size Size, center Point) UIElement {
	text.Properties.Size = size
	text.Properties.Center = center
	return text
}

func (text Text) SetParent(parent *Properties) UIElement {
	text.Properties.Parent = parent
	return text
}

func (text Text) GetProperties() Properties {
	return text.Properties
}

func drawText(img *image.RGBA, text []string, font string, fontSize float64, fontColor color.Color, x, y int) {

	// Load font
	fontBytes, err := ioutil.ReadFile(font)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Load freetype context
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(fontColor))

	// Draw the text
	pt := freetype.Pt(x, y+int(c.PointToFixed(fontSize)>>6))
	for _, s := range text {
		_, err := c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(fontSize * 1.5)
	}
}

func (text Text) ToString() string {
	return text.Properties.ToString() +
		text.StyleText.ToString()
}
