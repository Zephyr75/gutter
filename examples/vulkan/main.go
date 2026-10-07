// Command vulkan is the showcase for gutter: a widget tree, a host that draws
// it, and nothing in between.
//
// Everything gutter-shaped lives in this file. The tree below is rebuilt from
// plain Go structs on every frame; calling Draw on it appends one Cmd per rect
// to a DrawList and returns the clickable areas. host.go turns that list into
// Vulkan quads and never mentions a widget.
//
//	go run .          # from this directory
package main

import (
	"fmt"
	"image/color"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Zephyr75/gutter/ui"
)

// The demo's state. There is no widget-local state in gutter yet, so this is
// where a click leaves its mark; the next frame's tree reads it back.
var (
	page   = 0
	hearts = 3
	clicks = 0
)

const maxHearts = 5

var pages = []string{"Overview", "Widgets", "Layout", "About"}

func main() {
	host := NewHost("gutter", 1100, 700)
	defer host.Close()

	// One DrawList, reused. Reset keeps the backing array, so after the first
	// few frames the list itself stops allocating.
	var drawList ui.DrawList
	quads := 0

	for !host.ShouldClose() {
		host.Poll()
		input := host.Input()

		drawList.Reset()
		// No explicit Initialize: Draw does it, and it does it *after* reading
		// the viewport out of Input. Initialising first would size the root from
		// a stale viewport, which on the very first frame is gutter's default.
		areas := MainWindow(host, quads).Draw(&drawList, input)

		// Areas come back in paint order, so the last one under the cursor is
		// the one on top. Firing on the press edge means a held button acts once.
		if host.Clicked() {
			for i := len(areas) - 1; i >= 0; i-- {
				a := areas[i]
				if a.Function != nil && ui.MouseInBounds(input, a) {
					a.Function()
					break
				}
			}
		}

		// Reported one frame late on purpose. A label that reports something
		// which changes every frame is a new string, hence a new cached bitmap
		// and a new texture, every frame -- the one thing a host with a fixed
		// pool of descriptor slots cannot survive.
		quads = len(drawList.Cmds)

		host.Render(&drawList)
	}
}

// MainWindow is the whole UI: three stacked bands in a column.
func MainWindow(host *Host, quads int) ui.UIElement {
	return ui.Column{
		Style: ui.Style{Color: bg},
		Children: []ui.UIElement{
			topBar(host),
			body(),
			statusBar(quads),
		},
	}
}

// --------------------------------------------------------------------------
// Top bar: a title, a subtitle, and a quit button
// --------------------------------------------------------------------------

func topBar(host *Host) ui.UIElement {
	return ui.Row{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 11},
			Padding: topBarPadding,
		},
		Style: ui.Style{Color: panel},
		Children: []ui.UIElement{
			label("gutter", 26, fg, ui.Size{Scale: ui.ScaleRelative, Width: 30, Height: 100}),
			label("widgets in, draw list out", 14, muted,
				ui.Size{Scale: ui.ScaleRelative, Width: 55, Height: 100}),
			ui.Button{
				Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 15, Height: 100}},
				Style:      ui.Style{Color: red},
				Function:   func() { host.Quit() },
				Child:      centeredLabel("quit", 16, bg),
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
			Padding: bodyPadding,
		},
		Style:    ui.Style{Color: bg},
		Children: []ui.UIElement{nav(), content(), gallery()},
	}
}

func nav() ui.UIElement {
	buttons := make([]ui.UIElement, 0, len(pages)+1)
	for i, name := range pages {
		i, name := i, name

		fill, text := panel, muted
		if i == page {
			fill, text = blue, bg
		}

		buttons = append(buttons, ui.Button{
			Properties: ui.Properties{
				Size:   ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 12},
				Margin: navButtonMargin,
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
			Padding: panelPadding,
		},
		Style:    ui.Style{Color: bg},
		Children: buttons,
	}
}

func content() ui.UIElement {
	children := []ui.UIElement{
		label(pages[page], 30, fg, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: titleHeight()}),
	}
	if pages[page] == "Layout" {
		children = append(children, boxModel()...)
	} else {
		children = append(children, heartsDemo()...)
	}

	return ui.Column{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 56, Height: 100},
			Padding: panelPadding,
		},
		Style:    ui.Style{Color: panel},
		Children: children,
	}
}

