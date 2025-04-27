package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"framefold/pkg/framefold"
)

func main() {
	var (
		configPath   string
		sourceDir    string
		targetDir    string
		deleteSource bool
		showVersion  bool
	)

	// Parse command line flags
	flag.StringVar(&configPath, "config", "", "Path to configuration file (optional)")
	flag.StringVar(&sourceDir, "source", "", "Source directory containing photos")
	flag.StringVar(&targetDir, "target", "", "Target directory to organize photos")
	flag.BoolVar(&deleteSource, "delete-source", false, "Delete source files after successful copy (default: false)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	// Show version if requested
	if showVersion {
		fmt.Println(framefold.VersionInfo())
		os.Exit(0)
	}

	// Validate required flags
	if sourceDir == "" || targetDir == "" {
		log.Fatal("Both --source and --target flags are required")
	}

	// Load configuration
	config, err := framefold.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create and run processor
	processor, err := framefold.NewProcessor(sourceDir, targetDir, config, deleteSource)
	if err != nil {
		log.Fatal(err)
	}

	if err := processor.Process(); err != nil {
		log.Fatal(err)
	}

	// Print summary
	fmt.Println(processor.GetStats().String())
}
