package core

import (
	"fmt"
	"image"
	"runtime"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/Zephyr75/gutter/ui"
	"github.com/Zephyr75/gutter/utils"
)

type App struct {
	Name   string
	Width  int
	Height int
	Window *glfw.Window
}

func init() {
	// GLFW: This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func (app App) Quit() {
	app.Window.SetShouldClose(true)
}

func (app App) Run(widget func(app App) ui.UIElement) {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	utils.RESOLUTION_X = app.Width
	utils.RESOLUTION_Y = app.Height

	window, err := glfw.CreateWindow(utils.RESOLUTION_X, utils.RESOLUTION_Y, app.Name, nil, nil)
	if err != nil {
		panic(err)
	}

	app.Window = window

	window.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	var texture uint32
	{
		gl.GenTextures(1, &texture)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	}

	var framebuffer uint32
	{
		gl.GenFramebuffers(1, &framebuffer)
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	var (
		// The canvas the widgets rasterise into, and the same pixels with their
		// rows reversed for OpenGL. Both are allocated once per window size
		// rather than once per frame
		canvas  *image.RGBA
		flipped []byte
		texW    int
		texH    int

		areas     []ui.Area
		lastTree  uint64
		lastHover uint64
		lastDown  bool
		resized   = true
	)

	frames, redraws := 0, 0
	lastReport := glfw.GetTime()

	for !window.ShouldClose() {

		w, h := window.GetSize()
		if w == 0 || h == 0 {
			// Minimised: nothing to rasterise and nothing to blit
			glfw.PollEvents()
			continue
		}

		utils.RESOLUTION_X = w
		utils.RESOLUTION_Y = h

		if w != texW || h != texH {
			canvas = image.NewRGBA(image.Rect(0, 0, w, h))
			flipped = make([]byte, w*h*4)
			texW, texH = w, h

			// Reallocate the texture's storage to match, once, on resize
			gl.BindTexture(gl.TEXTURE_2D, texture)
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
			resized = true
		}

		cursorX, cursorY := window.GetCursorPos()

		// Dispatch input against the areas that were actually drawn last frame,
		// before rebuilding the tree, so a handler's effect shows up this frame.
		// The areas are in paint order, so the last one containing the cursor is
		// the topmost, and it is the only one that gets the click
		down := window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press
		if down && !lastDown {
			for i := len(areas) - 1; i >= 0; i-- {
				if inBounds(cursorX, cursorY, areas[i]) {
					if areas[i].Function != nil {
						areas[i].Function()
					}
					break
				}
			}
		}
		lastDown = down

		instance := widget(app)

		// A redraw is only worth it when the tree changed, when the set of
		// hovered widgets changed, or when the window was resized
		redraw := resized
		if instance != nil {
			tree := ui.TreeHash(instance)
			hover := hoverSignature(cursorX, cursorY, areas)
			redraw = redraw || tree != lastTree || hover != lastHover
			lastTree = tree
		}

		if redraw && instance != nil {
			clear(canvas.Pix)
			areas = dropEmpty(instance.Draw(canvas, window))

			flipVertical(canvas, flipped)
			gl.BindTexture(gl.TEXTURE_2D, texture)
			gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(flipped))

			lastHover = hoverSignature(cursorX, cursorY, areas)
			resized = false
			redraws++
		}

		// The blit happens every frame regardless: the default framebuffer is
		// swapped out from under us, so what it holds is not ours to keep
		gl.BlitFramebuffer(0, 0, int32(w), int32(h), 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.LINEAR)

		window.SwapBuffers()
		glfw.PollEvents()

		frames++
		if now := glfw.GetTime(); now-lastReport > 1 {
			fmt.Printf("\rFPS: %-6d redraws/s: %-6d ", frames, redraws)
			frames, redraws = 0, 0
			lastReport = now
		}
	}
}

func inBounds(x, y float64, area ui.Area) bool {
	return x > area.Left && x < area.Right && y > area.Top && y < area.Bottom
}

// hoverSignature identifies which widgets the cursor is over. It is compared
// between frames to decide whether the hover styling needs repainting, and it
// reads the cursor once rather than once per area.
func hoverSignature(x, y float64, areas []ui.Area) uint64 {
	h := ui.NewHasher()
	for i := range areas {
		h.Bool(inBounds(x, y, areas[i]))
	}
	return h.Sum()
}

// dropEmpty compacts the areas of widgets that report no clickable box, in
// place, so the hit test walks a short list.
func dropEmpty(areas []ui.Area) []ui.Area {
	n := 0
	for _, area := range areas {
		if area.Right > area.Left && area.Bottom > area.Top {
			areas[n] = area
			n++
		}
	}
	return areas[:n]
}

// flipVertical copies src into dst with its rows reversed, which is the order
// OpenGL wants. Writing into a buffer the caller owns is what keeps this off
// the allocator: imaging.FlipV returned a new full-size image every frame.
func flipVertical(src *image.RGBA, dst []byte) {
	height := src.Rect.Dy()
	rowLen := src.Rect.Dx() * 4
	for y := 0; y < height; y++ {
		start := (height - 1 - y) * src.Stride
		copy(dst[y*rowLen:(y+1)*rowLen], src.Pix[start:start+rowLen])
	}
}
