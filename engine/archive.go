package engine

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// A packaged game is a zip archive: the game's .gs sources and its assets, with
// main.gs at the root. The archive can be run directly, or appended to a copy of
// the Lumen binary to make a single self-contained executable.
//
// At startup a packaged game is unpacked into a cache directory and run from
// there, rather than being served out of the zip in memory. That is deliberate.
// Ghost resolves `import` and its own `io` module against the real filesystem,
// and Lumen cannot change how Ghost loads files. Serving Lumen's own loaders
// from a virtual filesystem while Ghost's imports still needed real paths would
// give a packaged game two different, silently disagreeing views of its own
// files. Unpacking costs a moment on first launch and keeps every path in the
// system pointing at the same thing.

const (
	// archiveMagic marks the trailer of a fused binary. It is written last so it
	// can be found by seeking backwards from the end of the file.
	archiveMagic = "LUMENPKG"

	// archiveTrailerSize is the magic plus the 8-byte archive offset.
	archiveTrailerSize = len(archiveMagic) + 8

	// ArchiveExtension is the suffix of a standalone packaged game.
	ArchiveExtension = ".lumen"

	// SourceExtension is the suffix Ghost source files carry. Ghost renamed it
	// from .ghost to .gs, and resolves `import "player"` to `player.gs` on its
	// own; this is only for the names Lumen itself builds or reports.
	SourceExtension = ".gs"

	// EntryFile is the file a game starts from. Every way of finding a game —
	// a directory, an archive, a fused binary, the working directory — looks
	// for this one name, so it is spelled once here rather than at each of them.
	EntryFile = "main" + SourceExtension
)

// PackageGame writes the directory at source into a .lumen archive at target.
func PackageGame(source, target string) error {
	entry := filepath.Join(source, EntryFile)

	if _, err := os.Stat(entry); err != nil {
		return fmt.Errorf("no %s in %s", EntryFile, source)
	}

	file, err := os.Create(target)

	if err != nil {
		return err
	}

	defer file.Close()

	return writeArchive(file, source)
}

// FuseGame writes a self-contained executable: a copy of the running Lumen
// binary with the game's archive appended and a trailer pointing at it. The
// operating system ignores trailing bytes in an executable, so the result runs
// as a normal program.
func FuseGame(source, target string) error {
	if _, err := os.Stat(filepath.Join(source, EntryFile)); err != nil {
		return fmt.Errorf("no %s in %s", EntryFile, source)
	}

	executable, err := os.Executable()

	if err != nil {
		return err
	}

	// Writing over the running engine would truncate the very file being read
	// from, and on most systems fails part way through with an error that says
	// nothing about why.
	if sameFile(executable, target) {
		return fmt.Errorf("refusing to build over the running engine: choose another -o path")
	}

	// A binary that already carries a game is fused again from its engine half,
	// so packaging twice does not nest one archive inside another.
	engine, engineSize, err := openEngineBinary(executable)

	if err != nil {
		return err
	}

	defer engine.Close()

	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)

	if err != nil {
		return err
	}

	// Closed explicitly at the end rather than deferred, because the file has
	// to be complete and closed before it is signed.
	if _, err := io.CopyN(output, engine, engineSize); err != nil {
		output.Close()

		return err
	}

	if err := writeArchive(output, source); err != nil {
		output.Close()

		return err
	}

	// The trailer records where the archive starts, so the runtime can find it
	// without scanning the whole file.
	trailer := make([]byte, 8)
	binary.LittleEndian.PutUint64(trailer, uint64(engineSize))

	if _, err := output.Write(trailer); err != nil {
		output.Close()

		return err
	}

	if _, err := output.WriteString(archiveMagic); err != nil {
		return err
	}

	// The file has to be closed before it can be signed: codesign rewrites it.
	if err := output.Close(); err != nil {
		return err
	}

	return signExecutable(target)
}

// signExecutable re-signs a freshly built binary on macOS. Every executable on
// Apple silicon carries a signature, appending a game to one invalidates it,
// and the system kills a binary whose signature does not match its contents —
// so a game built without this step dies instantly with nothing but "killed".
// An ad-hoc signature is enough to run locally; shipping to other machines
// wants a real developer identity and notarisation.
func signExecutable(target string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	codesign, err := exec.LookPath("codesign")

	if err != nil {
		return fmt.Errorf("built %s, but it cannot run until it is signed: codesign is not installed (it comes with the Xcode command line tools)", target)
	}

	command := exec.Command(codesign, "--force", "--sign", "-", target)

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("built %s, but signing it failed: %s: %s", target, err, strings.TrimSpace(string(output)))
	}

	return nil
}

// sameFile reports whether two paths lead to the same file, following the
// symlinks and relative paths that either of them might be written as.
func sameFile(first, second string) bool {
	firstInfo, err := os.Stat(first)

	if err != nil {
		return false
	}

	secondInfo, err := os.Stat(second)

	if err != nil {
		return false
	}

	return os.SameFile(firstInfo, secondInfo)
}

