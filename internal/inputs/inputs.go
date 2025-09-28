package inputs

import (
	"log"
	"os"

	"github.com/jakebark/corset/internal/config"
	"github.com/spf13/pflag"
)

type UserInput struct {
	Target      string
	Replace     bool
	Whitespace  bool
	IsDirectory bool
	MaxFiles    int
}

// ParseFlags returns pased CLI flags and arguments
func isDirectory(target string) bool {
	info, _ := os.Stat(target)
	return info.IsDir()
}

func ParseFlags() UserInput {
	var replace bool
	var whitespace bool

	pflag.BoolVarP(&replace, "replace", "r", false, "replace old files")
	pflag.BoolVarP(&whitespace, "whitespace", "w", false, "retain whitespace")
	pflag.Parse()

	if pflag.NArg() < 1 {
		log.Fatal("Error: Please specify a directory or file")
	}
	target := pflag.Arg(0)
	return UserInput{
		Target:      target,
		Replace:     replace,
		Whitespace:  whitespace,
		IsDirectory: isDirectory(target),
		MaxFiles:    config.DefaultMaxFiles,
	}
}
