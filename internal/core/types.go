package core

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
