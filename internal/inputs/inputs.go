package inputs

import (
	"log"
	"os"

	"github.com/jakebark/corset/internal/config"
	"github.com/spf13/pflag"
)

type UserInput struct {
	Target      string
	Delete      bool
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
	var delete bool
	var whitespace bool

	pflag.BoolVarP(&delete, "delete", "d", false, "Delete old files")
	pflag.BoolVarP(&whitespace, "whitespace", "w", false, "retain whitespace")
	pflag.Parse()

	if pflag.NArg() < 1 {
		log.Fatal("Error: Please specify a directory or file")
	}
	target := pflag.Arg(0)
	return UserInput{
		Target:      target,
		Delete:      delete,
		Whitespace:  whitespace,
		IsDirectory: isDirectory(target),
		MaxFiles:    config.DefaultMaxFiles,
	}
}
