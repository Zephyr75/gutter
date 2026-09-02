package main

// The Vulkan half of the demo. gutter never appears below the Render method:
// the library hands over a []ui.Cmd and this file turns each one into a quad.
//
// Structure is lifted from the how-to-vulkan reference renderer, minus
// everything a 2D overlay does not need: no vertex or index buffer (the quad is
// generated in the shader), no depth attachment (painter's order is the depth
// test), no per-frame uniform buffer (a 44-byte push constant carries the whole
// draw). What is left is instance/device/swapchain, a bindless texture array,
// and one pipeline.

import (
	"image"
	"log"
	"math"
	"runtime"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"

	"go-vulkan/vk"

	"github.com/Zephyr75/gutter/ui"

	"github.com/Zephyr75/gutter/examples/vulkan/shaders"
)

// framesInFlight is how many frames the CPU may record ahead of the GPU.
const framesInFlight = 2

// maxTextures sizes the bindless array. Every distinct Texture.Key gutter
// produces takes one slot: one per image file, one per (string, font, size,
// bucketed width) that has been drawn. gutter buckets text bitmap widths to a
// multiple of 64 precisely so a label that changes every frame does not mint a
// new key every frame.
const maxTextures = 512

// push mirrors the shader's push_constant block. All fields are 4-byte aligned
// in both languages, so Go's layout and std430's agree field for field; the
// total is 44 bytes, well inside the 128 every implementation guarantees.
type push struct {
	Color  [4]float32
	Screen [2]float32
	Pos    [2]float32
	Size   [2]float32
	Tex    int32
}

// hostTexture is one uploaded gutter Texture and the descriptor slot it lives in.
type hostTexture struct {
	image vk.Image
	alloc vk.VmaAllocation
	view  vk.ImageView
	slot  int32
}

type Host struct {
	window *glfw.Window

	instance       vk.Instance
	physicalDevice vk.PhysicalDevice
	device         vk.Device
	queue          vk.Queue
	queueFamily    uint32
	allocator      *vk.VmaAllocator
	surface        vk.SurfaceKHR

	swapchainCI vk.SwapchainCreateInfo
	swapchain   vk.SwapchainKHR
	imageFormat vk.Format
	images      []vk.Image
	imageViews  []vk.ImageView
	extent      vk.Extent2D
	needsResize bool

	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer
	fences         [framesInFlight]vk.Fence
	acquired       [framesInFlight]vk.Semaphore
	rendered       []vk.Semaphore // one per swapchain image
	frame          int

	vertModule vk.ShaderModule
	fragModule vk.ShaderModule
	setLayout  vk.DescriptorSetLayout
	pool       vk.DescriptorPool
	set        vk.DescriptorSet
	layout     vk.PipelineLayout
	pipeline   vk.Pipeline
	sampler    vk.Sampler

	textures  map[uint64]*hostTexture
	nextSlot  int32
	slotsFull bool

	// Input state. Cursor is reported in window coordinates and the draw list is
	// in framebuffer pixels, so on a scaled display the two differ.
	cursorX, cursorY float64
	mouseDown        bool
	clicked          bool
}

func init() {
	// Vulkan submission and GLFW both want a stable thread.
	runtime.LockOSThread()
}

func chk(err error) {
	if err != nil {
		log.Fatalf("vulkan: %v", err)
	}
}

