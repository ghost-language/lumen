package modules

import (
	"os"
	"path/filepath"
	"strings"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
)

var FilesystemMethods = map[string]*object.LibraryFunction{}
var FilesystemProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(FilesystemMethods, "setIdentity", filesystemSetIdentityMethod)
	modules.RegisterMethod(FilesystemMethods, "getSaveDirectory", filesystemGetSaveDirectoryMethod)
	modules.RegisterMethod(FilesystemMethods, "write", filesystemWriteMethod)
	modules.RegisterMethod(FilesystemMethods, "append", filesystemAppendMethod)
	modules.RegisterMethod(FilesystemMethods, "read", filesystemReadMethod)
	modules.RegisterMethod(FilesystemMethods, "exists", filesystemExistsMethod)
	modules.RegisterMethod(FilesystemMethods, "remove", filesystemRemoveMethod)
	modules.RegisterMethod(FilesystemMethods, "getDirectoryItems", filesystemGetDirectoryItemsMethod)
	modules.RegisterMethod(FilesystemMethods, "createDirectory", filesystemCreateDirectoryMethod)
	modules.RegisterMethod(FilesystemMethods, "readAsset", filesystemReadAssetMethod)
}

// Ghost's io module reads and writes next to a game's source files, which is the
// wrong place for saved games: a game installed read-only cannot write there,
// and a player's progress does not belong inside the program. This module writes
// to the platform's per-user config directory instead, under a name the game
// chooses with setIdentity().

// filesystemSetIdentityMethod names the folder a game's saves live in. Call it
// once during load(), before writing anything.
func filesystemSetIdentityMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.setIdentity", tok, args, 1); err != nil {
		return err
	}

	identity, err := text("filesystem.setIdentity", tok, args, 0)

	if err != nil {
		return err
	}

	if identityErr := engine.Lumen.SetSaveIdentity(identity); identityErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.setIdentity() %s", tok.Line, tok.Column, identityErr)
	}

	return value.NULL
}

func filesystemGetSaveDirectoryMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	directory, err := engine.Lumen.SaveDirectory()

	if err != nil {
		return object.NewError("%d:%d: runtime error: filesystem.getSaveDirectory() %s", tok.Line, tok.Column, err)
	}

	return &object.String{Value: directory}
}

func filesystemWriteMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return writeSaveFile("filesystem.write", tok, args, false)
}

func filesystemAppendMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return writeSaveFile("filesystem.append", tok, args, true)
}

func writeSaveFile(name string, tok token.Token, args []object.Object, appending bool) object.Object {
	if err := arity(name, tok, args, 2); err != nil {
		return err
	}

	relative, err := text(name, tok, args, 0)

	if err != nil {
		return err
	}

	contents, err := text(name, tok, args, 1)

	if err != nil {
		return err
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: %s() %s", tok.Line, tok.Column, name, pathErr)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
		return object.NewError("%d:%d: runtime error: %s() %s", tok.Line, tok.Column, name, mkdirErr)
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC

	if appending {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}

	file, openErr := os.OpenFile(path, flags, 0o644)

	if openErr != nil {
		return object.NewError("%d:%d: runtime error: %s() %s", tok.Line, tok.Column, name, openErr)
	}

	defer file.Close()

	if _, writeErr := file.WriteString(contents); writeErr != nil {
		return object.NewError("%d:%d: runtime error: %s() %s", tok.Line, tok.Column, name, writeErr)
	}

	return value.NULL
}

// filesystemReadMethod reads a saved file. A file that does not exist reads as
// null rather than an error, so a game can treat "no save yet" as a normal case.
func filesystemReadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.read", tok, args, 1); err != nil {
		return err
	}

	relative, err := text("filesystem.read", tok, args, 0)

	if err != nil {
		return err
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.read() %s", tok.Line, tok.Column, pathErr)
	}

	contents, readErr := os.ReadFile(path)

	if readErr != nil {
		if os.IsNotExist(readErr) {
			return value.NULL
		}

		return object.NewError("%d:%d: runtime error: filesystem.read() %s", tok.Line, tok.Column, readErr)
	}

	return &object.String{Value: string(contents)}
}

func filesystemExistsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.exists", tok, args, 1); err != nil {
		return err
	}

	relative, err := text("filesystem.exists", tok, args, 0)

	if err != nil {
		return err
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.exists() %s", tok.Line, tok.Column, pathErr)
	}

	_, statErr := os.Stat(path)

	return &object.Boolean{Value: statErr == nil}
}

func filesystemRemoveMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.remove", tok, args, 1); err != nil {
		return err
	}

	relative, err := text("filesystem.remove", tok, args, 0)

	if err != nil {
		return err
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.remove() %s", tok.Line, tok.Column, pathErr)
	}

	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return object.NewError("%d:%d: runtime error: filesystem.remove() %s", tok.Line, tok.Column, removeErr)
	}

	return value.NULL
}

func filesystemGetDirectoryItemsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	relative := ""

	if len(args) == 1 {
		given, err := text("filesystem.getDirectoryItems", tok, args, 0)

		if err != nil {
			return err
		}

		relative = given
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.getDirectoryItems() %s", tok.Line, tok.Column, pathErr)
	}

	entries, readErr := os.ReadDir(path)

	if readErr != nil {
		if os.IsNotExist(readErr) {
			return &object.List{Elements: []object.Object{}}
		}

		return object.NewError("%d:%d: runtime error: filesystem.getDirectoryItems() %s", tok.Line, tok.Column, readErr)
	}

	elements := make([]object.Object, 0, len(entries))

	for _, entry := range entries {
		elements = append(elements, &object.String{Value: entry.Name()})
	}

	return &object.List{Elements: elements}
}

func filesystemCreateDirectoryMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.createDirectory", tok, args, 1); err != nil {
		return err
	}

	relative, err := text("filesystem.createDirectory", tok, args, 0)

	if err != nil {
		return err
	}

	path, pathErr := savePath(relative)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.createDirectory() %s", tok.Line, tok.Column, pathErr)
	}

	if mkdirErr := os.MkdirAll(path, 0o755); mkdirErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.createDirectory() %s", tok.Line, tok.Column, mkdirErr)
	}

	return value.NULL
}

// filesystemReadAssetMethod reads a file shipped with the game, relative to the
// game's source directory. It is the read-only counterpart to the save methods,
// and is what map and dialogue data should be loaded with.
func filesystemReadAssetMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("filesystem.readAsset", tok, args, 1); err != nil {
		return err
	}

	relative, err := text("filesystem.readAsset", tok, args, 0)

	if err != nil {
		return err
	}

	contents, readErr := os.ReadFile(resolvePath(relative))

	if readErr != nil {
		return object.NewError("%d:%d: runtime error: filesystem.readAsset() %s", tok.Line, tok.Column, readErr)
	}

	return &object.String{Value: string(contents)}
}

// savePath resolves a game-relative save path, refusing anything that would
// escape the save directory. A save name comes from game code, but game code
// can build one out of player input, and a slot named "../../.bashrc" should
// not be able to reach outside the folder the game owns.
func savePath(relative string) (string, error) {
	directory, err := engine.Lumen.SaveDirectory()

	if err != nil {
		return "", err
	}

	cleaned := filepath.Clean(filepath.Join(directory, relative))

	if cleaned != directory && !strings.HasPrefix(cleaned, directory+string(os.PathSeparator)) {
		return "", os.ErrPermission
	}

	return cleaned, nil
}
