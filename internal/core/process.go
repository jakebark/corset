package core

import (
	"fmt"
	"github.com/jakebark/corset/internal/inputs"
)

type Policy struct {
	Version   string                   `json:"Version"`
	Statement []map[string]interface{} `json:"Statement"`
}

type Statement struct {
	Content map[string]interface{}
	Size    int
}

type WriteResult struct {
	Filename   string
	Size       int
	Statements int
}

func ProcessFiles(userInput inputs.UserInput, files []string) {
	allStatements := extractAllStatements(files)
	if len(allStatements) == 0 {
		fmt.Println("No policy statements found")
		return
	}

	packedFiles := packAllStatements(userInput, allStatements)
	writeOutputFiles(userInput, packedFiles, files)
}
