package main

import (
	"fmt"
	"math"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	im "github.com/AllenDang/cimgui-go/imgui"

	"bytes"
	_ "embed"
	"image"

	cte "github.com/AllenDang/cimgui-go/ImGuiColorTextEdit"
	"github.com/AllenDang/cimgui-go/imguizmo"
	_ "github.com/AllenDang/cimgui-go/imguizmo"
	_ "github.com/AllenDang/cimgui-go/immarkdown"
	_ "github.com/AllenDang/cimgui-go/imnodes"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/AllenDang/cimgui-go/utils"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
)

var currentBackend backend.Backend[sdlbackend.SDLWindowFlags]

func init() {
	runtime.LockOSThread()
}

func sizeof[T any]() int {
	var t T
	return (int)(unsafe.Sizeof(t))
}

func main3() {
	defer binsdl.Load().Unload() // sdl.LoadLibrary(sdl.Path())
	defer sdl.Quit()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	window, renderer, err := sdl.CreateWindowAndRenderer("Hello world", 500, 500, 0)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
	defer window.Destroy()

	renderer.SetDrawColor(255, 255, 255, 255)

mainloop:
	for true {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				break mainloop
			case sdl.EVENT_KEY_DOWN:
				if event.KeyboardEvent().Scancode == sdl.SCANCODE_ESCAPE {
					break mainloop
				}
			}
		}

		renderer.DebugText(50, 50, "Hello world")
		renderer.RenderLine(100, 100, 500, 500)
		renderer.Present()
	}
}

func ImGui_ImplOpenGL4_NewFrame() {
	panic("not implemented")
}

func ImGui_ImplSDL3_GetWindowSizeAndFramebufferScale(window *sdl.Window) (out_size, out_framebuffer_scale im.Vec2) {
	w, h := Must2(window.Size())

	if (window.Flags() & sdl.WINDOW_MINIMIZED) != 0 {
		w = 0
		h = 0
	}

	display_w, display_h := Must2(window.SizeInPixels())

	out_size = im.NewVec2(float32(w), float32(h))
	if w > 0 && h > 0 {
		out_framebuffer_scale = im.NewVec2(float32(display_w)/float32(w), float32(display_h)/float32(h))
	} else {
		out_framebuffer_scale = im.NewVec2(1, 1)
	}

	return
}

func ImGui_ImplSDL3_UpdateMonitors() {
	bd := ImGui_ImplSDL3_GetBackendData()
	if bd == nil {
		panic("Context or backend not initialized! Did you call ImGui_ImplSDL3_Init()?")
	}

	platform_io := im.CurrentIO()

	//platform_io.CData.Ctx.PlatformIO.Monitors.Resize(0)
	platform_io.Monitors.resize(0) // TODO what do I do here?
	//
	//platform_io.

	bd.WantUpdateMonitors = false

	displays := Must(sdl.GetDisplays())

	// int display_count;
	// SDL_DisplayID* displays = SDL_GetDisplays(&display_count);
	for n, display_id := range displays {
		// Warning: the validity of monitor DPI information on Windows depends on the application DPI awareness settings, which generally needs to be set in the manifest or at runtime.
		var monitor im.PlatformMonitor
		var r *sdl.Rect = Must(display_id.Bounds())

		pos := im.NewVec2(float32(r.X), float32(r.Y))
		monitor.SetMainPos(pos)
		monitor.SetWorkPos(pos)
		size := im.NewVec2(float32(r.W), float32(r.H))
		monitor.SetMainSize(size)
		monitor.SetWorkSize(size)

		r = Must(display_id.UsableBounds())
		if r.W > 0 && r.H > 0 {
			monitor.SetWorkPos(im.NewVec2(float32(r.X), float32(r.Y)))
			monitor.SetWorkSize(im.NewVec2(float32(r.W), float32(r.H)))
		}

		monitor.SetDpiScale(Must(display_id.ContentScale())) // See https://wiki.libsdl.org/SDL3/README-highdpi for details.
		monitor.SetPlatformHandle(uintptr(n))
		if monitor.DpiScale() <= 0 {
			continue // Some accessibility applications are declaring virtual monitors with a DPI of 0, see #7902.
		}

		platform_io.Monitors.push_back(monitor) // TODO what do I do here?
	}
}

