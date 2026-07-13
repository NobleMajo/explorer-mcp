package parent

import (
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

func formatSiblingProject(absPath, relPath string) string {
	var b strings.Builder
	b.WriteString(relPath)

	if hasGitMetadata(absPath) {
		b.WriteString(" @git")
	}

	for _, manifestFileName := range globals.ManifestLoaderFileNames() {
		if !globals.HasManifestFile(absPath, manifestFileName) {
			continue
		}
		tag, ok := globals.ManifestLoaderTags[manifestFileName]
		if !ok || tag == "" {
			continue
		}
		b.WriteString(" ")
		b.WriteString(tag)
	}

	return b.String()
}

func siblingRelativePath(entry string) string {
	if i := strings.Index(entry, " @"); i >= 0 {
		return entry[:i]
	}
	return entry
}
