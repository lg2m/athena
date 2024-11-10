package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lg2m/athena/internal/athena"
	"github.com/lg2m/athena/internal/athena/config"
)

func main() {
	// Define the config flag with shorthand -c and long form --config
	configPath := flag.String("config", "", "Path to the configuration file")
	flag.StringVar(configPath, "c", "", "Path to the configuration file (shorthand)")

	// Customize the usage message to reflect optional flag
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
	cfg, errors := config.LoadConfig(configPath)
	if len(errors) > 0 {
		for _, errMsg := range errors {
			fmt.Println("Config error:", errMsg)
		}
		// Depending on your preference, you might want to exit here
		os.Exit(1)
	}

	// Initialize Athena with the loaded configuration
	// ath, err := ui.NewAthena(
	// ui.WithFilePath(filePath),
	// Pass the configuration to NewAthena if required
	// ui.WithConfig(cfg),
	// )
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
