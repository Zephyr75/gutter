package test

import (

	"github.com/Zephyr75/gutter/ui"
)

var (
  counter int = 1
)


func MainWindow() ui.UIElement {
  return ui.Row{
          Style: ui.Style{
            Color: black,
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
                Color: black,
              },
              Children: []ui.UIElement{
                ui.Button{
                  Properties: ui.Properties{
                    Size: ui.Size{
                      Scale:  ui.ScaleRelative,
                      Width:  50,
                      Height: 50,
                    },
                    Function: func() {
                      counter += 1
                    },
                  },
                  Style: ui.Style{
                    Color: green,
                    CornerRadius: 25,
                  },
                },
                ui.Button{
                  Properties: ui.Properties{
                    Size: ui.Size{
                      Scale:  ui.ScaleRelative,
                      Width:  50,
                      Height: 50,
                    },
                    Function: func() {
                      counter -= 1
                    },
                  },
                  Style: ui.Style{
                    Color: red,
                    BorderColor: white,
                    BorderWidth: 10,
                    CornerRadius: 25,
                  },
                },
              },
            },
            ui.Button{
              Style: ui.Style{
                Color: red,
              },
              Child: ui.Text{
                Properties: ui.Properties{
                  Alignment: ui.AlignmentTopLeft,
                  //Padding:   ui.PaddingEqual(ui.ScalePixel, 100),
                  Size: ui.Size{
                    Scale:  ui.ScalePixel,
                    Width:  100,
                    Height: 50,
                  },
                },
                StyleText: ui.StyleText{
                  Font: "Comfortaa.ttf",
                  FontSize: counter,
                  FontColor: black,
                },
              },
            },
          },
        }
}