// NewHost opens a window and brings up everything needed to draw a gutter
// DrawList into it.
func NewHost(title string, width, height int) *Host {
	h := &Host{textures: map[uint64]*hostTexture{}}

	chk(glfw.Init())
	if !glfw.VulkanSupported() {
		log.Fatal("glfw: no Vulkan support")
	}
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	chk(err)
	h.window = window
	window.SetFramebufferSizeCallback(func(*glfw.Window, int, int) { h.needsResize = true })

	h.instance, err = vk.CreateInstance(vk.InstanceCreateInfo{
		AppName:    title,
		APIVersion: vk.ApiVersion13,
		Extensions: window.GetRequiredInstanceExtensions(),
	})
	chk(err)

	devices, err := vk.EnumeratePhysicalDevices(h.instance)
	chk(err)
	if len(devices) == 0 {
		log.Fatal("vulkan: no physical devices")
	}
	h.physicalDevice = devices[0]
	log.Printf("device: %s", vk.GetPhysicalDeviceProperties2(h.physicalDevice).DeviceName)

	for i, qf := range vk.GetPhysicalDeviceQueueFamilyProperties(h.physicalDevice) {
		if qf.QueueFlags&vk.QueueGraphics != 0 {
			h.queueFamily = uint32(i)
			break
		}
	}

	// Descriptor indexing is the only interesting feature here: the fragment
	// shader indexes a runtime-sized texture array by a push constant, and most
	// of its slots are empty at any given moment, hence PartiallyBound.
	h.device, err = vk.CreateDevice(h.physicalDevice, vk.DeviceCreateInfo{
		QueueCreateInfos: []vk.DeviceQueueCreateInfo{{QueueFamilyIndex: h.queueFamily, Priorities: []float32{1}}},
		Extensions:       []string{"VK_KHR_swapchain"},
		Features: vk.Features{
			DescriptorIndexing:                        true,
			ShaderSampledImageArrayNonUniformIndexing: true,
			DescriptorBindingVariableDescriptorCount:  true,
			DescriptorBindingPartiallyBound:           true,
			RuntimeDescriptorArray:                    true,
			Synchronization2:                          true,
			DynamicRendering:                          true,
		},
	})
	chk(err)
	h.queue = vk.GetDeviceQueue(h.device, h.queueFamily, 0)

	h.allocator = vk.VmaCreateAllocator(vk.VmaAllocatorCreateInfo{
		PhysicalDevice: h.physicalDevice,
		Device:         h.device,
		Instance:       h.instance,
	})

	// GLFW takes the instance as a pointer-kind value and returns a pointer to
	// the created surface handle, so both ends round-trip through
	// unsafe.Pointer. `go vet` reports "possible misuse of unsafe.Pointer" here:
	// these are opaque Vulkan handles, not Go pointers, so the warning does not
	// apply -- but it is the price of this binding pair.
	surfRaw, err := window.CreateWindowSurface((*byte)(unsafe.Pointer(h.instance)), nil)
	chk(err)
	h.surface = vk.SurfaceKHR(*(*uintptr)(unsafe.Pointer(surfRaw)))
	presentOK, err := vk.GetPhysicalDeviceSurfaceSupportKHR(h.physicalDevice, h.queueFamily, h.surface)
	chk(err)
	if !presentOK {
		log.Fatal("vulkan: graphics queue cannot present")
	}

	h.imageFormat = pickFormat(h.physicalDevice, h.surface)
	h.createSwapchain()

	h.commandPool, err = vk.CreateCommandPool(h.device, h.queueFamily, vk.CommandPoolCreateResetCommandBuffer)
	chk(err)
	h.commandBuffers, err = vk.AllocateCommandBuffers(h.device, h.commandPool, framesInFlight)
	chk(err)
	for i := 0; i < framesInFlight; i++ {
		h.fences[i], err = vk.CreateFence(h.device, vk.FenceCreateSignaled)
		chk(err)
		h.acquired[i], err = vk.CreateSemaphore(h.device)
		chk(err)
	}

	h.createDescriptors()
	h.createPipeline()

	h.sampler, err = vk.CreateSampler(h.device, vk.SamplerCreateInfo{
		MagFilter: vk.FilterLinear, MinFilter: vk.FilterLinear,
		MipmapMode:   vk.SamplerMipmapModeLinear,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
		MaxLod:       1,
	})
	chk(err)

	return h
}

// pickFormat prefers a UNORM swapchain. An _SRGB one would have the hardware
// encode on write, but gutter's colours and its PNGs are already sRGB-encoded
// bytes; UNORM presents them unchanged, so the window matches what the library's
// own CPU replay produces.
func pickFormat(pd vk.PhysicalDevice, s vk.SurfaceKHR) vk.Format {
	formats, err := vk.GetPhysicalDeviceSurfaceFormatsKHR(pd, s)
	chk(err)
	for _, want := range []vk.Format{vk.FormatB8G8R8A8Unorm, vk.FormatR8G8B8A8Unorm} {
		for _, f := range formats {
			if f.Format == want && f.ColorSpace == vk.ColorSpaceSrgbNonlinearKHR {
				return want
			}
		}
	}
	return formats[0].Format
}

