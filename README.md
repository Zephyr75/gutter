# Gutter :sweat_drops:
A Flutter inspired UI framework using Go with OpenGL 

# Documentation

```go
package main

import (
	"github.com/Zephyr75/gutter/test"
	"github.com/Zephyr75/gutter/app"
)

func main() {

  app := app.App {
    Name: "Gutter",
    Width: 800,
    Height: 600,
  }

  app.Run(test.MainWindow)

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
  Image: "white_on_black.png",
},
```


</details>
