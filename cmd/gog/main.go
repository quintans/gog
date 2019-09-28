package main

import (
	"flag"
	"fmt"
	"os"

	"github.con/quintans/gog/generator"

	_ "github.con/quintans/gog/plugins"
)

const version = "0.0.2"

var (
	fileName = flag.String("f", "", "file name to be parsed, overriding the environment variable GOFILE value")
	recur    = flag.Bool("r", false, "scan current dir and sub directories")
	ver      = flag.Bool("v", false, "version")
)

func main() {
	flag.Parse()

	if *ver {
		fmt.Println("gog version", version)
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
