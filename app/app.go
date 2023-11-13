package app

import (
	"fmt"
	"image"

	"runtime"

	// "sync"

	// "image/draw"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/Zephyr75/gutter/ui"
	"github.com/Zephyr75/gutter/utils"

	"github.com/disintegration/imaging"
)

type App struct {
  Name string
  Width int
  Height int
  Window *glfw.Window
}



func init() {
    // GLFW: This is needed to arrange that main() runs on main thread.
    // See documentation for functions that are only allowed to be called from the main thread.
    runtime.LockOSThread()
}

func (app App) Run(widget func(app App) ui.UIElement) {
    err := glfw.Init()
    if err != nil {
        panic(err)
    }
    defer glfw.Terminate()

    utils.RESOLUTION_X = app.Width
    utils.RESOLUTION_Y = app.Height

    window, err := glfw.CreateWindow(utils.RESOLUTION_X, utils.RESOLUTION_Y, "Gutter", nil, nil)
    if err != nil {
        panic(err)
    }

    app.Window = window

    window.MakeContextCurrent()

    err = gl.Init()
    if err != nil {
        panic(err)
    }

    var texture uint32
    {
        gl.GenTextures(1, &texture)

        gl.BindTexture(gl.TEXTURE_2D, texture)
        gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
        gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
        gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
        gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

        gl.BindImageTexture(0, texture, 0, false, 0, gl.WRITE_ONLY, gl.RGBA8)
    }

    var framebuffer uint32
    {
        gl.GenFramebuffers(1, &framebuffer)
        gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
        gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

        gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
        gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
    }


    ///
    i := 0
    time := glfw.GetTime()
    ///

    lastInstance := ui.Container{}.Initialize(ui.SkipAlignmentNone).(ui.UIElement)
    var flippedImg *image.NRGBA

    first := true

    lastMap := map[string]bool{}
    areas := []ui.Area{}

    for !window.ShouldClose() {

        var w, h = window.GetSize()

        utils.RESOLUTION_X = w
        utils.RESOLUTION_Y = h

        var img = image.NewRGBA(image.Rect(0, 0, w, h))

        // -------------------------
        // MODIFY OR LOAD IMAGE HERE
        // -------------------------

        instance := widget(app)
        
        lastInstance = instance


        equal := true
        for _, area := range areas {
          if ui.MouseInBounds(window, area) != lastMap[area.ToString()] {
            equal = false
          }
          if ui.MouseInBounds(window, area) && window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
            area.Function()
          }
        }
    
        if lastInstance.ToString() != instance.ToString() || !equal || first {
          areas = instance.Draw(img, window)
          // Remove all empty areas
          newAreas := []ui.Area{}
          for _, area := range areas {
            if area.Left != 0 || area.Right != 0 || area.Top != 0 || area.Bottom != 0 {
              newAreas = append(newAreas, area)
            }
          }
          areas = newAreas
          flippedImg = imaging.FlipV(img)
          first = false
        }
        for _, area := range areas {
          lastMap[area.ToString()] = ui.MouseInBounds(window, area)
        }
          
        // window.SetShouldClose(true)
        // -------------------------

        gl.BindTexture(gl.TEXTURE_2D, texture)
        gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(flippedImg.Pix))

        gl.BlitFramebuffer(0, 0, int32(w), int32(h), 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.LINEAR)

        window.SwapBuffers()
        glfw.PollEvents()

        ///
        i++
        if glfw.GetTime()-time > 1 {
          fmt.Printf("\rFPS: %d", i)
          i = 0
          time = glfw.GetTime()
        }
        ///



    }
}


func (app App) Quit() {
  app.Window.SetShouldClose(true)
}

