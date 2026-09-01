# Gutter :sweat_drops:
A Flutter inspired UI framework written in Go with OpenGL rendering

## Running the demo

```sh
go run .          # font and image paths in main.go are relative to the repo root
go test ./...     # layout, hashing and text, none of which need a window
OUT=/tmp/demo.png go test . -run TestDemoRenders   # rasterise the demo to a PNG
```

The demo prints `FPS` and `redraws/s` to the terminal. The second number is the one
that matters: the widget tree is rebuilt every frame, but it is only rasterised
when its hash or the set of hovered widgets actually changed, so an idle window
sits at zero redraws.

## Documentation

### Getting started
```go
package main

import (...)

func main() {
  myApp := app.App {
    Name: "Gutter",
    Width: 800,
    Height: 600,
  }
  myApp.Run(test.MainWindow)
}

func MainWindow(app core.App) ui.UIElement {
    return ui.Row{
        ...
    }
}

```

<details>
<summary>Widgets</summary>
  
### Row
```go
ui.Row{
    Style: ui.Style{
        Color: black,
    },
    Children: []ui.UIElement{
        ... 
    },
}
```

### Column
```go
ui.Row{
    Style: ui.Style{
        Color: black,
    },
    Children: []ui.UIElement{
        ... 
    },
}
```

### Button
```go
ui.Button{
    Properties: ui.Properties{
        Size: ui.Size{
            Scale:  ui.ScaleRelative,
            Width:  50,
            Height: 50,
        }, 
    },
    Style: ui.Style{
        BorderWidth: 10,
        BorderColor: white,
        CornerRadius: 25,
        Color: blue,
    },
    Image: "background.png",
    HoverImage: "hover.png",
    Function: func() {
        app.Quit()
    },
},
```

### Text

`Content` is the string to draw; embedded newlines start a new line. The text is
left-aligned inside the widget's own box and the block is centred vertically, so
padding is what positions it. It is clipped to that box rather than painted over
its neighbours.

```go
ui.Text{
    Properties: ui.Properties{
      Alignment: ui.AlignmentTopLeft,
      Padding: ui.PaddingSymmetric(ui.ScalePixel, 0, 10),
      Size: ui.Size{
        Scale:  ui.ScalePixel,
        Width:  100,
        Height: 50,
      },
    },
    StyleText: ui.StyleText{
      Font: "Comfortaa.ttf",
      FontSize: 15,
      FontColor: black,
    },
    Content: "hello\nworld",
}
```

### Container

```go
ui.Container{
    Style: ui.Style{
        Color: red,
    },
    Child: ui.Text{
        ...
    },
},
```

</details>
