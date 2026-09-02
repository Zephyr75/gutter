package ui

type Container struct {
	Properties Properties
	Style      Style
	Child      UIElement
	Image      string
}

func (container Container) Initialize(in Input, skip SkipAlignment) UIElement {
	container.Properties = DefaultProperties(container.Properties, in, skip, UIContainer)
	container.Style = DefaultStyle(container.Style)
	return container
}

func (container Container) Draw(dl *DrawList, in Input) []ClickArea {

	areas := []ClickArea{}

	if !container.Properties.Initialized {
		container = container.Initialize(in, SkipAlignmentNone).(Container)
	}

	container.Properties = ApplyLayout(container.Properties)

	if container.Child != nil {
		container.Child = container.Child.SetParent(&container.Properties)
		container.Child = container.Child.Initialize(in, SkipAlignmentNone)
	}

	if area, ok := Draw(dl, in, container); ok {
		areas = append(areas, area)
	}

	if container.Child != nil {
		props := container.Child.GetProperties()
		container.Child = container.Child.SetProperties(props.Size, container.Properties.Center)
		areas = append(areas, container.Child.Draw(dl, in)...)
	}

	return areas
}

func (container Container) SetProperties(size Size, center Point) UIElement {
	container.Properties.Size = size
	container.Properties.Center = center
	return container
}

func (container Container) SetParent(parent *Properties) UIElement {
	container.Properties.Parent = parent
	return container
}

func (container Container) GetProperties() Properties {
	return container.Properties
}

func (container Container) Hash(h *Hasher) {
	container.Properties.Hash(h)
	container.Style.Hash(h)
	h.String(container.Image)
	if container.Child != nil {
		h.Bool(true)
		container.Child.Hash(h)
		return
	}
	h.Bool(false)
}

func (container Container) ToString() string {
	result := container.Properties.ToString() +
		container.Style.ToString() +
		container.Image
	if container.Child != nil {
		result += container.Child.ToString()
	}
	return result
}
