module github.com/Zephyr75/gutter/examples/vulkan

go 1.26.3

require (
	github.com/Zephyr75/gutter v0.0.0
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20260628091122-0bd588dc30cf
	go-vulkan v0.0.0
)

require (
	github.com/goki/freetype v1.0.1 // indirect
	golang.org/x/image v0.6.0 // indirect
)

// The example is its own module so that gutter's own go.mod stays free of GLFW
// and Vulkan: the library emits a draw list and depends on no graphics API.
replace github.com/Zephyr75/gutter => ../..

// go-vulkan declares the module path "go-vulkan", so it can only be reached
// through a replace. Point this at your checkout of
// https://github.com/Zephyr75/go-vulkan.
replace go-vulkan => ../../../go-vulkan
