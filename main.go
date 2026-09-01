package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/Zephyr75/gutter/core"
	"github.com/Zephyr75/gutter/ui"
)

// The demo's state. The widget tree is rebuilt from these every frame; the host
// loop hashes the result and only repaints when the hash moves, so holding the
// window still costs no rasterisation at all.
var (
	page   = 0
	hearts = 3
	clicks = 0
)

const maxHearts = 5

var pages = []string{"Overview", "Widgets", "Layout", "About"}

func main() {
	app := core.App{
		Name:   "Gutter",
		Width:  1100,
		Height: 700,
	}
	app.Run(MainWindow)
}

func MainWindow(app core.App) ui.UIElement {
	return ui.Column{
		Style: ui.Style{Color: bg},
		Children: []ui.UIElement{
			topBar(app),
			body(),
			statusBar(),
		},
	}
}

// --------------------------------------------------------------------------
// Top bar: a title, a stretch of nothing, and a quit button
// --------------------------------------------------------------------------

func topBar(app core.App) ui.UIElement {
	return ui.Row{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 11},
			Padding: ui.PaddingSymmetric(ui.ScalePixel, 10, 16),
		},
		Style: ui.Style{Color: panel},
		Children: []ui.UIElement{
			label("gutter", 26, fg, ui.Size{Scale: ui.ScaleRelative, Width: 30, Height: 100}),
			label("a Flutter-shaped UI framework, in Go", 14, muted,
				ui.Size{Scale: ui.ScaleRelative, Width: 55, Height: 100}),
			ui.Button{
				Properties: ui.Properties{
					Size: ui.Size{Scale: ui.ScaleRelative, Width: 15, Height: 100},
				},
				Style:    ui.Style{Color: red},
				Function: func() { app.Quit() },
				Child:    centeredLabel("quit", 16, bg),
			},
		},
	}
}

// --------------------------------------------------------------------------
// Body: navigation on the left, content in the middle, images on the right
// --------------------------------------------------------------------------

func body() ui.UIElement {
	return ui.Row{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 82},
			Padding: ui.PaddingEqual(ui.ScalePixel, 8),
		},
		Style: ui.Style{Color: bg},
		Children: []ui.UIElement{
			nav(),
			content(),
			gallery(),
		},
	}
}

func nav() ui.UIElement {
	buttons := make([]ui.UIElement, 0, len(pages)+1)
	for i, name := range pages {
		i, name := i, name

		fill := panel
		text := muted
		if i == page {
			fill = blue
			text = bg
		}

		buttons = append(buttons, ui.Button{
			Properties: ui.Properties{
				Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 12},
				Padding: ui.PaddingSymmetric(ui.ScalePixel, 4, 0),
			},
			Style:    ui.Style{Color: fill},
			Function: func() { page = i },
			Child:    label(name, 16, text, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 100}),
		})
	}

	// Soak up the leftover height so the buttons stay at the top
	buttons = append(buttons, spacer(100, 100-12*len(pages)))

	return ui.Column{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 22, Height: 100},
			Padding: ui.PaddingEqual(ui.ScalePixel, 6),
		},
		Style:    ui.Style{Color: bg},
		Children: buttons,
	}
}

func content() ui.UIElement {
	return ui.Column{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 56, Height: 100},
			Padding: ui.PaddingEqual(ui.ScalePixel, 6),
		},
		Style: ui.Style{Color: panel},
		Children: []ui.UIElement{
			label(pages[page], 30, fg, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 16}),
			label(blurb(), 14, muted, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 34}),

			// Fixed-size children in a row keep their own size; the trailing
			// relative spacer takes whatever width is left over
			ui.Row{
				Properties: ui.Properties{
					Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
					Padding: ui.PaddingEqual(ui.ScalePixel, 8),
				},
				Style:    ui.Style{Color: transparent},
				Children: heartRow(),
			},

			ui.Row{
				Properties: ui.Properties{
					Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 18},
					Padding: ui.PaddingEqual(ui.ScalePixel, 8),
				},
				Style: ui.Style{Color: transparent},
				Children: []ui.UIElement{
					stepper("- damage", red, func() {
						if hearts > 0 {
							hearts--
						}
					}),
					spacer(6, 100),
					stepper("+ heal", green, func() {
						if hearts < maxHearts {
							hearts++
						}
					}),
					spacer(48, 100),
				},
			},

			spacer(100, 12),
		},
	}
}

func heartRow() []ui.UIElement {
	row := make([]ui.UIElement, 0, maxHearts+1)
	for i := 0; i < maxHearts; i++ {
		file := "heart_empty.png"
		if i < hearts {
			file = "heart_full.png"
		}
		row = append(row, ui.Container{
			Properties: ui.Properties{
				Size: ui.Size{Scale: ui.ScalePixel, Width: 56, Height: 56},
			},
			Style: ui.Style{Color: transparent},
			Image: file,
		})
	}
	// Without a relative child the fixed hearts would simply bunch at the left,
	// which is what we want here -- the spacer just makes that explicit
	row = append(row, spacer(100, 100))
	return row
}

