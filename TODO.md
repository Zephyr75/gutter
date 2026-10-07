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

- [x] **Track go-vulkan's API change**
  Barriers go through `vk.DependencyInfo{Image: ...}`. go-vulkan had removed
  `GetPhysicalDeviceSurfaceFormatsKHR` as unused, so it was restored there for
  `pickFormat`. go-vulkan's `BINDINGS_OVERDRIVE_GUTTER.md` now records which of
  the two callers needs each binding.

- [x] **Docs match the code**
  README no longer points at unexported helpers or miscounts the interface.
  `OVERVIEW.md` walks through the library and the example.

## Open

- [ ] **Wire overdrive to the draw list**
  Overdrive still pins gutter `v0.1.2` and the old `ui.Area` /
  window-in-`Draw` API, so it needs a new gutter tag first. Then only
  `core/ui.go` and `ui.slang`. Unit quad `(0,0)..(1,1)` instead of a
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
  Text clips to its *bucketed* width, so a string longer than its box can spill
  up to 63 px past the box's right edge.

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
  five fields twice; fully transparent `Cmd`s still cost the host a draw call;
  `UIImage` has no widget behind it (images are a field on the other widgets).

- [ ] **`StyleText` defaults to font `"Arial"`**
  `Font` is a file path, so the default never loads and an unstyled `Text` draws
  nothing, silently. Either embed a fallback font or make a missing font loud.

- [ ] **`textTexture` measures on every frame, even on a cache hit**
  The cache key holds the *measured* width, so the measure pass — a `DrawString`
  per line, plus a `strings.Split` — has to run before the lookup can happen.
  Keying on `maxWidth`/`maxHeight` instead would skip it entirely when cached.

## Deferred

Measure/arrange, intrinsic sizing, flex, glyph and image atlases, batched vertex
streaming. None needed until something above forces them.