func ImGui_ImplSDL3_NewFrame() {

	bd := ImGui_ImplSDL3_GetBackendData()
	if bd == nil {
		panic("Context or backend not initialized! Did you call ImGui_ImplSDL3_Init()?")
	}
	io := im.CurrentIO()

	// Setup main viewport size (every frame to accommodate for window resizing)
	displaySize, displayScale := ImGui_ImplSDL3_GetWindowSizeAndFramebufferScale(bd.Window)

	io.SetDisplaySize(displaySize)
	io.SetDisplayFramebufferScale(displayScale)

	// Update monitors
	if runtime.GOOS == "windows" {
		bd.WantUpdateMonitors = true // Keep polling under Windows to handle changes of work area when resizing task-bar (#8415)
	}

	if bd.WantUpdateMonitors {
		ImGui_ImplSDL3_UpdateMonitors()
	}
	// Setup time step (we could also use SDL_GetTicksNS() available since SDL3)
	// (Accept SDL_GetPerformanceCounter() not returning a monotonically increasing value. Happens in VMs and Emscripten, see #6189, #6114, #3644)
	frequency := sdl.GetPerformanceFrequency()
	current_time := sdl.GetPerformanceCounter()
	if current_time <= bd.Time {
		current_time = bd.Time + 1
	}

	if bd.Time > 0 {
		io.SetDeltaTime(float32(current_time-bd.Time) / float32(frequency))
	} else {
		io.SetDeltaTime(1 / 60)
	}

	bd.Time = current_time

	if bd.MousePendingLeaveFrame != 0 && bd.MousePendingLeaveFrame >= im.FrameCount() && bd.MouseButtonsDown == 0 {
		bd.MouseWindowID = 0
		bd.MousePendingLeaveFrame = 0
		io.AddMousePosEvent(-math.MaxFloat32, -math.MaxFloat32)
	}

	// Our io.AddMouseViewportEvent() calls will only be valid when not capturing.
	// Technically speaking testing for 'bd.MouseButtonsDown == 0' would be more rigorous, but testing for payload reduces noise and potential side-effects.
	if bd.MouseCanReportHoveredViewport && im.DragDropPayload() == nil {
		io.SetBackendFlags(io.BackendFlags() | im.BackendFlagsHasMouseHoveredViewport)
	} else {
		io.SetBackendFlags(io.BackendFlags() & ^im.BackendFlagsHasMouseHoveredViewport)
	}

	ImGui_ImplSDL3_UpdateMouseData()
	ImGui_ImplSDL3_UpdateMouseCursor()
	// Update game controllers (if enabled and available)
	ImGui_ImplSDL3_UpdateGamepads()
}

func ImGui_ImplOpenGL4_InitLoader() {
	// do we even need this?
}

type ImGui_ImplOpenGL4_Data = struct {
	GlVersion                 uint32   // Extracted at runtime using GL_MAJOR_VERSION, GL_MINOR_VERSION queries (e.g. 320 for GL 3.2)
	GlslVersionString         [32]byte // Specified by user or detected based on compile time GL settings.
	GlProfileIsES2            bool
	GlProfileIsES3            bool
	GlProfileIsCompat         bool
	GlProfileMask             int32
	MaxTextureSize            int32
	ShaderHandle              uint32
	AttribLocationTex         int32 // Uniforms location
	AttribLocationProjMtx     int32
	AttribLocationVtxPos      uint32 // Vertex attributes location
	AttribLocationVtxUV       uint32
	AttribLocationVtxColor    uint32
	VboHandle, ElementsHandle uint32
	VertexBufferSize          int
	IndexBufferSize           int
	HasPolygonMode            bool
	HasBindSampler            bool
	HasClipOrigin             bool
	UseBufferSubData          bool
	TempBuffer                []byte
}

func ImGui_ImplOpenGL4_GetBackendData() *ImGui_ImplOpenGL4_Data {
	if im.CurrentContext() != nil {
		return (*ImGui_ImplOpenGL4_Data)(unsafe.Pointer(im.CurrentIO().BackendRendererUserData()))
	}
	return nil
}

