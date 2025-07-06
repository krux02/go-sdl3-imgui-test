package main

import (
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/Zyko0/go-sdl3/sdl"
	"slices"
	"unsafe"
)

type ImGui_ImplSDL3_Data struct {
	MouseButtonsDown              uint32
	MouseWindowID                 sdl.WindowID
	BackendPlatformName           string
	MouseLastLeaveFrame           int32
	Window                        *sdl.Window
	WindowID                      sdl.WindowID
	Renderer                      *sdl.Renderer
	MouseCanUseGlobalState        bool
	MouseCanUseCapture            bool
	WantUpdateGamepadsList        bool
	WantUpdateMonitors            bool
	MouseCanReportHoveredViewport bool
	MousePendingLeaveFrame        int32
	MouseCursors                  [11]*sdl.Cursor
	Time                          uint64
}

func ImGui_ImplSDL3_GetBackendData() *ImGui_ImplSDL3_Data {
	if imgui.CurrentContext() != nil {
		return (*ImGui_ImplSDL3_Data)(unsafe.Pointer(imgui.CurrentIO().BackendPlatformUserData()))
	}
	return nil
}

func ImGui_ImplSDL3_GetViewportForWindowID(windowID sdl.WindowID) *imgui.Viewport {
	return imgui.FindViewportByPlatformHandle(uintptr(windowID))
}

func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func Must2[T1 any, T2 any](t1 T1, t2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return t1, t2
}

// for some reason these constants are missing in the sdl3 wrapper
// const (
// 	SDL_GL_CONTEXT_PROFILE_CORE          = 0x0001 /**< OpenGL Core Profile context */
// 	SDL_GL_CONTEXT_PROFILE_COMPATIBILITY = 0x0002 /**< OpenGL Compatibility Profile context */
// 	SDL_GL_CONTEXT_PROFILE_ES            = 0x0004 /**< GLX_CONTEXT_ES2_PROFILE_BIT_EXT */
// )

