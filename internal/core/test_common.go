package core

// testPolicy is a helper struct for testing that mirrors the Policy struct
// but can be used in test files without import cycles
type testPolicy struct {
	Version   string                   `json:"Version"`
	Statement []map[string]interface{} `json:"Statement"`
}