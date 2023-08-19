package ui


import (
	"os"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/go-gl/glfw/v3.3/glfw"
)




type Image struct {
	Properties Properties
	Style	   Style
	Child      UIElement
}

func (image Image) Initialize(skip SkipAlignment) UIElement {
  image.Properties = DefaultProperties(image.Properties, skip)
  return image
}

func (image Image) Draw(img *image.RGBA, window *glfw.Window) {

  if !image.Properties.Initialized {
    image = image.Initialize(SkipAlignmentNone).(Image)
  }

  image = ApplyRelative(image).(Image)

  image = ApplyAlignment(image).(Image)

  image = ApplyPadding(image).(Image)

  if image.Child != nil {
    image.Child = image.Child.SetParent(&image.Properties)
    image.Child = image.Child.Initialize(SkipAlignmentNone)
  }

  // if b > 200 {
  //   fmt.Println(button)
  // }

	Draw(img, window, image.Properties, image.Style)
	
	if image.Child != nil {
    props := image.Child.GetProperties()
		image.Child.SetProperties(props.Size, image.Properties.Center)
		image.Child.Draw(img, window)
	}
}


func (image Image) SetProperties(size Size, center Point) UIElement {
	image.Properties.Size = size
	image.Properties.Center = center
  return image
}

func (image Image) SetParent(parent *Properties) UIElement {
  image.Properties.Parent = parent
  return image
}

func (image Image) GetProperties() Properties {
  return image.Properties
}

func (image Image) Debug() {
	println(image.Properties.Center.Y)
}

func CreateTexture(file string) uint32 {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
  )

  // gl.GenerateMipmap(gl.TEXTURE_2D)

	return texture
}