// heartsDemo is the content of every page but Layout: images and buttons that
// change state
func heartsDemo() []ui.UIElement {
	return []ui.UIElement{
		label(blurb(), 14, muted, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 34}),

		// Fixed-size children in a row keep their own size; the trailing
		// relative spacer takes whatever width is left over
		ui.Row{
			Properties: ui.Properties{
				Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
				Padding: demoRowPadding,
			},
			Style:    ui.Style{Color: transparent},
			Children: heartRow(),
		},

		ui.Row{
			Properties: ui.Properties{
				Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 18},
				Padding: demoRowPadding,
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
	}
}

// boxModel is the Layout page. Two lists say where this window uses a margin
// and where it uses padding, read from the same variables the tree is built
// from. Below them, three identical squares sit side by side on a dark strip:
// one with padding, one with a margin, one with neither.
func boxModel() []ui.UIElement {
	margins := strings.Join([]string{
		"Margin: space between widgets",
		"- nav buttons: " + describe(navButtonMargin),
		"- gallery buttons: " + describe(galleryButtonMargin),
		"- margin square: " + describe(demoSpacing),
		"- lone children: " + describe(fillMargin) + ", " + describe(sideMargin),
		"- their parents: " + describe(parentGap),
		// Blank lines so both lists have the same height and so the same
		// top, since text is centred vertically in its box
		"", "",
	}, "\n")
	paddings := strings.Join([]string{
		"Padding: inner space to the edge",
		"- top bar: " + describe(topBarPadding),
		"- status bar: " + describe(statusBarPadding),
		"- body: " + describe(bodyPadding),
		"- the three panels: " + describe(panelPadding),
		"- heart + button rows: " + describe(demoRowPadding),
		"- text labels: " + describe(labelPadding),
		"- padding square: " + describe(demoSpacing),
	}, "\n")

	lists := ui.Row{
		Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 32}},
		Style:      ui.Style{Color: transparent},
		Children: []ui.UIElement{
			label(margins, 14, muted, ui.Size{Scale: ui.ScaleRelative, Width: 1, Height: 100}),
			label(paddings, 14, muted, ui.Size{Scale: ui.ScaleRelative, Width: 1, Height: 100}),
		},
	}

	// The squares are fixed-size, so the strip's leftover width goes to the
	// trailing spacer and the squares keep their size whatever the window does
	squares := ui.Row{
		Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 24}},
		Style:      ui.Style{Color: bg},
		Children: []ui.UIElement{
			square("padding", blue, ui.Spacing{}, demoSpacing),
			square("margin", green, demoSpacing, ui.Spacing{}),
			square("none", red, ui.Spacing{}, ui.Spacing{}),
			// spacer(100, 100),
		},
	}

	return []ui.UIElement{
		lists,
		squares,
		label(strings.Join([]string{
			"Padding (blue) insets the child; margin (green) pushes the",
			"neighbours away. Hover: only the squares react, never the gaps.",
		}, "\n"), 14, muted, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 8}),
		loneMargins(),
		label("No siblings: the margin is taken off the parent before sizing.", 14, muted,
			ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 5}),
	}
}

