
package ui


import (
	"image"
	_ "image/png"


	"github.com/Zephyr75/gutter/utils"
	"github.com/go-gl/glfw/v3.3/glfw"

  "github.com/nfnt/resize"
)




type Container struct {
	Properties Properties
	Style	   Style
	Child      UIElement
  Image       string
  ImageData  image.Image
}

func (container Container) Initialize(skip SkipAlignment) UIElement {
  container.Properties = DefaultProperties(container.Properties, skip, UIContainer)
  container.Style = DefaultStyle(container.Style)

  smallWidth := container.Properties.Size.Width - 2 * container.Style.BorderWidth
  smallHeight := container.Properties.Size.Height - 2 * container.Style.BorderWidth
  if container.Image != "" {
    texture, _ := utils.GetImageFromFilePath(container.Image)
    texture = resize.Resize(uint(smallWidth), uint(smallHeight), texture, resize.Lanczos3)
    container.ImageData = texture
  }
  return container
}

func (container Container) Draw(img *image.RGBA, window *glfw.Window) {

  if !container.Properties.Initialized {
    container = container.Initialize(SkipAlignmentNone).(Container)
  }

  container = ApplyRelative(container).(Container)

  container = ApplyAlignment(container).(Container)

  container = ApplyPadding(container).(Container)

  if container.Child != nil {
    container.Child = container.Child.SetParent(&container.Properties)
    container.Child = container.Child.Initialize(SkipAlignmentNone)
  }

  // if b > 200 {
  //   fmt.Println(button)
  // }

	Draw(img, window, container.Properties, container.Style, container.ImageData, nil)
	
	if container.Child != nil {
    props := container.Child.GetProperties()
		container.Child.SetProperties(props.Size, container.Properties.Center)
		container.Child.Draw(img, window)
	}
}


func (container Container) SetProperties(size Size, center Point) UIElement {
	container.Properties.Size = size
	container.Properties.Center = center
  return container
}

func (container Container) SetParent(parent *Properties) UIElement {
  container.Properties.Parent = parent
  return container
}

func (container Container) GetProperties() Properties {
  return container.Properties
}

func (container Container) Debug() {
	println(container.Properties.Center.Y)
}


