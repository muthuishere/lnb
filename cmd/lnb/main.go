package main

import (
	"fmt"
	"os"
	"path/filepath"

	"lnb/internal/oshandler"
)

func usage() {
	fmt.Println("Usage: lnb <file> [install|remove]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	filename := os.Args[1]
	action := "install"
	if len(os.Args) > 2 {
		action = os.Args[2]
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	handler := oshandler.New()
	if handler == nil {
		fmt.Println("Unsupported OS")
		os.Exit(1)
	}

	if err := handler.Handle(absPath, action); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
