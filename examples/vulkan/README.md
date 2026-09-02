# gutter on Vulkan

A runnable showcase: a widget tree, and a host that draws it.

```sh
go run .
```

`main.go` is the gutter half — a UI built from plain Go structs, rebuilt every
frame. `host.go` is the Vulkan half. They meet at one type:

```go
var dl ui.DrawList

dl.Reset()
areas := MainWindow(host, quads).Draw(&dl, host.Input())  // widgets -> geometry
host.Render(&dl)                                          // geometry -> quads
```

`Draw` appends one `ui.Cmd` per rect to the list and returns the clickable
`[]ui.Area`. gutter writes no pixels and imports no graphics API; the host never
mentions a widget.

## What the host does with a Cmd

Each `Cmd` becomes one alpha-blended quad, in list order:

- **No vertex buffer.** The six corners come from `gl_VertexIndex`; a 44-byte
  push constant carries the rect, the tint and a texture slot. A `Cmd` is a
  `vkCmdPushConstants` and a `vkCmdDraw`, nothing else.
- **No depth attachment.** The draw list is already back-to-front — that is what
  a painter's list is — so `CullModeNone` and no depth state at all.
- **No Y flip.** gutter's origin is top-left with Y down, which is also Vulkan's
  NDC orientation, so the vertex shader is a scale and a bias.
- **Bindless textures.** Each distinct `Texture.Key` is uploaded once and keeps a
  slot in a partially-bound `sampler2D[]`. gutter's caches return the same
  `*Texture` every frame, so steady state uploads nothing.
- **UNORM swapchain, not sRGB.** gutter's colours and its PNGs are already
  sRGB-encoded bytes; UNORM presents them unchanged, so the window matches what
  a CPU replay of the same draw list produces.

## The trap this demo walks into on purpose

The status bar reports the quad count *one frame late*. A label that reports
something which changes every frame is a new string, so a new cached bitmap, so
a new texture — every frame. A host with a fixed pool of descriptor slots dies
in seconds. gutter helps by bucketing text bitmap widths to a multiple of 64, but
it cannot help if the *content* changes. Report last frame's number, or don't
report it.

## Requirements

- A Vulkan 1.3 driver with `VK_KHR_swapchain` and descriptor indexing.
- [`Zephyr75/go-vulkan`](https://github.com/Zephyr75/go-vulkan) checked out
  next to this repository. It declares the module path `go-vulkan`, so `go.mod`
  reaches it through a `replace`; adjust that line if yours lives elsewhere.
- `glslc` only if you edit `shaders/*.vert|frag` — the `.spv` is committed and
  embedded. Rebuild with `go generate ./shaders`.

This example is its own Go module so that gutter's `go.mod` stays free of GLFW
and Vulkan.

`go vet` reports "possible misuse of unsafe.Pointer" twice, at the GLFW surface
call. Those are opaque Vulkan handles round-tripping between two bindings, not
Go pointers; the warning does not apply.
