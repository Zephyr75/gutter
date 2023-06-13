package main

import (
	// "fmt"
	"image"
	"image/color"

	"runtime"

  // "sync"

  // "image/draw"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"gutter/ui"
	"gutter/utils"

	"github.com/disintegration/imaging"
)

func init() {
    // GLFW: This is needed to arrange that main() runs on main thread.
    // See documentation for functions that are only allowed to be called from the main thread.
    runtime.LockOSThread()
}

func main() {
    err := glfw.Init()
    if err != nil {
        panic(err)
    }
    defer glfw.Terminate()

    window, err := glfw.CreateWindow(utils.RESOLUTION_X, utils.RESOLUTION_Y, "Gutter", nil, nil)
    if err != nil {
        panic(err)
    }

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



    for !window.ShouldClose() {

        var w, h = window.GetSize()

        utils.RESOLUTION_X = w
        utils.RESOLUTION_Y = h

        var img = image.NewRGBA(image.Rect(0, 0, w, h))

        // -------------------------
        // MODIFY OR LOAD IMAGE HERE
        // -------------------------

        
        green := color.RGBA{158, 206, 106, 255}
        white := color.RGBA{192, 202, 245, 255}
        blue := color.RGBA{122, 162, 247, 255}
        red := color.RGBA{247, 118, 142, 255}
        black := color.RGBA{36, 40, 59, 255}

        parent := ui.Row{
          Style: ui.Style{
            Color: white,
          },
          Children: []ui.UIElement{
            ui.Button{
              Properties: ui.Properties{
                Alignment: ui.AlignmentBottom,
                Size: ui.Size{
                  Scale:  ui.ScalePixel,
                  Width:  100,
                  Height: 100,
                }, 
              },
              Style: ui.Style{
                Color: green,
              },
            },
            ui.Column{
              Properties: ui.Properties{
                Padding: ui.PaddingSideBySide(ui.ScaleRelative, 0, 25, 25, 0),
              },
              Style: ui.Style{
                Color: white,
              },
              Children: []ui.UIElement{
                ui.Button{
                  Properties: ui.Properties{
                    Size: ui.Size{
                      Scale:  ui.ScaleRelative,
                      Width:  50,
                      Height: 50,
                    },
                  },
                  Style: ui.Style{
                    Color: green,
                  },
                },
                ui.Button{
                  Properties: ui.Properties{
                    Size: ui.Size{
                      Scale:  ui.ScaleRelative,
                      Width:  50,
                      Height: 50,
                    },
                  },
                  Style: ui.Style{
                    Color: red,
                  },
                },
              },
            },
            // ui.Button{
            //   Properties: ui.Properties{
            //     Size: ui.Size{
            //       Scale:  ui.ScaleRelative,
            //       Width:  50,
            //       Height: 100,
            //     }, 
            //   },
            //   Style: ui.Style{
            //     Color: green,
            //   },
            // },
            // ui.Button{
            //   Properties: ui.Properties{
            //     Size: ui.Size{
            //       Scale:  ui.ScalePixel,
            //       Width:  200,
            //       Height: 100,
            //     }, 
            //   },
            //   Style: ui.Style{
            //     Color: blue,
            //   },
            // },
            ui.Button{
              Style: ui.Style{
                Color: red,
              },
              // Child: ui.Button{
              //   Properties: ui.Properties{
              //     Size: ui.Size{
              //       Scale:  ui.ScalePixel,
              //       Width:  50,
              //       Height: 50,
              //     },
              //     Alignment: ui.AlignmentCenter,
              //   },
              //   Style: ui.Style{
              //     Color: blue,
              //   },
              // },

              Child: ui.Text{
                Properties: ui.Properties{
                  Alignment: ui.AlignmentCenter,
                  //Padding:   ui.PaddingEqual(ui.ScalePixel, 100),
                  Size: ui.Size{
                    Scale:  ui.ScalePixel,
                    Width:  100,
                    Height: 50,
                  },
                },
                StyleText: ui.StyleText{
                  Font: "JBMono.ttf",
                  FontSize: 20,
                  FontColor: black,
                },
              },
            },
          },
        }

        parent.Draw(img, window)

        exit := ui.Button{
          Properties: ui.Properties{
            Alignment: ui.AlignmentTopLeft,
            // Padding: ui.PaddingEqual(ui.ScaleRelative, 25), 
            Size: ui.Size{
              Scale:  ui.ScaleRelative,
              Width:  10,
              Height: 10,
            },
            Function: func() {
              window.SetShouldClose(true)
            },
          },
          Style: ui.Style{
            Color: blue,
          },
        }

        exit.Draw(img, window)


        flippedImg := imaging.FlipV(img)


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
          println("FPS:", i)
          i = 0
          time = glfw.GetTime()
        }
        ///



    }
}
