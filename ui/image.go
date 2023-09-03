package ui


import (
	"image"
	_ "image/png"


	"github.com/go-gl/glfw/v3.3/glfw"
)




type Image struct {
	Properties Properties
	Style	   Style
	Child      UIElement
  Name       string
}

func (image Image) Initialize(skip SkipAlignment) UIElement {
  image.Properties = DefaultProperties(image.Properties, skip, UIImage)
  image.Style = DefaultStyle(image.Style)
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

	Draw(img, window, image.Properties, image.Style, image.Name)
	
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


