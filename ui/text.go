package ui

// import (
// 	"image"
// 	"image/color"

// 	"github.com/go-gl/glfw/v3.3/glfw"

// 	"io/ioutil"
// 	"log"

// 	"github.com/goki/freetype"
// )


// type Text struct {
// 	Properties *Properties
// 	StyleText  StyleText
// }

// func (text Text) Draw(img *image.RGBA, window *glfw.Window) {
// 	//Draw(img, window, text.Properties, Style{})

// 	maxWidth, maxHeight := GetMaxDimensions(text.Properties, window)
// 	width, height := GetDimensions(text.Properties, maxWidth, maxHeight)
// 	centerX, centerY := GetCenter(text.Properties, width, height, maxWidth, maxHeight)
// 	

// 	drawText(img, []string{"Hello, World!"}, text.StyleText.Font, float64(text.StyleText.FontSize), text.StyleText.FontColor, centerX, centerY)

// }


// func (text Text) SetProperties(size Size, center Point) {
// 	text.Properties.MaxSize = size
// 	text.Properties.Center = center
// }

// func (text Text) Debug() {
// 	println(text.Properties.Center.Y)
// }

// func drawText(img *image.RGBA, text []string, font string, fontSize float64, fontColor color.Color, x, y int) {

// 	// Load font
// 	fontBytes, err := ioutil.ReadFile(font)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	f, err := freetype.ParseFont(fontBytes)

// 	// Load freetype context
// 	c := freetype.NewContext()
// 	c.SetDPI(72)
// 	c.SetFont(f)
// 	c.SetFontSize(fontSize)
// 	c.SetClip(img.Bounds())
// 	c.SetDst(img)
// 	c.SetSrc(image.NewUniform(fontColor))

// 	// Draw the text
// 	pt := freetype.Pt(x, y+int(c.PointToFixed(fontSize)>>6))
// 	for _, s := range text {
// 		_, err := c.DrawString(s, pt)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		pt.Y += c.PointToFixed(fontSize * 1.5)
// 	}
// }
