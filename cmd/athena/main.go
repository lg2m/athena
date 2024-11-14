package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lg2m/athena/internal/athena"
	"github.com/lg2m/athena/internal/athena/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to the configuration file (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-c config_path] <filename>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	// Check if the filename is provided
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	filePath := args[0]

	// Load the configuration
	cfg, errors := config.LoadConfig(&configPath)
	if len(errors) > 0 {
		for _, errMsg := range errors {
			fmt.Println("Config error:", errMsg)
		}
		os.Exit(1)
	}

	a, err := athena.NewAthena(cfg, filePath)
	if err != nil {
		fmt.Printf("Error initializing Athena: %v\n", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		fmt.Printf("Error running editor: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {

}
