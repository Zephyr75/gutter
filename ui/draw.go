package ui

import (
	// "fmt"
	"image"
	"image/color"
	// "sync"

  "math"

	"github.com/go-gl/glfw/v3.3/glfw"


  "image/draw"
  "os"
  "github.com/nfnt/resize"
)



func Draw(img *image.RGBA, window *glfw.Window, props Properties, style Style, file string, hoverFile string) {

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
  borderWidth := style.BorderWidth

  cornerRadius := style.CornerRadius

  smallWidth := width - 2 * borderWidth
  smallHeight := height - 2 * borderWidth

  // Create a mask with rounded corners
  mask := image.NewRGBA(image.Rect(0, 0, width, height))
  offsetMask := image.NewRGBA(image.Rect(0, 0, smallWidth, smallHeight))
  if cornerRadius > 0 {
    draw.Draw(mask, mask.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
    draw.Draw(offsetMask, offsetMask.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
  }


  if cornerRadius > 0 {
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
        offsetMask.Set(x-1, y-1, color.Transparent)
      }
      for x := 0; x <= int(l); x++ {
        offsetMask.Set(smallWidth-x, y-1, color.Transparent)
      }
      for x := 0; x <= int(l); x++ {
        offsetMask.Set(x-1, smallHeight-y, color.Transparent)
      }
      for x := 0; x <= int(l); x++ {
        offsetMask.Set(smallWidth-x, smallHeight-y, color.Transparent)
      }
    }
  }

  var col, colBorder color.Color
  col = color.RGBA{0, 0, 0, 0}
  colBorder = color.RGBA{0, 0, 0, 0}
  if style.Color != nil {
    col = style.Color
  }
  if style.BorderColor != nil {
    colBorder = style.BorderColor
  }

  var texture, hoverTexture, borderTexture, blackTexture image.Image
  texture = image.NewUniform(col)
  hoverTexture = image.NewUniform(col)
  borderTexture = image.NewUniform(colBorder)
  blackTexture = image.NewUniform(color.RGBA{0, 0, 0, 55})
  if file != "" {
    texture, _ = getImageFromFilePath(file)
    texture = resize.Resize(uint(smallWidth), uint(smallHeight), texture, resize.Lanczos3)
  }
  if hoverFile != "" {
    hoverTexture, _ = getImageFromFilePath(hoverFile)
    hoverTexture = resize.Resize(uint(smallWidth), uint(smallHeight), hoverTexture, resize.Lanczos3)
  }




  if style.ShadowWidth > 0 {
    shadowWidth := style.ShadowWidth
    shadowAlignment := style.ShadowAlignment
    shadowColor := style.ShadowColor
    if style.ShadowColor != nil {
      shadowColor = style.ShadowColor
    }
    shadowTexture := image.NewUniform(shadowColor)
    shadowRect := image.Rect(centerX - width/2, centerY - height/2, centerX + width/2, centerY + height/2)
    switch shadowAlignment {
    case AlignmentCenter:
      shadowRect = image.Rect(centerX - width/2, centerY - height/2, centerX + width/2, centerY + height/2)
    case AlignmentBottom:
      shadowRect = image.Rect(centerX - width/2, centerY - height/2 + shadowWidth, centerX + width/2, centerY + height/2 + shadowWidth)
    case AlignmentTop:
      shadowRect = image.Rect(centerX - width/2, centerY - height/2 - shadowWidth, centerX + width/2, centerY + height/2 - shadowWidth)
    case AlignmentLeft:
      shadowRect = image.Rect(centerX - width/2 - shadowWidth, centerY - height/2, centerX + width/2 - shadowWidth, centerY + height/2)
    case AlignmentRight:
      shadowRect = image.Rect(centerX - width/2 + shadowWidth, centerY - height/2, centerX + width/2 + shadowWidth, centerY + height/2)
    case AlignmentTopLeft:
      shadowRect = image.Rect(centerX - width/2 - shadowWidth, centerY - height/2 - shadowWidth, centerX + width/2 - shadowWidth, centerY + height/2 - shadowWidth)
    case AlignmentTopRight:
      shadowRect = image.Rect(centerX - width/2 + shadowWidth, centerY - height/2 - shadowWidth, centerX + width/2 + shadowWidth, centerY + height/2 - shadowWidth)
    case AlignmentBottomLeft:
      shadowRect = image.Rect(centerX - width/2 - shadowWidth, centerY - height/2 + shadowWidth, centerX + width/2 - shadowWidth, centerY + height/2 + shadowWidth)
    case AlignmentBottomRight:
      shadowRect = image.Rect(centerX - width/2 + shadowWidth, centerY - height/2 + shadowWidth, centerX + width/2 + shadowWidth, centerY + height/2 + shadowWidth)
    }
    draw.DrawMask(img, shadowRect, shadowTexture, image.Point{}, mask, image.Point{}, draw.Over)
  }
      

  rect := image.Rect(centerX - width/2, centerY - height/2, centerX + width/2, centerY + height/2)
  offsetRect := image.Rect(centerX - width/2 + borderWidth, centerY - height/2 + borderWidth, centerX + width/2 - borderWidth, centerY + height/2 - borderWidth)

  if style.BorderWidth > 0 {
    if style.CornerRadius > 0 {
      draw.DrawMask(img, rect, borderTexture, image.Point{}, mask, image.Point{}, draw.Over)
      
      if !darken {
        draw.DrawMask(img, offsetRect, texture, image.Point{}, offsetMask, image.Point{}, draw.Over)
      } else {
        if hoverFile != "" {
          draw.DrawMask(img, offsetRect, hoverTexture, image.Point{}, offsetMask, image.Point{}, draw.Over)
        } else {
          draw.DrawMask(img, offsetRect, texture, image.Point{}, offsetMask, image.Point{}, draw.Over)
          draw.DrawMask(img, offsetRect, blackTexture, image.Point{}, offsetMask, image.Point{}, draw.Over)
        }
      }
    } else {
      draw.Draw(img, rect, borderTexture, image.Point{}, draw.Src)

      if !darken {
        draw.Draw(img, offsetRect, texture, image.Point{}, draw.Over)
      } else {
        if hoverFile != "" {
          draw.Draw(img, offsetRect, hoverTexture, image.Point{}, draw.Over)
        } else {
          draw.Draw(img, offsetRect, texture, image.Point{}, draw.Over)
          draw.Draw(img, offsetRect, blackTexture, image.Point{}, draw.Over)
        }
      }
    }
  } else {
    if style.CornerRadius > 0 {
      if !darken {
        draw.DrawMask(img, rect, texture, image.Point{}, mask, image.Point{}, draw.Over)
      } else {
        if hoverFile != "" {
          draw.DrawMask(img, rect, hoverTexture, image.Point{}, mask, image.Point{}, draw.Over)
        } else {
          draw.DrawMask(img, rect, texture, image.Point{}, mask, image.Point{}, draw.Over)
          draw.DrawMask(img, rect, blackTexture, image.Point{}, mask, image.Point{}, draw.Over)
        }
      }
      
    } else {
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
