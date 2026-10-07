package ui

type Row struct {
	Properties Properties
	Style      Style
	Children   []UIElement
	Image      string
}

func (row Row) Initialize(input Input, skip SkipAlignment) UIElement {
	row.Properties = DefaultProperties(row.Properties, input, skip, UIRow)
	row.Style = DefaultStyle(row.Style)
	return row
}

func (row Row) Draw(drawList *DrawList, input Input) []ClickArea {

	areas := []ClickArea{}

	if !row.Properties.Initialized {
		row = row.Initialize(input, SkipAlignmentNone).(Row)
	}

	row.Properties = ApplyLayout(row.Properties)
	content := ContentBox(row.Properties)

	for i, child := range row.Children {
		child = child.SetParent(&content)
		row.Children[i] = child.Initialize(input, SkipAlignmentHoriz)
	}

	if area, ok := Draw(drawList, input, row); ok {
		areas = append(areas, area)
	}

	// Margins are empty space beside a child, so every child's left and right
	// margins come out first, then fixed-size children claim their width
	availableWidth := content.Size.Width
	maxWidth := content.Size.Width
	for _, child := range row.Children {
		childProps := child.GetProperties()
		_, right, _, left := pixels(childProps.Margin, content.Size.Width, content.Size.Height)
		availableWidth -= left + right
		if childProps.Size.Scale == ScalePixel {
			availableWidth -= childProps.Size.Width
		}
	}

	// What is left is shared among the relative children, in proportion to the
	// percentages they declared
	childrenWidth := 0
	for _, child := range row.Children {
		childProps := child.GetProperties()
		if childProps.Size.Scale == ScaleRelative {
			childrenWidth += childProps.Size.Width
		}
	}

	// Size and place every child in one pass, stepping over its margins
	currentX := content.Center.X - maxWidth/2
	for i, child := range row.Children {
		childProps := child.GetProperties()
		top, right, bottom, left := pixels(childProps.Margin, content.Size.Width, content.Size.Height)

		pixelWidth := childProps.Size.Width
		pixelHeight := childProps.Size.Height
		if childProps.Size.Scale == ScaleRelative {
			// childrenWidth is zero when every relative child declared a width of
			// zero, which used to divide by zero and take the process with it
			pixelWidth = 0
			if childrenWidth > 0 {
				pixelWidth = childProps.Size.Width * availableWidth / childrenWidth
			}
			// A relative child fills the row's height, less its own margins
			pixelHeight = content.Size.Height - top - bottom
		}

		currentX += left
		row.Children[i] = child.SetProperties(
			Size{
				Scale:  ScalePixel,
				Width:  pixelWidth,
				Height: pixelHeight,
			},
			Point{
				X: currentX + pixelWidth/2,
				Y: content.Center.Y,
			},
		)
		currentX += pixelWidth + right
	}

	for _, child := range row.Children {
		areas = append(areas, child.Draw(drawList, input)...)
	}

	return areas

}

func (row Row) SetProperties(size Size, center Point) UIElement {
	row.Properties.Size = size
	row.Properties.Center = center
	return row
}

func (row Row) SetParent(parent *Properties) UIElement {
	row.Properties.Parent = parent
	return row
}

func (row Row) GetProperties() Properties {
	return row.Properties
}

func (row Row) Hash(h *Hasher) {
	row.Properties.Hash(h)
	row.Style.Hash(h)
	h.String(row.Image)
	h.Int(len(row.Children))
	for _, child := range row.Children {
		child.Hash(h)
	}
}

func (row Row) ToString() string {
	result := row.Properties.ToString() + row.Style.ToString()
	for _, child := range row.Children {
		result += child.ToString()
	}
	return result
}
