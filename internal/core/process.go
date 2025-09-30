package core

import (
	"fmt"
	"github.com/jakebark/corset/internal/inputs"
)

func ProcessFiles(userInput inputs.UserInput, files []string) {
	allStatements := extractAllStatements(files)
	if len(allStatements) == 0 {
		fmt.Println("No policy statements found")
		return
	}

	packedFiles := packAllStatements(userInput, allStatements)
	buildOutput(userInput, packedFiles, files)
}
