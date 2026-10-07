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
	host := &Host{textures: map[uint64]*hostTexture{}}

	chk(glfw.Init())
	if !glfw.VulkanSupported() {
		log.Fatal("glfw: no Vulkan support")
	}
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	chk(err)
	host.window = window
	window.SetFramebufferSizeCallback(func(*glfw.Window, int, int) { host.needsResize = true })

	host.instance, err = vk.CreateInstance(vk.InstanceCreateInfo{
		AppName:    title,
		APIVersion: vk.ApiVersion13,
		Extensions: window.GetRequiredInstanceExtensions(),
	})
	chk(err)

	devices, err := vk.EnumeratePhysicalDevices(host.instance)
	chk(err)
	if len(devices) == 0 {
		log.Fatal("vulkan: no physical devices")
	}
	host.physicalDevice = devices[0]
	log.Printf("device: %s", vk.GetPhysicalDeviceProperties2(host.physicalDevice).DeviceName)

	for i, qf := range vk.GetPhysicalDeviceQueueFamilyProperties(host.physicalDevice) {
		if qf.QueueFlags&vk.QueueGraphics != 0 {
			host.queueFamily = uint32(i)
			break
		}
	}

	// Descriptor indexing is the only interesting feature here: the fragment
	// shader indexes a runtime-sized texture array by a push constant, and most
	// of its slots are empty at any given moment, hence PartiallyBound.
	host.device, err = vk.CreateDevice(host.physicalDevice, vk.DeviceCreateInfo{
		QueueCreateInfos: []vk.DeviceQueueCreateInfo{{QueueFamilyIndex: host.queueFamily, Priorities: []float32{1}}},
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
	host.queue = vk.GetDeviceQueue(host.device, host.queueFamily, 0)

	host.allocator = vk.VmaCreateAllocator(vk.VmaAllocatorCreateInfo{
		PhysicalDevice: host.physicalDevice,
		Device:         host.device,
		Instance:       host.instance,
	})

	// GLFW takes the instance as a pointer-kind value and returns a pointer to
	// the created surface handle, so both ends round-trip through
	// unsafe.Pointer. `go vet` reports "possible misuse of unsafe.Pointer" here:
	// these are opaque Vulkan handles, not Go pointers, so the warning does not
	// apply -- but it is the price of this binding pair.
	surfRaw, err := window.CreateWindowSurface((*byte)(unsafe.Pointer(host.instance)), nil)
	chk(err)
	host.surface = vk.SurfaceKHR(*(*uintptr)(unsafe.Pointer(surfRaw)))
	presentOK, err := vk.GetPhysicalDeviceSurfaceSupportKHR(host.physicalDevice, host.queueFamily, host.surface)
	chk(err)
	if !presentOK {
		log.Fatal("vulkan: graphics queue cannot present")
	}

	host.imageFormat = pickFormat(host.physicalDevice, host.surface)
	host.createSwapchain()

	host.commandPool, err = vk.CreateCommandPool(host.device, host.queueFamily, vk.CommandPoolCreateResetCommandBuffer)
	chk(err)
	host.commandBuffers, err = vk.AllocateCommandBuffers(host.device, host.commandPool, framesInFlight)
	chk(err)
	for i := 0; i < framesInFlight; i++ {
		host.fences[i], err = vk.CreateFence(host.device, vk.FenceCreateSignaled)
		chk(err)
		host.acquired[i], err = vk.CreateSemaphore(host.device)
		chk(err)
	}

	host.createDescriptors()
	host.createPipeline()

	host.sampler, err = vk.CreateSampler(host.device, vk.SamplerCreateInfo{
		MagFilter: vk.FilterLinear, MinFilter: vk.FilterLinear,
		MipmapMode:   vk.SamplerMipmapModeLinear,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
		MaxLod:       1,
	})
	chk(err)

	return host
}

// pickFormat prefers a UNORM swapchain. An _SRGB one would have the hardware
// encode on write, but gutter's colours and its PNGs are already sRGB-encoded
// bytes; UNORM presents them unchanged, so the window matches what the library's
// own CPU replay produces.
func pickFormat(physicalDevice vk.PhysicalDevice, surface vk.SurfaceKHR) vk.Format {
	formats, err := vk.GetPhysicalDeviceSurfaceFormatsKHR(physicalDevice, surface)
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

func (host *Host) createSwapchain() {
	caps, err := vk.GetPhysicalDeviceSurfaceCapabilitiesKHR(host.physicalDevice, host.surface)
	chk(err)

	w, ht := host.window.GetFramebufferSize()
	host.extent = vk.Extent2D{Width: uint32(w), Height: uint32(ht)}
	if caps.CurrentExtent.Width != 0xFFFFFFFF {
		host.extent = caps.CurrentExtent
	}

	host.swapchainCI = vk.SwapchainCreateInfo{
		Surface:         host.surface,
		MinImageCount:   caps.MinImageCount,
		ImageFormat:     host.imageFormat,
		ImageColorSpace: vk.ColorSpaceSrgbNonlinearKHR,
		ImageExtent:     host.extent,
		ImageUsage:      vk.ImageUsageColorAttachment,
		PreTransform:    vk.SurfaceTransformIdentityKHR,
		CompositeAlpha:  vk.CompositeAlphaOpaqueKHR,
		PresentMode:     vk.PresentModeFifoKHR,
		OldSwapchain:    host.swapchain,
	}
	newSwapchain, err := vk.CreateSwapchainKHR(host.device, host.swapchainCI)
	chk(err)
	if host.swapchain != 0 {
		vk.DestroySwapchainKHR(host.device, host.swapchain)
	}
	host.swapchain = newSwapchain

	for _, v := range host.imageViews {
		vk.DestroyImageView(host.device, v)
	}
	host.images, err = vk.GetSwapchainImagesKHR(host.device, host.swapchain)
	chk(err)
	host.imageViews = make([]vk.ImageView, len(host.images))
	for i := range host.images {
		host.imageViews[i], err = vk.CreateImageView(host.device, vk.ImageViewCreateInfo{
			Image: host.images[i], ViewType: vk.ImageViewType2D, Format: host.imageFormat,
			SubresourceRange: colorRange,
		})
		chk(err)
	}

	// Present waits on a per-image semaphore, so their count follows the images.
	for _, s := range host.rendered {
		vk.DestroySemaphore(host.device, s)
	}
	host.rendered = make([]vk.Semaphore, len(host.images))
	for i := range host.rendered {
		host.rendered[i], err = vk.CreateSemaphore(host.device)
		chk(err)
	}
}

var colorRange = vk.ImageSubresourceRange{AspectMask: vk.ImageAspectColor, LevelCount: 1, LayerCount: 1}

func (host *Host) createDescriptors() {
	var err error
	host.setLayout, err = vk.CreateDescriptorSetLayout(host.device, vk.DescriptorSetLayoutCreateInfo{
		UseBindingFlags: true,
		Bindings: []vk.DescriptorSetLayoutBinding{{
			Binding: 0, DescriptorType: vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: maxTextures, StageFlags: vk.ShaderStageFragment,
			BindingFlags: vk.DescriptorBindingVariableDescriptorCount | vk.DescriptorBindingPartiallyBound,
		}},
	})
	chk(err)
	host.pool, err = vk.CreateDescriptorPool(host.device, vk.DescriptorPoolCreateInfo{
		MaxSets:   1,
		PoolSizes: []vk.DescriptorPoolSize{{Type: vk.DescriptorTypeCombinedImageSampler, DescriptorCount: maxTextures}},
	})
	chk(err)
	sets, err := vk.AllocateDescriptorSets(host.device, vk.DescriptorSetAllocateInfo{
		Pool:           host.pool,
		Layouts:        []vk.DescriptorSetLayout{host.setLayout},
		VariableCounts: []uint32{maxTextures},
	})
	chk(err)
	host.set = sets[0]
}

func (host *Host) createPipeline() {
	var err error
	host.vertModule, err = vk.CreateShaderModule(host.device, shaders.Vert)
	chk(err)
	host.fragModule, err = vk.CreateShaderModule(host.device, shaders.Frag)
	chk(err)

	host.layout, err = vk.CreatePipelineLayout(host.device, vk.PipelineLayoutCreateInfo{
		SetLayouts:         []vk.DescriptorSetLayout{host.setLayout},
		PushConstantRanges: []vk.PushConstantRange{{StageFlags: vk.ShaderStageVertex, Size: uint32(unsafe.Sizeof(push{}))}},
	})
	chk(err)

	host.pipeline, err = vk.CreateGraphicsPipeline(host.device, vk.GraphicsPipelineCreateInfo{
		Layout: host.layout,
		Stages: []vk.PipelineShaderStageCreateInfo{
			{Stage: vk.ShaderStageVertex, Module: host.vertModule, Name: "main"},
			{Stage: vk.ShaderStageFragment, Module: host.fragModule, Name: "main"},
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
		Rendering: &vk.PipelineRenderingCreateInfo{ColorAttachmentFormats: []vk.Format{host.imageFormat}},
	})
	chk(err)
}

// ---- textures ------------------------------------------------------------

// upload creates an image for one gutter Texture and copies its pixels in. It is
// called at most once per Texture.Key for the life of the process, because
// gutter's own caches hand back the same *Texture (and so the same Key) every
// frame.
func (host *Host) upload(t *ui.Texture) *hostTexture {
	if ht, ok := host.textures[t.Key]; ok {
		return ht
	}
	if host.nextSlot >= maxTextures {
		if !host.slotsFull {
			log.Printf("texture slots exhausted at %d; further textures draw untextured", maxTextures)
			host.slotsFull = true
		}
		return nil
	}

	hostTex := &hostTexture{slot: host.nextSlot}
	host.nextSlot++

	var err error
	hostTex.image, hostTex.alloc, err = host.allocator.VmaCreateImage(vk.ImageCreateInfo{
		ImageType: vk.ImageType2D,
		Format:    vk.FormatR8G8B8A8Unorm,
		Extent:    vk.Extent3D{Width: uint32(t.W), Height: uint32(t.H), Depth: 1},
		Usage:     vk.ImageUsageTransferDst | vk.ImageUsageSampled,
	}, vk.VmaAllocationCreateInfo{Usage: vk.VmaMemoryUsageAuto})
	chk(err)
	hostTex.view, err = vk.CreateImageView(host.device, vk.ImageViewCreateInfo{
		Image: hostTex.image, ViewType: vk.ImageViewType2D, Format: vk.FormatR8G8B8A8Unorm,
		SubresourceRange: colorRange,
	})
	chk(err)

	// gutter guarantees Pixels is NRGBA with stride W*4, which is exactly what
	// an R8G8B8A8 upload wants -- no repacking, no row-by-row copy.
	pix := packed(t.Pixels)
	staging, stagingAlloc, stagingInfo, err := host.allocator.VmaCreateBuffer(
		vk.BufferCreateInfo{Size: uint64(len(pix)), Usage: vk.BufferUsageTransferSrc},
		vk.VmaAllocationCreateInfo{
			Flags: vk.VmaAllocationCreateHostAccessSequentialWrite | vk.VmaAllocationCreateMapped,
			Usage: vk.VmaMemoryUsageAuto,
		})
	chk(err)
	vk.MemCopy(stagingInfo.MappedData, pix)

	host.oneShot(func(cb vk.CommandBuffer) {
		vk.CmdPipelineBarrier2(cb, vk.DependencyInfo{Image: []vk.ImageMemoryBarrier2{{
			SrcStageMask: vk.PipelineStage2None, SrcAccessMask: vk.Access2None,
			DstStageMask: vk.PipelineStage2Transfer, DstAccessMask: vk.Access2TransferWrite,
			OldLayout: vk.ImageLayoutUndefined, NewLayout: vk.ImageLayoutTransferDstOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image: hostTex.image, SubresourceRange: colorRange,
		}}})
		vk.CmdCopyBufferToImage(cb, staging, hostTex.image, vk.ImageLayoutTransferDstOptimal, []vk.BufferImageCopy{{
			AspectMask: vk.ImageAspectColor, LayerCount: 1,
			ImageExtent: vk.Extent3D{Width: uint32(t.W), Height: uint32(t.H), Depth: 1},
		}})
		vk.CmdPipelineBarrier2(cb, vk.DependencyInfo{Image: []vk.ImageMemoryBarrier2{{
			SrcStageMask: vk.PipelineStage2Transfer, SrcAccessMask: vk.Access2TransferWrite,
			DstStageMask: vk.PipelineStage2FragmentShader, DstAccessMask: vk.Access2ShaderRead,
			OldLayout: vk.ImageLayoutTransferDstOptimal, NewLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image: hostTex.image, SubresourceRange: colorRange,
		}}})
	})
	host.allocator.VmaDestroyBuffer(staging, stagingAlloc)

	host.textures[t.Key] = hostTex
	return hostTex
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
func (host *Host) oneShot(record func(vk.CommandBuffer)) {
	fence, err := vk.CreateFence(host.device, 0)
	chk(err)
	bufs, err := vk.AllocateCommandBuffers(host.device, host.commandPool, 1)
	chk(err)
	cb := bufs[0]
	chk(vk.BeginCommandBuffer(cb, vk.CommandBufferUsageOneTimeSubmit))
	record(cb)
	chk(vk.EndCommandBuffer(cb))
	chk(vk.QueueSubmit2(host.queue, []vk.SubmitInfo2{{CommandBuffers: []vk.CommandBuffer{cb}}}, fence))
	chk(vk.WaitForFences(host.device, []vk.Fence{fence}, true, math.MaxUint64))
	vk.DestroyFence(host.device, fence)
}

// prepare uploads any texture in the list the host has not seen and writes its
// descriptor. Writing a descriptor in a set that an in-flight command buffer has
// bound is illegal without update-after-bind, so the rare frame that introduces
// a texture pays a DeviceWaitIdle. New textures appear on the first frame and
// then only when a new string or image shows up, so this is not a per-frame cost.
func (host *Host) prepare(drawList *ui.DrawList) {
	var writes []vk.WriteDescriptorSet
	for i := range drawList.Cmds {
		t := drawList.Cmds[i].Tex
		if t == nil {
			continue
		}
		if _, seen := host.textures[t.Key]; seen {
			continue
		}
		ht := host.upload(t)
		if ht == nil {
			continue
		}
		writes = append(writes, vk.WriteDescriptorSet{
			DstSet: host.set, DstBinding: 0, DstArrayElement: uint32(ht.slot),
			DescriptorType: vk.DescriptorTypeCombinedImageSampler,
			ImageInfo: []vk.DescriptorImageInfo{{
				Sampler: host.sampler, ImageView: ht.view, ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			}},
		})
	}
	if len(writes) > 0 {
		chk(vk.DeviceWaitIdle(host.device))
		vk.UpdateDescriptorSets(host.device, writes)
	}
}

// ---- frame ---------------------------------------------------------------

// Render draws one DrawList and presents it: one Draw per Cmd, in list order.
func (host *Host) Render(drawList *ui.DrawList) {
	if host.extent.Width == 0 || host.extent.Height == 0 {
		return
	}
	host.prepare(drawList)

	chk(vk.WaitForFences(host.device, []vk.Fence{host.fences[host.frame]}, true, math.MaxUint64))

	imageIndex, err := vk.AcquireNextImageKHR(host.device, host.swapchain, math.MaxUint64, host.acquired[host.frame], vk.Fence(0))
	if err == vk.ErrOutOfDateKHR {
		host.needsResize = true
		return
	}
	if err != vk.SuboptimalKHR {
		chk(err)
	}
	// Only reset the fence once the frame is definitely going to be submitted;
	// resetting before a bailout would deadlock the next wait on this slot.
	chk(vk.ResetFences(host.device, []vk.Fence{host.fences[host.frame]}))

	commandBuffer := host.commandBuffers[host.frame]
	chk(vk.ResetCommandBuffer(commandBuffer))
	chk(vk.BeginCommandBuffer(commandBuffer, vk.CommandBufferUsageOneTimeSubmit))

	vk.CmdPipelineBarrier2(commandBuffer, vk.DependencyInfo{Image: []vk.ImageMemoryBarrier2{{
		SrcStageMask: vk.PipelineStage2ColorAttachmentOutput, SrcAccessMask: vk.Access2None,
		DstStageMask: vk.PipelineStage2ColorAttachmentOutput, DstAccessMask: vk.Access2ColorAttachmentWrite,
		OldLayout: vk.ImageLayoutUndefined, NewLayout: vk.ImageLayoutColorAttachmentOptimal,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image: host.images[imageIndex], SubresourceRange: colorRange,
	}}})

	vk.CmdBeginRendering(commandBuffer, vk.RenderingInfo{
		RenderArea: vk.Rect2D{Extent: host.extent},
		LayerCount: 1,
		ColorAttachments: []vk.RenderingAttachmentInfo{{
			ImageView: host.imageViews[imageIndex], ImageLayout: vk.ImageLayoutColorAttachmentOptimal,
			LoadOp: vk.AttachmentLoadOpClear, StoreOp: vk.AttachmentStoreOpStore,
			ClearValue: vk.ClearColor(0, 0, 0, 1),
		}},
	})

	vk.CmdSetViewport(commandBuffer, vk.Viewport{Width: float32(host.extent.Width), Height: float32(host.extent.Height), MaxDepth: 1})
	vk.CmdSetScissor(commandBuffer, vk.Rect2D{Extent: host.extent})
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, host.pipeline)
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, host.layout, 0, []vk.DescriptorSet{host.set})

	screen := [2]float32{float32(host.extent.Width), float32(host.extent.Height)}
	pcSize := uint32(unsafe.Sizeof(push{}))
	for i := range drawList.Cmds {
		c := &drawList.Cmds[i]
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
			if ht, ok := host.textures[c.Tex.Key]; ok {
				pc.Tex = ht.slot
			}
		}
		vk.CmdPushConstants(commandBuffer, host.layout, vk.ShaderStageVertex, 0, pcSize, unsafe.Pointer(&pc))
		vk.CmdDraw(commandBuffer, 6, 1, 0, 0)
	}

	vk.CmdEndRendering(commandBuffer)

	vk.CmdPipelineBarrier2(commandBuffer, vk.DependencyInfo{Image: []vk.ImageMemoryBarrier2{{
		SrcStageMask: vk.PipelineStage2ColorAttachmentOutput, SrcAccessMask: vk.Access2ColorAttachmentWrite,
		DstStageMask: vk.PipelineStage2ColorAttachmentOutput, DstAccessMask: vk.Access2None,
		OldLayout: vk.ImageLayoutColorAttachmentOptimal, NewLayout: vk.ImageLayoutPresentSrcKHR,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored, DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image: host.images[imageIndex], SubresourceRange: colorRange,
	}}})
	chk(vk.EndCommandBuffer(commandBuffer))

	chk(vk.QueueSubmit2(host.queue, []vk.SubmitInfo2{{
		WaitSemaphores:   []vk.SemaphoreSubmitInfo{{Semaphore: host.acquired[host.frame], StageMask: vk.PipelineStage2ColorAttachmentOutput}},
		CommandBuffers:   []vk.CommandBuffer{commandBuffer},
		SignalSemaphores: []vk.SemaphoreSubmitInfo{{Semaphore: host.rendered[imageIndex], StageMask: vk.PipelineStage2AllCommands}},
	}}, host.fences[host.frame]))
	host.frame = (host.frame + 1) % framesInFlight

	if err := vk.QueuePresentKHR(host.queue, host.rendered[imageIndex], host.swapchain, imageIndex); err != nil {
		if err == vk.ErrOutOfDateKHR || err == vk.SuboptimalKHR {
			host.needsResize = true
		} else {
			chk(err)
		}
	}
}

// ---- input and loop ------------------------------------------------------

func (host *Host) ShouldClose() bool { return host.window.ShouldClose() }
func (host *Host) Quit()             { host.window.SetShouldClose(true) }

// Poll pumps events, resizes if needed, and latches a left-button press edge so
// a held button fires once rather than every frame.
func (host *Host) Poll() {
	glfw.PollEvents()

	down := host.window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press
	host.clicked = down && !host.mouseDown
	host.mouseDown = down

	x, y := host.window.GetCursorPos()
	// The cursor is in window coordinates; the draw list is in framebuffer
	// pixels. They are the same size only when the display scale is 1.
	winW, winH := host.window.GetSize()
	fbW, fbH := host.window.GetFramebufferSize()
	if winW > 0 && winH > 0 {
		x *= float64(fbW) / float64(winW)
		y *= float64(fbH) / float64(winH)
	}
	host.cursorX, host.cursorY = x, y

	if host.needsResize {
		for {
			w, ht := host.window.GetFramebufferSize()
			if w > 0 && ht > 0 {
				break
			}
			glfw.WaitEvents() // minimised
		}
		host.needsResize = false
		chk(vk.DeviceWaitIdle(host.device))
		host.createSwapchain()
	}
}

// Input is the snapshot gutter's widgets read. It is deliberately tiny -- a
// cursor and a viewport -- which is what lets ui/ stay free of any window or
// graphics dependency.
func (host *Host) Input() ui.Input {
	return ui.Input{
		CursorX: host.cursorX, CursorY: host.cursorY,
		Width: int(host.extent.Width), Height: int(host.extent.Height),
	}
}

// Clicked reports whether the left button went down this frame.
func (host *Host) Clicked() bool { return host.clicked }

func (host *Host) Close() {
	chk(vk.DeviceWaitIdle(host.device))
	for _, t := range host.textures {
		vk.DestroyImageView(host.device, t.view)
		host.allocator.VmaDestroyImage(t.image, t.alloc)
	}
	vk.DestroySampler(host.device, host.sampler)
	vk.DestroyPipeline(host.device, host.pipeline)
	vk.DestroyPipelineLayout(host.device, host.layout)
	vk.DestroyShaderModule(host.device, host.vertModule)
	vk.DestroyShaderModule(host.device, host.fragModule)
	vk.DestroyDescriptorPool(host.device, host.pool)
	vk.DestroyDescriptorSetLayout(host.device, host.setLayout)
	for i := 0; i < framesInFlight; i++ {
		vk.DestroyFence(host.device, host.fences[i])
		vk.DestroySemaphore(host.device, host.acquired[i])
	}
	for _, s := range host.rendered {
		vk.DestroySemaphore(host.device, s)
	}
	vk.DestroyCommandPool(host.device, host.commandPool)
	for _, v := range host.imageViews {
		vk.DestroyImageView(host.device, v)
	}
	vk.DestroySwapchainKHR(host.device, host.swapchain)
	vk.DestroySurfaceKHR(host.instance, host.surface)
	vk.VmaDestroyAllocator(host.allocator)
	vk.DestroyDevice(host.device)
	vk.DestroyInstance(host.instance)
	host.window.Destroy()
	glfw.Terminate()
}
