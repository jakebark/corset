package config

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

	// SCPBaseSizeMinified is the character overhead for minified SCP structure (minus Statement array)
	SCPBaseSizeMinified = 37 // len(SCPBaseStructure) - 2 for []

	// SCPBaseSizeWithWS is the character overhead for formatted SCP structure (minus Statement array)
	SCPBaseSizeWithWS = 46 // len(SCPBaseWithWS) - 2 for []

	// SCPVersion is the AWS SCP policy version
	SCPVersion = "2012-10-17"
)
