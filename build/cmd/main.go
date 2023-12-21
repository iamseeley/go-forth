package main

import (
	"flag"
	"log"
	"os"

	"go-forth2.0/build/internal/dev"
	"go-forth2.0/build/internal/static"
)

func main() {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	devCmd := flag.NewFlagSet("dev", flag.ExitOnError)

	if len(os.Args) < 2 {
		log.Println("expected subcommand: 'build' or 'dev'")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		static.BuildSite() // Call the build function
	case "dev":
		devCmd.Parse(os.Args[2:])
		dev.StartServer() // Call the dev function
	default:
		log.Println("expected subcommand: 'build' or 'dev'")
		os.Exit(1)
	}
}
