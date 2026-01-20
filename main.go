package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/riadafridishibly/npmclean/tui"
)

func tempDir() string {
	if runtime.GOOS == "darwin" {
		return "/tmp"
	}
	return os.TempDir()
}

func main() {
	logFile, err := os.CreateTemp(tempDir(), "npmclean-*.log")
	if err != nil {
		log.Fatalf("Error creating log file: %v", err)
	}
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[NPMCLN] ")
	log.SetOutput(logFile)

	fmt.Println("Logfile is being written in:", logFile.Name())

	var rootDir string
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		rootDir = cwd
	}

	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", rootDir, err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Path does not exist: %s\n", absPath)
		os.Exit(1)
	}

	app := tui.NewApp(absPath)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
