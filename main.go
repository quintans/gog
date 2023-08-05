package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/quintans/gog/config"
	"github.com/quintans/gog/generator"

	_ "github.com/quintans/gog/plugins"
)

const recurSuffix = "/..."

var (
	fileName = flag.String("f", "", "file name to be parsed, overriding the environment variable GOFILE value")
	dir      = flag.String("d", "", "dir to be parsed. If it ends with /...it will be recursive")
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
		generator.ScanAndGenerateFile(fileToParse)
		return
	}

	if *dir != "" {
		if strings.HasSuffix(*dir, recurSuffix) {
			generator.ScanDirAndSubDirs(strings.TrimSuffix(*dir, recurSuffix))
			return
		}

		generator.ScanDir(*dir)
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
