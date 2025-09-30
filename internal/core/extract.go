package core

import (
	"encoding/json"
	"os"
)

func extractAllStatements(files []string) []Statement {
	var allStatements []Statement
	for _, file := range files {
		statements := extractIndividualStatements(file)
		allStatements = append(allStatements, statements...)
	}
	return allStatements
}

func extractIndividualStatements(filename string) []Statement {
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