func ImGui_ImplSDL3_ProcessEvent(event *sdl.Event) bool {
	bd := ImGui_ImplSDL3_GetBackendData()
	//IM_ASSERT(bd != nullptr && "Context or backend not initialized! Did you call ImGui_ImplSDL2_Init()?");
	//ImGuiIO& io = ImGui::GetIO();
	io := imgui.CurrentIO()

	switch event.Type {
	case sdl.EVENT_MOUSE_MOTION:
		motionEvent := event.MouseMotionEvent()
		if ImGui_ImplSDL3_GetViewportForWindowID(motionEvent.WindowID) == nil {
			return false
		}

		mouse_pos := imgui.Vec2{X: motionEvent.X, Y: motionEvent.Y}

		if (io.ConfigFlags() & imgui.ConfigFlagsViewportsEnable) != 0 {
			window_x, window_y := Must2(Must(motionEvent.WindowID.Window()).Position())
			mouse_pos.X += (float32)(window_x)
			mouse_pos.Y += (float32)(window_y)
		}

		// if motionEvent.Which == sdl.TOUCH_MOUSEID {
		// 	io.AddMouseSourceEvent(ImGuiMouseSource_TouchScreen)
		// }

		io.AddMouseSourceEvent(imgui.MouseSourceMouse)
		io.AddMousePosEvent(mouse_pos.X, mouse_pos.Y)
	case sdl.EVENT_MOUSE_WHEEL:
		wheelEvent := event.MouseWheelEvent()
		if ImGui_ImplSDL3_GetViewportForWindowID(wheelEvent.WindowID) == nil {
			return false
		}

		//IMGUI_DEBUG_LOG("wheel %.2f %.2f, precise %.2f %.2f\n", (float)wheelEvent.x, (float)wheelEvent.y, wheelEvent.preciseX, wheelEvent.preciseY);
		wheel_x := -wheelEvent.X
		wheel_y := wheelEvent.Y

		// if wheelEvent.Which == sdl.TOUCH_MOUSEID {
		// 	io.AddMouseSourceEvent(imgui.MouseSourceTouchScreen)
		// }
		io.AddMouseSourceEvent(imgui.MouseSourceMouse)
		io.AddMouseWheelEvent(wheel_x, wheel_y)
	case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
		buttonEvent := event.MouseButtonEvent()
		if ImGui_ImplSDL3_GetViewportForWindowID(buttonEvent.WindowID) == nil {
			return false
		}
		var mouse_button int32 = -1
		if buttonEvent.Button == uint8(sdl.BUTTON_LEFT) {
			mouse_button = 0
		}
		if buttonEvent.Button == uint8(sdl.BUTTON_RIGHT) {
			mouse_button = 1
		}
		if buttonEvent.Button == uint8(sdl.BUTTON_MIDDLE) {
			mouse_button = 2
		}
		if buttonEvent.Button == uint8(sdl.BUTTON_X1) {
			mouse_button = 3
		}
		if buttonEvent.Button == uint8(sdl.BUTTON_X2) {
			mouse_button = 4
		}
		if mouse_button == -1 {
			return false
		}

		// if buttonEvent.Which == sdl.TOUCH_MOUSEID {
		// 	io.AddMouseSourceEvent(ImGuiMouseSource_TouchScreen)
		// } else {
		io.AddMouseSourceEvent(imgui.MouseSourceMouse)

		io.AddMouseButtonEvent(mouse_button, (event.Type == sdl.EVENT_MOUSE_BUTTON_DOWN))

		if event.Type == sdl.EVENT_MOUSE_BUTTON_DOWN {
			bd.MouseButtonsDown = bd.MouseButtonsDown | (1 << mouse_button)
		} else {
			bd.MouseButtonsDown = bd.MouseButtonsDown & ^(1 << mouse_button)
		}
	case sdl.EVENT_TEXT_INPUT:
		textEvent := event.TextInputEvent()
		if ImGui_ImplSDL3_GetViewportForWindowID(textEvent.WindowID) == nil {
			return false
		}
		io.AddInputCharactersUTF8(textEvent.Text)
	case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
		keyEvent := event.KeyboardEvent()
		if ImGui_ImplSDL3_GetViewportForWindowID(keyEvent.WindowID) == nil {
			return false
		}

		// <ImGui_ImplSDL3_UpdateKeyModifiers>
		sdl_key_mods := keyEvent.Mod
		io.AddKeyEvent(imgui.ModCtrl, (sdl_key_mods&sdl.KMOD_CTRL) != 0)
		io.AddKeyEvent(imgui.ModShift, (sdl_key_mods&sdl.KMOD_SHIFT) != 0)
		io.AddKeyEvent(imgui.ModAlt, (sdl_key_mods&sdl.KMOD_ALT) != 0)
		io.AddKeyEvent(imgui.ModSuper, (sdl_key_mods&sdl.KMOD_GUI) != 0)
		// </ImGui_ImplSDL3_UpdateKeyModifiers>

		//IMGUI_DEBUG_LOG("SDL_KEY_%s : key=%d ('%s'), scancode=%d ('%s'), mod=%X\n",
		//    (event.type == SDL_KEYDOWN) ? "DOWN" : "UP  ", event.key.keysym.sym, SDL_GetKeyName(event.key.keysym.sym), event.key.keysym.scancode, SDL_GetScancodeName(event.key.keysym.scancode), event.key.keysym.mod);
		var key imgui.Key = ImGui_ImplSDL3_KeyEventToImGuiKey(keyEvent.Key, keyEvent.Scancode)
		io.AddKeyEvent(key, event.Type == sdl.EVENT_KEY_DOWN)
		io.SetKeyEventNativeData(key, (int32)(keyEvent.Key), (int32)(keyEvent.Scancode)) // To support legacy indexing (<1.87 user code). Legacy backend uses SDLK_*** as indices to IsKeyXXX() functions.
	case sdl.EVENT_WINDOW_FOCUS_GAINED:
		io.AddFocusEvent(true)
	case sdl.EVENT_WINDOW_FOCUS_LOST:
		io.AddFocusEvent(false)
	case sdl.EVENT_WINDOW_RESIZED:
		viewport := ImGui_ImplSDL3_GetViewportForWindowID(event.WindowEvent().WindowID)
		if viewport == nil {
			return false
		}
		viewport.SetPlatformRequestResize(true)
	case sdl.EVENT_WINDOW_CLOSE_REQUESTED:
		viewport := ImGui_ImplSDL3_GetViewportForWindowID(event.WindowEvent().WindowID)
		if viewport.CData == nil {
			return false
		}
		viewport.SetPlatformRequestClose(true)
	case sdl.EVENT_WINDOW_MOVED:
		viewport := ImGui_ImplSDL3_GetViewportForWindowID(event.WindowEvent().WindowID)
		if viewport.CData == nil {
			return false
		}
		viewport.SetPlatformRequestMove(true)

		// - When capturing mouse, SDL will send a bunch of conflicting LEAVE/ENTER event on every mouse move, but the final ENTER tends to be right.
		// - However we won't get a correct LEAVE event for a captured window.
		// - In some cases, when detaching a window from main viewport SDL may send SDL_WINDOWEVENT_ENTER one frame too late,
		//   causing SDL_WINDOWEVENT_LEAVE on previous frame to interrupt drag operation by clear mouse position. This is why
		//   we delay process the SDL_WINDOWEVENT_LEAVE events by one frame. See issue #5012 for details.
	case sdl.EVENT_WINDOW_MOUSE_ENTER:
		bd.MouseWindowID = event.WindowEvent().WindowID
		bd.MouseLastLeaveFrame = 0
	case sdl.EVENT_WINDOW_MOUSE_LEAVE:
		bd.MouseLastLeaveFrame = imgui.FrameCount() + 1
	case sdl.EVENT_GAMEPAD_ADDED, sdl.EVENT_GAMEPAD_REMOVED:
		bd.WantUpdateGamepadsList = true
	default:
		return false
	}
	return true
}

