package main

import (
	"fmt"
	"runtime"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	"github.com/AllenDang/cimgui-go/imgui"

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

func ImGui_ImplSDL3_NewFrame() {
	panic("not implemented")
}

func ImGui_ImplOpenGL4_RenderDrawData(drawData *imgui.DrawData) {
	panic("not implemented")
}

func ImGui_ImplOpenGL4_Shutdown() {
	panic("not implemented")
}

func ImGui_ImplSDL3_Shutdown() {
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

	imgui.SetCurrentContext(imgui.CreateContext())

	io := imgui.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() | imgui.ConfigFlagsNavEnableKeyboard | imgui.ConfigFlagsNavEnableGamepad)

	imgui.StyleColorsDark()

	// imgui.WindowViewport()

	// Main loop
	var done = false
	for !done {
		// Poll and handle events (inputs, window resize, etc.)
		// You can read the io.WantCaptureMouse, io.WantCaptureKeyboard flags to tell if dear imgui wants to use your inputs.
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

		imgui.NewFrame()
		// rendering
		Loop()

		// Rendering
		imgui.Render()

		drawData := imgui.CurrentDrawData()

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

	currentBackend.SetBgColor(imgui.NewVec4(0.45, 0.55, 0.6, 1.0))
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

func InputTextCallback(data imgui.InputTextCallbackData) int {
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
	imgui.ClearSizeCallbackPool()
	imguizmo.BeginFrame()
	ShowWidgetsDemo()
	ShowPictureLoadingDemo()
	ShowImPlotDemo()
	ShowGizmoDemo()
	ShowCTEDemo()
}

func ShowWidgetsDemo() {
	if showDemoWindow {
		imgui.ShowDemoWindowV(&showDemoWindow)
	}

	imgui.SetNextWindowSizeV(imgui.NewVec2(300, 300), imgui.CondOnce)

	imgui.SetNextWindowSizeConstraintsV(imgui.Vec2{300, 300}, imgui.Vec2{500, 500}, func(data *imgui.SizeCallbackData) {
	}, 0)

	imgui.Begin("Window 1")
	if imgui.ButtonV("Click Me", imgui.NewVec2(80, 20)) {
		fmt.Println("Button clicked")
	}
	imgui.TextUnformatted("Unformatted text")
	imgui.Checkbox("Show demo window", &showDemoWindow)
	if imgui.BeginCombo("Combo", "Combo preview") {
		imgui.SelectableBoolPtr("Item 1", &selected)
		imgui.SelectableBool("Item 2")
		imgui.SelectableBool("Item 3")
		imgui.EndCombo()
	}

	if imgui.RadioButtonBool("Radio button1", selected) {
		selected = true
	}

	imgui.SameLine()

	if imgui.RadioButtonBool("Radio button2", !selected) {
		selected = false
	}

	imgui.InputTextWithHint("Name", "write your name here", &content, 0, InputTextCallback)
	imgui.Text(content)
	imgui.SliderInt("Slider int", &value3, 0, 100)
	imgui.DragInt("Drag int", &value1)
	imgui.DragInt2("Drag int2", &values)
	value1 = values[0]
	imgui.ColorEdit4("Color Edit3", &color4)
	imgui.End()
}

func ShowPictureLoadingDemo() {
	// demo of showing a picture
	basePos := imgui.MainViewport().Pos()
	imgui.SetNextWindowPosV(imgui.NewVec2(basePos.X+60, 600), imgui.CondOnce, imgui.NewVec2(0, 0))
	imgui.Begin("Image")
	imgui.Text(fmt.Sprintf("pointer = %v", texture.ID))
	imgui.ImageWithBgV(texture.ID, imgui.NewVec2(float32(texture.Width), float32(texture.Height)), imgui.NewVec2(0, 0), imgui.NewVec2(1, 1), imgui.NewVec4(0, 0, 0, 0), imgui.NewVec4(1, 1, 1, 1))
	imgui.End()
}

func ShowImPlotDemo() {
	basePos := imgui.MainViewport().Pos()
	imgui.SetNextWindowPosV(imgui.NewVec2(basePos.X+400, basePos.Y+60), imgui.CondOnce, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(500, 300), imgui.CondOnce)
	imgui.Begin("Plot window")
	if implot.BeginPlotV("Plot", imgui.NewVec2(-1, -1), 0) {
		implot.PlotBarsS64PtrInt("Bar", utils.SliceToPtr(barValues), int32(len(barValues)))
		implot.PlotLineS64PtrInt("Line", utils.SliceToPtr(barValues), int32(len(barValues)))
		implot.EndPlot()
	}
	imgui.End()
}

func ShowCTEDemo() {
	basePos := imgui.MainViewport().Pos()
	imgui.SetNextWindowPosV(imgui.NewVec2(basePos.X+800, basePos.Y+260), imgui.CondOnce, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(250, 400), imgui.CondOnce)
	imgui.Begin("Color Text Edit")

	if textEditor.Render("Color Text Edit") {
	}

	imgui.End()
}

func Image() *image.RGBA {
	return img
}
