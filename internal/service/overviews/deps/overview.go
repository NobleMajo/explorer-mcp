package deps

import "github.com/NobleMajo/explorer-mcp/internal/service/globals"

type Settings struct {
	ShowGoToolDeps bool
}

func DefaultSettings() Settings {
	return Settings{ShowGoToolDeps: true}
}

func buildDependencies(projectRootPath string, settings Settings, verbose bool) ([]string, error) {
	_ = verbose
	return globals.CollectManifestDependencies(projectRootPath, globals.ManifestDepsSettings{
		ShowGoToolDeps: settings.ShowGoToolDeps,
	})
}
