# Gutter
A **Flutter‑style declarative UI framework** written in Go that renders through a *draw list*.

> The framework is tiny (~2 k lines of pure Go) yet fully featured: it supports absolute/relative sizing, alignment, padding, image sprites and text rendering, and produces a flat list of quads for any host (Vulkan, OpenGL, SDL, …).  A single example application shows the whole pipeline.

---

## Architecture Overview
```
UITree ──► Layout Engine ──► DrawList ──► Host Rendering Pipeline
```
* **UITree** – Your code declares widgets as plain Go structs that implement `ui.UIElement`.  No reflection, no code generation.
* **Layout Engine** (`ApplyLayout`, `applyRelative`, `applyAlignment`, `applyPadding`) resolves relative sizes and alignment into absolute coordinates.  The engine runs once per frame for the whole tree.
* **DrawList** – Each widget contributes one or more `ui.Cmd`s (rect, colour + optional texture).  A single list is passed to a host; the host turns it into GPU draw calls.

---

## Quick Start
```sh
cd examples/vulkan && go run .
```
The demo creates a window and draws a simple UI.  You’ll see:
* No per‑frame image uploads – textures are cached by file path.
* Text is rendered once per unique string + font+size combination.
* Clickable areas come back from `Draw` and the host uses them for mouse events.

---

## Public API
The package exports a small set of types and helper functions.  The table below summarizes what you need to know when building your own UI tree.
| Type | Purpose | Key Methods |
|------|---------|-------------|
| `UIElement` | Interface that every widget implements | `Draw`, `SetProperties`, `GetProperties`, `Initialize`, `Hash`, `ToString` |
| `Row`, `Column`, `Container`, `Button`, `Text` | Concrete widgets | Constructed as plain structs (see examples). |
| `Size`, `Padding`, `Style`, `StyleText`, `Point`, `ClickArea` | Layout & styling helpers | `ScaleRelative / ScalePixel` for sizes, `PaddingSymmetric` etc. |
| `DrawList` | Mutable list of draw commands | `Add(rect, colour, texture)`, `Reset()` |
| `Input` | Host‑supplied cursor position and viewport size | Passed to `Draw` and `MouseInBounds`. |

### Creating a widget
```go
ui.Button{
    Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 20, Height: 100}},
    Style:      ui.Style{Color: red},
    Function:   func(){ /* your callback */ },
    Child:      centeredLabel("quit", 16, bg),
}
```
The widget is immutable – you construct a new value each frame.  The framework takes care of the rest.

---

## Extending Gutter
1. **Custom widgets** – Implement `ui.UIElement` yourself.  Just copy one of the existing structs, add fields, and implement the six interface methods (the most work is in `Initialize`, `Draw` and `Hash`).
2. **Stateful widgets** – Currently `Gutter` has no internal state; pass state to your widget via its constructor or store it in a higher‑level component that rebuilds the tree.
3. **Additional primitives** – You can add primitives such as `Image`, `Slider`, etc., following the patterns in `Button` / `Text`.

---

## Error Handling & Diagnostics
| Function | What can fail | Recommended handling |
|----------|---------------|-----------------------|
| `imageTexture(path)` | File not found or unreadable | The function returns `(nil, err)`.  Internally we ignore the error and fall back to a coloured quad.  If you want stricter checks, use `utils.GetImageFromFilePath` directly.
| `textContext(font, size)` | Font parse failure | Returns an error; `textTexture` discards it silently.  In production code you might want to log the error or provide a fallback font.

> **Note:** The public API deliberately does not surface errors from rendering – this keeps the host logic simple.  If you need more control, call the lower‑level helpers (`imageTexture`, `textContext`) yourself.

---

## Running Tests & Building
The repository currently contains no automated tests, but the example is a full integration test of the entire stack.  Build it with:
```sh
go run ./examples/vulkan
```
You can also build the library alone:
```sh
go test ./...   # will compile all packages
```

---

## Contributing
* Pull requests are welcome – focus on small, well‑scoped changes.
* Keep the public API stable; if you need a new primitive add it under `ui/` with clear documentation.
* Add unit tests for any new logic (layout calculations, caching, etc.).

---

## Acknowledgements
* Text rendering uses the [freetype](https://github.com/goki/freetype) package.
* The example is a port of the “Gutter” demo from the original Go‑Vulkan sample set.