// loneMargins shows a margin with no siblings to push: each dark parent holds
// one green child, which sizes and centres itself in what the margin leaves
func loneMargins() ui.UIElement {
	cases := []struct {
		caption string
		size    ui.Size
		margin  ui.Spacing
	}{
		{"fill, " + describe(fillMargin), ui.Size{}, fillMargin},
		{"50%, " + describe(sideMargin), ui.Size{Scale: ui.ScaleRelative, Width: 50, Height: 50}, sideMargin},
		{"60 px, " + describe(sideMargin), ui.Size{Scale: ui.ScalePixel, Width: 60, Height: 60}, sideMargin},
	}

	captions := make([]ui.UIElement, 0, len(cases))
	parents := make([]ui.UIElement, 0, len(cases))
	for _, c := range cases {
		cell := ui.Properties{
			Size:   ui.Size{Scale: ui.ScaleRelative, Width: 1, Height: 100},
			Margin: parentGap,
		}
		captions = append(captions, ui.Container{
			Properties: cell,
			Style:      ui.Style{Color: transparent},
			Child: ui.Text{
				StyleText: ui.StyleText{Font: asset(font), FontSize: 13, FontColor: fg},
				Content:   c.caption,
			},
		})
		parents = append(parents, ui.Container{
			Properties: cell,
			Style:      ui.Style{Color: bg},
			Child: ui.Container{
				Properties: ui.Properties{Size: c.size, Margin: c.margin},
				Style:      ui.Style{Color: green},
			},
		})
	}

	return ui.Column{
		Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 19}},
		Style:      ui.Style{Color: transparent},
		Children: []ui.UIElement{
			ui.Row{
				Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 25}},
				Style:      ui.Style{Color: transparent},
				Children:   captions,
			},
			ui.Row{
				Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 75}},
				Style:      ui.Style{Color: transparent},
				Children:   parents,
			},
		},
	}
}

// square is a fixed-size clickable box in colour c holding a darker child with
// its name, so padding shows as a band of c around the child and a margin as
// empty strip around the whole square
func square(name string, c color.RGBA, margin, padding ui.Spacing) ui.UIElement {
	return ui.Button{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 33, Height: 120},
			Margin:  margin,
			Padding: padding,
		},
		Style:    ui.Style{Color: c},
		Function: func() { clicks++ },
		Child: ui.Container{
			Style: ui.Style{Color: shade(c)},
			Child: centeredLabel(name, 14, fg),
		},
	}
}

// shade darkens c, for a child drawn inside a box of colour c
func shade(c color.RGBA) color.RGBA {
	return color.RGBA{c.R / 2, c.G / 2, c.B / 2, c.A}
}

func heartRow() []ui.UIElement {
	row := make([]ui.UIElement, 0, maxHearts+1)
	for i := 0; i < maxHearts; i++ {
		file := "heart_empty.png"
		if i < hearts {
			file = "heart_full.png"
		}
		row = append(row, ui.Container{
			Properties: ui.Properties{Size: ui.Size{Scale: ui.ScalePixel, Width: 56, Height: 56}},
			Style:      ui.Style{Color: transparent},
			Image:      asset(file),
		})
	}
	// The two heart PNGs are two textures however many hearts are on screen:
	// the same file path is the same Texture.Key, uploaded once.
	row = append(row, spacer(100, 100))
	return row
}

func gallery() ui.UIElement {
	return ui.Column{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 22, Height: 100},
			Padding: panelPadding,
		},
		Style: ui.Style{Color: bg},
		Children: []ui.UIElement{
			label("hover me", 13, muted, ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 8}),

			// A button with a dedicated hover image swaps the whole texture
			ui.Button{
				Properties: ui.Properties{
					Size:   ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
					Margin: galleryButtonMargin,
				},
				Image:      asset("white_on_black.png"),
				HoverImage: asset("black_on_white.png"),
				Function:   func() { clicks++ },
			},

			label("no hover image:\ntinted instead", 13, muted,
				ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 12}),

			// Without one, the widget emits a second translucent Cmd over the
			// first -- a tint costs a quad, not a texture
			ui.Button{
				Properties: ui.Properties{
					Size:   ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 20},
					Margin: galleryButtonMargin,
				},
				Image:    asset("button.png"),
				Function: func() { clicks++ },
			},

			spacer(100, 40),
		},
	}
}

// --------------------------------------------------------------------------
// Status bar
// --------------------------------------------------------------------------

