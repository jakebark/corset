package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jakebark/corset/internal/config"
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

type Processor struct {
	userInput inputs.UserInput
}

func NewProcessor(userInput inputs.UserInput) *Processor {
	return &Processor{userInput: userInput}
}

func (p *Processor) ProcessFiles(files []string) {
	allStatements := p.extractAllStatements(files)
	if len(allStatements) == 0 {
		fmt.Println("No policy statements found")
		return
	}

	packedFiles := p.packAllStatements(allStatements)
	p.writeOutputFiles(packedFiles, files)
}

func (p *Processor) extractAllStatements(files []string) []Statement {
	var allStatements []Statement
	for _, file := range files {
		statements := p.extractIndividualPolicies(file)
		allStatements = append(allStatements, statements...)
	}
	return allStatements
}

func (p *Processor) extractIndividualPolicies(filename string) []Statement {
	data, _ := os.ReadFile(filename)

	var policy Policy
	json.Unmarshal(data, &policy)

	var statements []Statement
	for _, stmt := range policy.Statement {
		stmtJSON, _ := json.Marshal(stmt)

		statements = append(statements, Statement{
			Content: stmt,
			Size:    len(stmtJSON),
		})
	}

	return statements
}

func (p *Processor) packAllStatements(statements []Statement) [][]Statement {
	baseSize := config.SCPBaseSizeMinified
	if p.userInput.Whitespace {
		baseSize = config.SCPBaseSizeWithWS
	}
	return p.packPolicies(statements, baseSize)
}

// first fit / bin pack
func (p *Processor) packPolicies(statements []Statement, baseSize int) [][]Statement {
	// Sort policies by size (largest first) for better bin packing
	sort.Slice(statements, func(i, j int) bool {
		return statements[i].Size > statements[j].Size
	})

	files := make([][]Statement, p.userInput.MaxFiles)
	fileSizes := make([]int, p.userInput.MaxFiles)

	// Initialize each file with base structure size
	for i := range fileSizes {
		fileSizes[i] = baseSize
	}

	// First Fit Decreasing algorithm
	for _, stmt := range statements {
		placed := false

		// Try to place in existing file with space
		for i := 0; i < p.userInput.MaxFiles; i++ {
			// Account for comma separator (except for first statement)
			separator := 0
			if len(files[i]) > 0 {
				separator = 1 // for comma
			}

			if fileSizes[i]+stmt.Size+separator <= config.MaxPolicySize {
				files[i] = append(files[i], stmt)
				fileSizes[i] += stmt.Size + separator
				placed = true
				break
			}
		}

		if !placed {
			return nil // Cannot fit all policies
		}
	}

	// Remove empty files
	var result [][]Statement
	for _, file := range files {
		if len(file) > 0 {
			result = append(result, file)
		}
	}

	return result
}

func (p *Processor) generateMultipleFiles(packedFiles [][]Statement, outputDir string) {
	fmt.Printf("Split into %d files:\n", len(packedFiles))

	for i, statements := range packedFiles {
		filename := filepath.Join(outputDir, fmt.Sprintf("corset%d.json", i+1))
		size := p.writeOutputFile(filename, statements)
		fmt.Printf("- %s (%d characters, %d statements)\n",
			filepath.Base(filename), size, len(statements))
	}

	if p.userInput.Delete {
		// Would need to track original files to delete them
		fmt.Println("Note: --delete flag not implemented for directory processing yet")
	}
}

func (p *Processor) writeOutputFiles(packedFiles [][]Statement, inputFiles []string) {
	outputDir := filepath.Dir(inputFiles[0])
	if len(inputFiles) > 1 {
		p.generateMultipleFiles(packedFiles, outputDir)
	} else {
		p.generateSingleFile(packedFiles, inputFiles[0])
	}
}

func (p *Processor) generateSingleFile(packedFiles [][]Statement, originalFile string) {
	if len(packedFiles) == 1 {
		// Single file output
		base := strings.TrimSuffix(originalFile, ".json")
		filename := base + config.CorsetSuffix + ".json"

		size := p.writeOutputFile(filename, packedFiles[0])

		fmt.Printf("%s %d characters\n", filepath.Base(filename), size)
	} else {
		// Multiple files needed
		fmt.Printf("Original Filename: %s\nSplit into %d files:\n", filepath.Base(originalFile), len(packedFiles))

		base := strings.TrimSuffix(originalFile, ".json")
		for i, statements := range packedFiles {
			filename := fmt.Sprintf("%s%s%d.json", base, config.CorsetSuffix, i+1)
			size := p.writeOutputFile(filename, statements)
			fmt.Printf("- %s (%d characters, %d statements)\n",
				filepath.Base(filename), size, len(statements))
		}
	}

	if p.userInput.Delete {
		os.Remove(originalFile)
	}
}

func (p *Processor) writeOutputFile(filename string, statements []Statement) int {
	policy := Policy{
		Version:   "2012-10-17",
		Statement: make([]map[string]interface{}, len(statements)),
	}

	for i, stmt := range statements {
		policy.Statement[i] = stmt.Content
	}

	var data []byte

	if p.userInput.Whitespace {
		data, _ = json.MarshalIndent(policy, "", "  ")
	} else {
		data, _ = json.Marshal(policy)
	}

	os.WriteFile(filename, data, 0644)

	return len(data)
}