func (h *Host) createSwapchain() {
	caps, err := vk.GetPhysicalDeviceSurfaceCapabilitiesKHR(h.physicalDevice, h.surface)
	chk(err)

	w, ht := h.window.GetFramebufferSize()
	h.extent = vk.Extent2D{Width: uint32(w), Height: uint32(ht)}
	if caps.CurrentExtent.Width != 0xFFFFFFFF {
		h.extent = caps.CurrentExtent
	}

	h.swapchainCI = vk.SwapchainCreateInfo{
		Surface:         h.surface,
		MinImageCount:   caps.MinImageCount,
		ImageFormat:     h.imageFormat,
		ImageColorSpace: vk.ColorSpaceSrgbNonlinearKHR,
		ImageExtent:     h.extent,
		ImageUsage:      vk.ImageUsageColorAttachment,
		PreTransform:    vk.SurfaceTransformIdentityKHR,
		CompositeAlpha:  vk.CompositeAlphaOpaqueKHR,
		PresentMode:     vk.PresentModeFifoKHR,
		OldSwapchain:    h.swapchain,
	}
	newSwapchain, err := vk.CreateSwapchainKHR(h.device, h.swapchainCI)
	chk(err)
	if h.swapchain != 0 {
		vk.DestroySwapchainKHR(h.device, h.swapchain)
	}
	h.swapchain = newSwapchain

	for _, v := range h.imageViews {
		vk.DestroyImageView(h.device, v)
	}
	h.images, err = vk.GetSwapchainImagesKHR(h.device, h.swapchain)
	chk(err)
	h.imageViews = make([]vk.ImageView, len(h.images))
	for i := range h.images {
		h.imageViews[i], err = vk.CreateImageView(h.device, vk.ImageViewCreateInfo{
			Image: h.images[i], ViewType: vk.ImageViewType2D, Format: h.imageFormat,
			SubresourceRange: colorRange,
		})
		chk(err)
	}

	// Present waits on a per-image semaphore, so their count follows the images.
	for _, s := range h.rendered {
		vk.DestroySemaphore(h.device, s)
	}
	h.rendered = make([]vk.Semaphore, len(h.images))
	for i := range h.rendered {
		h.rendered[i], err = vk.CreateSemaphore(h.device)
		chk(err)
	}
}

var colorRange = vk.ImageSubresourceRange{AspectMask: vk.ImageAspectColor, LevelCount: 1, LayerCount: 1}

func (h *Host) createDescriptors() {
	var err error
	h.setLayout, err = vk.CreateDescriptorSetLayout(h.device, vk.DescriptorSetLayoutCreateInfo{
		UseBindingFlags: true,
		Bindings: []vk.DescriptorSetLayoutBinding{{
			Binding: 0, DescriptorType: vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: maxTextures, StageFlags: vk.ShaderStageFragment,
			BindingFlags: vk.DescriptorBindingVariableDescriptorCount | vk.DescriptorBindingPartiallyBound,
		}},
	})
	chk(err)
	h.pool, err = vk.CreateDescriptorPool(h.device, vk.DescriptorPoolCreateInfo{
		MaxSets:   1,
		PoolSizes: []vk.DescriptorPoolSize{{Type: vk.DescriptorTypeCombinedImageSampler, DescriptorCount: maxTextures}},
	})
	chk(err)
	sets, err := vk.AllocateDescriptorSets(h.device, vk.DescriptorSetAllocateInfo{
		Pool:           h.pool,
		Layouts:        []vk.DescriptorSetLayout{h.setLayout},
		VariableCounts: []uint32{maxTextures},
	})
	chk(err)
	h.set = sets[0]
}

