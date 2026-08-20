package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

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
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [<filename>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.BoolVar(&flagHelp, "h", false, "display help information")
	flag.BoolVar(&flagVersion, "v", false, "display version information")
}

func main() {
	flag.Parse()

	if flagVersion {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), engine.Version)
		os.Exit(0)
	}

	if flagHelp {
		showHelp()
		os.Exit(0)
	}

	source, directory, err := readSource(flag.Args())

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

// readSource loads the game's entry file. With no argument, Lumen looks for a
// main.ghost next to the executable, which is how a packaged game starts.
func readSource(args []string) (string, string, error) {
	file := ""

	if len(args) == 0 {
		executable, err := os.Executable()

		if err != nil {
			return "", "", err
		}

		file = filepath.Join(filepath.Dir(executable), "main.ghost")
	} else {
		file = args[0]
	}

	// A directory is a convenience: a game is a folder with a main.ghost in it.
	if info, err := os.Stat(file); err == nil && info.IsDir() {
		file = filepath.Join(file, "main.ghost")
	}

	contents, err := os.ReadFile(file)

	if err != nil {
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
	fmt.Println("    lumen [flags] {file|directory}")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println()
	fmt.Println("    -h  show help")
	fmt.Println("    -v  show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Println("    lumen main.ghost")
	fmt.Println()
	fmt.Println("            Run a game from its entry file")
	fmt.Println()
	fmt.Println("    lumen examples/53_top_down")
	fmt.Println()
	fmt.Println("            Run the main.ghost inside a directory")
	fmt.Println()
}
