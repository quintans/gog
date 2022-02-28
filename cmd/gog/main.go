package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/quintans/gog/config"
	"github.com/quintans/gog/generator"

	_ "github.com/quintans/gog/plugins"
)

var (
	fileName = flag.String("f", "", "file name to be parsed, overriding the environment variable GOFILE value")
	recur    = flag.Bool("r", false, "scan current dir and sub directories")
	ver      = flag.Bool("v", false, "version")
)

func main() {
	flag.Parse()

	fmt.Println("gog version", config.Version)
	if *ver {
		return
	}

	fileToParse := getFileToParse()
	if fileToParse != "" {
		generator.ParseGoFileAndGenerateFile(fileToParse)
		return
	}

	if *recur {
		generator.ScanCurrentDirAndSubDirs()
		return
	}

	generator.ScanCurrentDir()
}

func getFileToParse() string {
	if *fileName != "" {
		return *fileName
	}

	return os.Getenv("GOFILE")
}