func (h *Host) createPipeline() {
	var err error
	h.vertModule, err = vk.CreateShaderModule(h.device, shaders.Vert)
	chk(err)
	h.fragModule, err = vk.CreateShaderModule(h.device, shaders.Frag)
	chk(err)

	h.layout, err = vk.CreatePipelineLayout(h.device, vk.PipelineLayoutCreateInfo{
		SetLayouts:         []vk.DescriptorSetLayout{h.setLayout},
		PushConstantRanges: []vk.PushConstantRange{{StageFlags: vk.ShaderStageVertex, Size: uint32(unsafe.Sizeof(push{}))}},
	})
	chk(err)

	h.pipeline, err = vk.CreateGraphicsPipeline(h.device, vk.GraphicsPipelineCreateInfo{
		Layout: h.layout,
		Stages: []vk.PipelineShaderStageCreateInfo{
			{Stage: vk.ShaderStageVertex, Module: h.vertModule, Name: "main"},
			{Stage: vk.ShaderStageFragment, Module: h.fragModule, Name: "main"},
		},
		// Empty, not nil: the quad comes from gl_VertexIndex, so there are no
		// bindings or attributes, but the struct itself is still required.
		VertexInputState:   &vk.PipelineVertexInputStateCreateInfo{},
		InputAssemblyState: &vk.PipelineInputAssemblyStateCreateInfo{Topology: vk.PrimitiveTopologyTriangleList},
		ViewportState:      &vk.PipelineViewportStateCreateInfo{ViewportCount: 1, ScissorCount: 1},
		// CullModeNone sidesteps the usual silent-geometry trap: no winding to
		// get wrong, and a UI quad is never seen from behind anyway.
		RasterizationState: &vk.PipelineRasterizationStateCreateInfo{PolygonMode: vk.PolygonModeFill, CullMode: vk.CullModeNone, LineWidth: 1},
		MultisampleState:   &vk.PipelineMultisampleStateCreateInfo{RasterizationSamples: vk.SampleCount1Bit},
		// No depth state and no depth attachment: the draw list is already in
		// back-to-front order, which is the whole point of a painter's list.
		ColorBlendState: &vk.PipelineColorBlendStateCreateInfo{
			Attachments: []vk.PipelineColorBlendAttachmentState{{
				BlendEnable:         true,
				SrcColorBlendFactor: vk.BlendFactorSrcAlpha,
				DstColorBlendFactor: vk.BlendFactorOneMinusSrcAlpha,
				ColorBlendOp:        vk.BlendOpAdd,
				SrcAlphaBlendFactor: vk.BlendFactorOne,
				DstAlphaBlendFactor: vk.BlendFactorOneMinusSrcAlpha,
				AlphaBlendOp:        vk.BlendOpAdd,
				ColorWriteMask:      0xF,
			}},
		},
		DynamicState: &vk.PipelineDynamicStateCreateInfo{
			DynamicStates: []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor},
		},
		Rendering: &vk.PipelineRenderingCreateInfo{ColorAttachmentFormats: []vk.Format{h.imageFormat}},
	})
	chk(err)
}

// ---- textures ------------------------------------------------------------

// upload creates an image for one gutter Texture and copies its pixels in. It is
// called at most once per Texture.Key for the life of the process, because
// gutter's own caches hand back the same *Texture (and so the same Key) every
// frame.
func (h *Host) upload(t *ui.Texture) *hostTexture {
	if ht, ok := h.textures[t.Key]; ok {
		return ht
	}
	if h.nextSlot >= maxTextures {
		if !h.slotsFull {
			log.Printf("texture slots exhausted at %d; further textures draw untextured", maxTextures)
			h.slotsFull = true
		}
		return nil
	}

	ht := &hostTexture{slot: h.nextSlot}
	h.nextSlot++

	var err error
	ht.image, ht.alloc, err = h.allocator.VmaCreateImage(vk.ImageCreateInfo{
		ImageType: vk.ImageType2D,
		Format:    vk.FormatR8G8B8A8Unorm,
		Extent:    vk.Extent3D{Width: uint32(t.W), Height: uint32(t.H), Depth: 1},
		Usage:     vk.ImageUsageTransferDst | vk.ImageUsageSampled,
	}, vk.VmaAllocationCreateInfo{Usage: vk.VmaMemoryUsageAuto})
	chk(err)
	ht.view, err = vk.CreateImageView(h.device, vk.ImageViewCreateInfo{
		Image: ht.image, ViewType: vk.ImageViewType2D, Format: vk.FormatR8G8B8A8Unorm,
		SubresourceRange: colorRange,
	})
	chk(err)

	// gutter guarantees Pixels is NRGBA with stride W*4, which is exactly what
	// an R8G8B8A8 upload wants -- no repacking, no row-by-row copy.
	pix := packed(t.Pixels)
	staging, stagingAlloc, stagingInfo, err := h.allocator.VmaCreateBuffer(
		vk.BufferCreateInfo{Size: uint64(len(pix)), Usage: vk.BufferUsageTransferSrc},
		vk.VmaAllocationCreateInfo{
			Flags: vk.VmaAllocationCreateHostAccessSequentialWrite | vk.VmaAllocationCreateMapped,
			Usage: vk.VmaMemoryUsageAuto,
		})
	chk(err)
	vk.MemCopy(stagingInfo.MappedData, pix)

	h.oneShot(func(cb vk.CommandBuffer) {
		vk.CmdPipelineBarrier2(cb, []vk.ImageMemoryBarrier2{{
			SrcStageMask: vk.PipelineStage2None, SrcAccessMask: vk.Access2None,
			DstStageMask: vk.PipelineStage2Transfer, DstAccessMask: vk.Access2TransferWrite,
			OldLayout: vk.ImageLayoutUndefined, NewLayout: vk.ImageLayoutTransferDstOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image: ht.image, SubresourceRange: colorRange,
		}})
		vk.CmdCopyBufferToImage(cb, staging, ht.image, vk.ImageLayoutTransferDstOptimal, []vk.BufferImageCopy{{
			AspectMask: vk.ImageAspectColor, LayerCount: 1,
			ImageExtent: vk.Extent3D{Width: uint32(t.W), Height: uint32(t.H), Depth: 1},
		}})
		vk.CmdPipelineBarrier2(cb, []vk.ImageMemoryBarrier2{{
			SrcStageMask: vk.PipelineStage2Transfer, SrcAccessMask: vk.Access2TransferWrite,
			DstStageMask: vk.PipelineStage2FragmentShader, DstAccessMask: vk.Access2ShaderRead,
			OldLayout: vk.ImageLayoutTransferDstOptimal, NewLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image: ht.image, SubresourceRange: colorRange,
		}})
	})
	h.allocator.VmaDestroyBuffer(staging, stagingAlloc)

	h.textures[t.Key] = ht
	return ht
}