type VtxAttribState struct {
	Enabled, Size, Type, Normalized, Stride int32
	Ptr                                     unsafe.Pointer
}

func (this *VtxAttribState) GetState(index uint32) {
	gl.GetVertexAttribiv(index, gl.VERTEX_ATTRIB_ARRAY_ENABLED, &this.Enabled)
	gl.GetVertexAttribiv(index, gl.VERTEX_ATTRIB_ARRAY_SIZE, &this.Size)
	gl.GetVertexAttribiv(index, gl.VERTEX_ATTRIB_ARRAY_TYPE, &this.Type)
	gl.GetVertexAttribiv(index, gl.VERTEX_ATTRIB_ARRAY_NORMALIZED, &this.Normalized)
	gl.GetVertexAttribiv(index, gl.VERTEX_ATTRIB_ARRAY_STRIDE, &this.Stride)
	gl.GetVertexAttribPointerv(index, gl.VERTEX_ATTRIB_ARRAY_POINTER, &this.Ptr)
}

func (this *VtxAttribState) SetState(index uint32) {
	gl.VertexAttribPointer(index, this.Size, uint32(this.Type), this.Normalized != 0, this.Stride, this.Ptr)
	if this.Enabled != 0 {
		gl.EnableVertexAttribArray(index)
	} else {
		gl.DisableVertexAttribArray(index)
	}
}

func ImGui_ImplOpenGL4_UpdateTexture(tex *im.TextureData) {
	panic("not implemented")
}

func ImGui_ImplOpenGL4_SetupRenderState(drawData *im.DrawData, fb_width, fb_height int, vertex_array_object uint32) {
	panic("not implemented")
}

