package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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

type Console struct {
	args []string
}

type Ghost *ghost.Ghost

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

	args := flag.Args()

	if flagVersion {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), "v0.0.1")
		os.Exit(0)
	}

	if flagHelp {
		showHelp()
		os.Exit(0)
	}

	console := &Console{args}

	var f *os.File
	var err error
	var rootDirectory string

	if len(console.args) == 0 {
		// Do we have a main.ghost file present?
		ex, err := os.Executable()

		if err != nil {
			panic(err)
		}

		rootDirectory = filepath.Dir(ex)
		mainFile, err := filepath.Abs(rootDirectory + "/main.ghost")

		if err != nil {
			panic(err)
		}

		f, err = os.Open(mainFile)

		if err != nil {
			log.Fatalf("could not find main.ghost: %s", err)
			showHelp()
			os.Exit(2)
		}
	} else {
		f, err = os.Open(console.args[0])

		rootDirectory = filepath.Dir(console.args[0])

		if err != nil {
			log.Fatalf("could not open file: %s: %s", err, console.args[0])
		}
	}

	b, err := io.ReadAll(f)

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read file: %s: %s", err, console.args[0])
		os.Exit(1)
	}

	lumen := engine.New("Lumen")

	// Register ghost modules
	modules.Register()

	lumen.Ghost = ghost.New()
	lumen.Ghost.SetSource(string(b))
	lumen.Ghost.SetDirectory(rootDirectory)
	result := lumen.Ghost.Execute()

	if _, ok := result.(*object.Error); ok {
		os.Exit(1)
	}

	lumen.Run()
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("    lumen [flags] {file}")
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
	fmt.Println("            Execute source file (main.ghost)")
	fmt.Println()
	fmt.Println()
}