func ImGui_ImplSDL3_KeyEventToImGuiKey(keycode sdl.Keycode, scancode sdl.Scancode) imgui.Key {
	switch keycode {
	case sdl.K_TAB:
		return imgui.KeyTab
	case sdl.K_LEFT:
		return imgui.KeyLeftArrow
	case sdl.K_RIGHT:
		return imgui.KeyRightArrow
	case sdl.K_UP:
		return imgui.KeyUpArrow
	case sdl.K_DOWN:
		return imgui.KeyDownArrow
	case sdl.K_PAGEUP:
		return imgui.KeyPageUp
	case sdl.K_PAGEDOWN:
		return imgui.KeyPageDown
	case sdl.K_HOME:
		return imgui.KeyHome
	case sdl.K_END:
		return imgui.KeyEnd
	case sdl.K_INSERT:
		return imgui.KeyInsert
	case sdl.K_DELETE:
		return imgui.KeyDelete
	case sdl.K_BACKSPACE:
		return imgui.KeyBackspace
	case sdl.K_SPACE:
		return imgui.KeySpace
	case sdl.K_RETURN:
		return imgui.KeyEnter
	case sdl.K_ESCAPE:
		return imgui.KeyEscape
	//case sdl.K_QUOTE: return imgui.KeyApostrophe;
	case sdl.K_COMMA:
		return imgui.KeyComma
	//case sdl.K_MINUS: return imgui.KeyMinus;
	case sdl.K_PERIOD:
		return imgui.KeyPeriod
	//case sdl.K_SLASH: return imgui.KeySlash;
	case sdl.K_SEMICOLON:
		return imgui.KeySemicolon
	//case sdl.K_EQUALS: return imgui.KeyEqual;
	//case sdl.K_LEFTBRACKET: return imgui.KeyLeftBracket;
	//case sdl.K_BACKSLASH: return imgui.KeyBackslash;
	//case sdl.K_RIGHTBRACKET: return imgui.KeyRightBracket;
	//case sdl.K_BACKQUOTE: return imgui.KeyGraveAccent;
	case sdl.K_CAPSLOCK:
		return imgui.KeyCapsLock
	case sdl.K_SCROLLLOCK:
		return imgui.KeyScrollLock
	case sdl.K_NUMLOCKCLEAR:
		return imgui.KeyNumLock
	case sdl.K_PRINTSCREEN:
		return imgui.KeyPrintScreen
	case sdl.K_PAUSE:
		return imgui.KeyPause
	case sdl.K_KP_0:
		return imgui.KeyKeypad0
	case sdl.K_KP_1:
		return imgui.KeyKeypad1
	case sdl.K_KP_2:
		return imgui.KeyKeypad2
	case sdl.K_KP_3:
		return imgui.KeyKeypad3
	case sdl.K_KP_4:
		return imgui.KeyKeypad4
	case sdl.K_KP_5:
		return imgui.KeyKeypad5
	case sdl.K_KP_6:
		return imgui.KeyKeypad6
	case sdl.K_KP_7:
		return imgui.KeyKeypad7
	case sdl.K_KP_8:
		return imgui.KeyKeypad8
	case sdl.K_KP_9:
		return imgui.KeyKeypad9
	case sdl.K_KP_PERIOD:
		return imgui.KeyKeypadDecimal
	case sdl.K_KP_DIVIDE:
		return imgui.KeyKeypadDivide
	case sdl.K_KP_MULTIPLY:
		return imgui.KeyKeypadMultiply
	case sdl.K_KP_MINUS:
		return imgui.KeyKeypadSubtract
	case sdl.K_KP_PLUS:
		return imgui.KeyKeypadAdd
	case sdl.K_KP_ENTER:
		return imgui.KeyKeypadEnter
	case sdl.K_KP_EQUALS:
		return imgui.KeyKeypadEqual
	case sdl.K_LCTRL:
		return imgui.KeyLeftCtrl
	case sdl.K_LSHIFT:
		return imgui.KeyLeftShift
	case sdl.K_LALT:
		return imgui.KeyLeftAlt
	case sdl.K_LGUI:
		return imgui.KeyLeftSuper
	case sdl.K_RCTRL:
		return imgui.KeyRightCtrl
	case sdl.K_RSHIFT:
		return imgui.KeyRightShift
	case sdl.K_RALT:
		return imgui.KeyRightAlt
	case sdl.K_RGUI:
		return imgui.KeyRightSuper
	case sdl.K_APPLICATION:
		return imgui.KeyMenu
	case sdl.K_0:
		return imgui.Key0
	case sdl.K_1:
		return imgui.Key1
	case sdl.K_2:
		return imgui.Key2
	case sdl.K_3:
		return imgui.Key3
	case sdl.K_4:
		return imgui.Key4
	case sdl.K_5:
		return imgui.Key5
	case sdl.K_6:
		return imgui.Key6
	case sdl.K_7:
		return imgui.Key7
	case sdl.K_8:
		return imgui.Key8
	case sdl.K_9:
		return imgui.Key9
	// case sdl.K_a:
	// 	return imgui.KeyA
	// case sdl.K_b:
	// 	return imgui.KeyB
	// case sdl.K_c:
	// 	return imgui.KeyC
	// case sdl.K_d:
	// 	return imgui.KeyD
	// case sdl.K_e:
	// 	return imgui.KeyE
	// case sdl.K_f:
	// 	return imgui.KeyF
	// case sdl.K_g:
	// 	return imgui.KeyG
	// case sdl.K_h:
	// 	return imgui.KeyH
	// case sdl.K_i:
	// 	return imgui.KeyI
	// case sdl.K_j:
	// 	return imgui.KeyJ
	// case sdl.K_k:
	// 	return imgui.KeyK
	// case sdl.K_l:
	// 	return imgui.KeyL
	// case sdl.K_m:
	// 	return imgui.KeyM
	// case sdl.K_n:
	// 	return imgui.KeyN
	// case sdl.K_o:
	// 	return imgui.KeyO
	// case sdl.K_p:
	// 	return imgui.KeyP
	// case sdl.K_q:
	// 	return imgui.KeyQ
	// case sdl.K_r:
	// 	return imgui.KeyR
	// case sdl.K_s:
	// 	return imgui.KeyS
	// case sdl.K_t:
	// 	return imgui.KeyT
	// case sdl.K_u:
	// 	return imgui.KeyU
	// case sdl.K_v:
	// 	return imgui.KeyV
	// case sdl.K_w:
	// 	return imgui.KeyW
	// case sdl.K_x:
	// 	return imgui.KeyX
	// case sdl.K_y:
	//  return imgui.KeyY
	// case sdl.K_z:
	// 	return imgui.KeyZ
	case sdl.K_F1:
		return imgui.KeyF1
	case sdl.K_F2:
		return imgui.KeyF2
	case sdl.K_F3:
		return imgui.KeyF3
	case sdl.K_F4:
		return imgui.KeyF4
	case sdl.K_F5:
		return imgui.KeyF5
	case sdl.K_F6:
		return imgui.KeyF6
	case sdl.K_F7:
		return imgui.KeyF7
	case sdl.K_F8:
		return imgui.KeyF8
	case sdl.K_F9:
		return imgui.KeyF9
	case sdl.K_F10:
		return imgui.KeyF10
	case sdl.K_F11:
		return imgui.KeyF11
	case sdl.K_F12:
		return imgui.KeyF12
	case sdl.K_F13:
		return imgui.KeyF13
	case sdl.K_F14:
		return imgui.KeyF14
	case sdl.K_F15:
		return imgui.KeyF15
	case sdl.K_F16:
		return imgui.KeyF16
	case sdl.K_F17:
		return imgui.KeyF17
	case sdl.K_F18:
		return imgui.KeyF18
	case sdl.K_F19:
		return imgui.KeyF19
	case sdl.K_F20:
		return imgui.KeyF20
	case sdl.K_F21:
		return imgui.KeyF21
	case sdl.K_F22:
		return imgui.KeyF22
	case sdl.K_F23:
		return imgui.KeyF23
	case sdl.K_F24:
		return imgui.KeyF24
	case sdl.K_AC_BACK:
		return imgui.KeyAppBack
	case sdl.K_AC_FORWARD:
		return imgui.KeyAppForward
	default:
		break
	}

	// Fallback to scancode
	switch scancode {
	case sdl.SCANCODE_GRAVE:
		return imgui.KeyGraveAccent
	case sdl.SCANCODE_MINUS:
		return imgui.KeyMinus
	case sdl.SCANCODE_EQUALS:
		return imgui.KeyEqual
	case sdl.SCANCODE_LEFTBRACKET:
		return imgui.KeyLeftBracket
	case sdl.SCANCODE_RIGHTBRACKET:
		return imgui.KeyRightBracket
	case sdl.SCANCODE_NONUSBACKSLASH:
		return imgui.KeyOem102
	case sdl.SCANCODE_BACKSLASH:
		return imgui.KeyBackslash
	case sdl.SCANCODE_SEMICOLON:
		return imgui.KeySemicolon
	case sdl.SCANCODE_APOSTROPHE:
		return imgui.KeyApostrophe
	case sdl.SCANCODE_COMMA:
		return imgui.KeyComma
	case sdl.SCANCODE_PERIOD:
		return imgui.KeyPeriod
	case sdl.SCANCODE_SLASH:
		return imgui.KeySlash
	default:
		break
	}
	return imgui.KeyNone
}

