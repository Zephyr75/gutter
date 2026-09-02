package ui

type Button struct {
	Properties Properties
	Style      Style
	Child      UIElement
	Function   func()
	Image      string
	HoverImage string
}

func (button Button) Initialize(in Input, skip SkipAlignment) UIElement {
	button.Properties = DefaultProperties(button.Properties, in, skip, UIButton)
	button.Style = DefaultStyle(button.Style)
	return button
}

func (button Button) Draw(dl *DrawList, in Input) []ClickArea {

	areas := []ClickArea{}

	if !button.Properties.Initialized {
		button = button.Initialize(in, SkipAlignmentNone).(Button)
	}

	button.Properties = ApplyLayout(button.Properties)

	if button.Child != nil {
		button.Child = button.Child.SetParent(&button.Properties)
		button.Child = button.Child.Initialize(in, SkipAlignmentNone)
	}

	if area, ok := Draw(dl, in, button); ok {
		areas = append(areas, area)
	}

	if button.Child != nil {
		props := button.Child.GetProperties()
		button.Child = button.Child.SetProperties(props.Size, button.Properties.Center)
		areas = append(areas, button.Child.Draw(dl, in)...)
	}

	return areas
}

func (button Button) SetProperties(size Size, center Point) UIElement {
	button.Properties.Size = size
	button.Properties.Center = center
	return button
}

func (button Button) SetParent(parent *Properties) UIElement {
	button.Properties.Parent = parent
	return button
}

func (button Button) GetProperties() Properties {
	return button.Properties
}

func (button Button) Hash(h *Hasher) {
	button.Properties.Hash(h)
	button.Style.Hash(h)
	h.String(button.Image)
	h.String(button.HoverImage)
	if button.Child != nil {
		h.Bool(true)
		button.Child.Hash(h)
		return
	}
	h.Bool(false)
}

func (button Button) ToString() string {
	result := button.Properties.ToString() +
		button.Style.ToString() +
		button.Image +
		button.HoverImage
	if button.Child != nil {
		result += button.Child.ToString()
	}
	return result
}

// TODO: split between draw and setup
// TODO: create a method to check if mouse is over button
// TODO: add a parameter to get the result
// TODO: call setup before toString and draw after
