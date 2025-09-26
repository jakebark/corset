package core

const (
	// DefaultMaxFiles is the default maximum number of files to split policies across
	DefaultMaxFiles = 5

	// MaxAllowedFiles is the maximum number of files allowed (AWS OU limit)
	MaxAllowedFiles = 5

	// MaxPolicySize is the AWS SCP character limit
	MaxPolicySize = 5120

	// CorsetSuffix is appended to output filenames
	CorsetSuffix = "_corset"

	// SCPBaseStructure is the minified base SCP structure
	SCPBaseStructure = `{"Version":"2012-10-17","Statement":[]}`

	// SCPBaseWithWS is the formatted base SCP structure with whitespace
	SCPBaseWithWS = `{
  "Version": "2012-10-17",
  "Statement": []
}`
)
