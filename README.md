# Gutter :sweat_drops:
A Flutter inspired UI framework written in Go with OpenGL rendering

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

```go
ui.Text{
    Properties: ui.Properties{
      Alignment: ui.AlignmentTopLeft,
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
