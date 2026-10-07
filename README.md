# Gutter
A **Flutter‑style declarative UI framework** written in Go that renders through a *draw list*.

> The library is small (~1.4 k lines of pure Go, one dependency: freetype) and supports absolute/relative sizing, alignment, padding, image sprites, hover and text rendering. It writes no pixels and imports no graphics API: it produces a flat list of quads that any host (Vulkan, OpenGL, SDL, …) can draw. `examples/vulkan` shows the whole pipeline.

For a walkthrough of how it works, the main functions, and how the example uses them, read [`OVERVIEW.md`](OVERVIEW.md).

---

## Architecture Overview
```
UITree ──► Layout Engine ──► DrawList + []ClickArea ──► Host Rendering Pipeline
```
* **UITree** – Your code declares widgets as plain Go structs that implement `ui.UIElement`, and rebuilds the tree every frame. No reflection, no code generation.
* **Layout Engine** (`ApplyLayout` = relative sizing → alignment → padding, plus the child distribution in `Row`/`Column`) resolves relative sizes and alignment into absolute pixel coordinates. It runs once per frame for the whole tree, as part of `Draw`.
* **DrawList** – Each widget appends zero or more `ui.Cmd`s (rect, straight‑alpha colour, optional texture). The host turns each into one alpha-blended quad, in list order.
* **ClickArea** – `Draw` also returns the rects of every `Button`, with its callback. The host hit-tests them against the cursor.

---

## Quick Start
```sh
cd examples/vulkan && go run .
```
The example needs a Vulkan 1.3 driver and a checkout of [`Zephyr75/go-vulkan`](https://github.com/Zephyr75/go-vulkan) next to this repository. It uses the current go-vulkan API: `CmdPipelineBarrier2` takes a `vk.DependencyInfo`, and `GetPhysicalDeviceSurfaceFormatsKHR` is bound (it was removed and then restored on 2026‑10‑07, so an older checkout won't build). See [`examples/vulkan/README.md`](examples/vulkan/README.md).

The demo creates a window and draws a small UI. You'll see:
* No per‑frame image uploads – textures are cached by file path and uploaded once.
* Text is rendered once per unique string + font + size combination.
* Clickable areas come back from `Draw` and the host uses them for mouse events.

---

## Public API
The package exports a small set of types and helper functions. The table below summarizes what you need to know when building your own UI tree.
| Type / function | Purpose | Key methods / fields |
|------|---------|-------------|
| `UIElement` | Interface that every widget implements | `Draw`, `Initialize`, `SetParent`, `SetProperties`, `GetProperties`, `Hash`, `ToString` |
| `Row`, `Column`, `Container`, `Button`, `Text` | Concrete widgets | Constructed as plain structs (see examples). |
| `Properties`, `Size`, `Padding`, `Alignment`, `Point` | Layout | `ScaleRelative` / `ScalePixel` for sizes, `PaddingEqual`, `PaddingSymmetric`, `PaddingSideBySide` |
| `Style`, `StyleText` | Styling | `Color`; `Font` (a `.ttf` path), `FontSize`, `FontColor` |
| `DrawList`, `Cmd`, `Rect`, `Texture` | The output: one quad per `Cmd` | `Add(rect, colour, texture)`, `Reset()`; hosts cache textures by `Texture.Key` |
| `Input` | Host‑supplied cursor position and viewport size, in framebuffer pixels | Passed to `Draw` and `MouseInBounds`. |
| `ClickArea`, `MouseInBounds` | Hit-testing | `Function` is the button's callback |
| `TreeHash`, `Hasher` | Change detection | `TreeHash(root)` changes when any widget's geometry, style or content does |
| `ClearTextureCache` | Drops cached image and text bitmaps | For reloading assets that changed on disk |

### Creating a widget
```go
ui.Button{
    Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 20, Height: 100}},
    Style:      ui.Style{Color: red},
    Function:   func(){ /* your callback */ },
    Child:      centeredLabel("quit", 16, bg),
}
```
The widget is immutable – you construct a new value each frame. The framework takes care of the rest.

---

## Extending Gutter
1. **Custom widgets** – Implement `ui.UIElement` yourself. Copy one of the existing structs, add fields, and implement the seven interface methods (most of the work is in `Draw`). The shared `ui.Draw` helper only knows the built-in widget types, so a custom widget's `Draw` appends its own `Cmd`s to the `DrawList`.
2. **Stateful widgets** – Gutter has no widget state yet. Keep state in your own variables and rebuild the tree from it each frame. The example does this with package-level vars.
3. **Additional primitives** – Add primitives such as `Slider` by following the patterns in `Button` / `Text`. Images don't need a widget of their own: set `Image` on a `Container`, `Button`, `Row` or `Column`.

---

## Error Handling & Diagnostics
Rendering errors are not surfaced through the public API. This keeps the host logic simple.
| Case | What happens |
|----------|---------------|
| Image file missing or undecodable | The widget falls back to a quad in its `Style.Color`. The failure is cached, so it costs one failed open, not one per frame. Call `utils.GetImageFromFilePath` yourself if you want the error. |
| Font missing or unparsable | The `Text` draws nothing. `StyleText.Font` defaults to `"Arial"`, which is treated as a path, so always set a real `.ttf` path. |
| Host runs out of texture slots | Host-specific. The example logs once and draws further textures untextured. |

---

## Running Tests & Building
The repository contains no automated tests yet. The example is a full integration test of the whole stack:
```sh
cd examples/vulkan && go run .
```
The example is its own Go module, so the root module stays free of GLFW and Vulkan. To build and vet the library alone:
```sh
go build ./... && go vet ./...
```

---

## Contributing
* Pull requests are welcome – focus on small, well‑scoped changes.
* Keep the public API stable; if you need a new primitive add it under `ui/` with clear documentation.
* Add unit tests for any new logic (layout calculations, caching, etc.).

---

## Acknowledgements
* Text rendering uses the [freetype](https://github.com/goki/freetype) package.
* The example's Vulkan host is based on the `how_to_vulkan` port in [go-vulkan](https://github.com/Zephyr75/go-vulkan), with everything a 2D overlay doesn't need removed.
