package parent

import (
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

func formatSiblingProject(absPath, relPath string, subfiles, subdirs []string) string {
	var b strings.Builder
	b.WriteString(relPath)

	flags, _ := globals.CollectSiblingProjectFlags(absPath, subfiles, subdirs)
	for _, flag := range flags {
		b.WriteString(" ")
		b.WriteString(flag)
	}

	return b.String()
}

func hasSiblingProjectFlags(absPath string, subfiles, subdirs []string) bool {
	flags, err := globals.CollectSiblingProjectFlags(absPath, subfiles, subdirs)
	return err == nil && len(flags) > 0
}

func siblingRelativePath(entry string) string {
	if i := strings.Index(entry, " @"); i >= 0 {
		return entry[:i]
	}
	return entry
}