// packed returns the tightly packed RGBA8 bytes of img. gutter's textures always
// have Stride == W*4 already; the copy is the fallback for anything that does not.
func packed(img *image.NRGBA) []byte {
	w, hgt := img.Bounds().Dx(), img.Bounds().Dy()
	if img.Stride == w*4 {
		return img.Pix[:w*hgt*4]
	}
	out := make([]byte, w*hgt*4)
	for y := 0; y < hgt; y++ {
		copy(out[y*w*4:(y+1)*w*4], img.Pix[y*img.Stride:])
	}
	return out
}

// oneShot records and runs a command buffer, waiting for it to finish.
func (h *Host) oneShot(record func(vk.CommandBuffer)) {
	fence, err := vk.CreateFence(h.device, 0)
	chk(err)
	bufs, err := vk.AllocateCommandBuffers(h.device, h.commandPool, 1)
	chk(err)
	cb := bufs[0]
	chk(vk.BeginCommandBuffer(cb, vk.CommandBufferUsageOneTimeSubmit))
	record(cb)
	chk(vk.EndCommandBuffer(cb))
	chk(vk.QueueSubmit2(h.queue, []vk.SubmitInfo2{{CommandBuffers: []vk.CommandBuffer{cb}}}, fence))
	chk(vk.WaitForFences(h.device, []vk.Fence{fence}, true, math.MaxUint64))
	vk.DestroyFence(h.device, fence)
}

// prepare uploads any texture in the list the host has not seen and writes its
// descriptor. Writing a descriptor in a set that an in-flight command buffer has
// bound is illegal without update-after-bind, so the rare frame that introduces
// a texture pays a DeviceWaitIdle. New textures appear on the first frame and
// then only when a new string or image shows up, so this is not a per-frame cost.
func (h *Host) prepare(dl *ui.DrawList) {
	var writes []vk.WriteDescriptorSet
	for i := range dl.Cmds {
		t := dl.Cmds[i].Tex
		if t == nil {
			continue
		}
		if _, seen := h.textures[t.Key]; seen {
			continue
		}
		ht := h.upload(t)
		if ht == nil {
			continue
		}
		writes = append(writes, vk.WriteDescriptorSet{
			DstSet: h.set, DstBinding: 0, DstArrayElement: uint32(ht.slot),
			DescriptorType: vk.DescriptorTypeCombinedImageSampler,
			ImageInfo: []vk.DescriptorImageInfo{{
				Sampler: h.sampler, ImageView: ht.view, ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			}},
		})
	}
	if len(writes) > 0 {
		chk(vk.DeviceWaitIdle(h.device))
		vk.UpdateDescriptorSets(h.device, writes)
	}
}