func gallery() ui.UIElement {
	return ui.Column{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 22, Height: 100},
			Padding: ui.PaddingEqual(ui.ScalePixel, 6),
		},
		Style: ui.Style{Color: bg},
		Children: []ui.UIElement{
			label("hover me", 13, muted, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 8}),

			// A button with a dedicated hover image swaps the whole texture;
			// both variants are decoded and resampled once, then cached
			ui.Button{
				Properties: ui.Properties{
					Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
					Padding: ui.PaddingEqual(ui.ScalePixel, 4),
				},
				Image:      "white_on_black.png",
				HoverImage: "black_on_white.png",
				Function:   func() { clicks++ },
			},

			label("no hover image:\ntinted instead", 13, muted,
				ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 12}),

			ui.Button{
				Properties: ui.Properties{
					Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
					Padding: ui.PaddingEqual(ui.ScalePixel, 4),
				},
				Image:    "button.png",
				Function: func() { clicks++ },
			},

			spacer(100, 40),
		},
	}
}

// --------------------------------------------------------------------------
// Status bar
// --------------------------------------------------------------------------

func statusBar() ui.UIElement {
	status := fmt.Sprintf("page %d/%d   hearts %d/%d   clicks %d",
		page+1, len(pages), hearts, maxHearts, clicks)

	return ui.Row{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 7},
			Padding: ui.PaddingSymmetric(ui.ScalePixel, 0, 16),
		},
		Style: ui.Style{Color: panel},
		Children: []ui.UIElement{
			label(status, 13, muted, ui.Size{Scale: ui.ScaleRelative, Width: 60, Height: 100}),
			label("click a nav entry, or the hearts buttons", 13, muted,
				ui.Size{Scale: ui.ScaleRelative, Width: 40, Height: 100}),
		},
	}
}

// --------------------------------------------------------------------------
// Small helpers, so the tree above reads as layout rather than as boilerplate
// --------------------------------------------------------------------------

func label(content string, size int, col color.Color, box ui.Size) ui.UIElement {
	return ui.Container{
		Properties: ui.Properties{Size: box},
		Style:      ui.Style{Color: transparent},
		Child: ui.Text{
			Properties: ui.Properties{Padding: ui.PaddingSymmetric(ui.ScalePixel, 0, 10)},
			StyleText:  ui.StyleText{Font: font, FontSize: size, FontColor: col},
			Content:    content,
		},
	}
}

func centeredLabel(content string, size int, col color.Color) ui.UIElement {
	// Text is left-aligned inside its own box, so padding is what centres it
	pad := (10 - len(content)) * 4
	if pad < 4 {
		pad = 4
	}
	return ui.Text{
		Properties: ui.Properties{Padding: ui.PaddingSymmetric(ui.ScalePixel, 0, pad)},
		StyleText:  ui.StyleText{Font: font, FontSize: size, FontColor: col},
		Content:    content,
	}
}

func stepper(content string, col color.Color, fn func()) ui.UIElement {
	return ui.Button{
		Properties: ui.Properties{
			Size: ui.Size{Scale: ui.ScalePixel, Width: 130, Height: 44},
		},
		Style:    ui.Style{Color: col},
		Function: fn,
		Child:    centeredLabel(content, 15, bg),
	}
}

func spacer(width, height int) ui.UIElement {
	if height < 1 {
		height = 1
	}
	return ui.Container{
		Properties: ui.Properties{
			Size: ui.Size{Scale: ui.ScaleRelative, Width: width, Height: height},
		},
		Style: ui.Style{Color: transparent},
	}
}

func blurb() string {
	return strings.Join([]string{
		"This tree is rebuilt every frame from plain Go structs.",
		"",
		"The host loop hashes it and repaints only when the hash",
		"changes, so an idle window rasterises nothing. Images and",
		"fonts are decoded and resampled once, then cached by size.",
		"",
		"Watch redraws/s in the terminal: it drops to zero when the",
		"mouse is still, and ticks up only on hover or a click.",
	}, "\n")
}

const font = "Comfortaa.ttf"

var (
	bg          = color.RGBA{26, 27, 38, 255}
	panel       = color.RGBA{36, 40, 59, 255}
	fg          = color.RGBA{192, 202, 245, 255}
	muted       = color.RGBA{130, 139, 184, 255}
	blue        = color.RGBA{122, 162, 247, 255}
	green       = color.RGBA{158, 206, 106, 255}
	red         = color.RGBA{247, 118, 142, 255}
	transparent = color.RGBA{0, 0, 0, 0}
)
