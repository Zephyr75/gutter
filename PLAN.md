# Gutter — analysis and implementation plan

Goal: make gutter the UI layer for **overdrive** (Go/Vulkan 1.3 game engine, `../overdrive/`).

Written 2026-09-01 against gutter `86c99f9` and overdrive `src/core/ui.go`.

---

## 0. Where things stand

Data flow today:

```
widget(app) -> ui.UIElement tree      (rebuilt every frame, in user code)
  tree.Draw(img, window)              (layout + hit-test + rasterise, one recursive pass)
    -> *image.RGBA, full window size  (CPU pixels)
    -> []ui.Area                      (flat list of clickable rects)
  imaging.FlipV(img)                  (full-image copy, for the Y-flip)
  Backend.UpdateTexture2D(...)        (full-screen RGBA upload, every frame)
  Draw(quad) with that texture        (one fullscreen textured quad in the main pass)
```

Overdrive already links `github.com/Zephyr75/gutter v0.1.2` and has the host code
(`src/core/ui.go`), but `src/main.go:81` calls `app.Run(&scene, nil, world)` — the widget
argument is `nil`, so **the UI is currently switched off**. That is a good place to be:
the API can change without breaking a running game.

The framework is ~1400 lines and the ideas in it are sound (Flutter-style composition,
declarative tree, relative/pixel dual sizing). What follows is what has to change before
it can carry a game HUD and an editor.

---

## 1. Findings

### 1.1 Correctness — verified by running the code

**F1. `ToString()` collides. It is the change-detection key and the hover map key.**
`Size.ToString()` is `strconv.Itoa(Width) + strconv.Itoa(Height)` (`ui/ui.go:95`).
`Size{1,23}` and `Size{12,3}` both produce `"123"`. Same for `Point` (`ui/ui.go:241`),
`Area` (`ui/ui.go:253`) and `Style` (`ui/ui.go:206`, four RGBA channels concatenated).
Verified: both strings compare equal.

Two consequences: the whole-tree diff at `core/app.go:130` misses real changes, so the
UI silently fails to redraw; and `lastMap[area.ToString()]` (`core/app.go:114,136`)
merges the hover state of distinct buttons.

**F2. Integer divide-by-zero in `Row`/`Column`.**
`ui/row.go:75` and `:92` divide by `childrenWidth`, which is the sum of the *relative*
children's widths. A child declared `Size{ScaleRelative, 0, 50}` (zero width, non-zero
height, so `DefaultProperties` does not substitute a default) enters the relative branch
while contributing 0 to the sum. Verified: `runtime error: integer divide by zero`.
`ui/column.go` has the mirror bug on `childrenHeight`.

**F3. The overlay is blank on every unchanged frame, in overdrive only.**
`src/core/ui.go` allocates a fresh `img` per call, but `instance.Draw(img, window)` runs
only inside the `lastInstance != ... || !equal` guard, while `imaging.FlipV(img)` and
`UpdateTexture2D` run unconditionally below it. So a frame where nothing changed uploads
a *blank* image. Gutter's own `core/app.go:133` gets this right by keeping `flippedImg`
across iterations. This will look like "the UI flickers / only appears when I move the
mouse" the moment a widget is passed in.

**F4. `Text` has no text.** `ui/text.go:39` hardcodes `[]string{"Hello, World!", "sdgsg"}`.
The `Text` struct carries only `Properties` and `StyleText`.

**F5. Held mouse button fires the callback every frame.** `core/app.go:118` /
`src/core/ui.go:58` call `area.Function()` whenever the button *is down*, not on the
press edge. Holding for one second at 144 fps calls the handler 144 times.

**F6. Hit test has no Z order.** `areas` is a flat slice; every area containing the
cursor fires. Overlapping widgets — which is what a menu over a HUD is — all trigger.

**F7. Errors are dropped.** `ui/draw.go:85,89` discard the image-load error; a missing
file yields a nil `image.Image` handed to `resize.Resize`. `ui/text.go:63` logs and
returns, so a bad font path is a silent no-op every frame.