func ImGui_ImplOpenGL4_RenderDrawData(drawData *im.DrawData) {
	// Avoid rendering when minimized, scale coordinates for retina displays (screen coordinates != framebuffer coordinates)
	var fb_width int = (int)(drawData.DisplaySize().X * drawData.FramebufferScale().X)
	var fb_height int = (int)(drawData.DisplaySize().Y * drawData.FramebufferScale().Y)
	if fb_width <= 0 || fb_height <= 0 {
		return
	}

	ImGui_ImplOpenGL4_InitLoader() // Lazily init loader if not already done for e.g. DLL boundaries.
	bd := ImGui_ImplOpenGL4_GetBackendData()

	// Catch up with texture updates. Most of the times, the list will have 1 element with an OK status, aka nothing to do.
	// (This almost always points to ImGui::GetPlatformIO().Textures[] but is part of ImDrawData to allow overriding or disabling texture updates).

	//im.CurrentIO().CData.Texture

	if drawData.Textures != nil {
		for /* ImTextureData* */ _, tex := range drawData.Textures {
			if tex.Status != im.TextureStatus_OK {
				ImGui_ImplOpenGL4_UpdateTexture(tex)
			}
		}
	}
	// Backup GL state
	var last_active_texture int32
	gl.GetIntegerv(gl.ACTIVE_TEXTURE, &last_active_texture)
	gl.ActiveTexture(gl.TEXTURE0)
	var last_program int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &last_program)
	var last_texture int32
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &last_texture)
	var last_sampler int32
	gl.GetIntegerv(gl.SAMPLER_BINDING, &last_sampler)

	var last_array_buffer int32
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, &last_array_buffer)

	//#ifndef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	// This is part of VAO on OpenGL 3.0+ and OpenGL ES 3.0+.
	var last_element_array_buffer int32
	gl.GetIntegerv(gl.ELEMENT_ARRAY_BUFFER_BINDING, &last_element_array_buffer)
	var last_vtx_attrib_state_pos VtxAttribState
	last_vtx_attrib_state_pos.GetState(bd.AttribLocationVtxPos)
	var last_vtx_attrib_state_uv VtxAttribState
	last_vtx_attrib_state_uv.GetState(bd.AttribLocationVtxUV)
	var last_vtx_attrib_state_color VtxAttribState
	last_vtx_attrib_state_color.GetState(bd.AttribLocationVtxColor)
	//#endif

	//#ifdef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	var last_vertex_array_object int32
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &last_vertex_array_object)
	//#endif

	//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_POLYGON_MODE
	var last_polygon_mode [2]int32
	if bd.HasPolygonMode {
		gl.GetIntegerv(gl.POLYGON_MODE, &last_polygon_mode[0])
	}
	// #endif

	var last_viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &last_viewport[0])
	var last_scissor_box [4]int32
	gl.GetIntegerv(gl.SCISSOR_BOX, &last_scissor_box[0])
	var last_blend_src_rgb int32
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &last_blend_src_rgb)
	var last_blend_dst_rgb int32
	gl.GetIntegerv(gl.BLEND_DST_RGB, &last_blend_dst_rgb)
	var last_blend_src_alpha int32
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &last_blend_src_alpha)
	var last_blend_dst_alpha int32
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &last_blend_dst_alpha)
	var last_blend_equation_rgb int32
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &last_blend_equation_rgb)
	var last_blend_equation_alpha int32
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &last_blend_equation_alpha)
	var last_enable_blend bool = gl.IsEnabled(gl.BLEND)
	var last_enable_cull_face bool = gl.IsEnabled(gl.CULL_FACE)
	var last_enable_depth_test bool = gl.IsEnabled(gl.DEPTH_TEST)
	var last_enable_stencil_test bool = gl.IsEnabled(gl.STENCIL_TEST)
	var last_enable_scissor_test bool = gl.IsEnabled(gl.SCISSOR_TEST)
	//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_PRIMITIVE_RESTART
	var last_enable_primitive_restart bool = gl.IsEnabled(gl.PRIMITIVE_RESTART)
	//#endif

	// Setup desired GL state
	// Recreate the VAO every time (this is to easily allow multiple GL contexts to be rendered to. VAO are not shared among GL contexts)
	// The renderer would actually work without any VAO bound, but then our VertexAttrib calls would overwrite the default one currently bound.
	var vertex_array_object uint32 = 0
	//#ifdef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	gl.GenVertexArrays(1, &vertex_array_object)
	//#endif
	ImGui_ImplOpenGL4_SetupRenderState(drawData, fb_width, fb_height, vertex_array_object)

	// Will project scissor/clipping rectangles into framebuffer space
	var clip_off im.Vec2 = drawData.DisplayPos()         // (0,0) unless using multi-viewports
	var clip_scale im.Vec2 = drawData.FramebufferScale() // (1,1) unless using retina display which are often (2,2)

	// Render command lists
	for n := (int32)(0); n < drawData.CmdListsCount(); n++ {
		// drawData.CmdListsCount

		//drawDrawData.CmdLists().

		var draw_list *im.DrawList = drawData.CData.CmdLists[n]

		// Upload vertex/index buffers
		// - OpenGL drivers are in a very sorry state nowadays....
		//   During 2021 we attempted to switch from glBufferData() to orphaning+glBufferSubData() following reports
		//   of leaks on Intel GPU when using multi-viewports on Windows.
		// - After this we kept hearing of various display corruptions issues. We started disabling on non-Intel GPU, but issues still got reported on Intel.
		// - We are now back to using exclusively glBufferData(). So bd.UseBufferSubData IS ALWAYS FALSE in this code.
		//   We are keeping the old code path for a while in case people finding new issues may want to test the bd.UseBufferSubData path.
		// - See https://github.com/ocornut/imgui/issues/4468 and please report any corruption issues.

		var vtx_buffer_size int = (int)(draw_list.VtxBuffer().Size) * sizeof[im.DrawVert]()
		var idx_buffer_size int = (int)(draw_list.IdxBuffer().Size) * sizeof[im.DrawIdx]()

		if bd.UseBufferSubData {
			if bd.VertexBufferSize < vtx_buffer_size {
				bd.VertexBufferSize = vtx_buffer_size
				gl.BufferData(gl.ARRAY_BUFFER, bd.VertexBufferSize, nil, gl.STREAM_DRAW)
			}
			if bd.IndexBufferSize < idx_buffer_size {
				bd.IndexBufferSize = idx_buffer_size
				gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, bd.IndexBufferSize, nil, gl.STREAM_DRAW)
			}
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, vtx_buffer_size, (unsafe.Pointer)(draw_list.VtxBuffer().Data))
			gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, idx_buffer_size, (unsafe.Pointer)(draw_list.IdxBuffer().Data))
		} else {
			gl.BufferData(gl.ARRAY_BUFFER, vtx_buffer_size, (unsafe.Pointer)(draw_list.VtxBuffer().Data), gl.STREAM_DRAW)
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, idx_buffer_size, (unsafe.Pointer)(draw_list.IdxBuffer().Data), gl.STREAM_DRAW)
		}

		for cmd_i := 0; cmd_i < draw_list.CmdBuffer().Size; cmd_i++ {
			var pcmd = &draw_list.CmdBuffer().Slice()[cmd_i]
			if pcmd.UserCallback != nil {
				// User callback, registered via ImDrawList::AddCallback()
				// (ImDrawCallback_ResetRenderState is a special callback value used by the user to request the renderer to reset render state.)
				if pcmd.UserCallback == im.DrawCallback_ResetRenderState {
					ImGui_ImplOpenGL4_SetupRenderState(drawData, fb_width, fb_height, vertex_array_object)
				} else {
					//pcmd.UserCallback(draw_list, pcmd)
					pcmd.UserCallback()
				}
			} else {
				// Project scissor/clipping rectangles into framebuffer space
				clip_min := im.NewVec2((pcmd.ClipRect().X-clip_off.X)*clip_scale.X, (pcmd.ClipRect().Y-clip_off.Y)*clip_scale.Y)
				clip_max := im.NewVec2((pcmd.ClipRect().Z-clip_off.X)*clip_scale.X, (pcmd.ClipRect().W-clip_off.Y)*clip_scale.Y)
				if clip_max.X <= clip_min.X || clip_max.Y <= clip_min.Y {
					continue
				}
				// Apply scissor/clipping rectangle (Y is inverted in OpenGL)
				gl.Scissor((int32)(clip_min.X), (int32)((float32)(fb_height)-clip_max.Y), (int32)(clip_max.X-clip_min.X), (int32)(clip_max.Y-clip_min.Y))

				// Bind texture, Draw
				gl.BindTexture(gl.TEXTURE_2D, (uint32)(pcmd.TexID()))
				//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_VTX_OFFSET
				//if (bd.GlVersion >= 320)

				var typeArg uint32
				if sizeof[im.DrawIdx]() == 2 {
					typeArg = gl.UNSIGNED_SHORT
				} else {
					typeArg = gl.UNSIGNED_INT
				}

				gl.DrawElementsBaseVertex(gl.TRIANGLES, int32(pcmd.ElemCount()), typeArg, (unsafe.Pointer)(uintptr(pcmd.IdxOffset())*uintptr(sizeof[im.DrawIdx]())), int32(pcmd.VtxOffset()))
				//else
				//#endif
				//gl.DrawElements(gl._TRIANGLES, (gl.sizei)(pcmd.ElemCount), typeArg, (void*)(intptr_t)(pcmd.IdxOffset * sizeof(ImDrawIdx)));
			}
		}
	}

	// Destroy the temporary VAO
	//#ifdef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	gl.DeleteVertexArrays(1, &vertex_array_object)
	//#endif

	// Restore modified GL state
	// This "glIsProgram()" check is required because if the program is "pending deletion" at the time of binding backup, it will have been deleted by now and will cause an OpenGL error. See #6220.
	if last_program == 0 || gl.IsProgram(uint32(last_program)) {
		gl.UseProgram(uint32(last_program))
	}
	gl.BindTexture(gl.TEXTURE_2D, uint32(last_texture))
	//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_BIND_SAMPLER
	if bd.GlVersion >= 330 || bd.GlProfileIsES3 {
		gl.BindSampler(0, uint32(last_sampler))
	}
	//#endif
	gl.ActiveTexture(uint32(last_active_texture))
	//#ifdef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	gl.BindVertexArray(uint32(last_vertex_array_object))
	//#endif
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(last_array_buffer))
	//#ifndef IMGUI_IMPL_OPENGL_USE_VERTEX_ARRAY
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, uint32(last_element_array_buffer))
	last_vtx_attrib_state_pos.SetState(bd.AttribLocationVtxPos)
	last_vtx_attrib_state_uv.SetState(bd.AttribLocationVtxUV)
	last_vtx_attrib_state_color.SetState(bd.AttribLocationVtxColor)
	//#endif
	gl.BlendEquationSeparate(uint32(last_blend_equation_rgb), uint32(last_blend_equation_alpha))
	gl.BlendFuncSeparate(uint32(last_blend_src_rgb), uint32(last_blend_dst_rgb), uint32(last_blend_src_alpha), uint32(last_blend_dst_alpha))
	if last_enable_blend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
	if last_enable_cull_face {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if last_enable_depth_test {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if last_enable_stencil_test {
		gl.Enable(gl.STENCIL_TEST)
	} else {
		gl.Disable(gl.STENCIL_TEST)
	}
	if last_enable_scissor_test {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}
	//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_PRIMITIVE_RESTART
	//if (!bd.GlProfileIsES3 && bd.GlVersion >= 310) {
	if last_enable_primitive_restart {
		gl.Enable(gl.PRIMITIVE_RESTART)
	} else {
		gl.Disable(gl.PRIMITIVE_RESTART)
	}
	//}
	//#endif

	//#ifdef IMGUI_IMPL_OPENGL_MAY_HAVE_POLYGON_MODE
	// Desktop OpenGL 3.0 and OpenGL 3.1 had separate polygon draw modes for front-facing and back-facing faces of polygons
	if bd.HasPolygonMode {
		if bd.GlVersion <= 310 || bd.GlProfileIsCompat {
			gl.PolygonMode(gl.FRONT, uint32(last_polygon_mode[0]))
			gl.PolygonMode(gl.BACK, uint32(last_polygon_mode[1]))
		} else {
			gl.PolygonMode(gl.FRONT_AND_BACK, uint32(last_polygon_mode[0]))
		}
	}
	//#endif // IMGUI_IMPL_OPENGL_MAY_HAVE_POLYGON_MODE

	gl.Viewport(last_viewport[0], last_viewport[1], last_viewport[2], last_viewport[3])
	gl.Scissor(last_scissor_box[0], last_scissor_box[1], last_scissor_box[2], last_scissor_box[3])
	//(void)bd; // Not all compilation paths use this
}

func ImGui_ImplOpenGL4_Shutdown() {
	panic("not implemented")
}

func ImGui_ImplSDL3_Shutdown() {
	panic("not implemented")
}

func ImGui_ImplSDL3_UpdateMouseData() {
	panic("not implemented")
}

func ImGui_ImplSDL3_UpdateMouseCursor() {
	panic("not implemented")
}

// Update game controllers (if enabled and available)
func ImGui_ImplSDL3_UpdateGamepads() {
	panic("not implemented")
}

var ViewportsEnable bool = false

func igUpdatePlatformWindows() {
	panic("not implemented")
}

func igRenderPlatformWindowsDefault(x, y int32) {
	panic("not implemented")
}

func main() {
	// ImGuiIO* io = igGetIO_Nil();
	//
	Must0(gl.Init())

	defer binsdl.Load().Unload() // sdl.LoadLibrary(sdl.Path())

	sdl.Init(sdl.INIT_AUDIO | sdl.INIT_JOYSTICK | sdl.INIT_VIDEO)
	defer sdl.Quit()

	sdl.GL_SetAttribute(sdl.GL_CONTEXT_FLAGS, 0)
	sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GL_SetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 6)

	window, err := sdl.CreateWindow("Hello world", 500, 500, sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	gl_context, err := sdl.GL_CreateContext(window)
	if err != nil {
		panic(err)
	}
	defer sdl.GL_DestroyContext(gl_context)
	sdl.GL_MakeCurrent(window, gl_context)

	windowID := Must(window.ID())

	im.SetCurrentContext(im.CreateContext())

	io := im.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() | im.ConfigFlagsNavEnableKeyboard | im.ConfigFlagsNavEnableGamepad)

	im.StyleColorsDark()

	// im.WindowViewport()

	// Main loop
	var done = false
	for !done {
		// Poll and handle events (inputs, window resize, etc.)
		// You can read the io.WantCaptureMouse, io.WantCaptureKeyboard flags to tell if dear im wants to use your inputs.
		// - When io.WantCaptureMouse is true, do not dispatch mouse input data to your main application, or clear/overwrite your copy of the mouse data.
		// - When io.WantCaptureKeyboard is true, do not dispatch keyboard input data to your main application, or clear/overwrite your copy of the keyboard data.
		// Generally you may always pass all inputs to dear imgui, and hide them from your application based on those two flags.
		var event sdl.Event
		for sdl.PollEvent(&event) {
			ImGui_ImplSDL3_ProcessEvent(&event)

			if event.Type == sdl.EVENT_QUIT {
				done = true
			}
			if event.Type == sdl.EVENT_WINDOW_CLOSE_REQUESTED && event.WindowEvent().WindowID == windowID {
				done = true
			}
		}

		// Start the Dear ImGui frame

		sizex, sizey := Must2(window.SizeInPixels())

		gl.Viewport(0, 0, sizex, sizey)

		// glClearColor(sdl_clear_color.x * sdl_clear_color.w, sdl_clear_color.y * sdl_clear_color.w, sdl_clear_color.z * sdl_clear_color.w, sdl_clear_color.w);
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		ImGui_ImplOpenGL4_NewFrame()
		ImGui_ImplSDL3_NewFrame()

		im.NewFrame()
		// rendering
		Loop()

		// Rendering
		im.Render()

		drawData := im.CurrentDrawData()

		ImGui_ImplOpenGL4_RenderDrawData(drawData)

		// Update and Render additional Platform Windows
		// (Platform functions may change the current OpenGL context, so we save/restore it to make it easier to paste this code elsewhere.
		//  For this specific demo app we could also call SDL_GL_MakeCurrent(window, gl_context) directly)
		if ViewportsEnable {
			igUpdatePlatformWindows()
			igRenderPlatformWindowsDefault(0, 0)
			sdl.GL_MakeCurrent(window, gl_context)
		}

		sdl.GL_SwapWindow(window)
	}

	// Cleanup
	ImGui_ImplOpenGL4_Shutdown()
	ImGui_ImplSDL3_Shutdown()

}

