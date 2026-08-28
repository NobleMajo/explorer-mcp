package workspacescan

import (
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

// CollectFlags returns project flags for a directory when enabled.
func CollectFlags(opts Options, absPath string, subfiles, subdirs []string) ([]string, error) {
	if !opts.CheckFlags {
		return nil, nil
	}
	return globals.CollectSiblingProjectFlags(absPath, subfiles, subdirs)
}

// HasFlags reports whether a directory has project flags.
func HasFlags(opts Options, absPath string, subfiles, subdirs []string) bool {
	flags, err := CollectFlags(opts, absPath, subfiles, subdirs)
	return err == nil && len(flags) > 0
}

// FormatPathWithFlags appends sorted project flags to a relative path.
func FormatPathWithFlags(relPath string, flags []string) string {
	if len(flags) == 0 {
		return relPath
	}
	var b strings.Builder
	b.WriteString(relPath)
	for _, flag := range flags {
		b.WriteString(" ")
		b.WriteString(flag)
	}
	return b.String()
}
