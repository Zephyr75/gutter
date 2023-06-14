
package test

import (
  "github.com/Zephyr75/gutter/ui"
)

func exitButton() ui.UIElement {
  return ui.Button{
          Properties: ui.Properties{
            Alignment: ui.AlignmentTopLeft,
            // Padding: ui.PaddingEqual(ui.ScaleRelative, 25), 
            Size: ui.Size{
              Scale:  ui.ScalePixel,
              Width:  50,
              Height: 30,
            },
            Function: func() {
              // window.SetShouldClose(true)
            },
          },
          Style: ui.Style{
            Color: blue,
          },
        }
}