// ---- frame ---------------------------------------------------------------

// Render draws one DrawList and presents it: one Draw per Cmd, in list order.
func (h *Host) Render(dl *ui.DrawList) {
	if h.extent.Width == 0 || h.extent.Height == 0 {
		return
	}
	h.prepare(dl)

	chk(vk.WaitForFences(h.device, []vk.Fence{h.fences[h.frame]}, true, math.MaxUint64))

	imageIndex, err := vk.AcquireNextImageKHR(h.device, h.swapchain, math.MaxUint64, h.acquired[h.frame], vk.Fence(0))
	if err == vk.ErrOutOfDateKHR {
		h.needsResize = true
		return
	}
	if err != vk.SuboptimalKHR {
		chk(err)
	}
	// Only reset the fence once the frame is definitely going to be submitted;
	// resetting before a bailout would deadlock the next wait on this slot.
	chk(vk.ResetFences(h.device, []vk.Fence{h.fences[h.frame]}))

	cb := h.commandBuffers[h.frame]
	chk(vk.ResetCommandBuffer(cb))
	chk(vk.BeginCommandBuffer(cb, vk.CommandBufferUsageOneTimeSubmit))

	vk.CmdPipelineBarrier2(cb, []vk.ImageMemoryBarrier2{{
		SrcStageMask: vk.PipelineStage2ColorAttachmentOutput, SrcAccessMask: vk.Access2None,
		DstStageMask: vk.PipelineStage2ColorAttachmentOutput, DstAccessMask: vk.Access2ColorAttachmentWrite,
		OldLayout: vk.ImageLayoutUndefined, NewLayout: vk.ImageLayoutColorAttachmentOptimal,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image: h.images[imageIndex], SubresourceRange: colorRange,
	}})

	vk.CmdBeginRendering(cb, vk.RenderingInfo{
		RenderArea: vk.Rect2D{Extent: h.extent},
		LayerCount: 1,
		ColorAttachments: []vk.RenderingAttachmentInfo{{
			ImageView: h.imageViews[imageIndex], ImageLayout: vk.ImageLayoutColorAttachmentOptimal,
			LoadOp: vk.AttachmentLoadOpClear, StoreOp: vk.AttachmentStoreOpStore,
			ClearValue: vk.ClearColor(0, 0, 0, 1),
		}},
	})

	vk.CmdSetViewport(cb, vk.Viewport{Width: float32(h.extent.Width), Height: float32(h.extent.Height), MaxDepth: 1})
	vk.CmdSetScissor(cb, vk.Rect2D{Extent: h.extent})
	vk.CmdBindPipeline(cb, vk.PipelineBindPointGraphics, h.pipeline)
	vk.CmdBindDescriptorSets(cb, vk.PipelineBindPointGraphics, h.layout, 0, []vk.DescriptorSet{h.set})

	screen := [2]float32{float32(h.extent.Width), float32(h.extent.Height)}
	pcSize := uint32(unsafe.Sizeof(push{}))
	for i := range dl.Cmds {
		c := &dl.Cmds[i]
		pc := push{
			Color: [4]float32{
				float32(c.Color.R) / 255, float32(c.Color.G) / 255,
				float32(c.Color.B) / 255, float32(c.Color.A) / 255,
			},
			Screen: screen,
			Pos:    [2]float32{float32(c.Rect.X), float32(c.Rect.Y)},
			Size:   [2]float32{float32(c.Rect.W), float32(c.Rect.H)},
			Tex:    -1,
		}
		if c.Tex != nil {
			if ht, ok := h.textures[c.Tex.Key]; ok {
				pc.Tex = ht.slot
			}
		}
		vk.CmdPushConstants(cb, h.layout, vk.ShaderStageVertex, 0, pcSize, unsafe.Pointer(&pc))
		vk.CmdDraw(cb, 6, 1, 0, 0)
	}

	vk.CmdEndRendering(cb)

	vk.CmdPipelineBarrier2(cb, []vk.ImageMemoryBarrier2{{
		SrcStageMask: vk.PipelineStage2ColorAttachmentOutput, SrcAccessMask: vk.Access2ColorAttachmentWrite,
		DstStageMask: vk.PipelineStage2ColorAttachmentOutput, DstAccessMask: vk.Access2None,
		OldLayout: vk.ImageLayoutColorAttachmentOptimal, NewLayout: vk.ImageLayoutPresentSrcKHR,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image: h.images[imageIndex], SubresourceRange: colorRange,
	}})
	chk(vk.EndCommandBuffer(cb))

	chk(vk.QueueSubmit2(h.queue, []vk.SubmitInfo2{{
		WaitSemaphores:   []vk.SemaphoreSubmitInfo{{Semaphore: h.acquired[h.frame], StageMask: vk.PipelineStage2ColorAttachmentOutput}},
		CommandBuffers:   []vk.CommandBuffer{cb},
		SignalSemaphores: []vk.SemaphoreSubmitInfo{{Semaphore: h.rendered[imageIndex], StageMask: vk.PipelineStage2AllCommands}},
	}}, h.fences[h.frame]))
	h.frame = (h.frame + 1) % framesInFlight

	if err := vk.QueuePresentKHR(h.queue, h.rendered[imageIndex], h.swapchain, imageIndex); err != nil {
		if err == vk.ErrOutOfDateKHR || err == vk.SuboptimalKHR {
			h.needsResize = true
		} else {
			chk(err)
		}
	}
}

