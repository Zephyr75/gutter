package ui

// Input is what the host reads from its window and hands to the widget tree:
// - cursor position such that the host can decide whether a button is hovered
// - current screen size to scale widget tree elements
type Input struct {
	CursorX, CursorY float64
	Width, Height    int
}

// Fallback viewport, used only when a host hands Draw an Input with no size --
// a blank window is a worse answer than a wrong-sized one.
const defaultViewportW, defaultViewportH = 1000, 500

// viewport is the size a root widget fills. Input is the single source of truth
// for it; nothing in this package keeps a mutable copy.
func (in Input) viewport() (int, int) {
	if in.Width > 0 && in.Height > 0 {
		return in.Width, in.Height
	}
	return defaultViewportW, defaultViewportH
}

func MouseInBounds(in Input, area ClickArea) bool {
	return in.CursorX > area.Left && in.CursorX < area.Right &&
		in.CursorY > area.Top && in.CursorY < area.Bottom
}
