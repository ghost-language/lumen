package modules

import (
	"errors"
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
		return engine.Value("filesystem.setIdentity", tok, "cannot use `%s` as a save folder", identity).
			WithHelp("the identity names one folder, so it cannot be empty or a path")
	}

	return value.NULL
}

func filesystemGetSaveDirectoryMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	directory, err := engine.Lumen.SaveDirectory()

	if err != nil {
		return engine.SystemFailure("filesystem.getSaveDirectory", tok, err)
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
		return saveFailure(name, tok, relative, pathErr)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
		return engine.SystemFailure(name, tok, mkdirErr)
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC

	if appending {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}

	file, openErr := os.OpenFile(path, flags, 0o644)

	if openErr != nil {
		return engine.SystemFailure(name, tok, openErr)
	}

	defer file.Close()

	if _, writeErr := file.WriteString(contents); writeErr != nil {
		return engine.SystemFailure(name, tok, writeErr)
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
		return saveFailure("filesystem.read", tok, relative, pathErr)
	}

	contents, readErr := os.ReadFile(path)

	if readErr != nil {
		if os.IsNotExist(readErr) {
			return value.NULL
		}

		return engine.SystemFailure("filesystem.read", tok, readErr)
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
		return saveFailure("filesystem.exists", tok, relative, pathErr)
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
		return saveFailure("filesystem.remove", tok, relative, pathErr)
	}

	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return engine.SystemFailure("filesystem.remove", tok, removeErr)
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
		return saveFailure("filesystem.getDirectoryItems", tok, relative, pathErr)
	}

	entries, readErr := os.ReadDir(path)

	if readErr != nil {
		if os.IsNotExist(readErr) {
			return &object.List{Elements: []object.Object{}}
		}

		return engine.SystemFailure("filesystem.getDirectoryItems", tok, readErr)
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
		return saveFailure("filesystem.createDirectory", tok, relative, pathErr)
	}

	if mkdirErr := os.MkdirAll(path, 0o755); mkdirErr != nil {
		return engine.SystemFailure("filesystem.createDirectory", tok, mkdirErr)
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

	resolved := resolvePath(relative)

	contents, readErr := os.ReadFile(resolved)

	if readErr != nil {
		return engine.AssetFailure("filesystem.readAsset", tok, relative, resolved, readErr)
	}

	return &object.String{Value: string(contents)}
}

// saveFailure reports a save path that could not be resolved. There are only
// two ways that happens, and they need different answers: the platform would
// not say where this user's data lives, or the game asked for somewhere outside
// the folder it owns — which is the case worth naming, because it is usually a
// slot name built out of something the player typed.
func saveFailure(name string, tok token.Token, relative string, failure error) *object.Error {
	if errors.Is(failure, os.ErrPermission) {
		return engine.Value(name, tok, "cannot reach `%s`", relative).
			WithHelp("a save path stays inside the game's own save folder, so it cannot climb out of it with `..`")
	}

	return engine.SystemFailure(name, tok, failure)
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
