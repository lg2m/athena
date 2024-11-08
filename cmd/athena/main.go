package main

import (
	"fmt"
	"os"

	"github.com/lg2m/athena/internal/ui"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ath <filename>")
		os.Exit(1)
	}

	ath, err := ui.NewAthena(
		ui.WithFilePath(os.Args[1]),
	)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		os.Exit(1)
	}
	if err := ath.Run(); err != nil {
		fmt.Printf("Error running editor: %v\n", err)
		os.Exit(1)
	}
}