func main2() {
	Initialize()

	currentBackend, _ = backend.CreateBackend(sdlbackend.NewSDLBackend())

	currentBackend.SetAfterCreateContextHook(AfterCreateContext)
	currentBackend.SetBeforeDestroyContextHook(BeforeDestroyContext)

	currentBackend.SetBgColor(im.NewVec4(0.45, 0.55, 0.6, 1.0))
	currentBackend.CreateWindow("Hello from cimgui-go", 1200, 900)

	currentBackend.SetDropCallback(func(p []string) {
		fmt.Printf("drop triggered: %v", p)
	})

	currentBackend.SetCloseCallback(func() {
		fmt.Println("window is closing")
	})

	fmt.Println("running the loop")

	// currentBackend.SetIcons(Image())

	currentBackend.Run(Loop)
}

var (
	showDemoWindow bool
	value1         int32
	value2         int32
	value3         int32
	values         [2]int32 = [2]int32{value1, value2}
	content        string   = "Let me try"
	r              float32
	g              float32
	b              float32
	a              float32
	color4         [4]float32 = [4]float32{r, g, b, a}
	selected       bool
	//go:embed test.jpeg
	imgData    []byte
	img        *image.RGBA
	texture    *backend.Texture
	barValues  []int64
	textEditor *cte.TextEditor
)