func ImGui_ImplSDL3_Init(window *sdl.Window, renderer *sdl.Renderer, sdl_gl_context unsafe.Pointer) bool {
	io := imgui.CurrentIO()

	// IMGUI_CHECKVERSION()
	if io.BackendPlatformUserData != nil {
		panic("Already initialized a platform backend!")
	}

	ver_linked := sdl.GetVersion()

	// Setup backend capabilities flags
	bd := &ImGui_ImplSDL3_Data{}

	bd.BackendPlatformName = fmt.Sprintf("imgui_impl_sdl3 (%d)", ver_linked)

	// backendPlatformName := fmt.Sprintf("imgui_impl_sdl3 (%d.%d.%d; %d.%d.%d)",
	// 	SDL_MAJOR_VERSION, SDL_MINOR_VERSION, SDL_MICRO_VERSION, SDL_VERSIONNUM_MAJOR(ver_linked), SDL_VERSIONNUM_MINOR(ver_linked), SDL_VERSIONNUM_MICRO(ver_linked))

	io.SetBackendRendererUserData(uintptr(unsafe.Pointer(bd)))
	io.SetBackendPlatformName(bd.BackendPlatformName)
	io.SetBackendFlags(io.BackendFlags() | imgui.BackendFlagsHasMouseCursors | imgui.BackendFlagsHasSetMousePos)

	// We can honor GetMouseCursor() values (optional)
	// We can honor io.WantSetMousePos requests (optional, rarely used)

	bd.Window = window
	bd.WindowID = Must(window.ID())
	bd.Renderer = renderer

	// Check and store if we are on a SDL backend that supports SDL_GetGlobalMouseState() and SDL_CaptureMouse()
	// ("wayland" and "rpi" don't support it, but we chose to use a white-list instead of a black-list)
	bd.MouseCanUseGlobalState = false
	bd.MouseCanUseCapture = false

	// SDL_HAS_CAPTURE_AND_GLOBAL_MOUSE
	if true {
		sdl_backend := sdl.GetCurrentVideoDriver()
		var capture_and_global_state_whitelist []string = []string{"windows", "cocoa", "x11", "DIVE", "VMAN"}
		if slices.Contains(capture_and_global_state_whitelist, sdl_backend) {
			bd.MouseCanUseGlobalState = true
			bd.MouseCanUseCapture = true
		}
	}

	// TODO actually implement this
	//var platform_io *imgui.PlatformIO = imgui.CurrentPlatformIO()
	// platform_io.CData.Platform_SetClipboardTextFn = C.ImGui_ImplSDL3_SetClipboardText
	// platform_io.CData.Platform_GetClipboardTextFn = C.ImGui_ImplSDL3_GetClipboardText
	// platform_io.CData.Platform_SetImeDataFn = C.ImGui_ImplSDL3_PlatformSetImeData
	// platform_io.CData.Platform_OpenInShellFn = func(context *imgui.Context, cstring url) { return sdl.OpenURL(url) == 0 }

	// // Gamepad handling
	// bd.GamepadMode = imgui.ImplSDL3_GamepadMode_AutoFirst

	bd.WantUpdateGamepadsList = true

	sdl.GetCursor()

	// Load mouse cursors
	bd.MouseCursors[imgui.MouseCursorArrow] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_DEFAULT))
	bd.MouseCursors[imgui.MouseCursorTextInput] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_TEXT))
	bd.MouseCursors[imgui.MouseCursorResizeAll] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_MOVE))
	bd.MouseCursors[imgui.MouseCursorResizeNS] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NS_RESIZE))
	bd.MouseCursors[imgui.MouseCursorResizeEW] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_EW_RESIZE))
	bd.MouseCursors[imgui.MouseCursorResizeNESW] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NESW_RESIZE))
	bd.MouseCursors[imgui.MouseCursorResizeNWSE] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NWSE_RESIZE))
	bd.MouseCursors[imgui.MouseCursorHand] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_POINTER))
	bd.MouseCursors[imgui.MouseCursorWait] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_WAIT))
	bd.MouseCursors[imgui.MouseCursorProgress] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_PROGRESS))
	bd.MouseCursors[imgui.MouseCursorNotAllowed] = Must(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NOT_ALLOWED))

	// Set platform dependent data in viewport
	// Our mouse update function expect PlatformHandle to be filled for the main viewport

	// TODO, what's a platform handle?
	// main_viewport := imgui.MainViewport()
	// imgui.ImplSDL3_SetupPlatformHandles(main_viewport, window)

	// From 2.0.5: Set SDL hint to receive mouse click events on window focus, otherwise SDL doesn't emit the event.
	// Without this, when clicking to gain focus, our widgets wouldn't activate even though they showed as hovered.
	// (This is unfortunately a global SDL setting, so enabling it might have a side-effect on your application.
	// It is unlikely to make a difference, but if your app absolutely needs to ignore the initial on-focus click:
	// you can ignore SDL_EVENT_MOUSE_BUTTON_DOWN events coming right after a SDL_WINDOWEVENT_FOCUS_GAINED)
	sdl.SetHint(sdl.HINT_MOUSE_FOCUS_CLICKTHROUGH, "1")

	// From 2.0.22: Disable auto-capture, this is preventing drag and drop across multiple windows (see #5710)
	// sdl.SetHint(sdl.HINT_MOUSE_AUTO_CAPTURE, "0")

	return true
}