// writeArchive zips everything under source into the given writer.
func writeArchive(target io.Writer, source string) error {
	archive := zip.NewWriter(target)

	root, err := filepath.Abs(source)

	if err != nil {
		return err
	}

	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(root, path)

		if err != nil {
			return err
		}

		// Zip entries always use forward slashes, whatever built them.
		entry, err := archive.Create(filepath.ToSlash(relative))

		if err != nil {
			return err
		}

		file, err := os.Open(path)

		if err != nil {
			return err
		}

		defer file.Close()

		_, err = io.Copy(entry, file)

		return err
	})

	if walkErr != nil {
		archive.Close()

		return walkErr
	}

	return archive.Close()
}

// openEngineBinary returns a reader over the engine portion of a binary, along
// with its length. For an ordinary binary that is the whole file; for a fused
// one it is everything before the appended archive.
func openEngineBinary(path string) (*os.File, int64, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, 0, err
	}

	offset, _, err := archiveBounds(file)

	if err != nil {
		file.Close()

		return nil, 0, err
	}

	if offset < 0 {
		info, err := file.Stat()

		if err != nil {
			file.Close()

			return nil, 0, err
		}

		return file, info.Size(), nil
	}

	return file, offset, nil
}

// archiveBounds reports where an appended archive starts and how long it is, or
// an offset of -1 when the file carries no archive.
func archiveBounds(file *os.File) (int64, int64, error) {
	info, err := file.Stat()

	if err != nil {
		return 0, 0, err
	}

	size := info.Size()

	if size < int64(archiveTrailerSize) {
		return -1, 0, nil
	}

	trailer := make([]byte, archiveTrailerSize)

	if _, err := file.ReadAt(trailer, size-int64(archiveTrailerSize)); err != nil {
		return 0, 0, err
	}

	if string(trailer[8:]) != archiveMagic {
		return -1, 0, nil
	}

	offset := int64(binary.LittleEndian.Uint64(trailer[:8]))

	if offset < 0 || offset >= size-int64(archiveTrailerSize) {
		return 0, 0, errors.New("packaged game has a corrupt trailer")
	}

	return offset, size - int64(archiveTrailerSize) - offset, nil
}

// FusedGame returns the directory of the game packaged into the running binary,
// unpacking it if needed. The second return value is false when the binary
// carries no game, which is the ordinary case for the engine itself.
func FusedGame() (string, bool, error) {
	executable, err := os.Executable()

	if err != nil {
		return "", false, err
	}

	file, err := os.Open(executable)

	if err != nil {
		return "", false, err
	}

	defer file.Close()

	offset, length, err := archiveBounds(file)

	if err != nil {
		return "", false, err
	}

	if offset < 0 {
		return "", false, nil
	}

	directory, err := unpack(io.NewSectionReader(file, offset, length), length, executable)

	if err != nil {
		return "", false, err
	}

	return directory, true, nil
}

// UnpackArchive unpacks a standalone .lumen archive and returns the directory to
// run from.
func UnpackArchive(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()

	info, err := file.Stat()

	if err != nil {
		return "", err
	}

	return unpack(io.NewSectionReader(file, 0, info.Size()), info.Size(), path)
}

// IsArchive reports whether a path looks like a packaged game.
func IsArchive(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ArchiveExtension)
}

// unpack extracts an archive into a cache directory named after its contents.
// Keying the directory on a hash of the archive means a second launch of an
// unchanged game reuses what is already on disk, and a changed game never reads
// a stale mixture of old and new files.
func unpack(archive *io.SectionReader, size int64, origin string) (string, error) {
	digest := sha256.New()

	if _, err := io.Copy(digest, archive); err != nil {
		return "", err
	}

	name := hex.EncodeToString(digest.Sum(nil))[:16]

	cache, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	directory := filepath.Join(cache, "lumen", "games", name)

	// A marker written only after a successful unpack distinguishes a complete
	// extraction from one interrupted part way through.
	marker := filepath.Join(directory, ".unpacked")

	if _, err := os.Stat(marker); err == nil {
		return directory, nil
	}

	if err := os.RemoveAll(directory); err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}

	reader, err := zip.NewReader(archive, size)

	if err != nil {
		return "", fmt.Errorf("%s is not a valid packaged game: %w", filepath.Base(origin), err)
	}

	for _, entry := range reader.File {
		if err := extract(entry, directory); err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(marker, []byte(name), 0o644); err != nil {
		return "", err
	}

	return directory, nil
}

// extract writes one archive entry, refusing any path that would land outside
// the destination. A zip is data from wherever the game came from, and an entry
// named "../../.bashrc" must not be able to reach the rest of the disk.
func extract(entry *zip.File, directory string) error {
	target := filepath.Join(directory, filepath.FromSlash(entry.Name))

	if target != directory && !strings.HasPrefix(target, directory+string(os.PathSeparator)) {
		return fmt.Errorf("packaged game contains an unsafe path: %s", entry.Name)
	}

	if entry.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	source, err := entry.Open()

	if err != nil {
		return err
	}

	defer source.Close()

	file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, source)

	return err
}
