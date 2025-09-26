package main

import (
	"log"

	"github.com/jakebark/corset/internal/core"
	"github.com/jakebark/corset/internal/inputs"
)

func main() {
	log.SetFlags(0) // remove timestamp from prints

	userInput := inputs.ParseFlags()

	var files []string
	if userInput.IsDirectory {
		files = core.FindJSONFilesInDirectory(userInput.Target)
	} else {
		files = []string{userInput.Target}
	}

	processor := core.NewProcessor(userInput)

	processor.ProcessFiles(files)
}
