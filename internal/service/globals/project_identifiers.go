package globals

import "strings"

func identifyMakefileProject(path string, subfiles []string, subdirs []string) ([]string, error) {
	_ = path
	_ = subdirs
	if containsAnyFileName(subfiles, "Makefile", "makefile", "GNUmakefile") {
		return []string{"@makefile"}, nil
	}
	return nil, nil
}

func identifyByFileNames(tag string, names ...string) ProjectIdentifier {
	return func(path string, subfiles []string, subdirs []string) ([]string, error) {
		_ = path
		_ = subdirs
		if containsAnyFileName(subfiles, names...) {
			return []string{tag}, nil
		}
		return nil, nil
	}
}

func identifyByMatcher(tag string, match func(string) bool) ProjectIdentifier {
	return func(path string, subfiles []string, subdirs []string) ([]string, error) {
		_ = path
		_ = subdirs
		for _, name := range subfiles {
			if match(name) {
				return []string{tag}, nil
			}
		}
		return nil, nil
	}
}

func containsFileName(subfiles []string, name string) bool {
	for _, fileName := range subfiles {
		if fileName == name {
			return true
		}
	}
	return false
}

func containsAnyFileName(subfiles []string, names ...string) bool {
	for _, name := range names {
		if containsFileName(subfiles, name) {
			return true
		}
	}
	return false
}

func isTSConfigFileName(name string) bool {
	if name == "tsconfig.json" {
		return true
	}
	if strings.HasSuffix(name, ".tsconfig.json") {
		return true
	}
	return strings.HasPrefix(name, "tsconfig.") && strings.HasSuffix(name, ".json")
}

func isDockerfileName(name string) bool {
	switch name {
	case "Dockerfile", "Containerfile", "Podmanfile":
		return true
	}
	if strings.HasPrefix(name, "Dockerfile.") || strings.HasPrefix(name, "Containerfile.") {
		return true
	}
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".dockerfile")
}

func isDockerComposeFileName(name string) bool {
	switch name {
	case "compose.yml", "compose.yaml",
		"docker-compose.yml", "docker-compose.yaml",
		"podman-compose.yml", "podman-compose.yaml":
		return true
	}
	if !hasYAMLExtension(name) {
		return false
	}
	if strings.HasPrefix(name, "docker-compose.") || strings.HasPrefix(name, "docker-compose-") {
		return true
	}
	if strings.HasPrefix(name, "compose.") || strings.HasPrefix(name, "compose-") {
		return true
	}
	return strings.HasPrefix(name, "podman-compose.")
}

func isGradleFileName(name string) bool {
	switch name {
	case "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts":
		return true
	}
	return false
}

func hasYAMLExtension(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

func isNamedConfigFile(name, stem string) bool {
	prefix := stem + "."
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	ext := strings.TrimPrefix(name, prefix)
	if ext == "" {
		return false
	}
	switch ext {
	case "ts", "js", "mjs", "cjs", "mts", "cts", "json":
		return true
	}
	return false
}

func hasNamedConfigFile(subfiles []string, stem string) bool {
	for _, name := range subfiles {
		if isNamedConfigFile(name, stem) {
			return true
		}
	}
	return false
}
