package main

import (
	"fmt"
	"math"
	"runtime"

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

	out_size = im.Vec2{float32(w), float32(h)}
	if w > 0 && h > 0 {
		out_framebuffer_scale = im.Vec2{float32(display_w) / float32(w), float32(display_h) / float32(h)}
	} else {
		out_framebuffer_scale = im.Vec2{1, 1}
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

	bd.WantUpdateMonitors = false

	displays := Must(sdl.GetDisplays())

	// int display_count;
	// SDL_DisplayID* displays = SDL_GetDisplays(&display_count);
	for n, display_id := range displays {
		// Warning: the validity of monitor DPI information on Windows depends on the application DPI awareness settings, which generally needs to be set in the manifest or at runtime.
		var monitor im.PlatformMonitor
		var r *sdl.Rect = Must(display_id.Bounds())

		pos := im.Vec2{float32(r.X), float32(r.Y)}
		monitor.SetMainPos(pos)
		monitor.SetWorkPos(pos)
		size := im.Vec2{float32(r.W), float32(r.H)}
		monitor.SetMainSize(size)
		monitor.SetWorkSize(size)

		r = Must(display_id.UsableBounds())
		if r.W > 0 && r.H > 0 {
			monitor.SetWorkPos(im.Vec2{float32(r.X), float32(r.Y)})
			monitor.SetWorkSize(im.Vec2{float32(r.W), float32(r.H)})
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

func ImGui_ImplOpenGL4_RenderDrawData(drawData *im.DrawData) {
	panic("not implemented")
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

// ImGui_ImplOpenGL3_RenderDrawData(igGetDrawData())
// func ImGui_Impl

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

	im.SetNextWindowSizeConstraintsV(im.Vec2{300, 300}, im.Vec2{500, 500}, func(data *im.SizeCallbackData) {
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
