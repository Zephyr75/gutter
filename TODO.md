# TODO

## Done

- [x] **Avoid resizing on each frame**
  Images upload once at native size; the sampler scales them onto the quad.
  `GetScaledImage` and its variant cache are gone.

- [x] **Avoid reloading on each frame**
  Decoded images, parsed fonts, freetype contexts and rendered strings are all
  cached. The context cache is the one that matters: it keeps freetype's glyph
  cache alive across frames.

- [x] **Emit geometry instead of pixels**
  Widgets append to a `DrawList` instead of compositing into an `*image.RGBA`.
  One `Cmd` per rect, host turns each into an alpha-blended quad. No fullscreen
  texture upload.

- [x] **Drop the `utils.RESOLUTION_X/Y` globals**
  `Initialize` takes an `Input`, so the viewport is a parameter, not package
  state. Kills the bug where initializing before drawing read a stale size.

- [x] **A runnable example**
  `examples/vulkan` — widget tree plus a Vulkan host, its own Go module so
  gutter's `go.mod` stays free of GLFW and Vulkan.

## Open

- [ ] **Wire overdrive to the draw list**
  Only `core/ui.go` and `ui.slang`. Unit quad `(0,0)..(1,1)` instead of a
  clip-space fullscreen one; per `Cmd` look up `cache[Tex.Key]`, build a `Model`
  onto `Cmd.Rect`, set the tint, one `Draw` each. `DrawUniforms` is size-locked,
  so reuse `MatDiffuse` + `MatMetallic` for the tint rather than adding a field.
  Watch the two traps: bindless slots leak on any texture resize (hence gutter's
  64-px width bucketing), and wrong quad winding vanishes silently.

- [ ] **Stop boxing every widget in the layout walk**
  `SetProperties`, `SetParent` and `Initialize` return `UIElement` by value, so
  each costs a heap allocation per widget per frame — ~490 for a 72-widget tree.
  Largest remaining cost in gutter.

- [ ] **Widget state**
  There is none, so a click has nowhere to land but a package-level var in the
  host. Needs stable widget keys and a store the tree can read between frames.
  Would also let `ClickArea` go back to being pure geometry instead of carrying
  a `func()`.

- [ ] **Clip rects, UV rects, corner radius, border width in `Cmd`**
  Each is addable without changing `Cmd`'s shape. Only text is clipped today,
  and that happens during rasterisation; everything else can overflow its parent.

- [ ] **Replace the `UIType` type switch**
  `ui/draw.go` switches on `props.Type` and type-asserts back to the concrete
  widget to reach `Style` and `Image`. An interface method would do it without
  the assertion.

- [ ] **`Center{0,0}` is the "unset" sentinel**
  `DefaultProperties` treats a zero centre as undeclared, so a widget genuinely
  centred on the origin cannot say so.

- [ ] **Clean up**
  Deprecated `ApplyPadding`/`ApplyAlignment`/`ApplyRelative` wrappers; `ToString`
  methods superseded by `Hash`; `textKey` and the `Hasher` sum encode the same
  five fields twice; fully transparent `Cmd`s still cost the host a draw call.

## Deferred

Measure/arrange, intrinsic sizing, flex, glyph and image atlases, batched vertex
streaming. None needed until something above forces them.
