//go:build !wayland && (linux || freebsd || openbsd || netbsd) && !js && !wasm && !test_web_driver
// +build !wayland
// +build linux freebsd openbsd netbsd
// +build !js
// +build !wasm
// +build !test_web_driver

package glfw

import "strconv"

// GetWindowHandle returns the window handle. Only implemented for X11 currently.
func (w *window) GetWindowHandle() string {
	xid := uint(w.viewport.GetX11Window())
	return "x11:" + strconv.FormatUint(uint64(xid), 16)
}