func statusBar(quads int) ui.UIElement {
	status := fmt.Sprintf("page %d/%d   hearts %d/%d   clicks %d   quads %d",
		page+1, len(pages), hearts, maxHearts, clicks, quads)

	return ui.Row{
		Properties: ui.Properties{
			Size:    ui.Size{Scale: ui.ScaleRelative, Width: 100, Height: 7},
			Padding: statusBarPadding,
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
			Properties: ui.Properties{Padding: labelPadding},
			StyleText:  ui.StyleText{Font: asset(font), FontSize: size, FontColor: col},
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
		StyleText:  ui.StyleText{Font: asset(font), FontSize: size, FontColor: col},
		Content:    content,
	}
}

func stepper(content string, col color.Color, fn func()) ui.UIElement {
	return ui.Button{
		Properties: ui.Properties{Size: ui.Size{Scale: ui.ScalePixel, Width: 130, Height: 44}},
		Style:      ui.Style{Color: col},
		Function:   fn,
		Child:      centeredLabel(content, 15, bg),
	}
}

func spacer(width, height int) ui.UIElement {
	if height < 1 {
		height = 1
	}
	return ui.Container{
		Properties: ui.Properties{Size: ui.Size{Scale: ui.ScaleRelative, Width: width, Height: height}},
		Style:      ui.Style{Color: transparent},
	}
}

// The spacing used across the tree. They are named so the Layout page can list
// them: its text is built from these values, so it cannot drift from the UI.
var (
	topBarPadding       = ui.PaddingSymmetric(ui.ScalePixel, 10, 16)
	statusBarPadding    = ui.PaddingSymmetric(ui.ScalePixel, 0, 16)
	bodyPadding         = ui.PaddingEqual(ui.ScalePixel, 8)
	panelPadding        = ui.PaddingEqual(ui.ScalePixel, 6)
	demoRowPadding      = ui.PaddingEqual(ui.ScalePixel, 8)
	labelPadding        = ui.PaddingSymmetric(ui.ScalePixel, 0, 10)
	navButtonMargin     = ui.PaddingSymmetric(ui.ScalePixel, 4, 0)
	galleryButtonMargin = ui.PaddingEqual(ui.ScalePixel, 4)
	demoSpacing         = ui.PaddingEqual(ui.ScalePixel, 20)
	fillMargin          = ui.PaddingEqual(ui.ScalePixel, 10)
	sideMargin          = ui.PaddingSideBySide(ui.ScalePixel, 0, 0, 0, 40)
	parentGap           = ui.PaddingSymmetric(ui.ScalePixel, 0, 6)
)

// describe renders a pixel Spacing compactly: "8 px" when every side matches,
// "10 px v, 16 px h" when it is symmetric, the non-zero sides by name otherwise
func describe(p ui.Spacing) string {
	if p.Top == p.Bottom && p.Left == p.Right {
		if p.Top == p.Left {
			return fmt.Sprintf("%d px", p.Top)
		}
		parts := []string{}
		if p.Top != 0 {
			parts = append(parts, fmt.Sprintf("%d px v", p.Top))
		}
		if p.Left != 0 {
			parts = append(parts, fmt.Sprintf("%d px h", p.Left))
		}
		return strings.Join(parts, ", ")
	}
	parts := []string{}
	for _, side := range []struct {
		px   int
		name string
	}{{p.Top, "top"}, {p.Right, "right"}, {p.Bottom, "bottom"}, {p.Left, "left"}} {
		if side.px != 0 {
			parts = append(parts, fmt.Sprintf("%d px %s", side.px, side.name))
		}
	}
	return strings.Join(parts, ", ")
}

// titleHeight leaves the Layout page more room for its two lists
func titleHeight() int {
	if pages[page] == "Layout" {
		return 12
	}
	return 16
}

func blurb() string {
	return strings.Join([]string{
		"This tree is rebuilt every frame from plain Go structs.",
		"",
		"Drawing it appends one Cmd per rect to a DrawList; the",
		"host turns each into an alpha-blended quad. Nothing is",
		"rasterised per frame and no fullscreen texture is uploaded.",
		"",
		"Images go up once at native size and the sampler scales",
		"them. A string becomes one white coverage bitmap, tinted",
		"by the Cmd -- so a colour change costs no new texture.",
	}, "\n")
}

const font = "Comfortaa.ttf"

// The demo's assets live at the repository root, two directories up. Resolving
// them from this file's own path means `go run .` works from anywhere.
var repoRoot = func() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}()

func asset(name string) string { return filepath.Join(repoRoot, name) }

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