// Initialize prepares global variables. Call before anything related to backend.
func Initialize() {
	imgImage, _, _ := image.Decode(bytes.NewReader(imgData))
	img = backend.ImageToRgba(imgImage)

	for i := 0; i < 10; i++ {
		barValues = append(barValues, int64(i+1))
	}
}

func InputTextCallback(data im.InputTextCallbackData) int {
	fmt.Println("got call back")
	return 0
}

func AfterCreateContext() {
	texture = backend.NewTextureFromRgba(img)
	implot.CreateContext()
	textEditor = cte.NewTextEditor()
	textEditor.SetLanguageDefinition(cte.Cpp)
	textEditor.SetText(`// Colorize a C++ file
#include <iostream>
ImGui::Text("Hello World")`)
}

func BeforeDestroyContext() {
	implot.DestroyContext()
}

func Loop() {
	im.ClearSizeCallbackPool()
	imguizmo.BeginFrame()
	ShowWidgetsDemo()
	ShowPictureLoadingDemo()
	ShowImPlotDemo()
	ShowGizmoDemo()
	ShowCTEDemo()
}

func ShowWidgetsDemo() {
	if showDemoWindow {
		im.ShowDemoWindowV(&showDemoWindow)
	}

	im.SetNextWindowSizeV(im.NewVec2(300, 300), im.CondOnce)

	im.SetNextWindowSizeConstraintsV(im.NewVec2(300, 300), im.NewVec2(500, 500), func(data *im.SizeCallbackData) {
	}, 0)

	im.Begin("Window 1")
	if im.ButtonV("Click Me", im.NewVec2(80, 20)) {
		fmt.Println("Button clicked")
	}
	im.TextUnformatted("Unformatted text")
	im.Checkbox("Show demo window", &showDemoWindow)
	if im.BeginCombo("Combo", "Combo preview") {
		im.SelectableBoolPtr("Item 1", &selected)
		im.SelectableBool("Item 2")
		im.SelectableBool("Item 3")
		im.EndCombo()
	}

	if im.RadioButtonBool("Radio button1", selected) {
		selected = true
	}

	im.SameLine()

	if im.RadioButtonBool("Radio button2", !selected) {
		selected = false
	}

	im.InputTextWithHint("Name", "write your name here", &content, 0, InputTextCallback)
	im.Text(content)
	im.SliderInt("Slider int", &value3, 0, 100)
	im.DragInt("Drag int", &value1)
	im.DragInt2("Drag int2", &values)
	value1 = values[0]
	im.ColorEdit4("Color Edit3", &color4)
	im.End()
}

