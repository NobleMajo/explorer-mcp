package parent

import (
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/workspacescan"
)

var siblingFlagOptions = workspacescan.Options{CheckFlags: true}

func formatSiblingProject(absPath, relPath string, subfiles, subdirs []string) string {
	flags, _ := workspacescan.CollectFlags(siblingFlagOptions, absPath, subfiles, subdirs)
	return workspacescan.FormatPathWithFlags(relPath, flags)
}

func hasSiblingProjectFlags(absPath string, subfiles, subdirs []string) bool {
	return workspacescan.HasFlags(siblingFlagOptions, absPath, subfiles, subdirs)
}

func siblingRelativePath(entry string) string {
	if i := strings.Index(entry, " @"); i >= 0 {
		return entry[:i]
	}
	return entry
}