// ---- input and loop ------------------------------------------------------

func (h *Host) ShouldClose() bool { return h.window.ShouldClose() }
func (h *Host) Quit()             { h.window.SetShouldClose(true) }

// Poll pumps events, resizes if needed, and latches a left-button press edge so
// a held button fires once rather than every frame.
func (h *Host) Poll() {
	glfw.PollEvents()

	down := h.window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press
	h.clicked = down && !h.mouseDown
	h.mouseDown = down

	x, y := h.window.GetCursorPos()
	// The cursor is in window coordinates; the draw list is in framebuffer
	// pixels. They are the same size only when the display scale is 1.
	winW, winH := h.window.GetSize()
	fbW, fbH := h.window.GetFramebufferSize()
	if winW > 0 && winH > 0 {
		x *= float64(fbW) / float64(winW)
		y *= float64(fbH) / float64(winH)
	}
	h.cursorX, h.cursorY = x, y

	if h.needsResize {
		for {
			w, ht := h.window.GetFramebufferSize()
			if w > 0 && ht > 0 {
				break
			}
			glfw.WaitEvents() // minimised
		}
		h.needsResize = false
		chk(vk.DeviceWaitIdle(h.device))
		h.createSwapchain()
	}
}

// Input is the snapshot gutter's widgets read. It is deliberately tiny -- a
// cursor and a viewport -- which is what lets ui/ stay free of any window or
// graphics dependency.
func (h *Host) Input() ui.Input {
	return ui.Input{
		CursorX: h.cursorX, CursorY: h.cursorY,
		Width: int(h.extent.Width), Height: int(h.extent.Height),
	}
}

// Clicked reports whether the left button went down this frame.
func (h *Host) Clicked() bool { return h.clicked }

func (h *Host) Close() {
	chk(vk.DeviceWaitIdle(h.device))
	for _, t := range h.textures {
		vk.DestroyImageView(h.device, t.view)
		h.allocator.VmaDestroyImage(t.image, t.alloc)
	}
	vk.DestroySampler(h.device, h.sampler)
	vk.DestroyPipeline(h.device, h.pipeline)
	vk.DestroyPipelineLayout(h.device, h.layout)
	vk.DestroyShaderModule(h.device, h.vertModule)
	vk.DestroyShaderModule(h.device, h.fragModule)
	vk.DestroyDescriptorPool(h.device, h.pool)
	vk.DestroyDescriptorSetLayout(h.device, h.setLayout)
	for i := 0; i < framesInFlight; i++ {
		vk.DestroyFence(h.device, h.fences[i])
		vk.DestroySemaphore(h.device, h.acquired[i])
	}
	for _, s := range h.rendered {
		vk.DestroySemaphore(h.device, s)
	}
	vk.DestroyCommandPool(h.device, h.commandPool)
	for _, v := range h.imageViews {
		vk.DestroyImageView(h.device, v)
	}
	vk.DestroySwapchainKHR(h.device, h.swapchain)
	vk.DestroySurfaceKHR(h.instance, h.surface)
	vk.VmaDestroyAllocator(h.allocator)
	vk.DestroyDevice(h.device)
	vk.DestroyInstance(h.instance)
	h.window.Destroy()
	glfw.Terminate()
}
