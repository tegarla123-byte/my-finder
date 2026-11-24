package main

import (
	"os"
	"runtime"
)

func main() {
	// Win32 UI message loop and STA COM dialogs must stay pinned to the main OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	enablePerMonitorDPI()

	// Initialize COM STA for modern Windows IFileOpenDialog folder picker.
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED|COINIT_DISABLE_OLE1DDE)
	if hr == 0 || hr == 1 { // S_OK or S_FALSE
		defer procCoUninitialize.Call()
	}

	store := newStore("")
	_ = store.Load()

	app := newApp(store)
	os.Exit(app.Run())
}
