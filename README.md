# Gutter :sweat_drops:
A Flutter inspired UI framework written in Go, rendering through a draw list

## What it is

gutter turns a declarative tree of widgets into a flat list of quads. It does not
own a window, a graphics API or a rasteriser: the host walks the list and draws
it however it likes. That is what lets it sit on top of a Vulkan engine without
dragging a second graphics stack into it.

A runnable demo, on Vulkan, is in [`examples/vulkan`](examples/vulkan):

```sh
cd examples/vulkan && go run .
```

## Getting started

Build a tree, hand it a list and an input snapshot, draw what comes back:

```go
var list ui.DrawList

func frame(cursorX, cursorY float64, width, height int) []ui.Area {
    list.Reset() // keeps the backing array across frames

    tree := MainWindow()
    return tree.Draw(&list, ui.Input{
        CursorX: cursorX, CursorY: cursorY,
        Width:   width,   Height:  height,
    })
}

func MainWindow() ui.UIElement {
    return ui.Row{
        ...
    }
}
```

`Draw` returns the clickable `[]ui.Area` in paint order, so the last one
containing the cursor is the topmost. Dispatching clicks is the host's job — fire
`area.Function` on the press edge, not while the button is held.

## Consuming the list

```go
type Rect struct{ X, Y, W, H int }   // pixels, origin top-left, Y down

type Cmd struct {
    Rect  Rect
    Color color.NRGBA // tint; the fill colour when Tex is nil
    Tex   *Texture    // nil = solid colour
}

type Texture struct {
    Key    uint64       // stable across frames; cache the GPU handle by this
    Pixels *image.NRGBA // straight alpha, Stride == W*4, uploadable as RGBA8
    W, H   int
}
```

One `Cmd` is one alpha-blended quad. `Color` multiplies `Tex`, so a single white
text bitmap serves every colour it is drawn in. Upload a `Texture` the first time
you see its `Key` and look it up afterwards; images are uploaded once at native
size and stretched by the sampler, so nothing is resampled per widget.

Blend straight alpha (`SrcAlpha` / `OneMinusSrcAlpha`). Draw in list order — that
is back to front.

One caution for hosts that reallocate a texture when its size changes: cached text
bitmaps have their width rounded up to a multiple of 64 precisely so a string that
changes every frame keeps the same byte size. Do not defeat that by keying your
cache on anything finer.

## Widgets

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
left-aligned against the widget's left edge and centred on its vertical axis, so
padding is what positions it. It is clipped to the widget's box, during
rasterisation, rather than painted over its neighbours.

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
