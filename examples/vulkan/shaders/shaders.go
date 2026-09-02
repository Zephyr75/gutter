// Package shaders embeds the offline-compiled SPIR-V for the UI pipeline.
// Rebuild with `go generate ./shaders` after editing the GLSL.
package shaders

import _ "embed"

//go:generate glslc --target-env=vulkan1.3 ui.vert -o ui.vert.spv
//go:generate glslc --target-env=vulkan1.3 ui.frag -o ui.frag.spv

//go:embed ui.vert.spv
var Vert []byte

//go:embed ui.frag.spv
var Frag []byte
