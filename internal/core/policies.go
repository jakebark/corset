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

type PolicyStatement struct {
	Content          map[string]interface{}
	Size             int // character count
	OriginalFilename string
}

type Processor struct {
	userInput inputs.UserInput
}

func NewProcessor(userInput inputs.UserInput) *Processor {
	return &Processor{userInput: userInput}
}

func (p *Processor) ProcessFiles(files []string) {
	// Get user confirmation for destructive operations
	if p.userInput.Delete {
		fmt.Print("This will delete the original files. Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled.")
		}
	}

	// Collect all policies from all files
	allStatements := []PolicyStatement{}

	for _, file := range files {
		statements := p.extractIndividualPolicies(file)
		allStatements = append(allStatements, statements...)
	}

	if len(allStatements) == 0 {
		fmt.Println("No policy statements found")
		return
	}

	// Calculate base structure size
	baseSize := p.calculateBaseSize()

	// Pack policies into files using bin packing algorithm
	packedFiles := p.packPolicies(allStatements, baseSize)

	// Generate output files
	outputDir := filepath.Dir(files[0])
	if len(files) > 1 {
		// Multiple input files, use corset1.json naming
		p.generateMultipleFiles(packedFiles, outputDir)
	} else {
		// Single input file, use filename_corset.json naming
		p.generateSingleFile(packedFiles, files[0])
	}
}

func (p *Processor) extractIndividualPolicies(filename string) []PolicyStatement {
	data, _ := os.ReadFile(filename)

	var policy Policy
	json.Unmarshal(data, &policy)

	var statements []PolicyStatement
	for _, stmt := range policy.Statement {
		stmtJSON, _ := json.Marshal(stmt)

		statements = append(statements, PolicyStatement{
			Content:          stmt,
			Size:             len(stmtJSON),
			OriginalFilename: filename,
		})
	}

	return statements
}

// replace this with flat value
func (p *Processor) calculateBaseSize() int {
	if p.userInput.Whitespace {
		return len(config.SCPBaseWithWS) - 2 // Subtract the [] from Statement
	}
	return len(config.SCPBaseStructure) - 2 // Subtract the [] from Statement
}

// first fit / bin pack
func (p *Processor) packPolicies(statements []PolicyStatement, baseSize int) [][]PolicyStatement {
	// Sort policies by size (largest first) for better bin packing
	sort.Slice(statements, func(i, j int) bool {
		return statements[i].Size > statements[j].Size
	})

	files := make([][]PolicyStatement, p.userInput.MaxFiles)
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
	var result [][]PolicyStatement
	for _, file := range files {
		if len(file) > 0 {
			result = append(result, file)
		}
	}

	return result
}

func (p *Processor) generateMultipleFiles(packedFiles [][]PolicyStatement, outputDir string) {
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

func (p *Processor) generateSingleFile(packedFiles [][]PolicyStatement, originalFile string) {
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

func (p *Processor) writeOutputFile(filename string, statements []PolicyStatement) int {
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