**F8. The README documents fields that do not exist.** `Style` has exactly one field,
`Color` (`ui/ui.go:213`). `BorderWidth`, `BorderColor`, `CornerRadius`, `ShadowWidth`
appear in the README and are commented out in `main.go`; commit `50fa912` ("Removed
borders and corners") took them out.

### 1.2 Performance

**F9. Per-frame allocation and upload proportional to window area, unconditionally.**
At 1920x1080: `image.NewRGBA` = 8.3 MB, `imaging.FlipV` = another 8.3 MB, plus an 8.3 MB
texture upload. At 144 fps that is ~3.6 GB/s of allocation and PCIe traffic for a HUD
that changes when the health bar moves. Note this happens *even with `widget == nil`*
today — overdrive pays the full cost for no UI at all.

**F10. Images are decoded from disk and Lanczos3-resampled every draw.**
`ui/draw.go:85-90`: `os.Open` + `image.Decode` + `resize.Resize(..., Lanczos3)` per image
element per redraw. This is the entirety of `TODO.md`, and it is worth roughly the 2x/1.5x
the TODO estimates.

**F11. The font file is read and parsed every draw.** `ui/text.go:63-70`:
`ioutil.ReadFile` + `freetype.ParseFont`, then a fresh `freetype.Context`, per Text
widget per frame. No glyph cache.

**F12. Change detection costs a full-tree string build.** `ToString()` concatenates the
entire tree into one string every frame to compare it against the last one — allocation
proportional to tree size, to decide whether to skip work.

**F13. Every layout step copies every widget.** `SetProperties`/`SetParent` return
`UIElement` by value (`ui/row.go:114`, etc.), so `ApplyRelative`, `ApplyAlignment`,
`ApplyPadding` each copy the struct into a fresh interface value — three heap allocations
per widget per frame, before any drawing.

### 1.3 Architecture

**F14. The widget layer imports GLFW.** Every file in `ui/` takes `*glfw.Window` in
`Draw` to read the cursor. Overdrive's first architectural invariant is that *nothing
above `renderer/` imports a graphics API*; gutter drags GLFW into the layer that should
be the most portable. It also makes `ui/` untestable without a window — the tests written
while investigating this had to pass `nil`.

**F15. Layout, hit-testing and painting are one pass.** `Draw` does `ApplyRelative` ->
`ApplyAlignment` -> `ApplyPadding` -> paint -> recurse. Because there is no measure step,
a widget can never size itself to its content (a button cannot be "as wide as its label"),
there is no clipping to the parent rect, and there can be no scroll view.

**F16. `Draw()` type-switches on a `UIType` byte tag.** `ui/draw.go:58-72` asserts
`element.(Button)`, `element.(Row)` … to fetch `Style` and `Image`. Adding a widget means
editing `draw.go`. This is a hand-rolled vtable in a language that has interfaces; `Text`
is missing from the switch, which is why a `Text` never gets a background.

**F17. Global mutable state, one UI per process.** `utils.RESOLUTION_X/Y` are package
globals written by the frame loop (`core/app.go:99`); overdrive's `src/core/ui.go` keeps
`lastInstance`, `lastMap`, `areas`, `uiTexture` at package scope. No second viewport, no
render-to-texture UI (in-world screens, minimaps), not testable in parallel.

**F18. No way to hold state between frames.** The tree is rebuilt from scratch each frame,
but there is no identity, so nothing can remember a scroll offset, a text cursor, a
checkbox, or which widget has focus. This is the hard blocker on every interactive widget
beyond a button.

**F19. Y-flip costs a full copy.** The rasteriser works Y-down and the whole image is
flipped for the GPU. Vulkan can take a negative-height viewport, or the overlay's quad
can simply flip its V coordinate — `quadVertices` in `src/core/ui.go` is five lines.

**F20. Window size, not framebuffer size.** `glfw.GetSize` is used throughout. On a HiDPI
display the UI rasterises at half resolution and is upscaled — blurry. `GetFramebufferSize`
is the correct source, with a DPI scale factor threaded through `Size`/`Padding`.

**F21. `gutter/core` drags OpenGL into overdrive's dependency graph.** `core/app.go`
imports `go-gl/gl/all-core`; overdrive imports `gutter/ui` but the module is one unit.
`ui` should have no graphics dependency at all, with the GLFW+GL host demoted to an
example.

### 1.4 Hygiene

**F22. No tests.** Layout is pure integer arithmetic on structs — the easiest thing in the
codebase to test, and where two of the three verified bugs live.

**F23. Dead code paths.** `ui/row.go:48` and `:92` branch on `Size.Scale == ScaleRelative`
after `ApplyRelative` has already rewritten every size to `ScalePixel` — unreachable.
Commented-out blocks in `button.go`, `container.go`, `draw.go`, `app.go`.

**F24. `go 1.20` in `go.mod`** vs overdrive's `1.26.3`. No `replace` directive for local
iteration, so every gutter change needs a tag+publish round trip before overdrive sees it.

---

## 2. Target architecture

Three layers, each with one job, and no graphics API above the bottom one:

```
    user code            declarative tree of Widgets (unchanged in spirit)
    ---------------------------------------------------------------------
    ui/          layout: Measure -> Arrange, produces a Draw List
                 input:  hit test against the arranged tree, from an Input snapshot
                 state:  keyed store, survives frames (hover/active/focus/scroll/text)
    ---------------------------------------------------------------------
    render/      consumes a Draw List. Two implementations:
                 - raster: into an *image.RGBA (what exists today, kept for tests + CPU fallback)
                 - gpu:    vertex + index buffers, one atlas texture, N clipped batches
    ---------------------------------------------------------------------
    host/        GLFW window + backend. In overdrive this is core/ui.go against renderer.Backend.
```

The **draw list** is the pivot. Instead of `Draw(img, window)` writing pixels, layout emits
a flat, sorted slice:

```go
type Cmd struct {
    Rect     RectF     // arranged, in physical pixels
    Clip     RectF     // parent's clip, intersected down the tree
    Color    color.RGBA
    Tex      TextureID // 0 = solid; otherwise atlas page
    UV       RectF
    Corner   float32   // rounded corners, as a shader/rasteriser parameter
    Border   float32
    BorderColor color.RGBA
}
type DrawList struct { Cmds []Cmd; Verts []Vertex; Idx []uint32 }
```

Why this and not "fix the rasteriser":

- It removes the full-screen upload. The GPU path uploads a few KB of vertices instead of
  8 MB of pixels, and it composites *inside* overdrive's main pass, so the UI can blend
  with the scene, respect depth, and be drawn to an offscreen target for an in-world screen.
- It satisfies overdrive's invariant 1 by construction: `ui/` produces data, `renderer/`
  turns it into draws. No graphics types cross the line.
- It keeps a CPU rasteriser as a *consumer* of the same list, so golden-image tests stay
  possible with no GPU — the same property that makes `go test ./...` work in overdrive today.
- Rounded corners, borders and shadows (F8, the features that were deleted) become a
  fragment-shader SDF instead of per-pixel CPU work. That is why they were expensive enough
  to remove; on the GPU they are nearly free.

Text follows the same shape: a font atlas baked once (or on demand per size), glyphs
emitted as quads into the same list. Nothing per-frame touches the disk.

---

## 3. Plan

Five phases. Phases 1-2 are worth doing whatever else is decided; they are small, they fix
real bugs, and they do not constrain the later design. Phase 3 is the fork in the road.

**Status as of 2026-09-01.** Phase 1 is done bar two partials; Phase 2 has one item
finished and one half-finished; Phases 3-5 are untouched.

| phase | state |
| --- | --- |
| 1 — correctness and cost | done, except 1.9 and 1.10, plus three items overdrive has not taken |
| 2 — decouple, then test | 2.6 done, 2.7 partial, the other six untouched |
| 3 — measure/arrange/paint | not started |
| 4 — GPU path | not started |
| 5 — widget set | not started |

The only work outstanding that touches overdrive is 1.3, 1.5 and 1.6: they are done in
gutter's own host loop, but `src/core/ui.go` keeps its own copy of that loop and still
allocates a canvas per frame, uploads unconditionally (so the overlay blanks on every
unchanged frame) and fires handlers while the button is merely held.

### Phase 1 — Correctness and cost, no API change — **done, two partials**

Small, mechanical, unblocks turning the UI on in overdrive at all.

Measured on a Row of 12 image-backed containers rasterised into a 1280x720
canvas, one frame being tree rebuild + change check + rasterise:

| | before (`86c99f9`) | after |
| --- | --- | --- |
| time per frame | 19.6 ms | 0.17 ms |
| allocated per frame | 15.5 MB | 22 KB |
| allocations per frame | 2557 | 223 |

And the redraw check on its own, over a 72-widget tree:

| | `ToString` | `TreeHash` |
| --- | --- | --- |
| time | 16.7 us | 10.7 us |
| allocations | 747 | 1 |

Items 1.3, 1.5 and 1.6 are done in gutter's own host loop (`core/app.go`);
overdrive's copy in `src/core/ui.go` still needs them, and that is the only
outstanding work in that repo.

| # | Change | Files | Status |
|---|---|---|---|
| 1.1 | Replace every `ToString()` with `Hash() uint64` over a canonical binary encoding (FNV-1a over `binary.Write` of the fields, or `maphash`). Delimit fields so `{1,23}` and `{12,3}` differ. Fixes **F1**. | `ui/ui.go`, all widgets | **done** |
| 1.2 | Guard the relative-size division: if the relative sum is 0, treat every relative child as 0 px rather than dividing. Fixes **F2**. | `ui/row.go:75,92`, `ui/column.go` | **done** |
| 1.3 | Hoist `img` and the flipped buffer out of the per-frame path in overdrive: keep one `*image.RGBA` and one texture, reuse them, and skip `FlipV` + `UpdateTexture2D` entirely when nothing changed. Fixes **F3** and most of **F9**. Also early-out when `widget == nil`. | `../overdrive/src/core/ui.go` | **done** gutter / todo overdrive |
| 1.4 | Add a `Text` field to `ui.Text` (`Content []string` or `Content string` + wrapping later). Fixes **F4**. | `ui/text.go` | **done** |
| 1.5 | Fire `Function` on the press *edge*: track the previous frame's button state and require `down && !wasDown` inside the area. Fixes **F5**. | `core/app.go`, `../overdrive/src/core/ui.go` | **done** gutter / todo overdrive |
| 1.6 | Iterate `areas` in reverse (last drawn = topmost) and stop at the first hit. Fixes **F6**, correctly enough until Phase 3 gives real Z order. | same | **done** gutter / todo overdrive |
| 1.7 | Cache decoded images and their resized variants: `map[struct{path string; w,h int}]image.Image`, with the decode cached separately from the resample. Fixes **F10** and both `TODO.md` items. | `utils/loader.go`, `ui/draw.go` | **done** |
| 1.8 | Cache parsed fonts by path and `freetype.Context` by (font, size, colour). Fixes **F11**. | `ui/text.go` | **done** |
| 1.9 | Return errors from image and font loading instead of discarding them; draw a magenta placeholder rect for a missing asset so it is visible rather than silent. Fixes **F7**. | `ui/draw.go`, `utils/loader.go` | **partial** |
| 1.10 | Add a `replace github.com/Zephyr75/gutter => ../../gutter` to overdrive's `go.mod` for local iteration; bump gutter to `go 1.26`. Fixes **F24**. | `../overdrive/src/go.mod`, `go.mod` | **partial** |

What the two partials still owe:

- **1.9** — a missing image no longer panics and the failure is cached, but it falls back
  silently to the widget's background colour. There is no visible placeholder, so a
  mistyped asset path still looks like a styling choice rather than a mistake.
- **1.10** — gutter is on `go 1.21`. The `replace` directive in overdrive's `go.mod` is
  unwritten, that repo being off limits for now.

**Acceptance:** met inside gutter — `go run .` holds 180 FPS with `redraws/s` at zero
while the mouse is still, and a clean frame allocates nothing beyond the rebuilt tree.
Not yet met in overdrive, which cannot satisfy it until it takes 1.3, 1.5 and 1.6.

### Phase 2 — Decouple, then test — **barely started**

The refactor that makes everything after it possible. This is the one not to skip: the
GLFW decoupling and the layout tests are the safety net Phase 3 gets rewritten against.

| # | Change | Files | Status |
|---|---|---|---|
| 2.1 | Define `ui.Input` — cursor position, button states (down/pressed/released), scroll delta, key and rune events, modifiers, window and framebuffer size, DPI scale. Populate it in the host from GLFW; pass it to widgets instead of `*glfw.Window`. Fixes **F14**. | new `ui/input.go`, all widgets, both hosts | todo |
| 2.2 | Delete GLFW and `go-gl/gl` from `ui/`'s imports. Move `core/` to `examples/glfw-gl/` so the library itself has no graphics dependency. Fixes **F21**. | `core/`, `ui/` | todo |
| 2.3 | Thread `Size`/`Padding`/window dimensions through a context struct instead of `utils.RESOLUTION_X/Y`; give overdrive's UI state a `type UIState struct` owned by `App` rather than package globals. Fixes **F17**. | `utils/settings.go`, `ui/ui.go`, `../overdrive/src/core/ui.go` | todo |
| 2.4 | Use `GetFramebufferSize` and carry a `Scale float32` so a 2x display rasterises at 2x. Fixes **F20**. | hosts, `ui/input.go` | todo |
| 2.5 | Give `UIElement` a `Style() Style` and `Background() (image string, hover string)` method (or fold both into a `Paint(*DrawList)` method) and delete the `UIType` type switch. Fixes **F16**; makes `Text` paintable. | `ui/ui.go`, `ui/draw.go`, all widgets | todo |
| 2.6 | Delete the unreachable `ScaleRelative` branches after `ApplyRelative`. Fixes **F23**. | `ui/row.go`, `ui/column.go` | **done** |
| 2.7 | **Table-driven layout tests.** Given a tree and a viewport, assert every widget's arranged rect. Cover: pixel/relative mixes, all nine alignments, both padding scales, nested rows in columns, zero-size children, over-subscribed children (percentages summing past 100). Fixes **F22**. | new `ui/layout_test.go` | **partial** |
| 2.8 | Hit-test tests: overlapping widgets, edge-exact coordinates, click-through on transparent parents. | new `ui/input_test.go` | todo |

Where 2.7 actually stands: `ui/layout_test.go` covers pixel/relative mixes, zero-width and
zero-height children, proportional splitting, a fixed child keeping its own size, and
`ApplyLayout` agreeing with the three steps it replaced. It exercises exactly one of the
nine alignments (`TopLeft`), and does not yet cover both padding scales, rows nested in
columns, or children whose percentages sum past 100.

Two test files exist that the plan did not ask for, both written to pin down bugs found
while doing Phase 1: `ui/text_test.go` (the cached freetype context must draw identically
the first time — full hinting made the first string at each size render nothing) and
`main_test.go` (rasterises the demo with no window and no GL context).

**Acceptance:** `go test ./...` passes with no window — already true. `ui/` imports nothing
from `go-gl` — not true; all seven files in `ui/` still import GLFW.

### Phase 3 — Measure / Arrange / Paint, and the draw list — **not started**

This is the redesign. Do it once Phase 2's tests exist, because they are the safety net.

| # | Change |
|---|---|
| 3.1 | Split `Draw` into `Measure(constraints) Size` and `Arrange(rect)`. Constraints are min/max in both axes (Flutter's `BoxConstraints` is the right model, and gutter is already Flutter-shaped). |
| 3.2 | Introduce intrinsic sizing: `ScaleContent` alongside `ScalePixel`/`ScaleRelative`. `Text` measures its string; `Container` measures its child. This is what makes a label-sized button possible. |
| 3.3 | Replace the percentage-of-parent flex with an explicit `Flex int` on children (`Expanded`/`Flexible`), plus `MainAxisAlignment`, `CrossAxisAlignment` and `Gap`. Percentages stay supported but stop being the only mechanism. |
| 3.4 | Add `Paint(dl *DrawList, clip RectF)`. Every widget appends commands; nothing writes pixels. Clip rects intersect down the tree, so a child cannot escape its parent (**F15**). |
| 3.5 | Move the raster path behind `render/raster`, consuming a `DrawList`. Keep it — it is the test oracle and the headless path. |
| 3.6 | Order: paint order *is* Z order; hit test walks the arranged tree back-to-front and returns the topmost hit. Replaces the flat `[]Area` (**F6**). |
| 3.7 | Widget identity + a `State` store keyed by a stable path (parent id + child index + optional user `Key`), holding hover/active/focus, scroll offset, text cursor, animation clocks (**F18**). |
| 3.8 | Retire whole-tree `Hash()` diffing: with a draw list, a redraw is cheap enough that the diff earns nothing. Keep a `dirty` flag driven by the state store instead (**F12**). |

**Acceptance:** a golden-image test suite — render a fixed tree to RGBA, compare against a
committed PNG. Scroll, clipping and content-sized text all demonstrable.

### Phase 4 — GPU path in overdrive — **not started**

| # | Change |
|---|---|
| 4.1 | `renderer.Backend` gains what a UI batch needs: a dynamic vertex/index buffer updated per frame, and a scissor set per batch (`SetViewportScissor` already exists inside a pass — check whether it is usable on the backbuffer, not only on an atlas target). |
| 4.2 | A `ui.slang` shader: textured, tinted quad with an SDF rounded-rect + border term. Restores **F8**'s deleted features nearly for free. Note: rounded corners and borders were prototyped on the CPU path (cached antialiased alpha masks, masking only the four corner squares) and then reverted, so nothing of that remains in the tree. |
| 4.3 | Font atlas: bake glyphs once per (font, size) into an R8 atlas; text becomes quads in the same batch. Optionally MSDF, if arbitrary scaling matters. |
| 4.4 | Image atlas for widget backgrounds, so a HUD is one or two draw calls, not one per widget. |
| 4.5 | Drop the fullscreen quad and `UpdateTexture2D` from `core/ui.go` (**F9**, **F19** — the Y-flip becomes a UV convention). |
| 4.6 | Keep the raster path selectable via `configs/*.toml`, matching overdrive's convention that every runtime knob lives in the TOML file. |

**Acceptance:** UI cost in a RenderDoc capture is a handful of draws and <100 KB of upload;
FPS with the HUD on is within noise of the HUD off.

### Phase 5 — Widgets an engine actually needs — **not started**

Once state and text work, these are ordinary:

- **Game:** label, image, progress/health bar, nine-slice panel, tooltip, modal, menu list with keyboard navigation, gamepad focus traversal.
- **Editor:** text input, checkbox, radio, slider, drag-float, dropdown, collapsible header, scroll view, tab bar, tree view (scene hierarchy), colour picker, property grid, docking split, a plot widget for the frame timings `core/app.go` already prints.
- **Cross-cutting:** theming (a `Theme` struct resolved at build time, so `Style` fields can be left zero), animation/transition driven by the state store's clock, and a debug inspector that draws the layout rects of the UI itself.

---

## 4. Sequencing advice

Phase 1 is a day and makes overdrive's UI usable. Phase 2 is the one to not skip — the
GLFW decoupling and the layout tests are what make Phase 3 safe, and Phase 3 is a rewrite
of the core loop. Phase 4 only pays off once there is enough UI on screen to be expensive;
before that the fixed Phase-1 raster path is fine.

If time is short, the highest value per hour is: **1.1, 1.3, 1.7, 1.8, 2.7**.

## 5. Risks

- **Phase 3 is a breaking API change.** `Draw(img, window)` disappears from the public
  interface. Overdrive is the only consumer and its widget tree is 60 lines in `main.go`,
  so the cost is bounded — but tag gutter `v0.2.0` and let overdrive pin the old version
  until it is ready.
- **The draw list must not leak graphics types.** `Cmd` holds a `TextureID` (an integer),
  never a `renderer.TextureHandle` or a `vk.Image`. Resolution happens in the host.
- **Font rendering is the deep end.** freetype + a hand-rolled atlas is fine for Latin at
  fixed sizes; shaping, RTL, emoji and hinting are not in scope and should be declared out
  of scope explicitly rather than discovered later.
- **Do not add a reactive/diffing layer** (React-style vdom). The tree is rebuilt per frame
  and that is correct for a game; the missing piece is the *state store* (3.7), not a
  reconciler.

## 6. Open questions for the owner

1. **Immediate or retained?** The current design is immediate-mode in spirit (rebuild every
   frame) with retained-mode diffing bolted on. Phase 3 as written commits to immediate-mode
   + a keyed state store — Dear ImGui's model with Flutter's syntax. Confirm that is the
   intent before 3.7.
2. **Is the CPU rasteriser worth keeping** past Phase 4, or is it purely a test oracle?
   Keeping it costs a second implementation of every visual feature.
3. **Editor or game HUD first?** Editor UI wants text input, docking and scroll; a HUD wants
   nine-slice, bars and animation. Phase 5 splits cleanly either way, but the order changes.
4. **DPI policy:** rasterise at physical resolution (crisp, more pixels) or at logical
   resolution and upscale (cheap, blurry)? Phase 2.4 assumes the former.
