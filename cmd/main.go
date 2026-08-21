package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"ghostlang.org/x/ghost/ghost"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/lumen/engine"
	"ghostlang.org/x/lumen/modules"
)

var (
	flagVersion bool
	flagHelp    bool
)

func init() {
	flag.Usage = func() {
		showHelp()
		os.Exit(0)
	}

	flag.BoolVar(&flagHelp, "h", false, "display help information")
	flag.BoolVar(&flagVersion, "v", false, "display version information")
}

func main() {
	// Subcommands are matched before flags are parsed. Go's flag package stops
	// at the first non-flag argument, so `lumen package mygame -o out` would
	// otherwise leave the -o unparsed and silently ignored.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "package":
			run(packageGame(os.Args[2:]))
		case "build":
			run(buildGame("build", os.Args[2:]))
		case "fuse":
			// The older name for the same thing. `build` is what everyone
			// reaches for first, so that is the one the help talks about, and
			// this stays an alias so existing scripts keep working.
			run(buildGame("fuse", os.Args[2:]))
		}
	}

	flag.Parse()

	args := flag.Args()

	if flagVersion {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), engine.Version)
		os.Exit(0)
	}

	if flagHelp {
		showHelp()
		os.Exit(0)
	}

	source, directory, err := readSource(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "lumen: %s\n", err)
		os.Exit(1)
	}

	lumen := engine.New("Lumen")

	modules.Register()

	lumen.Ghost = ghost.New()
	lumen.Ghost.SetSource(source)
	lumen.Ghost.SetDirectory(directory)

	// The game's source runs before the loop starts. It defines the callbacks
	// the loop will drive and does any set-up that does not need the window.
	if _, failed := lumen.Ghost.Execute().(*object.Error); failed {
		os.Exit(1)
	}

	lumen.Run()
}

// run reports an error from a subcommand and exits, since a subcommand never
// falls through to starting a game.
func run(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "lumen: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// packageGame builds a .lumen archive from a game directory.
func packageGame(args []string) error {
	source, target, err := subcommandArgs("package", args)

	if err != nil {
		return err
	}

	if target == "" {
		target = filepath.Base(source) + engine.ArchiveExtension
	}

	if err := engine.PackageGame(source, target); err != nil {
		return err
	}

	fmt.Printf("packaged %s -> %s\n", source, target)

	return nil
}

// buildGame builds a standalone executable from a game directory.
func buildGame(name string, args []string) error {
	source, target, err := subcommandArgs(name, args)

	if err != nil {
		return err
	}

	if target == "" {
		target = filepath.Base(source)

		if runtime.GOOS == "windows" {
			target += ".exe"
		}
	}

	if err := engine.FuseGame(source, target); err != nil {
		return err
	}

	fmt.Printf("built %s -> %s\n", source, target)

	return nil
}

// subcommandArgs reads a subcommand's game directory and optional -o output.
// The flag is accepted on either side of the directory, because both readings
// are natural and neither should be silently ignored.
func subcommandArgs(name string, args []string) (string, string, error) {
	usage := fmt.Errorf("usage: lumen %s <directory> [-o output]", name)

	directory := ""
	output := ""

	for index := 0; index < len(args); index++ {
		argument := args[index]

		switch {
		case strings.HasPrefix(argument, "-o="):
			output = strings.TrimPrefix(argument, "-o=")
		case argument == "-o":
			index++

			if index >= len(args) {
				return "", "", usage
			}

			output = args[index]
		case strings.HasPrefix(argument, "-"):
			return "", "", usage
		case directory != "":
			return "", "", usage
		default:
			directory = argument
		}
	}

	if directory == "" {
		return "", "", usage
	}

	return filepath.Clean(directory), output, nil
}

// readSource loads the game's entry file. Lumen looks for a game in four places,
// in order: a game packaged into this very binary, then whatever path was given
// on the command line, then a main.ghost next to the executable, and finally a
// main.ghost in the working directory.
func readSource(args []string) (string, string, error) {
	if directory, fused, err := engine.FusedGame(); err != nil {
		return "", "", err
	} else if fused {
		return readEntry(filepath.Join(directory, "main.ghost"))
	}

	if len(args) > 0 {
		return readTarget(args[0])
	}

	executable, err := os.Executable()

	if err != nil {
		return "", "", err
	}

	beside := filepath.Join(filepath.Dir(executable), "main.ghost")

	if _, err := os.Stat(beside); err == nil {
		return readEntry(beside)
	}

	if _, err := os.Stat("main.ghost"); err == nil {
		return readEntry("main.ghost")
	}

	return "", "", fmt.Errorf("no game given, and no main.ghost found\n\nRun `lumen -h` for usage")
}

// readTarget resolves a path that may be an archive, a directory, or a file.
func readTarget(target string) (string, string, error) {
	if engine.IsArchive(target) {
		directory, err := engine.UnpackArchive(target)

		if err != nil {
			return "", "", err
		}

		return readEntry(filepath.Join(directory, "main.ghost"))
	}

	info, err := os.Stat(target)

	// Naming something that is not there is nearly always a mistyped path or a
	// subcommand that does not exist, and neither is helped by being told that
	// a game needs a main.ghost.
	if err != nil {
		return "", "", fmt.Errorf("no such file or directory: %s\n\nRun `lumen -h` for usage", target)
	}

	// A game is usually a folder with a main.ghost in it.
	if info.IsDir() {
		return readEntry(filepath.Join(target, "main.ghost"))
	}

	return readEntry(target)
}

// readEntry reads an entry file and returns it with the directory it lives in,
// which is what asset and import paths resolve against.
func readEntry(file string) (string, string, error) {
	contents, err := os.ReadFile(file)

	if err != nil {
		if strings.HasSuffix(file, "main.ghost") {
			return "", "", fmt.Errorf("could not read %s: a game needs a main.ghost", file)
		}

		return "", "", fmt.Errorf("could not read %s: %w", file, err)
	}

	directory, err := filepath.Abs(filepath.Dir(file))

	if err != nil {
		return "", "", err
	}

	return string(contents), directory, nil
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("    lumen [flags] [file|directory|archive]")
	fmt.Println("    lumen build <directory> [-o game]")
	fmt.Println("    lumen package <directory> [-o game.lumen]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println()
	fmt.Println("    -h  show help")
	fmt.Println("    -v  show version")
	fmt.Println("    -o  output path for build and package")
	fmt.Println()
	fmt.Println("Running a game:")
	fmt.Println()
	fmt.Println("    lumen main.ghost          run a game from its entry file")
	fmt.Println("    lumen examples/60_rpg     run the main.ghost inside a directory")
	fmt.Println("    lumen game.lumen          run a packaged game")
	fmt.Println("    lumen                     run the main.ghost beside the binary,")
	fmt.Println("                              or in the working directory")
	fmt.Println()
	fmt.Println("Shipping a game:")
	fmt.Println()
	fmt.Println("    lumen build mygame -o mygame")
	fmt.Println()
	fmt.Println("            Build a standalone executable with the engine and the")
	fmt.Println("            game in one file. Players just run it. Builds for the")
	fmt.Println("            machine it runs on. (`lumen fuse` does the same thing.)")
	fmt.Println()
	fmt.Println("    lumen package mygame -o mygame.lumen")
	fmt.Println()
	fmt.Println("            Build a single-file archive. Players run it with")
	fmt.Println("            `lumen mygame.lumen`.")
	fmt.Println()
}