func ShowPictureLoadingDemo() {
	// demo of showing a picture
	basePos := im.MainViewport().Pos()
	im.SetNextWindowPosV(im.NewVec2(basePos.X+60, 600), im.CondOnce, im.NewVec2(0, 0))
	im.Begin("Image")
	im.Text(fmt.Sprintf("pointer = %v", texture.ID))
	im.ImageWithBgV(texture.ID, im.NewVec2(float32(texture.Width), float32(texture.Height)), im.NewVec2(0, 0), im.NewVec2(1, 1), im.NewVec4(0, 0, 0, 0), im.NewVec4(1, 1, 1, 1))
	im.End()
}

func ShowImPlotDemo() {
	basePos := im.MainViewport().Pos()
	im.SetNextWindowPosV(im.NewVec2(basePos.X+400, basePos.Y+60), im.CondOnce, im.NewVec2(0, 0))
	im.SetNextWindowSizeV(im.NewVec2(500, 300), im.CondOnce)
	im.Begin("Plot window")
	if implot.BeginPlotV("Plot", im.NewVec2(-1, -1), 0) {
		implot.PlotBarsS64PtrInt("Bar", utils.SliceToPtr(barValues), int32(len(barValues)))
		implot.PlotLineS64PtrInt("Line", utils.SliceToPtr(barValues), int32(len(barValues)))
		implot.EndPlot()
	}
	im.End()
}

func ShowCTEDemo() {
	basePos := im.MainViewport().Pos()
	im.SetNextWindowPosV(im.NewVec2(basePos.X+800, basePos.Y+260), im.CondOnce, im.NewVec2(0, 0))
	im.SetNextWindowSizeV(im.NewVec2(250, 400), im.CondOnce)
	im.Begin("Color Text Edit")

	if textEditor.Render("Color Text Edit") {
	}

	im.End()
}

func Image() *image.RGBA {
	return img
}
