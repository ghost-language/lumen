package modules

import (
	"net/url"
	"os/exec"
	"runtime"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
)

var SystemMethods = map[string]*object.LibraryFunction{}
var SystemProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(SystemMethods, "getClipboardText", systemGetClipboardTextMethod)
	modules.RegisterMethod(SystemMethods, "setClipboardText", systemSetClipboardTextMethod)
	modules.RegisterMethod(SystemMethods, "openUrl", systemOpenUrlMethod)
	modules.RegisterMethod(SystemMethods, "getPowerInfo", systemGetPowerInfoMethod)

	modules.RegisterProperty(SystemProperties, "os", systemOsProperty)
	modules.RegisterProperty(SystemProperties, "processorCount", systemProcessorCountProperty)
}

func systemGetClipboardTextMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	clipboard, err := sdl.GetClipboardText()

	if err != nil {
		return object.NewError("%d:%d: runtime error: system.getClipboardText() %s", tok.Line, tok.Column, err)
	}

	return &object.String{Value: clipboard}
}

func systemSetClipboardTextMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("system.setClipboardText", tok, args, 1); err != nil {
		return err
	}

	contents, err := text("system.setClipboardText", tok, args, 0)

	if err != nil {
		return err
	}

	if clipboardErr := sdl.SetClipboardText(contents); clipboardErr != nil {
		return object.NewError("%d:%d: runtime error: system.setClipboardText() %s", tok.Line, tok.Column, clipboardErr)
	}

	return value.NULL
}

// systemOpenUrlMethod hands a URL to the player's browser, for credits screens
// and links out to a game's own site. Only http and https URLs are opened:
// handing an arbitrary string to the platform opener would let a malformed or
// hostile string reach the shell as something other than a web address.
func systemOpenUrlMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("system.openUrl", tok, args, 1); err != nil {
		return err
	}

	address, err := text("system.openUrl", tok, args, 0)

	if err != nil {
		return err
	}

	parsed, parseErr := url.Parse(address)

	if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return object.NewError("%d:%d: runtime error: system.openUrl() expects an http or https URL. got=%s", tok.Line, tok.Column, address)
	}

	var command *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", parsed.String())
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", parsed.String())
	default:
		command = exec.Command("xdg-open", parsed.String())
	}

	if startErr := command.Start(); startErr != nil {
		return object.NewError("%d:%d: runtime error: system.openUrl() %s", tok.Line, tok.Column, startErr)
	}

	return value.NULL
}

// systemGetPowerInfoMethod reports battery state as [state, percent, seconds].
// A game can use it to dim effects or warn before a laptop dies mid-dungeon.
func systemGetPowerInfoMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	state, seconds, percent := sdl.GetPowerInfo()

	names := map[int]string{
		sdl.POWERSTATE_UNKNOWN:    "unknown",
		sdl.POWERSTATE_ON_BATTERY: "battery",
		sdl.POWERSTATE_NO_BATTERY: "nobattery",
		sdl.POWERSTATE_CHARGING:   "charging",
		sdl.POWERSTATE_CHARGED:    "charged",
	}

	name, ok := names[state]

	if !ok {
		name = "unknown"
	}

	return &object.List{Elements: []object.Object{
		&object.String{Value: name},
		object.NewInt(int64(percent)),
		object.NewInt(int64(seconds)),
	}}
}

func systemOsProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.String{Value: runtime.GOOS}
}

func systemProcessorCountProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(sdl.GetCPUCount()))
}
